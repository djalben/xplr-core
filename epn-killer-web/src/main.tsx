import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './app';
import './styles.css';

class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error: Error | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: 40, fontFamily: 'monospace', color: '#ef4444', background: '#0a0a0f', minHeight: '100vh' }}>
          <h1 style={{ color: '#fafafa' }}>XPLR — Ошибка приложения</h1>
          <pre style={{ whiteSpace: 'pre-wrap', marginTop: 16 }}>
            {this.state.error?.message}
          </pre>
          <p style={{ color: '#71717a', marginTop: 24 }}>
            Откройте консоль браузера (F12) для деталей.
          </p>
        </div>
      );
    }
    return this.props.children;
  }
}

const root = document.getElementById('root');

if (!root) {
  document.body.innerHTML =
    '<div style="padding:40px;font-family:monospace;color:#ef4444">Fatal: #root element not found</div>';
} else {
  ReactDOM.createRoot(root).render(
    <React.StrictMode>
      <ErrorBoundary>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </ErrorBoundary>
    </React.StrictMode>
  );
}
