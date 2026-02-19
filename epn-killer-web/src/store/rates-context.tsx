import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react';
import apiClient, { API_BASE_URL } from '../api/axios';

interface Rates {
  usd: number;
  eur: number;
}

interface RatesContextType {
  rates: Rates;
  setRates: (rates: Rates) => void;
  refreshRates: () => Promise<void>;
  getRate: (currency: 'USD' | 'EUR') => number;
}

const DEFAULT_RATES: Rates = { usd: 89.45, eur: 97.82 };

const RatesContext = createContext<RatesContextType | undefined>(undefined);

export const RatesProvider = ({ children }: { children: ReactNode }) => {
  const [rates, setRatesState] = useState<Rates>(() => {
    try {
      const saved = localStorage.getItem('xplr_rates');
      if (saved) return JSON.parse(saved);
    } catch {}
    return DEFAULT_RATES;
  });

  useEffect(() => {
    localStorage.setItem('xplr_rates', JSON.stringify(rates));
  }, [rates]);

  const refreshRates = useCallback(async () => {
    try {
      const res = await apiClient.get(`${API_BASE_URL}/rates`);
      if (res.data?.usd && res.data?.eur) {
        setRatesState({ usd: Number(res.data.usd), eur: Number(res.data.eur) });
      }
    } catch {
      // Keep current rates on failure
    }
  }, []);

  // Try to fetch rates on mount
  useEffect(() => {
    refreshRates();
  }, [refreshRates]);

  const setRates = useCallback((newRates: Rates) => {
    setRatesState(newRates);
  }, []);

  const getRate = useCallback((currency: 'USD' | 'EUR') => {
    return currency === 'USD' ? rates.usd : rates.eur;
  }, [rates]);

  return (
    <RatesContext.Provider value={{ rates, setRates, refreshRates, getRate }}>
      {children}
    </RatesContext.Provider>
  );
};

export const useRates = () => {
  const context = useContext(RatesContext);
  if (context === undefined) {
    throw new Error('useRates must be used within a RatesProvider');
  }
  return context;
};
