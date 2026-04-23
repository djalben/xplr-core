import { useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../store/auth-context';
import { DashboardLayout } from '../../components/dashboard-layout';
import { Shield, Loader2 } from 'lucide-react';

type GrantMessage = {
  type: 'xplr_admin_grant';
};

export const AdminEntryPage = () => {
  const navigate = useNavigate();
  const { isAdmin, authReady } = useAuth();

  const origin = useMemo(() => window.location.origin, []);

  useEffect(() => {
    if (!authReady) return;
    if (!isAdmin) {
      navigate('/dashboard', { replace: true });
      return;
    }

    const onMessage = (event: MessageEvent) => {
      if (event.origin !== origin) return;
      if (event.source !== window.opener) return;

      const data = event.data as Partial<GrantMessage> | null;
      if (!data || data.type !== 'xplr_admin_grant') return;

      sessionStorage.setItem('_xplr_staff', 'granted');
      navigate('/admin/dashboard', { replace: true });
    };

    window.addEventListener('message', onMessage);
    return () => window.removeEventListener('message', onMessage);
  }, [authReady, isAdmin, navigate, origin]);

  return (
    <DashboardLayout>
      <div className="stagger-fade-in max-w-xl">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center">
            <Shield className="w-7 h-7 text-blue-300" />
          </div>
          <div className="min-w-0">
            <h1 className="text-2xl md:text-3xl font-bold text-white">Админка</h1>
            <p className="text-slate-400 text-sm">Ожидание подтверждения доступа…</p>
          </div>
        </div>

        <div className="glass-card p-6 flex items-center gap-3 text-slate-300">
          <Loader2 className="w-5 h-5 animate-spin text-blue-400" />
          <span className="text-sm">Пожалуйста, не закрывайте вкладку — выполняется вход.</span>
        </div>
      </div>
    </DashboardLayout>
  );
};

