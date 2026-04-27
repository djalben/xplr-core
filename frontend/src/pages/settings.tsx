import { useState, useEffect, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { LANG_KEY } from '../i18n';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import apiClient from '../services/axios';
import { useAuth } from '../store/auth-context';
import { QRCodeSVG } from 'qrcode.react';
import { 
  User, Lock, Bell, Shield, Eye, EyeOff, Save, Check, Smartphone, Globe, Mail,
  FileText, Upload, CheckCircle, AlertCircle, MessageCircle, LogOut, Loader2,
  MonitorSmartphone, ShieldCheck, ShieldAlert, ExternalLink, Clock, X, Pencil,
  Copy, PartyPopper
} from 'lucide-react';

type SettingsTab = 'profile' | 'security' | 'notifications' | 'kyc' | 'language';

interface ProfileData {
  id: number;
  email: string;
  display_name: string;
  is_verified: boolean;
  verification_status: string;
  two_factor_enabled: boolean;
  telegram_linked: boolean;
  role: string;
  is_admin: boolean;
}

interface SessionData {
  id: number;
  ip: string;
  location: string;
  device: string;
  last_active: string;
}

interface NotifPrefs {
  notify_transactions: boolean;
  notify_balance: boolean;
  notify_security: boolean;
}

const Toggle = ({ checked, onChange, disabled = false }: { checked: boolean; onChange: (v: boolean) => void; disabled?: boolean }) => (
  <label className={`relative inline-flex items-center ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}>
    <input type="checkbox" checked={checked} onChange={(e) => !disabled && onChange(e.target.checked)} className="sr-only peer" disabled={disabled} />
    <div className="w-11 h-6 bg-white/10 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-500/50 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-white/20 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
  </label>
);

const Toast = ({ msg, type }: { msg: string; type: 'ok' | 'err' }) => (
  <div className={`fixed top-6 right-6 z-50 flex items-center gap-3 px-5 py-4 rounded-2xl text-sm font-medium shadow-2xl backdrop-blur-lg border ${
    type === 'ok' ? 'bg-emerald-500/90 text-white border-emerald-400/30' : 'bg-red-500/90 text-white border-red-400/30'
  }`}>
    {type === 'ok' ? <CheckCircle className="w-5 h-5 shrink-0" /> : <AlertCircle className="w-5 h-5 shrink-0" />}
    {msg}
  </div>
);

// ══════════════════════════════════════
// TELEGRAM LINK MODAL
// ══════════════════════════════════════
const TelegramLinkModal = ({ onClose, onSuccess }: { onClose: () => void; onSuccess: () => void }) => {
  const [link, setLink] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);
  const [success, setSuccess] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await apiClient.get('/user/settings/telegram-link');
        if (!cancelled) {
          setLink(res.data?.link || '');
          setCode(res.data?.code || '');
          setLoading(false);
        }
      } catch {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  // Polling for link status
  useEffect(() => {
    if (!code || success) return;
    pollRef.current = setInterval(async () => {
      try {
        const res = await apiClient.get('/user/settings/telegram/check-status');
        if (res.data?.linked) {
          if (pollRef.current) clearInterval(pollRef.current);
          setSuccess(true);
          setTimeout(() => { onSuccess(); }, 2000);
        }
      } catch { /* ignore */ }
    }, 2500);
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, [code, success, onSuccess]);

  const copyCode = () => {
    navigator.clipboard.writeText(code).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div className="relative w-full max-w-md bg-slate-900/95 border border-white/10 rounded-2xl shadow-2xl p-6 space-y-5" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-bold text-white flex items-center gap-2">
            <MessageCircle className="w-5 h-5 text-blue-400" />
            Привязка Telegram
          </h2>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors"><X className="w-5 h-5" /></button>
        </div>

        {loading ? (
          <div className="flex justify-center py-12"><Loader2 className="w-8 h-8 animate-spin text-blue-400" /></div>
        ) : success ? (
          <div className="flex flex-col items-center gap-4 py-8">
            <div className="w-16 h-16 rounded-full bg-emerald-500/20 flex items-center justify-center">
              <CheckCircle className="w-10 h-10 text-emerald-400" />
            </div>
            <p className="text-xl font-bold text-white">Telegram привязан!</p>
            <p className="text-sm text-slate-400">Уведомления теперь приходят в оба канала</p>
          </div>
        ) : (
          <>
            {/* QR Code */}
            <div className="flex flex-col items-center gap-3">
              <div className="bg-white p-3 rounded-xl">
                <QRCodeSVG value={link} size={180} level="M" />
              </div>
              <p className="text-xs text-slate-500">Отсканируйте QR-код камерой телефона</p>
            </div>

            {/* Open Telegram button */}
            <a
              href={link}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center justify-center gap-2 w-full py-3 bg-blue-500 hover:bg-blue-600 text-white font-semibold rounded-xl transition-colors text-sm"
            >
              <ExternalLink className="w-4 h-4" />
              Открыть Telegram
            </a>

            {/* Manual code */}
            <div className="bg-white/[0.03] border border-white/5 rounded-xl p-4 space-y-2">
              <p className="text-xs text-slate-400">Или отправьте боту <span className="text-blue-400 font-medium">@xplr_notify_bot</span> команду:</p>
              <div className="flex items-center gap-2">
                <code className="flex-1 bg-slate-800 text-blue-300 px-3 py-2 rounded-lg text-sm font-mono truncate">/start {code}</code>
                <button
                  onClick={copyCode}
                  className="shrink-0 px-3 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg transition-colors"
                  title="Скопировать код"
                >
                  {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                </button>
              </div>
            </div>

            {/* Footer */}
            <p className="text-xs text-slate-500 text-center">
              Если Telegram не установлен — <a href="https://web.telegram.org" target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:underline">используйте Telegram Web</a>
            </p>

            {/* Polling indicator */}
            <div className="flex items-center justify-center gap-2 text-xs text-slate-500">
              <Loader2 className="w-3 h-3 animate-spin" />
              Ожидание привязки...
            </div>
          </>
        )}
      </div>
    </div>
  );
};

// ══════════════════════════════════════
// TELEGRAM CARD (used inside ProfileTab)
// ══════════════════════════════════════
const TelegramCard = ({ profile, reload, showToast }: { profile: ProfileData | null; reload: () => void; showToast: (m: string, t: 'ok' | 'err') => void }) => {
  const { t } = useTranslation();
  const [showModal, setShowModal] = useState(false);

  const handleLinkSuccess = () => {
    setShowModal(false);
    showToast('✅ Telegram успешно привязан!', 'ok');
    reload();
  };

  return (
    <>
      <div className="glass-card p-4 sm:p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><MessageCircle className="w-5 h-5 text-blue-400" />{t('settings.telegram.title')}</h3>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          {profile?.telegram_linked ? (
            <>
              <div className="flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-emerald-400" /><p className="text-white font-medium">{t('settings.telegram.connected')}</p></div>
              <div className="flex items-center gap-2">
                <span className="px-3 py-1.5 bg-emerald-500/20 text-emerald-400 text-xs font-medium rounded-lg">{t('settings.telegram.connected')}</span>
                <button
                  onClick={async () => {
                    try {
                      await apiClient.post('/user/settings/telegram/unlink');
                      showToast('Telegram отвязан', 'ok');
                      reload();
                    } catch { showToast('Ошибка отвязки Telegram', 'err'); }
                  }}
                  className="px-3 py-1.5 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 text-xs font-medium rounded-lg transition-colors"
                >
                  Отвязать
                </button>
              </div>
            </>
          ) : (
            <>
              <p className="text-sm text-slate-400">{t('settings.telegram.linkDesc')}</p>
              <button onClick={() => setShowModal(true)} className="flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium rounded-xl transition-colors whitespace-nowrap w-full sm:w-auto">
                <MessageCircle className="w-3.5 h-3.5" />
                {t('settings.telegram.link')}
              </button>
            </>
          )}
        </div>
      </div>
      {showModal && <TelegramLinkModal onClose={() => setShowModal(false)} onSuccess={handleLinkSuccess} />}
    </>
  );
};

// ══════════════════════════════════════
// PROFILE TAB
// ══════════════════════════════════════
const ProfileTab = ({ profile, reload, showToast, setActiveTab }: { profile: ProfileData | null; reload: () => void; showToast: (m: string, t: 'ok' | 'err') => void; setActiveTab: (tab: SettingsTab) => void }) => {
  const { t } = useTranslation();
  const { updateUserName } = useAuth();
  const [displayName, setDisplayName] = useState(profile?.display_name || '');
  const [isEditingName, setIsEditingName] = useState(false);
  const [saving, setSaving] = useState(false);
  const [emailVerifyOpen, setEmailVerifyOpen] = useState(false);
  const [emailCode, setEmailCode] = useState('');
  const [emailSending, setEmailSending] = useState(false);
  const [emailVerifying, setEmailVerifying] = useState(false);
  useEffect(() => { if (profile) setDisplayName(profile.display_name || ''); }, [profile]);

  const handleSave = async () => {
    if (!displayName.trim()) return;
    setSaving(true);
    try {
      await apiClient.patch('/user/settings/profile', { display_name: displayName.trim() });
      updateUserName(displayName.trim());
      setIsEditingName(false);
      showToast(t('settings.profileSection.profileUpdated'), 'ok');
      reload();
    } catch { showToast(t('settings.profileSection.saveError'), 'err'); }
    finally { setSaving(false); }
  };

  const handleRequestEmailCode = async () => {
    setEmailSending(true);
    try {
      await apiClient.post('/user/settings/verify-email-request');
      setEmailVerifyOpen(true);
      showToast(t('settings.emailVerify.codeSent'), 'ok');
    } catch { showToast(t('settings.emailVerify.sendError'), 'err'); }
    finally { setEmailSending(false); }
  };

  const handleConfirmEmailCode = async () => {
    if (emailCode.length !== 6) return;
    setEmailVerifying(true);
    try {
      await apiClient.post('/user/settings/verify-email-confirm', { code: emailCode });
      showToast(t('settings.emailVerify.verified'), 'ok');
      setEmailVerifyOpen(false);
      setEmailCode('');
      reload();
    } catch { showToast(t('settings.emailVerify.invalidCode'), 'err'); }
    finally { setEmailVerifying(false); }
  };

  const verStatus = profile?.verification_status || 'pending';
  const verColors: Record<string, string> = { pending: 'bg-orange-500/20 text-orange-400', verified: 'bg-emerald-500/20 text-emerald-400', rejected: 'bg-red-500/20 text-red-400' };
  const verLabel = t(`settings.verification.${verStatus}` as any) || t('settings.verification.pending');

  return (
    <div className="space-y-6">
      <div className="glass-card p-4 sm:p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><User className="w-5 h-5 text-blue-400" />{t('settings.profileSection.title')}</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-2">{t('settings.profileSection.email')}</label>
            <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
              <input type="text" value={profile?.email || ''} readOnly className="xplr-input w-full min-w-0 bg-white/[0.02] cursor-not-allowed" />
              {profile?.is_verified
                ? <span className="flex items-center gap-1 text-xs text-emerald-400 whitespace-nowrap shrink-0"><CheckCircle className="w-3.5 h-3.5" />{t('settings.profileSection.verified')}</span>
                : <button onClick={handleRequestEmailCode} disabled={emailSending} className="flex items-center justify-center gap-1.5 px-3 py-2 bg-blue-500/20 hover:bg-blue-500/30 border border-blue-500/30 text-blue-400 text-xs font-medium rounded-lg transition-colors whitespace-nowrap w-full sm:w-auto shrink-0">
                    {emailSending ? <Loader2 className="w-3 h-3 animate-spin" /> : <Mail className="w-3 h-3" />}{t('settings.emailVerify.verify')}
                  </button>}
            </div>
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-2">{t('settings.profileSection.displayName')}</label>
            <div className="flex items-center gap-2 w-full min-w-0">
              <input
                type="text"
                value={displayName}
                onChange={e => setDisplayName(e.target.value)}
                placeholder={t('settings.profileSection.displayNamePlaceholder')}
                readOnly={!isEditingName}
                onKeyDown={e => { if (e.key === 'Enter' && isEditingName) handleSave(); if (e.key === 'Escape') { setDisplayName(profile?.display_name || ''); setIsEditingName(false); } }}
                className={`xplr-input w-full min-w-0 transition-colors ${
                  isEditingName ? 'border-blue-500/50 bg-white/[0.04]' : 'bg-white/[0.02] cursor-default'
                }`}
              />
              {!isEditingName ? (
                <button
                  onClick={() => setIsEditingName(true)}
                  className="shrink-0 p-2.5 hover:bg-white/10 rounded-xl transition-colors group"
                  title={t('settings.profileSection.edit')}
                >
                  <Pencil className="w-4 h-4 text-slate-400 group-hover:text-blue-400 transition-colors" />
                </button>
              ) : (
                <button
                  onClick={handleSave}
                  disabled={saving || !displayName.trim()}
                  className="shrink-0 flex items-center gap-1.5 px-4 py-2.5 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white text-sm font-medium rounded-xl transition-colors whitespace-nowrap"
                >
                  {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                  {t('settings.profileSection.saveProfile')}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="glass-card p-4 sm:p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><FileText className="w-5 h-5 text-purple-400" />{t('settings.verification.title')}</h3>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <span className={`px-3 py-1 rounded-full text-xs font-medium shrink-0 ${verColors[verStatus] || verColors.pending}`}>{verLabel}</span>
            <p className="text-sm text-slate-400 truncate">{verStatus === 'verified' ? t('settings.verification.accountVerified') : t('settings.verification.needVerification')}</p>
          </div>
          {verStatus !== 'verified' && <button onClick={() => setActiveTab('kyc')} className="w-full sm:w-auto px-4 py-2.5 bg-purple-500 hover:bg-purple-600 text-white text-sm font-medium rounded-xl transition-colors whitespace-nowrap text-center">{t('settings.verification.startVerification')}</button>}
        </div>
      </div>

      <TelegramCard profile={profile} reload={reload} showToast={showToast} />

      {/* Email Verification Modal */}
      {emailVerifyOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="glass-card p-8 max-w-sm w-full mx-4 relative">
            <button onClick={() => { setEmailVerifyOpen(false); setEmailCode(''); }} className="absolute top-4 right-4 text-slate-400 hover:text-white"><X className="w-5 h-5" /></button>
            <div className="text-center mb-6">
              <Mail className="w-12 h-12 text-blue-400 mx-auto mb-3" />
              <h3 className="text-lg font-bold text-white mb-1">{t('settings.emailVerify.title')}</h3>
              <p className="text-sm text-slate-400">{t('settings.emailVerify.description')}</p>
            </div>
            <input
              type="text"
              value={emailCode}
              onChange={e => setEmailCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
              placeholder="000000"
              className="xplr-input w-full text-center tracking-[0.5em] text-2xl font-bold mb-4"
              maxLength={6}
              autoFocus
            />
            <button
              onClick={handleConfirmEmailCode}
              disabled={emailVerifying || emailCode.length !== 6}
              className="w-full px-4 py-3 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white font-medium rounded-xl transition-colors flex items-center justify-center gap-2"
            >
              {emailVerifying ? <Loader2 className="w-4 h-4 animate-spin" /> : <CheckCircle className="w-4 h-4" />}
              {t('settings.emailVerify.confirm')}
            </button>
            <button onClick={handleRequestEmailCode} disabled={emailSending} className="w-full mt-3 text-sm text-slate-400 hover:text-blue-400 transition-colors">
              {emailSending ? t('settings.emailVerify.resending') : t('settings.emailVerify.resend')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

// ══════════════════════════════════════
// SECURITY TAB
// ══════════════════════════════════════
const SecurityTab = ({ profile, reload, showToast }: { profile: ProfileData | null; reload: () => void; showToast: (m: string, t: 'ok' | 'err') => void }) => {
  const { t } = useTranslation();
  const [oldPw, setOldPw] = useState('');
  const [newPw, setNewPw] = useState('');
  const [confirmPw, setConfirmPw] = useState('');
  const [showOld, setShowOld] = useState(false);
  const [showNew, setShowNew] = useState(false);
  const [pwSaving, setPwSaving] = useState(false);
  const [totpSecret, setTotpSecret] = useState('');
  const [totpURI, setTotpURI] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [totpSaving, setTotpSaving] = useState(false);
  const [sessions, setSessions] = useState<SessionData[]>([]);
  const [logoutSaving, setLogoutSaving] = useState(false);

  const loadSessions = useCallback(async () => {
    try { const res = await apiClient.get('/user/settings/sessions'); setSessions(res.data || []); } catch { /* ignore */ }
  }, []);
  useEffect(() => { loadSessions(); }, [loadSessions]);

  const handleChangePassword = async () => {
    if (newPw !== confirmPw) { showToast(t('settings.password.mismatch'), 'err'); return; }
    if (newPw.length < 8) { showToast(t('settings.password.tooShort'), 'err'); return; }
    setPwSaving(true);
    try {
      await apiClient.post('/user/settings/change-password', { old_password: oldPw, new_password: newPw });
      showToast(t('settings.password.changed'), 'ok');
      setOldPw(''); setNewPw(''); setConfirmPw('');
    } catch (err: any) {
      const msg = typeof err.response?.data === 'string' ? err.response.data : t('settings.password.error');
      showToast(msg, 'err');
    } finally { setPwSaving(false); }
  };

  const handleSetup2FA = async () => {
    try { const res = await apiClient.post('/user/settings/2fa/setup'); setTotpSecret(res.data.secret); setTotpURI(res.data.otp_uri); }
    catch { showToast(t('settings.twoFa.setupError'), 'err'); }
  };

  const handleVerify2FA = async () => {
    if (totpCode.length !== 6) { showToast(t('settings.twoFa.enterCode'), 'err'); return; }
    setTotpSaving(true);
    try {
      await apiClient.post('/user/settings/2fa/verify', { code: totpCode });
      showToast(t('settings.twoFa.enabledToast'), 'ok');
      setTotpSecret(''); setTotpURI(''); setTotpCode(''); reload();
    } catch (err: any) {
      showToast(typeof err.response?.data === 'string' ? err.response.data : t('settings.twoFa.invalidCode'), 'err');
    } finally { setTotpSaving(false); }
  };

  const handleDisable2FA = async () => {
    try { await apiClient.post('/user/settings/2fa/disable'); showToast(t('settings.twoFa.disabledToast'), 'ok'); reload(); }
    catch { showToast(t('settings.error'), 'err'); }
  };

  const handleLogoutAll = async () => {
    setLogoutSaving(true);
    try {
      await apiClient.post('/user/settings/logout-all');
      // JWT у нас stateless (в localStorage). Сервер не может «убить» уже выданный токен,
      // поэтому для MVP делаем явный logout на клиенте.
      localStorage.removeItem('token');
      showToast(t('settings.sessions.logoutAllDone'), 'ok');
      loadSessions();
      window.location.href = '/auth';
    }
    catch { showToast(t('settings.error'), 'err'); }
    finally { setLogoutSaving(false); }
  };

  return (
    <div className="space-y-6">
      {/* Change Password */}
      <div className="glass-card p-4 sm:p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><Lock className="w-5 h-5 text-amber-400" />{t('settings.password.title')}</h3>
        <div className="space-y-3 max-w-md">
          <div className="relative">
            <input type={showOld ? 'text' : 'password'} value={oldPw} onChange={e => setOldPw(e.target.value)} placeholder={t('settings.password.current')} className="xplr-input w-full pr-10" />
            <button onClick={() => setShowOld(!showOld)} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white">{showOld ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}</button>
          </div>
          <div className="relative">
            <input type={showNew ? 'text' : 'password'} value={newPw} onChange={e => setNewPw(e.target.value)} placeholder={t('settings.password.new')} className="xplr-input w-full pr-10" />
            <button onClick={() => setShowNew(!showNew)} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white">{showNew ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}</button>
          </div>
          <input type="password" value={confirmPw} onChange={e => setConfirmPw(e.target.value)} placeholder={t('settings.password.confirm')} className="xplr-input w-full" />
          <button onClick={handleChangePassword} disabled={pwSaving || !oldPw || !newPw || !confirmPw} className="flex items-center gap-2 px-5 py-2.5 bg-amber-500 hover:bg-amber-600 disabled:opacity-50 text-white text-sm font-medium rounded-xl transition-colors">
            {pwSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Lock className="w-4 h-4" />}{t('settings.password.change')}
          </button>
        </div>
      </div>

      {/* 2FA */}
      <div className="glass-card p-4 sm:p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><Smartphone className="w-5 h-5 text-emerald-400" />{t('settings.twoFa.title')}</h3>
        {profile?.two_factor_enabled ? (
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-3">
              <ShieldCheck className="w-6 h-6 text-emerald-400 shrink-0" />
              <div><p className="text-white font-medium">{t('settings.twoFa.enabled')}</p><p className="text-sm text-slate-400">{t('settings.twoFa.enabledDesc')}</p></div>
            </div>
            <button onClick={handleDisable2FA} className="w-full sm:w-auto px-4 py-2.5 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 text-sm font-medium rounded-xl transition-colors text-center">{t('settings.twoFa.disable')}</button>
          </div>
        ) : totpSecret ? (
          <div className="space-y-4">
            <p className="text-sm text-slate-400">{t('settings.twoFa.scanQr')}</p>
            <div className="grid md:grid-cols-2 gap-6">
              <div className="flex flex-col items-center gap-3 p-4 bg-white rounded-xl">
                <img src={`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(totpURI)}`} alt="QR" className="w-44 h-44" />
              </div>
              <div className="space-y-3">
                <div><label className="block text-xs text-slate-400 mb-1">{t('settings.twoFa.secretLabel')}</label><code className="block p-3 bg-white/5 rounded-lg text-xs text-blue-400 font-mono break-all select-all">{totpSecret}</code></div>
                <input type="text" value={totpCode} onChange={e => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))} placeholder={t('settings.twoFa.enterCode')} className="xplr-input w-full text-center tracking-widest text-lg" maxLength={6} />
                <button onClick={handleVerify2FA} disabled={totpSaving || totpCode.length !== 6} className="w-full px-4 py-3 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-50 text-white font-medium rounded-xl transition-colors flex items-center justify-center gap-2">
                  {totpSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <ShieldCheck className="w-4 h-4" />}{t('settings.twoFa.confirmAndEnable')}
                </button>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-3">
              <ShieldAlert className="w-6 h-6 text-orange-400 shrink-0" />
              <div><p className="text-white font-medium">{t('settings.twoFa.notConfigured')}</p><p className="text-sm text-slate-400">{t('settings.twoFa.notConfiguredDesc')}</p></div>
            </div>
            <button onClick={handleSetup2FA} className="w-full sm:w-auto px-4 py-2.5 bg-emerald-500 hover:bg-emerald-600 text-white text-sm font-medium rounded-xl transition-colors text-center">{t('settings.twoFa.setup')}</button>
          </div>
        )}
      </div>

      {/* Sessions */}
      <div className="glass-card p-4 sm:p-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-4">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2"><MonitorSmartphone className="w-5 h-5 text-blue-400" />{t('settings.sessions.title')}</h3>
          <button onClick={handleLogoutAll} disabled={logoutSaving} className="flex items-center justify-center gap-2 px-3 py-2 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 text-xs font-medium rounded-lg transition-colors w-full sm:w-auto">
            {logoutSaving ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <LogOut className="w-3.5 h-3.5" />}{t('settings.sessions.logoutAll')}
          </button>
        </div>
        {sessions.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-b border-white/10">
                <th className="text-left px-3 py-2 text-slate-400 font-medium text-xs">{t('settings.sessions.date')}</th>
                <th className="text-left px-3 py-2 text-slate-400 font-medium text-xs">{t('settings.sessions.ip')}</th>
                <th className="text-left px-3 py-2 text-slate-400 font-medium text-xs">{t('settings.sessions.device')}</th>
              </tr></thead>
              <tbody>{sessions.map(s => (
                <tr key={s.id} className="border-b border-white/5">
                  <td className="px-3 py-2.5 text-slate-300 text-xs">{s.last_active ? new Date(s.last_active).toLocaleString() : '—'}</td>
                  <td className="px-3 py-2.5 font-mono text-slate-400 text-xs">{s.ip || '—'}</td>
                  <td className="px-3 py-2.5 text-slate-400 text-xs">{s.device ? (s.device.length > 60 ? s.device.slice(0, 60) + '…' : s.device) : '—'}</td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        ) : <p className="text-sm text-slate-500 text-center py-4">{t('settings.sessions.noSessions')}</p>}
      </div>
    </div>
  );
};

// ══════════════════════════════════════
// NOTIFICATIONS TAB
// ══════════════════════════════════════
const CHANNEL_OPTIONS: { value: string; label: string; desc: string }[] = [
  { value: 'both', label: 'Email + Telegram', desc: 'Уведомления в оба канала' },
  { value: 'email', label: 'Только Email', desc: 'Уведомления только на почту' },
  { value: 'telegram', label: 'Только Telegram', desc: 'Уведомления только в Telegram' },
];

const NotificationsTab = ({ showToast, telegramLinked }: { showToast: (m: string, t: 'ok' | 'err') => void; telegramLinked: boolean }) => {
  const { t } = useTranslation();
  const [prefs, setPrefs] = useState<NotifPrefs>({ notify_transactions: true, notify_balance: true, notify_security: true });
  const [notifChannel, setNotifChannel] = useState('both');
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    apiClient.get('/user/settings/notifications').then(res => {
      setPrefs({ notify_transactions: res.data.notify_transactions, notify_balance: res.data.notify_balance, notify_security: res.data.notify_security });
      setNotifChannel(res.data.notification_pref || 'both');
      setLoaded(true);
    }).catch(() => setLoaded(true));
  }, []);

  const savePrefs = async (updated: NotifPrefs) => {
    setPrefs(updated);
    try { await apiClient.patch('/user/settings/notifications', updated); }
    catch { showToast(t('settings.notif.saveError'), 'err'); }
  };

  const saveChannel = async (ch: string) => {
    setNotifChannel(ch);
    try { await apiClient.patch('/user/settings/notifications', { notification_pref: ch }); showToast('Канал уведомлений обновлён', 'ok'); }
    catch { showToast('Необходимо оставить хотя бы один способ связи', 'err'); }
  };

  if (!loaded) return <div className="flex justify-center py-12"><Loader2 className="w-6 h-6 animate-spin text-blue-400" /></div>;

  const items = [
    { key: 'notify_transactions' as const, label: t('settings.notif.transactions'), desc: t('settings.notif.transactionsDesc') },
    { key: 'notify_balance' as const, label: t('settings.notif.balance'), desc: t('settings.notif.balanceDesc') },
    { key: 'notify_security' as const, label: t('settings.notif.security'), desc: t('settings.notif.securityDesc') },
  ];

  return (
    <div className="space-y-6">
      {/* Notification Channel Selector */}
      <div className="glass-card p-4 sm:p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><Globe className="w-5 h-5 text-purple-400" />Канал уведомлений</h3>
        <div className="grid gap-3 sm:grid-cols-3">
          {CHANNEL_OPTIONS.map(opt => {
            const needsTG = opt.value === 'both' || opt.value === 'telegram';
            const isDisabled = needsTG && !telegramLinked;
            return (
              <div key={opt.value} className="relative group">
                <button
                  onClick={() => !isDisabled && saveChannel(opt.value)}
                  disabled={isDisabled}
                  className={`w-full p-4 rounded-xl border text-left transition-all ${
                    isDisabled
                      ? 'border-white/5 bg-white/[0.01] opacity-40 cursor-not-allowed'
                      : notifChannel === opt.value
                        ? 'border-blue-500/50 bg-blue-500/10'
                        : 'border-white/5 bg-white/[0.02] hover:bg-white/[0.05]'
                  }`}
                >
                  <p className={`text-sm font-medium ${isDisabled ? 'text-slate-500' : notifChannel === opt.value ? 'text-blue-400' : 'text-white'}`}>{opt.label}</p>
                  <p className="text-xs text-slate-500 mt-1">{opt.desc}</p>
                </button>
                {isDisabled && (
                  <div className="absolute -top-8 left-1/2 -translate-x-1/2 px-3 py-1.5 bg-slate-800 border border-white/10 rounded-lg text-xs text-slate-300 whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                    Сначала привяжите Telegram в блоке выше
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Toggle Prefs */}
      <div className="glass-card p-4 sm:p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><Bell className="w-5 h-5 text-blue-400" />{t('settings.notif.title')}</h3>
        <div className="space-y-4">
          {items.map(item => (
            <div key={item.key} className="flex items-center justify-between gap-4 p-4 rounded-xl bg-white/[0.03] border border-white/5">
              <div className="min-w-0"><p className="text-white font-medium">{item.label}</p><p className="text-sm text-slate-500">{item.desc}</p></div>
              <Toggle checked={prefs[item.key]} onChange={v => savePrefs({ ...prefs, [item.key]: v })} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════
// KYC TAB — wired to POST /user/settings/kyc
// ══════════════════════════════════════
const KYCTab = ({ profile, reload, showToast }: { profile: ProfileData | null; reload: () => void; showToast: (m: string, t: 'ok' | 'err') => void }) => {
  const { t } = useTranslation();
  const [step, setStep] = useState(1);
  const [country, setCountry] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [birthDate, setBirthDate] = useState('');
  const [address, setAddress] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [existingKyc, setExistingKyc] = useState<any>(null);
  const verStatus = profile?.verification_status || 'pending';

  useEffect(() => {
    apiClient.get('/user/settings/kyc').then(res => { if (res.data) setExistingKyc(res.data); }).catch(() => {});
  }, []);

  if (verStatus === 'verified') {
    return (
      <div className="glass-card p-8 text-center">
        <CheckCircle className="w-16 h-16 text-emerald-400 mx-auto mb-4" />
        <h3 className="text-xl font-bold text-white mb-2">{t('settings.kycForm.accountVerified')}</h3>
        <p className="text-slate-400">{t('settings.kycForm.fullAccess')}</p>
      </div>
    );
  }

  if (existingKyc && existingKyc.status === 'pending') {
    return (
      <div className="glass-card p-8 text-center">
        <Clock className="w-16 h-16 text-orange-400 mx-auto mb-4" />
        <h3 className="text-xl font-bold text-white mb-2">{t('settings.verification.pending')}</h3>
        <p className="text-slate-400">{t('settings.kycForm.alreadyPending')}</p>
      </div>
    );
  }

  const handleSubmit = async () => {
    if (!country || !firstName.trim() || !lastName.trim()) {
      showToast(t('settings.kycForm.fillRequired'), 'err');
      return;
    }
    setSubmitting(true);
    try {
      await apiClient.post('/user/settings/kyc', { country, first_name: firstName.trim(), last_name: lastName.trim(), birth_date: birthDate, address: address.trim() });
      showToast(t('settings.kycForm.submitSuccess'), 'ok');
      reload();
      setExistingKyc({ status: 'pending' });
    } catch {
      showToast(t('settings.kycForm.submitError'), 'err');
    } finally { setSubmitting(false); }
  };

  const steps = [
    { id: 1, title: t('settings.kycForm.stepCitizenship'), status: step > 1 ? 'completed' : step === 1 ? 'current' : 'pending' },
    { id: 2, title: t('settings.kycForm.stepPersonalData'), status: step > 2 ? 'completed' : step === 2 ? 'current' : 'pending' },
    { id: 3, title: t('settings.kycForm.stepDocuments'), status: step === 3 ? 'current' : 'pending' },
  ];

  return (
    <div className="space-y-6">
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-6">{t('settings.kycForm.title')}</h3>
        <div className="flex items-center justify-between mb-8 overflow-x-auto pb-2">
          {steps.map((s, i) => (
            <div key={s.id} className="flex items-center shrink-0">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${
                s.status === 'completed' ? 'bg-emerald-500 text-white' : s.status === 'current' ? 'bg-blue-500 text-white' : 'bg-white/10 text-slate-500'
              }`}>{s.status === 'completed' ? <Check className="w-5 h-5" /> : s.id}</div>
              <span className={`ml-2 sm:ml-3 text-sm sm:text-base hidden sm:inline ${s.status === 'current' ? 'text-white' : 'text-slate-500'}`}>{s.title}</span>
              {i < steps.length - 1 && <div className="w-8 sm:w-16 md:w-24 h-0.5 bg-white/10 mx-2 sm:mx-4" />}
            </div>
          ))}
        </div>
      </div>

      {step === 1 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">{t('settings.kycForm.stepCitizenship')}</h3>
          <select value={country} onChange={e => setCountry(e.target.value)} className="xplr-select w-full mb-4">
            <option value="">{t('settings.kycForm.selectCountry')}</option>
            <option value="RU">{t('settings.kycForm.countries.RU')}</option>
            <option value="BY">{t('settings.kycForm.countries.BY')}</option>
            <option value="KZ">{t('settings.kycForm.countries.KZ')}</option>
            <option value="UA">{t('settings.kycForm.countries.UA')}</option>
            <option value="US">{t('settings.kycForm.countries.US')}</option>
            <option value="DE">{t('settings.kycForm.countries.DE')}</option>
          </select>
          <button onClick={() => country && setStep(2)} disabled={!country} className="w-full px-4 py-3 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white font-medium rounded-xl transition-colors">{t('settings.kycForm.continue')}</button>
        </div>
      )}

      {step === 2 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">{t('settings.kycForm.stepPersonalData')}</h3>
          <div className="grid gap-4 mb-6">
            <div className="grid md:grid-cols-2 gap-4">
              <div><label className="block text-sm text-slate-400 mb-2">{t('settings.kycForm.firstName')}</label><input type="text" value={firstName} onChange={e => setFirstName(e.target.value)} className="xplr-input w-full" /></div>
              <div><label className="block text-sm text-slate-400 mb-2">{t('settings.kycForm.lastName')}</label><input type="text" value={lastName} onChange={e => setLastName(e.target.value)} className="xplr-input w-full" /></div>
            </div>
            <div><label className="block text-sm text-slate-400 mb-2">{t('settings.kycForm.birthDate')}</label><input type="date" value={birthDate} onChange={e => setBirthDate(e.target.value)} className="xplr-input w-full" /></div>
            <div><label className="block text-sm text-slate-400 mb-2">{t('settings.kycForm.address')}</label><input type="text" value={address} onChange={e => setAddress(e.target.value)} className="xplr-input w-full" /></div>
          </div>
          <div className="flex gap-3">
            <button onClick={() => setStep(1)} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl transition-colors">{t('settings.kycForm.back')}</button>
            <button onClick={() => { if (firstName.trim() && lastName.trim()) setStep(3); else showToast(t('settings.kycForm.fillRequired'), 'err'); }} className="flex-1 px-4 py-3 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-xl transition-colors">{t('settings.kycForm.continue')}</button>
          </div>
        </div>
      )}

      {step === 3 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">{t('settings.kycForm.uploadDocs')}</h3>
          <div className="space-y-4 mb-6">
            {[
              { id: 'passport', label: t('settings.kycForm.passport'), desc: t('settings.kycForm.passportDesc') },
              { id: 'address', label: t('settings.kycForm.addressProof'), desc: t('settings.kycForm.addressProofDesc') },
              { id: 'selfie', label: t('settings.kycForm.selfie'), desc: t('settings.kycForm.selfieDesc') },
            ].map(doc => (
              <div key={doc.id} className="p-4 rounded-xl bg-white/[0.03] border border-white/10 border-dashed">
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <Upload className="w-5 h-5 text-slate-400 shrink-0" />
                    <div className="min-w-0"><p className="text-white font-medium">{doc.label}</p><p className="text-sm text-slate-500 truncate">{doc.desc}</p></div>
                  </div>
                  <button className="w-full sm:w-auto px-4 py-2.5 glass-card hover:bg-white/10 text-slate-300 text-sm rounded-xl transition-colors text-center shrink-0">{t('settings.kycForm.upload')}</button>
                </div>
              </div>
            ))}
          </div>
          <div className="flex gap-3">
            <button onClick={() => setStep(2)} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl transition-colors">{t('settings.kycForm.back')}</button>
            <button onClick={handleSubmit} disabled={submitting} className="flex-1 px-4 py-3 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-50 text-white font-medium rounded-xl transition-colors flex items-center justify-center gap-2">
              {submitting && <Loader2 className="w-4 h-4 animate-spin" />}{t('settings.kycForm.submitReview')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

// ══════════════════════════════════════
// LANGUAGE TAB
// ══════════════════════════════════════
const LanguageTab = () => {
  const { t, i18n } = useTranslation();
  const currentLang = i18n.language;
  const languages = [
    { code: 'ru', name: 'Русский (RU)', flag: '🇷🇺' },
    { code: 'en', name: 'English (EN)', flag: '🇺🇸' },
    { code: 'es', name: 'Español (ES)', flag: '🇪🇸' },
    { code: 'pt', name: 'Português (PT)', flag: '🇧🇷' },
    { code: 'tr', name: 'Türkçe (TR)', flag: '🇹🇷' },
    { code: 'zh', name: '中文 (ZH)', flag: '🇨🇳' },
  ];
  const handleChange = (code: string) => { i18n.changeLanguage(code); localStorage.setItem(LANG_KEY, code); };

  return (
    <div className="glass-card p-4 sm:p-6">
      <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2"><Globe className="w-5 h-5 text-blue-400" />{t('settings.languageTitle')}</h3>
      <div className="space-y-3">
        {languages.map(lang => (
          <button key={lang.code} onClick={() => handleChange(lang.code)} className={`w-full flex items-center justify-between p-4 rounded-xl transition-all duration-150 ${
            currentLang === lang.code ? 'bg-blue-500/20 border border-blue-500/50' : 'bg-white/[0.03] border border-white/5 hover:border-white/10'
          }`}>
            <div className="flex items-center gap-3"><span className="text-2xl">{lang.flag}</span><span className="text-white font-medium">{lang.name}</span></div>
            {currentLang === lang.code && <CheckCircle className="w-5 h-5 text-blue-400" />}
          </button>
        ))}
      </div>
    </div>
  );
};

// ══════════════════════════════════════
// MAIN SETTINGS PAGE
// ══════════════════════════════════════
export const SettingsPage = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<SettingsTab>('profile');
  const [profile, setProfile] = useState<ProfileData | null>(null);
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null);

  const showToast = (msg: string, type: 'ok' | 'err' = 'ok') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 4000);
  };

  const loadProfile = useCallback(async () => {
    try { const res = await apiClient.get('/user/settings/profile'); setProfile(res.data); } catch { /* ignore */ }
  }, []);
  useEffect(() => { loadProfile(); }, [loadProfile]);
  
  const tabs = [
    { id: 'profile' as const, label: t('settings.profile'), icon: User },
    { id: 'security' as const, label: t('settings.security'), icon: Shield },
    { id: 'notifications' as const, label: t('settings.notifications'), icon: Bell },
    { id: 'kyc' as const, label: t('settings.kyc'), icon: FileText },
    { id: 'language' as const, label: t('settings.language'), icon: Globe },
  ];

  return (
    <DashboardLayout>
      <div className="stagger-fade-in w-full max-w-4xl overflow-hidden">
        <BackButton />
        {toast && <Toast msg={toast.msg} type={toast.type} />}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">{t('settings.title')}</h1>
          <p className="text-slate-400">{t('settings.subtitle')}</p>
        </div>
        <div className="flex flex-wrap gap-2 mb-8">
          {tabs.map(tab => (
            <button key={tab.id} onClick={() => setActiveTab(tab.id)} className={`flex items-center gap-2 px-4 py-3 rounded-xl font-medium transition-all whitespace-nowrap ${
              activeTab === tab.id ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/25' : 'glass-card text-slate-400 hover:text-white'
            }`}><tab.icon className="w-4 h-4" />{tab.label}</button>
          ))}
        </div>
        {activeTab === 'profile' && <ProfileTab profile={profile} reload={loadProfile} showToast={showToast} setActiveTab={setActiveTab} />}
        {activeTab === 'security' && <SecurityTab profile={profile} reload={loadProfile} showToast={showToast} />}
        {activeTab === 'notifications' && <NotificationsTab showToast={showToast} telegramLinked={profile?.telegram_linked ?? false} />}
        {activeTab === 'kyc' && <KYCTab profile={profile} reload={loadProfile} showToast={showToast} />}
        {activeTab === 'language' && <LanguageTab />}
      </div>
    </DashboardLayout>
  );
};
