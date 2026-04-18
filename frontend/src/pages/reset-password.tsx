import React, { useState, useMemo } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Lock, Eye, EyeOff, CheckCircle, Check, X, AlertTriangle } from 'lucide-react';
import { resetPassword } from '../services/auth';

const passwordRules = [
  { key: 'length', label: 'Минимум 8 символов', test: (p: string) => p.length >= 8 },
  { key: 'upper', label: 'Заглавная буква', test: (p: string) => /[A-Z]/.test(p) },
  { key: 'digit', label: 'Цифра', test: (p: string) => /\d/.test(p) },
  { key: 'special', label: 'Спецсимвол (!@#$%^&*)', test: (p: string) => /[!@#$%^&*]/.test(p) },
] as const;

export const ResetPasswordPage = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token') || '';

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const pwChecks = useMemo(
    () => passwordRules.map((r) => ({ ...r, pass: r.test(password) })),
    [password],
  );
  const allPwValid = pwChecks.every((c) => c.pass);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!token) {
      setError('Недействительная ссылка для сброса пароля.');
      return;
    }
    if (!allPwValid) {
      setError('Пароль не соответствует требованиям.');
      return;
    }
    if (password !== confirmPassword) {
      setError('Пароли не совпадают.');
      return;
    }

    setIsLoading(true);
    try {
      await resetPassword(token, password);
      setSuccess(true);
    } catch (err: any) {
      const data = err.response?.data;
      if (typeof data === 'string' && data.includes('expired')) {
        setError('Ссылка истекла. Запросите сброс пароля повторно.');
      } else if (typeof data === 'string' && data.includes('Invalid')) {
        setError('Недействительная или уже использованная ссылка.');
      } else if (err.code === 'ERR_NETWORK') {
        setError('Сервер недоступен.');
      } else {
        setError(typeof data === 'string' ? data : 'Ошибка при сбросе пароля.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="min-h-screen bg-transparent flex items-center justify-center p-4 overflow-hidden relative z-2">
        <div className="absolute inset-0 bg-[#05050a]/70" />
        <div className="relative z-10 w-full max-w-md text-center">
          <AlertTriangle className="w-16 h-16 text-amber-400 mx-auto mb-4" />
          <h2 className="text-xl font-bold text-white mb-2">Недействительная ссылка</h2>
          <p className="text-slate-400 text-sm mb-6">Ссылка для сброса пароля отсутствует или повреждена.</p>
          <Link to="/forgot-password">
            <button className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white font-semibold rounded-xl">
              Запросить новую ссылку
            </button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-transparent flex items-center justify-center p-4 overflow-hidden relative z-2">
      <div className="absolute inset-0 bg-[#05050a]/70" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(30,58,138,0.15)_0%,transparent_60%)]" />

      <div className="relative z-10 w-full max-w-md animate-fade-in-up">
        <Link to="/auth">
          <button className="flex items-center gap-2 text-slate-500 hover:text-slate-300 transition-colors mb-6 text-sm font-medium">
            <ArrowLeft className="w-4 h-4" />
            Назад к входу
          </button>
        </Link>

        <div className="relative rounded-[22px] p-8 overflow-hidden
          bg-gradient-to-br from-[#0d1528] via-[#0a1025] to-[#0c0f20]
          shadow-[0_8px_60px_-12px_rgba(30,64,175,0.25)]
          border border-white/[0.08]">
          <div className="absolute inset-0 bg-gradient-to-br from-white/[0.06] via-transparent to-transparent pointer-events-none rounded-[22px]" />
          <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-blue-400/30 to-transparent" />

          <div className="relative z-10">
            {success ? (
              <div className="text-center py-4">
                <CheckCircle className="w-16 h-16 text-emerald-400 mx-auto mb-4" />
                <h2 className="text-xl font-bold text-white mb-2">Пароль изменён</h2>
                <p className="text-slate-400 text-sm mb-6">Теперь вы можете войти с новым паролем.</p>
                <button
                  onClick={() => navigate('/auth')}
                  className="w-full py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white font-semibold rounded-xl hover:from-blue-400 hover:to-indigo-500 transition-all shadow-lg shadow-blue-500/20"
                >
                  Войти
                </button>
              </div>
            ) : (
              <>
                <div className="text-center mb-6">
                  <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-blue-500/20 to-indigo-500/20 border border-blue-500/30 flex items-center justify-center mx-auto mb-4">
                    <Lock className="w-7 h-7 text-blue-400" />
                  </div>
                  <h2 className="text-xl font-bold text-white mb-1">Новый пароль</h2>
                  <p className="text-slate-400 text-sm">Придумайте надёжный пароль</p>
                </div>

                {error && (
                  <div className="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm text-center">
                    {error}
                  </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                  <div className="relative group">
                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-400 transition-colors z-10" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      placeholder="Новый пароль"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="w-full bg-white/[0.04] border border-white/[0.10] rounded-xl py-4 pl-12 pr-14
                        text-white placeholder-slate-600
                        focus:outline-none focus:border-blue-400 focus:bg-white/[0.06] focus:ring-2 focus:ring-blue-500/60
                        transition-colors duration-150 text-base font-mono tracking-widest"
                      required
                      autoComplete="new-password"
                      autoFocus
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword((v) => !v)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white transition-colors p-2 rounded-lg hover:bg-white/[0.06] z-10"
                      tabIndex={-1}
                    >
                      {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                    </button>
                  </div>

                  {password.length > 0 && (
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

                  <div className="relative group">
                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-400 transition-colors z-10" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      placeholder="Подтвердите пароль"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="w-full bg-white/[0.04] border border-white/[0.10] rounded-xl py-4 pl-12 pr-4
                        text-white placeholder-slate-600
                        focus:outline-none focus:border-blue-400 focus:bg-white/[0.06] focus:ring-2 focus:ring-blue-500/60
                        transition-colors duration-150 text-base font-mono tracking-widest"
                      required
                      autoComplete="new-password"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={isLoading || !allPwValid || password !== confirmPassword}
                    className="w-full py-4 rounded-xl font-semibold text-base
                      bg-gradient-to-r from-blue-500 to-indigo-600 text-white
                      hover:from-blue-400 hover:to-indigo-500
                      disabled:opacity-50 disabled:cursor-not-allowed
                      transition-all duration-150
                      shadow-lg shadow-blue-500/20 hover:shadow-blue-500/30"
                  >
                    {isLoading ? 'Сохранение...' : 'Установить новый пароль'}
                  </button>
                </form>
              </>
            )}
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
