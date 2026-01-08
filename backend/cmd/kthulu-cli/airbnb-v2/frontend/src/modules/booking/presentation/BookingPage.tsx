import React, { useState } from 'react';
import { Card, Typography } from 'antd';
import { BookingList } from './components/BookingList';
import { BookingForm } from './components/BookingForm';
import { useBookings } from '../application/useBookings';
import { Booking } from '../domain/Booking';

const { Title } = Typography;

const BookingPage: React.FC = () => {
  const { data, loading, create, update, remove } = useBookings();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingBooking, setEditingBooking] = useState<Booking | undefined>(undefined);

  const handleCreate = () => {
    setEditingBooking(undefined);
    setIsModalVisible(true);
  };

  const handleEdit = (record: Booking) => {
    setEditingBooking(record);
    setIsModalVisible(true);
  };

  const handleSubmit = async (values: any) => {
    if (editingBooking) {
      await update(editingBooking.id, values);
    } else {
      await create(values);
    }
    setIsModalVisible(false);
  };

  return (
    <div className="p-6">
      <Card>
        <div className="flex justify-between items-center mb-6">
          <Title level={2}>bookings</Title>
        </div>

        <BookingList
          data={data}
          loading={loading}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={remove}
        />

        <BookingForm
          visible={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          onSubmit={handleSubmit}
          initialValues={editingBooking}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default BookingPage;
