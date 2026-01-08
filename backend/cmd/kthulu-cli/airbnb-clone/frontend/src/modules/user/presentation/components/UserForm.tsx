import React, { useEffect } from 'react';
import { Modal, Form, Input, InputNumber, Switch, DatePicker } from 'antd';
import { User } from '../../domain/User';
import dayjs from 'dayjs';

interface UserFormProps {
  visible: boolean;
  onCancel: () => void;
  onSubmit: (values: any) => void;
  initialValues?: User;
  loading?: boolean;
}

export const UserForm: React.FC<UserFormProps> = ({
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
      title={initialValues ? `Edit ${initialValues.id}` : 'Create User'}
      onCancel={onCancel}
      onOk={handleOk}
      confirmLoading={loading}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="name"
          label="Name"
          rules={[{ required: true, message: 'Please input Name!' }]}
          valuePropName="value"
        >
          <Input />
        </Form.Item>
      </Form>
    </Modal>
  );
};
