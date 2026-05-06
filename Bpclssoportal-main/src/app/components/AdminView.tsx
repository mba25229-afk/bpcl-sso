import { useState, useEffect } from 'react';
import { Loader2, Plus, Edit2, KeyRound, Trash2 } from 'lucide-react';
import { api } from '../../api/client';

interface User {
  id: string;
  employee_id: string;
  email: string;
  name: string;
  role: string;
  territory_code: string | null;
  is_active: boolean;
}

export function AdminView() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [formData, setFormData] = useState({
    employee_id: '',
    email: '',
    name: '',
    role: 'ro_manager',
    territory_code: '',
    password: '',
  });
  const [passwordData, setPasswordData] = useState({ new_password: '' });

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listUsers({ limit: 100 });
      setUsers(res.users || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingUser) {
        await api.updateUser(editingUser.id, {
          name: formData.name,
          role: formData.role,
          territory_code: formData.territory_code || undefined,
          is_active: true,
        });
      } else {
        await api.createUser({
          employee_id: formData.employee_id,
          email: formData.email,
          name: formData.name,
          role: formData.role,
          territory_code: formData.territory_code || undefined,
          password: formData.password,
        });
      }
      setShowModal(false);
      setEditingUser(null);
      setFormData({ employee_id: '', email: '', name: '', role: 'ro_manager', territory_code: '', password: '' });
      loadUsers();
    } catch (err: any) {
      setError(err.message || 'Failed to save user');
    }
  };

  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingUser) return;
    try {
      await api.resetPassword(editingUser.id, passwordData);
      setShowPasswordModal(false);
      setPasswordData({ new_password: '' });
      alert('Password reset successfully');
    } catch (err: any) {
      setError(err.message || 'Failed to reset password');
    }
  };

  const openEdit = (user: User) => {
    setEditingUser(user);
    setFormData({
      employee_id: user.employee_id,
      email: user.email,
      name: user.name,
      role: user.role,
      territory_code: user.territory_code || '',
      password: '',
    });
    setShowModal(true);
  };

  const openResetPassword = (user: User) => {
    setEditingUser(user);
    setShowPasswordModal(true);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="animate-spin text-blue-600" size={40} />
      </div>
    );
  }

  if (error && !users.length) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <p className="text-red-600">{error}</p>
        <button onClick={loadUsers} className="mt-4 px-4 py-2 bg-red-600 text-white rounded-lg">
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">User Management</h2>
        <button
          onClick={() => {
            setEditingUser(null);
            setFormData({ employee_id: '', email: '', name: '', role: 'ro_manager', territory_code: '', password: '' });
            setShowModal(true);
          }}
          className="flex items-center gap-2 px-4 py-2 text-white rounded-lg"
          style={{ backgroundColor: '#007BC9' }}
        >
          <Plus size={20} />
          Add User
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">
          {error}
        </div>
      )}

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-6 py-3 text-left text-sm text-gray-700">Employee ID</th>
                <th className="px-6 py-3 text-left text-sm text-gray-700">Name</th>
                <th className="px-6 py-3 text-left text-sm text-gray-700">Role</th>
                <th className="px-6 py-3 text-left text-sm text-gray-700">Territory</th>
                <th className="px-6 py-3 text-left text-sm text-gray-700">Status</th>
                <th className="px-6 py-3 text-center text-sm text-gray-700">Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="px-6 py-4 text-gray-900">{user.employee_id}</td>
                  <td className="px-6 py-4 text-gray-900">{user.name}</td>
                  <td className="px-6 py-4 text-gray-500">{user.role}</td>
                  <td className="px-6 py-4 text-gray-500">{user.territory_code || '-'}</td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2 py-1 rounded-full text-xs ${
                        user.is_active
                          ? 'bg-green-100 text-green-700'
                          : 'bg-gray-100 text-gray-700'
                      }`}
                    >
                      {user.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-center gap-2">
                      <button
                        onClick={() => openEdit(user)}
                        className="p-2 text-gray-500 hover:text-gray-700"
                      >
                        <Edit2 size={18} />
                      </button>
                      <button
                        onClick={() => openResetPassword(user)}
                        className="p-2 text-gray-500 hover:text-gray-700"
                      >
                        <KeyRound size={18} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-xl font-bold mb-4">
              {editingUser ? 'Edit User' : 'Add User'}
            </h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              {!editingUser && (
                <>
                  <div>
                    <label className="block text-sm text-gray-700 mb-1">Employee ID</label>
                    <input
                      type="text"
                      value={formData.employee_id}
                      onChange={(e) => setFormData({ ...formData, employee_id: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm text-gray-700 mb-1">Email</label>
                    <input
                      type="email"
                      value={formData.email}
                      onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                      required
                    />
                  </div>
                </>
              )}
              <div>
                <label className="block text-sm text-gray-700 mb-1">Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-gray-700 mb-1">Role</label>
                <select
                  value={formData.role}
                  onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                >
                  <option value="ro_manager">RO Manager</option>
                  <option value="territory_manager">Territory Manager</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-700 mb-1">Territory Code</label>
                <input
                  type="text"
                  value={formData.territory_code}
                  onChange={(e) => setFormData({ ...formData, territory_code: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                />
              </div>
              {!editingUser && (
                <div>
                  <label className="block text-sm text-gray-700 mb-1">Password</label>
                  <input
                    type="password"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    required
                    minLength={8}
                  />
                  <p className="text-xs text-gray-500 mt-1">Min 8 chars, uppercase+lowercase+number</p>
                </div>
              )}
              <div className="flex justify-end gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-white rounded-lg"
                  style={{ backgroundColor: '#007BC9' }}
                >
                  Save
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showPasswordModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-xl font-bold mb-4">Reset Password</h3>
            <form onSubmit={handleResetPassword} className="space-y-4">
              <div>
                <label className="block text-sm text-gray-700 mb-1">New Password</label>
                <input
                  type="password"
                  value={passwordData.new_password}
                  onChange={(e) => setPasswordData({ ...passwordData, new_password: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                  required
                  minLength={8}
                />
              </div>
              <div className="flex justify-end gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => setShowPasswordModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-white rounded-lg"
                  style={{ backgroundColor: '#007BC9' }}
                >
                  Reset
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}