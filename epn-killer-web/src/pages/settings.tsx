import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { LANG_KEY } from '../i18n';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import apiClient from '../api/axios';
import { 
  User,
  Mail,
  Lock,
  Bell,
  Shield,
  Eye,
  EyeOff,
  Save,
  Check,
  Smartphone,
  Globe,
  FileText,
  Upload,
  CheckCircle,
  Clock,
  AlertCircle,
  MessageCircle,
  LogOut,
  Loader2,
  MonitorSmartphone,
  ShieldCheck,
  ShieldAlert,
  ExternalLink
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

// ── Toggle Switch ──
const Toggle = ({ checked, onChange, disabled = false }: { checked: boolean; onChange: (v: boolean) => void; disabled?: boolean }) => (
  <label className={`relative inline-flex items-center ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}>
    <input type="checkbox" checked={checked} onChange={(e) => !disabled && onChange(e.target.checked)} className="sr-only peer" disabled={disabled} />
    <div className="w-11 h-6 bg-white/10 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-500/50 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-white/20 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
  </label>
);

// ── Toast ──
const Toast = ({ msg, type }: { msg: string; type: 'ok' | 'err' }) => (
  <div className={`fixed top-6 right-6 z-50 flex items-center gap-3 px-5 py-4 rounded-2xl text-sm font-medium shadow-2xl backdrop-blur-lg border ${
    type === 'ok' ? 'bg-emerald-500/90 text-white border-emerald-400/30' : 'bg-red-500/90 text-white border-red-400/30'
  }`}>
    {type === 'ok' ? <CheckCircle className="w-5 h-5 shrink-0" /> : <AlertCircle className="w-5 h-5 shrink-0" />}
    {msg}
  </div>
);

// ══════════════════════════════════════
// PROFILE TAB
// ══════════════════════════════════════
const ProfileTab = ({ profile, reload, showToast }: { profile: ProfileData | null; reload: () => void; showToast: (m: string, t: 'ok' | 'err') => void }) => {
  const [displayName, setDisplayName] = useState(profile?.display_name || '');
  const [saving, setSaving] = useState(false);

  useEffect(() => { if (profile) setDisplayName(profile.display_name || ''); }, [profile]);

  const handleSave = async () => {
    if (!displayName.trim()) return;
    setSaving(true);
    try {
      await apiClient.patch('/user/settings/profile', { display_name: displayName.trim() });
      showToast('Профиль обновлён', 'ok');
      reload();
    } catch { showToast('Ошибка сохранения', 'err'); }
    finally { setSaving(false); }
  };

  const verStatus = profile?.verification_status || 'pending';
  const verColors: Record<string, string> = {
    pending: 'bg-orange-500/20 text-orange-400',
    verified: 'bg-emerald-500/20 text-emerald-400',
    rejected: 'bg-red-500/20 text-red-400',
  };

  return (
    <div className="space-y-6">
      {/* Profile Info */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <User className="w-5 h-5 text-blue-400" />
          Профиль
        </h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-2">Email</label>
            <div className="flex items-center gap-3">
              <input type="text" value={profile?.email || ''} readOnly className="xplr-input w-full bg-white/[0.02] cursor-not-allowed" />
              {profile?.is_verified ? (
                <span className="flex items-center gap-1 text-xs text-emerald-400 whitespace-nowrap"><CheckCircle className="w-3.5 h-3.5" />Подтверждён</span>
              ) : (
                <span className="flex items-center gap-1 text-xs text-red-400 whitespace-nowrap"><AlertCircle className="w-3.5 h-3.5" />Не подтверждён</span>
              )}
            </div>
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-2">Отображаемое имя</label>
            <input type="text" value={displayName} onChange={e => setDisplayName(e.target.value)} placeholder="Ваше имя" className="xplr-input w-full" />
          </div>
          <button onClick={handleSave} disabled={saving || !displayName.trim()} className="flex items-center gap-2 px-5 py-2.5 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white text-sm font-medium rounded-xl transition-colors">
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            Сохранить
          </button>
        </div>
      </div>

      {/* Verification Status */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <FileText className="w-5 h-5 text-purple-400" />
          Верификация
        </h3>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className={`px-3 py-1 rounded-full text-xs font-medium ${verColors[verStatus] || verColors.pending}`}>
              {verStatus === 'pending' ? 'Ожидает' : verStatus === 'verified' ? 'Верифицирован' : 'Отклонено'}
            </span>
            <p className="text-sm text-slate-400">
              {verStatus === 'verified' ? 'Ваш аккаунт верифицирован' : 'Пройдите верификацию для полного доступа'}
            </p>
          </div>
          {verStatus !== 'verified' && (
            <button className="px-4 py-2 bg-purple-500 hover:bg-purple-600 text-white text-sm font-medium rounded-lg transition-colors whitespace-nowrap">
              Пройти верификацию
            </button>
          )}
        </div>
      </div>

      {/* Telegram */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <MessageCircle className="w-5 h-5 text-blue-400" />
          Telegram
        </h3>
        <div className="flex items-center justify-between">
          {profile?.telegram_linked ? (
            <>
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-emerald-400" />
                <p className="text-white font-medium">Подключен</p>
              </div>
              <button className="px-4 py-2 glass-card hover:bg-white/10 text-slate-300 text-sm rounded-lg transition-colors">
                Отключить
              </button>
            </>
          ) : (
            <>
              <p className="text-sm text-slate-400">Привяжите Telegram для уведомлений и 2FA</p>
              <a href="https://t.me/your_telegram_link" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium rounded-lg transition-colors">
                Привязать <ExternalLink className="w-3.5 h-3.5" />
              </a>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════
// SECURITY TAB
// ══════════════════════════════════════
const SecurityTab = ({ profile, reload, showToast }: { profile: ProfileData | null; reload: () => void; showToast: (m: string, t: 'ok' | 'err') => void }) => {
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
    try {
      const res = await apiClient.get('/user/settings/sessions');
      setSessions(res.data || []);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => { loadSessions(); }, [loadSessions]);

  const handleChangePassword = async () => {
    if (newPw !== confirmPw) { showToast('Пароли не совпадают', 'err'); return; }
    if (newPw.length < 8) { showToast('Минимум 8 символов', 'err'); return; }
    setPwSaving(true);
    try {
      await apiClient.post('/user/settings/change-password', { old_password: oldPw, new_password: newPw });
      showToast('Пароль изменён', 'ok');
      setOldPw(''); setNewPw(''); setConfirmPw('');
    } catch (err: any) {
      const msg = typeof err.response?.data === 'string' ? err.response.data : 'Ошибка смены пароля';
      showToast(msg, 'err');
    } finally { setPwSaving(false); }
  };

  const handleSetup2FA = async () => {
    try {
      const res = await apiClient.post('/user/settings/2fa/setup');
      setTotpSecret(res.data.secret);
      setTotpURI(res.data.otp_uri);
    } catch { showToast('Ошибка настройки 2FA', 'err'); }
  };

  const handleVerify2FA = async () => {
    if (totpCode.length !== 6) { showToast('Введите 6-значный код', 'err'); return; }
    setTotpSaving(true);
    try {
      await apiClient.post('/user/settings/2fa/verify', { code: totpCode });
      showToast('2FA включена!', 'ok');
      setTotpSecret(''); setTotpURI(''); setTotpCode('');
      reload();
    } catch (err: any) {
      const msg = typeof err.response?.data === 'string' ? err.response.data : 'Неверный код';
      showToast(msg, 'err');
    } finally { setTotpSaving(false); }
  };

  const handleDisable2FA = async () => {
    try {
      await apiClient.post('/user/settings/2fa/disable');
      showToast('2FA отключена', 'ok');
      reload();
    } catch { showToast('Ошибка', 'err'); }
  };

  const handleLogoutAll = async () => {
    setLogoutSaving(true);
    try {
      await apiClient.post('/user/settings/logout-all');
      showToast('Все сессии завершены', 'ok');
      loadSessions();
    } catch { showToast('Ошибка', 'err'); }
    finally { setLogoutSaving(false); }
  };

  return (
    <div className="space-y-6">
      {/* Change Password */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Lock className="w-5 h-5 text-amber-400" />
          Смена пароля
        </h3>
        <div className="space-y-3 max-w-md">
          <div className="relative">
            <input type={showOld ? 'text' : 'password'} value={oldPw} onChange={e => setOldPw(e.target.value)} placeholder="Текущий пароль" className="xplr-input w-full pr-10" />
            <button onClick={() => setShowOld(!showOld)} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white">
              {showOld ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
          <div className="relative">
            <input type={showNew ? 'text' : 'password'} value={newPw} onChange={e => setNewPw(e.target.value)} placeholder="Новый пароль (мин. 8 символов)" className="xplr-input w-full pr-10" />
            <button onClick={() => setShowNew(!showNew)} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white">
              {showNew ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
          <input type="password" value={confirmPw} onChange={e => setConfirmPw(e.target.value)} placeholder="Подтвердите новый пароль" className="xplr-input w-full" />
          <button onClick={handleChangePassword} disabled={pwSaving || !oldPw || !newPw || !confirmPw} className="flex items-center gap-2 px-5 py-2.5 bg-amber-500 hover:bg-amber-600 disabled:opacity-50 text-white text-sm font-medium rounded-xl transition-colors">
            {pwSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Lock className="w-4 h-4" />}
            Сменить пароль
          </button>
        </div>
      </div>

      {/* 2FA Google Authenticator */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Smartphone className="w-5 h-5 text-emerald-400" />
          Google Authenticator (2FA)
        </h3>
        {profile?.two_factor_enabled ? (
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <ShieldCheck className="w-6 h-6 text-emerald-400" />
              <div>
                <p className="text-white font-medium">2FA включена</p>
                <p className="text-sm text-slate-400">Ваш аккаунт защищён двухфакторной аутентификацией</p>
              </div>
            </div>
            <button onClick={handleDisable2FA} className="px-4 py-2 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 text-sm font-medium rounded-lg transition-colors">
              Отключить
            </button>
          </div>
        ) : totpSecret ? (
          <div className="space-y-4">
            <p className="text-sm text-slate-400">Отсканируйте QR-код или введите секрет в Google Authenticator:</p>
            <div className="grid md:grid-cols-2 gap-6">
              <div className="flex flex-col items-center gap-3 p-4 bg-white rounded-xl">
                <img src={`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(totpURI)}`} alt="QR" className="w-44 h-44" />
              </div>
              <div className="space-y-3">
                <div>
                  <label className="block text-xs text-slate-400 mb-1">Секрет (ручной ввод)</label>
                  <code className="block p-3 bg-white/5 rounded-lg text-xs text-blue-400 font-mono break-all select-all">{totpSecret}</code>
                </div>
                <input type="text" value={totpCode} onChange={e => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))} placeholder="Введите 6-значный код" className="xplr-input w-full text-center tracking-widest text-lg" maxLength={6} />
                <button onClick={handleVerify2FA} disabled={totpSaving || totpCode.length !== 6} className="w-full px-4 py-3 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-50 text-white font-medium rounded-xl transition-colors flex items-center justify-center gap-2">
                  {totpSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <ShieldCheck className="w-4 h-4" />}
                  Подтвердить и включить
                </button>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <ShieldAlert className="w-6 h-6 text-orange-400" />
              <div>
                <p className="text-white font-medium">2FA не настроена</p>
                <p className="text-sm text-slate-400">Рекомендуем включить для защиты аккаунта</p>
              </div>
            </div>
            <button onClick={handleSetup2FA} className="px-4 py-2 bg-emerald-500 hover:bg-emerald-600 text-white text-sm font-medium rounded-lg transition-colors">
              Настроить
            </button>
          </div>
        )}
      </div>

      {/* Sessions */}
      <div className="glass-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <MonitorSmartphone className="w-5 h-5 text-blue-400" />
            Последняя активность
          </h3>
          <button onClick={handleLogoutAll} disabled={logoutSaving} className="flex items-center gap-2 px-3 py-1.5 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 text-xs font-medium rounded-lg transition-colors">
            {logoutSaving ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <LogOut className="w-3.5 h-3.5" />}
            Выйти со всех устройств
          </button>
        </div>
        {sessions.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/10">
                  <th className="text-left px-3 py-2 text-slate-400 font-medium text-xs">Дата</th>
                  <th className="text-left px-3 py-2 text-slate-400 font-medium text-xs">IP</th>
                  <th className="text-left px-3 py-2 text-slate-400 font-medium text-xs">Устройство</th>
                </tr>
              </thead>
              <tbody>
                {sessions.map(s => (
                  <tr key={s.id} className="border-b border-white/5">
                    <td className="px-3 py-2.5 text-slate-300 text-xs">{s.last_active ? new Date(s.last_active).toLocaleString('ru-RU') : '—'}</td>
                    <td className="px-3 py-2.5 font-mono text-slate-400 text-xs">{s.ip || '—'}</td>
                    <td className="px-3 py-2.5 text-slate-400 text-xs">{s.device || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-slate-500 text-center py-4">Нет данных о сессиях</p>
        )}
      </div>
    </div>
  );
};

// ══════════════════════════════════════
// NOTIFICATIONS TAB
// ══════════════════════════════════════
const NotificationsTab = ({ showToast }: { showToast: (m: string, t: 'ok' | 'err') => void }) => {
  const [prefs, setPrefs] = useState<NotifPrefs>({ notify_transactions: true, notify_balance: true, notify_security: true });
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    apiClient.get('/user/settings/notifications').then(res => { setPrefs(res.data); setLoaded(true); }).catch(() => setLoaded(true));
  }, []);

  const savePrefs = async (updated: NotifPrefs) => {
    setPrefs(updated);
    try {
      await apiClient.patch('/user/settings/notifications', updated);
    } catch { showToast('Ошибка сохранения', 'err'); }
  };

  if (!loaded) return <div className="flex justify-center py-12"><Loader2 className="w-6 h-6 animate-spin text-blue-400" /></div>;

  return (
    <div className="space-y-6">
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Bell className="w-5 h-5 text-blue-400" />
          Уведомления
        </h3>
        <div className="space-y-4">
          {[
            { key: 'notify_transactions' as const, label: 'Транзакции', desc: 'Операции по картам, списания, пополнения' },
            { key: 'notify_balance' as const, label: 'Баланс', desc: 'Изменения баланса, низкий остаток' },
            { key: 'notify_security' as const, label: 'Безопасность', desc: 'Входы, смена пароля, подозрительные действия' },
          ].map(item => (
            <div key={item.key} className="flex items-center justify-between p-4 rounded-xl bg-white/[0.03] border border-white/5">
              <div>
                <p className="text-white font-medium">{item.label}</p>
                <p className="text-sm text-slate-500">{item.desc}</p>
              </div>
              <Toggle checked={prefs[item.key]} onChange={v => savePrefs({ ...prefs, [item.key]: v })} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════
// KYC TAB
// ══════════════════════════════════════
const KYCTab = ({ profile }: { profile: ProfileData | null }) => {
  const [step, setStep] = useState(1);
  const [country, setCountry] = useState('');
  const verStatus = profile?.verification_status || 'pending';

  if (verStatus === 'verified') {
    return (
      <div className="glass-card p-8 text-center">
        <CheckCircle className="w-16 h-16 text-emerald-400 mx-auto mb-4" />
        <h3 className="text-xl font-bold text-white mb-2">Аккаунт верифицирован</h3>
        <p className="text-slate-400">Вам доступен полный функционал платформы.</p>
      </div>
    );
  }

  const steps = [
    { id: 1, title: 'Гражданство', status: step > 1 ? 'completed' : step === 1 ? 'current' : 'pending' },
    { id: 2, title: 'Личные данные', status: step > 2 ? 'completed' : step === 2 ? 'current' : 'pending' },
    { id: 3, title: 'Документы', status: step === 3 ? 'current' : 'pending' },
  ];

  return (
    <div className="space-y-6">
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-6">Верификация аккаунта</h3>
        <div className="flex items-center justify-between mb-8">
          {steps.map((s, i) => (
            <div key={s.id} className="flex items-center">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${
                s.status === 'completed' ? 'bg-emerald-500 text-white' : s.status === 'current' ? 'bg-blue-500 text-white' : 'bg-white/10 text-slate-500'
              }`}>
                {s.status === 'completed' ? <Check className="w-5 h-5" /> : s.id}
              </div>
              <span className={`ml-3 ${s.status === 'current' ? 'text-white' : 'text-slate-500'}`}>{s.title}</span>
              {i < steps.length - 1 && <div className="w-16 md:w-24 h-0.5 bg-white/10 mx-4" />}
            </div>
          ))}
        </div>
      </div>
      {step === 1 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Выберите гражданство</h3>
          <select value={country} onChange={e => setCountry(e.target.value)} className="xplr-select w-full mb-4">
            <option value="">Выберите страну</option>
            <option value="RU">Россия</option>
            <option value="BY">Беларусь</option>
            <option value="KZ">Казахстан</option>
            <option value="UA">Украина</option>
            <option value="US">США</option>
            <option value="DE">Германия</option>
          </select>
          <button onClick={() => country && setStep(2)} disabled={!country} className="w-full px-4 py-3 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white font-medium rounded-xl transition-colors">Продолжить</button>
        </div>
      )}
      {step === 2 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Личные данные</h3>
          <div className="grid gap-4 mb-6">
            <div className="grid md:grid-cols-2 gap-4">
              <div><label className="block text-sm text-slate-400 mb-2">Имя</label><input type="text" className="xplr-input w-full" placeholder="Иван" /></div>
              <div><label className="block text-sm text-slate-400 mb-2">Фамилия</label><input type="text" className="xplr-input w-full" placeholder="Иванов" /></div>
            </div>
            <div><label className="block text-sm text-slate-400 mb-2">Дата рождения</label><input type="date" className="xplr-input w-full" /></div>
            <div><label className="block text-sm text-slate-400 mb-2">Адрес</label><input type="text" className="xplr-input w-full" placeholder="Город, улица, дом" /></div>
          </div>
          <div className="flex gap-3">
            <button onClick={() => setStep(1)} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl transition-colors">Назад</button>
            <button onClick={() => setStep(3)} className="flex-1 px-4 py-3 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-xl transition-colors">Продолжить</button>
          </div>
        </div>
      )}
      {step === 3 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Загрузка документов</h3>
          <div className="space-y-4 mb-6">
            {[
              { id: 'passport', label: 'Гос. паспорт', desc: 'Фото первого разворота' },
              { id: 'address', label: 'Подтверждение адреса', desc: 'Квитанция или выписка' },
              { id: 'selfie', label: 'Селфи с документом', desc: 'Держите паспорт рядом с лицом' },
            ].map(doc => (
              <div key={doc.id} className="p-4 rounded-xl bg-white/[0.03] border border-white/10 border-dashed">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Upload className="w-5 h-5 text-slate-400" />
                    <div><p className="text-white font-medium">{doc.label}</p><p className="text-sm text-slate-500">{doc.desc}</p></div>
                  </div>
                  <button className="px-4 py-2 glass-card hover:bg-white/10 text-slate-300 text-sm rounded-lg transition-colors">Загрузить</button>
                </div>
              </div>
            ))}
          </div>
          <div className="flex gap-3">
            <button onClick={() => setStep(2)} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl transition-colors">Назад</button>
            <button className="flex-1 px-4 py-3 bg-emerald-500 hover:bg-emerald-600 text-white font-medium rounded-xl transition-colors">Отправить на проверку</button>
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
    <div className="glass-card p-6">
      <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
        <Globe className="w-5 h-5 text-blue-400" />
        {t('settings.languageTitle')}
      </h3>
      <div className="space-y-3">
        {languages.map(lang => (
          <button key={lang.code} onClick={() => handleChange(lang.code)} className={`w-full flex items-center justify-between p-4 rounded-xl transition-all duration-150 ${
            currentLang === lang.code ? 'bg-blue-500/20 border border-blue-500/50' : 'bg-white/[0.03] border border-white/5 hover:border-white/10'
          }`}>
            <div className="flex items-center gap-3">
              <span className="text-2xl">{lang.flag}</span>
              <span className="text-white font-medium">{lang.name}</span>
            </div>
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
    try {
      const res = await apiClient.get('/user/settings/profile');
      setProfile(res.data);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => { loadProfile(); }, [loadProfile]);
  
  const tabs = [
    { id: 'profile' as const, label: 'Профиль', icon: User },
    { id: 'security' as const, label: t('settings.security'), icon: Shield },
    { id: 'notifications' as const, label: t('settings.notifications'), icon: Bell },
    { id: 'kyc' as const, label: t('settings.kyc'), icon: FileText },
    { id: 'language' as const, label: t('settings.language'), icon: Globe },
  ];

  return (
    <DashboardLayout>
      <div className="stagger-fade-in max-w-4xl">
        <BackButton />

        {toast && <Toast msg={toast.msg} type={toast.type} />}
        
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">{t('settings.title')}</h1>
          <p className="text-slate-400">{t('settings.subtitle')}</p>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-8 overflow-x-auto pb-2">
          {tabs.map(tab => (
            <button key={tab.id} onClick={() => setActiveTab(tab.id)} className={`flex items-center gap-2 px-4 py-3 rounded-xl font-medium transition-all whitespace-nowrap ${
              activeTab === tab.id ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/25' : 'glass-card text-slate-400 hover:text-white'
            }`}>
              <tab.icon className="w-4 h-4" />
              {tab.label}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        {activeTab === 'profile' && <ProfileTab profile={profile} reload={loadProfile} showToast={showToast} />}
        {activeTab === 'security' && <SecurityTab profile={profile} reload={loadProfile} showToast={showToast} />}
        {activeTab === 'notifications' && <NotificationsTab showToast={showToast} />}
        {activeTab === 'kyc' && <KYCTab profile={profile} />}
        {activeTab === 'language' && <LanguageTab />}
      </div>
    </DashboardLayout>
  );
};
