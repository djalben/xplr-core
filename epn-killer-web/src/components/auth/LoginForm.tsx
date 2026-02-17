import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../../api/axios';
import { theme } from '../../theme/theme';

interface LoginFormProps {
  onSwitchToRegister?: () => void;
}

const LoginForm: React.FC<LoginFormProps> = ({ onSwitchToRegister }) => {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!email || !password) {
      setError('Please enter email and password');
      return;
    }

    setIsLoading(true);

    try {
      const response = await axios.post(`${API_BASE_URL}/auth/login`, {
        email,
        password,
      }, {
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // Save token to localStorage
      const token = response.data.token;
      localStorage.setItem('token', token);

      setSuccess('Login successful!');

      // Redirect to Dashboard after a brief delay
      setTimeout(() => {
        navigate('/dashboard');
      }, 500);
    } catch (error: any) {
      if (error.response?.status === 401) {
        setError('Invalid email or password');
      } else {
        const errorMessage = error.response?.data?.message || error.message || 'Failed to login';
        setError(errorMessage);
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div style={{
      width: '100vw',
      height: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      backgroundColor: theme.colors.background,
      margin: 0,
      padding: 0,
      position: 'fixed',
      top: 0,
      left: 0,
      zIndex: 9999
    }}>
      {/* Web-Native Fluid Grid - Apple Premium Design */}
      <div style={{
        width: '100%',
        maxWidth: '500px',
        backgroundColor: 'transparent',
        backdropFilter: 'blur(20px)',
        borderRadius: '40px',
        padding: '48px',
        boxShadow: 'none',
        border: 'none'
      }}>
        <h1 style={{
          fontWeight: '700',
          fontSize: '64px',
          letterSpacing: '-2px',
          color: theme.colors.accent,
          textAlign: 'center',
          marginBottom: '64px',
          textTransform: 'uppercase'
        }}>
          XPLR
        </h1>

        <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '32px' }}>
          <div>
            <label style={{
              display: 'block',
              fontSize: '11px',
              fontWeight: '400',
              letterSpacing: '2px',
              color: '#71717a',
              textTransform: 'uppercase',
              marginBottom: '12px'
            }}>
              Email
            </label>
            <input
              type="email"
              placeholder="your@email.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              style={{
                width: '100%',
                height: '56px',
                padding: '0 24px',
                backgroundColor: 'rgba(255, 255, 255, 0.03)',
                border: 'none',
                borderRadius: '12px',
                color: '#ffffff',
                fontSize: '18px',
                outline: 'none'
              }}
              required
            />
          </div>

          <div>
            <label style={{
              display: 'block',
              fontSize: '11px',
              fontWeight: '400',
              letterSpacing: '2px',
              color: '#71717a',
              textTransform: 'uppercase',
              marginBottom: '12px'
            }}>
              Password
            </label>
            <div style={{ position: 'relative' }}>
              <input
                type={showPassword ? 'text' : 'password'}
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                style={{
                  width: '100%',
                  height: '56px',
                  padding: '0 56px 0 24px',
                  backgroundColor: 'rgba(255, 255, 255, 0.03)',
                  border: 'none',
                  borderRadius: '12px',
                  color: '#ffffff',
                  fontSize: '18px',
                  outline: 'none'
                }}
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                style={{
                  position: 'absolute',
                  right: '20px',
                  top: '50%',
                  transform: 'translateY(-50%)',
                  cursor: 'pointer',
                  color: '#71717a',
                  background: 'none',
                  border: 'none',
                  padding: '0',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? (
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" style={{ width: '20px', height: '20px' }}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                  </svg>
                ) : (
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" style={{ width: '20px', height: '20px' }}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                )}
              </button>
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            style={{
              width: '100%',
              height: '56px',
              backgroundColor: '#00e096',
              color: '#000000',
              fontWeight: '700',
              textTransform: 'uppercase',
              letterSpacing: '-0.5px',
              borderRadius: '12px',
              border: 'none',
              cursor: isLoading ? 'not-allowed' : 'pointer',
              opacity: isLoading ? 0.5 : 1,
              marginTop: '8px',
              transition: 'all 0.3s'
            }}
          >
            {isLoading ? 'Logging in...' : 'LOGIN'}
          </button>

          {/* Error/Success Messages */}
          {error && (
            <div style={{
              marginTop: '16px',
              padding: '12px 16px',
              backgroundColor: 'rgba(239, 68, 68, 0.1)',
              border: '1px solid rgba(239, 68, 68, 0.3)',
              borderRadius: '12px',
              color: '#ef4444',
              fontSize: '14px',
              textAlign: 'center',
              fontWeight: '500'
            }}>
              {error}
            </div>
          )}
          {success && (
            <div style={{
              marginTop: '16px',
              padding: '12px 16px',
              backgroundColor: 'rgba(34, 197, 94, 0.1)',
              border: '1px solid rgba(34, 197, 94, 0.3)',
              borderRadius: '12px',
              color: '#22c55e',
              fontSize: '14px',
              textAlign: 'center',
              fontWeight: '500'
            }}>
              {success}
            </div>
          )}
        </form>

        <div style={{ marginTop: '32px', textAlign: 'center' }}>
          <button
            onClick={onSwitchToRegister || (() => navigate('/register'))}
            style={{
              fontSize: '11px',
              fontWeight: '400',
              letterSpacing: '2px',
              color: '#71717a',
              textTransform: 'uppercase',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: '0'
            }}
          >
            No account? <span style={{ color: '#ffffff' }}>Register</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default LoginForm;
