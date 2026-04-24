import { useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Eye, EyeOff, Loader2, ToggleLeft, ToggleRight, Lock, Save } from 'lucide-react';

type SystemSettingRow = {
  key: string;
  value: string;
  description: string;
  boolValue?: boolean | null;
  updatedAt?: string;
};

type VPNInfraDraft = {
  ip: string;
  cpu: string;
  ramMb: string;
  diskGb: string;
  diskType: string;
  expiresAt: string;
  costEur: string;
  limitGb: string;
  location: string;
  os: string;
};

const VPN_KEYS = {
  ip: 'vpn_server_ip',
  cpu: 'vpn_server_cpu',
  ramMb: 'vpn_server_ram_mb',
  diskGb: 'vpn_server_disk_gb',
  diskType: 'vpn_server_disk_type',
  expiresAt: 'vpn_server_expires_at',
  costEur: 'vpn_server_cost_eur',
  limitGb: 'vpn_server_limit_gb',
  location: 'vpn_server_location',
  os: 'vpn_server_os',
} as const;

export const AdminSystemSettingsPage = () => {
  const [sbpEnabled, setSbpEnabled] = useState<boolean | null>(null);
  const [sbpLoading, setSbpLoading] = useState(false);
  const [sbpError, setSbpError] = useState('');

  const [pin, setPin] = useState('');
  const [pinSaving, setPinSaving] = useState(false);
  const [pinOk, setPinOk] = useState(false);
  const [pinError, setPinError] = useState('');
  const [pinVisible, setPinVisible] = useState(false);

  const [vpnLoading, setVpnLoading] = useState(false);
  const [vpnSaving, setVpnSaving] = useState(false);
  const [vpnOk, setVpnOk] = useState(false);
  const [vpnError, setVpnError] = useState('');
  const [vpnDraft, setVpnDraft] = useState<VPNInfraDraft>({
    ip: '',
    cpu: '',
    ramMb: '',
    diskGb: '',
    diskType: '',
    expiresAt: '',
    costEur: '',
    limitGb: '',
    location: '',
    os: '',
  });

  const sbpLabel = useMemo(() => {
    if (sbpEnabled === null) return 'загрузка...';
    return sbpEnabled ? 'включено' : 'выключено';
  }, [sbpEnabled]);

  const loadSBP = async () => {
    setSbpError('');
    try {
      const res = await apiClient.get<{ enabled: boolean }>('/admin/sbp-topup');
      setSbpEnabled(Boolean(res.data.enabled));
    } catch {
      setSbpEnabled(null);
      setSbpError('Не удалось загрузить статус СБП');
    }
  };

  useEffect(() => {
    loadSBP();
    loadVPNInfra();
  }, []);

  const loadVPNInfra = async () => {
    setVpnLoading(true);
    setVpnError('');
    try {
      const res = await apiClient.get<SystemSettingRow[]>('/admin/system-settings');
      const rows = res.data || [];
      const m = new Map(rows.map((r) => [r.key, String(r.value ?? '')]));

      setVpnDraft({
        ip: m.get(VPN_KEYS.ip) || '',
        cpu: m.get(VPN_KEYS.cpu) || '',
        ramMb: m.get(VPN_KEYS.ramMb) || '',
        diskGb: m.get(VPN_KEYS.diskGb) || '',
        diskType: m.get(VPN_KEYS.diskType) || '',
        expiresAt: m.get(VPN_KEYS.expiresAt) || '',
        costEur: m.get(VPN_KEYS.costEur) || '',
        limitGb: m.get(VPN_KEYS.limitGb) || '',
        location: m.get(VPN_KEYS.location) || '',
        os: m.get(VPN_KEYS.os) || '',
      });
    } catch {
      setVpnError('Не удалось загрузить настройки VPN');
    } finally {
      setVpnLoading(false);
    }
  };

  const saveVPNInfra = async () => {
    setVpnSaving(true);
    setVpnOk(false);
    setVpnError('');
    try {
      const pairs: Array<[string, string]> = [
        [VPN_KEYS.ip, vpnDraft.ip],
        [VPN_KEYS.cpu, vpnDraft.cpu],
        [VPN_KEYS.ramMb, vpnDraft.ramMb],
        [VPN_KEYS.diskGb, vpnDraft.diskGb],
        [VPN_KEYS.diskType, vpnDraft.diskType],
        [VPN_KEYS.expiresAt, vpnDraft.expiresAt],
        [VPN_KEYS.costEur, vpnDraft.costEur],
        [VPN_KEYS.limitGb, vpnDraft.limitGb],
        [VPN_KEYS.location, vpnDraft.location],
        [VPN_KEYS.os, vpnDraft.os],
      ];

      // последовательно, чтобы не DDOS'ить admin ручки и проще отлавливать ошибки
      for (const [key, value] of pairs) {
        // eslint-disable-next-line no-await-in-loop
        await apiClient.patch(`/admin/system-settings/${encodeURIComponent(key)}`, { value: String(value ?? '') });
      }

      setVpnOk(true);
      window.setTimeout(() => setVpnOk(false), 2500);
    } catch {
      setVpnError('Не удалось сохранить настройки VPN');
    } finally {
      setVpnSaving(false);
    }
  };

  const toggleSBP = async () => {
    if (sbpEnabled === null) return;
    const next = !sbpEnabled;
    setSbpLoading(true);
    setSbpError('');
    setSbpEnabled(next);
    try {
      const res = await apiClient.patch<{ enabled: boolean }>('/admin/sbp-topup', { enabled: next });
      setSbpEnabled(Boolean(res.data.enabled));
    } catch {
      setSbpEnabled(!next);
      setSbpError('Не удалось переключить СБП');
    } finally {
      setSbpLoading(false);
    }
  };

  const savePIN = async () => {
    setPinSaving(true);
    setPinOk(false);
    setPinError('');
    try {
      await apiClient.patch('/admin/staff-pin', { pin });
      setPin('');
      setPinVisible(false);
      setPinOk(true);
      window.setTimeout(() => setPinOk(false), 2500);
    } catch {
      setPinError('Не удалось изменить PIN');
    } finally {
      setPinSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Системные настройки</h1>
          <p className="text-sm text-slate-400 mt-1">PIN админки и глобальные переключатели</p>
        </div>
      </div>

      {sbpError ? <p className="text-sm text-red-400">{sbpError}</p> : null}
      {pinError ? <p className="text-sm text-red-400">{pinError}</p> : null}
      {vpnError ? <p className="text-sm text-red-400">{vpnError}</p> : null}

      <div className="glass-card p-6">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <p className="text-white font-semibold">Пополнение через СБП</p>
            <p className="text-sm text-slate-400 mt-1">
              Глобальное включение/выключение пополнения. Сейчас: <span className="text-slate-200">{sbpLabel}</span>
            </p>
          </div>

          <button
            onClick={toggleSBP}
            disabled={sbpEnabled === null || sbpLoading}
            className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {sbpEnabled ? <ToggleRight className="w-5 h-5 text-emerald-400" /> : <ToggleLeft className="w-5 h-5 text-slate-400" />}
            {sbpLoading ? '...' : sbpEnabled ? 'Выключить' : 'Включить'}
          </button>
        </div>
      </div>

      <div className="glass-card p-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500 to-purple-600 flex items-center justify-center">
            <Lock className="w-5 h-5 text-white" />
          </div>
          <div>
            <p className="text-white font-semibold">PIN админки</p>
            <p className="text-sm text-slate-400">Задать новый PIN для входа (4 цифры)</p>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
          <div className="relative w-full sm:max-w-[240px]">
            <input
              value={pin}
              onChange={(e) => { setPin(e.target.value.replace(/[^\d]/g, '').slice(0, 4)); setPinError(''); }}
              inputMode="numeric"
              pattern="[0-9]*"
              type={pinVisible ? 'text' : 'password'}
              maxLength={4}
              placeholder="0000"
              className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 pr-12 text-white text-center text-lg tracking-[0.3em] font-mono outline-none focus:border-blue-500/40"
            />
            <button
              type="button"
              onClick={() => setPinVisible((v) => !v)}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-white/5 transition-colors"
              title={pinVisible ? 'Скрыть PIN' : 'Показать PIN'}
            >
              {pinVisible ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
          <button
            onClick={savePIN}
            disabled={pinSaving || pin.length !== 4}
            className="px-5 py-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
          >
            {pinSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            {pinSaving ? '...' : 'Сохранить'}
          </button>
          {pinOk ? <span className="text-sm text-emerald-400">Сохранено</span> : null}
        </div>
      </div>

      <div className="glass-card p-6">
        <div className="flex items-start justify-between gap-4 flex-wrap mb-4">
          <div>
            <p className="text-white font-semibold">VPN сервер (инфраструктура)</p>
            <p className="text-sm text-slate-400 mt-1">Оверрайд данных для карточки VPN‑сервера (без Aeza API)</p>
          </div>
          <div className="flex items-center gap-2">
            {vpnOk ? <span className="text-sm text-emerald-400">Сохранено</span> : null}
            <button
              onClick={loadVPNInfra}
              disabled={vpnLoading || vpnSaving}
              className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 inline-flex items-center gap-2"
            >
              {vpnLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
              Обновить
            </button>
            <button
              onClick={saveVPNInfra}
              disabled={vpnSaving || vpnLoading}
              className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-sm font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
            >
              {vpnSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
              Сохранить
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="text-xs text-slate-500">IP</label>
            <input
              value={vpnDraft.ip}
              onChange={(e) => setVpnDraft((p) => ({ ...p, ip: e.target.value }))}
              placeholder="109.120.157.144"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Оплачен до (RFC3339)</label>
            <input
              value={vpnDraft.expiresAt}
              onChange={(e) => setVpnDraft((p) => ({ ...p, expiresAt: e.target.value }))}
              placeholder="2026-05-18T00:00:00Z"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">CPU (vCPU)</label>
            <input
              value={vpnDraft.cpu}
              onChange={(e) => setVpnDraft((p) => ({ ...p, cpu: e.target.value.replace(/[^\d]/g, '') }))}
              placeholder="1"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">RAM (MB)</label>
            <input
              value={vpnDraft.ramMb}
              onChange={(e) => setVpnDraft((p) => ({ ...p, ramMb: e.target.value.replace(/[^\d]/g, '') }))}
              placeholder="2048"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Диск (GB)</label>
            <input
              value={vpnDraft.diskGb}
              onChange={(e) => setVpnDraft((p) => ({ ...p, diskGb: e.target.value.replace(/[^\d]/g, '') }))}
              placeholder="30"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Тип диска</label>
            <input
              value={vpnDraft.diskType}
              onChange={(e) => setVpnDraft((p) => ({ ...p, diskType: e.target.value }))}
              placeholder="SSD"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Стоимость (EUR/мес)</label>
            <input
              value={vpnDraft.costEur}
              onChange={(e) => setVpnDraft((p) => ({ ...p, costEur: e.target.value.replace(/[^0-9.]/g, '') }))}
              placeholder="4.94"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Лимит трафика (GB)</label>
            <input
              value={vpnDraft.limitGb}
              onChange={(e) => setVpnDraft((p) => ({ ...p, limitGb: e.target.value.replace(/[^0-9.]/g, '') }))}
              placeholder="30"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Локация</label>
            <input
              value={vpnDraft.location}
              onChange={(e) => setVpnDraft((p) => ({ ...p, location: e.target.value }))}
              placeholder="Stockholm"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">OS</label>
            <input
              value={vpnDraft.os}
              onChange={(e) => setVpnDraft((p) => ({ ...p, os: e.target.value }))}
              placeholder="Ubuntu"
              className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
            />
          </div>
        </div>
      </div>
    </div>
  );
};

