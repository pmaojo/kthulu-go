import React from 'react';
import { Table, Button, Space, Popconfirm } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { User } from '../../domain/User';

interface UserListProps {
  data: User[];
  loading: boolean;
  onEdit: (record: User) => void;
  onDelete: (id: number) => void;
  onCreate: () => void;
}

export const UserList: React.FC<UserListProps> = ({
  data,
  loading,
  onEdit,
  onDelete,
  onCreate,
}) => {
  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: User) => (
        <Space>
          <Button
            icon={<EditOutlined />}
            onClick={() => onEdit(record)}
            type="text"
          />
          <Popconfirm
            title={`Are you sure you want to delete this ${record.id}?`}
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
      <div style={ { marginBottom: 16, display: 'flex', justifyContent: 'flex-end' } }>
        <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>
          Add User
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
