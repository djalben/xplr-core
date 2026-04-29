import { useState, useEffect, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useRates } from '../store/rates-context';
import { DashboardLayout } from '../components/dashboard-layout';
import { ModalPortal } from '../components/modal-portal';
import { BackButton } from '../components/back-button';
import { getWallet, transferWalletToCard } from '../services/wallet';
import { getUserCards, issuePersonalCard, getCardDetails, updateCardStatus, toggleAutoPay, getCardSubscriptions, toggleSubscription, getTelegramStatus, type Card as BackendCard, type CardSubscription } from '../services/cards';
import { getTierInfo, type TierInfo } from '../services/tier';
import { 
  Plus, 
  CreditCard as CardIcon,
  Wifi,
  Eye,
  Copy,
  Trash2,
  Check,
  X,
  Smartphone,
  Apple,
  ChevronRight,
  Banknote,
  ArrowUpDown,
  ShoppingBag,
  CreditCard,
  ChevronDown,
  Shield,
  Bell,
  MessageSquare,
  ToggleLeft,
  ToggleRight,
  Store,
  AlertTriangle
} from 'lucide-react';

interface PersonalCard {
  id: string;
  type: 'subscriptions' | 'travel' | 'premium';
  name: string;
  holderName: string;
  number: string;
  expiry: string;
  cvv: string;
  balance: number;
  currency: string;
  cardNetwork: 'visa' | 'mastercard';
  color: 'blue' | 'purple' | 'gold';
  price: string;
  bin: string;
  last4: string;
  isAutoPayEnabled: boolean;
}

// Professional Visa Logo
const VisaLogo = ({ className = "h-8 w-auto", color = "#1A1F71" }: { className?: string; color?: string }) => (
  <svg viewBox="0 0 780 500" className={className} preserveAspectRatio="xMidYMid meet">
    <path 
      d="M293.2 348.7l33.4-195.2h53.3l-33.4 195.2h-53.3zM534.4 157.6c-10.5-4.1-27.1-8.5-47.7-8.5-52.6 0-89.7 26.3-90 64-0.3 27.9 26.4 43.4 46.6 52.7 20.7 9.5 27.7 15.6 27.6 24.1-0.1 13-16.6 18.9-31.9 18.9-21.3 0-32.6-2.9-50.1-10.2l-6.9-3.1-7.5 43.6c12.4 5.4 35.4 10.1 59.3 10.4 55.9 0 92.2-26 92.7-66.2 0.3-22.1-14-38.9-44.6-52.8-18.6-9-30-15-29.9-24.1 0-8.1 9.7-16.7 30.5-16.7 17.4-0.3 30.1 3.5 39.9 7.4l4.8 2.2 7.2-41.7zM651.4 153.5h-41.2c-12.8 0-22.3 3.5-27.9 16.2l-79.2 178h56l11.2-29.2h68.4c1.6 6.8 6.5 29.2 6.5 29.2h49.5l-43.3-194.2zm-65.8 125.3c4.4-11.2 21.4-54.5 21.4-54.5-0.3 0.5 4.4-11.3 7.1-18.7l3.6 16.9s10.3 46.9 12.4 56.3h-44.5zM231.4 153.5l-52.2 133.2-5.5-27c-9.7-30.9-39.8-64.4-73.5-81.2l47.6 169h56.5l84.1-194h-57z" 
      fill={color}
    />
    <path 
      d="M131.9 153.5H46.6l-0.7 4.1c67 16.1 111.4 55 129.7 101.7l-18.7-89.6c-3.2-12.2-12.6-15.8-25-16.2z" 
      fill="#F7B600"
    />
  </svg>
);

// Professional Mastercard Logo
export const MastercardLogo = ({ className = "h-8 w-auto" }: { className?: string }) => (
  <svg viewBox="0 0 780 500" className={className} preserveAspectRatio="xMidYMid meet">
    <circle cx="250" cy="250" r="150" fill="#EB001B"/>
    <circle cx="530" cy="250" r="150" fill="#F79E1B"/>
    <path 
      d="M390 120.8C351.5 152.6 326 200.5 326 250c0 49.5 25.5 97.4 64 129.2 38.5-31.8 64-79.7 64-129.2 0-49.5-25.5-97.4-64-129.2z" 
      fill="#FF5F00"
    />
  </svg>
);

