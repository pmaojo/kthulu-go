import { useState, useEffect, useCallback } from 'react';
import { Booking } from '../domain/Booking';
import { BookingService } from '../infrastructure/BookingService';
import { message } from 'antd';

export const useBookings = () => {
  const [data, setData] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchBookings = useCallback(async () => {
    setLoading(true);
    try {
      const result = await BookingService.list();
      setData(result);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch bookings');
      message.error('Failed to load bookings');
    } finally {
      setLoading(false);
    }
  }, []);

  const createBooking = async (booking: Omit<Booking, 'id'>) => {
    try {
      await BookingService.create(booking);
      message.success('Booking created successfully');
      fetchBookings();
    } catch (err: any) {
      message.error('Failed to create Booking');
      throw err;
    }
  };

  const updateBooking = async (id: number, booking: Partial<Booking>) => {
    try {
      await BookingService.update(id, booking);
      message.success('Booking updated successfully');
      fetchBookings();
    } catch (err: any) {
      message.error('Failed to update Booking');
      throw err;
    }
  };

  const deleteBooking = async (id: number) => {
    try {
      await BookingService.delete(id);
      message.success('Booking deleted successfully');
      fetchBookings();
    } catch (err: any) {
      message.error('Failed to delete Booking');
      throw err;
    }
  };

  useEffect(() => {
    fetchBookings();
  }, [fetchBookings]);

  return {
    data,
    loading,
    error,
    refresh: fetchBookings,
    create: createBooking,
    update: updateBooking,
    remove: deleteBooking,
  };
};
