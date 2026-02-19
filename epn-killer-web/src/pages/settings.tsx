import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { LANG_KEY } from '../i18n';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  User,
  Mail,
  Lock,
  Bell,
  Shield,
  Eye,
  EyeOff,
  Save,
  Camera,
  Check,
  Copy,
  RefreshCw,
  Smartphone,
  Globe,
  Key,
  FileText,
  Upload,
  CheckCircle,
  Clock,
  AlertCircle,
  MessageCircle,
  CreditCard,
  Zap
} from 'lucide-react';

type SettingsTab = 'security' | 'kyc' | 'notifications' | 'language';

// Toggle Switch Component
const Toggle = ({ checked, onChange, disabled = false }: { checked: boolean; onChange: (v: boolean) => void; disabled?: boolean }) => (
  <label className={`relative inline-flex items-center ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}>
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => !disabled && onChange(e.target.checked)}
      className="sr-only peer"
      disabled={disabled}
    />
    <div className="w-11 h-6 bg-white/10 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-500/50 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-white/20 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
  </label>
);

// Security Tab
const SecurityTab = () => {
  const [showApiToken, setShowApiToken] = useState(false);
  const [copiedToken, setCopiedToken] = useState(false);
  const [settings, setSettings] = useState({
    telegramRequisites: true,
    twoFactorLogin: true,
    twoFactorRequisites: false,
    twoFactorWithdraw: true,
    twoFactorTransfer: true
  });

  const apiToken = 'xplr_sk_live_4d8a9f2c3e5b7a1d0f9e8c7b6a5d4e3f2';
  
  const copyToken = () => {
    navigator.clipboard.writeText(apiToken);
    setCopiedToken(true);
    setTimeout(() => setCopiedToken(false), 2000);
  };

  const activityLog = [
    { date: '18.02.2026, 14:32', ip: '185.24.54.251', location: 'Россия, Москва', browser: 'Safari 17.2' },
    { date: '17.02.2026, 09:15', ip: '185.24.54.251', location: 'Россия, Москва', browser: 'Chrome 121' },
    { date: '15.02.2026, 21:48', ip: '91.108.32.123', location: 'Россия, СПб', browser: 'Firefox 122' },
  ];

  return (
    <div className="space-y-6">
      {/* Email Verification */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Mail className="w-5 h-5 text-blue-400" />
          Подтверждение Email
        </h3>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-white font-medium">aalabin5@gmail.com</p>
            <span className="inline-flex items-center gap-1 text-xs text-red-400 mt-1">
              <AlertCircle className="w-3 h-3" />
              Не подтверждён
            </span>
          </div>
          <button className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium rounded-lg transition-colors">
            Подтвердить адрес
          </button>
        </div>
      </div>

      {/* API Token */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Key className="w-5 h-5 text-purple-400" />
          API токен
        </h3>
        <p className="text-sm text-slate-400 mb-4">
          Используйте API токен для полной автоматизации работы с картами и платежами
        </p>
        <div className="flex gap-2 mb-4">
          <div className="flex-1 relative">
            <input
              type={showApiToken ? 'text' : 'password'}
              value={apiToken}
              readOnly
              className="xplr-input w-full pr-10 font-mono text-sm"
            />
            <button
              onClick={() => setShowApiToken(!showApiToken)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white transition-colors"
            >
              {showApiToken ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
          <button
            onClick={copyToken}
            className={`px-4 py-2 rounded-lg transition-colors flex items-center gap-2 ${
              copiedToken ? 'bg-emerald-500 text-white' : 'glass-card hover:bg-white/10 text-white'
            }`}
          >
            {copiedToken ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
          </button>
          <button className="px-4 py-2 glass-card hover:bg-white/10 text-white rounded-lg flex items-center gap-2">
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Telegram Bot */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <MessageCircle className="w-5 h-5 text-blue-400" />
          Telegram Bot
        </h3>
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-emerald-400" />
              <p className="text-white font-medium">Подключен: @aalabin</p>
            </div>
            <p className="text-sm text-slate-400 mt-1">2FA и управление через бота</p>
          </div>
          <button className="px-4 py-2 glass-card hover:bg-white/10 text-slate-300 text-sm rounded-lg transition-colors">
            Отключить
          </button>
        </div>
      </div>

      {/* Google Authenticator */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Smartphone className="w-5 h-5 text-emerald-400" />
          Google Authenticator
        </h3>
        <div className="grid md:grid-cols-2 gap-6">
          <div className="flex items-center justify-center p-4 bg-white rounded-xl">
            {/* QR Code placeholder */}
            <div className="w-32 h-32 bg-slate-100 rounded flex items-center justify-center text-slate-400">
              QR Code
            </div>
          </div>
          <div>
            <p className="text-sm text-slate-400 mb-4">
              Отсканируйте QR-код в приложении Google Authenticator и введите код подтверждения
            </p>
            <input
              type="text"
              placeholder="Введите 6-значный код"
              className="xplr-input w-full mb-3 text-center tracking-widest"
              maxLength={6}
            />
            <button className="w-full px-4 py-3 bg-emerald-500 hover:bg-emerald-600 text-white font-medium rounded-xl transition-colors">
              Подтвердить
            </button>
          </div>
        </div>
      </div>

      {/* 2FA Settings */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Shield className="w-5 h-5 text-amber-400" />
          Настройки 2FA
        </h3>
        <div className="space-y-4">
          {[
            { key: 'telegramRequisites', label: 'Разрешить просмотр реквизитов в Telegram', desc: 'Быстрый доступ к данным карт через бота' },
            { key: 'twoFactorLogin', label: '2FA на вход', desc: 'Подтверждение при каждом входе в аккаунт' },
            { key: 'twoFactorRequisites', label: '2FA для просмотра реквизитов', desc: 'Дополнительная защита данных карт' },
            { key: 'twoFactorWithdraw', label: '2FA для вывода денег', desc: 'Подтверждение вывода средств' },
            { key: 'twoFactorTransfer', label: '2FA для перевода другому пользователю', desc: 'Защита переводов между аккаунтами' },
          ].map(item => (
            <div key={item.key} className="flex items-center justify-between p-4 rounded-xl bg-white/[0.03] border border-white/5">
              <div>
                <p className="text-white font-medium">{item.label}</p>
                <p className="text-sm text-slate-500">{item.desc}</p>
              </div>
              <Toggle
                checked={settings[item.key as keyof typeof settings]}
                onChange={(v) => setSettings({ ...settings, [item.key]: v })}
              />
            </div>
          ))}
        </div>
      </div>

      {/* Activity Log */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Clock className="w-5 h-5 text-blue-400" />
          Последняя активность
        </h3>
        <div className="overflow-x-auto">
          <table className="xplr-table min-w-[500px]">
            <thead>
              <tr>
                <th>Дата/Время</th>
                <th>IP</th>
                <th>Локация</th>
                <th>Браузер</th>
              </tr>
            </thead>
            <tbody>
              {activityLog.map((log, i) => (
                <tr key={i}>
                  <td className="text-slate-300">{log.date}</td>
                  <td className="font-mono text-slate-400">{log.ip}</td>
                  <td className="text-slate-400">{log.location}</td>
                  <td className="text-slate-400">{log.browser}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

// KYC Tab
const KYCTab = () => {
  const [step, setStep] = useState(1);
  const [country, setCountry] = useState('');

  const steps = [
    { id: 1, title: 'Гражданство', status: 'current' },
    { id: 2, title: 'Личные данные', status: 'pending' },
    { id: 3, title: 'Документы', status: 'pending' },
  ];

  return (
    <div className="space-y-6">
      {/* Progress Steps */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-6">Верификация аккаунта</h3>
        <div className="flex items-center justify-between mb-8">
          {steps.map((s, i) => (
            <div key={s.id} className="flex items-center">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${
                s.status === 'completed' ? 'bg-emerald-500 text-white' :
                s.status === 'current' ? 'bg-blue-500 text-white' :
                'bg-white/10 text-slate-500'
              }`}>
                {s.status === 'completed' ? <Check className="w-5 h-5" /> : s.id}
              </div>
              <span className={`ml-3 ${s.status === 'current' ? 'text-white' : 'text-slate-500'}`}>
                {s.title}
              </span>
              {i < steps.length - 1 && (
                <div className="w-16 md:w-24 h-0.5 bg-white/10 mx-4" />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Step 1: Citizenship */}
      {step === 1 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Выберите гражданство</h3>
          <select
            value={country}
            onChange={(e) => setCountry(e.target.value)}
            className="xplr-select w-full mb-4"
          >
            <option value="">Выберите страну</option>
            <option value="RU">🇷🇺 Россия</option>
            <option value="BY">🇧🇾 Беларусь</option>
            <option value="KZ">🇰🇿 Казахстан</option>
            <option value="UA">🇺🇦 Украина</option>
            <option value="US">🇺🇸 США</option>
            <option value="DE">🇩🇪 Германия</option>
          </select>
          <button
            onClick={() => country && setStep(2)}
            disabled={!country}
            className="w-full px-4 py-3 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium rounded-xl transition-colors"
          >
            Продолжить
          </button>
        </div>
      )}

      {/* Step 2: Personal Data */}
      {step === 2 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Личные данные</h3>
          <div className="grid gap-4 mb-6">
            <div className="grid md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-slate-400 mb-2">Имя</label>
                <input type="text" className="xplr-input w-full" placeholder="Иван" />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-2">Фамилия</label>
                <input type="text" className="xplr-input w-full" placeholder="Иванов" />
              </div>
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-2">Дата рождения</label>
              <input type="date" className="xplr-input w-full" />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-2">Адрес проживания</label>
              <input type="text" className="xplr-input w-full" placeholder="Город, улица, дом" />
            </div>
          </div>
          <div className="flex gap-3">
            <button
              onClick={() => setStep(1)}
              className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl transition-colors"
            >
              Назад
            </button>
            <button
              onClick={() => setStep(3)}
              className="flex-1 px-4 py-3 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-xl transition-colors"
            >
              Продолжить
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Documents */}
      {step === 3 && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Загрузка документов</h3>
          <div className="space-y-4 mb-6">
            {[
              { id: 'passport', label: 'Гос. паспорт', desc: 'Фото первого разворота' },
              { id: 'address', label: 'Подтверждение адреса', desc: 'Квитанция или выписка' },
              { id: 'selfie', label: 'Селфи с документом', desc: 'Держите паспорт рядом с лицом' },
            ].map((doc) => (
              <div key={doc.id} className="p-4 rounded-xl bg-white/[0.03] border border-white/10 border-dashed">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Upload className="w-5 h-5 text-slate-400" />
                    <div>
                      <p className="text-white font-medium">{doc.label}</p>
                      <p className="text-sm text-slate-500">{doc.desc}</p>
                    </div>
                  </div>
                  <button className="px-4 py-2 glass-card hover:bg-white/10 text-slate-300 text-sm rounded-lg transition-colors">
                    Загрузить
                  </button>
                </div>
              </div>
            ))}
          </div>
          <div className="flex gap-3">
            <button
              onClick={() => setStep(2)}
              className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl transition-colors"
            >
              Назад
            </button>
            <button className="flex-1 px-4 py-3 bg-emerald-500 hover:bg-emerald-600 text-white font-medium rounded-xl transition-colors">
              Отправить на проверку
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

// Notifications Tab
const NotificationsTab = () => {
  const [channels, setChannels] = useState({
    email: true,
    telegram: true,
    push: false
  });

  const [events, setEvents] = useState({
    login: true,
    ticket: true,
    restricted: true,
    teamDecisions: false,
    joinRequest: true,
    cardBlock: true,
    cardOperations: true,
    codes3ds: true,
    topupSuccess: true,
    topupError: true,
    topupLow: true
  });

  return (
    <div className="space-y-6">
      {/* Channels */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Каналы уведомлений</h3>
        <div className="grid gap-4">
          {[
            { key: 'email', label: 'Почта', icon: Mail },
            { key: 'telegram', label: 'Telegram', icon: MessageCircle },
            { key: 'push', label: 'Push-уведомления', icon: Bell },
          ].map((channel) => (
            <div key={channel.key} className="flex items-center justify-between p-4 rounded-xl bg-white/[0.03] border border-white/5">
              <div className="flex items-center gap-3">
                <channel.icon className="w-5 h-5 text-blue-400" />
                <span className="text-white font-medium">{channel.label}</span>
              </div>
              <Toggle
                checked={channels[channel.key as keyof typeof channels]}
                onChange={(v) => setChannels({ ...channels, [channel.key]: v })}
              />
            </div>
          ))}
        </div>
      </div>

      {/* Event Categories */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Системные</h3>
        <div className="space-y-3">
          {[
            { key: 'login', label: 'Вход в аккаунт' },
            { key: 'ticket', label: 'Ответ на тикет' },
            { key: 'restricted', label: 'Ограниченные операции' },
          ].map((event) => (
            <div key={event.key} className="flex items-center justify-between py-2">
              <span className="text-slate-300">{event.label}</span>
              <Toggle
                checked={events[event.key as keyof typeof events]}
                onChange={(v) => setEvents({ ...events, [event.key]: v })}
              />
            </div>
          ))}
        </div>
      </div>

      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Командные</h3>
        <div className="space-y-3">
          {[
            { key: 'teamDecisions', label: 'Решения владельца' },
            { key: 'joinRequest', label: 'Запрос на вступление' },
          ].map((event) => (
            <div key={event.key} className="flex items-center justify-between py-2">
              <span className="text-slate-300">{event.label}</span>
              <Toggle
                checked={events[event.key as keyof typeof events]}
                onChange={(v) => setEvents({ ...events, [event.key]: v })}
              />
            </div>
          ))}
        </div>
      </div>

      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Карты</h3>
        <div className="space-y-3">
          {[
            { key: 'cardBlock', label: 'Блокировка карты' },
            { key: 'cardOperations', label: 'Операции по карте' },
            { key: 'codes3ds', label: '3DS коды' },
          ].map((event) => (
            <div key={event.key} className="flex items-center justify-between py-2">
              <span className="text-slate-300">{event.label}</span>
              <Toggle
                checked={events[event.key as keyof typeof events]}
                onChange={(v) => setEvents({ ...events, [event.key]: v })}
              />
            </div>
          ))}
        </div>
      </div>

      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Автопополнение</h3>
        <div className="space-y-3">
          {[
            { key: 'topupSuccess', label: 'Успешное пополнение' },
            { key: 'topupError', label: 'Ошибка пополнения' },
            { key: 'topupLow', label: 'Недостаток средств' },
          ].map((event) => (
            <div key={event.key} className="flex items-center justify-between py-2">
              <span className="text-slate-300">{event.label}</span>
              <Toggle
                checked={events[event.key as keyof typeof events]}
                onChange={(v) => setEvents({ ...events, [event.key]: v })}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// Language Tab
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

  const handleChange = (code: string) => {
    i18n.changeLanguage(code);
    localStorage.setItem(LANG_KEY, code);
  };

  return (
    <div className="glass-card p-6">
      <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
        <Globe className="w-5 h-5 text-blue-400" />
        {t('settings.languageTitle')}
      </h3>
      <div className="space-y-3">
        {languages.map((lang) => (
          <button
            key={lang.code}
            onClick={() => handleChange(lang.code)}
            className={`w-full flex items-center justify-between p-4 rounded-xl transition-all duration-150 ${
              currentLang === lang.code
                ? 'bg-blue-500/20 border border-blue-500/50'
                : 'bg-white/[0.03] border border-white/5 hover:border-white/10'
            }`}
          >
            <div className="flex items-center gap-3">
              <span className="text-2xl">{lang.flag}</span>
              <span className="text-white font-medium">{lang.name}</span>
            </div>
            {currentLang === lang.code && (
              <CheckCircle className="w-5 h-5 text-blue-400" />
            )}
          </button>
        ))}
      </div>
    </div>
  );
};

export const SettingsPage = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<SettingsTab>('security');
  const [saved, setSaved] = useState(false);
  
  const tabs = [
    { id: 'security', label: t('settings.security'), icon: Shield },
    { id: 'kyc', label: t('settings.kyc'), icon: FileText },
    { id: 'notifications', label: t('settings.notifications'), icon: Bell },
    { id: 'language', label: t('settings.language'), icon: Globe },
  ];

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <DashboardLayout>
      <div className="stagger-fade-in max-w-4xl">
        <BackButton />
        
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">{t('settings.title')}</h1>
          <p className="text-slate-400">{t('settings.subtitle')}</p>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-8 overflow-x-auto pb-2">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as SettingsTab)}
              className={`flex items-center gap-2 px-4 py-3 rounded-xl font-medium transition-all whitespace-nowrap ${
                activeTab === tab.id
                  ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/25'
                  : 'glass-card text-slate-400 hover:text-white'
              }`}
            >
              <tab.icon className="w-4 h-4" />
              {tab.label}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        {activeTab === 'security' && <SecurityTab />}
        {activeTab === 'kyc' && <KYCTab />}
        {activeTab === 'notifications' && <NotificationsTab />}
        {activeTab === 'language' && <LanguageTab />}

        {/* Save Button */}
        <button 
          onClick={handleSave}
          className={`w-full mt-8 py-4 rounded-xl font-semibold text-lg transition-all flex items-center justify-center gap-2 ${
            saved 
              ? 'bg-emerald-500 text-white' 
              : 'gradient-accent text-white hover:shadow-lg hover:shadow-blue-500/25'
          }`}
        >
          {saved ? <Check className="w-5 h-5" /> : <Save className="w-5 h-5" />}
          {saved ? t('settings.saved') : t('settings.save')}
        </button>
      </div>
    </DashboardLayout>
  );
};