// Card Issue Modal
const CardIssueModal = ({ 
  card, 
  onClose,
  onIssue,
  isIssuing = false,
  walletBalance = 0
}: { 
  card: { type: string; name: string; price: string; currency: string; description: string; features: { title: string; items: string }[]; conditions: { label: string; value: string }[]; capabilities: { label: string; value: string; link?: boolean }[] };
  onClose: () => void;
  onIssue?: (cardType: 'subscriptions' | 'travel' | 'premium', priceUsd: number) => void;
  isIssuing?: boolean;
  walletBalance?: number;
}) => {
  const [showProhibited, setShowProhibited] = useState(false);
  const [selectedCurrency, setSelectedCurrency] = useState<'USD' | 'EUR'>(card.currency === 'EUR' ? 'EUR' : 'USD');
  const { rates } = useRates();

  const currentRate = selectedCurrency === 'USD' ? rates.usd : rates.eur;
  const currencySymbol = selectedCurrency === 'USD' ? '$' : '€';
  
  const { t } = useTranslation();
  const prohibitedOperations = [
    t('cards.prohibitedList.finance'),
    t('cards.prohibitedList.crypto'),
    t('cards.prohibitedList.gambling'),
    t('cards.prohibitedList.adult'),
    t('cards.prohibitedList.giftCards'),
    t('cards.prohibitedList.russian'),
  ];

  // Determine which payment methods are available
  const hasApplePay = card.type === 'travel' || card.type === 'premium';
  const hasGooglePay = true; // All cards support Google Pay

  return (
    <ModalPortal>
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
      <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 rounded-2xl w-full max-w-[440px] max-h-[90dvh] flex flex-col animate-scale-in shadow-2xl shadow-black/60">
        {/* Fixed header */}
        <div className="shrink-0 p-5 pb-3 border-b border-white/[0.06]">
          <div className="flex justify-center mb-3">
            {card.type === 'subscriptions' && <SubscriptionsCardVisual mini={false} currencySymbol={currencySymbol} />}
            {card.type === 'travel' && <TravelCardVisual mini={false} currencySymbol={currencySymbol} />}
            {card.type === 'premium' && <PremiumCardVisual mini={false} currencySymbol={currencySymbol} />}
          </div>
          
          {/* Payment method badges */}
          <div className="flex justify-center gap-3 mb-3">
            {hasApplePay && (
              <div className="px-3 py-1 bg-white/10 rounded-lg flex items-center gap-1.5 border border-white/5">
                <Apple className="w-3.5 h-3.5" />
                <span className="text-[11px] font-medium">Pay</span>
              </div>
            )}
            {hasGooglePay && (
              <div className="px-3 py-1 bg-white/10 rounded-lg flex items-center gap-1.5 border border-white/5">
                <span className="text-xs font-medium text-blue-400">G</span>
                <span className="text-[11px] font-medium">Pay</span>
              </div>
            )}
          </div>
          
          <h2 className="text-lg font-bold text-white text-center mb-0.5">{card.name}</h2>
          <p className="text-slate-400 text-xs text-center mb-2">{card.description}</p>
          <p className="text-xl font-bold text-blue-400 text-center mb-2">{card.price}</p>
          
          {/* Currency selector */}
          <div className="flex items-center justify-center gap-2 mb-2">
            <button
              onClick={() => setSelectedCurrency('USD')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                selectedCurrency === 'USD'
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                  : 'bg-white/5 text-slate-400 border border-white/10 hover:bg-white/10'
              }`}
            >
              $ USD
            </button>
            <button
              onClick={() => setSelectedCurrency('EUR')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                selectedCurrency === 'EUR'
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                  : 'bg-white/5 text-slate-400 border border-white/10 hover:bg-white/10'
              }`}
            >
              € EUR
            </button>
          </div>

          {/* Exchange rate - dynamic */}
          <p className="text-center text-xs text-slate-400">
            {card.type === 'premium' ? t('cards.bestRate') : t('cards.currentRate')} <span className="text-blue-400 font-medium">{currencySymbol}1 = {currentRate.toFixed(2)} ₽</span>
          </p>
        </div>
        
        {/* Scrollable body */}
        <div className="flex-1 overflow-y-auto min-h-0 px-5 py-3 space-y-3">
          {/* Features */}
          {card.features.map((feature, i) => (
            <div key={i} className="flex items-start gap-2.5 p-2.5 bg-white/5 rounded-xl">
              <div className="w-5 h-5 rounded-full bg-blue-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                <Check className="w-3 h-3 text-blue-400" />
              </div>
              <div>
                <p className="text-white font-medium text-xs">{feature.title}</p>
                <p className="text-slate-400 text-[11px]">{feature.items}</p>
              </div>
            </div>
          ))}
        
          {/* Conditions */}
          <div>
            <h3 className="text-white font-semibold text-sm mb-2">{t('cards.issueConditions')}</h3>
            <ul className="space-y-1.5">
              {card.conditions.map((cond, i) => (
                <li key={i} className="flex items-center gap-2 text-xs">
                  <span className="text-slate-400">•</span>
                  <span className="text-slate-300">{cond.label} – <span className="text-white">{cond.value}</span></span>
                </li>
              ))}
            </ul>
          </div>
        
          {/* Capabilities */}
          <div>
            <h3 className="text-white font-semibold text-sm mb-2">{t('cards.capabilities')}</h3>
            <ul className="space-y-1.5">
              {card.capabilities.map((cap, i) => (
                <li key={i} className="flex items-center gap-2 text-xs">
                  <span className="text-slate-400">•</span>
                  <span className="text-slate-300">{cap.label} – {cap.link ? (
                    <span className="text-blue-400 cursor-pointer hover:underline">{cap.value}</span>
                  ) : (
                    <span className="text-white">{cap.value}</span>
                  )}</span>
                </li>
              ))}
            </ul>
          </div>
        
          {/* Prohibited Operations */}
          <div>
            <button 
              onClick={() => setShowProhibited(!showProhibited)}
              className="flex items-center justify-between w-full py-2 border-t border-white/10"
            >
              <span className="text-white font-semibold text-sm">{t('cards.prohibitedOps')}</span>
              <ChevronDown className={`w-4 h-4 text-slate-400 transition-transform ${showProhibited ? 'rotate-180' : ''}`} />
            </button>
            {showProhibited && (
              <ul className="space-y-1.5 pb-2 animate-fade-in">
                {prohibitedOperations.map((op, i) => (
                  <li key={i} className="flex items-center gap-2 text-xs">
                    <span className="text-slate-400">•</span>
                    <span className="text-slate-400">{op}</span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
        
        {/* Fixed footer button */}
        <div className="shrink-0 p-5 pt-3 border-t border-white/[0.06]">
          {/* Wallet balance info */}
          <div className="flex items-center justify-between mb-3 px-1">
            <span className="text-xs text-slate-400">У вас на балансе:</span>
            <span className={`text-sm font-bold ${walletBalance > 0 ? 'text-emerald-400' : 'text-red-400'}`}>
              ${walletBalance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </span>
          </div>
          <button 
            onClick={() => {
              if (onIssue) {
                const priceNum = parseFloat(card.price.replace(/[^0-9.]/g, '')) || 0;
                onIssue(card.type as 'subscriptions' | 'travel' | 'premium', priceNum);
              } else {
                onClose();
              }
            }}
            disabled={isIssuing}
            className="w-full py-3.5 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isIssuing ? 'Выпускаем карту...' : `${t('cards.issueCard')} ${card.price}`}
          </button>
        </div>
      </div>
    </div>
    </ModalPortal>
  );
};

// Subscriptions Card Visual - colorful with service icons (realistic bank card style)
export const SubscriptionsCardVisual = ({ mini = true, currencySymbol, bin, last4 }: { mini?: boolean; currencySymbol?: string; bin?: string; last4?: string }) => (
  <div className={`relative ${mini ? 'w-full aspect-[1.586/1]' : 'w-72 h-44'} rounded-2xl overflow-hidden shadow-2xl`}>
    {/* Gradient background */}
    <div className="absolute inset-0">
      <div className="absolute inset-0 bg-gradient-to-br from-pink-500 via-purple-500 to-blue-600" />
    </div>

    {/* Digital service icons overlay — thin contour watermark */}
    <div className="absolute inset-0 opacity-[0.10] pointer-events-none">
      <svg viewBox="0 0 280 170" className="w-full h-full" preserveAspectRatio="xMidYMid slice">
        {/* Play triangle */}
        <polygon points="30,25 30,50 50,37.5" fill="none" stroke="white" strokeWidth="0.8" />
        {/* Pause bars */}
        <rect x="220" y="20" width="4" height="18" rx="1" fill="none" stroke="white" strokeWidth="0.7" />
        <rect x="228" y="20" width="4" height="18" rx="1" fill="none" stroke="white" strokeWidth="0.7" />
        {/* Cloud */}
        <path d="M140,35 a12,12 0 0,1 24,0 a10,10 0 0,1 10,10 h-44 a10,10 0 0,1 10,-10z" fill="none" stroke="white" strokeWidth="0.7" />
        {/* Music note */}
        <path d="M70,90 v-22 l16,-5 v22" fill="none" stroke="white" strokeWidth="0.7" />
        <circle cx="70" cy="90" r="4" fill="none" stroke="white" strokeWidth="0.7" />
        <circle cx="86" cy="85" r="4" fill="none" stroke="white" strokeWidth="0.7" />
        {/* Wi-Fi arcs */}
        <path d="M200,120 a8,8 0 0,1 16,0" fill="none" stroke="white" strokeWidth="0.6" />
        <path d="M196,115 a14,14 0 0,1 24,0" fill="none" stroke="white" strokeWidth="0.6" />
        <circle cx="208" cy="123" r="1.5" fill="white" />
        {/* Code brackets */}
        <path d="M250,80 l-8,12 l8,12" fill="none" stroke="white" strokeWidth="0.7" />
        <path d="M262,80 l8,12 l-8,12" fill="none" stroke="white" strokeWidth="0.7" />
        {/* Small play */}
        <polygon points="120,120 120,140 135,130" fill="none" stroke="white" strokeWidth="0.6" />
        {/* Gear */}
        <circle cx="40" cy="130" r="7" fill="none" stroke="white" strokeWidth="0.6" />
        <circle cx="40" cy="130" r="3" fill="none" stroke="white" strokeWidth="0.5" />
      </svg>
    </div>

    {/* Card content */}
    <div className="relative h-full p-4 flex flex-col justify-between">
      {/* Top row - branding and currency */}
      <div className="flex items-start justify-between">
        <div>
          <p className="text-white/90 text-sm font-bold tracking-[0.2em] leading-none">XPLR</p>
          <p className="text-white/60 text-[7px] font-light tracking-[0.25em] uppercase leading-none mt-0.5">Explorer</p>
        </div>
        <span className="text-white text-sm font-bold">{currencySymbol ?? '€'}</span>
      </div>
      
      {/* Card number at bottom */}
      <div className="mt-auto">
        <p className="text-white/40 text-[7px] font-light tracking-[0.15em] uppercase leading-none mb-1.5">БЕЗ ГРАНИЦ</p>
        <p className="text-white/50 text-[10px] mb-0.5">Card number</p>
        <p className="text-white font-mono text-sm tracking-widest">{bin ? `${bin.slice(0, 4)} ${bin.slice(4, 6)}** **** ${last4 || '****'}` : '**** **** **** ****'}</p>
      </div>
      
      {/* Mastercard logo at bottom right */}
      <div className="absolute bottom-4 right-4">
        <MastercardLogo className="h-7 w-auto" />
      </div>
    </div>
  </div>
);

// Travel Card Visual - blue gradient (realistic bank card style)
export const TravelCardVisual = ({ mini = true, currencySymbol, bin, last4 }: { mini?: boolean; currencySymbol?: string; bin?: string; last4?: string }) => (
  <div className={`relative ${mini ? 'w-full aspect-[1.586/1]' : 'w-72 h-44'} rounded-2xl overflow-hidden shadow-2xl`}>
    {/* Blue gradient background */}
    <div className="absolute inset-0">
      <div className="absolute inset-0 bg-gradient-to-br from-blue-400 via-blue-500 to-blue-700" />
    </div>

    {/* Travel silhouettes overlay — thin contour watermark */}
    <div className="absolute inset-0 opacity-[0.10] pointer-events-none">
      <svg viewBox="0 0 280 170" className="w-full h-full" preserveAspectRatio="xMidYMid slice">
        {/* Suitcase */}
        <rect x="30" y="55" width="28" height="22" rx="3" fill="none" stroke="white" strokeWidth="0.8" />
        <rect x="38" y="48" width="12" height="8" rx="2" fill="none" stroke="white" strokeWidth="0.7" />
        {/* Palm tree */}
        <line x1="220" y1="130" x2="220" y2="85" stroke="white" strokeWidth="0.8" />
        <path d="M220,85 q-18,-5 -22,-20" fill="none" stroke="white" strokeWidth="0.7" />
        <path d="M220,85 q18,-5 22,-20" fill="none" stroke="white" strokeWidth="0.7" />
        <path d="M220,88 q-20,2 -26,-12" fill="none" stroke="white" strokeWidth="0.6" />
        <path d="M220,88 q20,2 26,-12" fill="none" stroke="white" strokeWidth="0.6" />
        {/* Sun */}
        <circle cx="130" cy="35" r="12" fill="none" stroke="white" strokeWidth="0.8" />
        <line x1="130" y1="18" x2="130" y2="13" stroke="white" strokeWidth="0.6" />
        <line x1="130" y1="52" x2="130" y2="57" stroke="white" strokeWidth="0.6" />
        <line x1="113" y1="35" x2="108" y2="35" stroke="white" strokeWidth="0.6" />
        <line x1="147" y1="35" x2="152" y2="35" stroke="white" strokeWidth="0.6" />
        <line x1="118" y1="23" x2="115" y2="20" stroke="white" strokeWidth="0.5" />
        <line x1="142" y1="23" x2="145" y2="20" stroke="white" strokeWidth="0.5" />
        <line x1="118" y1="47" x2="115" y2="50" stroke="white" strokeWidth="0.5" />
        <line x1="142" y1="47" x2="145" y2="50" stroke="white" strokeWidth="0.5" />
        {/* Airplane */}
        <path d="M60,120 l30,-15 l-5,5 l15,0 l-30,15 l5,-5 l-15,0z" fill="none" stroke="white" strokeWidth="0.7" />
        {/* Compass circle */}
        <circle cx="250" cy="45" r="10" fill="none" stroke="white" strokeWidth="0.6" />
        <line x1="250" y1="37" x2="250" y2="53" stroke="white" strokeWidth="0.5" />
        <line x1="242" y1="45" x2="258" y2="45" stroke="white" strokeWidth="0.5" />
      </svg>
    </div>

    {/* Card content */}
    <div className="relative h-full p-4 flex flex-col justify-between">
      {/* Top row - branding and currency */}
      <div className="flex items-start justify-between">
        <div>
          <p className="text-white/90 text-sm font-bold tracking-[0.2em] leading-none">XPLR</p>
          <p className="text-white/60 text-[7px] font-light tracking-[0.25em] uppercase leading-none mt-0.5">Explorer</p>
        </div>
        <span className="text-white text-sm font-bold">{currencySymbol ?? '$'}</span>
      </div>
      
      {/* Card number at bottom */}
      <div className="mt-auto">
        <p className="text-white/40 text-[7px] font-light tracking-[0.15em] uppercase leading-none mb-1.5">БЕЗ ГРАНИЦ</p>
        <p className="text-white/50 text-[10px] mb-0.5">Card number</p>
        <p className="text-white font-mono text-sm tracking-widest">{bin ? `${bin.slice(0, 4)} ${bin.slice(4, 6)}** **** ${last4 || '****'}` : '**** **** **** ****'}</p>
      </div>
      
      {/* Mastercard logo at bottom right */}
      <div className="absolute bottom-4 right-4">
        <MastercardLogo className="h-7 w-auto" />
      </div>
    </div>
  </div>
);

// Premium Card Visual - XPLR PRIME: deep black matte, neural texture, platinum chip, Power Beam
export const PremiumCardVisual = ({ mini = true, currencySymbol, bin, last4 }: { mini?: boolean; currencySymbol?: string; bin?: string; last4?: string }) => (
  <div className={`relative ${mini ? 'w-full aspect-[1.586/1]' : 'w-72 h-44'} rounded-2xl overflow-hidden shadow-2xl`}>
    {/* Deep matte black background */}
    <div className="absolute inset-0 bg-gradient-to-br from-[#0a0a0a] via-[#111111] to-[#080808]">
      {/* Subtle neural/geometric mesh texture */}
      <div className="absolute inset-0 opacity-[0.12]">
        <svg viewBox="0 0 200 120" className="w-full h-full" preserveAspectRatio="xMidYMid slice">
          <defs>
            <pattern id="neural-mesh" width="16" height="16" patternUnits="userSpaceOnUse">
              <path d="M 16 0 L 0 0 0 16" fill="none" stroke="rgba(255,255,255,0.15)" strokeWidth="0.25"/>
              <circle cx="0" cy="0" r="0.6" fill="rgba(255,255,255,0.2)" />
              <circle cx="16" cy="16" r="0.6" fill="rgba(255,255,255,0.2)" />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#neural-mesh)" />
          {/* Neural connection lines */}
          <line x1="20" y1="20" x2="60" y2="40" stroke="rgba(255,255,255,0.06)" strokeWidth="0.3" />
          <line x1="60" y1="40" x2="100" y2="25" stroke="rgba(255,255,255,0.06)" strokeWidth="0.3" />
          <line x1="100" y1="25" x2="140" y2="50" stroke="rgba(255,255,255,0.06)" strokeWidth="0.3" />
          <line x1="140" y1="50" x2="180" y2="35" stroke="rgba(255,255,255,0.06)" strokeWidth="0.3" />
          <line x1="30" y1="70" x2="70" y2="90" stroke="rgba(255,255,255,0.05)" strokeWidth="0.3" />
          <line x1="70" y1="90" x2="120" y2="75" stroke="rgba(255,255,255,0.05)" strokeWidth="0.3" />
          <line x1="120" y1="75" x2="170" y2="95" stroke="rgba(255,255,255,0.05)" strokeWidth="0.3" />
        </svg>
      </div>

      {/* ── Power Beam — vertical red energy line ── */}
      <div className="absolute top-0 bottom-0 right-10 w-[3px] bg-gradient-to-b from-transparent via-red-600 to-transparent" />
      {/* Glow layer around the beam */}
      <div className="absolute top-0 bottom-0 right-[38px] w-[7px] bg-gradient-to-b from-transparent via-red-500/40 to-transparent blur-sm" />
      <div className="absolute top-0 bottom-0 right-[36px] w-[15px] bg-gradient-to-b from-transparent via-red-500/15 to-transparent blur-md" />
    </div>

    {/* Card content */}
    <div className="relative h-full p-4 flex flex-col justify-between">
      {/* Top row — XPLR PRIME branding and currency */}
      <div className="flex items-start justify-between">
        <div>
          <p className="text-white font-bold text-lg tracking-[0.25em] leading-none">XPLR PRIME</p>
          <p className="text-white/40 text-[7px] font-light tracking-[0.25em] uppercase leading-none mt-0.5">Explorer</p>
        </div>
        <div className="w-8 h-8 rounded-full bg-white/[0.06] backdrop-blur-sm flex items-center justify-center border border-white/[0.08]">
          <span className="text-white text-sm font-bold">{currencySymbol ?? '$'}</span>
        </div>
      </div>

      {/* Platinum / Silver EMV Chip */}
      <div className="w-10 h-7 rounded-md bg-gradient-to-br from-slate-300 via-slate-200 to-slate-400 mt-1 shadow-lg shadow-white/10">
        <div className="w-full h-full flex">
          <div className="w-1/3 border-r border-slate-400/50" />
          <div className="w-1/3 border-r border-slate-400/50 flex flex-col">
            <div className="h-1/2 border-b border-slate-400/50" />
            <div className="h-1/2" />
          </div>
          <div className="w-1/3" />
        </div>
      </div>

      {/* Card number at bottom */}
      <div className="mt-auto">
        <p className="text-white/25 text-[7px] font-light tracking-[0.15em] uppercase leading-none mb-1.5">БЕЗ ГРАНИЦ</p>
        <p className="text-white/25 text-[10px] mb-0.5">Card number</p>
        <p className="text-white/90 font-mono text-sm tracking-widest">{bin ? `${bin.slice(0, 4)} ${bin.slice(4, 6)}** **** ${last4 || '****'}` : '**** **** **** ****'}</p>
      </div>

      {/* Mastercard logo at bottom right */}
      <div className="absolute bottom-4 right-4">
        <MastercardLogo className="h-7 w-auto" />
      </div>
    </div>
  </div>
);

// Card Type Selection for Personal cards
const PersonalCardTypeCard = ({ 
  type, 
  name, 
  description, 
  currency, 
  currencySymbol,
  price, 
  usdRate,
  eurRate,
  onSelect 
}: { 
  type: 'subscriptions' | 'travel' | 'premium';
  name: string; 
  description: string; 
  currency: string;
  currencySymbol: string;
  price: string;
  usdRate: number;
  eurRate: number;
  onSelect: () => void;
}) => {
  const { t } = useTranslation();
  return (
    <div className="relative rounded-2xl border border-white/[0.08] bg-[#0d0d14]/80 backdrop-blur-xl overflow-hidden flex flex-col">
      {/* Accent top edge */}
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-blue-500/40 to-transparent" />

      {/* Card preview */}
      <div className="p-4 pb-0">
        {type === 'subscriptions' && <SubscriptionsCardVisual mini={true} />}
        {type === 'travel' && <TravelCardVisual mini={true} />}
        {type === 'premium' && <PremiumCardVisual mini={true} />}
      </div>

      {/* Info block */}
      <div className="p-4 pt-4 flex-1 flex flex-col">
        <h4 className="text-white font-semibold text-center mb-1">{name}</h4>
        <p className="text-slate-400 text-xs text-center mb-4">{description}</p>
        
        {/* Price and rate */}
        <div className="grid grid-cols-2 gap-2 mb-5">
          <div className="p-2.5 bg-white/[0.04] rounded-xl text-center border border-white/[0.06]">
            <p className="text-blue-400 font-bold text-lg">{price}</p>
            <p className="text-slate-500 text-[10px]">{t('cards.cost')}</p>
          </div>
          <div className="p-2.5 bg-white/[0.04] rounded-xl text-center flex flex-col items-center justify-center border border-white/[0.06]">
            <p className="text-white font-medium text-[11px]">USD: {usdRate.toFixed(2)} ₽</p>
            <p className="text-white font-medium text-[11px]">EUR: {eurRate.toFixed(2)} ₽</p>
            <p className="text-slate-500 text-[10px] mt-0.5">{t('cards.rate')}</p>
          </div>
        </div>
        
        {/* Button strictly under info */}
        <button 
          onClick={onSelect}
          className="w-full py-3.5 bg-gradient-to-r from-blue-500 to-indigo-600 hover:from-blue-400 hover:to-indigo-500 text-white font-semibold rounded-xl transition-all text-sm shadow-lg shadow-blue-500/20 hover:shadow-blue-500/30 mt-auto"
        >
          {t('cards.issueCard')} {price}
        </button>
      </div>
    </div>
  );
};

// Inline SVG bank logos — crisp at any resolution
const BankLogos: Record<string, React.ReactNode> = {
  sbp: (
    <svg viewBox="0 0 40 40" className="w-8 h-8">
      <defs>
        <linearGradient id="sbp-g" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#5B57A2" />
          <stop offset="35%" stopColor="#D90751" />
          <stop offset="65%" stopColor="#FAB718" />
          <stop offset="100%" stopColor="#0FA8D6" />
        </linearGradient>
      </defs>
      <rect rx="8" width="40" height="40" fill="url(#sbp-g)" />
      <text x="20" y="26" textAnchor="middle" fill="white" fontSize="14" fontWeight="700" fontFamily="system-ui">СБП</text>
    </svg>
  ),
  sber: (
    <svg viewBox="0 0 40 40" className="w-8 h-8">
      <circle cx="20" cy="20" r="20" fill="#21A038" />
      <path d="M20 8 L20 20 L30 20" stroke="white" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round" fill="none" />
      <circle cx="20" cy="20" r="11" stroke="white" strokeWidth="2.5" fill="none" />
    </svg>
  ),
  tbank: (
    <svg viewBox="0 0 40 40" className="w-8 h-8">
      <rect rx="8" width="40" height="40" fill="#FFDD2D" />
      <text x="20" y="27" textAnchor="middle" fill="#333" fontSize="20" fontWeight="800" fontFamily="system-ui">T</text>
    </svg>
  ),
  alfa: (
    <svg viewBox="0 0 40 40" className="w-8 h-8">
      <rect rx="8" width="40" height="40" fill="#EF3124" />
      <text x="20" y="28" textAnchor="middle" fill="white" fontSize="22" fontWeight="800" fontFamily="system-ui">A</text>
    </svg>
  ),
  vtb: (
    <svg viewBox="0 0 40 40" className="w-8 h-8">
      <rect rx="8" width="40" height="40" fill="#002882" />
      <rect x="8" y="12" width="24" height="3.5" rx="1.5" fill="white" />
      <rect x="8" y="18.5" width="24" height="3.5" rx="1.5" fill="white" />
      <rect x="8" y="25" width="24" height="3.5" rx="1.5" fill="white" />
    </svg>
  ),
};

// Bank Logo Button — inline SVG, unified size, hover/active
const BankLogoButton = ({
  bank,
  selected,
  onClick,
}: {
  bank: { id: string; name: string };
  selected: boolean;
  onClick: () => void;
}) => (
  <button
    onClick={onClick}
    className={`
      flex flex-col items-center gap-1.5 p-2 rounded-xl transition-all duration-150 cursor-pointer
      ${selected
        ? 'bg-blue-500/10 border-2 border-blue-500 scale-105 shadow-lg shadow-blue-500/20'
        : 'bg-white/[0.03] border border-white/[0.08] hover:bg-white/[0.06] hover:border-white/15 active:scale-95'
      }
    `}
  >
    <div className="w-8 h-8 rounded-lg flex items-center justify-center overflow-hidden shrink-0">
      {BankLogos[bank.id] ?? <span className="text-white font-bold text-xs">{bank.name[0]}</span>}
    </div>
    <span className={`text-[10px] font-medium leading-tight ${selected ? 'text-white' : 'text-slate-400'}`}>
      {bank.name}
    </span>
  </button>
);

// Wallet-to-Card Transfer Modal — internal transfer, no banks/СБП
const WalletTopUpModal = ({ card, walletBalance, onClose, onTransfer }: {
  card: PersonalCard;
  walletBalance: number;      // master_balance in USD
  onClose: () => void;
  onTransfer: (amount: number) => void; // amount in card's currency
}) => {
  const { rates } = useRates();
  const [amount, setAmount] = useState('');
  const [isTransferring, setIsTransferring] = useState(false);

  const currencySymbol = card.currency;
  // walletBalance is in USD; convert to card currency
  const availableInCurrency = card.currency === '€'
    ? (rates.eur > 0 ? walletBalance * rates.usd / rates.eur : 0)
    : walletBalance; // USD → USD, no conversion

  const numAmount = parseFloat(amount) || 0;
  const isInsufficient = numAmount > 0 && numAmount > availableInCurrency;
  const canTransfer = numAmount > 0 && !isInsufficient && !isTransferring;

  const presets = [10, 25, 50, 100, 250, 500];

  const handleTransfer = () => {
    if (!canTransfer) return;
    setIsTransferring(true);
    onTransfer(numAmount);
  };

  return (
    <ModalPortal>
    <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center p-0 sm:p-4">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-md" onClick={onClose} />

      <div className="relative bg-[#0b0b14]/95 border border-white/[0.10] rounded-t-2xl sm:rounded-2xl w-full max-w-[380px] flex flex-col animate-scale-in shadow-[0_24px_80px_-12px_rgba(0,0,0,0.8)]">
        {/* Glass accent */}
        <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-blue-400/40 to-transparent rounded-t-2xl" />

        {/* Header */}
        <div className="px-5 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-emerald-500/20 to-blue-500/20 border border-emerald-500/30 flex items-center justify-center">
              <ArrowUpDown className="w-4 h-4 text-emerald-400" />
            </div>
            <div>
              <h3 className="text-base font-semibold text-white leading-tight">Перевод на карту</h3>
              <p className="text-[11px] text-slate-400">{card.name}</p>
            </div>
          </div>
          <button onClick={onClose} className="p-1.5 hover:bg-white/10 rounded-lg transition-colors">
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Body */}
        <div className="px-5 pb-2 space-y-4">
          {/* Wallet balance info */}
          <div className="p-3 rounded-xl bg-white/[0.04] border border-white/[0.06]">
            <div className="flex items-center justify-between">
              <span className="text-xs text-slate-400">Доступно в Кошельке:</span>
              <span className="text-white font-bold text-sm">
                {currencySymbol}{availableInCurrency.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </span>
            </div>
          </div>

          {/* Amount input */}
          <div>
            <label className="block text-xs text-slate-400 mb-1.5">Сумма перевода</label>
            <div className="relative">
              <span className={`absolute left-4 top-1/2 -translate-y-1/2 text-lg font-bold ${isInsufficient ? 'text-red-400' : 'text-blue-400'}`}>
                {currencySymbol}
              </span>
              <input
                type="number"
                inputMode="decimal"
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className={`w-full h-14 pl-12 pr-4 bg-white/[0.04] rounded-xl text-white text-2xl font-bold focus:outline-none transition-colors placeholder:text-slate-600 ${
                  isInsufficient
                    ? 'border-2 border-red-500 focus:border-red-500 focus:ring-1 focus:ring-red-500/50 text-red-300'
                    : 'border border-white/[0.10] focus:border-blue-400 focus:ring-1 focus:ring-blue-400/50'
                }`}
              />
            </div>
            {isInsufficient && (
              <p className="text-red-400 text-xs mt-1.5 font-medium">Недостаточно средств в Кошельке</p>
            )}
          </div>

          {/* Quick presets */}
          <div className="flex flex-wrap gap-1.5">
            {presets.map(p => (
              <button
                key={p}
                onClick={() => setAmount(String(p))}
                className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                  parseFloat(amount) === p
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                    : 'bg-white/[0.04] text-slate-400 border border-white/[0.08] hover:bg-white/[0.08]'
                }`}
              >
                {currencySymbol}{p}
              </button>
            ))}
          </div>
        </div>

        {/* Transfer button — large, thumb-friendly */}
        <div className="px-5 pt-3 pb-5">
          <button
            onClick={handleTransfer}
            disabled={!canTransfer}
            className="w-full py-4 bg-gradient-to-r from-blue-500 to-indigo-600 hover:from-blue-400 hover:to-indigo-500 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/20 hover:shadow-blue-500/30 disabled:opacity-40 disabled:cursor-not-allowed text-base"
          >
            {isTransferring ? 'Переводим...' : `Перевести на карту${numAmount > 0 ? ` ${currencySymbol}${numAmount.toLocaleString()}` : ''}`}
          </button>
        </div>
      </div>
    </div>
    </ModalPortal>
  );
};

// ══════════════════════════════════════════════════════════════
// Subscription Dashboard — auto-pay toggle + tracked merchants
// ══════════════════════════════════════════════════════════════
const SubscriptionDashboard = ({ cardId }: { cardId: number }) => {
  const [subs, setSubs] = useState<CardSubscription[]>([]);
  const [autoPayEnabled, setAutoPayEnabled] = useState(true);
  const [loading, setLoading] = useState(true);
  const [toggling, setToggling] = useState<number | null>(null);

  const fetchSubs = useCallback(async () => {
    try {
      const data = await getCardSubscriptions(cardId);
      setSubs(data.subscriptions || []);
      setAutoPayEnabled(data.is_auto_pay_enabled);
    } catch { /* ignore */ }
    setLoading(false);
  }, [cardId]);

  useEffect(() => { fetchSubs(); }, [fetchSubs]);

  const handleToggleAutoPay = async () => {
    const next = !autoPayEnabled;
    setAutoPayEnabled(next);
    try {
      await toggleAutoPay(cardId, next);
    } catch {
      setAutoPayEnabled(!next);
    }
  };

  const handleToggleSub = async (sub: CardSubscription) => {
    setToggling(sub.id);
    const next = !sub.is_allowed;
    setSubs(prev => prev.map(s => s.id === sub.id ? { ...s, is_allowed: next } : s));
    try {
      await toggleSubscription(cardId, sub.id, next);
    } catch {
      setSubs(prev => prev.map(s => s.id === sub.id ? { ...s, is_allowed: !next } : s));
    }
    setToggling(null);
  };

  if (loading) return null;

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2 mb-1">
        <Shield className="w-4 h-4 text-blue-400" />
        <h3 className="text-white font-semibold text-sm">Контроль подписок</h3>
      </div>

      {/* Global auto-pay toggle */}
      <div className="p-3 bg-white/5 rounded-xl border border-white/10 flex items-center justify-between">
        <div>
          <p className="text-white text-sm font-medium">Автосписания</p>
          <p className="text-xs text-slate-400">Разрешить recurring-платежи</p>
        </div>
        <button onClick={handleToggleAutoPay} className="transition-colors">
          {autoPayEnabled ? (
            <ToggleRight className="w-8 h-8 text-emerald-400" />
          ) : (
            <ToggleLeft className="w-8 h-8 text-red-400" />
          )}
        </button>
      </div>

      {!autoPayEnabled && (
        <div className="p-2.5 bg-red-500/10 rounded-lg border border-red-500/30">
          <p className="text-xs text-red-300">⛔ Все автосписания заблокированы. Подписки не смогут снимать средства с этой карты.</p>
        </div>
      )}

      {/* Tracked merchant subscriptions */}
      {subs.length > 0 && (
        <div className="space-y-2">
          <p className="text-xs text-slate-400 font-medium">Активные подписки</p>
          {subs.map(sub => (
            <div key={sub.id} className="p-2.5 bg-white/5 rounded-xl border border-white/10 flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0 flex-1">
                <Store className="w-4 h-4 text-slate-400 shrink-0" />
                <div className="min-w-0">
                  <p className="text-white text-sm truncate">{sub.merchant_name}</p>
                  <p className="text-xs text-slate-500">
                    {sub.last_currency === 'USD' ? '$' : sub.last_currency}{sub.last_amount} · {sub.charge_count}x
                  </p>
                </div>
              </div>
              <button
                onClick={() => handleToggleSub(sub)}
                disabled={toggling === sub.id}
                className="shrink-0 transition-colors"
              >
                {sub.is_allowed ? (
                  <ToggleRight className="w-7 h-7 text-emerald-400" />
                ) : (
                  <ToggleLeft className="w-7 h-7 text-red-400" />
                )}
              </button>
            </div>
          ))}
        </div>
      )}

      {subs.length === 0 && (
        <p className="text-xs text-slate-500 text-center py-2">Подписки появятся автоматически после первого списания</p>
      )}
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// 3DS Code Display — WebSocket real-time delivery
// ══════════════════════════════════════════════════════════════
const ThreeDSCodeDisplay = () => {
  const [code, setCode] = useState<{ code: string; merchant: string; ts: string } | null>(null);
  const [copied, setCopied] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) return;

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const host = import.meta.env.VITE_API_URL?.replace(/^https?:\/\//, '') || window.location.host;
    const wsUrl = `${proto}://${host}/api/v1/user/3ds-ws?token=${encodeURIComponent(token)}`;

    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === '3ds_code' && msg.data?.code) {
            setCode({
              code: msg.data.code,
              merchant: msg.data.merchant_name || '',
              ts: new Date().toLocaleTimeString(),
            });
            setCopied(false);
          }
        } catch { /* ignore parse errors */ }
      };

      ws.onerror = () => { /* silent */ };
      ws.onclose = () => { wsRef.current = null; };
    } catch { /* WebSocket not supported or blocked */ }

    return () => { wsRef.current?.close(); };
  }, []);

  const handleCopy = () => {
    if (!code) return;
    navigator.clipboard.writeText(code.code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (!code) return null;

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Bell className="w-4 h-4 text-amber-400 animate-pulse" />
        <h3 className="text-white font-semibold text-sm">3DS Код</h3>
      </div>
      <div className="p-3 bg-amber-500/10 rounded-xl border border-amber-500/30">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-amber-300 font-mono text-2xl font-bold tracking-widest">{code.code}</p>
            {code.merchant && <p className="text-xs text-amber-400/70 mt-1">{code.merchant}</p>}
            <p className="text-xs text-slate-500 mt-0.5">{code.ts}</p>
          </div>
          <button
            onClick={handleCopy}
            className="p-2 bg-amber-500/20 hover:bg-amber-500/30 rounded-lg transition-colors"
          >
            {copied ? <Check className="w-5 h-5 text-emerald-400" /> : <Copy className="w-5 h-5 text-amber-400" />}
          </button>
        </div>
        <p className="text-xs text-amber-400/60 mt-2">⚠️ Не сообщайте код третьим лицам</p>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// Mandatory Telegram Modal — blocks payment if TG not linked
// ══════════════════════════════════════════════════════════════
const TelegramRequiredModal = ({ onClose, onActivate }: { onClose: () => void; onActivate: () => void }) => (
  <ModalPortal>
    <div className="fixed inset-0 z-[110] flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
      <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-[420px] animate-scale-in shadow-2xl shadow-black/60">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-12 h-12 rounded-xl bg-blue-500/20 border border-blue-500/30 flex items-center justify-center">
            <MessageSquare className="w-6 h-6 text-blue-400" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-white">Telegram не привязан</h3>
            <p className="text-xs text-slate-400">Требуется для 3DS кодов</p>
          </div>
        </div>

        <div className="p-3 bg-amber-500/10 rounded-xl border border-amber-500/30 mb-4">
          <div className="flex items-start gap-2">
            <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
            <p className="text-sm text-amber-200">
              Для получения 3DS кодов безопасности необходимо привязать Telegram. Без этого оплата может быть невозможна.
            </p>
          </div>
        </div>

        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-3 bg-white/10 hover:bg-white/15 text-slate-300 font-medium rounded-xl transition-colors"
          >
            Позже
          </button>
          <button
            onClick={onActivate}
            className="flex-1 px-4 py-3 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-xl transition-colors flex items-center justify-center gap-2"
          >
            <MessageSquare className="w-4 h-4" />
            Привязать
          </button>
        </div>
      </div>
    </div>
  </ModalPortal>
);

// Card Details Modal with Billing Address
const CardDetailsModal = ({ 
  card, 
  onClose 
}: { 
  card: PersonalCard; 
  onClose: () => void;
}) => {
  const [copiedField, setCopiedField] = useState<string | null>(null);

  const handleCopy = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const billingAddress = "XPLR Tech Solutions LLC, 15/1 Vardanants St, Yerevan 0010, Armenia";

  return (
    <ModalPortal>
      <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
        <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
        
        <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 rounded-2xl w-full max-w-[440px] max-h-[90dvh] flex flex-col animate-scale-in shadow-2xl shadow-black/60">
          {/* Header */}
          <div className="shrink-0 p-5 pb-3 border-b border-white/[0.06]">
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-lg font-bold text-white">Card Details</h2>
              <button onClick={onClose} className="p-1.5 hover:bg-white/10 rounded-lg transition-colors">
                <X className="w-5 h-5 text-slate-400" />
              </button>
            </div>
            
            {/* Card preview */}
            <div className="flex justify-center mb-3">
              {card.type === 'subscriptions' && <SubscriptionsCardVisual mini={false} currencySymbol={card.currency} bin={card.bin} last4={card.last4} />}
              {card.type === 'travel' && <TravelCardVisual mini={false} currencySymbol={card.currency} bin={card.bin} last4={card.last4} />}
              {card.type === 'premium' && <PremiumCardVisual mini={false} currencySymbol={card.currency} bin={card.bin} last4={card.last4} />}
            </div>
          </div>

          {/* Scrollable body */}
          <div className="flex-1 overflow-y-auto min-h-0 px-5 py-4 space-y-4">
            {/* Card Information */}
            <div className="space-y-3">
              <h3 className="text-white font-semibold text-sm mb-3">Card Information</h3>
              
              {/* Card Number */}
              <div className="p-3 bg-white/5 rounded-xl border border-white/10">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-slate-400">Card Number</span>
                  <button 
                    onClick={() => handleCopy(card.number.replace(/\s/g, ''), 'number')}
                    className="p-1.5 hover:bg-white/10 rounded-lg transition-colors"
                  >
                    {copiedField === 'number' ? (
                      <Check className="w-4 h-4 text-emerald-400" />
                    ) : (
                      <Copy className="w-4 h-4 text-slate-400" />
                    )}
                  </button>
                </div>
                <p className="text-white font-mono text-base tracking-wider">{card.number}</p>
              </div>

              {/* Expiry and CVV */}
              <div className="grid grid-cols-2 gap-3">
                <div className="p-3 bg-white/5 rounded-xl border border-white/10">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs text-slate-400">Expiry Date</span>
                    <button 
                      onClick={() => handleCopy(card.expiry, 'expiry')}
                      className="p-1.5 hover:bg-white/10 rounded-lg transition-colors"
                    >
                      {copiedField === 'expiry' ? (
                        <Check className="w-4 h-4 text-emerald-400" />
                      ) : (
                        <Copy className="w-4 h-4 text-slate-400" />
                      )}
                    </button>
                  </div>
                  <p className="text-white font-mono text-base">{card.expiry}</p>
                </div>

                <div className="p-3 bg-white/5 rounded-xl border border-white/10">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs text-slate-400">CVV</span>
                    <button 
                      onClick={() => handleCopy(card.cvv, 'cvv')}
                      className="p-1.5 hover:bg-white/10 rounded-lg transition-colors"
                    >
                      {copiedField === 'cvv' ? (
                        <Check className="w-4 h-4 text-emerald-400" />
                      ) : (
                        <Copy className="w-4 h-4 text-slate-400" />
                      )}
                    </button>
                  </div>
                  <p className="text-white font-mono text-base">{card.cvv}</p>
                </div>
              </div>

              {/* Cardholder Name */}
              <div className="p-3 bg-white/5 rounded-xl border border-white/10">
                <span className="text-xs text-slate-400 block mb-1">Cardholder Name</span>
                <p className="text-white font-medium">{card.holderName}</p>
              </div>

              {/* Balance */}
              <div className="p-3 bg-emerald-500/10 rounded-xl border border-emerald-500/30">
                <span className="text-xs text-emerald-400 block mb-1">Current Balance</span>
                <p className="text-white font-bold text-xl">{card.currency}{card.balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</p>
              </div>
            </div>

            {/* Billing Address */}
            <div className="pt-3 border-t border-white/10">
              <h3 className="text-white font-semibold text-sm mb-3">Billing Address</h3>
              <div className="p-4 bg-blue-500/10 rounded-xl border border-blue-500/30">
                <p className="text-white text-sm leading-relaxed mb-3">{billingAddress}</p>
                <button 
                  onClick={() => handleCopy(billingAddress, 'address')}
                  className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-500/20 hover:bg-blue-500/30 text-blue-400 rounded-lg transition-colors text-sm font-medium border border-blue-500/40"
                >
                  {copiedField === 'address' ? (
                    <>
                      <Check className="w-4 h-4" />
                      Address Copied!
                    </>
                  ) : (
                    <>
                      <Copy className="w-4 h-4" />
                      Copy Address
                    </>
                  )}
                </button>
              </div>
              <p className="text-xs text-slate-400 mt-2 text-center">Use this address for billing when making online payments</p>
            </div>

            {/* Subscription Dashboard + Anti-Drain */}
            <SubscriptionDashboard cardId={Number(card.id)} />

            {/* 3DS Code Display */}
            <ThreeDSCodeDisplay />
          </div>

          {/* Footer */}
          <div className="shrink-0 p-5 pt-3 border-t border-white/[0.06]">
            <button 
              onClick={onClose}
              className="w-full py-3 bg-white/10 hover:bg-white/20 text-white font-semibold rounded-xl transition-all"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
};

// Realistic Credit Card display for "Your Cards" section
const RealisticCreditCard = ({ 
  card, 
  onClose, 
  onTopUp,
  onApplePay,
  onGooglePay,
  onToggleAutoPay
}: { 
  card: PersonalCard; 
  onClose: () => void;
  onTopUp: () => void;
  onApplePay: () => void;
  onGooglePay: () => void;
  onToggleAutoPay: (cardId: string, enabled: boolean) => void;
}) => {
  const { t } = useTranslation();
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  
  const canAddToWallet = card.type === 'travel' || card.type === 'premium';
  const showSubProtection = card.type === 'subscriptions' || card.type === 'premium';

  return (
    <>
      {showDetailsModal && (
        <CardDetailsModal card={card} onClose={() => setShowDetailsModal(false)} />
      )}
      
      <div className="group">
        {/* Card visual based on type */}
        <div className="relative">
          {card.type === 'subscriptions' && <SubscriptionsCardVisual mini={true} bin={card.bin} last4={card.last4} />}
          {card.type === 'travel' && <TravelCardVisual mini={true} bin={card.bin} last4={card.last4} />}
          {card.type === 'premium' && <PremiumCardVisual mini={true} bin={card.bin} last4={card.last4} />}
          
          {/* Overlay for card details */}
          <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl flex items-center justify-center gap-2">
            <button 
              onClick={() => setShowDetailsModal(true)}
              className="p-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors"
              title="View card details"
            >
              <Eye className="w-5 h-5 text-white" />
            </button>
          </div>
        </div>
        
        {/* Card Info */}
        <div className="mt-3 p-3 bg-white/5 rounded-xl">
          <div className="flex items-center justify-between mb-2">
            <span className="text-white font-medium text-sm">{card.name}</span>
            <span className="text-emerald-400 font-bold">{card.currency}{card.balance.toLocaleString()}</span>
          </div>
          <p className="text-xs text-slate-400">**** **** **** {card.number.slice(-4)}</p>
        </div>

      {/* Action Buttons */}
      <div className="mt-3 flex gap-2">
        <button 
          onClick={onTopUp}
          className="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 rounded-xl transition-colors text-sm border border-emerald-500/30"
        >
          <Banknote className="w-4 h-4" />
          Пополнить с кошелька
        </button>
        <button 
          onClick={onClose}
          className="px-3 py-2.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 rounded-xl transition-colors"
        >
          <X className="w-4 h-4" />
        </button>
      </div>
      
      {canAddToWallet && (
        <div className="mt-2 flex gap-2">
          <button onClick={onApplePay} className="flex-1 flex items-center justify-center gap-1 px-3 py-2 glass-card hover:bg-white/10 text-white rounded-lg transition-colors text-xs">
            <Apple className="w-3 h-3" />
            Apple Pay
          </button>
          <button onClick={onGooglePay} className="flex-1 flex items-center justify-center gap-1 px-3 py-2 glass-card hover:bg-white/10 text-white rounded-lg transition-colors text-xs">
            <Smartphone className="w-3 h-3" />
            Google Pay
          </button>
        </div>
      )}

      {/* Subscription Protection Toggle — only for subscriptions/premium, hidden for travel */}
      {showSubProtection && (
        <div className="mt-2 px-3 py-2 bg-white/[0.03] border border-white/[0.06] rounded-xl">
          <button
            onClick={() => onToggleAutoPay(card.id, !card.isAutoPayEnabled)}
            className="w-full flex items-center gap-2.5 group/toggle"
          >
            <div className={`shrink-0 w-7 h-7 rounded-lg flex items-center justify-center transition-colors ${
              card.isAutoPayEnabled ? 'bg-amber-500/20 border border-amber-500/40' : 'bg-emerald-500/20 border border-emerald-500/40'
            }`}>
              <Shield className={`w-3.5 h-3.5 transition-colors ${card.isAutoPayEnabled ? 'text-amber-400' : 'text-emerald-400'}`} />
            </div>
            <div className="flex-1 text-left min-w-0">
              <span className="text-xs font-medium text-white block leading-tight">Защита от автоподписок</span>
              <span className="text-[10px] text-slate-500 leading-tight block">Блокирует рекуррентные списания (Netflix, Figma и др.)</span>
            </div>
            <div className={`shrink-0 w-9 h-5 rounded-full transition-colors relative ${
              !card.isAutoPayEnabled ? 'bg-emerald-500' : 'bg-white/20'
            }`}>
              <div className={`absolute top-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform ${
                !card.isAutoPayEnabled ? 'translate-x-4' : 'translate-x-0.5'
              }`} />
            </div>
          </button>
        </div>
      )}
      </div>
    </>
  );
};

// Close Card Modal
const CloseCardModal = ({ card, onClose, onConfirm }: { card: PersonalCard; onClose: () => void; onConfirm: () => void }) => {
  const { t } = useTranslation();
  return (
  <ModalPortal>
  <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
    <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
    <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-[440px] animate-scale-in shadow-2xl shadow-black/60">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-xl bg-red-500/20 border border-red-500/30 flex items-center justify-center">
          <Trash2 className="w-6 h-6 text-red-400" />
        </div>
        <div>
          <h3 className="text-xl font-semibold text-white">{t('cards.closeModal.title')}</h3>
          <p className="text-sm text-slate-400">{card.name}</p>
        </div>
      </div>
      <p className="text-slate-300 mb-6">{t('cards.closeModal.warning')}</p>
      <div className="flex gap-3">
        <button onClick={onClose} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl">{t('cards.closeModal.cancel')}</button>
        <button onClick={onConfirm} className="flex-1 px-4 py-3 bg-red-500 hover:bg-red-600 text-white font-medium rounded-xl">{t('cards.closeModal.close')}</button>
      </div>
    </div>
  </div>
  </ModalPortal>
  );
};

// Payment Method Modal
const PaymentMethodModal = ({ type, onClose }: { type: 'apple' | 'google'; onClose: () => void }) => (
  <ModalPortal>
  <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
    <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
    <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-[440px] animate-scale-in shadow-2xl shadow-black/60">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-12 h-12 rounded-xl bg-white/10 flex items-center justify-center">
          {type === 'apple' ? <Apple className="w-6 h-6 text-white" /> : <Smartphone className="w-6 h-6 text-white" />}
        </div>
        <h3 className="text-xl font-semibold text-white">{type === 'apple' ? 'Apple Pay' : 'Google Pay'}</h3>
      </div>
      <ol className="space-y-2 text-slate-400 mb-6 text-sm">
        <li>1. Откройте {type === 'apple' ? 'Wallet' : 'Google Pay'}</li>
        <li>2. Нажмите "Добавить карту"</li>
        <li>3. Введите данные карты</li>
        <li>4. Подтвердите по SMS</li>
      </ol>
      <button onClick={onClose} className="w-full px-4 py-3 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-xl">Понятно</button>
    </div>
  </div>
  </ModalPortal>
);

// Bank Fees Tooltip
const BankFeesTooltip = ({ fees }: { fees: { name: string; value: string }[] }) => (
  <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-80 p-4 bg-[#1a1a24] border border-white/10 rounded-xl shadow-2xl z-50 animate-fade-in">
    <h4 className="text-amber-400 font-semibold mb-3">Дополнительные тарифы банка</h4>
    <div className="space-y-2">
      {fees.map((fee, i) => (
        <div key={i} className="flex items-center justify-between text-sm">
          <span className="text-slate-300">{fee.name}</span>
          <span className="text-white font-medium">{fee.value}</span>
        </div>
      ))}
    </div>
    <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 w-4 h-4 bg-[#1a1a24] border-r border-b border-white/10 transform rotate-45" />
  </div>
);

export const CardsPage = () => {
  const { rates } = useRates();
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'my-cards' | 'new-card'>('my-cards');
  const [closeCardModal, setCloseCardModal] = useState<PersonalCard | null>(null);
  const [topUpModal, setTopUpModal] = useState<PersonalCard | null>(null);
  const [paymentModal, setPaymentModal] = useState<{ type: 'apple' | 'google' } | null>(null);
  const [issueModal, setIssueModal] = useState<any>(null);
  const [walletBalance, setWalletBalance] = useState(0);
  const [personalCards, setPersonalCards] = useState<PersonalCard[]>([]);
  const [isLoadingCards, setIsLoadingCards] = useState(true);
  const [isIssuing, setIsIssuing] = useState(false);
  const [tierInfo, setTierInfo] = useState<TierInfo | null>(null);
  const [telegramLinked, setTelegramLinked] = useState<boolean | null>(null);
  const [showTgModal, setShowTgModal] = useState(false);
  const [pendingTopUpCard, setPendingTopUpCard] = useState<PersonalCard | null>(null);
  const [allSubscriptions, setAllSubscriptions] = useState<Array<CardSubscription & { card_id: number; card_last4: string }>>([]);
  const [isLoadingSubs, setIsLoadingSubs] = useState(false);

  // Map backend card type to PersonalCard color
  const typeColorMap: Record<string, 'blue' | 'purple' | 'gold'> = {
    subscriptions: 'blue', services: 'blue',
    travel: 'purple',
    premium: 'gold',
  };

  // Map backend Card → PersonalCard
  const mapBackendCard = (bc: BackendCard, details?: { full_number: string; cvv: string; expiry: string }): PersonalCard => {
    const slug = bc.service_slug || bc.category || 'subscriptions';
    const typeMap: Record<string, 'subscriptions' | 'travel' | 'premium'> = {
      subscriptions: 'subscriptions', services: 'subscriptions',
      travel: 'travel', premium: 'premium',
    };
    const cardType = typeMap[slug] || 'subscriptions';
    const isEur = slug === 'subscriptions' || slug === 'services';
    return {
      id: String(bc.id),
      type: cardType,
      name: bc.nickname || 'Карта',
      holderName: 'XPLR USER',
      number: details?.full_number
        ? details.full_number.replace(/(.{4})/g, '$1 ').trim()
        : `**** **** **** ${bc.last_4_digits}`,
      expiry: details?.expiry || '—',
      cvv: details?.cvv || '***',
      balance: parseFloat(bc.card_balance || '0'),
      currency: isEur ? '€' : '$',
      cardNetwork: (bc.card_type || '').toLowerCase().includes('visa') ? 'visa' : 'mastercard',
      color: typeColorMap[slug] || 'blue',
      price: '',
      bin: bc.bin || '',
      last4: bc.last_4_digits || '',
      isAutoPayEnabled: bc.is_auto_pay_enabled !== false,
    };
  };

  // Fetch cards from backend (only ACTIVE)
  const fetchCards = async () => {
    try {
      const backendCards = await getUserCards();
      const activeCards = (backendCards || []).filter(c => c.card_status === 'ACTIVE');

      // Fetch details for each card in parallel
      const mapped = await Promise.all(
        activeCards.map(async (bc) => {
          try {
            const details = await getCardDetails(bc.id);
            return mapBackendCard(bc, details);
          } catch {
            return mapBackendCard(bc);
          }
        })
      );
      // Sort: subscriptions first, travel second, premium last
      const typeOrder: Record<string, number> = { subscriptions: 0, travel: 1, premium: 2 };
      mapped.sort((a, b) => (typeOrder[a.type] ?? 9) - (typeOrder[b.type] ?? 9));
      setPersonalCards(mapped);
      // Fetch subscriptions after cards are loaded
      fetchAllSubscriptions(mapped);
    } catch (err) {
      console.error('Failed to fetch cards:', err);
    } finally {
      setIsLoadingCards(false);
    }
  };

  // Fetch all subscriptions across all cards
  const fetchAllSubscriptions = async (cards: PersonalCard[]) => {
    setIsLoadingSubs(true);
    try {
      const results = await Promise.all(
        cards
          .filter(c => c.type === 'subscriptions' || c.type === 'premium')
          .map(async (card) => {
            try {
              const data = await getCardSubscriptions(parseInt(card.id));
              return (data.subscriptions || []).map(sub => ({
                ...sub,
                card_id: parseInt(card.id),
                card_last4: card.last4,
              }));
            } catch {
              return [];
            }
          })
      );
      setAllSubscriptions(results.flat());
    } catch {
      setAllSubscriptions([]);
    } finally {
      setIsLoadingSubs(false);
    }
  };

  // Fetch wallet + cards + tier info on mount, auto-refresh wallet every 30s
  useEffect(() => {
    fetchCards();
    getTierInfo().then(setTierInfo).catch(() => {});
    getTelegramStatus().then(d => setTelegramLinked(d.telegram_linked)).catch(() => setTelegramLinked(false));
    const refreshWallet = () => {
      getWallet()
        .then((v) => setWalletBalance(Number(v.master_balance) || 0))
        .catch(() => {});
    };
    refreshWallet();
    const interval = setInterval(refreshWallet, 30000);
    return () => clearInterval(interval);
  }, []);

  // Handle card issuance via real API
  const handleIssueCard = async (cardType: 'subscriptions' | 'travel' | 'premium', priceUsd: number) => {
    // Check card limit before issuing
    if (tierInfo && !tierInfo.can_issue_more) {
      alert(`Достигнут лимит карт для вашего уровня (${tierInfo.current_cards}/${tierInfo.card_limit}). Улучшите уровень до GOLD для выпуска до 15 карт.`);
      return;
    }

    setIsIssuing(true);
    try {
      const result = await issuePersonalCard(cardType, priceUsd);
      if (result.successful_count > 0) {
        await fetchCards(); // Refresh card list from DB
        // Refresh tier info and wallet balance after card issue
        getTierInfo().then(setTierInfo).catch(() => {});
        getWallet().then((v) => setWalletBalance(Number(v.master_balance) || 0)).catch(() => {});
        setIssueModal(null);
        setActiveTab('my-cards');
      } else {
        const msg = result.results?.[0]?.message || 'Ошибка выпуска карты';
        alert(msg);
      }
    } catch (err: any) {
      const msg = err?.response?.data || err?.message || 'Ошибка выпуска карты';
      alert(typeof msg === 'string' ? msg : JSON.stringify(msg));
    } finally {
      setIsIssuing(false);
    }
  };

  // Handle card close via real API — refunds card_balance back to wallet
  const handleCloseCard = async (card: PersonalCard) => {
    try {
      await updateCardStatus(parseInt(card.id), 'CLOSED');
      setPersonalCards(prev => prev.filter(c => c.id !== card.id));
      setCloseCardModal(null);
      // Refresh wallet balance (card_balance was refunded to master_balance)
      getWallet().then(v => setWalletBalance(Number(v.master_balance) || 0)).catch(() => {});
    } catch (err) {
      console.error('Failed to close card:', err);
      setCloseCardModal(null);
    }
  };

  // Handle auto-pay toggle with optimistic UI
  const handleToggleAutoPay = async (cardId: string, enabled: boolean) => {
    // Optimistic: flip immediately
    setPersonalCards(prev =>
      prev.map(c => c.id === cardId ? { ...c, isAutoPayEnabled: enabled } : c)
    );
    try {
      await toggleAutoPay(parseInt(cardId), enabled);
    } catch (err) {
      // Revert on error
      setPersonalCards(prev =>
        prev.map(c => c.id === cardId ? { ...c, isAutoPayEnabled: !enabled } : c)
      );
      console.error('Failed to toggle auto-pay:', err);
    }
  };

  // Handle wallet-to-card transfer with optimistic UI
  const handleTransfer = async (card: PersonalCard, amountInCurrency: number) => {
    // walletBalance is in USD; convert transfer amount to USD equivalent for optimistic deduction
    const usdEquivalent = card.currency === '€'
      ? amountInCurrency * (rates.eur / rates.usd)
      : amountInCurrency; // already USD

    // Optimistic update — instant visual feedback
    setPersonalCards(prev =>
      prev.map(c => c.id === card.id ? { ...c, balance: c.balance + amountInCurrency } : c)
    );
    setWalletBalance(prev => prev - usdEquivalent);
    setTopUpModal(null);

    try {
      const updatedWallet = await transferWalletToCard(card.id, amountInCurrency, card.currency);
      setWalletBalance(Number(updatedWallet.master_balance) || 0);
    } catch (err) {
      // Revert on error
      setPersonalCards(prev =>
        prev.map(c => c.id === card.id ? { ...c, balance: c.balance - amountInCurrency } : c)
      );
      setWalletBalance(prev => prev + usdEquivalent);
      console.error('Transfer failed:', err);
    }
  };

  const cardTypesForIssue = [
    { 
      type: 'subscriptions' as const,
      name: 'Карта для подписок',
      description: 'Карта в евро, подходит для оплаты сервисов',
      currency: 'EUR',
      currencySymbol: '€',
      price: '$34',
      features: [
        { title: 'Онлайн сервисы', items: 'Netflix, Patreon, Apple Music, Disney+' },
        { title: 'Нейросети', items: 'ChatGPT, Grok, DeepL, Midjourney, Gemini, DeepSeek, Veo 3' },
        { title: 'Покупки на маркетплейсах', items: 'Amazon, Aliexpress, Ebay и др.' },
        { title: 'Для бизнеса и работы', items: 'Adobe Creative Cloud, Canva, Notion, Miro' },
      ],
      conditions: [
        { label: 'Выпуск карты', value: '$34' },
        { label: 'Первый год обслуживания', value: '$0' },
        { label: 'Комиссия за пополнение', value: '0%' },
        { label: 'Комиссия за транзакцию', value: '$0.25' },
        { label: 'Обслуживание после 1 года', value: '$34' },
      ],
      capabilities: [
        { label: '3D Secure', value: 'в приложении' },
        { label: 'Apple Pay', value: 'нет' },
        { label: 'Google Pay', value: 'да', link: true },
      ],
    },
    { 
      type: 'travel' as const,
      name: 'Карта для путешествий',
      description: 'С возможностью привязать Apple Pay и Google Pay',
      currency: 'USD',
      currencySymbol: '$',
      price: '$45',
      features: [
        { title: 'Бронирование и оплата отелей', items: 'Booking, AirBnb, Trip.com и другие' },
        { title: 'Покупка авиабилетов', items: 'Google Flights, Skyscanner, Kayak, Momondo' },
        { title: 'Оплата покупок через терминалы', items: 'Оплачивайте покупки в любых магазинах по всему миру, которые поддерживают Apple Pay и Google Pay' },
      ],
      conditions: [
        { label: 'Выпуск карты', value: '$45' },
        { label: 'Первый год обслуживания', value: '$0' },
        { label: 'Комиссия за пополнение', value: '0%' },
        { label: 'Комиссия за транзакцию', value: '$0.25' },
        { label: 'Обслуживание после 1 года', value: '$45' },
      ],
      capabilities: [
        { label: '3D Secure', value: 'в приложении' },
        { label: 'Apple Pay', value: 'да', link: true },
        { label: 'Google Pay', value: 'да', link: true },
      ],
    },
    { 
      type: 'premium' as const,
      name: 'Премиальная карта',
      description: 'Для тех, кто совершает много покупок за границей',
      currency: 'USD',
      currencySymbol: '$',
      price: '$168',
      features: [
        { title: 'Для покупок и подписок', items: 'Booking, Grab, Uber, Trip.com и любые другие сервисы' },
        { title: 'Покупка авиабилетов', items: 'Google Flights, Skyscanner, Kayak, Momondo' },
        { title: 'Самый выгодный курс валют', items: 'Возможность покупать больше' },
      ],
      conditions: [
        { label: 'Выпуск карты', value: '$168' },
        { label: 'Первый год обслуживания', value: '$0' },
        { label: 'Комиссия за пополнение', value: '0%' },
        { label: 'Комиссия за транзакцию', value: '$0.25' },
        { label: 'Обслуживание после 1 года', value: '$168' },
      ],
      capabilities: [
        { label: '3D Secure', value: 'в приложении' },
        { label: 'Apple Pay', value: 'да', link: true },
        { label: 'Google Pay', value: 'да', link: true },
      ],
    },
  ];

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />
        
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-white mb-2">{t('cards.title')}</h1>
            <p className="text-slate-400 text-sm">Управление картами и выпуск новых</p>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 p-1 bg-white/[0.04] rounded-xl mb-8 w-fit">
          <button
            onClick={() => setActiveTab('my-cards')}
            className={`flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium transition-all ${
              activeTab === 'my-cards'
                ? 'bg-white/10 text-white shadow-sm'
                : 'text-slate-400 hover:text-white hover:bg-white/[0.04]'
            }`}
          >
            <CardIcon className="w-4 h-4" />
            Мои карты
          </button>
          <button
            onClick={() => setActiveTab('new-card')}
            className={`flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium transition-all ${
              activeTab === 'new-card'
                ? 'bg-white/10 text-white shadow-sm'
                : 'text-slate-400 hover:text-white hover:bg-white/[0.04]'
            }`}
          >
            <Plus className="w-4 h-4" />
            Новая карта
          </button>
        </div>

        {/* Tab: Мои карты */}
        {activeTab === 'my-cards' && (
          <div>
            {isLoadingCards ? (
              <div className="flex items-center justify-center py-16">
                <div className="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin" />
              </div>
            ) : personalCards.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
                {personalCards.map(card => (
                  <RealisticCreditCard 
                    key={card.id} 
                    card={card}
                    onClose={() => setCloseCardModal(card)}
                    onTopUp={() => {
                      if (telegramLinked === false) {
                        setPendingTopUpCard(card);
                        setShowTgModal(true);
                      } else {
                        setTopUpModal(card);
                      }
                    }}
                    onApplePay={() => setPaymentModal({ type: 'apple' })}
                    onGooglePay={() => setPaymentModal({ type: 'google' })}
                    onToggleAutoPay={handleToggleAutoPay}
                  />
                ))}
              </div>
            ) : (
              <div className="glass-card p-12 text-center">
                <CardIcon className="w-12 h-12 text-slate-600 mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-white mb-2">Нет активных карт</h3>
                <p className="text-slate-400 text-sm mb-4">Выпустите первую карту, чтобы начать</p>
                <button onClick={() => setActiveTab('new-card')} className="px-6 py-2.5 bg-gradient-to-r from-blue-500 to-blue-600 text-white font-medium rounded-xl text-sm">
                  Выпустить карту
                </button>
              </div>
            )}

            {/* Subscription Hub — below card grid */}
            {personalCards.length > 0 && (
              <div className="mt-8 p-5 bg-white/[0.03] border border-white/[0.06] rounded-2xl">
                <div className="flex items-center gap-3 mb-3">
                  <div className="w-9 h-9 rounded-xl bg-amber-500/15 border border-amber-500/30 flex items-center justify-center">
                    <Shield className="w-4.5 h-4.5 text-amber-400" />
                  </div>
                  <div>
                    <h3 className="text-white font-semibold text-sm">Защита от автоподписок</h3>
                    <p className="text-[11px] text-slate-500 leading-snug">Управляйте списаниями каждого сервиса индивидуально. Отключайте ненужные подписки, не блокируя карту целиком.</p>
                  </div>
                </div>

                {isLoadingSubs ? (
                  <div className="flex items-center justify-center py-6">
                    <div className="w-5 h-5 border-2 border-amber-500/30 border-t-amber-500 rounded-full animate-spin" />
                  </div>
                ) : allSubscriptions.length > 0 ? (
                  <div className="space-y-1.5 mt-3">
                    {allSubscriptions.map(sub => (
                      <div key={`${sub.card_id}-${sub.id}`} className="flex items-center gap-3 px-3 py-2.5 bg-white/[0.03] rounded-xl hover:bg-white/[0.05] transition-colors">
                        <div className="w-7 h-7 rounded-lg bg-white/10 flex items-center justify-center shrink-0">
                          <Store className="w-3.5 h-3.5 text-slate-400" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <span className="text-xs font-medium text-white block truncate">{sub.merchant_name}</span>
                          <span className="text-[10px] text-slate-500">Карта **** {sub.card_last4} · {sub.charge_count} списан{sub.charge_count === 1 ? 'ие' : 'ий'}</span>
                        </div>
                        <button
                          onClick={async () => {
                            const newAllowed = !sub.is_allowed;
                            setAllSubscriptions(prev => prev.map(s =>
                              s.id === sub.id && s.card_id === sub.card_id ? { ...s, is_allowed: newAllowed } : s
                            ));
                            try {
                              await toggleSubscription(sub.card_id, sub.id, newAllowed);
                            } catch {
                              setAllSubscriptions(prev => prev.map(s =>
                                s.id === sub.id && s.card_id === sub.card_id ? { ...s, is_allowed: !newAllowed } : s
                              ));
                            }
                          }}
                          className="shrink-0"
                        >
                          <div className={`w-9 h-5 rounded-full transition-colors relative ${
                            sub.is_allowed ? 'bg-emerald-500' : 'bg-red-500/60'
                          }`}>
                            <div className={`absolute top-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform ${
                              sub.is_allowed ? 'translate-x-4' : 'translate-x-0.5'
                            }`} />
                          </div>
                        </button>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-xs text-slate-500 text-center py-4">Подписки пока не обнаружены. Они появятся автоматически после первого рекуррентного списания.</p>
                )}
              </div>
            )}
          </div>
        )}

        {/* Tab: Новая карта */}
        {activeTab === 'new-card' && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {cardTypesForIssue.map(ct => (
              <PersonalCardTypeCard
                key={ct.type}
                type={ct.type}
                name={ct.name}
                description={ct.description}
                currency={ct.currency}
                currencySymbol={ct.currencySymbol}
                price={ct.price}
                usdRate={rates.usd}
                eurRate={rates.eur}
                onSelect={() => setIssueModal(ct)}
              />
            ))}
          </div>
        )}

        {closeCardModal && <CloseCardModal card={closeCardModal} onClose={() => setCloseCardModal(null)} onConfirm={() => handleCloseCard(closeCardModal)} />}
        {topUpModal && <WalletTopUpModal card={topUpModal} walletBalance={walletBalance} onClose={() => setTopUpModal(null)} onTransfer={(amt) => handleTransfer(topUpModal, amt)} />}
        {paymentModal && <PaymentMethodModal type={paymentModal.type} onClose={() => setPaymentModal(null)} />}
        {issueModal && <CardIssueModal card={issueModal} onClose={() => setIssueModal(null)} onIssue={handleIssueCard} isIssuing={isIssuing} walletBalance={walletBalance} />}
        {showTgModal && (
          <TelegramRequiredModal
            onClose={() => { setShowTgModal(false); setPendingTopUpCard(null); }}
            onActivate={() => { setShowTgModal(false); window.location.href = '/settings'; }}
          />
        )}
      </div>

      <style>{`
        @keyframes scale-in { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
        .animate-scale-in { animation: scale-in 0.15s ease-out forwards; }
        @keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
        .animate-fade-in { animation: fade-in 0.15s ease-out forwards; }
      `}</style>
    </DashboardLayout>
  );
};
