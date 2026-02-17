// XPLR Premium — тёмная премиальная схема с акцентами в стиле "X"
export const theme = {
  colors: {
    // Фоны (глубокие тёмные тона)
    background: '#0a0a0b',
    backgroundSecondary: '#111113',
    backgroundCard: '#161619',
    backgroundElevated: '#1c1c21',

    // Акценты X (острый контраст: бирюза/циан + мягкий золотой)
    accent: '#06b6d4',
    accentHover: '#22d3ee',
    accentMuted: 'rgba(6, 182, 212, 0.15)',
    accentBorder: 'rgba(6, 182, 212, 0.35)',
    gold: '#c9a227',
    goldMuted: 'rgba(201, 162, 39, 0.12)',

    // Семантика
    success: '#10b981',
    error: '#ef4444',
    warning: '#f59e0b',

    // Текст
    textPrimary: '#fafafa',
    textSecondary: '#a1a1aa',
    textMuted: '#71717a',

    // Границы
    border: '#27272a',
    borderFocus: '#06b6d4',
    borderError: '#ef4444',
  },

  fonts: {
    regular: 'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
    mono: 'ui-monospace, "Cascadia Code", "Courier New", monospace',
  },

  fontSizes: {
    xs: 12,
    sm: 14,
    md: 16,
    lg: 18,
    xl: 24,
    xxl: 32,
    huge: 48,
  },

  spacing: {
    xs: 4,
    sm: 8,
    md: 16,
    lg: 24,
    xl: 32,
  },

  borderRadius: {
    sm: 6,
    md: 10,
    lg: 14,
  },
};

export type AppMode = 'professional' | 'personal';
export const XPLR_STORAGE_MODE = 'xplr_app_mode';

export default theme;
