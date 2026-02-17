import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../api/axios';
import { theme, XPLR_STORAGE_MODE } from '../theme/theme';
import SidebarLayout from '../components/SidebarLayout';

interface Card {
  id: number;
  nickname: string;
  last_4_digits: string;
  bin: string;
  card_status: string;
  daily_spend_limit: number;
  auto_replenish_enabled?: boolean;
  auto_replenish_threshold?: string;
  auto_replenish_amount?: string;
  card_balance?: string;
  card_type?: string;
  category?: string;
  service_slug?: string;
  team_id?: number;
}

const CardsPage: React.FC = () => {
  const navigate = useNavigate();
  const [cards, setCards] = useState<Card[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const appMode = localStorage.getItem(XPLR_STORAGE_MODE) || 'professional';
  const isProfessional = appMode === 'professional';

  // Card creation state
  const [newCardType, setNewCardType] = useState<'VISA' | 'MasterCard'>('VISA');
  const [newCardCategory, setNewCardCategory] = useState<'arbitrage' | 'travel' | 'services'>('arbitrage');
  const [newCardNickname, setNewCardNickname] = useState('');
  const [newCardCount, setNewCardCount] = useState(1);
  const [isCreatingCard, setIsCreatingCard] = useState(false);

  const arbBins = [
    { bin: '487013', label: 'VISA 487013', fee: 5.00, topup: 2.5 },
    { bin: '539453', label: 'MC 539453', fee: 4.50, topup: 2.0 },
    { bin: '414720', label: 'VISA 414720', fee: 6.00, topup: 3.0 },
  ];
  const [selectedBin, setSelectedBin] = useState(arbBins[0].bin);

  const [exchangeRates, setExchangeRates] = useState<{ currency_from: string; currency_to: string; final_rate: string }[]>([]);

  // Card details
  const [revealedCardDetails, setRevealedCardDetails] = useState<Record<number, { full_number: string; cvv: string; expiry: string } | null>>({});
  const [loadingCardDetails, setLoadingCardDetails] = useState<Record<number, boolean>>({});

  const getConfig = () => ({ headers: { Authorization: `Bearer ${localStorage.getItem('token')}` } });

  useEffect(() => {
    fetchCards();
    fetchRates();
  }, []);

  useEffect(() => { if (toast) { const t = setTimeout(() => setToast(null), 3000); return () => clearTimeout(t); } }, [toast]);

  const fetchCards = async () => {
    try {
      const res = await axios.get(`${API_BASE_URL}/user/cards`, getConfig());
      setCards(Array.isArray(res.data) ? res.data : []);
    } catch (e) { console.error('Error fetching cards:', e); }
    finally { setIsLoading(false); }
  };

  const fetchRates = async () => {
    try {
      const res = await axios.get(`${API_BASE_URL}/exchange-rates`, getConfig());
      setExchangeRates(Array.isArray(res.data) ? res.data : []);
    } catch (e) { /* ignore */ }
  };

  const handleCreateCard = async () => {
    if (isCreatingCard) return;
    setIsCreatingCard(true);
    try {
      const requestData = {
        count: isProfessional ? newCardCount : 1,
        nickname: newCardNickname,
        daily_limit: 500,
        merchant_name: newCardNickname || 'Default',
        card_type: newCardType,
        category: isProfessional ? 'arbitrage' : newCardCategory,
      };
      await axios.post(`${API_BASE_URL}/user/cards/issue`, requestData, getConfig());
      await fetchCards();
      setToast({ message: 'Карта успешно создана!', type: 'success' });
      setTimeout(() => { setShowCreateModal(false); setNewCardNickname(''); }, 200);
    } catch (error: any) {
      setToast({ message: error?.response?.data || 'Ошибка создания карты', type: 'error' });
    } finally { setIsCreatingCard(false); }
  };

  const handleToggleBlock = async (cardId: number, currentStatus: string) => {
    const action = currentStatus === 'BLOCKED' ? 'unblock' : 'block';
    try {
      await axios.patch(`${API_BASE_URL}/user/cards/${cardId}/status`, { action }, getConfig());
      await fetchCards();
      setToast({ message: `Card ${action}ed`, type: 'success' });
    } catch { setToast({ message: 'Failed to update card', type: 'error' }); }
  };

  const handleReveal = async (cardId: number) => {
    if (revealedCardDetails[cardId]) { setRevealedCardDetails(p => ({ ...p, [cardId]: null })); return; }
    setLoadingCardDetails(p => ({ ...p, [cardId]: true }));
    try {
      const res = await axios.get(`${API_BASE_URL}/user/cards/${cardId}/mock-details`, getConfig());
      setRevealedCardDetails(p => ({ ...p, [cardId]: res.data }));
    } catch { setToast({ message: 'Failed to load details', type: 'error' }); }
    finally { setLoadingCardDetails(p => ({ ...p, [cardId]: false })); }
  };

  const filteredCards = isProfessional
    ? (cards ?? []).filter(c => (c.category || 'arbitrage') === 'arbitrage')
    : (cards ?? []).filter(c => { const cat = c.category || 'arbitrage'; return cat === 'travel' || cat === 'services'; });

  const catColors: Record<string, { bg: string; border: string }> = {
    arbitrage: { bg: 'rgba(30, 58, 95, 0.7)', border: 'rgba(59, 130, 246, 0.3)' },
    travel:    { bg: 'rgba(19, 78, 74, 0.7)', border: 'rgba(20, 184, 166, 0.3)' },
    services:  { bg: 'rgba(40, 40, 50, 0.7)', border: 'rgba(120, 120, 140, 0.3)' },
  };

  if (isLoading) {
    return <SidebarLayout><div style={{ padding: 40, textAlign: 'center', color: theme.colors.textSecondary }}>Загрузка...</div></SidebarLayout>;
  }

  return (
    <SidebarLayout>
      {/* Toast */}
      {toast && (
        <div style={{
          position: 'fixed', top: 20, right: 20, zIndex: 99999,
          padding: '14px 24px', borderRadius: 12,
          backgroundColor: toast.type === 'success' ? 'rgba(16, 185, 129, 0.9)' : 'rgba(239, 68, 68, 0.9)',
          color: '#fff', fontWeight: 600, fontSize: 14, backdropFilter: 'blur(12px)'
        }}>{toast.message}</div>
      )}

      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <button onClick={() => navigate('/dashboard')} style={{
            padding: '8px 16px', backgroundColor: 'rgba(255,255,255,0.04)',
            border: `1px solid ${theme.colors.border}`, borderRadius: 8,
            color: theme.colors.textSecondary, fontSize: 13, cursor: 'pointer', fontWeight: 600
          }}>← Назад</button>
          <div>
            <h1 style={{ margin: 0, fontSize: 24, fontWeight: 700 }}>
              {isProfessional ? 'Арбитражные карты' : 'Личные карты'}
            </h1>
            <p style={{ margin: '4px 0 0', color: theme.colors.textSecondary, fontSize: 13 }}>
              {filteredCards.length} карт
            </p>
          </div>
        </div>
        <button onClick={() => { setNewCardCategory(isProfessional ? 'arbitrage' : 'travel'); setNewCardCount(1); setShowCreateModal(true); }}
          style={{
            padding: '12px 24px', backgroundColor: theme.colors.accent,
            color: theme.colors.background, border: 'none', borderRadius: 8,
            fontWeight: 700, fontSize: 14, cursor: 'pointer'
          }}>
          {isProfessional ? '+ Массовый выпуск' : '+ Выпустить карту'}
        </button>
      </div>

      {/* Fee info (arbitrage only) */}
      {isProfessional && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 12, marginBottom: 24 }}>
          {arbBins.map(b => (
            <div key={b.bin} style={{
              padding: '12px 16px', backgroundColor: theme.colors.backgroundCard,
              border: `1px solid ${theme.colors.border}`, borderRadius: 10, fontSize: 12
            }}>
              <div style={{ color: '#fff', fontWeight: 700, marginBottom: 4 }}>{b.label}</div>
              <div style={{ color: theme.colors.textSecondary }}>Выпуск: ${b.fee.toFixed(2)} · Пополнение: {b.topup}%</div>
            </div>
          ))}
        </div>
      )}

      {/* Cards Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 20 }}>
        {filteredCards.length === 0 ? (
          <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: 60, color: theme.colors.textSecondary, fontSize: 15 }}>
            Нет карт. Нажмите кнопку выпуска.
          </div>
        ) : filteredCards.map(card => {
          const cc = catColors[(card.category || 'arbitrage')] || catColors.arbitrage;
          const details = revealedCardDetails[card.id];
          const isLoadingDet = loadingCardDetails[card.id];
          return (
            <div key={card.id} style={{
              backgroundColor: card.card_status === 'BLOCKED' ? theme.colors.backgroundCard : cc.bg,
              backdropFilter: 'blur(20px)',
              border: card.card_status === 'BLOCKED' ? `1px solid ${theme.colors.error}` : `1px solid ${cc.border}`,
              borderRadius: theme.borderRadius.lg, padding: 20,
              opacity: card.card_status === 'CLOSED' ? 0.5 : card.card_status === 'BLOCKED' ? 0.7 : 1,
              transition: '0.3s'
            }}>
              {/* Card header */}
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ fontWeight: 700, fontStyle: 'italic', color: '#fff' }}>
                    {card.bin?.startsWith('4') ? 'VISA' : 'MC'}
                  </span>
                  {isProfessional && card.bin && (
                    <span style={{ fontSize: 10, padding: '3px 6px', borderRadius: 4, background: 'rgba(59,130,246,0.15)', color: '#3b82f6', fontFamily: theme.fonts.mono }}>
                      BIN {card.bin}
                    </span>
                  )}
                </div>
                <span style={{
                  fontSize: 10, padding: '4px 8px', borderRadius: 4, fontWeight: 700,
                  background: card.card_status === 'ACTIVE' ? theme.colors.accentMuted : `${theme.colors.error}30`,
                  color: card.card_status === 'ACTIVE' ? theme.colors.accent : theme.colors.error
                }}>{card.card_status}</span>
              </div>

              {/* Card number */}
              <div style={{ fontFamily: theme.fonts.mono, fontSize: 17, letterSpacing: 2, color: theme.colors.textSecondary, marginBottom: 10, cursor: 'pointer' }}
                onClick={() => handleReveal(card.id)}>
                {isLoadingDet ? '...' : details ? details.full_number.replace(/(.{4})/g, '$1 ').trim() : `•••• •••• •••• ${card.last_4_digits}`}
              </div>

              {details && (
                <div style={{ display: 'flex', gap: 20, fontSize: 13, color: '#fff', marginBottom: 10, padding: '10px 12px', backgroundColor: 'rgba(0,0,0,0.25)', borderRadius: 8, fontFamily: theme.fonts.mono }}>
                  <div><span style={{ color: '#888', fontSize: 10 }}>CVV</span><br />{details.cvv}</div>
                  <div><span style={{ color: '#888', fontSize: 10 }}>EXPIRY</span><br />{details.expiry}</div>
                </div>
              )}

              {/* Card info */}
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: theme.colors.textSecondary, marginBottom: 12 }}>
                <span>{card.nickname}</span>
                <span>Лимит: ${card.daily_spend_limit || 0}/day</span>
              </div>

              {/* Actions */}
              <div style={{ display: 'flex', gap: 8 }}>
                {card.card_status !== 'CLOSED' && (
                  <button onClick={() => handleToggleBlock(card.id, card.card_status)}
                    style={{
                      flex: 1, padding: '8px 12px', fontSize: 12, fontWeight: 600, cursor: 'pointer',
                      borderRadius: 6, border: 'none', transition: '0.2s',
                      backgroundColor: card.card_status === 'BLOCKED' ? 'rgba(0,224,150,0.15)' : 'rgba(255,59,59,0.15)',
                      color: card.card_status === 'BLOCKED' ? '#00e096' : '#ff3b3b'
                    }}>
                    {card.card_status === 'BLOCKED' ? 'Разблокировать' : 'Заблокировать'}
                  </button>
                )}
                <button onClick={() => handleReveal(card.id)}
                  style={{
                    flex: 1, padding: '8px 12px', fontSize: 12, fontWeight: 600, cursor: 'pointer',
                    borderRadius: 6, border: `1px solid ${theme.colors.border}`,
                    backgroundColor: 'transparent', color: theme.colors.textSecondary
                  }}>
                  {details ? 'Скрыть' : 'Показать'}
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {/* ===== CREATE CARD MODAL ===== */}
      {showCreateModal && (
        <div style={{
          position: 'fixed', top: 0, left: 0, width: '100vw', height: '100vh',
          backgroundColor: 'rgba(0,0,0,0.8)', backdropFilter: 'blur(10px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 10000
        }} onClick={() => setShowCreateModal(false)}>
          <div style={{
            backgroundColor: theme.colors.backgroundElevated, backdropFilter: 'blur(40px)',
            borderRadius: 24, padding: 36, width: '90%', maxWidth: 480,
            border: `1px solid ${theme.colors.border}`, boxShadow: '0 20px 60px rgba(0,0,0,0.5)'
          }} onClick={e => e.stopPropagation()}>
            <h2 style={{ margin: '0 0 24px', fontSize: 24, fontWeight: 700 }}>
              {isProfessional ? 'Массовый выпуск' : 'Выпуск карты'}
            </h2>

            {/* BIN / Category selector */}
            <div style={{ marginBottom: 20 }}>
              <label style={{ display: 'block', fontSize: 11, fontWeight: 600, color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 10 }}>
                {isProfessional ? 'Выберите BIN' : 'Тип карты'}
              </label>
              {isProfessional ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {arbBins.map(b => (
                    <button key={b.bin} onClick={() => { setSelectedBin(b.bin); setNewCardType(b.bin.startsWith('4') ? 'VISA' : 'MasterCard'); }}
                      style={{
                        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                        padding: '14px 16px',
                        backgroundColor: selectedBin === b.bin ? 'rgba(59,130,246,0.2)' : 'rgba(255,255,255,0.04)',
                        border: selectedBin === b.bin ? '2px solid #3b82f6' : `2px solid ${theme.colors.border}`,
                        borderRadius: 12, cursor: 'pointer', textAlign: 'left'
                      }}>
                      <div>
                        <div style={{ color: selectedBin === b.bin ? '#3b82f6' : '#fff', fontWeight: 600, fontSize: 14 }}>{b.label}</div>
                        <div style={{ color: theme.colors.textSecondary, fontSize: 11, marginTop: 2 }}>Пополнение: {b.topup}%</div>
                      </div>
                      <div style={{ color: selectedBin === b.bin ? '#3b82f6' : theme.colors.textSecondary, fontSize: 15, fontWeight: 700 }}>${b.fee.toFixed(2)}</div>
                    </button>
                  ))}
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {([
                    { key: 'travel' as const, label: 'Для путешествий', desc: 'Отели, авиабилеты • 1 год', feeUsd: 3 },
                    { key: 'services' as const, label: 'Для зарубежных сервисов', desc: 'Подписки, онлайн-сервисы • 1 год', feeUsd: 2 },
                  ]).map(cat => {
                    const rubRate = (exchangeRates ?? []).find(r => r.currency_to === 'USD');
                    const rate = rubRate ? parseFloat(rubRate.final_rate) : 99;
                    const priceRub = Math.ceil(cat.feeUsd * rate);
                    return (
                      <button key={cat.key} onClick={() => setNewCardCategory(cat.key)}
                        style={{
                          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                          padding: '14px 16px',
                          backgroundColor: newCardCategory === cat.key ? 'rgba(0,224,150,0.15)' : 'rgba(255,255,255,0.04)',
                          border: newCardCategory === cat.key ? '2px solid #00e096' : `2px solid ${theme.colors.border}`,
                          borderRadius: 12, cursor: 'pointer', textAlign: 'left'
                        }}>
                        <div>
                          <div style={{ color: newCardCategory === cat.key ? '#00e096' : '#fff', fontWeight: 600, fontSize: 14 }}>{cat.label}</div>
                          <div style={{ color: theme.colors.textSecondary, fontSize: 11, marginTop: 2 }}>{cat.desc}</div>
                        </div>
                        <div style={{ color: newCardCategory === cat.key ? '#00e096' : theme.colors.textSecondary, fontSize: 15, fontWeight: 700 }}>{priceRub}₽</div>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Quantity (arbitrage only) */}
            {isProfessional && (
              <div style={{ marginBottom: 20 }}>
                <label style={{ display: 'block', fontSize: 11, fontWeight: 600, color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 8 }}>
                  Количество (1–100)
                </label>
                <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                  <input type="number" min={1} max={100} value={newCardCount}
                    onChange={e => setNewCardCount(Math.max(1, Math.min(100, parseInt(e.target.value) || 1)))}
                    style={{
                      width: 70, padding: 10, backgroundColor: 'rgba(255,255,255,0.04)',
                      border: `1px solid ${theme.colors.border}`, borderRadius: 8,
                      color: '#3b82f6', fontWeight: 700, fontSize: 16, textAlign: 'center', outline: 'none'
                    }} />
                  {[5, 10, 25, 50].map(q => (
                    <button key={q} onClick={() => setNewCardCount(q)}
                      style={{
                        padding: '8px 14px', fontSize: 13, fontWeight: 700, cursor: 'pointer',
                        borderRadius: 8, transition: '0.2s',
                        backgroundColor: newCardCount === q ? 'rgba(59,130,246,0.2)' : 'rgba(255,255,255,0.04)',
                        border: newCardCount === q ? '1px solid #3b82f6' : `1px solid ${theme.colors.border}`,
                        color: newCardCount === q ? '#3b82f6' : theme.colors.textSecondary
                      }}>{q}</button>
                  ))}
                </div>
                <div style={{ fontSize: 11, color: theme.colors.textSecondary, marginTop: 6 }}>
                  Итого: ${((arbBins.find(b => b.bin === selectedBin)?.fee ?? 5) * newCardCount).toFixed(2)}
                </div>
              </div>
            )}

            {/* Card Type (personal only — VISA/MC) */}
            {!isProfessional && (
              <div style={{ marginBottom: 20 }}>
                <label style={{ display: 'block', fontSize: 11, fontWeight: 600, color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 8 }}>Card Type</label>
                <div style={{ display: 'flex', gap: 10 }}>
                  {(['VISA', 'MasterCard'] as const).map(t => (
                    <button key={t} onClick={() => setNewCardType(t)} style={{
                      flex: 1, padding: 14, borderRadius: 10, fontWeight: 700, fontSize: 15, cursor: 'pointer',
                      backgroundColor: newCardType === t ? 'rgba(0,224,150,0.15)' : 'rgba(255,255,255,0.04)',
                      border: newCardType === t ? '2px solid #00e096' : `2px solid ${theme.colors.border}`,
                      color: newCardType === t ? '#00e096' : theme.colors.textSecondary
                    }}>{t}</button>
                  ))}
                </div>
              </div>
            )}

            {/* Nickname */}
            <div style={{ marginBottom: 24 }}>
              <label style={{ display: 'block', fontSize: 11, fontWeight: 600, color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 8 }}>Nickname</label>
              <input type="text" placeholder="e.g. FB Campaign" value={newCardNickname}
                onChange={e => setNewCardNickname(e.target.value)}
                style={{
                  width: '100%', padding: 14, backgroundColor: 'rgba(255,255,255,0.04)',
                  border: `1px solid ${theme.colors.border}`, borderRadius: 10,
                  color: '#fff', fontSize: 15, outline: 'none', boxSizing: 'border-box'
                }} />
            </div>

            {/* Buttons */}
            <div style={{ display: 'flex', gap: 12 }}>
              <button onClick={() => setShowCreateModal(false)} style={{
                flex: 1, padding: 14, borderRadius: 10, fontWeight: 600, fontSize: 15, cursor: 'pointer',
                backgroundColor: 'rgba(255,255,255,0.04)', border: `1px solid ${theme.colors.border}`, color: theme.colors.textSecondary
              }}>Отмена</button>
              <button onClick={handleCreateCard} disabled={isCreatingCard} style={{
                flex: 1, padding: 14, borderRadius: 10, fontWeight: 700, fontSize: 15,
                cursor: isCreatingCard ? 'not-allowed' : 'pointer', opacity: isCreatingCard ? 0.5 : 1,
                backgroundColor: isProfessional ? '#3b82f6' : '#00e096',
                border: 'none', color: isProfessional ? '#fff' : '#000'
              }}>
                {isCreatingCard ? 'Создание...' : isProfessional
                  ? `Выпустить ${newCardCount} — $${((arbBins.find(b => b.bin === selectedBin)?.fee ?? 5) * newCardCount).toFixed(2)}`
                  : (() => { const r = (exchangeRates ?? []).find(r => r.currency_to === 'USD'); const rate = r ? parseFloat(r.final_rate) : 99; return `Выпустить — ${Math.ceil((newCardCategory === 'travel' ? 3 : 2) * rate)}₽`; })()
                }
              </button>
            </div>
          </div>
        </div>
      )}
    </SidebarLayout>
  );
};

export default CardsPage;
