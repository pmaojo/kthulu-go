import React, { useState } from 'react';
import { Card, Typography } from 'antd';
import { TodoList } from './components/TodoList';
import { TodoForm } from './components/TodoForm';
import { useTodos } from '../application/useTodos';
import { Todo } from '../domain/Todo';

const { Title } = Typography;

const TodoPage: React.FC = () => {
  const { data, loading, create, update, remove } = useTodos();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingTodo, setEditingTodo] = useState<Todo | undefined>(undefined);

  const handleCreate = () => {
    setEditingTodo(undefined);
    setIsModalVisible(true);
  };

  const handleEdit = (record: Todo) => {
    setEditingTodo(record);
    setIsModalVisible(true);
  };

  const handleSubmit = async (values: any) => {
    if (editingTodo) {
      await update(editingTodo.id, values);
    } else {
      await create(values);
    }
    setIsModalVisible(false);
  };

  return (
    <div className="p-6">
      <Card>
        <div className="flex justify-between items-center mb-6">
          <Title level={2}>todos</Title>
        </div>

        <TodoList
          data={data}
          loading={loading}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={remove}
        />

        <TodoForm
          visible={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          onSubmit={handleSubmit}
          initialValues={editingTodo}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default TodoPage;
