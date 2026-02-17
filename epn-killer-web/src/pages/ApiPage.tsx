import React from 'react';
import { useNavigate } from 'react-router-dom';
import { theme } from '../theme/theme';
import SidebarLayout from '../components/SidebarLayout';

const ApiPage: React.FC = () => {
  const navigate = useNavigate();

  return (
    <SidebarLayout>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 32 }}>
        <button onClick={() => navigate('/dashboard')} style={{
          padding: '8px 16px', backgroundColor: 'rgba(255,255,255,0.04)',
          border: `1px solid ${theme.colors.border}`, borderRadius: 8,
          color: theme.colors.textSecondary, fontSize: 13, cursor: 'pointer', fontWeight: 600
        }}>← Назад</button>
        <h1 style={{ margin: 0, fontSize: 24, fontWeight: 700 }}>API & Trackers</h1>
      </div>

      {/* Content */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(340px, 1fr))', gap: 24 }}>
        {/* API Base URL */}
        <div style={{
          backgroundColor: theme.colors.backgroundCard, borderRadius: 14,
          border: `1px solid ${theme.colors.border}`, padding: 28
        }}>
          <div style={{ fontSize: 40, marginBottom: 16 }}>🔌</div>
          <h3 style={{ margin: '0 0 8px', fontSize: 18, fontWeight: 700 }}>REST API</h3>
          <p style={{ color: theme.colors.textSecondary, fontSize: 13, lineHeight: 1.6, margin: '0 0 20px' }}>
            Используйте API для автоматизации выпуска карт, управления балансами и получения транзакций.
          </p>
          <div style={{
            padding: 14, backgroundColor: 'rgba(255,255,255,0.03)', borderRadius: 10,
            border: `1px solid ${theme.colors.border}`, fontFamily: theme.fonts.mono, fontSize: 13
          }}>
            <div style={{ color: theme.colors.textSecondary, fontSize: 10, marginBottom: 6, textTransform: 'uppercase' }}>Base URL</div>
            <code style={{ color: theme.colors.accent }}>https://xplr-web.vercel.app/api/v1/</code>
          </div>
        </div>

        {/* Webhooks */}
        <div style={{
          backgroundColor: theme.colors.backgroundCard, borderRadius: 14,
          border: `1px solid ${theme.colors.border}`, padding: 28
        }}>
          <div style={{ fontSize: 40, marginBottom: 16 }}>🔔</div>
          <h3 style={{ margin: '0 0 8px', fontSize: 18, fontWeight: 700 }}>Webhooks</h3>
          <p style={{ color: theme.colors.textSecondary, fontSize: 13, lineHeight: 1.6, margin: '0 0 20px' }}>
            Настройте уведомления о транзакциях, изменениях статуса карт и пополнениях в реальном времени.
          </p>
          <div style={{
            padding: 14, backgroundColor: 'rgba(255,255,255,0.03)', borderRadius: 10,
            border: `1px solid ${theme.colors.border}`, fontSize: 13, color: theme.colors.textMuted
          }}>
            Скоро будет доступно
          </div>
        </div>

        {/* Trackers */}
        <div style={{
          backgroundColor: theme.colors.backgroundCard, borderRadius: 14,
          border: `1px solid ${theme.colors.border}`, padding: 28
        }}>
          <div style={{ fontSize: 40, marginBottom: 16 }}>📊</div>
          <h3 style={{ margin: '0 0 8px', fontSize: 18, fontWeight: 700 }}>Trackers</h3>
          <p style={{ color: theme.colors.textSecondary, fontSize: 13, lineHeight: 1.6, margin: '0 0 20px' }}>
            Подключите Keitaro, Binom или другой трекер для автоматического управления рекламными расходами.
          </p>
          <div style={{
            padding: 14, backgroundColor: 'rgba(255,255,255,0.03)', borderRadius: 10,
            border: `1px solid ${theme.colors.border}`, fontSize: 13, color: theme.colors.textMuted
          }}>
            Скоро будет доступно
          </div>
        </div>
      </div>

      {/* Docs section */}
      <div style={{
        marginTop: 32, padding: 24,
        backgroundColor: theme.colors.backgroundCard, borderRadius: 14,
        border: `1px solid ${theme.colors.border}`
      }}>
        <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 700 }}>Быстрый старт</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          {[
            { method: 'GET', path: '/user/me', desc: 'Получить данные пользователя' },
            { method: 'GET', path: '/user/cards', desc: 'Список всех карт' },
            { method: 'POST', path: '/user/cards/issue', desc: 'Выпустить карты' },
            { method: 'POST', path: '/user/topup', desc: 'Пополнить баланс' },
          ].map(ep => (
            <div key={ep.path} style={{
              padding: 12, borderRadius: 8,
              backgroundColor: 'rgba(255,255,255,0.02)',
              border: `1px solid ${theme.colors.border}`, fontSize: 13
            }}>
              <span style={{
                display: 'inline-block', padding: '2px 8px', borderRadius: 4, fontSize: 10,
                fontWeight: 700, fontFamily: theme.fonts.mono, marginRight: 8,
                backgroundColor: ep.method === 'GET' ? 'rgba(59,130,246,0.15)' : 'rgba(0,224,150,0.15)',
                color: ep.method === 'GET' ? '#3b82f6' : '#00e096'
              }}>{ep.method}</span>
              <code style={{ color: theme.colors.textPrimary, fontFamily: theme.fonts.mono }}>{ep.path}</code>
              <div style={{ color: theme.colors.textSecondary, fontSize: 11, marginTop: 4 }}>{ep.desc}</div>
            </div>
          ))}
        </div>
      </div>
    </SidebarLayout>
  );
};

export default ApiPage;
