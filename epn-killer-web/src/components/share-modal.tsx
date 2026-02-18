import { useState } from 'react';
import { X, Check, Send, MessageCircle, Link2, Share2 } from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';

interface ShareModalProps {
  url: string;
  title?: string;
  text?: string;
  onClose: () => void;
}

export const ShareModal = ({ url, title = 'Поделиться', text, onClose }: ShareModalProps) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const shareText = encodeURIComponent((text ?? 'Присоединяйся к XPLR и получи бонус!') + ' ' + url);
  const shareUrl = encodeURIComponent(url);

  return (
    <div className="fixed inset-0 z-[60] flex items-end sm:items-center justify-center p-4 pb-6 sm:pb-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
      <div className="relative w-full max-w-sm animate-slide-up">
        {/* Power Beam — decorative red line */}
        <div className="absolute -top-px left-1/2 -translate-x-1/2 w-16 h-[2px] bg-gradient-to-r from-transparent via-red-500 to-transparent rounded-full" />

        <div className="bg-[#050507]/95 backdrop-blur-3xl border border-white/10 rounded-2xl overflow-hidden shadow-2xl shadow-black/60">
          {/* Header */}
          <div className="flex items-center justify-between p-5 pb-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/20 flex items-center justify-center">
                <Share2 className="w-5 h-5 text-blue-400" />
              </div>
              <div>
                <h3 className="text-white font-semibold">{title}</h3>
                <p className="text-slate-500 text-xs">Пригласите друзей в XPLR</p>
              </div>
            </div>
            <button onClick={onClose} className="p-2 hover:bg-white/10 rounded-lg transition-colors">
              <X className="w-5 h-5 text-slate-400" />
            </button>
          </div>

          {/* QR Code */}
          <div className="flex justify-center px-5 pb-4">
            <div className="relative bg-white rounded-2xl p-3 shadow-lg">
              <QRCodeSVG
                value={url}
                size={160}
                bgColor="#ffffff"
                fgColor="#0a0a0f"
                level="M"
                includeMargin={false}
              />
              {/* XPLR logo overlay in center */}
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-[#0a0a12] to-[#050508] flex items-center justify-center shadow-md border-2 border-white">
                  <span className="text-blue-500 font-extrabold text-sm">X</span>
                </div>
              </div>
            </div>
          </div>

          {/* Copy link */}
          <div className="px-5 pb-4">
            <button
              onClick={handleCopy}
              className={`w-full flex items-center gap-3 p-3.5 rounded-xl border transition-all duration-200 ${
                copied
                  ? 'bg-emerald-500/10 border-emerald-500/30'
                  : 'bg-white/[0.03] border-white/[0.08] hover:bg-white/[0.06] hover:border-white/15 active:scale-[0.98]'
              }`}
            >
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${
                copied ? 'bg-emerald-500/20' : 'bg-white/[0.06]'
              }`}>
                {copied ? <Check className="w-5 h-5 text-emerald-400" /> : <Link2 className="w-5 h-5 text-slate-300" />}
              </div>
              <div className="flex-1 text-left min-w-0">
                <p className={`text-sm font-medium ${copied ? 'text-emerald-400' : 'text-white'}`}>
                  {copied ? 'Скопировано!' : 'Копировать ссылку'}
                </p>
                <p className="text-xs text-slate-500 truncate">{url}</p>
              </div>
            </button>
          </div>

          {/* Divider */}
          <div className="px-5">
            <div className="h-px bg-white/[0.06]" />
          </div>

          {/* Social grid */}
          <div className="p-5 pt-4">
            <p className="text-slate-500 text-xs font-medium uppercase tracking-wider mb-3">Соцсети</p>
            <div className="grid grid-cols-3 gap-3">
              {/* Telegram */}
              <a
                href={`https://t.me/share/url?url=${shareUrl}&text=${shareText}`}
                target="_blank"
                rel="noopener noreferrer"
                className="flex flex-col items-center gap-2 p-3 rounded-xl bg-white/[0.03] border border-white/[0.06] hover:bg-sky-500/10 hover:border-sky-500/20 transition-all active:scale-95"
              >
                <div className="w-11 h-11 rounded-full bg-gradient-to-br from-sky-400 to-sky-600 flex items-center justify-center shadow-lg shadow-sky-500/20">
                  <Send className="w-5 h-5 text-white" />
                </div>
                <span className="text-[11px] text-slate-400 font-medium">Telegram</span>
              </a>

              {/* WhatsApp */}
              <a
                href={`https://wa.me/?text=${shareText}`}
                target="_blank"
                rel="noopener noreferrer"
                className="flex flex-col items-center gap-2 p-3 rounded-xl bg-white/[0.03] border border-white/[0.06] hover:bg-emerald-500/10 hover:border-emerald-500/20 transition-all active:scale-95"
              >
                <div className="w-11 h-11 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 flex items-center justify-center shadow-lg shadow-emerald-500/20">
                  <MessageCircle className="w-5 h-5 text-white" />
                </div>
                <span className="text-[11px] text-slate-400 font-medium">WhatsApp</span>
              </a>

              {/* Twitter / X */}
              <a
                href={`https://twitter.com/intent/tweet?text=${shareText}`}
                target="_blank"
                rel="noopener noreferrer"
                className="flex flex-col items-center gap-2 p-3 rounded-xl bg-white/[0.03] border border-white/[0.06] hover:bg-white/[0.08] hover:border-white/15 transition-all active:scale-95"
              >
                <div className="w-11 h-11 rounded-full bg-gradient-to-br from-slate-600 to-slate-800 flex items-center justify-center shadow-lg shadow-black/30">
                  <span className="text-white font-bold text-base">𝕏</span>
                </div>
                <span className="text-[11px] text-slate-400 font-medium">Twitter</span>
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
