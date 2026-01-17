import React, { useEffect } from 'react';
import { Modal, Form, Input, InputNumber, Switch, DatePicker } from 'antd';
import { Todo } from '../../domain/Todo';
import dayjs from 'dayjs';

interface TodoFormProps {
  visible: boolean;
  onCancel: () => void;
  onSubmit: (values: any) => void;
  initialValues?: Todo;
  loading?: boolean;
}

export const TodoForm: React.FC<TodoFormProps> = ({
  visible,
  onCancel,
  onSubmit,
  initialValues,
  loading,
}) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (visible && initialValues) {
      const values = { ...initialValues };
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
      title={initialValues ? `Edit ${initialValues.id}` : 'Create Todo'}
      onCancel={onCancel}
      onOk={handleOk}
      confirmLoading={loading}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="title"
          label="Title"
          rules={[{ required: true, message: 'Please input Title!' }]}
          valuePropName="value"
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="completed"
          label="Completed"
          rules={[{ required: true, message: 'Please input Completed!' }]}
          valuePropName="checked"
        >
          <Switch />
          
        </Form.Item>
      </Form>
    </Modal>
  );
};
