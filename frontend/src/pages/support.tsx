import { useState, useEffect, useRef, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import apiClient from '../api/axios';
import {
  MessageCircle,
  Send,
  CheckCircle,
  AlertCircle,
  Loader2,
  CreditCard,
  Wallet,
  ShieldCheck,
  HelpCircle,
  X,
  Headphones,
  Lock,
} from 'lucide-react';

interface ChatMessage {
  id: number;
  conversation_id: number;
  sender_type: 'user' | 'admin';
  sender_name: string;
  body: string;
  created_at: string;
}

interface Conversation {
  id: number;
  user_id: number;
  topic: string;
  status: string;
  created_at: string;
  updated_at: string;
}

const POLL_INTERVAL = 5000;

export const SupportPage = () => {
  const { t } = useTranslation();
  const [conversation, setConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const showToast = (msg: string, type: 'ok' | 'err' = 'ok') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 4000);
  };

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  // Try to load existing open conversation on mount
  useEffect(() => {
    apiClient.post('/user/chat/start', { topic: '__check__' })
      .then(res => {
        if (res.data?.conversation?.status === 'open' && res.data.conversation.topic !== '__check__') {
          setConversation(res.data.conversation);
          setMessages(res.data.messages || []);
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  // Poll for new messages when conversation is active (also poll 'closed' briefly to get final system msg)
  useEffect(() => {
    if (!conversation) return;

    const poll = () => {
      apiClient.get(`/user/chat/messages/${conversation.id}`)
        .then(res => {
          const newMsgs: ChatMessage[] = res.data?.messages || [];
          setMessages(prev => {
            if (newMsgs.length !== prev.length) return newMsgs;
            return prev;
          });
          // Update conversation status
          if (res.data?.conversation) setConversation(res.data.conversation);
        })
        .catch(() => {});
    };

    pollRef.current = setInterval(poll, POLL_INTERVAL);
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, [conversation]);

  // Auto-scroll on new messages
  useEffect(() => { scrollToBottom(); }, [messages, scrollToBottom]);

  const startConversation = async (topic: string) => {
    setLoading(true);
    try {
      const res = await apiClient.post('/user/chat/start', { topic });
      setConversation(res.data.conversation);
      setMessages(res.data.messages || []);
    } catch {
      showToast(t('support.error'), 'err');
    } finally {
      setLoading(false);
    }
  };

  const handleSend = async () => {
    if (!input.trim() || sending || !conversation) return;
    const text = input.trim();
    setSending(true);
    setInput('');
    try {
      const res = await apiClient.post(`/user/chat/send/${conversation.id}`, { message: text });
      setMessages(prev => [...prev, res.data]);
    } catch {
      showToast(t('support.sendError'), 'err');
      setInput(text);
    } finally {
      setSending(false);
    }
  };

  const handleClose = async () => {
    if (!conversation) return;
    try {
      await apiClient.post(`/user/chat/close/${conversation.id}`);
      setConversation(null);
      setMessages([]);
      showToast(t('support.chatClosed'), 'ok');
    } catch {
      showToast(t('support.error'), 'err');
    }
  };

  const topics = [
    { key: 'cards', icon: CreditCard, color: 'text-blue-400 bg-blue-500/20' },
    { key: 'payments', icon: Wallet, color: 'text-emerald-400 bg-emerald-500/20' },
    { key: 'security', icon: ShieldCheck, color: 'text-amber-400 bg-amber-500/20' },
    { key: 'other', icon: HelpCircle, color: 'text-purple-400 bg-purple-500/20' },
  ];

  const fmtTime = (iso: string) => {
    const d = new Date(iso);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex justify-center items-center py-24">
          <Loader2 className="w-8 h-8 animate-spin text-blue-400" />
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="stagger-fade-in w-full max-w-3xl mx-auto overflow-hidden">
        <BackButton />

        {/* Toast */}
        {toast && (
          <div className={`fixed top-6 right-6 z-50 flex items-center gap-3 px-5 py-4 rounded-2xl text-sm font-medium shadow-2xl backdrop-blur-lg border animate-slide-in ${
            toast.type === 'ok' ? 'bg-emerald-500/90 text-white border-emerald-400/30' : 'bg-red-500/90 text-white border-red-400/30'
          }`}>
            {toast.type === 'ok' ? <CheckCircle className="w-5 h-5 shrink-0" /> : <AlertCircle className="w-5 h-5 shrink-0" />}
            {toast.msg}
          </div>
        )}

        {/* ─── No active conversation → Topic picker ─── */}
        {!conversation && (
          <>
            <div className="mb-8 text-center">
              <div className="w-16 h-16 rounded-2xl bg-blue-500/20 flex items-center justify-center mx-auto mb-4">
                <Headphones className="w-8 h-8 text-blue-400" />
              </div>
              <h1 className="text-3xl font-bold text-white mb-2">{t('support.title')}</h1>
              <p className="text-slate-400">{t('support.subtitle')}</p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 max-w-xl mx-auto">
              {topics.map(tp => (
                <button
                  key={tp.key}
                  onClick={() => startConversation(tp.key)}
                  className="glass-card p-5 card-hover group text-left"
                >
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-3 ${tp.color} group-hover:scale-110 transition-transform`}>
                    <tp.icon className="w-6 h-6" />
                  </div>
                  <h3 className="text-white font-semibold mb-1">{t(`support.topics.${tp.key}`)}</h3>
                  <p className="text-slate-500 text-sm">{t(`support.topics.${tp.key}Desc`)}</p>
                </button>
              ))}
            </div>

            <div className="mt-8 text-center">
              <p className="text-slate-500 text-sm">
                {t('support.avgResponse')} <span className="text-white font-medium">{t('support.avgResponseTime')}</span>
              </p>
            </div>
          </>
        )}

        {/* ─── Active conversation → Chat ─── */}
        {conversation && (
          <div className="glass-card overflow-hidden flex flex-col" style={{ height: 'calc(100dvh - 180px)', minHeight: 400 }}>
            {/* Chat header */}
            <div className="flex items-center justify-between p-4 border-b border-white/10">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-blue-500/20 flex items-center justify-center">
                  <MessageCircle className="w-5 h-5 text-blue-400" />
                </div>
                <div>
                  <h3 className="text-white font-semibold text-sm">{t('support.chatTitle')}</h3>
                  <p className="text-slate-500 text-xs">{t(`support.topics.${conversation.topic}`)}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {conversation.status === 'open' ? (
                  <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-500/20 text-emerald-400 text-xs font-medium">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                    {t('support.online')}
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-slate-500/20 text-slate-400 text-xs font-medium">
                    <Lock className="w-3 h-3" />
                    Решён
                  </span>
                )}
                {conversation.status === 'open' && (
                  <button onClick={handleClose} className="p-2 hover:bg-white/10 rounded-lg transition-colors" title={t('support.closeChat')}>
                    <X className="w-4 h-4 text-slate-400" />
                  </button>
                )}
              </div>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {/* System welcome */}
              <div className="flex justify-center">
                <span className="px-3 py-1.5 rounded-full bg-white/[0.05] text-slate-500 text-xs">
                  {t('support.welcomeMsg')}
                </span>
              </div>

              {messages.map(msg => (
                <div key={msg.id} className={`flex ${msg.sender_type === 'user' ? 'justify-end' : 'justify-start'}`}>
                  <div className={`max-w-[80%] rounded-2xl px-4 py-2.5 ${
                    msg.sender_type === 'user'
                      ? 'bg-blue-500 text-white rounded-br-md'
                      : 'bg-white/[0.06] text-white border border-white/10 rounded-bl-md'
                  }`}>
                    {msg.sender_type === 'admin' && (
                      <p className="text-xs text-blue-400 font-medium mb-1">{msg.sender_name}</p>
                    )}
                    <p className="text-sm whitespace-pre-wrap break-words">{msg.body}</p>
                    <p className={`text-[10px] mt-1 ${msg.sender_type === 'user' ? 'text-blue-200' : 'text-slate-500'}`}>
                      {fmtTime(msg.created_at)}
                    </p>
                  </div>
                </div>
              ))}
              <div ref={messagesEndRef} />
            </div>

            {/* Input or Closed banner */}
            {conversation.status === 'open' ? (
              <div className="p-3 border-t border-white/10">
                <div className="flex items-end gap-2">
                  <textarea
                    value={input}
                    onChange={e => setInput(e.target.value)}
                    onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); } }}
                    placeholder={t('support.placeholder')}
                    rows={1}
                    className="xplr-input flex-1 resize-none min-h-[44px] max-h-32 py-2.5"
                    disabled={sending}
                  />
                  <button
                    onClick={handleSend}
                    disabled={!input.trim() || sending}
                    className="shrink-0 w-11 h-11 flex items-center justify-center rounded-xl bg-blue-500 hover:bg-blue-600 disabled:opacity-40 disabled:cursor-not-allowed text-white transition-colors"
                  >
                    {sending ? <Loader2 className="w-5 h-5 animate-spin" /> : <Send className="w-5 h-5" />}
                  </button>
                </div>
              </div>
            ) : (
              <div className="p-4 border-t border-white/10 text-center">
                <div className="flex items-center justify-center gap-2 text-slate-400 mb-3">
                  <CheckCircle className="w-5 h-5 text-emerald-400" />
                  <span className="text-sm font-medium">Вопрос решён</span>
                </div>
                <button
                  onClick={() => { setConversation(null); setMessages([]); }}
                  className="px-6 py-2.5 rounded-xl bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium transition-colors"
                >
                  Создать новый запрос
                </button>
              </div>
            )}
          </div>
        )}
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
