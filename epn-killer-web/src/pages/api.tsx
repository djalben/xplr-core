import { useState } from 'react';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  Key,
  RefreshCw,
  Copy,
  Check,
  Eye,
  EyeOff,
  Code,
  Terminal,
  Zap,
  Shield,
  Clock
} from 'lucide-react';

const CodeBlock = ({ code, language }: { code: string; language: string }) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative rounded-xl overflow-hidden bg-[#0a0a0c] border border-white/10">
      <div className="flex items-center justify-between px-4 py-2 bg-white/5 border-b border-white/10">
        <span className="text-xs text-gray-500 font-mono">{language}</span>
        <button 
          onClick={handleCopy}
          className="flex items-center gap-1 text-xs text-gray-400 hover:text-white transition-colors"
        >
          {copied ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
          {copied ? 'Скопировано' : 'Копировать'}
        </button>
      </div>
      <pre className="p-4 overflow-x-auto text-sm">
        <code className="text-gray-300 font-mono">{code}</code>
      </pre>
    </div>
  );
};

const EndpointCard = ({ method, path, description }: { method: string; path: string; description: string }) => {
  const methodColors: Record<string, string> = {
    GET: 'bg-green-500/20 text-green-400',
    POST: 'bg-blue-500/20 text-blue-400',
    PUT: 'bg-yellow-500/20 text-yellow-400',
    DELETE: 'bg-red-500/20 text-red-400',
  };

  return (
    <div className="flex items-start gap-4 p-4 rounded-xl bg-white/[0.02] border border-white/5 hover:border-white/10 transition-colors">
      <span className={`px-2.5 py-1 rounded-md text-xs font-bold ${methodColors[method]}`}>
        {method}
      </span>
      <div className="flex-1">
        <p className="text-white font-mono text-sm mb-1">{path}</p>
        <p className="text-sm text-gray-500">{description}</p>
      </div>
    </div>
  );
};

