import React, { useState } from 'react';
import { Card, Typography } from 'antd';
import { PropertyList } from './components/PropertyList';
import { PropertyForm } from './components/PropertyForm';
import { usePropertys } from '../application/usePropertys';
import { Property } from '../domain/Property';

const { Title } = Typography;

const PropertyPage: React.FC = () => {
  const { data, loading, create, update, remove } = usePropertys();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingProperty, setEditingProperty] = useState<Property | undefined>(undefined);

  const handleCreate = () => {
    setEditingProperty(undefined);
    setIsModalVisible(true);
  };

  const handleEdit = (record: Property) => {
    setEditingProperty(record);
    setIsModalVisible(true);
  };

  const handleSubmit = async (values: any) => {
    if (editingProperty) {
      await update(editingProperty.id, values);
    } else {
      await create(values);
    }
    setIsModalVisible(false);
  };

  return (
    <div className="p-6">
      <Card>
        <div className="flex justify-between items-center mb-6">
          <Title level={2}>properties</Title>
        </div>

        <PropertyList
          data={data}
          loading={loading}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={remove}
        />

        <PropertyForm
          visible={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          onSubmit={handleSubmit}
          initialValues={editingProperty}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default PropertyPage;
