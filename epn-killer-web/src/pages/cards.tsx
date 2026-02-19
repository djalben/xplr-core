import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMode } from '../store/mode-context';
import { useRates } from '../store/rates-context';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  Plus, 
  CreditCard as CardIcon,
  Wifi,
  Eye,
  EyeOff,
  Copy,
  Trash2,
  Pause,
  Check,
  X,
  Smartphone,
  Apple,
  LayoutGrid,
  List,
  ChevronRight,
  Banknote,
  ArrowUpDown,
  DollarSign,
  Plane,
  ShoppingBag,
  Monitor,
  Bot,
  Briefcase,
  CreditCard,
  ChevronDown
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
}

interface ArbitrageCard {
  id: string;
  bin: string;
  last4: string;
  fullNumber: string;
  expiry: string;
  cvv: string;
  budget: number;
  spent: number;
  status: 'active' | 'paused' | 'depleted';
  cardType: 'visa' | 'mastercard';
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
const MastercardLogo = ({ className = "h-8 w-auto" }: { className?: string }) => (
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
  onClose 
}: { 
  card: { type: string; name: string; price: string; currency: string; description: string; features: { title: string; items: string }[]; conditions: { label: string; value: string }[]; capabilities: { label: string; value: string; link?: boolean }[] };
  onClose: () => void;
}) => {
  const [showProhibited, setShowProhibited] = useState(false);
  const [selectedCurrency, setSelectedCurrency] = useState<'USD' | 'EUR'>(card.currency === 'EUR' ? 'EUR' : 'USD');
  const { rates } = useRates();

  const currentRate = selectedCurrency === 'USD' ? rates.usd : rates.eur;
  const currencySymbol = selectedCurrency === 'USD' ? '$' : '€';
  
  const prohibitedOperations = [
    'Финансовые инвестиции и управление активами',
    'Покупка виртуальных валют',
    'Азартные игры',
    'Транзакции, связанные с продуктами для взрослых',
    'Покупка подарочных карт',
    'Оплата российских сервисов и покупок на территории РФ'
  ];

  // Determine which payment methods are available
  const hasApplePay = card.type === 'travel' || card.type === 'premium';
  const hasGooglePay = true; // All cards support Google Pay

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
      <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 rounded-2xl w-full max-w-md max-h-[90dvh] overflow-y-auto animate-scale-in shadow-2xl shadow-black/60">
        {/* Card Preview */}
        <div className="p-6 pb-4">
          <div className="flex justify-center mb-4">
            {card.type === 'subscriptions' && <SubscriptionsCardVisual mini={false} />}
            {card.type === 'travel' && <TravelCardVisual mini={false} />}
            {card.type === 'premium' && <PremiumCardVisual mini={false} />}
          </div>
          
          {/* Payment method badges */}
          <div className="flex justify-center gap-3 mb-4">
            {hasApplePay && (
              <div className="px-3 py-1.5 bg-white/10 rounded-lg flex items-center gap-1.5 border border-white/5">
                <Apple className="w-4 h-4" />
                <span className="text-xs font-medium">Pay</span>
              </div>
            )}
            {hasGooglePay && (
              <div className="px-3 py-1.5 bg-white/10 rounded-lg flex items-center gap-1.5 border border-white/5">
                <span className="text-sm font-medium text-blue-400">G</span>
                <span className="text-xs font-medium">Pay</span>
              </div>
            )}
          </div>
          
          <h2 className="text-xl font-bold text-white text-center mb-1">{card.name}</h2>
          <p className="text-slate-400 text-sm text-center mb-4">{card.description}</p>
          <p className="text-2xl font-bold text-blue-400 text-center mb-2">{card.price}</p>
          
          {/* Currency selector */}
          <div className="flex items-center justify-center gap-2 mb-3">
            <button
              onClick={() => setSelectedCurrency('USD')}
              className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all ${
                selectedCurrency === 'USD'
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                  : 'bg-white/5 text-slate-400 border border-white/10 hover:bg-white/10'
              }`}
            >
              $ USD
            </button>
            <button
              onClick={() => setSelectedCurrency('EUR')}
              className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all ${
                selectedCurrency === 'EUR'
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                  : 'bg-white/5 text-slate-400 border border-white/10 hover:bg-white/10'
              }`}
            >
              € EUR
            </button>
          </div>

          {/* Exchange rate - dynamic */}
          <div className="text-center text-sm text-slate-400 mb-6">
            {card.type === 'premium' ? 'Лучший курс:' : 'Актуальный курс:'} <span className="text-blue-400">
              {currencySymbol}1 = {currentRate.toFixed(2)} ₽
            </span>
          </div>
        </div>
        
        {/* Features */}
        <div className="px-6 pb-4 space-y-3">
          {card.features.map((feature, i) => (
            <div key={i} className="flex items-start gap-3 p-3 bg-white/5 rounded-xl">
              <div className="w-6 h-6 rounded-full bg-blue-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                <Check className="w-4 h-4 text-blue-400" />
              </div>
              <div>
                <p className="text-white font-medium text-sm">{feature.title}</p>
                <p className="text-slate-400 text-xs">{feature.items}</p>
              </div>
            </div>
          ))}
        </div>
        
        {/* Conditions */}
        <div className="px-6 pb-4">
          <h3 className="text-white font-semibold mb-3">Условия выпуска:</h3>
          <ul className="space-y-2">
            {card.conditions.map((cond, i) => (
              <li key={i} className="flex items-center gap-2 text-sm">
                <span className="text-slate-400">•</span>
                <span className="text-slate-300">{cond.label} – <span className="text-white">{cond.value}</span></span>
              </li>
            ))}
          </ul>
        </div>
        
        {/* Capabilities */}
        <div className="px-6 pb-4">
          <h3 className="text-white font-semibold mb-3">Возможности:</h3>
          <ul className="space-y-2">
            {card.capabilities.map((cap, i) => (
              <li key={i} className="flex items-center gap-2 text-sm">
                <span className="text-slate-400">•</span>
                <span className="text-slate-300">{cap.label} – {cap.link ? (
                  <span className="text-blue-400 cursor-pointer hover:underline">{cap.value} <span className="text-slate-500 text-xs">(как привязать?)</span></span>
                ) : (
                  <span className="text-white">{cap.value}</span>
                )}</span>
              </li>
            ))}
          </ul>
        </div>
        
        {/* Prohibited Operations */}
        <div className="px-6 pb-4">
          <button 
            onClick={() => setShowProhibited(!showProhibited)}
            className="flex items-center justify-between w-full py-3 border-t border-white/10"
          >
            <span className="text-white font-semibold">Запрещенные операции</span>
            <ChevronDown className={`w-5 h-5 text-slate-400 transition-transform ${showProhibited ? 'rotate-180' : ''}`} />
          </button>
          {showProhibited && (
            <ul className="space-y-2 pb-4 animate-fade-in">
              {prohibitedOperations.map((op, i) => (
                <li key={i} className="flex items-center gap-2 text-sm">
                  <span className="text-slate-400">•</span>
                  <span className="text-slate-400">{op}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
        
        {/* Issue Button */}
        <div className="p-6 pt-2">
          <button 
            onClick={onClose}
            className="w-full py-4 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/20"
          >
            Выпустить карту {card.price}
          </button>
        </div>
      </div>
    </div>
  );
};

// Subscriptions Card Visual - colorful with service icons (realistic bank card style)
const SubscriptionsCardVisual = ({ mini = true }: { mini?: boolean }) => (
  <div className={`relative ${mini ? 'w-full aspect-[1.586/1]' : 'w-72 h-44'} rounded-2xl overflow-hidden shadow-2xl`}>
    {/* Realistic gradient background */}
    <div className="absolute inset-0">
      {/* Base gradient - pink/purple/blue like the screenshot */}
      <div className="absolute inset-0 bg-gradient-to-br from-pink-500 via-purple-500 to-blue-600" />
      
      {/* Colorful abstract shapes overlay */}
      <div className="absolute inset-0">
        <div className="absolute top-0 right-0 w-32 h-32 bg-blue-400/50 rounded-full blur-2xl translate-x-8 -translate-y-8" />
        <div className="absolute bottom-0 left-0 w-28 h-28 bg-pink-400/50 rounded-full blur-2xl -translate-x-8 translate-y-8" />
        <div className="absolute top-1/3 left-1/3 w-24 h-24 bg-purple-300/40 rounded-full blur-xl" />
        <div className="absolute bottom-1/4 right-1/4 w-20 h-20 bg-cyan-400/30 rounded-full blur-xl" />
      </div>
      
      {/* Floating service icon squares */}
      <div className="absolute top-8 left-6 w-8 h-8 bg-white/25 backdrop-blur-sm rounded-lg flex items-center justify-center shadow-lg">
        <Monitor className="w-4 h-4 text-white" />
      </div>
      <div className="absolute top-6 left-16 w-7 h-7 bg-white/20 backdrop-blur-sm rounded-lg flex items-center justify-center">
        <Bot className="w-3.5 h-3.5 text-white" />
      </div>
      <div className="absolute top-14 right-16 w-7 h-7 bg-white/20 backdrop-blur-sm rounded-lg flex items-center justify-center">
        <Briefcase className="w-3.5 h-3.5 text-white" />
      </div>
    </div>
    
    {/* Card content */}
    <div className="relative h-full p-4 flex flex-col justify-between">
      {/* Top row - branding and currency */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-full bg-white/25 backdrop-blur-sm flex items-center justify-center">
            <span className="text-[10px]">🌍</span>
          </div>
          <span className="text-white/90 text-xs font-medium tracking-wide">Без границ.</span>
        </div>
        <div className="w-8 h-8 rounded-full bg-white/25 backdrop-blur-sm flex items-center justify-center border border-white/20">
          <span className="text-white text-sm font-bold">€</span>
        </div>
      </div>
      
      {/* Card number at bottom */}
      <div className="mt-auto">
        <p className="text-white/50 text-[10px] mb-0.5">Номер карты</p>
        <p className="text-white font-mono text-sm tracking-widest">**** **** **** 1234</p>
      </div>
      
      {/* Mastercard logo at bottom right */}
      <div className="absolute bottom-4 right-4">
        <MastercardLogo className="h-7 w-auto" />
      </div>
    </div>
  </div>
);

// Travel Card Visual - blue gradient (realistic bank card style)
const TravelCardVisual = ({ mini = true }: { mini?: boolean }) => (
  <div className={`relative ${mini ? 'w-full aspect-[1.586/1]' : 'w-72 h-44'} rounded-2xl overflow-hidden shadow-2xl`}>
    {/* Blue gradient background */}
    <div className="absolute inset-0">
      <div className="absolute inset-0 bg-gradient-to-br from-blue-400 via-blue-500 to-blue-700" />
      
      {/* Soft light effects */}
      <div className="absolute top-0 right-0 w-40 h-40 bg-blue-300/40 rounded-full blur-3xl -translate-y-12 translate-x-12" />
      <div className="absolute bottom-0 left-0 w-32 h-32 bg-blue-300/30 rounded-full blur-2xl translate-y-8 -translate-x-8" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-48 h-48 bg-blue-200/20 rounded-full blur-3xl" />
      
      {/* Subtle globe pattern */}
      <div className="absolute inset-0 opacity-15">
        <svg viewBox="0 0 200 200" className="w-full h-full">
          <circle cx="100" cy="100" r="70" fill="none" stroke="white" strokeWidth="0.5" />
          <ellipse cx="100" cy="100" rx="70" ry="25" fill="none" stroke="white" strokeWidth="0.5" />
          <ellipse cx="100" cy="100" rx="25" ry="70" fill="none" stroke="white" strokeWidth="0.5" />
          <line x1="30" y1="100" x2="170" y2="100" stroke="white" strokeWidth="0.5" />
          <line x1="100" y1="30" x2="100" y2="170" stroke="white" strokeWidth="0.5" />
        </svg>
      </div>
    </div>
    
    {/* Card content */}
    <div className="relative h-full p-4 flex flex-col justify-between">
      {/* Top row - branding and currency */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-full bg-white/25 backdrop-blur-sm flex items-center justify-center">
            <Plane className="w-3 h-3 text-white" />
          </div>
          <span className="text-white/90 text-xs font-medium tracking-wide">Без границ.</span>
        </div>
        <div className="w-8 h-8 rounded-full bg-white/25 backdrop-blur-sm flex items-center justify-center border border-white/20">
          <span className="text-white text-sm font-bold">$</span>
        </div>
      </div>
      
      {/* Card number at bottom */}
      <div className="mt-auto">
        <p className="text-white/50 text-[10px] mb-0.5">Номер карты</p>
        <p className="text-white font-mono text-sm tracking-widest">**** **** **** 1234</p>
      </div>
      
      {/* Mastercard logo at bottom right */}
      <div className="absolute bottom-4 right-4">
        <MastercardLogo className="h-7 w-auto" />
      </div>
    </div>
  </div>
);

// Premium Card Visual - XPLR PRIME: deep black matte, neural texture, platinum chip, Power Beam
const PremiumCardVisual = ({ mini = true }: { mini?: boolean }) => (
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
          <p className="text-white font-bold text-lg tracking-[0.25em]">XPLR PRIME</p>
          <p className="text-white/25 text-[8px] tracking-[0.3em] uppercase">Platinum Edition</p>
        </div>
        <div className="w-8 h-8 rounded-full bg-white/[0.06] backdrop-blur-sm flex items-center justify-center border border-white/[0.08]">
          <span className="text-white text-sm font-bold">$</span>
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
        <p className="text-white/25 text-[10px] mb-0.5">Номер карты</p>
        <p className="text-white/90 font-mono text-sm tracking-widest">**** **** **** 1234</p>
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
  exchangeRate,
  onSelect 
}: { 
  type: 'subscriptions' | 'travel' | 'premium';
  name: string; 
  description: string; 
  currency: string;
  currencySymbol: string;
  price: string;
  exchangeRate: string;
  onSelect: () => void;
}) => {
  return (
    <div className="glass-card p-4 card-hover">
      {/* Card preview */}
      <div className="mb-4">
        {type === 'subscriptions' && <SubscriptionsCardVisual mini={true} />}
        {type === 'travel' && <TravelCardVisual mini={true} />}
        {type === 'premium' && <PremiumCardVisual mini={true} />}
      </div>
      
      <h4 className="text-white font-semibold text-center mb-1">{name}</h4>
      <p className="text-slate-400 text-xs text-center mb-3">{description}</p>
      
      {/* Price and rate */}
      <div className="grid grid-cols-2 gap-2 mb-4">
        <div className="p-2 bg-white/5 rounded-lg text-center">
          <p className="text-blue-400 font-bold text-lg">{price}</p>
          <p className="text-slate-500 text-[10px]">Стоимость</p>
        </div>
        <div className="p-2 bg-white/5 rounded-lg text-center">
          <p className="text-white font-medium text-sm">{exchangeRate}</p>
          <p className="text-slate-500 text-[10px]">Курс</p>
        </div>
      </div>
      
      <button 
        onClick={onSelect}
        className="w-full py-3 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white font-medium rounded-xl transition-all text-sm"
      >
        Выпустить карту {price}
      </button>
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
      flex flex-col items-center gap-1.5 p-2 rounded-xl transition-all duration-200 cursor-pointer
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

// Top-up Modal with bidirectional conversion
const TopUpModal = ({ card, onClose }: { card: PersonalCard; onClose: () => void }) => {
  const [rubAmount, setRubAmount] = useState('');
  const [foreignAmount, setForeignAmount] = useState('');
  const [selectedBank, setSelectedBank] = useState('sbp');
  const { rates } = useRates();
  
  const currencySymbol = card.currency === '€' ? '€' : '$';
  const currencyCode = card.currency === '€' ? 'EUR' : 'USD';
  const exchangeRate = card.currency === '€' ? rates.eur : rates.usd;

  const banks = [
    { id: 'sbp', name: 'СБП' },
    { id: 'sber', name: 'Сбер' },
    { id: 'tbank', name: 'Т-Банк' },
    { id: 'alfa', name: 'Альфа' },
    { id: 'vtb', name: 'ВТБ' },
  ];

  const handleRubChange = (value: string) => {
    setRubAmount(value);
    if (value && !isNaN(parseFloat(value))) {
      setForeignAmount((parseFloat(value) / exchangeRate).toFixed(2));
    } else {
      setForeignAmount('');
    }
  };

  const handleForeignChange = (value: string) => {
    setForeignAmount(value);
    if (value && !isNaN(parseFloat(value))) {
      setRubAmount((parseFloat(value) * exchangeRate).toFixed(0));
    } else {
      setRubAmount('');
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Dense dark overlay — 80% black + heavy blur */}
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />

      {/* Modal panel — deep opaque glass */}
      <div className="relative bg-[#0d0d0f]/95 backdrop-blur-3xl border border-white/10 p-5 rounded-2xl w-full max-w-md animate-scale-in max-h-[90dvh] overflow-y-auto shadow-2xl shadow-black/60">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-blue-500/20 border border-emerald-500/30 flex items-center justify-center">
              <Banknote className="w-5 h-5 text-emerald-400" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-white">Пополнить карту</h3>
              <p className="text-xs text-slate-400">{card.name}</p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-white/10 rounded-lg transition-colors">
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>
        
        <div className="mb-4">
          <label className="block text-sm text-slate-400 mb-2">Способ оплаты</label>
          <div className="grid grid-cols-5 gap-2">
            {banks.map((bank) => (
              <BankLogoButton key={bank.id} bank={bank} selected={selectedBank === bank.id} onClick={() => setSelectedBank(bank.id)} />
            ))}
          </div>
        </div>

        <div className="space-y-3 mb-4">
          <div>
            <label className="block text-sm text-slate-400 mb-1.5">Сумма в рублях</label>
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-300 text-lg font-bold">₽</span>
              <input
                type="number"
                placeholder="10 000"
                value={rubAmount}
                onChange={(e) => handleRubChange(e.target.value)}
                className="w-full h-12 pl-12 pr-4 bg-white/[0.04] border border-white/10 rounded-xl text-white text-lg font-semibold focus:outline-none focus:border-blue-500/50 transition-all placeholder:text-slate-600"
              />
            </div>
          </div>
          
          <div className="flex items-center justify-center py-1">
            <div className="flex items-center gap-3">
              <div className="h-px w-10 bg-white/10" />
              <div className="w-8 h-8 rounded-full bg-white/[0.04] border border-white/10 flex items-center justify-center">
                <span className="text-xs font-bold text-slate-300">₽→{currencySymbol}</span>
              </div>
              <div className="h-px w-10 bg-white/10" />
            </div>
          </div>

          <div>
            <label className="block text-sm text-slate-400 mb-1.5">Получите на карту</label>
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 text-emerald-400 text-lg font-bold">{currencySymbol}</span>
              <input
                type="number"
                placeholder="0.00"
                value={foreignAmount}
                onChange={(e) => handleForeignChange(e.target.value)}
                className="w-full h-12 pl-12 pr-4 bg-emerald-500/[0.04] border border-emerald-500/20 rounded-xl text-emerald-400 text-lg font-semibold focus:outline-none focus:border-emerald-500/50 transition-all placeholder:text-emerald-900"
              />
            </div>
          </div>
        </div>

        <div className="p-3 rounded-xl bg-white/[0.04] border border-white/10 mb-4">
          <div className="flex items-center justify-between text-sm">
            <span className="text-slate-400">Курс:</span>
            <span className="text-white font-bold">1 {currencyCode} = {exchangeRate.toFixed(2)} ₽</span>
          </div>
        </div>
        
        <div className="flex gap-3">
          <button onClick={onClose} className="flex-1 px-4 py-3 bg-white/[0.04] hover:bg-white/[0.08] border border-white/10 text-slate-300 font-medium rounded-xl transition-colors">
            Отмена
          </button>
          <button className="flex-1 px-4 py-3 bg-gradient-to-r from-emerald-500 to-blue-500 hover:from-emerald-400 hover:to-blue-400 text-white font-semibold rounded-xl transition-all shadow-lg shadow-emerald-500/20 flex items-center justify-center gap-2">
            <Banknote className="w-4 h-4" />
            Пополнить
          </button>
        </div>
      </div>
    </div>
  );
};

// Realistic Credit Card display for "Your Cards" section
const RealisticCreditCard = ({ 
  card, 
  onClose, 
  onTopUp,
  onApplePay,
  onGooglePay
}: { 
  card: PersonalCard; 
  onClose: () => void;
  onTopUp: () => void;
  onApplePay: () => void;
  onGooglePay: () => void;
}) => {
  const [showDetails, setShowDetails] = useState(false);
  const [copied, setCopied] = useState(false);
  
  const canAddToWallet = card.type === 'travel' || card.type === 'premium';

  const handleCopy = () => {
    if (showDetails) {
      navigator.clipboard.writeText(card.number.replace(/\s/g, ''));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const formatCardNumber = (num: string, show: boolean) => {
    if (show) return num;
    return `**** **** **** ${num.slice(-4)}`;
  };

  return (
    <div className="group">
      {/* Card visual based on type */}
      <div className="relative">
        {card.type === 'subscriptions' && <SubscriptionsCardVisual mini={true} />}
        {card.type === 'travel' && <TravelCardVisual mini={true} />}
        {card.type === 'premium' && <PremiumCardVisual mini={true} />}
        
        {/* Overlay for card details */}
        <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl flex items-center justify-center gap-2">
          <button 
            onClick={() => setShowDetails(!showDetails)}
            className="p-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors"
          >
            {showDetails ? <EyeOff className="w-5 h-5 text-white" /> : <Eye className="w-5 h-5 text-white" />}
          </button>
          {showDetails && (
            <button onClick={handleCopy} className="p-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors">
              {copied ? <Check className="w-5 h-5 text-emerald-400" /> : <Copy className="w-5 h-5 text-white" />}
            </button>
          )}
        </div>
      </div>
      
      {/* Card Info */}
      <div className="mt-3 p-3 bg-white/5 rounded-xl">
        <div className="flex items-center justify-between mb-2">
          <span className="text-white font-medium text-sm">{card.name}</span>
          <span className="text-emerald-400 font-bold">{card.currency}{card.balance.toLocaleString()}</span>
        </div>
        {showDetails && (
          <div className="space-y-1 text-xs text-slate-400 animate-fade-in">
            <p>Номер: <span className="text-white font-mono">{card.number}</span></p>
            <p>Срок: <span className="text-white">{card.expiry}</span> | CVV: <span className="text-white">{card.cvv}</span></p>
          </div>
        )}
      </div>

      {/* Action Buttons */}
      <div className="mt-3 flex gap-2">
        <button 
          onClick={onTopUp}
          className="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 rounded-xl transition-colors text-sm border border-emerald-500/30"
        >
          <Banknote className="w-4 h-4" />
          Пополнить
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
    </div>
  );
};

// Close Card Modal
const CloseCardModal = ({ card, onClose, onConfirm }: { card: PersonalCard; onClose: () => void; onConfirm: () => void }) => (
  <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
    <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-md animate-scale-in shadow-2xl shadow-black/60">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-12 h-12 rounded-xl bg-red-500/20 flex items-center justify-center">
          <Trash2 className="w-6 h-6 text-red-400" />
        </div>
        <div>
          <h3 className="text-xl font-semibold text-white">Закрыть карту?</h3>
          <p className="text-sm text-slate-400">{card.name}</p>
        </div>
      </div>
      <p className="text-slate-300 mb-6">Это действие нельзя отменить. Остаток средств будет возвращён.</p>
      <div className="flex gap-3">
        <button onClick={onClose} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl">Отмена</button>
        <button onClick={onConfirm} className="flex-1 px-4 py-3 bg-red-500 hover:bg-red-600 text-white font-medium rounded-xl">Закрыть</button>
      </div>
    </div>
  </div>
);

// Payment Method Modal
const PaymentMethodModal = ({ type, onClose }: { type: 'apple' | 'google'; onClose: () => void }) => (
  <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
    <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-md animate-scale-in shadow-2xl shadow-black/60">
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

// Arbitrage Card Row
const ArbitrageCardTypeRow = ({ cardType, bin, network, price, topUpFee, fees, onIssue }: { 
  cardType: string; bin: string; network: 'visa' | 'mastercard'; price: number; topUpFee: number; 
  fees: { name: string; value: string }[]; onIssue: () => void;
}) => {
  const [showFees, setShowFees] = useState(false);

  return (
    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 p-4 bg-white/[0.02] hover:bg-white/[0.04] border border-white/10 rounded-xl transition-all">
      <div className="flex items-center gap-4">
        <div className="px-3 py-1.5 bg-amber-500/10 border border-amber-500/30 rounded-lg">
          <span className="text-amber-400 font-mono text-sm">{bin}</span>
        </div>
        <div className="flex items-center gap-2">
          <div className={`w-14 h-9 rounded-md flex items-center justify-center ${network === 'visa' ? 'bg-white' : 'bg-gradient-to-r from-red-500 to-yellow-500'}`}>
            {network === 'visa' ? <VisaLogo className="h-5 w-auto" /> : <MastercardLogo className="h-6 w-auto" />}
          </div>
          <span className="text-white font-medium">{network === 'visa' ? 'Visa' : 'MasterCard'}</span>
        </div>
      </div>
      <div className="flex items-center gap-3 flex-wrap">
        <div className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg">
          <span className="text-slate-400 text-sm">Карта: </span><span className="text-white">${price}</span>
        </div>
        <div className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg">
          <span className="text-slate-400 text-sm">Пополнение: </span><span className="text-white">{topUpFee}%</span>
        </div>
        <div className="relative">
          <button
            onMouseEnter={() => setShowFees(true)}
            onMouseLeave={() => setShowFees(false)}
            className="px-3 py-1.5 bg-transparent border border-amber-500/50 text-amber-400 hover:bg-amber-500/10 rounded-lg text-sm"
          >
            Доп. тарифы
          </button>
          {showFees && <BankFeesTooltip fees={fees} />}
        </div>
        <button onClick={onIssue} className="px-6 py-2.5 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-lg">
          Выпустить
        </button>
      </div>
    </div>
  );
};

// Arbitrage Card display
const ArbitrageCardRow = ({ card }: { card: ArbitrageCard }) => {
  const [showDetails, setShowDetails] = useState(false);
  const [copied, setCopied] = useState(false);
  const usagePercent = (card.spent / card.budget) * 100;

  return (
    <tr className="hover:bg-white/[0.02]">
      <td className="py-4 px-4"><span className="font-mono text-white">{card.bin}</span></td>
      <td className="py-4 px-4">
        <div className="flex items-center gap-2">
          <div className={`w-10 h-6 rounded flex items-center justify-center ${card.cardType === 'visa' ? 'bg-white' : 'bg-gradient-to-r from-red-500 to-yellow-500'}`}>
            {card.cardType === 'visa' ? <VisaLogo className="h-3 w-auto" /> : <MastercardLogo className="h-4 w-auto" />}
          </div>
          <span className="font-mono text-slate-400">{showDetails ? card.fullNumber : `•••• ${card.last4}`}</span>
          <button onClick={() => setShowDetails(!showDetails)} className="p-1.5 hover:bg-white/10 rounded">
            {showDetails ? <EyeOff className="w-4 h-4 text-slate-400" /> : <Eye className="w-4 h-4 text-slate-400" />}
          </button>
        </div>
      </td>
      <td className="py-4 px-4">
        <div className="flex items-center gap-3">
          <div className="flex-1 max-w-[120px] h-1.5 bg-white/10 rounded-full overflow-hidden">
            <div className={`h-full rounded-full ${usagePercent > 90 ? 'bg-red-500' : usagePercent > 70 ? 'bg-yellow-500' : 'bg-blue-500'}`} style={{ width: `${usagePercent}%` }} />
          </div>
          <span className="text-sm text-slate-400">${card.spent.toLocaleString()} / ${card.budget.toLocaleString()}</span>
        </div>
      </td>
      <td className="py-4 px-4">
        <span className={`badge ${card.status === 'active' ? 'badge-success' : card.status === 'paused' ? 'badge-warning' : 'badge-error'}`}>
          {card.status === 'active' ? 'Активна' : card.status === 'paused' ? 'Пауза' : 'Исчерпана'}
        </span>
      </td>
      <td className="py-4 px-4">
        <div className="flex items-center gap-2">
          <button onClick={() => { navigator.clipboard.writeText(`${card.fullNumber} ${card.expiry} ${card.cvv}`); setCopied(true); setTimeout(() => setCopied(false), 2000); }} className="p-2 hover:bg-white/10 rounded-lg">
            {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4 text-slate-400" />}
          </button>
          <button className="p-2 hover:bg-white/10 rounded-lg"><Pause className="w-4 h-4 text-slate-400" /></button>
          <button className="p-2 hover:bg-red-500/20 rounded-lg"><Trash2 className="w-4 h-4 text-slate-400" /></button>
        </div>
      </td>
    </tr>
  );
};

// Grade Display
const GradeDisplay = () => {
  const grades = [
    { name: 'Стандарт', commission: '6.7%', color: 'bg-slate-500' },
    { name: 'Серебро', commission: '6.0%', color: 'bg-slate-400' },
    { name: 'Золото', commission: '5.0%', color: 'bg-amber-500' },
    { name: 'Платина', commission: '4.0%', color: 'bg-purple-400' },
    { name: 'Блэк', commission: '3.0%', color: 'bg-slate-900' },
  ];

  return (
    <div className="glass-card p-6 mb-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="block-title mb-0">Грейд и комиссии</h3>
        <span className="px-3 py-1 bg-slate-500/20 border border-slate-500/30 text-slate-400 rounded-full text-sm">Стандарт</span>
      </div>
      <div className="h-2 bg-white/10 rounded-full overflow-hidden mb-4">
        <div className="h-full w-1/4 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full" />
      </div>
      <div className="flex gap-1">
        {grades.map((g, i) => (
          <div key={g.name} className="flex-1 text-center">
            <div className={`h-2 ${g.color} ${i === 0 ? 'rounded-l-full' : ''} ${i === 4 ? 'rounded-r-full' : ''} ${i === 0 ? 'opacity-100' : 'opacity-30'}`} />
            <p className="text-xs text-slate-500 mt-2">{g.name}</p>
            <p className="text-[10px] text-slate-600">{g.commission}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export const CardsPage = () => {
  const { mode } = useMode();
  const { rates } = useRates();
  const navigate = useNavigate();
  const [closeCardModal, setCloseCardModal] = useState<PersonalCard | null>(null);
  const [topUpModal, setTopUpModal] = useState<PersonalCard | null>(null);
  const [paymentModal, setPaymentModal] = useState<{ type: 'apple' | 'google' } | null>(null);
  const [issueModal, setIssueModal] = useState<any>(null);
  const [viewMode, setViewMode] = useState<'table' | 'grid'>('table');

  const personalCards: PersonalCard[] = [
    { id: '1', type: 'subscriptions', name: 'Карта для подписок', holderName: 'IVAN PETROV', number: '4521 8834 2291 7432', expiry: '12/26', cvv: '847', balance: 1250.50, currency: '€', cardNetwork: 'mastercard', color: 'blue', price: '2 990₽' },
    { id: '2', type: 'travel', name: 'Карта для путешествий', holderName: 'IVAN PETROV', number: '5234 1192 8847 0923', expiry: '08/27', cvv: '312', balance: 3840.00, currency: '$', cardNetwork: 'mastercard', color: 'purple', price: '3 990₽' },
    { id: '3', type: 'premium', name: 'Премиальная карта', holderName: 'IVAN PETROV', number: '3782 8224 6310 0052', expiry: '03/28', cvv: '921', balance: 12580.00, currency: '$', cardNetwork: 'mastercard', color: 'gold', price: '14 990₽' },
  ];

  const arbitrageCards: ArbitrageCard[] = [
    { id: '1', bin: '486555', last4: '4521', fullNumber: '4865 5512 3456 4521', expiry: '12/26', cvv: '847', budget: 5000, spent: 3240, status: 'active', cardType: 'visa' },
    { id: '2', bin: '536025', last4: '7832', fullNumber: '5360 2534 5678 7832', expiry: '11/26', cvv: '293', budget: 5000, spent: 4890, status: 'active', cardType: 'mastercard' },
  ];

  const cardTypesForIssue = [
    { 
      type: 'subscriptions' as const,
      name: 'Карта для подписок',
      description: 'Карта в евро, подходит для оплаты сервисов',
      currency: 'EUR',
      currencySymbol: '€',
      price: '2 990 ₽',
      exchangeRate: `USD: ${rates.usd.toFixed(2)}₽ | EUR: ${rates.eur.toFixed(2)}₽`,
      features: [
        { title: 'Онлайн сервисы', items: 'Netflix, Patreon, Apple Music, Disney+' },
        { title: 'Нейросети', items: 'ChatGPT, Grok, DeepL, Midjourney, Gemini, DeepSeek, Veo 3' },
        { title: 'Покупки на маркетплейсах', items: 'Amazon, Aliexpress, Ebay и др.' },
        { title: 'Для бизнеса и работы', items: 'Adobe Creative Cloud, Canva, Notion, Miro' },
      ],
      conditions: [
        { label: 'Выпуск карты', value: '2 990 ₽' },
        { label: 'Первый год обслуживания', value: '0 ₽' },
        { label: 'Комиссия за пополнение', value: '0%' },
        { label: 'Комиссия за транзакцию', value: '€0.25' },
        { label: 'Обслуживание после 1 года', value: '2 990 ₽' },
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
      price: '3 990 ₽',
      exchangeRate: `USD: ${rates.usd.toFixed(2)}₽ | EUR: ${rates.eur.toFixed(2)}₽`,
      features: [
        { title: 'Бронирование и оплата отелей', items: 'Booking, AirBnb, Trip.com и другие' },
        { title: 'Покупка авиабилетов', items: 'Google Flights, Skyscanner, Kayak, Momondo' },
        { title: 'Оплата покупок через терминалы', items: 'Оплачивайте покупки в любых магазинах по всему миру, которые поддерживают Apple Pay и Google Pay' },
      ],
      conditions: [
        { label: 'Выпуск карты', value: '3 990 ₽' },
        { label: 'Первый год обслуживания', value: '0 ₽' },
        { label: 'Комиссия за пополнение', value: '0%' },
        { label: 'Комиссия за транзакцию', value: '$0.25' },
        { label: 'Обслуживание после 1 года', value: '3 990 ₽' },
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
      price: '14 990 ₽',
      exchangeRate: `USD: ${rates.usd.toFixed(2)}₽ | EUR: ${rates.eur.toFixed(2)}₽`,
      features: [
        { title: 'Для покупок и подписок', items: 'Booking, Grab, Uber, Trip.com и любые другие сервисы' },
        { title: 'Покупка авиабилетов', items: 'Google Flights, Skyscanner, Kayak, Momondo' },
        { title: 'Самый выгодный курс валют', items: 'Возможность покупать больше' },
      ],
      conditions: [
        { label: 'Выпуск карты', value: '14 990 ₽' },
        { label: 'Первый год обслуживания', value: '0 ₽' },
        { label: 'Комиссия за пополнение', value: '0%' },
        { label: 'Комиссия за транзакцию', value: '$0.25' },
        { label: 'Обслуживание после 1 года', value: '14 990 ₽' },
      ],
      capabilities: [
        { label: '3D Secure', value: 'в приложении' },
        { label: 'Apple Pay', value: 'да', link: true },
        { label: 'Google Pay', value: 'да', link: true },
      ],
    },
  ];

  const bankFees = [
    { name: 'Комиссия за транзакцию', value: '$0.2' },
    { name: 'Международная транзакция', value: '$0.5 + 1.2%' },
    { name: 'Отмена транзакции', value: '$0.2' },
    { name: 'Отмена международной транзакции', value: '$0.8' },
    { name: 'Возврат средств', value: '$0.2 + 1%' },
    { name: 'Возврат средств (международный)', value: '$0.5 + 1%' },
    { name: 'Отклонение платежа', value: '$0.1' },
    { name: 'Отклонение международного платежа', value: '$0.4' },
  ];

  const availableCardTypes = [
    { id: 'visa-1', cardType: 'Visa', bin: '4865 55** *', network: 'visa' as const, price: 4, topUpFee: 6.7 },
    { id: 'mastercard-1', cardType: 'MasterCard', bin: '5360 25** *', network: 'mastercard' as const, price: 4, topUpFee: 6.7 },
  ];

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />
        
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-white mb-2">Карты</h1>
            <p className="text-slate-400 text-sm">
              {mode === 'PERSONAL' ? 'Виртуальные карты для личных нужд' : 'Карты для арбитража'}
            </p>
          </div>
          {mode === 'ARBITRAGE' && (
            <div className="flex items-center gap-3">
              <div className="flex items-center glass-card p-1">
                <button onClick={() => setViewMode('table')} className={`p-2.5 rounded-lg ${viewMode === 'table' ? 'bg-blue-500/20 text-blue-400' : 'text-slate-400'}`}>
                  <List className="w-5 h-5" />
                </button>
                <button onClick={() => setViewMode('grid')} className={`p-2.5 rounded-lg ${viewMode === 'grid' ? 'bg-blue-500/20 text-blue-400' : 'text-slate-400'}`}>
                  <LayoutGrid className="w-5 h-5" />
                </button>
              </div>
            </div>
          )}
        </div>

        {mode === 'PERSONAL' && (
          <>
            <div className="mb-8">
              <h3 className="section-header flex items-center gap-2"><Plus className="w-4 h-4" />Выпустить новую карту</h3>
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
                    exchangeRate={ct.exchangeRate}
                    onSelect={() => setIssueModal(ct)}
                  />
                ))}
              </div>
            </div>

            <div>
              <h3 className="section-header flex items-center gap-2"><CardIcon className="w-4 h-4" />Ваши карты</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
                {personalCards.map(card => (
                  <RealisticCreditCard 
                    key={card.id} 
                    card={card}
                    onClose={() => setCloseCardModal(card)}
                    onTopUp={() => setTopUpModal(card)}
                    onApplePay={() => setPaymentModal({ type: 'apple' })}
                    onGooglePay={() => setPaymentModal({ type: 'google' })}
                  />
                ))}
              </div>
            </div>
          </>
        )}

        {mode === 'ARBITRAGE' && (
          <>
            <GradeDisplay />
            <div className="mb-8">
              <h3 className="section-header flex items-center gap-2"><Plus className="w-4 h-4" />Выпустить карту</h3>
              <div className="space-y-3">
                {availableCardTypes.map(ct => (
                  <ArbitrageCardTypeRow
                    key={ct.id}
                    cardType={ct.cardType}
                    bin={ct.bin}
                    network={ct.network}
                    price={ct.price}
                    topUpFee={ct.topUpFee}
                    fees={bankFees}
                    onIssue={() => navigate(`/card-issue?type=${ct.network}&bin=${encodeURIComponent(ct.bin)}`)}
                  />
                ))}
              </div>
            </div>

            <h3 className="section-header flex items-center gap-2"><CardIcon className="w-4 h-4" />Активные карты</h3>
            <div className="glass-card overflow-x-auto">
              <table className="xplr-table min-w-[800px]">
                <thead><tr><th>BIN</th><th>Карта</th><th>Бюджет</th><th>Статус</th><th>Действия</th></tr></thead>
                <tbody>{arbitrageCards.map(card => <ArbitrageCardRow key={card.id} card={card} />)}</tbody>
              </table>
            </div>
          </>
        )}

        {closeCardModal && <CloseCardModal card={closeCardModal} onClose={() => setCloseCardModal(null)} onConfirm={() => setCloseCardModal(null)} />}
        {topUpModal && <TopUpModal card={topUpModal} onClose={() => setTopUpModal(null)} />}
        {paymentModal && <PaymentMethodModal type={paymentModal.type} onClose={() => setPaymentModal(null)} />}
        {issueModal && <CardIssueModal card={issueModal} onClose={() => setIssueModal(null)} />}
      </div>

      <style>{`
        @keyframes scale-in { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
        .animate-scale-in { animation: scale-in 0.2s ease-out forwards; }
        @keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
        .animate-fade-in { animation: fade-in 0.3s ease-out forwards; }
      `}</style>
    </DashboardLayout>
  );
};
