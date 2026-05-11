import { useState } from 'react';
import { api } from '../../api/client';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from './ui/dialog';
import { InputOTP, InputOTPGroup, InputOTPSlot } from './ui/input-otp';

type Step = 'email' | 'otp' | 'done';

export function ForgotPasswordModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [step, setStep] = useState<Step>('email');
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  function reset() {
    setStep('email');
    setEmail('');
    setOtp('');
    setNewPassword('');
    setError('');
    setLoading(false);
  }

  function handleClose() {
    reset();
    onClose();
  }

  async function handleSendOTP(e: React.FormEvent) {
    e.preventDefault();
    if (!email.trim()) {
      setError('Email is required');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await api.forgotPassword(email.trim());
      setStep('otp');
    } catch {
      setError('Failed to send OTP. Please try again.');
    } finally {
      setLoading(false);
    }
  }

  async function handleReset(e: React.FormEvent) {
    e.preventDefault();
    if (otp.length !== 6) {
      setError('Enter the 6-digit OTP');
      return;
    }
    if (!newPassword) {
      setError('New password is required');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await api.resetForgottenPassword(email, otp, newPassword);
      setStep('done');
    } catch (err: any) {
      if (err.status === 401) {
        setError('Invalid or expired OTP. Please request a new one.');
      } else if (err.status === 400) {
        setError(err.message || 'Password must be at least 8 characters.');
      } else {
        setError('Something went wrong. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose(); }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Reset Password</DialogTitle>
          {step === 'email' && (
            <DialogDescription>
              Enter your registered email address to receive a one-time password.
            </DialogDescription>
          )}
          {step === 'otp' && (
            <DialogDescription>
              Enter the 6-digit OTP sent to <strong>{email}</strong> and choose a new password.
            </DialogDescription>
          )}
        </DialogHeader>

        {step === 'email' && (
          <form onSubmit={handleSendOTP} className="space-y-4">
            <div>
              <label className="block text-sm text-gray-700 mb-1">Email Address</label>
              <input
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none"
                autoComplete="email"
                autoFocus
              />
            </div>
            {error && (
              <p className="text-sm text-red-600">{error}</p>
            )}
            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 rounded-lg text-gray-900 transition-all hover:shadow-md disabled:opacity-60 disabled:cursor-not-allowed"
              style={{ backgroundColor: '#FFE000' }}
            >
              {loading ? 'Sending…' : 'Send OTP'}
            </button>
          </form>
        )}

        {step === 'otp' && (
          <form onSubmit={handleReset} className="space-y-4">
            <div>
              <label className="block text-sm text-gray-700 mb-2">One-Time Password</label>
              <InputOTP maxLength={6} value={otp} onChange={setOtp}>
                <InputOTPGroup>
                  {[0, 1, 2, 3, 4, 5].map((i) => (
                    <InputOTPSlot key={i} index={i} />
                  ))}
                </InputOTPGroup>
              </InputOTP>
            </div>
            <div>
              <label className="block text-sm text-gray-700 mb-1">New Password</label>
              <input
                type="password"
                placeholder="New password (min 8 chars)"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none"
                autoComplete="new-password"
              />
            </div>
            {error && (
              <p className="text-sm text-red-600">{error}</p>
            )}
            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 rounded-lg text-gray-900 transition-all hover:shadow-md disabled:opacity-60 disabled:cursor-not-allowed"
              style={{ backgroundColor: '#FFE000' }}
            >
              {loading ? 'Resetting…' : 'Reset Password'}
            </button>
            <button
              type="button"
              onClick={() => { setStep('email'); setOtp(''); setError(''); }}
              className="w-full py-2 text-sm text-gray-500 hover:text-gray-700"
            >
              ← Back
            </button>
          </form>
        )}

        {step === 'done' && (
          <div className="space-y-4 text-center">
            <p className="text-green-700 font-medium">Password reset successfully!</p>
            <p className="text-sm text-gray-600">You can now sign in with your new password.</p>
            <button
              type="button"
              onClick={handleClose}
              className="w-full py-3 rounded-lg text-gray-900 transition-all hover:shadow-md"
              style={{ backgroundColor: '#FFE000' }}
            >
              Back to Sign In
            </button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
