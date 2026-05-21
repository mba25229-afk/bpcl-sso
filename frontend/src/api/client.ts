const API_BASE = import.meta.env.VITE_API_URL || '';

interface RequestOptions extends RequestInit {
  body?: any;
}

async function request(endpoint: string, options: RequestOptions = {}) {
  const { body, headers: customHeaders, ...rest } = options;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(customHeaders as Record<string, string>),
  };

  const token = localStorage.getItem('token');
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...rest,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

export const api = {
  // Auth
  login: (employeeId: string, password: string) =>
    request('/api/v1/auth/login', {
      method: 'POST',
      body: { employee_id: employeeId, password },
    }),

  me: () => request('/api/v1/auth/me'),

  forgotPassword: (email: string) =>
    request('/api/v1/auth/forgot-password', {
      method: 'POST',
      body: { email },
    }),

  resetPassword: (email: string, otp: string, newPassword: string) =>
    request('/api/v1/auth/reset-password', {
      method: 'POST',
      body: { email, otp, new_password: newPassword },
    }),

  // Outlets
  getOutlets: () => request('/api/v1/outlets'),

  getOutlet: (cc: string) => request(`/api/v1/outlets/${cc}`),

  getPerformance: (cc: string) => request(`/api/v1/outlets/${cc}/performance`),

  getAnalysis: (cc: string) => request(`/api/v1/outlets/${cc}/analysis`),

  getTargets: (cc: string) => request(`/api/v1/outlets/${cc}/targets`),

  setTargets: (cc: string, targets: any) =>
    request(`/api/v1/outlets/${cc}/targets`, {
      method: 'PUT',
      body: targets,
    }),

  // Territory
  getTerritoryOutlets: () => request('/api/v1/territory/outlets'),

  getTerritorySummary: () => request('/api/v1/territory/summary'),

  // Competition
  getLeaderboard: () => request('/api/v1/competition/leaderboard'),

  getDealerScorecard: (competitionId: string, cc: string) =>
    request(`/api/v1/competition/${competitionId}/dealers/${cc}`),

  // Admin
  getUsers: () => request('/api/v1/admin/users'),

  createUser: (user: any) =>
    request('/api/v1/admin/users', { method: 'POST', body: user }),

  updateUser: (id: string, user: any) =>
    request(`/api/v1/admin/users/${id}`, { method: 'PUT', body: user }),

  resetUserPassword: (id: string) =>
    request(`/api/v1/admin/users/${id}/reset-password`, { method: 'PUT' }),

  getDealers: () => request('/api/v1/admin/dealers'),

  getETLStatus: () => request('/api/v1/admin/etl/status'),

  triggerETL: () => request('/api/v1/admin/etl/trigger', { method: 'POST' }),

  // Crystal
  getDashboard: () => request('/api/v1/dashboard'),

  getScorecard: (ccCode: string) => request(`/api/v1/portal/${ccCode}/scorecard`),

  getDealersList: () => request('/api/v1/crystal/dealers'),

  getMTD: (ccCode: string) => request(`/api/v1/crystal/dealers/${ccCode}/mtd`),
};
