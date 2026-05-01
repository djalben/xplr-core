import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';
import { API_BASE_URL } from '../services/axios';

type Status = 'loading' | 'success' | 'error';

export const VerifyEmailPage = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [status, setStatus] = useState<Status>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const token = searchParams.get('token');
    if (!token) {
      setStatus('error');
      setMessage('Отсутствует токен подтверждения.');
      return;
    }

    const verify = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/auth/verify-email?token=${encodeURIComponent(token)}`);
        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || 'Verification failed');
        }
        setStatus('success');
        setMessage('Email успешно подтверждён! Перенаправляем...');
        setTimeout(() => navigate('/auth?verified=1', { replace: true }), 2500);
      } catch (err: any) {
        setStatus('error');
        const msg = err.message || 'Ошибка подтверждения';
        if (msg.includes('already used')) {
          // Token might have been pre-opened by email client scanners. Treat as success UX-wise.
          setStatus('success');
          setMessage('Email уже подтверждён. Перенаправляем...');
          setTimeout(() => navigate('/auth?verified=1', { replace: true }), 1000);
        } else if (msg.includes('expired')) {
          setMessage('Ссылка истекла. Запросите новое письмо подтверждения.');
        } else {
          setMessage('Не удалось подтвердить email. Попробуйте ещё раз.');
        }
      }
    };

    verify();
  }, [searchParams, navigate]);

  return (
    <div className="min-h-screen bg-transparent flex items-center justify-center p-4 relative z-2">
      <div className="absolute inset-0 bg-[#05050a]/70" />
      <div className="relative z-10 w-full max-w-sm">
        <div className="rounded-2xl bg-[#0d1528]/90 border border-white/[0.08] p-8 text-center shadow-2xl">
          {status === 'loading' && (
            <>
              <Loader2 className="w-12 h-12 text-blue-400 animate-spin mx-auto mb-4" />
              <h2 className="text-lg font-bold text-white mb-2">Подтверждение email...</h2>
              <p className="text-sm text-slate-400">Пожалуйста, подождите</p>
            </>
          )}
          {status === 'success' && (
            <>
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center">
                <CheckCircle className="w-8 h-8 text-emerald-400" />
              </div>
              <h2 className="text-lg font-bold text-white mb-2">Email подтверждён!</h2>
              <p className="text-sm text-slate-400">{message}</p>
            </>
          )}
          {status === 'error' && (
            <>
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center justify-center">
                <XCircle className="w-8 h-8 text-red-400" />
              </div>
              <h2 className="text-lg font-bold text-white mb-2">Ошибка</h2>
              <p className="text-sm text-slate-400 mb-4">{message}</p>
              <button
                onClick={() => navigate('/auth', { replace: true })}
                className="px-6 py-2.5 rounded-xl bg-blue-500/20 border border-blue-500/20 text-blue-400 text-sm font-medium hover:bg-blue-500/30 transition-colors"
              >
                Перейти к входу
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
