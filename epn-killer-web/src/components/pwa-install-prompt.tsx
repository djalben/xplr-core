import { useState, useEffect } from 'react';
import { X, Share, Plus, Download } from 'lucide-react';
import { ModalPortal } from './modal-portal';

interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

const isIOS = () =>
  /iPad|iPhone|iPod/.test(navigator.userAgent) && !(window as any).MSStream;

const isStandalone = () =>
  window.matchMedia('(display-mode: standalone)').matches ||
  (navigator as any).standalone === true;

export const PWAInstallPrompt = () => {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [showIOSModal, setShowIOSModal] = useState(false);
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    // Don't show if already installed as PWA
    if (isStandalone()) return;

    // Check if user previously dismissed
    const wasDismissed = sessionStorage.getItem('pwa-prompt-dismissed');
    if (wasDismissed) {
      setDismissed(true);
      return;
    }

    // Android / Chrome: intercept beforeinstallprompt
    const handler = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e as BeforeInstallPromptEvent);
    };
    window.addEventListener('beforeinstallprompt', handler);

    // iOS: show custom modal after 3s delay
    if (isIOS()) {
      const timer = setTimeout(() => setShowIOSModal(true), 3000);
      return () => {
        clearTimeout(timer);
        window.removeEventListener('beforeinstallprompt', handler);
      };
    }

    return () => window.removeEventListener('beforeinstallprompt', handler);
  }, []);

  const handleAndroidInstall = async () => {
    if (!deferredPrompt) return;
    await deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    console.log('[PWA] Install outcome:', outcome);
    setDeferredPrompt(null);
  };

  const handleDismiss = () => {
    setDismissed(true);
    setShowIOSModal(false);
    setDeferredPrompt(null);
    sessionStorage.setItem('pwa-prompt-dismissed', '1');
  };

  // Nothing to show
  if (dismissed || isStandalone()) return null;
  if (!deferredPrompt && !showIOSModal) return null;

  // ── Android banner ──
  if (deferredPrompt) {
    return (
      <div className="fixed bottom-4 left-4 right-4 z-[100] md:left-auto md:right-4 md:w-96 animate-slide-up">
        <div className="bg-[#0d0d12]/95 backdrop-blur-2xl border border-white/10 rounded-2xl p-4 shadow-2xl shadow-black/50 flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shrink-0 shadow-lg shadow-blue-500/20">
            <span className="text-white font-bold text-xl">X</span>
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-white font-semibold text-sm">Установить XPLR</p>
            <p className="text-slate-400 text-xs truncate">Быстрый доступ с экрана «Домой»</p>
          </div>
          <button
            onClick={handleAndroidInstall}
            className="px-4 py-2 bg-blue-500 hover:bg-blue-400 text-white text-sm font-semibold rounded-xl transition-colors shrink-0 flex items-center gap-1.5"
          >
            <Download className="w-4 h-4" />
            Да
          </button>
          <button onClick={handleDismiss} className="p-1.5 hover:bg-white/10 rounded-lg transition-colors shrink-0">
            <X className="w-4 h-4 text-slate-500" />
          </button>
        </div>
      </div>
    );
  }

  // ── iOS Safari instruction modal ──
  if (showIOSModal) {
    return (
      <ModalPortal>
      <div className="fixed inset-0 z-[100] flex items-end justify-center p-4 pb-8 md:items-center">
        <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={handleDismiss} />
        <div className="relative bg-[#0d0d12]/95 backdrop-blur-3xl border border-white/10 rounded-2xl p-6 w-full max-w-sm shadow-2xl shadow-black/60 animate-slide-up">
          {/* Header */}
          <div className="flex items-center justify-between mb-5">
            <div className="flex items-center gap-3">
              <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-lg shadow-blue-500/20">
                <span className="text-white font-bold text-lg">X</span>
              </div>
              <div>
                <p className="text-white font-semibold">Установить XPLR</p>
                <p className="text-slate-400 text-xs">Добавить на экран «Домой»</p>
              </div>
            </div>
            <button onClick={handleDismiss} className="p-2 hover:bg-white/10 rounded-lg transition-colors">
              <X className="w-5 h-5 text-slate-400" />
            </button>
          </div>

          {/* Steps */}
          <div className="space-y-4">
            <div className="flex items-center gap-4 p-3 bg-white/[0.04] rounded-xl border border-white/[0.06]">
              <div className="w-10 h-10 rounded-lg bg-blue-500/15 flex items-center justify-center shrink-0">
                <Share className="w-5 h-5 text-blue-400" />
              </div>
              <div>
                <p className="text-white text-sm font-medium">1. Нажмите «Поделиться»</p>
                <p className="text-slate-500 text-xs">Иконка внизу экрана Safari</p>
              </div>
            </div>

            <div className="flex items-center gap-4 p-3 bg-white/[0.04] rounded-xl border border-white/[0.06]">
              <div className="w-10 h-10 rounded-lg bg-blue-500/15 flex items-center justify-center shrink-0">
                <Plus className="w-5 h-5 text-blue-400" />
              </div>
              <div>
                <p className="text-white text-sm font-medium">2. «На экран Домой»</p>
                <p className="text-slate-500 text-xs">Прокрутите вниз и выберите</p>
              </div>
            </div>

            <div className="flex items-center gap-4 p-3 bg-white/[0.04] rounded-xl border border-white/[0.06]">
              <div className="w-10 h-10 rounded-lg bg-emerald-500/15 flex items-center justify-center shrink-0">
                <Download className="w-5 h-5 text-emerald-400" />
              </div>
              <div>
                <p className="text-white text-sm font-medium">3. Нажмите «Добавить»</p>
                <p className="text-slate-500 text-xs">Приложение появится на экране</p>
              </div>
            </div>
          </div>

          <button
            onClick={handleDismiss}
            className="w-full mt-5 py-3 bg-white/[0.06] hover:bg-white/[0.1] border border-white/10 text-slate-300 font-medium rounded-xl transition-colors text-sm"
          >
            Понятно
          </button>
        </div>
      </div>
      </ModalPortal>
    );
  }

  return null;
};
