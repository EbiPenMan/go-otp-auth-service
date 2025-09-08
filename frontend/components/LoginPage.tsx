
import React, { useState, FormEvent } from 'react';
import { useAuth } from '../context/AuthContext';
import { sendOTP, verifyOTP } from '../services/apiService';
import Spinner from './Spinner';

enum LoginStep {
  EnterPhone,
  EnterOTP,
}

const LoginPage: React.FC = () => {
  const [step, setStep] = useState<LoginStep>(LoginStep.EnterPhone);
  const [phoneNumber, setPhoneNumber] = useState('');
  const [otp, setOtp] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuth();

  const handlePhoneSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);
    try {
      await sendOTP(phoneNumber);
      setStep(LoginStep.EnterOTP);
    } catch (err) {
      setError('Failed to send OTP. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleOtpSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);
    try {
      const { token } = await verifyOTP(phoneNumber, otp);
      login(token);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center pt-16">
      <div className="w-full max-w-md p-8 space-y-8 bg-gray-800 rounded-lg shadow-lg">
        <div>
          <h2 className="text-3xl font-extrabold text-center text-white">
            Sign in to your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-400">
            {step === LoginStep.EnterPhone 
              ? 'Enter your phone number to receive an OTP' 
              : `Enter the OTP sent to ${phoneNumber}`}
          </p>
        </div>
        
        {error && <div className="p-3 text-sm text-red-200 bg-red-800 bg-opacity-50 rounded-md">{error}</div>}

        {step === LoginStep.EnterPhone ? (
          <form className="mt-8 space-y-6" onSubmit={handlePhoneSubmit}>
            <div>
              <label htmlFor="phone-number" className="sr-only">Phone Number</label>
              <input
                id="phone-number"
                name="phone-number"
                type="tel"
                autoComplete="tel"
                required
                className="w-full px-3 py-2 text-white bg-gray-700 border border-gray-600 rounded-md placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="Phone Number (e.g., +1555...)"
                value={phoneNumber}
                onChange={(e) => setPhoneNumber(e.target.value)}
              />
            </div>
            <button
              type="submit"
              disabled={isLoading}
              className="w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 focus:ring-offset-gray-900 disabled:bg-indigo-800 disabled:cursor-not-allowed"
            >
              {isLoading ? <Spinner /> : 'Send OTP'}
            </button>
          </form>
        ) : (
          <form className="mt-8 space-y-6" onSubmit={handleOtpSubmit}>
            <div>
              <label htmlFor="otp" className="sr-only">OTP</label>
              <input
                id="otp"
                name="otp"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                required
                className="w-full px-3 py-2 text-white bg-gray-700 border border-gray-600 rounded-md placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="6-digit OTP (use 123456)"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
              />
            </div>
            <div className="flex items-center justify-between">
              <button
                type="button"
                onClick={() => { setStep(LoginStep.EnterPhone); setError(null); }}
                className="text-sm font-medium text-indigo-400 hover:text-indigo-300"
              >
                Change phone number
              </button>
            </div>
            <button
              type="submit"
              disabled={isLoading}
              className="w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 focus:ring-offset-gray-900 disabled:bg-indigo-800 disabled:cursor-not-allowed"
            >
              {isLoading ? <Spinner /> : 'Verify & Sign In'}
            </button>
          </form>
        )}
      </div>
    </div>
  );
};

export default LoginPage;
