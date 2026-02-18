import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react';

type Mode = 'PERSONAL' | 'ARBITRAGE';

const STORAGE_KEY = 'xplr_app_mode';

interface ModeContextType {
  mode: Mode;
  toggleMode: () => void;
  setMode: (mode: Mode) => void;
}

const ModeContext = createContext<ModeContextType | undefined>(undefined);

export const ModeProvider = ({ children }: { children: ReactNode }) => {
  const [mode, setModeState] = useState<Mode>(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    return (saved === 'PERSONAL' || saved === 'ARBITRAGE') ? saved : 'PERSONAL';
  });

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, mode);
  }, [mode]);

  const toggleMode = useCallback(() => {
    setModeState(prev => prev === 'PERSONAL' ? 'ARBITRAGE' : 'PERSONAL');
  }, []);

  const setMode = useCallback((newMode: Mode) => {
    setModeState(newMode);
  }, []);

  return (
    <ModeContext.Provider value={{ mode, toggleMode, setMode }}>
      {children}
    </ModeContext.Provider>
  );
};

export const useMode = () => {
  const context = useContext(ModeContext);
  if (context === undefined) {
    throw new Error('useMode must be used within a ModeProvider');
  }
  return context;
};
