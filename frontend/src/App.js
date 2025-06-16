import React, { useState, useEffect, useCallback } from 'react';
import './App.css';

const API_URL = 'http://localhost:8080/api/users';

function App() {
  const [users, setUsers] = useState([]);
  const [formData, setFormData] = useState({ id: null, name: '', email: '' });
  const [isEditing, setIsEditing] = useState(false);

  const fetchUsers = useCallback(async () => {
    try {
      const response = await fetch(API_URL);
      const data = await response.json();
      setUsers(data || []);
    } catch (error) {
      console.error('ユーザーの取得に失敗しました:', error);
      setUsers([]);
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.name || !formData.email) {
      alert('名前とメールアドレスは必須です。');
      return;
    }

    const method = isEditing ? 'PUT' : 'POST';
    const url = isEditing ? `${API_URL}/${formData.id}` : API_URL;

    try {
      const response = await fetch(url, {
        method: method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: formData.name, email: formData.email }),
      });

      if (!response.ok) {
        throw new Error('リクエストに失敗しました');
      }

      setFormData({ id: null, name: '', email: '' });
      setIsEditing(false);
      fetchUsers();

    } catch (error) {
      console.error('ユーザーの保存に失敗しました:', error);
    }
  };

  const handleEdit = (user) => {
    setIsEditing(true);
    setFormData({ id: user.id, name: user.name, email: user.email });
  };

  const handleDelete = async (id) => {
    if (window.confirm('本当にこのユーザーを削除しますか？')) {
      try {
        const response = await fetch(`${API_URL}/${id}`, {
          method: 'DELETE',
        });

        if (!response.ok) {
          throw new Error('削除に失敗しました');
        }
        fetchUsers();
      } catch (error) {
        console.error('ユーザーの削除に失敗しました:', error);
      }
    }
  };
  
  const cancelEdit = () => {
    setIsEditing(false);
    setFormData({ id: null, name: '', email: '' });
  };

  return (
    <div className="App">
      <h1>Go & React CRUD アプリケーション</h1>
      <form onSubmit={handleSubmit}>
        <h2>{isEditing ? 'ユーザー編集' : 'ユーザー作成'}</h2>
        <input
          type="text"
          name="name"
          placeholder="名前"
          value={formData.name}
          onChange={handleInputChange}
        />
        <input
          type="email"
          name="email"
          placeholder="メールアドレス"
          value={formData.email}
          onChange={handleInputChange}
        />
        <button type="submit">{isEditing ? '更新' : '作成'}</button>
        {isEditing && <button type="button" onClick={cancelEdit}>キャンセル</button>}
      </form>
      <h2>ユーザー一覧</h2>
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>名前</th>
            <th>メールアドレス</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr key={user.id}>
              <td>{user.id}</td>
              <td>{user.name}</td>
              <td>{user.email}</td>
              <td>
                <button onClick={() => handleEdit(user)}>編集</button>
                <button onClick={() => handleDelete(user.id)}>削除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default App;
