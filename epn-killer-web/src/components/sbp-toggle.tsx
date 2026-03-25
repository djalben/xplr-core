import { useState, useEffect } from 'react';
import { getSystemSettings, updateSystemSetting } from '../api/system-settings';

export const SBPToggle = () => {
  const [sbpEnabled, setSbpEnabled] = useState(true);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const settings = await getSystemSettings();
        const sbpSetting = settings.find(s => s.key === 'sbp_enabled');
        if (sbpSetting) {
          setSbpEnabled(sbpSetting.value === 'true');
        }
      } catch (error) {
        console.error('Failed to fetch SBP settings:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchSettings();
  }, []);

  const handleToggle = async () => {
    setUpdating(true);
    try {
      const newValue = !sbpEnabled;
      await updateSystemSetting('sbp_enabled', newValue ? 'true' : 'false');
      setSbpEnabled(newValue);
    } catch (error) {
      console.error('Failed to update SBP setting:', error);
    } finally {
      setUpdating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-between p-4 bg-white/5 rounded-xl border border-white/10 animate-pulse">
        <div className="h-5 bg-white/10 rounded w-32" />
        <div className="h-6 bg-white/10 rounded w-12" />
      </div>
    );
  }

  return (
    <div className="flex items-center justify-between p-4 bg-white/5 rounded-xl border border-white/10">
      <div>
        <h4 className="text-white font-medium text-sm mb-1">Пополнение через СБП</h4>
        <p className="text-xs text-slate-400">
          {sbpEnabled ? 'Пользователи могут пополнять кошелек через СБП' : 'Пополнение через СБП отключено'}
        </p>
      </div>
      
      <button
        onClick={handleToggle}
        disabled={updating}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-[#0b0b14] disabled:opacity-50 ${
          sbpEnabled ? 'bg-emerald-500' : 'bg-slate-600'
        }`}
      >
        <span
          className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
            sbpEnabled ? 'translate-x-6' : 'translate-x-1'
          }`}
        />
      </button>
    </div>
  );
};
