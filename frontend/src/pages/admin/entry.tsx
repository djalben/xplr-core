import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../store/auth-context';
import { AdminLayout } from '../../components/admin-layout';
import { Shield, Loader2 } from 'lucide-react';

type GrantMessage = {
  type: 'xplr_admin_grant';
};

type ReadyMessage = {
  type: 'xplr_admin_ready';
};

export const AdminEntryPage = () => {
  const navigate = useNavigate();
  const { isAdmin, authReady } = useAuth();
  const [granted, setGranted] = useState(false);
  const grantedOnce = useRef(false);

  const origin = useMemo(() => window.location.origin, []);

  useEffect(() => {
    const grant = () => {
      if (grantedOnce.current) return;
      grantedOnce.current = true;
      sessionStorage.setItem('_xplr_staff', 'granted');
      setGranted(true);
    };

    const onMessage = (event: MessageEvent) => {
      if (event.origin !== origin) return;
      const data = event.data as Partial<GrantMessage> | null;
      if (!data || data.type !== 'xplr_admin_grant') return;
      grant();
    };

    const bc = new BroadcastChannel('xplr_admin');
    const onBC = (event: MessageEvent) => {
      const data = event.data as Partial<GrantMessage> | null;
      if (!data || data.type !== 'xplr_admin_grant') return;
      grant();
    };

    bc.addEventListener('message', onBC);
    window.addEventListener('message', onMessage);

    // Handshake: tell the opener we're ready to receive the grant.
    try {
      bc.postMessage({ type: 'xplr_admin_ready' satisfies ReadyMessage['type'] } as ReadyMessage);
    } catch {
      // ignore
    }

    return () => {
      bc.removeEventListener('message', onBC);
      bc.close();
      window.removeEventListener('message', onMessage);
    };
  }, [origin]);

  useEffect(() => {
    if (!authReady) return;
    if (!isAdmin) {
      navigate('/dashboard', { replace: true });
      return;
    }
    if (!granted) return;
    navigate('/admin/dashboard', { replace: true });
  }, [authReady, granted, isAdmin, navigate]);

  return (
    <AdminLayout>
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
    </AdminLayout>
  );
};

