const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface ApiError extends Error {
  status: number;
  code: string;
}

function makeApiError(msg: string, status: number, code: string): ApiError {
  return Object.assign(new Error(msg), { status, code }) as ApiError;
}

async function refreshTokens(): Promise<string | null> {
  const refresh = localStorage.getItem('refreshToken');
  if (!refresh) return null;
  try {
    const res = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh }),
    });
    if (!res.ok) return null;
    const data = await res.json();
    localStorage.setItem('accessToken', data.access_token);
    return data.access_token;
  } catch {
    return null;
  }
}

function clearSession() {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
  localStorage.removeItem('bpcl_user');
}

async function request<T = any>(
  path: string,
  options: RequestInit = {},
  retry = true,
): Promise<T> {
  const token = localStorage.getItem('accessToken');
  const headers: Record<string, string> = {
    ...(options.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(options.headers as Record<string, string> || {}),
  };

  const res = await fetch(`${BASE_URL}${path}`, { ...options, headers });

  if (res.status === 401 && retry) {
    const newToken = await refreshTokens();
    if (!newToken) {
      clearSession();
      window.location.href = '/login?reason=session_expired';
      throw makeApiError('Session expired', 401, 'SESSION_EXPIRED');
    }
    return request<T>(path, options, false);
  }

  if (res.status === 401) {
    clearSession();
    window.location.href = '/login?reason=session_expired';
    throw makeApiError('Unauthorized', 401, 'UNAUTHORIZED');
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: 'UNKNOWN' }));
    throw makeApiError(err.error || 'API error', res.status, err.code || 'UNKNOWN');
  }

  return res.status === 204 ? (null as T) : res.json();
}

export const crystalApi = {
  login: (employeeId: string, password: string) =>
    request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ employee_id: employeeId, password }),
    }),

  refresh: (refreshToken: string) =>
    request('/api/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    }),

  me: () => request('/api/v1/auth/me'),

  listDealers: () => request('/api/v1/crystal/dealers'),

  getMTD: (ccCode: string, date?: string) =>
    request(`/api/v1/crystal/dealers/${ccCode}/mtd${date ? `?date=${date}` : ''}`),

  getMTDDashboard: (month?: string) =>
    request(`/api/v1/crystal/dashboard${month ? `?month=${month}` : ''}`),

  getCronStatus: () => request('/api/v1/health/cron-status'),
};

export default crystalApi;
