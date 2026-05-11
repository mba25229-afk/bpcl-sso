import { useEffect, useRef, useCallback } from 'react';

const INACTIVITY_MS = 20 * 60 * 1000;
const WARN_BEFORE_MS = 60 * 1000;
const ACTIVITY_EVENTS = ['mousedown', 'keydown', 'touchstart', 'scroll'] as const;

export function useInactivityLogout(onLogout: () => void) {
  const logoutTimer = useRef<ReturnType<typeof setTimeout>>();
  const warnTimer = useRef<ReturnType<typeof setTimeout>>();

  const resetTimers = useCallback(() => {
    clearTimeout(logoutTimer.current);
    clearTimeout(warnTimer.current);

    warnTimer.current = setTimeout(() => {
      window.dispatchEvent(new CustomEvent('session-warning'));
    }, INACTIVITY_MS - WARN_BEFORE_MS);

    logoutTimer.current = setTimeout(() => {
      localStorage.removeItem('bpcl_token');
      localStorage.removeItem('bpcl_user');
      onLogout();
    }, INACTIVITY_MS);
  }, [onLogout]);

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
