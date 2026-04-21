import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Mail, CheckCircle } from 'lucide-react';
import { requestPasswordReset } from '../services/auth';

export const ForgotPasswordPage = () => {
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState('');
  const [cooldown, setCooldown] = useState(0);

  const startCooldown = () => {
    setCooldown(60);
    const timer = setInterval(() => {
      setCooldown((prev) => {
        if (prev <= 1) { clearInterval(timer); return 0; }
        return prev - 1;
      });
    }, 1000);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!email.trim()) {
      setError('Введите email');
      return;
    }
    if (cooldown > 0) return;
    setIsLoading(true);
    try {
      await requestPasswordReset(email.trim());
      setSent(true);
      startCooldown();
    } catch (err: any) {
      if (err.code === 'ERR_NETWORK') {
        setError('Сервер недоступен. Проверьте подключение.');
      } else if (err.response?.status === 429) {
        setError('Слишком много попыток. Подождите несколько минут.');
        startCooldown();
      } else {
        // Always show success to not leak email existence
        setSent(true);
        startCooldown();
      }
    } finally {
      setIsLoading(false);
    }
  };

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
            {sent ? (
              <div className="text-center py-4">
                <CheckCircle className="w-16 h-16 text-emerald-400 mx-auto mb-4" />
                <h2 className="text-xl font-bold text-white mb-2">Письмо отправлено</h2>
                <p className="text-slate-400 text-sm mb-6">
                  Если аккаунт с email <span className="text-white font-medium">{email}</span> существует, мы отправили ссылку для сброса пароля.
                </p>
                <p className="text-slate-500 text-xs mb-6">Проверьте папку «Спам», если письмо не пришло.</p>
                <Link to="/auth">
                  <button className="w-full py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white font-semibold rounded-xl hover:from-blue-400 hover:to-indigo-500 transition-all shadow-lg shadow-blue-500/20">
                    Вернуться к входу
                  </button>
                </Link>
              </div>
            ) : (
              <>
                <div className="text-center mb-6">
                  <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-blue-500/20 to-indigo-500/20 border border-blue-500/30 flex items-center justify-center mx-auto mb-4">
                    <Mail className="w-7 h-7 text-blue-400" />
                  </div>
                  <h2 className="text-xl font-bold text-white mb-1">Сброс пароля</h2>
                  <p className="text-slate-400 text-sm">Введите email, указанный при регистрации</p>
                </div>

                {error && (
                  <div className="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm text-center">
                    {error}
                  </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                  <div className="relative group">
                    <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-400 transition-colors z-10" />
                    <input
                      type="email"
                      placeholder="Email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="w-full bg-white/[0.04] border border-white/[0.10] rounded-xl py-4 pl-12 pr-4
                        text-white placeholder-slate-600
                        focus:outline-none focus:border-blue-400 focus:bg-white/[0.06] focus:ring-2 focus:ring-blue-500/60
                        transition-colors duration-150 text-base font-mono tracking-wide"
                      required
                      autoComplete="email"
                      autoFocus
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={isLoading}
                    className="w-full py-4 rounded-xl font-semibold text-base
                      bg-gradient-to-r from-blue-500 to-indigo-600 text-white
                      hover:from-blue-400 hover:to-indigo-500
                      disabled:opacity-50 disabled:cursor-not-allowed
                      transition-all duration-150
                      shadow-lg shadow-blue-500/20 hover:shadow-blue-500/30"
                  >
                    {isLoading ? 'Отправка...' : 'Отправить ссылку для сброса'}
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
