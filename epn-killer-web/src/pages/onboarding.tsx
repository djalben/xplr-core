import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../store/auth-context';
import { NeuralBackground } from '../components/neural-background';
import { User, Users, CreditCard, MessageCircle, BarChart3, Zap, Shield, ArrowRight } from 'lucide-react';

export const OnboardingPage = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { completeOnboarding } = useAuth();

  const handleSelect = (mode: 'personal' | 'business') => {
    completeOnboarding(mode);
    navigate('/dashboard');
  };

  return (
    <div className="min-h-[100dvh] bg-[#050507] flex flex-col items-center justify-center p-4 overflow-hidden relative">
      <div className="absolute inset-0 bg-gradient-to-br from-[#05050a] via-[#080812] to-[#04040a]" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(30,58,138,0.12)_0%,transparent_60%)]" />
      <NeuralBackground reducedDensity />

      <div className="relative z-10 w-full max-w-3xl">
        {/* Logo */}
        <div className="flex justify-center mb-6">
          <div className="flex items-center gap-3">
            <div className="relative w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center overflow-hidden shadow-lg shadow-blue-500/20">
              <div className="absolute inset-0 bg-gradient-to-tr from-transparent via-white/20 to-transparent" />
              <span className="text-white font-bold text-2xl tracking-tighter relative z-10">X</span>
            </div>
            <span className="text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-400 bg-clip-text text-transparent">XPLR</span>
          </div>
        </div>

        <h1 className="text-2xl md:text-3xl font-bold text-white text-center mb-2">{t('onboarding.title')}</h1>
        <p className="text-slate-500 text-center mb-10 text-sm md:text-base">{t('onboarding.subtitle')}</p>

        {/* Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {/* SOLO */}
          <button
            onClick={() => handleSelect('personal')}
            className="group relative rounded-2xl p-6 text-left transition-all duration-150
              bg-gradient-to-br from-[#0d1528]/90 via-[#0a1025]/95 to-[#0c0f20]/90
              backdrop-blur-2xl border border-white/[0.08]
              hover:border-blue-500/40 hover:shadow-[0_0_40px_-8px_rgba(59,130,246,0.2)]"
          >
            <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-blue-400/20 to-transparent" />

            <div className="flex items-center gap-3 mb-5">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-indigo-500/20 border border-blue-500/20 flex items-center justify-center">
                <User className="w-6 h-6 text-blue-400" />
              </div>
              <div>
                <h2 className="text-lg font-bold text-white">{t('onboarding.solo')}</h2>
                <p className="text-[11px] text-slate-500 uppercase tracking-wider">{t('onboarding.soloLabel')}</p>
              </div>
            </div>

            <p className="text-sm text-slate-400 leading-relaxed mb-6">
              {t('onboarding.soloDesc')}
            </p>

            <div className="space-y-2.5 mb-6">
              {[
                { icon: <CreditCard className="w-4 h-4 text-blue-400" />, text: t('onboarding.soloFeature1') },
                { icon: <Zap className="w-4 h-4 text-amber-400" />, text: t('onboarding.soloFeature2') },
                { icon: <Shield className="w-4 h-4 text-emerald-400" />, text: t('onboarding.soloFeature3') },
              ].map((f, i) => (
                <div key={i} className="flex items-center gap-2.5">
                  {f.icon}
                  <span className="text-xs text-slate-300">{f.text}</span>
                </div>
              ))}
            </div>

            <div className="flex items-center gap-2 text-blue-400 text-sm font-semibold group-hover:gap-3 transition-all duration-150">
              {t('onboarding.select')} <ArrowRight className="w-4 h-4" />
            </div>
          </button>

          {/* TEAM PRIME */}
          <button
            onClick={() => handleSelect('business')}
            className="group relative rounded-2xl p-6 text-left transition-all duration-150
              bg-gradient-to-br from-[#0d1528]/90 via-[#0a1025]/95 to-[#0c0f20]/90
              backdrop-blur-2xl border border-white/[0.08]
              hover:border-red-500/40 hover:shadow-[0_0_40px_-8px_rgba(239,68,68,0.2)]"
          >
            <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-red-400/20 to-transparent" />

            <div className="flex items-center gap-3 mb-5">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-red-500/20 to-orange-500/20 border border-red-500/20 flex items-center justify-center">
                <Users className="w-6 h-6 text-red-400" />
              </div>
              <div>
                <h2 className="text-lg font-bold text-white">{t('onboarding.teamPrime')}</h2>
                <p className="text-[11px] text-slate-500 uppercase tracking-wider">{t('onboarding.teamLabel')}</p>
              </div>
            </div>

            <p className="text-sm text-slate-400 leading-relaxed mb-6">
              {t('onboarding.teamDesc')}
            </p>

            <div className="space-y-2.5 mb-6">
              {[
                { icon: <Users className="w-4 h-4 text-red-400" />, text: t('onboarding.teamFeature1') },
                { icon: <MessageCircle className="w-4 h-4 text-orange-400" />, text: t('onboarding.teamFeature2') },
                { icon: <BarChart3 className="w-4 h-4 text-purple-400" />, text: t('onboarding.teamFeature3') },
              ].map((f, i) => (
                <div key={i} className="flex items-center gap-2.5">
                  {f.icon}
                  <span className="text-xs text-slate-300">{f.text}</span>
                </div>
              ))}
            </div>

            <div className="flex items-center gap-2 text-red-400 text-sm font-semibold group-hover:gap-3 transition-all duration-150">
              {t('onboarding.select')} <ArrowRight className="w-4 h-4" />
            </div>
          </button>
        </div>
      </div>
    </div>
  );
};
