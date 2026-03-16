import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import apiClient from '../api/axios';
import { 
  MessageCircle,
  Mail,
  Send,
  ExternalLink,
  Clock,
  CheckCircle,
  AlertCircle,
  Loader2,
  ShieldAlert,
  Settings
} from 'lucide-react';

export const SupportPage = () => {
  const navigate = useNavigate();
  const [message, setMessage] = useState('');
  const [sending, setSending] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null);
  const [telegramLinked, setTelegramLinked] = useState<boolean | null>(null);

  useEffect(() => {
    apiClient.get('/user/settings/profile')
      .then(res => setTelegramLinked(!!res.data?.telegram_linked))
      .catch(() => setTelegramLinked(false));
  }, []);

  const formBlocked = telegramLinked === false;

  const showToast = (msg: string, type: 'ok' | 'err' = 'ok') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 5000);
  };

  const handleSend = async () => {
    if (!message.trim() || sending) return;
    setSending(true);
    try {
      await apiClient.post('/user/support', { message: message.trim() });
      setMessage('');
      showToast('Сообщение отправлено! Мы ответим вам в течение 24 часов.', 'ok');
    } catch (err: any) {
      const msg = err.response?.data || 'Не удалось отправить сообщение. Попробуйте позже.';
      showToast(typeof msg === 'string' ? msg : 'Ошибка отправки', 'err');
    } finally {
      setSending(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="stagger-fade-in max-w-4xl mx-auto">
        <BackButton />
        
        {/* Toast */}
        {toast && (
          <div className={`fixed top-6 right-6 z-50 flex items-center gap-3 px-5 py-4 rounded-2xl text-sm font-medium shadow-2xl backdrop-blur-lg border animate-slide-in ${
            toast.type === 'ok'
              ? 'bg-emerald-500/90 text-white border-emerald-400/30'
              : 'bg-red-500/90 text-white border-red-400/30'
          }`}>
            {toast.type === 'ok' ? <CheckCircle className="w-5 h-5 shrink-0" /> : <AlertCircle className="w-5 h-5 shrink-0" />}
            {toast.msg}
          </div>
        )}

        {/* Header */}
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-white mb-2">Поддержка</h1>
          <p className="text-slate-500">Мы на связи 24/7. Выберите удобный способ обращения.</p>
        </div>

        {/* Contact Options — only Telegram + Email, centered */}
        <div className="grid md:grid-cols-2 gap-4 mb-8 max-w-2xl mx-auto">
          <a 
            href="https://t.me/your_telegram_link" 
            target="_blank" 
            rel="noopener noreferrer"
            className="glass-card p-6 card-hover group text-center"
          >
            <div className="w-14 h-14 rounded-xl bg-blue-500/20 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform mx-auto">
              <MessageCircle className="w-7 h-7 text-blue-400" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-1">Telegram</h3>
            <p className="text-slate-500 text-sm mb-3">Быстрый ответ в чате</p>
            <div className="flex items-center justify-center gap-2 text-blue-500 text-sm font-medium">
              Написать в Telegram
              <ExternalLink className="w-4 h-4" />
            </div>
          </a>

          <a 
            href="mailto:admin@xplr.pro"
            className="glass-card p-6 card-hover group text-center"
          >
            <div className="w-14 h-14 rounded-xl bg-emerald-500/20 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform mx-auto">
              <Mail className="w-7 h-7 text-emerald-400" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-1">Email</h3>
            <p className="text-slate-500 text-sm mb-3">Ответ в течение 24 часов</p>
            <div className="flex items-center justify-center gap-2 text-emerald-500 text-sm font-medium">
              admin@xplr.pro
            </div>
          </a>
        </div>

        {/* Quick Message Form */}
        <div className="glass-card p-6 mb-8 max-w-2xl mx-auto relative overflow-hidden">
          <h3 className="text-lg font-semibold text-white mb-4">Быстрое сообщение</h3>
          <textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder="Опишите вашу проблему или вопрос..."
            rows={4}
            className="xplr-input w-full mb-4 resize-none"
            disabled={sending || formBlocked}
          />
          <div className="flex items-center justify-between">
            <p className="text-sm text-slate-400">
              Среднее время ответа: <span className="text-white font-medium">до 24 часов</span>
            </p>
            <button 
              onClick={handleSend}
              disabled={!message.trim() || sending || formBlocked}
              className="flex items-center gap-2 px-6 py-3 gradient-accent text-white font-medium rounded-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed hover:shadow-lg hover:shadow-blue-500/25"
            >
              {sending ? <Loader2 className="w-5 h-5 animate-spin" /> : <Send className="w-5 h-5" />}
              {sending ? 'Отправка...' : 'Отправить'}
            </button>
          </div>

          {/* Overlay: Telegram not linked */}
          {formBlocked && (
            <div className="absolute inset-0 z-10 flex flex-col items-center justify-center gap-4 bg-slate-950/80 backdrop-blur-sm rounded-2xl px-6 text-center">
              <div className="w-14 h-14 rounded-full bg-amber-500/20 flex items-center justify-center">
                <ShieldAlert className="w-7 h-7 text-amber-400" />
              </div>
              <p className="text-white font-semibold text-lg leading-snug max-w-sm">
                Для отправки сообщений в поддержку необходимо подключить Telegram
              </p>
              <p className="text-slate-400 text-sm max-w-xs">
                Привяжите Telegram-бота в настройках профиля, чтобы получать ответы и отправлять обращения.
              </p>
              <button
                onClick={() => navigate('/settings')}
                className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-500 to-violet-500 hover:from-blue-600 hover:to-violet-600 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40"
              >
                <Settings className="w-5 h-5" />
                Перейти к привязке
              </button>
            </div>
          )}
        </div>

        {/* Working Hours */}
        <div className="glass-card p-6 max-w-2xl mx-auto">
          <div className="flex items-center gap-3 mb-4">
            <Clock className="w-5 h-5 text-blue-500" />
            <h3 className="text-lg font-semibold text-white">Время работы</h3>
          </div>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="p-4 bg-white/[0.03] rounded-xl border border-white/10">
              <p className="font-medium text-white mb-1">Telegram</p>
              <p className="text-emerald-400 font-semibold">24/7</p>
              <p className="text-sm text-slate-500 mt-1">Круглосуточно без выходных</p>
            </div>
            <div className="p-4 bg-white/[0.03] rounded-xl border border-white/10">
              <p className="font-medium text-white mb-1">Email поддержка</p>
              <p className="text-blue-400 font-semibold">09:00 — 21:00 МСК</p>
              <p className="text-sm text-slate-500 mt-1">Ответ в течение 24 часов</p>
            </div>
          </div>
        </div>
      </div>

      <style>{`
        @keyframes slide-in {
          from { opacity: 0; transform: translateX(40px); }
          to { opacity: 1; transform: translateX(0); }
        }
        .animate-slide-in {
          animation: slide-in 0.3s ease-out forwards;
        }
      `}</style>
    </DashboardLayout>
  );
};
