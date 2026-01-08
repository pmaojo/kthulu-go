import React, { useState } from 'react';
import { Card, Typography } from 'antd';
import { UserList } from './components/UserList';
import { UserForm } from './components/UserForm';
import { useUsers } from '../application/useUsers';
import { User } from '../domain/User';

const { Title } = Typography;

const UserPage: React.FC = () => {
  const { data, loading, create, update, remove } = useUsers();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | undefined>(undefined);

  const handleCreate = () => {
    setEditingUser(undefined);
    setIsModalVisible(true);
  };

  const handleEdit = (record: User) => {
    setEditingUser(record);
    setIsModalVisible(true);
  };

  const handleSubmit = async (values: any) => {
    if (editingUser) {
      await update(editingUser.id, values);
    } else {
      await create(values);
    }
    setIsModalVisible(false);
  };

  return (
    <div className="p-6">
      <Card>
        <div className="flex justify-between items-center mb-6">
          <Title level={2}>users</Title>
        </div>

        <UserList
          data={data}
          loading={loading}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={remove}
        />

        <UserForm
          visible={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          onSubmit={handleSubmit}
          initialValues={editingUser}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default UserPage;
