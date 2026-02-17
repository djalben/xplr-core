// XPLR Premium — тёмная премиальная схема с акцентами в стиле "X"
export const theme = {
  colors: {
    // Фоны (Deep Navy)
    background: 'rgb(10, 10, 14)',
    backgroundSecondary: 'rgb(14, 15, 22)',
    backgroundCard: 'rgb(20, 21, 30)',
    backgroundElevated: 'rgb(26, 27, 38)',

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
