import { useState } from 'react';
import { api } from '../api/client';
import bpclLogo from '../imports/Bharat_Petroleum-Logo.wine.png';

export function LoginPage({ onLogin }: { onLogin: () => void }) {
  const [employeeId, setEmployeeId] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!employeeId || !password) {
      setError('Employee ID and password are required');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const data = await api.login(employeeId, password);
      localStorage.setItem('bpcl_token', data.token);
      localStorage.setItem('bpcl_user', JSON.stringify(data.user));
      onLogin();
    } catch (err: any) {
      if (err.status === 401) {
        setError('Invalid employee ID or password');
      } else {
        setError('Login failed. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
          <div className="flex flex-col items-center mb-8">
            <img src={bpclLogo} alt="BPCL Logo" className="h-16 w-auto mb-3" />
            <h1 className="text-2xl text-gray-900">BPCL Insight</h1>
            <p className="text-sm text-gray-500 mt-1">Sales Officer Dashboard</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="block text-sm text-gray-700 mb-1">Employee ID</label>
              <input
                type="text"
                placeholder="e.g. EMP10001"
                value={employeeId}
                onChange={(e) => setEmployeeId(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none"
                onFocus={(e) => (e.target.style.boxShadow = '0 0 0 2px rgba(0,123,201,0.2)')}
                onBlur={(e) => (e.target.style.boxShadow = '')}
                autoComplete="username"
              />
            </div>

            <div>
              <label className="block text-sm text-gray-700 mb-1">Password</label>
              <input
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none"
                onFocus={(e) => (e.target.style.boxShadow = '0 0 0 2px rgba(0,123,201,0.2)')}
                onBlur={(e) => (e.target.style.boxShadow = '')}
                autoComplete="current-password"
              />
            </div>

            {error && (
              <div className="px-4 py-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 rounded-lg text-gray-900 transition-all hover:shadow-md disabled:opacity-60 disabled:cursor-not-allowed"
              style={{ backgroundColor: '#FFE000' }}
            >
              {loading ? 'Signing in…' : 'Sign In'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
