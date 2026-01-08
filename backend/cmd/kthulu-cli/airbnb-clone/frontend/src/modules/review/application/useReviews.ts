import { useState, useEffect, useCallback } from 'react';
import { Review } from '../domain/Review';
import { ReviewService } from '../infrastructure/ReviewService';
import { message } from 'antd';

export const useReviews = () => {
  const [data, setData] = useState<Review[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchReviews = useCallback(async () => {
    setLoading(true);
    try {
      const result = await ReviewService.list();
      setData(result);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch reviews');
      message.error('Failed to load reviews');
    } finally {
      setLoading(false);
    }
  }, []);

  const createReview = async (review: Omit<Review, 'id'>) => {
    try {
      await ReviewService.create(review);
      message.success('Review created successfully');
      fetchReviews();
    } catch (err: any) {
      message.error('Failed to create Review');
      throw err;
    }
  };

  const updateReview = async (id: number, review: Partial<Review>) => {
    try {
      await ReviewService.update(id, review);
      message.success('Review updated successfully');
      fetchReviews();
    } catch (err: any) {
      message.error('Failed to update Review');
      throw err;
    }
  };

  const deleteReview = async (id: number) => {
    try {
      await ReviewService.delete(id);
      message.success('Review deleted successfully');
      fetchReviews();
    } catch (err: any) {
      message.error('Failed to delete Review');
      throw err;
    }
  };

  useEffect(() => {
    fetchReviews();
  }, [fetchReviews]);

  return {
    data,
    loading,
    error,
    refresh: fetchReviews,
    create: createReview,
    update: updateReview,
    remove: deleteReview,
  };
};
