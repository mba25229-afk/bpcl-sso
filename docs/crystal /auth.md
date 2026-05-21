# Auth Design — JWT + 20-Minute Inactivity Logout

## Decision
**JWT (stateless)**. Tokens signed with HS256. No server-side session store.

**Why not sessions?** Crystal has no horizontal scaling complexity yet. JWT is simpler for the cron/API pattern. If multi-device invalidation becomes a requirement, revisit.

---

## Token Architecture

```
Access Token  → short-lived (20 min) — used for every API call
Refresh Token → long-lived (7 days) — used only to get a new access token
```

### Why two tokens?
A single 20-min JWT would log users out mid-work every 20 minutes regardless of activity. That's not "inactivity logout" — that's forced logout. The correct pattern:
- **Access token** expires in 20 min
- **Frontend resets the 20-min inactivity timer on any user interaction**
- If user is active → silently refresh the access token before it expires
- If user is inactive for 20 min → do NOT refresh → next API call fails → redirect to login

---

## Implementation

### Backend: `src/middleware/auth.ts`

```typescript
import jwt from 'jsonwebtoken';
import { Request, Response, NextFunction } from 'express';

const ACCESS_EXPIRY  = '20m';
const REFRESH_EXPIRY = '7d';
const SECRET = process.env.JWT_SECRET!; // must be set — crash on missing

if (!SECRET) throw new Error('JWT_SECRET env var is required');

export function signAccessToken(payload: { userId: string; role: string }) {
  return jwt.sign(payload, SECRET, { expiresIn: ACCESS_EXPIRY });
}

export function signRefreshToken(payload: { userId: string }) {
  return jwt.sign(payload, SECRET, { expiresIn: REFRESH_EXPIRY });
}

export function verifyToken(req: Request, res: Response, next: NextFunction) {
  const header = req.headers['authorization'];
  if (!header?.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Missing token' });
  }
  try {
    const token = header.split(' ')[1];
    const decoded = jwt.verify(token, SECRET);
    (req as any).user = decoded;
    next();
  } catch (err) {
    // TokenExpiredError is the inactivity case — return 401, frontend handles redirect
    return res.status(401).json({ error: 'Token expired or invalid' });
  }
}
```

### Backend: `/api/auth/refresh` route

```typescript
// POST /api/auth/refresh
// Body: { refreshToken: string }
// Returns: { accessToken: string }
router.post('/refresh', async (req, res) => {
  const { refreshToken } = req.body;
  if (!refreshToken) return res.status(400).json({ error: 'Missing refresh token' });
  try {
    const decoded = jwt.verify(refreshToken, SECRET) as { userId: string };
    const accessToken = signAccessToken({ userId: decoded.userId, role: 'user' });
    res.json({ accessToken });
  } catch {
    res.status(401).json({ error: 'Invalid or expired refresh token — re-login required' });
  }
});
```

---

### Frontend: Inactivity Timer — `src/hooks/useInactivityLogout.ts`

```typescript
import { useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

const INACTIVITY_MS = 20 * 60 * 1000; // 20 minutes
const WARN_BEFORE_MS = 60 * 1000;      // warn 1 min before logout
const ACTIVITY_EVENTS = ['mousedown', 'keydown', 'touchstart', 'scroll'];

export function useInactivityLogout() {
  const navigate    = useNavigate();
  const logoutTimer = useRef<ReturnType<typeof setTimeout>>();
  const warnTimer   = useRef<ReturnType<typeof setTimeout>>();

  const resetTimers = useCallback(() => {
    clearTimeout(logoutTimer.current);
    clearTimeout(warnTimer.current);

    warnTimer.current = setTimeout(() => {
      // Show a "You'll be logged out in 1 minute" toast
      window.dispatchEvent(new CustomEvent('session-warning'));
    }, INACTIVITY_MS - WARN_BEFORE_MS);

    logoutTimer.current = setTimeout(() => {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      navigate('/login?reason=inactivity');
    }, INACTIVITY_MS);
  }, [navigate]);

  useEffect(() => {
    resetTimers();
    ACTIVITY_EVENTS.forEach(e => window.addEventListener(e, resetTimers));
    return () => {
      clearTimeout(logoutTimer.current);
      clearTimeout(warnTimer.current);
      ACTIVITY_EVENTS.forEach(e => window.removeEventListener(e, resetTimers));
    };
  }, [resetTimers]);
}
```

### Frontend: Silent Token Refresh — `src/lib/api.ts`

```typescript
// Axios interceptor — silently refreshes access token before it expires
// Called on every 401 response

import axios from 'axios';

const api = axios.create({ baseURL: '/api' });

api.interceptors.request.use(config => {
  const token = localStorage.getItem('accessToken');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  res => res,
  async error => {
    const original = error.config;
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true;
      const refresh = localStorage.getItem('refreshToken');
      if (!refresh) {
        window.location.href = '/login?reason=session_expired';
        return Promise.reject(error);
      }
      try {
        const { data } = await axios.post('/api/auth/refresh', { refreshToken: refresh });
        localStorage.setItem('accessToken', data.accessToken);
        original.headers.Authorization = `Bearer ${data.accessToken}`;
        return api(original);
      } catch {
        localStorage.clear();
        window.location.href = '/login?reason=session_expired';
      }
    }
    return Promise.reject(error);
  }
);

export default api;
```

---

## Forgot Password Flow

```
1. User submits email → POST /api/auth/forgot-password
2. Backend: generate 6-char OTP, store hash in DB with 15-min expiry
3. Send OTP via email (use Nodemailer + SMTP env vars)
4. User submits OTP + new password → POST /api/auth/reset-password
5. Backend: verify OTP hash, check expiry, bcrypt new password, invalidate OTP
```

```sql
CREATE TABLE password_reset_tokens (
  id         SERIAL PRIMARY KEY,
  user_id    INTEGER REFERENCES users(id),
  token_hash TEXT NOT NULL,       -- bcrypt hash of OTP
  expires_at TIMESTAMPTZ NOT NULL,
  used       BOOLEAN DEFAULT FALSE
);
```

**Do not** email JWT reset links. OTP is simpler and safer for this user base.

---

## Hard Rules
1. `JWT_SECRET` minimum 32 chars, random, never committed to git.
2. Refresh tokens stored in `localStorage` is acceptable for this internal tool. For public-facing: `httpOnly` cookie.
3. Access token = 20 min. Do not change without updating the inactivity hook.
4. On logout (manual or inactivity): clear both tokens from localStorage immediately.
5. The `/api/auth/refresh` endpoint must NOT be protected by `verifyToken` middleware.
