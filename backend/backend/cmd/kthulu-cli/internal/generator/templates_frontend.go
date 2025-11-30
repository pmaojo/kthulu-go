package generator

import (
	"strings"
	"text/template"
)

// FrontendTemplateData holds data for rendering frontend templates
type FrontendTemplateData struct {
	Name       string
	Title      string // Capitalized name (e.g., "Product")
	PluralName string // Pluralized name (e.g., "products")
	Fields     []FrontendField
}

// FrontendField represents a field in the frontend entity
type FrontendField struct {
	Name     string
	Type     string
	Label    string
	Required bool
}

// ParseFrontendFields converts CLI field strings (field:type) to FrontendField structs
func ParseFrontendFields(rawFields []string) []FrontendField {
	fields := make([]FrontendField, 0, len(rawFields))
	for _, f := range rawFields {
		parts := strings.Split(f, ":")
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		typ := parts[1]
		tsType := "string"

		switch typ {
		case "int", "float", "number":
			tsType = "number"
		case "bool", "boolean":
			tsType = "boolean"
		case "time", "date":
			tsType = "string" // Dates often handled as strings in JSON
		}

		fields = append(fields, FrontendField{
			Name:     name,
			Type:     tsType,
			Label:    Capitalize(name),
			Required: true, // Default to true for simplicity
		})
	}
	return fields
}