export const ApiPage = () => {
  const [showKey, setShowKey] = useState(false);
  const [copied, setCopied] = useState(false);

  const apiKey = 'xplr_live_sk_7f8g9h0j1k2l3m4n5o6p7q8r9s0t';
  const maskedKey = 'xplr_live_sk_••••••••••••••••••••••••';

  const handleCopyKey = () => {
    navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const endpoints = [
    { method: 'POST', path: '/v1/cards/issue', description: 'Выпустить новую виртуальную карту' },
    { method: 'GET', path: '/v1/cards/{id}', description: 'Получить данные карты' },
    { method: 'GET', path: '/v1/cards', description: 'Список всех карт с пагинацией' },
    { method: 'PUT', path: '/v1/cards/{id}/pause', description: 'Приостановить карту' },
    { method: 'PUT', path: '/v1/cards/{id}/resume', description: 'Возобновить карту' },
    { method: 'DELETE', path: '/v1/cards/{id}', description: 'Удалить/закрыть карту' },
    { method: 'GET', path: '/v1/transactions', description: 'Список транзакций' },
    { method: 'GET', path: '/v1/balance', description: 'Получить баланс аккаунта' },
    { method: 'POST', path: '/v1/webhooks', description: 'Настроить вебхуки' },
  ];

  const curlExample = `curl -X POST https://api.xplr.io/v1/cards/issue \\
  -H "Authorization: Bearer ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "bin": "459312",
    "budget": 5000,
    "currency": "USD",
    "metadata": {
      "purpose": "advertising"
    }
  }'`;

  const pythonExample = `import requests

headers = {
    "Authorization": "Bearer ${apiKey}",
    "Content-Type": "application/json"
}

response = requests.post(
    "https://api.xplr.io/v1/cards/issue",
    headers=headers,
    json={
        "bin": "459312",
        "budget": 5000,
        "currency": "USD"
    }
)

card = response.json()
print(f"Карта создана: {card['id']}")`;

  const nodeExample = `const response = await fetch('https://api.xplr.io/v1/cards/issue', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ${apiKey}',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    bin: '459312',
    budget: 5000,
    currency: 'USD'
  })
});

const card = await response.json();
console.log(\`Карта создана: \${card.id}\`);`;

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />
        
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-800 mb-2">API</h1>
          <p className="text-slate-500">Интегрируйте XPLR в ваши приложения</p>
        </div>

        {/* API Key Section */}
        <div className="glass-card p-6 mb-8">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center border border-blue-500/30">
              <Key className="w-6 h-6 text-blue-400" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-white">API ключ</h3>
              <p className="text-sm text-gray-400">Используйте этот ключ для аутентификации запросов</p>
            </div>
          </div>
          
          <div className="flex gap-3">
            <div className="flex-1 flex items-center gap-3 bg-[#0a0a0c] rounded-xl px-4 py-3 border border-white/10 font-mono text-sm">
              <span className="text-gray-300 truncate">{showKey ? apiKey : maskedKey}</span>
            </div>
            <button 
              onClick={() => setShowKey(!showKey)}
              className="px-4 py-3 rounded-xl bg-white/5 hover:bg-white/10 text-gray-400 transition-colors"
              title={showKey ? 'Скрыть' : 'Показать'}
            >
              {showKey ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
            </button>
            <button 
              onClick={handleCopyKey}
              className={`px-4 py-3 rounded-xl transition-all ${
                copied 
                  ? 'bg-green-500/20 text-green-400' 
                  : 'bg-white/5 hover:bg-white/10 text-gray-400'
              }`}
              title="Копировать"
            >
              {copied ? <Check className="w-5 h-5" /> : <Copy className="w-5 h-5" />}
            </button>
            <button className="px-4 py-3 rounded-xl bg-white/5 hover:bg-white/10 text-gray-400 transition-colors" title="Перегенерировать">
              <RefreshCw className="w-5 h-5" />
            </button>
          </div>

          <div className="flex items-center gap-6 mt-4 pt-4 border-t border-white/10">
            <div className="flex items-center gap-2 text-sm">
              <Shield className="w-4 h-4 text-green-400" />
              <span className="text-gray-400">Боевой режим</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Clock className="w-4 h-4 text-gray-500" />
              <span className="text-gray-400">Создан: 1 дек, 2024</span>
            </div>
          </div>
        </div>

        {/* Rate Limits */}
        <div className="glass-card p-6 mb-8">
          <h3 className="block-title flex items-center gap-2">
            <Zap className="w-5 h-5 text-yellow-400" />
            Лимиты запросов
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5">
              <p className="text-2xl font-bold text-white mb-1">1 000</p>
              <p className="text-sm text-gray-500">Запросов в минуту</p>
            </div>
            <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5">
              <p className="text-2xl font-bold text-white mb-1">50 000</p>
              <p className="text-sm text-gray-500">Запросов в день</p>
            </div>
            <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5">
              <p className="text-2xl font-bold text-white mb-1">100</p>
              <p className="text-sm text-gray-500">Одновременных соединений</p>
            </div>
          </div>
          <p className="text-sm text-gray-500 mt-4">
            Нужны выше лимиты? <a href="#" className="text-blue-400 hover:text-blue-300">Свяжитесь с нами</a> для обсуждения enterprise-плана.
          </p>
        </div>

        {/* Endpoints */}
        <div className="glass-card p-6 mb-8">
          <h3 className="block-title flex items-center gap-2">
            <Terminal className="w-5 h-5 text-blue-400" />
            Эндпоинты
          </h3>
          <div className="space-y-3">
            {endpoints.map(ep => (
              <EndpointCard key={`${ep.method}-${ep.path}`} {...ep} />
            ))}
          </div>
        </div>

        {/* Code Examples */}
        <div className="glass-card p-6">
          <h3 className="block-title flex items-center gap-2">
            <Code className="w-5 h-5 text-green-400" />
            Примеры кода
          </h3>
          
          <div className="space-y-6">
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-3">cURL</h4>
              <CodeBlock code={curlExample} language="bash" />
            </div>
            
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-3">Python</h4>
              <CodeBlock code={pythonExample} language="python" />
            </div>
            
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-3">Node.js / JavaScript</h4>
              <CodeBlock code={nodeExample} language="javascript" />
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
};
