import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { NeuralBackground } from '../components/neural-background';
import { Wifi, Eye, EyeOff, Lock, Mail, ChevronRight, CreditCard, ArrowLeft } from 'lucide-react';
import { login, register } from '../api/auth';

type AuthMode = 'login' | 'register';

export const AuthPage = () => {
  const navigate = useNavigate();
  const [mode, setMode] = useState<AuthMode>('login');
  const [showPassword, setShowPassword] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email || !password) {
      setError('Заполните все поля');
      return;
    }

    if (mode === 'register' && password !== confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    setIsLoading(true);
    try {
      console.log('[Auth] Submitting:', { mode, email, password: '***' });
      if (mode === 'login') {
        const res = await login({ email, password });
        console.log('[Auth] Login response:', res);
      } else {
        const res = await register({ email, password });
        console.log('[Auth] Register response:', res);
      }
      // Token is saved by auth.ts, verify it's there
      const savedToken = localStorage.getItem('token');
      console.log('[Auth] Token saved:', savedToken ? 'yes' : 'NO');
      if (!savedToken) {
        setError('Сервер не вернул токен');
        return;
      }
      navigate('/dashboard');
    } catch (err: any) {
      console.error('[Auth] Error:', err.response?.status, err.response?.data, err.message);
      // Backend returns plain text errors via http.Error(), not JSON
      const data = err.response?.data;
      let msg: string;
      if (typeof data === 'string' && data.trim().length > 0) {
        msg = data.trim();
      } else if (data?.message) {
        msg = data.message;
      } else {
        msg = err.message || 'Ошибка авторизации';
      }
      setError(msg);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-[#0a0a0f] via-[#0f0f18] to-[#12121a] flex items-center justify-center p-4 overflow-hidden">
      <NeuralBackground />

      <div className="relative z-10 w-full max-w-md animate-fade-in-up">
        {/* Back to landing */}
        <Link to="/">
          <button className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors mb-6 text-sm font-medium">
            <ArrowLeft className="w-4 h-4" />
            На главную
          </button>
        </Link>

        {/* Logo */}
        <div className="flex justify-center mb-8">
          <div className="flex items-center gap-3">
            <div className="relative w-12 h-12 rounded-xl gradient-accent flex items-center justify-center overflow-hidden shadow-lg shadow-blue-500/30">
              <div className="absolute inset-0 bg-gradient-to-tr from-transparent via-white/20 to-transparent" />
              <span className="text-white font-bold text-2xl tracking-tighter">X</span>
            </div>
            <span className="text-3xl font-bold tracking-tight gradient-text">XPLR</span>
          </div>
        </div>

        {/* Mode Toggle */}
        <div className="flex justify-center mb-8">
          <div className="glass-card p-1.5 flex shadow-lg">
            <button
              onClick={() => { setMode('login'); setError(''); }}
              className={`px-8 py-3 text-sm font-semibold rounded-xl transition-all duration-300 ${
                mode === 'login'
                  ? 'gradient-accent text-white shadow-lg shadow-blue-500/25'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              Вход
            </button>
            <button
              onClick={() => { setMode('register'); setError(''); }}
              className={`px-8 py-3 text-sm font-semibold rounded-xl transition-all duration-300 ${
                mode === 'register'
                  ? 'gradient-accent text-white shadow-lg shadow-blue-500/25'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              Регистрация
            </button>
          </div>
        </div>

        {/* Error message */}
        {error && (
          <div className="mb-4 p-3 rounded-xl bg-red-500/15 border border-red-500/30 text-red-400 text-sm text-center">
            {error}
          </div>
        )}

        {/* Card Form - 3D Bank Card Style */}
        <div className="auth-card-3d">
          <form
            onSubmit={handleSubmit}
            className="auth-card-inner relative rounded-[24px] p-8 overflow-hidden
              bg-gradient-to-br from-blue-600 via-indigo-600 to-violet-700
              shadow-2xl shadow-blue-500/20
              border border-white/20"
          >
            {/* Card shine overlay */}
            <div className="absolute inset-0 bg-gradient-to-br from-white/20 via-transparent to-transparent pointer-events-none" />

            {/* Card pattern */}
            <div className="absolute inset-0 opacity-[0.06]">
              <svg width="100%" height="100%" viewBox="0 0 400 400">
                <defs>
                  <pattern id="auth-pattern" patternUnits="userSpaceOnUse" width="30" height="30" patternTransform="rotate(45)">
                    <circle cx="15" cy="15" r="1" fill="white" />
                  </pattern>
                </defs>
                <rect width="100%" height="100%" fill="url(#auth-pattern)" />
              </svg>
            </div>

            {/* Card Content */}
            <div className="relative z-10">
              {/* Header with logo and chip */}
              <div className="flex items-start justify-between mb-8">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-xl bg-white/15 flex items-center justify-center backdrop-blur-sm border border-white/20">
                    <CreditCard className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <span className="text-xl font-bold text-white block">XPLR</span>
                    <span className="text-xs text-white/60">Финтех платформа</span>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <div className="nfc-waves">
                    <Wifi className="w-6 h-6 text-white/50 rotate-90" />
                  </div>
                  <div className="chip-effect w-12 h-9 flex items-center justify-center rounded-md">
                    <div className="w-8 h-5 border border-yellow-600/50 rounded-sm bg-gradient-to-br from-yellow-200/30 to-transparent" />
                  </div>
                </div>
              </div>

              {/* Form Fields */}
              <div className="space-y-4">
                {/* Email */}
                <div className="relative group">
                  <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/50 group-focus-within:text-white transition-colors" />
                  <input
                    type="email"
                    placeholder="Email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-white/10 border border-white/20 rounded-xl py-4 pl-12 pr-4 text-white placeholder-white/40 focus:outline-none focus:border-white/50 focus:bg-white/15 transition-all text-base font-mono tracking-wide backdrop-blur-sm"
                    required
                  />
                </div>

                {/* Password */}
                <div className="relative group">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/50 group-focus-within:text-white transition-colors" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Пароль"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full bg-white/10 border border-white/20 rounded-xl py-4 pl-12 pr-14 text-white placeholder-white/40 focus:outline-none focus:border-white/50 focus:bg-white/15 transition-all text-base font-mono tracking-widest backdrop-blur-sm"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-4 top-1/2 -translate-y-1/2 text-white/50 hover:text-white transition-colors p-1"
                  >
                    {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>

                {/* Confirm Password - register only */}
                {mode === 'register' && (
                  <div className="relative group">
                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/50 group-focus-within:text-white transition-colors" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      placeholder="Подтвердите пароль"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="w-full bg-white/10 border border-white/20 rounded-xl py-4 pl-12 pr-4 text-white placeholder-white/40 focus:outline-none focus:border-white/50 focus:bg-white/15 transition-all text-base font-mono tracking-widest backdrop-blur-sm"
                      required
                    />
                  </div>
                )}
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={isLoading}
                className="w-full mt-8 py-4 rounded-xl font-semibold text-lg
                  bg-white text-blue-600
                  hover:bg-slate-100
                  disabled:opacity-60 disabled:cursor-not-allowed
                  transition-all duration-300
                  shadow-lg
                  flex items-center justify-center gap-2 group
                  relative overflow-hidden"
              >
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-blue-100 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-700" />
                <span className="relative z-10">
                  {isLoading ? 'Загрузка...' : mode === 'login' ? 'Войти в аккаунт' : 'Создать аккаунт'}
                </span>
                {!isLoading && <ChevronRight className="w-5 h-5 relative z-10 group-hover:translate-x-1 transition-transform" />}
              </button>

              {/* Card footer */}
              <div className="mt-6 pt-4 border-t border-white/20 flex items-center justify-between text-xs text-white/50">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                  <span>Безопасное соединение</span>
                </div>
                <span className="font-mono">XPLR v2.0</span>
              </div>
            </div>
          </form>
        </div>

        {/* Additional links */}
        <div className="mt-6 text-center">
          {mode === 'register' && (
            <p className="text-sm text-slate-500">
              Регистрируясь, вы соглашаетесь с{' '}
              <a href="#" className="text-blue-400 hover:text-blue-300 transition-colors">условиями использования</a>
            </p>
          )}
        </div>

        {/* Features */}
        <div className="mt-8 grid grid-cols-3 gap-4 text-center">
          <div className="glass-card p-4">
            <div className="text-2xl mb-1">🔒</div>
            <p className="text-xs text-slate-400">Безопасность</p>
          </div>
          <div className="glass-card p-4">
            <div className="text-2xl mb-1">💳</div>
            <p className="text-xs text-slate-400">Виртуальные карты</p>
          </div>
          <div className="glass-card p-4">
            <div className="text-2xl mb-1">⚡</div>
            <p className="text-xs text-slate-400">Мгновенно</p>
          </div>
        </div>
      </div>

      <style>{`
        @keyframes fade-in-up {
          from { opacity: 0; transform: translateY(30px); }
          to { opacity: 1; transform: translateY(0); }
        }
        .animate-fade-in-up {
          animation: fade-in-up 0.8s ease-out forwards;
        }
      `}</style>
    </div>
  );
};
