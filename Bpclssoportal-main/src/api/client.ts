const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface ApiError extends Error {
  status: number;
  code: string;
}

async function request(path: string, options: RequestInit = {}): Promise<any> {
  const token = localStorage.getItem('bpcl_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  // For FormData, don't set Content-Type (browser sets it with boundary)
  if (options.body instanceof FormData) {
    delete headers['Content-Type'];
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: { ...headers, ...(options.headers as Record<string, string> || {}) },
  });

  if (res.status === 401) {
    localStorage.removeItem('bpcl_token');
    localStorage.removeItem('bpcl_user');
    window.dispatchEvent(new CustomEvent('bpcl-session-expired'));
    throw Object.assign(new Error('Unauthorized'), { status: 401, code: 'UNAUTHORIZED' }) as ApiError;
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: 'UNKNOWN' }));
    throw Object.assign(new Error(err.error || 'API error'), { status: res.status, code: err.code || 'UNKNOWN' }) as ApiError;
  }

  return res.status === 204 ? null : res.json();
}

export const api = {
  login: (employeeId: string, password: string) =>
    request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ employee_id: employeeId, password }),
    }),

  me: () => request('/api/v1/auth/me'),

  getOutlet: (cc: string) => request(`/api/v1/outlets/${cc}`),

  getPerformance: (cc: string, period: string) =>
    request(`/api/v1/outlets/${cc}/performance?period=${period}`),

  getAnalysis: (cc: string, from: string, to: string) =>
    request(`/api/v1/outlets/${cc}/analysis?from=${from}&to=${to}`),

  getTargets: (cc: string, period: string) =>
    request(`/api/v1/outlets/${cc}/targets?period=${period}`),

  setTargets: (cc: string, body: unknown) =>
    request(`/api/v1/outlets/${cc}/targets`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  getLeaderboard: (territory: string) =>
    request(`/api/v1/competition/leaderboard?territory=${encodeURIComponent(territory)}`),

  // Crystal scoring engine — rank_overall for this CC + month
  getCrystalScorecard: (cc: string, month: string) =>
    request(`/api/v1/portal/${cc}/scorecard?month=${month}`),

  getDealerScorecard: (cc: string, competitionId: string) =>
    request(`/api/v1/competition/dealers/${cc}/scorecard?competition_id=${competitionId}`),

  getTerritorySummary: (period: string) =>
    request(`/api/v1/territory/summary?period=${period}`),

  getTrend: (cc: string, months: number = 6) =>
    request(`/api/v1/outlets/${cc}/trend?months=${months}`),

  uploadFile: (formData: FormData) =>
    request('/api/v1/uploads', { method: 'POST', body: formData }),

  getUploads: (cc?: string) =>
    request(`/api/v1/uploads${cc ? `?cc_number=${cc}` : ''}`),

  uploadDelhiMaster: (formData: FormData) =>
    request('/api/v1/uploads/delhi-master', { method: 'POST', body: formData }),

  getMarketShareStatus: (competitionId: string) =>
    request(`/api/v1/competition/${competitionId}/market-share-status`),

  recomputeScores: (competitionId: string) =>
    request(`/api/v1/competition/${competitionId}/recompute?confirm=true`, { method: 'POST' }),

  listUsers: (params: { role?: string; territory?: string; limit?: number; offset?: number } = {}) =>
    request(`/api/v1/admin/users?${new URLSearchParams(params as any).toString()}`),

  createUser: (body: { employee_id: string; email: string; name: string; role: string; territory_code?: string; password: string }) =>
    request('/api/v1/admin/users', { method: 'POST', body: JSON.stringify(body) }),

  updateUser: (id: string, body: { name: string; role: string; territory_code?: string; is_active: boolean }) =>
    request(`/api/v1/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(body) }),

  resetPassword: (id: string, body: { new_password: string }) =>
    request(`/api/v1/admin/users/${id}/reset-password`, { method: 'PUT', body: JSON.stringify(body) }),

  forgotPassword: (email: string) =>
    request('/api/v1/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),

  resetForgottenPassword: (email: string, otp: string, newPassword: string) =>
    request('/api/v1/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ email, otp, new_password: newPassword }),
    }),
};
