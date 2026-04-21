import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  ArrowLeft,
  CreditCard,
  Check,
  AlertCircle,
  ChevronRight,
  Wallet,
  DollarSign,
  Info
} from 'lucide-react';

// Professional Visa Logo
const VisaLogo = () => (
  <svg viewBox="0 0 780 500" className="h-8 w-auto" preserveAspectRatio="xMidYMid meet">
    <path 
      d="M293.2 348.7l33.4-195.2h53.3l-33.4 195.2h-53.3zM534.4 157.6c-10.5-4.1-27.1-8.5-47.7-8.5-52.6 0-89.7 26.3-90 64-0.3 27.9 26.4 43.4 46.6 52.7 20.7 9.5 27.7 15.6 27.6 24.1-0.1 13-16.6 18.9-31.9 18.9-21.3 0-32.6-2.9-50.1-10.2l-6.9-3.1-7.5 43.6c12.4 5.4 35.4 10.1 59.3 10.4 55.9 0 92.2-26 92.7-66.2 0.3-22.1-14-38.9-44.6-52.8-18.6-9-30-15-29.9-24.1 0-8.1 9.7-16.7 30.5-16.7 17.4-0.3 30.1 3.5 39.9 7.4l4.8 2.2 7.2-41.7zM651.4 153.5h-41.2c-12.8 0-22.3 3.5-27.9 16.2l-79.2 178h56l11.2-29.2h68.4c1.6 6.8 6.5 29.2 6.5 29.2h49.5l-43.3-194.2zm-65.8 125.3c4.4-11.2 21.4-54.5 21.4-54.5-0.3 0.5 4.4-11.3 7.1-18.7l3.6 16.9s10.3 46.9 12.4 56.3h-44.5zM231.4 153.5l-52.2 133.2-5.5-27c-9.7-30.9-39.8-64.4-73.5-81.2l47.6 169h56.5l84.1-194h-57z" 
      fill="#1A1F71"
    />
    <path 
      d="M131.9 153.5H46.6l-0.7 4.1c67 16.1 111.4 55 129.7 101.7l-18.7-89.6c-3.2-12.2-12.6-15.8-25-16.2z" 
      fill="#F7B600"
    />
  </svg>
);

// Professional Mastercard Logo
const MastercardLogo = () => (
  <svg viewBox="0 0 780 500" className="h-8 w-auto" preserveAspectRatio="xMidYMid meet">
    <circle cx="250" cy="250" r="150" fill="#EB001B"/>
    <circle cx="530" cy="250" r="150" fill="#F79E1B"/>
    <path 
      d="M390 120.8C351.5 152.6 326 200.5 326 250c0 49.5 25.5 97.4 64 129.2 38.5-31.8 64-79.7 64-129.2 0-49.5-25.5-97.4-64-129.2z" 
      fill="#FF5F00"
    />
  </svg>
);

interface CardIssuePageProps {
  cardType: 'visa' | 'mastercard';
  bin: string;
}

