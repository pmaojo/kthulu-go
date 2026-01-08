import React, { useState } from 'react';
import { Card, Typography } from 'antd';
import { ReviewList } from './components/ReviewList';
import { ReviewForm } from './components/ReviewForm';
import { useReviews } from '../application/useReviews';
import { Review } from '../domain/Review';

const { Title } = Typography;

const ReviewPage: React.FC = () => {
  const { data, loading, create, update, remove } = useReviews();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingReview, setEditingReview] = useState<Review | undefined>(undefined);

  const handleCreate = () => {
    setEditingReview(undefined);
    setIsModalVisible(true);
  };

  const handleEdit = (record: Review) => {
    setEditingReview(record);
    setIsModalVisible(true);
  };

  const handleSubmit = async (values: any) => {
    if (editingReview) {
      await update(editingReview.id, values);
    } else {
      await create(values);
    }
    setIsModalVisible(false);
  };

  return (
    <div className="p-6">
      <Card>
        <div className="flex justify-between items-center mb-6">
          <Title level={2}>reviews</Title>
        </div>

        <ReviewList
          data={data}
          loading={loading}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={remove}
        />

        <ReviewForm
          visible={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          onSubmit={handleSubmit}
          initialValues={editingReview}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default ReviewPage;