var (
	// Domain: Entity Interface
	frontendDomainTemplate = template.Must(template.New("frontendDomain").Parse(`export interface {{.Title}} {
  id: number;
  created_at?: string;
  updated_at?: string;
{{range .Fields}}  {{.Name}}: {{.Type}};
{{end}}}

export interface {{.Title}}Filter {
  query?: string;
  page?: number;
  pageSize?: number;
}
`))

	// Infrastructure: API Service
	frontendInfraTemplate = template.Must(template.New("frontendInfra").Parse(`import { {{.Title}}, {{.Title}}Filter } from '../domain/{{.Title}}';
import { api } from '@/services/api'; // Assuming a central axios/fetch wrapper exists

const BASE_URL = '/{{.PluralName}}';

export const {{.Title}}Service = {
  list: async (filter?: {{.Title}}Filter): Promise<{{.Title}}[]> => {
    const params = new URLSearchParams();
    if (filter?.query) params.append('q', filter.query);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

    const response = await api.get<{{.Title}}[]>(` + "`" + `${BASE_URL}?${params.toString()}` + "`" + `);
    return response.data;
  },

  get: async (id: number): Promise<{{.Title}}> => {
    const response = await api.get<{{.Title}}>(` + "`" + `${BASE_URL}/${id}` + "`" + `);
    return response.data;
  },

  create: async (data: Omit<{{.Title}}, 'id'>): Promise<{{.Title}}> => {
    const response = await api.post<{{.Title}}>(BASE_URL, data);
    return response.data;
  },

  update: async (id: number, data: Partial<{{.Title}}>): Promise<{{.Title}}> => {
    const response = await api.put<{{.Title}}>(` + "`" + `${BASE_URL}/${id}` + "`" + `, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(` + "`" + `${BASE_URL}/${id}` + "`" + `);
  },
};
`))

	// Application: Custom Hook (UseCase)
	frontendApplicationTemplate = template.Must(template.New("frontendApp").Parse(`import { useState, useEffect, useCallback } from 'react';
import { {{.Title}} } from '../domain/{{.Title}}';
import { {{.Title}}Service } from '../infrastructure/{{.Title}}Service';
import { message } from 'antd';

export const use{{.Title}}s = () => {
  const [data, setData] = useState<{{.Title}}[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetch{{.Title}}s = useCallback(async () => {
    setLoading(true);
    try {
      const result = await {{.Title}}Service.list();
      setData(result);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch {{.PluralName}}');
      message.error('Failed to load {{.PluralName}}');
    } finally {
      setLoading(false);
    }
  }, []);

  const create{{.Title}} = async ({{.Name}}: Omit<{{.Title}}, 'id'>) => {
    try {
      await {{.Title}}Service.create({{.Name}});
      message.success('{{.Title}} created successfully');
      fetch{{.Title}}s();
    } catch (err: any) {
      message.error('Failed to create {{.Title}}');
      throw err;
    }
  };

  const update{{.Title}} = async (id: number, {{.Name}}: Partial<{{.Title}}>) => {
    try {
      await {{.Title}}Service.update(id, {{.Name}});
      message.success('{{.Title}} updated successfully');
      fetch{{.Title}}s();
    } catch (err: any) {
      message.error('Failed to update {{.Title}}');
      throw err;
    }
  };

  const delete{{.Title}} = async (id: number) => {
    try {
      await {{.Title}}Service.delete(id);
      message.success('{{.Title}} deleted successfully');
      fetch{{.Title}}s();
    } catch (err: any) {
      message.error('Failed to delete {{.Title}}');
      throw err;
    }
  };

  useEffect(() => {
    fetch{{.Title}}s();
  }, [fetch{{.Title}}s]);

  return {
    data,
    loading,
    error,
    refresh: fetch{{.Title}}s,
    create: create{{.Title}},
    update: update{{.Title}},
    remove: delete{{.Title}},
  };
};
`))

	// Presentation: List Component
	frontendListTemplate = template.Must(template.New("frontendList").Parse(`import React from 'react';
import { Table, Button, Space, Popconfirm } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { {{.Title}} } from '../../domain/{{.Title}}';

interface {{.Title}}ListProps {
  data: {{.Title}}[];
  loading: boolean;
  onEdit: (record: {{.Title}}) => void;
  onDelete: (id: number) => void;
  onCreate: () => void;
}

export const {{.Title}}List: React.FC<{{.Title}}ListProps> = ({
  data,
  loading,
  onEdit,
  onDelete,
  onCreate,
}) => {
  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
{{range .Fields}}    { title: '{{.Label}}', dataIndex: '{{.Name}}', key: '{{.Name}}' },
{{end}}    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: {{.Title}}) => (
        <Space>
          <Button
            icon={<EditOutlined />}
            onClick={() => onEdit(record)}
            type="text"
          />
          <Popconfirm
            title="Are you sure you want to delete this {{.Name}}?"
            onConfirm={() => onDelete(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button icon={<DeleteOutlined />} danger type="text" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>
          Add {{.Title}}
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
      />
    </div>
  );
};
`))

	// Presentation: Form Component
	frontendFormTemplate = template.Must(template.New("frontendForm").Parse(`import React, { useEffect } from 'react';
import { Modal, Form, Input, InputNumber, Switch, DatePicker } from 'antd';
import { {{.Title}} } from '../../domain/{{.Title}}';
import dayjs from 'dayjs';

interface {{.Title}}FormProps {
  visible: boolean;
  onCancel: () => void;
  onSubmit: (values: any) => void;
  initialValues?: {{.Title}};
  loading?: boolean;
}

export const {{.Title}}Form: React.FC<{{.Title}}FormProps> = ({
  visible,
  onCancel,
  onSubmit,
  initialValues,
  loading,
}) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (visible && initialValues) {
      // Handle date fields conversion if needed
      const values = { ...initialValues };
      // Example: if there are date fields
      // if (values.someDate) values.someDate = dayjs(values.someDate);
      form.setFieldsValue(values);
    } else {
      form.resetFields();
    }
  }, [visible, initialValues, form]);

  const handleOk = () => {
    form.validateFields().then((values) => {
      onSubmit(values);
    });
  };

  return (
    <Modal
      open={visible}
      title={initialValues ? 'Edit {{.Title}}' : 'Create {{.Title}}'}
      onCancel={onCancel}
      onOk={handleOk}
      confirmLoading={loading}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
{{range .Fields}}        <Form.Item
          name="{{.Name}}"
          label="{{.Label}}"
          rules={[{ required: {{.Required}}, message: 'Please input {{.Label}}!' }]}
          valuePropName="{{if eq .Type "boolean"}}checked{{else}}value{{end}}"
        >
          {{if eq .Type "number"}}<InputNumber style={{ width: '100%' }} />
          {{else if eq .Type "boolean"}}<Switch />
          {{else}}<Input />{{end}}
        </Form.Item>
{{end}}      </Form>
    </Modal>
  );
};
`))

	// Presentation: Page Component
	frontendPageTemplate = template.Must(template.New("frontendPage").Parse(`import React, { useState } from 'react';
import { Card, Typography } from 'antd';
import { {{.Title}}List } from './components/{{.Title}}List';
import { {{.Title}}Form } from './components/{{.Title}}Form';
import { use{{.Title}}s } from '../application/use{{.Title}}s';
import { {{.Title}} } from '../domain/{{.Title}}';

const { Title } = Typography;

const {{.Title}}Page: React.FC = () => {
  const { data, loading, create, update, remove } = use{{.Title}}s();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editing{{.Title}}, setEditing{{.Title}}] = useState<{{.Title}} | undefined>(undefined);

  const handleCreate = () => {
    setEditing{{.Title}}(undefined);
    setIsModalVisible(true);
  };

  const handleEdit = (record: {{.Title}}) => {
    setEditing{{.Title}}(record);
    setIsModalVisible(true);
  };

  const handleSubmit = async (values: any) => {
    if (editing{{.Title}}) {
      await update(editing{{.Title}}.id, values);
    } else {
      await create(values);
    }
    setIsModalVisible(false);
  };

  return (
    <div className="p-6">
      <Card>
        <div className="flex justify-between items-center mb-6">
          <Title level={2}>{{.PluralName}}</Title>
        </div>

        <{{.Title}}List
          data={data}
          loading={loading}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={remove}
        />

        <{{.Title}}Form
          visible={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          onSubmit={handleSubmit}
          initialValues={editing{{.Title}}}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default {{.Title}}Page;
`))

	// Index: Module Registration
	frontendIndexTemplate = template.Must(template.New("frontendIndex").Parse(`import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

// Lazy load the main page component
const {{.Title}}Page = lazy(() => import('./presentation/{{.Title}}Page'));

const module: Module = {
  routes: [
    {
      path: '/{{.PluralName}}',
      Component: {{.Title}}Page,
    },
  ],
  components: {},
};

registerModule(module);
`))
)

func GenerateFrontendContent(tmpl *template.Template, data FrontendTemplateData) (string, error) {
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