export const CardIssuePage = () => {
  const navigate = useNavigate();
  const [quantity, setQuantity] = useState(1);
  const [budget, setBudget] = useState(1000);
  const [cardName, setCardName] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  
  // Get card type from URL params (simplified)
  const urlParams = new URLSearchParams(window.location.search);
  const cardType = (urlParams.get('type') || 'visa') as 'visa' | 'mastercard';
  const bin = urlParams.get('bin') || (cardType === 'visa' ? '4865 55**' : '5360 25**');
  
  const cardPrice = 4; // $4 per card
  const topUpFee = 6.7; // 6.7%
  const totalCardCost = quantity * cardPrice;
  const totalBudget = quantity * budget;
  const topUpAmount = totalBudget * (topUpFee / 100);
  const grandTotal = totalCardCost + totalBudget + topUpAmount;

  const handleIssue = async () => {
    setIsProcessing(true);
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 2000));
    setIsProcessing(false);
    navigate('/cards');
  };

  return (
    <DashboardLayout>
      <div className="stagger-fade-in max-w-3xl mx-auto">
        <BackButton />
        
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl md:text-3xl font-bold text-white mb-2">Выпуск карты</h1>
          <p className="text-slate-400">Настройте параметры новой виртуальной карты</p>
        </div>

        {/* Card Preview */}
        <div className="glass-card p-6 mb-6">
          <div className="flex items-center gap-4 mb-6">
            <div className={`w-16 h-10 rounded-lg flex items-center justify-center ${
              cardType === 'visa' ? 'bg-white' : 'bg-gradient-to-r from-red-500 to-yellow-500'
            }`}>
              {cardType === 'visa' ? <VisaLogo /> : <MastercardLogo />}
            </div>
            <div>
              <h3 className="text-white font-semibold text-lg">
                {cardType === 'visa' ? 'Visa' : 'MasterCard'}
              </h3>
              <p className="text-slate-400 text-sm font-mono">{bin} *</p>
            </div>
          </div>

          {/* Card Info */}
          <div className="grid grid-cols-2 gap-4 p-4 bg-white/5 rounded-xl">
            <div>
              <span className="text-slate-500 text-xs">Стоимость карты</span>
              <p className="text-white font-semibold">${cardPrice}</p>
            </div>
            <div>
              <span className="text-slate-500 text-xs">Комиссия пополнения</span>
              <p className="text-white font-semibold">{topUpFee}%</p>
            </div>
          </div>
        </div>

        {/* Configuration Form */}
        <div className="glass-card p-6 mb-6">
          <h3 className="text-white font-semibold mb-4 flex items-center gap-2">
            <CreditCard className="w-5 h-5 text-blue-400" />
            Параметры выпуска
          </h3>

          <div className="space-y-5">
            {/* Card Name */}
            <div>
              <label className="block text-sm text-slate-400 mb-2">
                Название карты (опционально)
              </label>
              <input
                type="text"
                placeholder="Например: Google Ads #1"
                value={cardName}
                onChange={(e) => setCardName(e.target.value)}
                className="w-full h-12 px-4 bg-white/5 border border-white/10 rounded-xl text-white focus:outline-none focus:border-blue-500/50 transition-all placeholder:text-slate-600"
              />
            </div>

            {/* Quantity */}
            <div>
              <label className="block text-sm text-slate-400 mb-2">
                Количество карт: <span className="text-white font-semibold">{quantity}</span>
              </label>
              <input
                type="range"
                min="1"
                max="100"
                value={quantity}
                onChange={(e) => setQuantity(parseInt(e.target.value))}
                className="w-full h-2 bg-white/10 rounded-lg appearance-none cursor-pointer accent-blue-500"
              />
              <div className="flex justify-between text-xs text-slate-500 mt-1">
                <span>1</span>
                <span>25</span>
                <span>50</span>
                <span>75</span>
                <span>100</span>
              </div>
            </div>

            {/* Budget per Card */}
            <div>
              <label className="block text-sm text-slate-400 mb-2">
                Начальный бюджет на карту ($)
              </label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">$</span>
                <input
                  type="number"
                  min="100"
                  max="50000"
                  value={budget}
                  onChange={(e) => setBudget(parseInt(e.target.value) || 0)}
                  className="w-full h-12 pl-10 pr-4 bg-white/5 border border-white/10 rounded-xl text-white focus:outline-none focus:border-blue-500/50 transition-all"
                />
              </div>
              <div className="flex gap-2 mt-2">
                {[500, 1000, 2500, 5000, 10000].map(val => (
                  <button
                    key={val}
                    onClick={() => setBudget(val)}
                    className={`flex-1 py-2 text-xs rounded-lg transition-colors ${
                      budget === val 
                        ? 'bg-blue-500/20 border border-blue-500/50 text-blue-400' 
                        : 'bg-white/5 border border-white/10 text-slate-400 hover:bg-white/10'
                    }`}
                  >
                    ${val.toLocaleString()}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Summary */}
        <div className="glass-card p-6 mb-6 bg-gradient-to-br from-blue-500/10 to-purple-500/10 border-blue-500/30">
          <h3 className="text-white font-semibold mb-4 flex items-center gap-2">
            <Wallet className="w-5 h-5 text-blue-400" />
            Итого к оплате
          </h3>

          <div className="space-y-3">
            <div className="flex justify-between items-center text-sm">
              <span className="text-slate-400">Карт к выпуску</span>
              <span className="text-white font-medium">{quantity}</span>
            </div>
            <div className="flex justify-between items-center text-sm">
              <span className="text-slate-400">Стоимость карт ({quantity} × ${cardPrice})</span>
              <span className="text-white font-medium">${totalCardCost.toFixed(2)}</span>
            </div>
            <div className="flex justify-between items-center text-sm">
              <span className="text-slate-400">Начальный бюджет ({quantity} × ${budget.toLocaleString()})</span>
              <span className="text-white font-medium">${totalBudget.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center text-sm">
              <span className="text-slate-400">Комиссия пополнения ({topUpFee}%)</span>
              <span className="text-white font-medium">${topUpAmount.toFixed(2)}</span>
            </div>
            
            <div className="border-t border-white/10 pt-3 mt-3">
              <div className="flex justify-between items-center">
                <span className="text-slate-300 font-medium">Итого</span>
                <span className="text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
                  ${grandTotal.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Info Notice */}
        <div className="flex items-start gap-3 p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl mb-6">
          <AlertCircle className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
          <div className="text-sm">
            <p className="text-amber-400 font-medium mb-1">Важная информация</p>
            <p className="text-slate-400">
              После выпуска карты средства будут списаны с вашего баланса. 
              Карта будет готова к использованию в течение 1-2 минут.
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-4">
          <button 
            onClick={() => navigate('/cards')}
            className="flex-1 px-6 py-4 bg-white/5 hover:bg-white/10 border border-white/10 text-slate-300 font-medium rounded-xl transition-colors"
          >
            Отмена
          </button>
          <button 
            onClick={handleIssue}
            disabled={isProcessing}
            className="flex-1 px-6 py-4 bg-gradient-to-r from-blue-500 to-purple-500 hover:from-blue-600 hover:to-purple-600 text-white font-medium rounded-xl transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isProcessing ? (
              <>
                <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Выпускаем...
              </>
            ) : (
              <>
                <CreditCard className="w-5 h-5" />
                Выпустить {quantity > 1 ? `${quantity} карт` : 'карту'}
              </>
            )}
          </button>
        </div>
      </div>
    </DashboardLayout>
  );
};
