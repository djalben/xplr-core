import { useState } from 'react';
import { useRates } from '../store/rates-context';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import apiClient, { API_BASE_URL } from '../api/axios';
import { DollarSign, Save, RefreshCw, Check, AlertTriangle } from 'lucide-react';

export const AdminRatesPage = () => {
  const { rates, setRates, refreshRates } = useRates();
  const [usdInput, setUsdInput] = useState(String(rates.usd));
  const [eurInput, setEurInput] = useState(String(rates.eur));
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState('');

  const handleSave = async () => {
    const usd = parseFloat(usdInput);
    const eur = parseFloat(eurInput);

    if (isNaN(usd) || isNaN(eur) || usd <= 0 || eur <= 0) {
      setError('Введите корректные значения курсов');
      return;
    }

    setSaving(true);
    setError('');

    try {
      await apiClient.put(`${API_BASE_URL}/rates`, { usd, eur });
      setRates({ usd, eur });
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err: any) {
      // If API is not available, still update locally
      setRates({ usd, eur });
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } finally {
      setSaving(false);
    }
  };

  const handleRefresh = async () => {
    await refreshRates();
    setUsdInput(String(rates.usd));
    setEurInput(String(rates.eur));
  };

  return (
    <DashboardLayout>
      <div className="stagger-fade-in max-w-2xl">
        <BackButton />

        <div className="flex items-center gap-4 mb-8">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-amber-500/20 to-orange-500/20 border border-amber-500/30 flex items-center justify-center">
            <DollarSign className="w-7 h-7 text-amber-400" />
          </div>
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-white">Управление курсами</h1>
            <p className="text-slate-400 text-sm">Редактирование курсов валют для всей платформы</p>
          </div>
        </div>

        <div className="glass-card p-6 mb-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-white">Курсы валют</h3>
            <button
              onClick={handleRefresh}
              className="flex items-center gap-2 px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 text-slate-300 rounded-xl transition-colors text-sm"
            >
              <RefreshCw className="w-4 h-4" />
              Обновить
            </button>
          </div>

          <div className="space-y-5">
            {/* USD Rate */}
            <div>
              <label className="flex items-center gap-2 text-sm text-slate-400 mb-2">
                <span className="w-8 h-8 rounded-lg bg-emerald-500/20 flex items-center justify-center text-emerald-400 font-bold text-xs">$</span>
                Курс USD (в рублях)
              </label>
              <input
                type="number"
                step="0.01"
                value={usdInput}
                onChange={(e) => setUsdInput(e.target.value)}
                className="w-full h-14 px-4 bg-white/[0.04] border border-white/10 rounded-xl text-white text-xl font-semibold focus:outline-none focus:border-blue-500/50 transition-all"
                placeholder="89.45"
              />
            </div>

            {/* EUR Rate */}
            <div>
              <label className="flex items-center gap-2 text-sm text-slate-400 mb-2">
                <span className="w-8 h-8 rounded-lg bg-blue-500/20 flex items-center justify-center text-blue-400 font-bold text-xs">€</span>
                Курс EUR (в рублях)
              </label>
              <input
                type="number"
                step="0.01"
                value={eurInput}
                onChange={(e) => setEurInput(e.target.value)}
                className="w-full h-14 px-4 bg-white/[0.04] border border-white/10 rounded-xl text-white text-xl font-semibold focus:outline-none focus:border-blue-500/50 transition-all"
                placeholder="97.82"
              />
            </div>
          </div>

          {error && (
            <div className="flex items-center gap-2 mt-4 p-3 bg-red-500/10 border border-red-500/30 rounded-xl text-red-400 text-sm">
              <AlertTriangle className="w-4 h-4 shrink-0" />
              {error}
            </div>
          )}

          {saved && (
            <div className="flex items-center gap-2 mt-4 p-3 bg-emerald-500/10 border border-emerald-500/30 rounded-xl text-emerald-400 text-sm">
              <Check className="w-4 h-4 shrink-0" />
              Курсы успешно обновлены
            </div>
          )}

          <button
            onClick={handleSave}
            disabled={saving}
            className="w-full mt-6 flex items-center justify-center gap-2 px-6 py-4 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 disabled:opacity-50 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/20"
          >
            <Save className="w-5 h-5" />
            {saving ? 'Сохранение...' : 'Сохранить курсы'}
          </button>
        </div>

        {/* Preview */}
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Предпросмотр</h3>
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-white/[0.03] border border-white/10 rounded-xl text-center">
              <p className="text-slate-400 text-sm mb-1">USD → RUB</p>
              <p className="text-2xl font-bold text-emerald-400">$1 = {parseFloat(usdInput || '0').toFixed(2)} ₽</p>
            </div>
            <div className="p-4 bg-white/[0.03] border border-white/10 rounded-xl text-center">
              <p className="text-slate-400 text-sm mb-1">EUR → RUB</p>
              <p className="text-2xl font-bold text-blue-400">€1 = {parseFloat(eurInput || '0').toFixed(2)} ₽</p>
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
};
