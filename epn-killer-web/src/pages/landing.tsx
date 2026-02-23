import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { 
  CreditCard, 
  Users, 
  Globe, 
  Shield, 
  Zap, 
  ArrowRight,
  Check,
  Star,
  ChevronDown,
  ChevronUp,
  Copy,
  Headphones,
  Code,
  Layers,
  TrendingUp,
  Smartphone,
  Building,
  Clock,
  MessageCircle,
  Play
} from 'lucide-react';

// Static ambient background — no canvas, no animation loop
const LandingNeuralBackground = () => (
  <div
    className="fixed inset-0 pointer-events-none"
    style={{
      zIndex: 0,
      background: `
        radial-gradient(ellipse 60% 50% at 20% 20%, rgba(59,130,246,0.10) 0%, transparent 70%),
        radial-gradient(ellipse 50% 40% at 80% 50%, rgba(139,92,246,0.09) 0%, transparent 70%),
        radial-gradient(ellipse 45% 50% at 50% 85%, rgba(99,102,241,0.06) 0%, transparent 70%)
      `,
    }}
  />
);

// Floating 3D Card Component — matches dashboard card branding
const FloatingCard = ({ className = '', delay = 0, variant = 'blue' }: { className?: string; delay?: number; variant?: 'blue' | 'purple' | 'gold' }) => {
  const gradients = {
    blue: 'bg-gradient-to-br from-blue-500 via-blue-600 to-indigo-700',
    purple: 'bg-gradient-to-br from-purple-500 via-violet-600 to-indigo-700',
    gold: 'bg-gradient-to-br from-amber-400 via-orange-500 to-rose-500',
  };

  return (
    <div 
      className={`relative w-72 h-44 rounded-2xl ${gradients[variant]} shadow-2xl shadow-blue-500/30 overflow-hidden ${className}`}
      style={{
        animation: `float ${4 + delay}s ease-in-out infinite`,
        animationDelay: `${delay}s`,
        transform: 'perspective(1000px) rotateY(-5deg) rotateX(5deg)',
      }}
    >
      {/* Glossy overlay */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/30 via-transparent to-transparent rounded-2xl" />
      
      {/* Card content */}
      <div className="relative h-full p-5 flex flex-col justify-between">
        {/* Top row — XPLR + EXPLORER */}
        <div className="flex items-start justify-between">
          <div>
            <p className="text-white/90 text-base font-bold tracking-[0.2em] leading-none">XPLR</p>
            <p className="text-white/60 text-[8px] font-light tracking-[0.25em] uppercase leading-none mt-0.5">Explorer</p>
          </div>
          {/* NFC icon */}
          <svg className="w-6 h-6 text-white/40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1.5-2.5" strokeLinecap="round" />
            <path d="M6.5 16.5A4.5 4.5 0 0 0 13 12c0-2.76-1.5-4-3-5" strokeLinecap="round" />
          </svg>
        </div>

        {/* Bottom — slogan + card number */}
        <div>
          <p className="text-white/40 text-[8px] font-light tracking-[0.15em] uppercase leading-none mb-1.5">БЕЗ ГРАНИЦ</p>
          <p className="text-white/80 font-mono text-sm tracking-widest">•••• •••• •••• 4521</p>
        </div>
      </div>

      {/* Mastercard-style circles bottom right */}
      <div className="absolute bottom-4 right-4 flex -space-x-2">
        <div className="w-6 h-6 rounded-full bg-white/20" />
        <div className="w-6 h-6 rounded-full bg-white/15" />
      </div>
    </div>
  );
};

// Brand SVG Logos
const BrandLogos: Record<string, React.ReactNode> = {
  Netflix: (
    <svg viewBox="0 0 111 30" className="w-16 h-auto" fill="#E50914">
      <path d="M105.062 14.28L111 30c-1.75-.25-3.499-.563-5.28-.845l-3.345-8.686-3.437 7.969c-1.687-.282-3.344-.376-5.031-.595l6.031-13.874L94.468 0h5.063l3.062 7.874L105.875 0h5.124l-5.937 14.28zM90.47 0h-4.594v27.25c1.5.094 3.062.156 4.594.343V0zm-8.563 26.937c-4.187-.281-8.375-.53-12.656-.625V0h4.687v21.875c2.688.062 5.375.28 7.969.405v4.657zM64.25 10.657v4.687h-6.406V26H53.22V0h13.125v4.687h-8.5v5.97h6.406zm-18.906-5.97V26.25c-1.563 0-3.156 0-4.688.062V4.687h-4.844V0h14.406v4.687h-4.874zM30.75 0v21.875c2.75.156 5.531.281 8.281.562v4.532L26.125 26V0H30.75zM21.969 5.25v4.781h-6.531v4.657h6.5V19.5h-6.531v5.28c-1.531.157-3.094.375-4.625.595V0h11.219l-.032 5.25zM0 0v30h4.687V4.687H8.28V0H0z"/>
    </svg>
  ),
  Spotify: (
    <svg viewBox="0 0 168 168" className="w-12 h-12" fill="#1DB954">
      <path d="M84 0C37.7 0 0 37.7 0 84s37.7 84 84 84 84-37.7 84-84S130.3 0 84 0zm38.6 121.2c-1.5 2.4-4.7 3.2-7.1 1.7-19.4-11.9-43.9-14.6-72.7-8-2.8.6-5.5-1.1-6.2-3.9-.6-2.8 1.1-5.5 3.9-6.2 31.5-7.2 58.5-4.1 80.6 9.3 2.4 1.5 3.2 4.7 1.5 7.1zm10.3-22.9c-1.9 3-5.9 4-8.9 2.1-22.2-13.6-56.1-17.6-82.4-9.6-3.4 1-7-1-8-4.4s1-7 4.4-8c30-9.2 67.4-4.7 93 11 3 1.9 4 5.9 2.1 8.9zm.9-23.8c-26.6-15.8-70.6-17.3-96-9.6-4.1 1.2-8.4-1.1-9.6-5.2-1.2-4.1 1.1-8.4 5.2-9.6 29.2-8.8 77.7-7.1 108.4 11.1 3.7 2.2 4.9 6.9 2.7 10.6-2.2 3.7-6.9 4.9-10.6 2.7z"/>
    </svg>
  ),
  ChatGPT: (
    <svg viewBox="0 0 41 41" className="w-12 h-12" fill="none">
      <path d="M37.5324 16.8707C37.9808 15.5241 38.1363 14.0974 37.9886 12.6859C37.8409 11.2744 37.3934 9.91076 36.676 8.68622C35.6126 6.83404 33.9882 5.3676 32.0373 4.4985C30.0864 3.62941 27.9098 3.40259 25.8215 3.85078C24.8796 2.7893 23.7219 1.94125 22.4257 1.36341C21.1295 0.785575 19.7249 0.491269 18.3058 0.500197C16.1708 0.495044 14.0893 1.16803 12.3614 2.42214C10.6335 3.67624 9.34853 5.44666 8.6917 7.47815C7.30085 7.76286 5.98686 8.3414 4.8377 9.17505C3.68854 10.0087 2.73073 11.0782 2.02839 12.312C0.956464 14.1591 0.498905 16.2988 0.721698 18.4228C0.944492 20.5467 1.83612 22.5449 3.268 24.1293C2.81966 25.4759 2.66413 26.9026 2.81182 28.3141C2.95951 29.7256 3.40701 31.0892 4.12437 32.3138C5.18791 34.1659 6.8123 35.6322 8.76321 36.5013C10.7141 37.3704 12.8907 37.5973 14.9789 37.1492C15.9208 38.2107 17.0786 39.0587 18.3747 39.6366C19.6709 40.2144 21.0755 40.5765 22.4947 40.4998C24.6371 40.5054 26.7251 39.8321 28.4577 38.5765C30.1903 37.3209 31.4788 35.548 32.1377 33.5125C33.5286 33.2278 34.8426 32.6493 35.9917 31.8156C37.1409 30.982 38.0987 29.9125 38.801 28.6787C39.8729 26.8316 40.3305 24.6919 40.1077 22.568C39.8849 20.4441 38.9932 18.4458 37.5612 16.8614L37.5324 16.8707ZM22.4978 37.8849C20.7443 37.8874 19.0459 37.2733 17.6994 36.1501C17.7601 36.117 17.8666 36.0586 17.936 36.0161L25.9004 31.4156C26.1003 31.3019 26.2663 31.137 26.381 30.9378C26.4957 30.7386 26.5765 30.5116 26.5765 30.3918V19.5614L29.9561 21.5093C29.9735 21.5184 29.9877 21.5322 29.9975 21.5492C30.0073 21.5661 30.0123 21.5855 30.0119 21.6052V30.5229C30.0094 32.4719 29.2339 34.3407 27.8525 35.7221C26.4711 37.1035 24.6023 37.879 22.6533 37.8815L22.4978 37.8849ZM6.39227 31.0064C5.51397 29.4888 5.19742 27.7107 5.49804 25.9832C5.55718 26.0187 5.66048 26.0818 5.73461 26.1244L13.699 30.7248C13.8975 30.8408 14.1233 30.902 14.3532 30.902C14.5831 30.902 14.8089 30.8408 15.0073 30.7248L24.731 25.1103V29.0063C24.7321 29.0267 24.7282 29.0468 24.7176 29.0641C24.707 29.0814 24.6982 29.0954 24.6802 29.1046L16.6247 33.7497C14.9191 34.7146 12.9112 35.0356 10.9865 34.6531C9.06191 34.2706 7.35442 33.2116 6.18824 31.6649L6.39227 31.0064ZM4.29707 13.6194C5.17156 12.0998 6.55279 10.9364 8.19885 10.3327C8.19885 10.4013 8.19491 10.5236 8.19491 10.6154V19.8945C8.19456 20.0788 8.24722 20.2597 8.34663 20.4157C8.44604 20.5717 8.58819 20.6961 8.75545 20.7742L18.479 26.3888L15.0993 28.3367C15.0825 28.3476 15.0633 28.3543 15.0433 28.3559C15.0232 28.3576 15.003 28.3543 14.9849 28.3462L6.92945 23.7018C5.22713 22.7373 3.93626 21.1773 3.28696 19.3238C2.63766 17.4703 2.67376 15.4498 3.38908 13.6194H4.29707ZM32.0487 20.1918L22.3251 14.5765L25.7047 12.6286C25.7215 12.6177 25.7407 12.611 25.7607 12.6094C25.7808 12.6077 25.801 12.611 25.8191 12.6191L33.8746 17.2635C35.1675 18.0111 36.2176 19.1139 36.8992 20.4427C37.5807 21.7715 37.8628 23.2694 37.7108 24.7551C37.5588 26.2408 36.9934 27.6525 36.1112 28.8279C35.2291 30.0034 34.0375 30.8965 32.6741 31.4049V22.0758C32.6755 21.8921 32.6236 21.7118 32.5257 21.5557C32.4263 21.3997 32.2841 21.2755 32.1169 21.1969L32.0487 20.1918ZM35.3804 14.9828C35.3213 14.9473 35.218 14.8842 35.1438 14.8416L27.1795 10.2412C26.9809 10.1252 26.7551 10.064 26.5253 10.064C26.2954 10.064 26.0696 10.1252 25.8712 10.2412L16.1475 15.8557V11.9598C16.1464 11.9393 16.1508 11.9192 16.1601 11.9019C16.1693 11.8846 16.1831 11.8707 16.2001 11.8615L24.2555 7.21626C25.5476 6.46825 27.0181 6.10047 28.5042 6.15292C29.9903 6.20536 31.4322 6.67586 32.6706 7.51352C33.909 8.35118 34.8934 9.52283 35.5118 10.8978C36.1302 12.2728 36.3569 13.7942 36.166 15.2907L35.3804 14.9828ZM14.1434 21.4037L10.7614 19.4559C10.7439 19.4467 10.7297 19.4329 10.7199 19.4159C10.71 19.399 10.705 19.3796 10.7054 19.3599V10.4421C10.708 8.95097 11.1254 7.49103 11.9132 6.23225C12.701 4.97349 13.9216 3.96799 15.3163 3.33149C16.7109 2.69499 18.2647 2.45448 19.7926 2.63635C21.3205 2.81821 22.7734 3.41471 23.9844 4.36122C23.9237 4.39432 23.8173 4.45266 23.7479 4.49519L15.7835 9.09559C15.5836 9.2093 15.4176 9.37419 15.3029 9.57341C15.1882 9.77263 15.1291 9.99863 15.1316 10.2281L14.1434 21.4037ZM16.1475 17.3867L20.4386 14.9828L24.7297 17.3867V22.1946L20.4386 24.5986L16.1475 22.1946V17.3867Z" fill="#10A37F"/>
    </svg>
  ),
  Uber: (
    <svg viewBox="0 0 118 24" className="w-16 h-auto" fill="#000">
      <path d="M31.7 1.3h-4.7v11.3c0 2.8-1.8 4.6-4.4 4.6-2.7 0-4.4-1.8-4.4-4.6V1.3h-4.7V13c0 5.1 3.6 8.6 9.1 8.6 5.5 0 9.1-3.5 9.1-8.6V1.3zM38 1.3h-4.6v20.2H38v-7.8c0-3.2 2-5.3 4.9-5.3.7 0 1.4.1 2 .3l.6-4.4c-.6-.2-1.3-.3-2-.3-2.6 0-4.5 1.3-5.5 3.3V1.3zM60.4 11.4c0-6-4.2-10.2-10.1-10.2-6 0-10.4 4.2-10.4 10.2s4.4 10.2 10.4 10.2c5.9 0 10.1-4.2 10.1-10.2zm-4.7 0c0 3.5-2.2 5.9-5.4 5.9s-5.7-2.4-5.7-5.9 2.4-5.9 5.7-5.9 5.4 2.4 5.4 5.9zM71.3 1.3h-4.7v20.2h4.7V1.3zM87.8 1.3h-4.5v20.2h4.5v-7c0-3.8 2.2-5.8 5.2-5.8 2.7 0 4.3 1.8 4.3 4.8v8h4.6v-9.2c0-4.7-3.3-8-8-8-2.9 0-5 1.4-6.1 3.5V1.3zM118 11.4c0-6-4.2-10.2-10.2-10.2-6 0-10.4 4.2-10.4 10.2s4.4 10.2 10.4 10.2c5.9 0 10.2-4.2 10.2-10.2zm-4.7 0c0 3.5-2.2 5.9-5.5 5.9s-5.7-2.4-5.7-5.9 2.4-5.9 5.7-5.9 5.5 2.4 5.5 5.9z"/>
    </svg>
  ),
  Airbnb: (
    <svg viewBox="0 0 102 32" className="w-14 h-auto" fill="#FF5A5F">
      <path d="M29.24 22.68c-.16-.39-.31-.8-.47-1.15l-.74-1.67-.03-.03c-2.2-4.8-4.55-9.68-7.04-14.48l-.1-.2c-.25-.47-.5-.99-.76-1.47-.32-.57-.63-1.18-1.14-1.76-.96-1.12-2.35-1.76-3.87-1.76-1.56 0-2.97.67-3.87 1.76-.47.54-.78 1.15-1.1 1.72-.26.5-.52 1-.77 1.51l-.1.2c-2.48 4.8-4.83 9.68-7.04 14.48l-.02.03c-.16.35-.31.74-.47 1.15-.15.37-.29.74-.39 1.12-.3 1.17-.21 2.37.25 3.47.88 2.15 2.91 3.59 5.31 3.75.34.02.68.02 1.04-.02.54-.05 1.08-.15 1.6-.35 1.31-.46 2.54-1.25 3.91-2.51 1.65 1.53 3.11 2.35 4.58 2.73.76.21 1.53.31 2.28.21 2.43-.21 4.42-1.72 5.2-3.86.45-1.1.55-2.3.25-3.47-.1-.41-.22-.79-.37-1.16zm-14.86 2.71c-1.4-1.53-2.35-3.12-2.76-4.61-.2-.73-.25-1.39-.16-2.01.07-.53.23-1.01.49-1.45.55-.93 1.48-1.58 2.56-1.79.06-.02.16-.02.25-.02 1.09 0 2.14.47 2.95 1.32.57.6 1.05 1.35 1.45 2.22.23-.52.49-1.02.81-1.47.91-1.33 2.14-2.07 3.47-2.07.09 0 .21 0 .3.02 1.08.16 2.02.86 2.57 1.79.26.44.42.92.49 1.45.09.62.05 1.28-.16 2.01-.41 1.49-1.35 3.08-2.76 4.61-1.08 1.18-2.32 2.19-3.73 3.01-1.4-.82-2.64-1.83-3.77-3.01z"/>
    </svg>
  ),
  Booking: (
    <svg viewBox="0 0 300 50" className="w-20 h-auto" fill="#003580">
      <path d="M22.5 8.3h13.8c6.9 0 12.5 2 12.5 9.3 0 4.6-2.5 7.1-6.3 8.3v.2c5.2.8 8.3 4.2 8.3 9.4 0 8.8-7.1 11.7-15 11.7H22.5V8.3zm10.4 15.4h3.8c3.3 0 5.4-1.3 5.4-4.2 0-3.1-2.1-4.2-5.4-4.2h-3.8v8.4zm0 16.7h4.2c3.8 0 6.3-1.5 6.3-5 0-3.3-2.5-5-6.7-5h-3.8v10zM54.2 26.7c0-12.1 7.1-19.2 17.1-19.2 10 0 17.1 7.1 17.1 19.2s-7.1 19.2-17.1 19.2c-10-.1-17.1-7.1-17.1-19.2zm23.7 0c0-7.1-2.5-11.7-6.7-11.7s-6.7 4.6-6.7 11.7 2.5 11.7 6.7 11.7 6.7-4.6 6.7-11.7zM92.9 26.7c0-12.1 7.1-19.2 17.1-19.2 10 0 17.1 7.1 17.1 19.2s-7.1 19.2-17.1 19.2c-10-.1-17.1-7.1-17.1-19.2zm23.8 0c0-7.1-2.5-11.7-6.7-11.7s-6.7 4.6-6.7 11.7 2.5 11.7 6.7 11.7 6.7-4.6 6.7-11.7zM131.7 8.3h10.4v19.2h.2L155.2 8.3h12.9l-14.6 18.8L169.6 47h-13.3l-10-15-4.2 5.4V47h-10.4V8.3zM175 8.3h10.4V47H175V8.3zM191.7 8.3h10.4l10.8 22.5h.2V8.3h9.6V47h-10l-11.3-23.3h-.2V47h-9.6V8.3zM228.3 26.7c0-11.7 7.5-19.2 17.9-19.2 6.3 0 11.3 2.5 14.2 7.5l-8.3 5c-1.7-2.5-3.3-4.6-6.3-4.6-4.6 0-7.1 4.6-7.1 11.3 0 7.1 2.5 11.7 7.1 11.7 3.3 0 5.4-2.5 6.7-5l8.8 4.6c-3.3 5.8-8.3 8.8-15.4 8.8-10.4-.1-17.6-7.5-17.6-19.1z"/>
    </svg>
  ),
  Steam: (
    <svg viewBox="0 0 233 233" className="w-12 h-12" fill="#1b2838">
      <path d="M116.5 0C53.8 0 2.7 46.5.1 106.1l62.6 25.9c5.3-3.6 11.7-5.7 18.6-5.7 1.2 0 2.5.1 3.7.2l27.8-40.3v-.6c0-24.9 20.2-45.2 45.2-45.2s45.2 20.2 45.2 45.2-20.2 45.2-45.2 45.2h-1.1l-39.6 28.3c0 .9.1 1.9.1 2.8 0 18.7-15.2 33.9-33.9 33.9s-33.9-15.2-33.9-33.9c0-.3 0-.6 0-.9l-45.2-18.7C23.3 199.7 65.4 233 116.5 233 180.8 233 233 180.8 233 116.5S180.8 0 116.5 0z"/>
      <path d="M73.7 193.3l-14.4-6c2.5 5.2 6.8 9.6 12.5 12.2 12.4 5.1 26.7-.8 31.8-13.2 2.5-6 2.4-12.6-.1-18.6-2.5-6-7.2-10.7-13.2-13.2-5.9-2.5-12.2-2.4-17.8-.4l14.9 6.2c9.1 3.8 13.5 14.4 9.7 23.6-3.8 9.1-14.3 13.4-23.4 9.4z" fill="#fff"/>
      <path d="M203.5 85.6c0-16.6-13.5-30.1-30.1-30.1s-30.1 13.5-30.1 30.1 13.5 30.1 30.1 30.1 30.1-13.5 30.1-30.1zm-53.5 0c0-12.9 10.5-23.4 23.4-23.4s23.4 10.5 23.4 23.4-10.5 23.4-23.4 23.4-23.4-10.5-23.4-23.4z"/>
    </svg>
  ),
  Adobe: (
    <svg viewBox="0 0 30 26" className="w-12 h-12" fill="#FF0000">
      <path d="M19 0h11v26zM11 0H0v26zM15 9.6L22 26h-4.5l-2.1-5.5h-5.3z"/>
    </svg>
  ),
  Trip: (
    <svg viewBox="0 0 100 30" className="w-16 h-auto" fill="#287DFA">
      <circle cx="15" cy="15" r="15"/>
      <text x="35" y="22" fontFamily="Arial" fontSize="20" fontWeight="bold" fill="#287DFA">Trip.com</text>
    </svg>
  ),
  Bolt: (
    <svg viewBox="0 0 100 40" className="w-16 h-auto" fill="#34D186">
      <circle cx="20" cy="20" r="18" fill="#34D186"/>
      <path d="M15 25l8-15h-4l2-5-8 15h4l-2 5z" fill="#fff"/>
      <text x="42" y="28" fontFamily="Arial" fontSize="22" fontWeight="bold" fill="#34D186">Bolt</text>
    </svg>
  ),
  Grab: (
    <svg viewBox="0 0 100 40" className="w-16 h-auto" fill="#00B14F">
      <circle cx="20" cy="20" r="18" fill="#00B14F"/>
      <text x="42" y="28" fontFamily="Arial" fontSize="22" fontWeight="bold" fill="#00B14F">Grab</text>
    </svg>
  ),
};

// Service Logo Component - renders real brand logos
const ServiceLogo = ({ name, color }: { name: string; color: string }) => {
  const logo = BrandLogos[name];
  
  if (logo) {
    return (
      <div className="flex items-center justify-center w-20 h-16 rounded-xl bg-white p-2 shadow-lg transition-transform hover:scale-110">
        {logo}
      </div>
    );
  }
  
  // Fallback to letter if no logo
  return (
    <div className={`flex items-center justify-center w-16 h-16 rounded-xl ${color} text-white font-bold text-xl shadow-lg transition-transform hover:scale-110`}>
      {name.charAt(0)}
    </div>
  );
};

// Stats counter
const StatCounter = ({ value, label, suffix = '' }: { value: string; label: string; suffix?: string }) => (
  <div className="text-center">
    <div className="text-3xl md:text-4xl font-bold text-white mb-2">
      {value}<span className="text-blue-400">{suffix}</span>
    </div>
    <div className="text-slate-400 text-sm">{label}</div>
  </div>
);

// Feature card
const FeatureCard = ({ icon: Icon, title, description }: { icon: React.ElementType; title: string; description: string }) => (
  <div className="glass-card p-6 card-hover group">
    <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
      <Icon className="w-7 h-7 text-blue-400" />
    </div>
    <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
    <p className="text-slate-400 text-sm leading-relaxed">{description}</p>
  </div>
);

// Card type card
const CardTypeCard = ({ type, currency, price, features, popular = false }: { 
  type: string; 
  currency: string; 
  price: string; 
  features: string[];
  popular?: boolean;
}) => (
  <div className={`glass-card p-6 relative ${popular ? 'border-blue-500/50 shadow-lg shadow-blue-500/20' : ''}`}>
    {popular && (
      <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-4 py-1 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full text-xs font-semibold text-white shadow-lg">
        Выбор 72% клиентов
      </div>
    )}
    <div className="text-sm text-slate-500 mb-1">{currency}</div>
    <h3 className="text-xl font-bold text-white mb-2">{type}</h3>
    <div className="text-3xl font-bold text-white mb-4">
      {price}<span className="text-lg text-slate-500">₽</span>
    </div>
    <ul className="space-y-2 mb-6">
      {features.map((feature, i) => (
        <li key={i} className="flex items-center gap-2 text-sm text-slate-300">
          <Check className="w-4 h-4 text-emerald-400 flex-shrink-0" />
          {feature}
        </li>
      ))}
    </ul>
    <Link to="/auth">
      <button className={`w-full py-3 rounded-xl font-medium transition-all ${
        popular 
          ? 'bg-gradient-to-r from-blue-500 to-purple-500 text-white hover:shadow-lg hover:shadow-blue-500/25' 
          : 'bg-white/10 text-white hover:bg-white/15 border border-white/10'
      }`}>
        Выпустить карту
      </button>
    </Link>
  </div>
);

// Grade tier
const GradeTier = ({ name, commission, volume, active = false }: { name: string; commission: string; volume: string; active?: boolean }) => (
  <div className={`p-4 rounded-xl border transition-all ${
    active 
      ? 'bg-blue-500/20 border-blue-500/50 shadow-lg shadow-blue-500/20' 
      : 'bg-white/5 border-white/10 hover:border-white/20'
  }`}>
    <div className="font-semibold text-white mb-1">{name}</div>
    <div className="text-2xl font-bold text-blue-400 mb-1">{commission}</div>
    <div className="text-xs text-slate-500">от {volume}</div>
  </div>
);

// Step card
const StepCard = ({ number, title, description }: { number: number; title: string; description: string }) => (
  <div className="relative">
    <div className="relative z-10 w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-white font-bold text-lg mb-4 shadow-lg shadow-blue-500/30">
      {number}
    </div>
    <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
    <p className="text-slate-400 text-sm">{description}</p>
  </div>
);

// Review card
const ReviewCard = ({ name, role, text, rating }: { name: string; role: string; text: string; rating: number }) => (
  <div className="glass-card p-6">
    <div className="flex items-center gap-4 mb-4">
      <div className="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-white font-bold shadow-lg">
        {name[0]}
      </div>
      <div>
        <div className="font-semibold text-white">{name}</div>
        <div className="text-sm text-slate-500">{role}</div>
      </div>
    </div>
    <div className="flex gap-1 mb-3">
      {[...Array(5)].map((_, i) => (
        <Star key={i} className={`w-4 h-4 ${i < rating ? 'text-amber-400 fill-amber-400' : 'text-slate-600'}`} />
      ))}
    </div>
    <p className="text-slate-300 text-sm leading-relaxed">{text}</p>
  </div>
);

// FAQ Item
const FAQItem = ({ question, answer }: { question: string; answer: string }) => {
  const [isOpen, setIsOpen] = useState(false);
  
  return (
    <div className="border-b border-white/10">
      <button 
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between py-5 text-left"
      >
        <span className="font-medium text-white pr-4">{question}</span>
        {isOpen ? (
          <ChevronUp className="w-5 h-5 text-slate-500 flex-shrink-0" />
        ) : (
          <ChevronDown className="w-5 h-5 text-slate-500 flex-shrink-0" />
        )}
      </button>
      {isOpen && (
        <div className="pb-5 text-slate-400 text-sm leading-relaxed animate-fade-in">
          {answer}
        </div>
      )}
    </div>
  );
};

// Service icons for marquee
const services = [
  { name: 'Netflix', color: 'bg-red-600' },
  { name: 'Spotify', color: 'bg-green-600' },
  { name: 'ChatGPT', color: 'bg-emerald-600' },
  { name: 'Uber', color: 'bg-slate-700' },
  { name: 'Bolt', color: 'bg-green-500' },
  { name: 'Grab', color: 'bg-green-600' },
  { name: 'Airbnb', color: 'bg-rose-500' },
  { name: 'Booking', color: 'bg-blue-600' },
  { name: 'Steam', color: 'bg-slate-600' },
  { name: 'Adobe', color: 'bg-red-500' },
];

export const LandingPage = () => {
  const [copied, setCopied] = useState(false);

  const copyReferralLink = () => {
    navigator.clipboard.writeText('https://xplr.io/ref/USER123');
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const faqs = [
    {
      question: 'Как работает виртуальная карта?',
      answer: 'Виртуальная карта — это полноценная банковская карта без физического носителя. Вы получаете номер карты, срок действия и CVV-код, которые можно использовать для онлайн-платежей. Карта привязана к вашему счёту в XPLR и работает везде, где принимают Visa/Mastercard.'
    },
    {
      question: 'Что можно оплачивать картами XPLR?',
      answer: 'Картами XPLR можно оплачивать более 560 сервисов: Netflix, Spotify, ChatGPT, Midjourney, Booking, Airbnb, Steam, Adobe и многие другие. Также карты работают для оплаты рекламы в Facebook, Google, TikTok и других платформах.'
    },
    {
      question: 'Безопасно ли использовать XPLR?',
      answer: 'Да, абсолютно безопасно. Все транзакции защищены 3D Secure, данные карт зашифрованы по стандарту PCI DSS. Каждая карта имеет отдельный баланс и лимиты, что минимизирует риски. Вы можете закрыть карту в любой момент.'
    },
    {
      question: 'Как пополнить карту?',
      answer: 'Пополнение происходит через Систему Быстрых Платежей (СБП) с любого российского банка. Средства зачисляются мгновенно. Комиссия за пополнение — 0₽. Минимальная сумма пополнения — 500₽.'
    },
    {
      question: 'Какая комиссия за использование карт?',
      answer: 'Для личного использования комиссия составляет 3.5% от суммы операции. Для арбитража действует грейдовая система: от 6.7% (Стандарт) до 3% (Блэк). Чем больше оборот, тем ниже комиссия.'
    },
    {
      question: 'Можно ли добавить карту в Apple Pay?',
      answer: 'Да, карты типов "Путешествия" и "Премиум" можно добавить в Apple Pay и Google Pay для бесконтактной оплаты. Инструкция по добавлению доступна в личном кабинете после выпуска карты.'
    },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#0a0a0f] via-[#0f0f18] to-[#12121a] relative overflow-x-hidden">
      <LandingNeuralBackground />
      
      {/* Content wrapper */}
      <div className="relative z-10">
        
        {/* Navigation */}
        <nav className="fixed top-0 left-0 right-0 z-50 px-6 py-4 bg-[#0a0a0f]/80 backdrop-blur-xl border-b border-white/[0.08]">
          <div className="max-w-7xl mx-auto flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg shadow-blue-500/30">
                <span className="text-white font-bold text-lg">X</span>
              </div>
              <span className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                XPLR
              </span>
            </div>
            <div className="hidden md:flex items-center gap-8">
              <a href="#personal" className="text-slate-400 hover:text-white transition-colors font-medium">Для личного</a>
              <a href="#arbitrage" className="text-slate-400 hover:text-white transition-colors font-medium">Для арбитража</a>
              <a href="#how-it-works" className="text-slate-400 hover:text-white transition-colors font-medium">Как это работает</a>
              <a href="#faq" className="text-slate-400 hover:text-white transition-colors font-medium">FAQ</a>
            </div>
            <div className="flex items-center gap-3">
              <Link to="/auth">
                <button className="px-6 py-2.5 bg-gradient-to-r from-blue-500 to-purple-500 rounded-xl text-white font-medium hover:shadow-lg hover:shadow-blue-500/30 transition-all">
                  Войти
                </button>
              </Link>
            </div>
          </div>
        </nav>

        {/* Hero Section */}
        <section className="min-h-screen flex items-center pt-24 pb-20 px-6">
          <div className="max-w-7xl mx-auto w-full">
            <div className="grid lg:grid-cols-2 gap-12 items-center">
              <div className="space-y-8">
                <div className="inline-flex items-center gap-2 px-4 py-2 bg-blue-500/20 border border-blue-500/30 rounded-full text-blue-400 text-sm font-medium">
                  <Zap className="w-4 h-4" />
                  Выпуск карты за 2 минуты
                </div>
                
                <h1 className="text-4xl md:text-6xl font-bold text-white leading-tight">
                  Виртуальные карты для оплаты{' '}
                  <span className="bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
                    по всему миру
                  </span>
                </h1>
                
                <p className="text-xl text-slate-400 leading-relaxed max-w-xl">
                  Оплачивайте подписки, путешествуйте, запускайте рекламу — одна платформа для всего. Принимается в 190 странах.
                </p>
                
                <div className="flex flex-col sm:flex-row gap-4">
                  <Link to="/auth">
                    <button className="px-8 py-4 bg-gradient-to-r from-blue-500 to-purple-500 rounded-xl text-white font-semibold text-lg hover:shadow-xl hover:shadow-blue-500/30 transition-all flex items-center justify-center gap-2 w-full sm:w-auto">
                      Выпустить карту
                      <ArrowRight className="w-5 h-5" />
                    </button>
                  </Link>
                  <a href="#how-it-works">
                    <button className="px-8 py-4 bg-white/10 border border-white/10 rounded-xl text-white font-semibold text-lg hover:bg-white/15 transition-all w-full sm:w-auto flex items-center justify-center gap-2">
                      <Play className="w-5 h-5" />
                      Узнать больше
                    </button>
                  </a>
                </div>

                {/* Stats */}
                <div className="flex items-center gap-8 pt-8 border-t border-white/10">
                  <StatCounter value="100K" suffix="+" label="Пользователей" />
                  <StatCounter value="190" suffix="+" label="Стран" />
                  <StatCounter value="24/7" label="Поддержка" />
                </div>
              </div>

              {/* Floating Cards */}
              <div className="relative h-[500px] hidden lg:block">
                <FloatingCard className="absolute top-0 left-8" variant="blue" delay={0} />
                <FloatingCard className="absolute top-24 right-0" variant="purple" delay={0.5} />
                <FloatingCard className="absolute bottom-12 left-16" variant="gold" delay={1} />
              </div>
            </div>
          </div>
        </section>

        {/* Services Marquee */}
        <section className="py-12 bg-white/[0.02] border-y border-white/[0.05] overflow-hidden">
          <div className="max-w-7xl mx-auto px-6 text-center mb-8">
            <p className="text-slate-400 font-medium">Оплачивайте более <span className="text-blue-400 font-bold">560+ сервисов</span> по всему миру</p>
          </div>
          <div className="relative">
            <div className="flex gap-8 marquee-content">
              {[...services, ...services].map((service, i) => (
                <ServiceLogo key={i} name={service.name} color={service.color} />
              ))}
            </div>
          </div>
        </section>

        {/* Advantages */}
        <section className="py-24 px-6">
          <div className="max-w-7xl mx-auto">
            <div className="grid md:grid-cols-4 gap-6">
              <FeatureCard 
                icon={Clock}
                title="Выпуск за 2 минуты"
                description="Мгновенная выдача виртуальной карты сразу после регистрации"
              />
              <FeatureCard 
                icon={Building}
                title="Пополнение через СБП"
                description="Переводите рубли напрямую с любого российского банка без комиссии"
              />
              <FeatureCard 
                icon={Smartphone}
                title="Apple Pay / Google Pay"
                description="Добавьте карту в телефон для бесконтактной оплаты где угодно"
              />
              <FeatureCard 
                icon={Globe}
                title="190 стран мира"
                description="Карты принимаются везде, где работают Visa и Mastercard"
              />
            </div>
          </div>
        </section>

        {/* Personal Section */}
        <section id="personal" className="py-24 px-6 bg-white/[0.02]">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <div className="inline-flex items-center gap-2 px-4 py-2 bg-emerald-500/20 border border-emerald-500/30 rounded-full text-emerald-400 text-sm font-medium mb-6">
                <CreditCard className="w-4 h-4" />
                Для личного использования
              </div>
              <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
                Оплачивайте подписки и путешествуйте без границ
              </h2>
              <p className="text-xl text-slate-400 max-w-2xl mx-auto">
                Карты для Netflix, Spotify, ChatGPT, Uber, Booking, Airbnb и тысяч других сервисов
              </p>
            </div>

            {/* Card Types */}
            <div className="grid md:grid-cols-3 gap-6 mb-16">
              <CardTypeCard 
                type="Подписки"
                currency="EUR"
                price="2 990"
                features={['Выпуск за 2 минуты', 'Пополнение через СБП', '3D Secure защита', 'Netflix, Spotify и др.']}
              />
              <CardTypeCard 
                type="Путешествия"
                currency="USD"
                price="3 990"
                features={['Apple Pay / Google Pay', 'Работает в 190 странах', '0₽ комиссия за пополнение', 'Booking, Airbnb, Uber']}
                popular
              />
              <CardTypeCard 
                type="Премиум"
                currency="USD/EUR"
                price="14 990"
                features={['Безлимитные операции', 'Приоритетная поддержка 24/7', 'Все возможности', 'VIP условия']}
              />
            </div>

            <div className="text-center">
              <Link to="/auth">
                <button className="px-8 py-4 bg-gradient-to-r from-blue-500 to-purple-500 rounded-xl text-white font-semibold text-lg hover:shadow-xl hover:shadow-blue-500/30 transition-all inline-flex items-center gap-2">
                  Выпустить карту сейчас
                  <ArrowRight className="w-5 h-5" />
                </button>
              </Link>
            </div>
          </div>
        </section>

        {/* Arbitrage Section */}
        <section id="arbitrage" className="py-24 px-6">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <div className="inline-flex items-center gap-2 px-4 py-2 bg-purple-500/20 border border-purple-500/30 rounded-full text-purple-400 text-sm font-medium mb-6">
                <TrendingUp className="w-4 h-4" />
                Для арбитража и рекламы
              </div>
              <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
                Масштабируйте рекламные кампании
              </h2>
              <p className="text-xl text-slate-400 max-w-2xl mx-auto">
                Виртуальные карты для Facebook, Google, TikTok Ads и других рекламных платформ
              </p>
            </div>

            {/* Grade System */}
            <div className="glass-card p-8 mb-16">
              <h3 className="text-xl font-bold text-white mb-6 text-center">Грейдовая система комиссий</h3>
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <GradeTier name="Standard" commission="6.7%" volume="$0" />
                <GradeTier name="Silver" commission="6.0%" volume="$1K" />
                <GradeTier name="Gold" commission="5.0%" volume="$10K" active />
                <GradeTier name="Platinum" commission="4.0%" volume="$50K" />
                <GradeTier name="Black" commission="3.0%" volume="$100K" />
              </div>
            </div>

            {/* Arbitrage Features */}
            <div className="grid md:grid-cols-3 gap-6 mb-12">
              <FeatureCard 
                icon={Layers}
                title="Массовый выпуск"
                description="Выпускайте от 1 до 100 карт одной кнопкой для ваших рекламных кампаний"
              />
              <FeatureCard 
                icon={Users}
                title="Командные аккаунты"
                description="Добавляйте участников команды и распределяйте лимиты между ними"
              />
              <FeatureCard 
                icon={Code}
                title="API интеграция"
                description="Полнофункциональный API для автоматизации выпуска и управления картами"
              />
            </div>

            <div className="text-center">
              <Link to="/auth">
                <button className="px-8 py-4 bg-gradient-to-r from-purple-500 to-pink-500 rounded-xl text-white font-semibold text-lg hover:shadow-xl hover:shadow-purple-500/30 transition-all inline-flex items-center gap-2">
                  Создать команду
                  <ArrowRight className="w-5 h-5" />
                </button>
              </Link>
            </div>
          </div>
        </section>

        {/* How It Works */}
        <section id="how-it-works" className="py-24 px-6 bg-white/[0.02]">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
                Как это работает
              </h2>
              <p className="text-xl text-slate-400">
                Четыре простых шага до вашей виртуальной карты
              </p>
            </div>

            <div className="grid md:grid-cols-4 gap-8 relative">
              {/* Connection line — SVG for precise control across all 4 step badges */}
              <svg className="hidden md:block absolute top-[23px] left-0 right-0 h-[6px] overflow-visible pointer-events-none" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="stepLineGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#3b82f6" />
                    <stop offset="50%" stopColor="#8b5cf6" />
                    <stop offset="100%" stopColor="#ec4899" />
                  </linearGradient>
                  <filter id="stepLineGlow">
                    <feGaussianBlur stdDeviation="2" result="blur" />
                    <feMerge><feMergeNode in="blur" /><feMergeNode in="SourceGraphic" /></feMerge>
                  </filter>
                </defs>
                {/* Glow layer */}
                <line x1="12.5%" y1="50%" x2="87.5%" y2="50%" stroke="url(#stepLineGrad)" strokeWidth="4" strokeLinecap="round" opacity="0.3" filter="url(#stepLineGlow)" />
                {/* Main line */}
                <line x1="12.5%" y1="50%" x2="87.5%" y2="50%" stroke="url(#stepLineGrad)" strokeWidth="2.5" strokeLinecap="round" opacity="0.7" />
              </svg>
              
              <StepCard 
                number={1}
                title="Зарегистрируйтесь"
                description="Создайте аккаунт за 2 минуты с email и паролем"
              />
              <StepCard 
                number={2}
                title="Выберите тип карты"
                description="Подберите карту под ваши задачи: подписки, путешествия или арбитраж"
              />
              <StepCard 
                number={3}
                title="Пополните через СБП"
                description="Переведите рубли с любого российского банка мгновенно"
              />
              <StepCard 
                number={4}
                title="Оплачивайте"
                description="Используйте карту для оплаты по всему миру"
              />
            </div>
          </div>
        </section>

        {/* Reviews */}
        <section id="reviews" className="py-24 px-6">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
                Что говорят пользователи
              </h2>
              <p className="text-xl text-slate-400">
                Более 100 000 довольных клиентов
              </p>
            </div>

            <div className="grid md:grid-cols-3 gap-6">
              <ReviewCard 
                name="Александр"
                role="Предприниматель"
                text="Наконец-то нормальный сервис для оплаты зарубежных подписок! Выпустил карту за пару минут, уже оплатил ChatGPT и Netflix. Всё работает как часы."
                rating={5}
              />
              <ReviewCard 
                name="Мария"
                role="Маркетолог"
                text="Использую для арбитража Facebook Ads. Массовый выпуск карт — это то, что нужно. Комиссии адекватные, поддержка отвечает быстро. Рекомендую!"
                rating={5}
              />
              <ReviewCard 
                name="Дмитрий"
                role="Путешественник"
                text="Брал карту в поездку по Европе. Работает везде: Booking, Uber, рестораны. Apple Pay добавил за минуту. Теперь всегда с XPLR."
                rating={5}
              />
            </div>
          </div>
        </section>

        {/* FAQ */}
        <section id="faq" className="py-24 px-6 bg-white/[0.02]">
          <div className="max-w-3xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
                Частые вопросы
              </h2>
              <p className="text-xl text-slate-400">
                Ответы на популярные вопросы о XPLR
              </p>
            </div>

            <div className="glass-card p-6 md:p-8">
              {faqs.map((faq, i) => (
                <FAQItem key={i} question={faq.question} answer={faq.answer} />
              ))}
            </div>
          </div>
        </section>

        {/* Partner Program */}
        <section className="py-24 px-6">
          <div className="max-w-4xl mx-auto">
            <div className="bg-gradient-to-br from-blue-600 via-indigo-600 to-violet-700 rounded-3xl p-8 md:p-12 text-center relative overflow-hidden shadow-2xl shadow-blue-500/20">
              {/* Background decoration */}
              <div className="absolute top-0 left-0 w-64 h-64 bg-white/10 rounded-full blur-3xl -translate-x-1/2 -translate-y-1/2" />
              <div className="absolute bottom-0 right-0 w-64 h-64 bg-white/10 rounded-full blur-3xl translate-x-1/2 translate-y-1/2" />
              
              <div className="relative z-10">
                <div className="inline-flex items-center gap-2 px-4 py-2 bg-white/20 rounded-full text-white text-sm font-medium mb-6">
                  <TrendingUp className="w-4 h-4" />
                  Партнёрская программа
                </div>
                
                <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
                  Зарабатывайте с нами
                </h2>
                
                <p className="text-xl text-white/80 mb-8 max-w-lg mx-auto">
                  Приглашайте друзей и получайте <span className="text-white font-bold">$10</span> за каждого нового пользователя
                </p>

                <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-8">
                  <div className="flex-1 max-w-md w-full">
                    <div className="flex items-center gap-2 p-3 bg-white/10 border border-white/20 rounded-xl backdrop-blur">
                      <input 
                        type="text" 
                        value="https://xplr.io/ref/USER123" 
                        readOnly 
                        className="flex-1 bg-transparent text-white text-sm outline-none"
                      />
                      <button 
                        onClick={copyReferralLink}
                        className={`px-4 py-2 rounded-lg font-medium transition-all flex items-center gap-2 ${
                          copied 
                            ? 'bg-emerald-500 text-white' 
                            : 'bg-white text-slate-800 hover:bg-slate-100'
                        }`}
                      >
                        <Copy className="w-4 h-4" />
                        {copied ? 'Скопировано!' : 'Копировать'}
                      </button>
                    </div>
                  </div>
                </div>

                <Link to="/auth">
                  <button className="px-8 py-4 bg-white rounded-xl text-slate-800 font-semibold text-lg hover:bg-slate-100 transition-all inline-flex items-center gap-2 shadow-lg">
                    Присоединиться к программе
                    <ArrowRight className="w-5 h-5" />
                  </button>
                </Link>
              </div>
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="py-24 px-6">
          <div className="max-w-4xl mx-auto text-center">
            <h2 className="text-3xl md:text-5xl font-bold text-white mb-6">
              Готовы начать?
            </h2>
            <p className="text-xl text-slate-400 mb-8">
              Присоединяйтесь к 100 000+ пользователей XPLR
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <Link to="/auth">
                <button className="px-10 py-5 bg-gradient-to-r from-blue-500 to-purple-500 rounded-xl text-white font-semibold text-xl hover:shadow-2xl hover:shadow-blue-500/30 transition-all flex items-center gap-3">
                  Создать аккаунт бесплатно
                  <ArrowRight className="w-6 h-6" />
                </button>
              </Link>
            </div>
          </div>
        </section>

        {/* Footer */}
        <footer className="py-16 px-6 bg-[#0a0a0f] border-t border-white/[0.08]">
          <div className="max-w-7xl mx-auto">
            <div className="grid md:grid-cols-4 gap-12 mb-12">
              {/* Logo */}
              <div>
                <div className="flex items-center gap-3 mb-4">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg shadow-blue-500/25">
                    <span className="text-white font-bold text-lg">X</span>
                  </div>
                  <span className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                    XPLR
                  </span>
                </div>
                <p className="text-slate-500 text-sm">
                  Финтех-платформа для виртуальных карт нового поколения
                </p>
              </div>

              {/* Links */}
              <div>
                <h4 className="font-semibold text-white mb-4">Продукт</h4>
                <ul className="space-y-2">
                  <li><a href="#personal" className="text-slate-400 hover:text-white transition-colors text-sm">Для личного</a></li>
                  <li><a href="#arbitrage" className="text-slate-400 hover:text-white transition-colors text-sm">Для арбитража</a></li>
                  <li><a href="#" className="text-slate-400 hover:text-white transition-colors text-sm">Тарифы</a></li>
                  <li><a href="#" className="text-slate-400 hover:text-white transition-colors text-sm">API</a></li>
                </ul>
              </div>

              <div>
                <h4 className="font-semibold text-white mb-4">Поддержка</h4>
                <ul className="space-y-2">
                  <li><a href="#faq" className="text-slate-400 hover:text-white transition-colors text-sm">FAQ</a></li>
                  <li><a href="#" className="text-slate-400 hover:text-white transition-colors text-sm">Контакты</a></li>
                  <li><a href="#" className="text-slate-400 hover:text-white transition-colors text-sm flex items-center gap-2">
                    <MessageCircle className="w-4 h-4" /> Telegram
                  </a></li>
                </ul>
              </div>

              <div>
                <h4 className="font-semibold text-white mb-4">Юридическая информация</h4>
                <ul className="space-y-2">
                  <li><a href="#" className="text-slate-400 hover:text-white transition-colors text-sm">Политика конфиденциальности</a></li>
                  <li><a href="#" className="text-slate-400 hover:text-white transition-colors text-sm">Условия использования</a></li>
                  <li><a href="#" className="text-slate-400 hover:text-white transition-colors text-sm">AML/KYC</a></li>
                </ul>
              </div>
            </div>

            <div className="pt-8 border-t border-white/[0.08] flex flex-col md:flex-row items-center justify-between gap-4">
              <p className="text-slate-500 text-sm">
                © 2024 XPLR. Все права защищены.
              </p>
              <div className="flex items-center gap-4">
                <Headphones className="w-5 h-5 text-slate-500" />
                <span className="text-slate-400 text-sm">Поддержка 24/7</span>
              </div>
            </div>
          </div>
        </footer>
      </div>

      {/* CSS for floating animation */}
      <style>{`
        @keyframes float {
          0%, 100% {
            transform: perspective(1000px) rotateY(-5deg) rotateX(5deg) translateY(0);
          }
          50% {
            transform: perspective(1000px) rotateY(-5deg) rotateX(5deg) translateY(-20px);
          }
        }
        @keyframes fade-in {
          from { opacity: 0; transform: translateY(-10px); }
          to { opacity: 1; transform: translateY(0); }
        }
        .animate-fade-in {
          animation: fade-in 0.15s ease-out forwards;
        }
      `}</style>
    </div>
  );
};
