import React, { useEffect } from 'react';
import { Modal, Form, Input, InputNumber, Switch, DatePicker } from 'antd';
import { Booking } from '../../domain/Booking';
import dayjs from 'dayjs';

interface BookingFormProps {
  visible: boolean;
  onCancel: () => void;
  onSubmit: (values: any) => void;
  initialValues?: Booking;
  loading?: boolean;
}

export const BookingForm: React.FC<BookingFormProps> = ({
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
      title={initialValues ? `Edit ${initialValues.id}` : 'Create Booking'}
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
