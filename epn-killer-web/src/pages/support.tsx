import { useState } from 'react';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  MessageCircle,
  Mail,
  Phone,
  Send,
  ExternalLink,
  Clock,
  CheckCircle
} from 'lucide-react';

export const SupportPage = () => {
  const [message, setMessage] = useState('');
  const [sent, setSent] = useState(false);

  const handleSend = () => {
    if (message.trim()) {
      setSent(true);
      setMessage('');
      setTimeout(() => setSent(false), 3000);
    }
  };

  return (
    <DashboardLayout>
      <div className="stagger-fade-in max-w-4xl">
        <BackButton />
        
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Поддержка</h1>
          <p className="text-slate-500">Мы на связи 24/7. Выберите удобный способ обращения.</p>
        </div>

        {/* Contact Options */}
        <div className="grid md:grid-cols-3 gap-4 mb-8">
          <a 
            href="https://t.me/xplr_support" 
            target="_blank" 
            rel="noopener noreferrer"
            className="glass-card p-6 card-hover group"
          >
            <div className="w-14 h-14 rounded-xl bg-blue-500/20 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
              <MessageCircle className="w-7 h-7 text-blue-400" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-1">Telegram</h3>
            <p className="text-slate-500 text-sm mb-3">Быстрый ответ в чате</p>
            <div className="flex items-center gap-2 text-blue-500 text-sm font-medium">
              @xplr_support
              <ExternalLink className="w-4 h-4" />
            </div>
          </a>

          <a 
            href="mailto:support@xplr.io"
            className="glass-card p-6 card-hover group"
          >
            <div className="w-14 h-14 rounded-xl bg-emerald-500/20 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
              <Mail className="w-7 h-7 text-emerald-400" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-1">Email</h3>
            <p className="text-slate-500 text-sm mb-3">Ответ в течение 2 часов</p>
            <div className="flex items-center gap-2 text-emerald-500 text-sm font-medium">
              support@xplr.io
            </div>
          </a>

          <a 
            href="https://wa.me/79991234567"
            target="_blank" 
            rel="noopener noreferrer"
            className="glass-card p-6 card-hover group"
          >
            <div className="w-14 h-14 rounded-xl bg-green-500/20 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
              <Phone className="w-7 h-7 text-green-400" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-1">WhatsApp</h3>
            <p className="text-slate-500 text-sm mb-3">Для срочных вопросов</p>
            <div className="flex items-center gap-2 text-green-500 text-sm font-medium">
              +7 (999) 123-45-67
              <ExternalLink className="w-4 h-4" />
            </div>
          </a>
        </div>

        {/* Quick Message Form */}
        <div className="glass-card p-6 mb-8">
          <h3 className="text-lg font-semibold text-white mb-4">Быстрое сообщение</h3>
          
          {sent ? (
            <div className="flex items-center gap-3 p-6 bg-emerald-500/10 rounded-xl border border-emerald-500/30">
              <CheckCircle className="w-8 h-8 text-emerald-400" />
              <div>
                <p className="font-semibold text-emerald-400">Сообщение отправлено!</p>
                <p className="text-sm text-emerald-500/70">Мы ответим в ближайшее время</p>
              </div>
            </div>
          ) : (
            <>
              <textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                placeholder="Опишите вашу проблему или вопрос..."
                rows={4}
                className="xplr-input w-full mb-4 resize-none"
              />
              <div className="flex items-center justify-between">
                <p className="text-sm text-slate-400">
                  Среднее время ответа: <span className="text-white font-medium">15 минут</span>
                </p>
                <button 
                  onClick={handleSend}
                  disabled={!message.trim()}
                  className="flex items-center gap-2 px-6 py-3 gradient-accent text-white font-medium rounded-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed hover:shadow-lg hover:shadow-blue-500/25"
                >
                  <Send className="w-5 h-5" />
                  Отправить
                </button>
              </div>
            </>
          )}
        </div>

        {/* Working Hours */}
        <div className="glass-card p-6">
          <div className="flex items-center gap-3 mb-4">
            <Clock className="w-5 h-5 text-blue-500" />
            <h3 className="text-lg font-semibold text-white">Время работы</h3>
          </div>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="p-4 bg-white/[0.03] rounded-xl border border-white/10">
              <p className="font-medium text-white mb-1">Telegram & WhatsApp</p>
              <p className="text-emerald-400 font-semibold">24/7</p>
              <p className="text-sm text-slate-500 mt-1">Круглосуточно без выходных</p>
            </div>
            <div className="p-4 bg-white/[0.03] rounded-xl border border-white/10">
              <p className="font-medium text-white mb-1">Email поддержка</p>
              <p className="text-blue-400 font-semibold">09:00 — 21:00 МСК</p>
              <p className="text-sm text-slate-500 mt-1">Ответ в течение 2 часов</p>
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
};
