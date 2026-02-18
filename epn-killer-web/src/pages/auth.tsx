import React, { useState, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { NeuralBackground } from '../components/neural-background';
import { Wifi, Eye, EyeOff, Lock, Mail, ChevronRight, ArrowLeft, Check, X } from 'lucide-react';
import { login, register } from '../api/auth';

type AuthMode = 'login' | 'register';

/* ── Password strength rules ── */
const passwordRules = [
  { key: 'length', label: 'Минимум 8 символов', test: (p: string) => p.length >= 8 },
  { key: 'upper', label: 'Заглавная буква (A-Z)', test: (p: string) => /[A-Z]/.test(p) },
  { key: 'digit', label: 'Цифра (0-9)', test: (p: string) => /\d/.test(p) },
  { key: 'special', label: 'Спецсимвол (!@#$%^&*)', test: (p: string) => /[!@#$%^&*]/.test(p) },
] as const;

/* ── Translate common backend errors ── */
const translateError = (raw: string): string => {
  const map: Record<string, string> = {
    'Email cannot be empty': 'Email не может быть пустым',
    'Password must be at least 8 characters': 'Пароль должен содержать минимум 8 символов',
    'Invalid request body': 'Неверный формат запроса',
    'email already registered': 'Этот Email уже зарегистрирован',
    'Email already registered': 'Этот Email уже зарегистрирован',
    'invalid email or password': 'Неверный email или пароль',
    'Invalid email or password': 'Неверный email или пароль',
    'user not found': 'Пользователь не найден',
    'User not found': 'Пользователь не найден',
  };
  const trimmed = raw.trim();
  return map[trimmed] ?? trimmed;
};

export const AuthPage = () => {
  const navigate = useNavigate();
  const [mode, setMode] = useState<AuthMode>('login');
  const [showPassword, setShowPassword] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  /* Live password validation (register only) */
  const pwChecks = useMemo(
    () => passwordRules.map((r) => ({ ...r, pass: r.test(password) })),
    [password],
  );
  const allPwValid = pwChecks.every((c) => c.pass);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email || !password) {
      setError('Заполните все поля');
      return;
    }

    if (mode === 'register') {
      if (!allPwValid) {
        setError('Пароль не соответствует требованиям безопасности');
        return;
      }
      if (password !== confirmPassword) {
        setError('Пароли не совпадают');
        return;
      }
    }

    setIsLoading(true);
    try {
      const payload = { email, password };
      console.log('[Auth] Sending registration data:', { mode, email, password: '***' });

      if (mode === 'login') {
        const res = await login(payload);
        console.log('[Auth] Login response:', res);
      } else {
        const res = await register(payload);
        console.log('[Auth] Register response:', res);
      }

      const savedToken = localStorage.getItem('token');
      console.log('[Auth] Token saved:', savedToken ? 'yes' : 'NO');
      if (!savedToken) {
        setError('Сервер не вернул токен. Попробуйте ещё раз.');
        return;
      }
      navigate('/dashboard');
    } catch (err: any) {
      console.error('[Auth] Registration error details:', err.response);
      console.error('[Auth] Status:', err.response?.status, 'Data:', err.response?.data, 'Msg:', err.message);

      const data = err.response?.data;
      let msg: string;
      if (typeof data === 'string' && data.trim().length > 0) {
        msg = translateError(data);
      } else if (data?.message) {
        msg = translateError(data.message);
      } else if (err.code === 'ERR_NETWORK') {
        msg = 'Сервер недоступен. Проверьте подключение к интернету.';
      } else {
        msg = err.message || 'Ошибка авторизации';
      }
      setError(msg);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#050507] flex items-center justify-center p-4 overflow-hidden relative">
      {/* Deep dark gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-br from-[#05050a] via-[#080812] to-[#04040a]" />
      {/* Subtle radial accent */}
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(30,58,138,0.15)_0%,transparent_60%)]" />

      <NeuralBackground reducedDensity />

      <div className="relative z-10 w-full max-w-md animate-fade-in-up">
        {/* Back to landing */}
        <Link to="/">
          <button className="flex items-center gap-2 text-slate-500 hover:text-slate-300 transition-colors mb-6 text-sm font-medium">
            <ArrowLeft className="w-4 h-4" />
            На главную
          </button>
        </Link>

        {/* Logo */}
        <div className="flex justify-center mb-8">
          <div className="flex items-center gap-3">
            <div className="relative w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center overflow-hidden shadow-lg shadow-blue-500/20">
              <div className="absolute inset-0 bg-gradient-to-tr from-transparent via-white/20 to-transparent" />
              <span className="text-white font-bold text-2xl tracking-tighter relative z-10">X</span>
            </div>
            <span className="text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-400 bg-clip-text text-transparent">XPLR</span>
          </div>
        </div>

        {/* Mode Toggle */}
        <div className="flex justify-center mb-6">
          <div className="bg-white/[0.04] backdrop-blur-xl border border-white/[0.06] rounded-2xl p-1.5 flex shadow-2xl">
            <button
              type="button"
              onClick={() => { setMode('login'); setError(''); }}
              className={`px-8 py-3 text-sm font-semibold rounded-xl transition-all duration-300 ${
                mode === 'login'
                  ? 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/25'
                  : 'text-slate-500 hover:text-slate-300'
              }`}
            >
              Вход
            </button>
            <button
              type="button"
              onClick={() => { setMode('register'); setError(''); }}
              className={`px-8 py-3 text-sm font-semibold rounded-xl transition-all duration-300 ${
                mode === 'register'
                  ? 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/25'
                  : 'text-slate-500 hover:text-slate-300'
              }`}
            >
              Регистрация
            </button>
          </div>
        </div>

        {/* Error message */}
        {error && (
          <div className="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm text-center backdrop-blur-sm">
            {error}
          </div>
        )}

        {/* ── Bank Card Form (Glassmorphism) ── */}
        <div className="perspective-[1200px]">
          <form
            onSubmit={handleSubmit}
            className="relative rounded-[22px] p-8 overflow-hidden
              bg-gradient-to-br from-[#0d1528]/90 via-[#0a1025]/95 to-[#0c0f20]/90
              backdrop-blur-2xl
              shadow-[0_8px_60px_-12px_rgba(30,64,175,0.25)]
              border border-white/[0.08]
              transition-transform duration-500 hover:scale-[1.005]"
          >
            {/* Glass shine overlay */}
            <div className="absolute inset-0 bg-gradient-to-br from-white/[0.06] via-transparent to-transparent pointer-events-none rounded-[22px]" />
            {/* Subtle holographic edge */}
            <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-blue-400/30 to-transparent" />

            {/* Card pattern dots */}
            <div className="absolute inset-0 opacity-[0.03] pointer-events-none">
              <svg width="100%" height="100%" viewBox="0 0 400 400">
                <defs>
                  <pattern id="auth-dots" patternUnits="userSpaceOnUse" width="24" height="24" patternTransform="rotate(45)">
                    <circle cx="12" cy="12" r="0.8" fill="white" />
                  </pattern>
                </defs>
                <rect width="100%" height="100%" fill="url(#auth-dots)" />
              </svg>
            </div>

            {/* ── Card header: logo + chip + NFC ── */}
            <div className="relative z-10">
              <div className="flex items-start justify-between mb-8">
                {/* Left: XPLR branding */}
                <div className="flex items-center gap-3">
                  <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-blue-500/20 to-indigo-500/20 border border-white/[0.08] flex items-center justify-center backdrop-blur-sm">
                    <span className="text-blue-400 font-bold text-lg">X</span>
                  </div>
                  <div>
                    <span className="text-lg font-bold text-white/90 block tracking-wide">XPLR</span>
                    <span className="text-[10px] text-white/30 uppercase tracking-widest">Финтех платформа</span>
                  </div>
                </div>

                {/* Right: NFC + Chip */}
                <div className="flex items-center gap-3">
                  <Wifi className="w-5 h-5 text-white/20 rotate-90" />
                  {/* Card chip */}
                  <div className="w-11 h-8 rounded-md bg-gradient-to-br from-yellow-600/30 via-yellow-500/20 to-yellow-700/30 border border-yellow-500/20 flex items-center justify-center overflow-hidden relative">
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-yellow-300/10 to-transparent" />
                    <div className="w-6 h-4 border border-yellow-600/30 rounded-sm" />
                  </div>
                </div>
              </div>

              {/* ── Form Fields ── */}
              <div className="space-y-4">
                {/* Email */}
                <div className="relative group">
                  <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-400 transition-colors z-10" />
                  <input
                    type="email"
                    placeholder="Email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-white/[0.04] border border-white/[0.08] rounded-xl py-4 pl-12 pr-4
                      text-white placeholder-slate-600
                      focus:outline-none focus:border-blue-500/40 focus:bg-white/[0.06] focus:shadow-[0_0_20px_-4px_rgba(59,130,246,0.15)]
                      transition-all text-base font-mono tracking-wide"
                    required
                    autoComplete="email"
                  />
                </div>

                {/* Password */}
                <div className="relative group">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-400 transition-colors z-10" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Пароль"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full bg-white/[0.04] border border-white/[0.08] rounded-xl py-4 pl-12 pr-14
                      text-white placeholder-slate-600
                      focus:outline-none focus:border-blue-500/40 focus:bg-white/[0.06] focus:shadow-[0_0_20px_-4px_rgba(59,130,246,0.15)]
                      transition-all text-base font-mono tracking-widest"
                    required
                    autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                  />
                  <button
                    type="button"
                    onClick={(e) => { e.preventDefault(); e.stopPropagation(); setShowPassword((v) => !v); }}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white transition-colors p-2 rounded-lg hover:bg-white/[0.06] z-10"
                    tabIndex={-1}
                    aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                  >
                    {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>

                {/* Password strength (register only) */}
                {mode === 'register' && password.length > 0 && (
                  <div className="grid grid-cols-2 gap-x-4 gap-y-1.5 px-1 py-2">
                    {pwChecks.map((c) => (
                      <div key={c.key} className="flex items-center gap-1.5">
                        {c.pass
                          ? <Check className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                          : <X className="w-3.5 h-3.5 text-slate-600 shrink-0" />}
                        <span className={`text-[11px] ${c.pass ? 'text-emerald-400/80' : 'text-slate-600'}`}>{c.label}</span>
                      </div>
                    ))}
                  </div>
                )}

                {/* Confirm Password - register only */}
                {mode === 'register' && (
                  <div className="relative group">
                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-400 transition-colors z-10" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      placeholder="Подтвердите пароль"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="w-full bg-white/[0.04] border border-white/[0.08] rounded-xl py-4 pl-12 pr-4
                        text-white placeholder-slate-600
                        focus:outline-none focus:border-blue-500/40 focus:bg-white/[0.06] focus:shadow-[0_0_20px_-4px_rgba(59,130,246,0.15)]
                        transition-all text-base font-mono tracking-widest"
                      required
                      autoComplete="new-password"
                    />
                  </div>
                )}
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={isLoading}
                className="w-full mt-8 py-4 rounded-xl font-semibold text-base
                  bg-gradient-to-r from-blue-500 to-indigo-600 text-white
                  hover:from-blue-400 hover:to-indigo-500
                  disabled:opacity-50 disabled:cursor-not-allowed
                  transition-all duration-300
                  shadow-lg shadow-blue-500/20 hover:shadow-blue-500/30
                  flex items-center justify-center gap-2 group
                  relative overflow-hidden"
              >
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-700" />
                <span className="relative z-10">
                  {isLoading ? 'Загрузка...' : mode === 'login' ? 'Войти в аккаунт' : 'Создать аккаунт'}
                </span>
                {!isLoading && <ChevronRight className="w-5 h-5 relative z-10 group-hover:translate-x-1 transition-transform" />}
              </button>

              {/* Card number footer (decorative) */}
              <div className="mt-6 pt-4 border-t border-white/[0.06] flex items-center justify-between text-[11px] text-white/25">
                <div className="flex items-center gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-emerald-500/60 animate-pulse" />
                  <span>Безопасное соединение</span>
                </div>
                <span className="font-mono tracking-widest">**** **** **** XPLR</span>
              </div>
            </div>
          </form>
        </div>

        {/* Additional links */}
        <div className="mt-5 text-center">
          {mode === 'register' && (
            <p className="text-xs text-slate-600">
              Регистрируясь, вы соглашаетесь с{' '}
              <a href="#" className="text-blue-500/70 hover:text-blue-400 transition-colors">условиями использования</a>
            </p>
          )}
        </div>

        {/* Features */}
        <div className="mt-6 grid grid-cols-3 gap-3 text-center">
          {[
            { icon: '🔒', label: 'Безопасность' },
            { icon: '💳', label: 'Виртуальные карты' },
            { icon: '⚡', label: 'Мгновенно' },
          ].map((f) => (
            <div key={f.label} className="bg-white/[0.02] backdrop-blur-sm border border-white/[0.05] rounded-xl p-4">
              <div className="text-xl mb-1">{f.icon}</div>
              <p className="text-[11px] text-slate-600">{f.label}</p>
            </div>
          ))}
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
