import { renderHook } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { useInactivityLogout } from '../useInactivityLogout';

describe('useInactivityLogout', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem('bpcl_token', 'tok');
    localStorage.setItem('bpcl_user', '{}');
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    localStorage.clear();
  });

  it('clears tokens after 20 minutes of inactivity', () => {
    const onLogout = vi.fn();
    renderHook(() => useInactivityLogout(onLogout));

    vi.advanceTimersByTime(20 * 60 * 1000);

    expect(localStorage.getItem('bpcl_token')).toBeNull();
    expect(localStorage.getItem('bpcl_user')).toBeNull();
    expect(onLogout).toHaveBeenCalledTimes(1);
  });

  it('fires session-warning event at 19 minutes', () => {
    const onLogout = vi.fn();
    renderHook(() => useInactivityLogout(onLogout));

    const listener = vi.fn();
    window.addEventListener('session-warning', listener);

    vi.advanceTimersByTime(19 * 60 * 1000);
    expect(listener).toHaveBeenCalledTimes(1);

    window.removeEventListener('session-warning', listener);
  });

  it('resets timer on user interaction (does not log out after 15 min + activity + 15 min)', () => {
    const onLogout = vi.fn();
    renderHook(() => useInactivityLogout(onLogout));

    vi.advanceTimersByTime(15 * 60 * 1000);
    // simulate activity — mousedown event resets timer
    window.dispatchEvent(new MouseEvent('mousedown'));
    vi.advanceTimersByTime(15 * 60 * 1000);

    // 15 min after reset — should NOT be logged out yet (20 min inactivity required)
    expect(localStorage.getItem('bpcl_token')).toBe('tok');
    expect(onLogout).not.toHaveBeenCalled();
  });
});
