import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getReferralStats, ReferralStats } from '../api/referrals';

const Referrals: React.FC = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState<ReferralStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    fetchStats();
  }, []);

  const fetchStats = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        navigate('/login');
        return;
      }
      const data = await getReferralStats();
      setStats(data);
      setIsLoading(false);
    } catch (error) {
      console.error('Error fetching referral stats:', error);
      setIsLoading(false);
    }
  };

  const copyReferralCode = () => {
    if (stats?.referral_code) {
      navigator.clipboard.writeText(stats.referral_code);
      alert('Реферальный код скопирован!');
    }
  };

  if (isLoading) {
    return (
      <div style={{ padding: '40px', textAlign: 'center', color: '#888c95' }}>
        Загрузка...
      </div>
    );
  }

  return (
    <div style={{
      minHeight: '100vh',
      backgroundColor: '#000',
      color: '#ffffff',
      padding: '30px'
    }}>
      <h1 style={{ margin: '0 0 30px 0', fontSize: '24px', fontWeight: '700' }}>Реферальная программа</h1>

      {/* Stats Cards */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
        gap: '20px',
        marginBottom: '30px'
      }}>
        <div style={{
          backgroundColor: 'rgba(255, 255, 255, 0.03)',
          borderRadius: '12px',
          padding: '24px',
          border: '1px solid rgba(255, 255, 255, 0.1)'
        }}>
          <div style={{ fontSize: '12px', color: '#888c95', marginBottom: '8px', textTransform: 'uppercase' }}>
            Всего рефералов
          </div>
          <div style={{ fontSize: '32px', fontWeight: '700', color: '#ffffff' }}>
            {stats?.total_referrals || 0}
          </div>
        </div>

        <div style={{
          backgroundColor: 'rgba(255, 255, 255, 0.03)',
          borderRadius: '12px',
          padding: '24px',
          border: '1px solid rgba(255, 255, 255, 0.1)'
        }}>
          <div style={{ fontSize: '12px', color: '#888c95', marginBottom: '8px', textTransform: 'uppercase' }}>
            Активных
          </div>
          <div style={{ fontSize: '32px', fontWeight: '700', color: '#00e096' }}>
            {stats?.active_referrals || 0}
          </div>
        </div>

        <div style={{
          backgroundColor: 'rgba(255, 255, 255, 0.03)',
          borderRadius: '12px',
          padding: '24px',
          border: '1px solid rgba(255, 255, 255, 0.1)'
        }}>
          <div style={{ fontSize: '12px', color: '#888c95', marginBottom: '8px', textTransform: 'uppercase' }}>
            Заработано
          </div>
          <div style={{ fontSize: '32px', fontWeight: '700', color: '#00e096' }}>
            ${parseFloat(stats?.total_commission || '0').toFixed(2)}
          </div>
        </div>
      </div>

      {/* Referral Code */}
      <div style={{
        backgroundColor: 'rgba(255, 255, 255, 0.03)',
        borderRadius: '12px',
        padding: '30px',
        border: '1px solid rgba(255, 255, 255, 0.1)',
        marginBottom: '30px'
      }}>
        <h2 style={{ margin: '0 0 20px 0', fontSize: '18px', fontWeight: '600' }}>Ваш реферальный код</h2>
        <div style={{
          display: 'flex',
          gap: '12px',
          alignItems: 'center'
        }}>
          <div style={{
            flex: 1,
            padding: '16px',
            backgroundColor: 'rgba(255, 255, 255, 0.05)',
            borderRadius: '8px',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            fontFamily: 'monospace',
            fontSize: '18px',
            color: '#00e096',
            fontWeight: '600',
            letterSpacing: '2px'
          }}>
            {stats?.referral_code || 'Генерация...'}
          </div>
          <button
            onClick={copyReferralCode}
            style={{
              padding: '16px 24px',
              backgroundColor: '#00e096',
              color: '#000',
              border: 'none',
              borderRadius: '8px',
              fontWeight: '600',
              cursor: 'pointer',
              fontSize: '14px'
            }}>
            Копировать
          </button>
        </div>
        <div style={{
          marginTop: '16px',
          padding: '12px',
          backgroundColor: 'rgba(0, 224, 150, 0.1)',
          borderRadius: '8px',
          fontSize: '13px',
          color: '#888c95',
          lineHeight: '1.6'
        }}>
          Поделитесь этим кодом с друзьями. За каждого приглашенного пользователя вы получите комиссию с его транзакций.
        </div>
      </div>

      {/* How it works */}
      <div style={{
        backgroundColor: 'rgba(255, 255, 255, 0.03)',
        borderRadius: '12px',
        padding: '30px',
        border: '1px solid rgba(255, 255, 255, 0.1)'
      }}>
        <h2 style={{ margin: '0 0 20px 0', fontSize: '18px', fontWeight: '600' }}>Как это работает</h2>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              backgroundColor: 'rgba(0, 224, 150, 0.2)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
              fontSize: '18px'
            }}>
              1
            </div>
            <div>
              <div style={{ fontWeight: '600', marginBottom: '4px' }}>Поделитесь реферальным кодом</div>
              <div style={{ fontSize: '13px', color: '#888c95' }}>
                Отправьте ваш реферальный код друзьям или разместите его в социальных сетях
              </div>
            </div>
          </div>
          <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              backgroundColor: 'rgba(0, 224, 150, 0.2)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
              fontSize: '18px'
            }}>
              2
            </div>
            <div>
              <div style={{ fontWeight: '600', marginBottom: '4px' }}>Реферал регистрируется</div>
              <div style={{ fontSize: '13px', color: '#888c95' }}>
                При регистрации ваш друг должен указать ваш реферальный код
              </div>
            </div>
          </div>
          <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              backgroundColor: 'rgba(0, 224, 150, 0.2)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
              fontSize: '18px'
            }}>
              3
            </div>
            <div>
              <div style={{ fontWeight: '600', marginBottom: '4px' }}>Получайте комиссию</div>
              <div style={{ fontSize: '13px', color: '#888c95' }}>
                Вы будете получать процент от всех транзакций ваших рефералов
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Referrals;
