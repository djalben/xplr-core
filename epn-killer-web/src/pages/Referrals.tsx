import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../api/axios';
import { theme } from '../theme/theme';

interface ReferralStats {
  total_referrals: number;
  active_referrals: number;
  total_commission: string;
  referral_code: string;
}

interface ReferralUser {
  id: number;
  email: string;
  status: string;
  commission: string;
  created_at: string;
}

const REFERRAL_BASE = 'https://xplr-web.vercel.app/register?ref=';

const Referrals: React.FC = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState<ReferralStats | null>(null);
  const [referrals, setReferrals] = useState<ReferralUser[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) { navigate('/login'); return; }
      const config = { headers: { Authorization: `Bearer ${token}` } };
      const res = await axios.get(`${API_BASE_URL}/user/referrals`, config);
      setStats(res.data);
      // referrals list is optional, may not exist yet
      try {
        const listRes = await axios.get(`${API_BASE_URL}/user/referrals/list`, config);
        setReferrals(Array.isArray(listRes.data) ? listRes.data : []);
      } catch { /* endpoint might not exist yet */ }
      setIsLoading(false);
    } catch (error) {
      console.error('Error fetching referral stats:', error);
      setIsLoading(false);
    }
  };

  const referralLink = stats?.referral_code ? `${REFERRAL_BASE}${stats.referral_code}` : '';

  const copyLink = () => {
    if (!referralLink) return;
    navigator.clipboard.writeText(referralLink);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (isLoading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', backgroundColor: theme.colors.background, color: theme.colors.textPrimary }}>
        Loading...
      </div>
    );
  }

  const revShareRate = '5%';

  return (
    <div style={{ minHeight: '100vh', backgroundColor: theme.colors.background, color: theme.colors.textPrimary, padding: '30px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '30px' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '28px', fontWeight: '800', letterSpacing: '-0.5px' }}>
            Referral Program
          </h1>
          <p style={{ margin: '5px 0 0', color: theme.colors.textSecondary, fontSize: '14px' }}>
            Invite friends & earn together
          </p>
        </div>
        <button onClick={() => navigate('/dashboard')} style={{
          padding: '10px 20px', backgroundColor: 'rgba(255,255,255,0.05)',
          border: `1px solid ${theme.colors.border}`, borderRadius: '8px',
          color: theme.colors.textSecondary, fontSize: '13px', cursor: 'pointer', fontWeight: '600'
        }}>← Dashboard</button>
      </div>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px', marginBottom: '30px' }}>
        {[
          { label: 'Total Referrals', value: stats?.total_referrals || 0, icon: '👥', color: '#3b82f6' },
          { label: 'Earned Bonuses', value: `$${parseFloat(stats?.total_commission || '0').toFixed(2)}`, icon: '💰', color: '#00e096' },
          { label: `RevShare (${revShareRate})`, value: revShareRate, icon: '📈', color: '#8b5cf6' },
        ].map(c => (
          <div key={c.label} style={{
            backgroundColor: theme.colors.backgroundCard,
            border: `1px solid ${theme.colors.border}`,
            borderRadius: '16px', padding: '24px', backdropFilter: 'blur(20px)'
          }}>
            <div style={{ fontSize: '12px', color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: '1px', marginBottom: '12px' }}>
              {c.icon} {c.label}
            </div>
            <div style={{ fontSize: '28px', fontWeight: '800', color: c.color }}>
              {c.value}
            </div>
          </div>
        ))}
      </div>

      {/* Referral Link */}
      <div style={{
        backgroundColor: theme.colors.backgroundCard,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: '16px', padding: '28px', marginBottom: '30px', backdropFilter: 'blur(20px)'
      }}>
        <h2 style={{ margin: '0 0 16px', fontSize: '18px', fontWeight: '700' }}>Your Referral Link</h2>
        <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
          <div style={{
            flex: 1, padding: '14px 16px',
            backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '10px',
            border: '1px solid rgba(255,255,255,0.1)', fontFamily: 'monospace',
            fontSize: '13px', color: '#00e096', fontWeight: '600',
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap'
          }}>
            {referralLink || 'Generating...'}
          </div>
          <button onClick={copyLink} style={{
            padding: '14px 24px', backgroundColor: copied ? '#14b8a6' : '#00e096',
            color: '#0a0a0a', border: 'none', borderRadius: '10px',
            fontWeight: '700', cursor: 'pointer', fontSize: '13px',
            whiteSpace: 'nowrap', transition: '0.2s'
          }}>
            {copied ? '✅ Copied!' : 'Copy Link'}
          </button>
        </div>
        <div style={{
          marginTop: '14px', padding: '12px 16px',
          backgroundColor: 'rgba(0, 224, 150, 0.08)', borderRadius: '10px',
          fontSize: '13px', color: theme.colors.textSecondary, lineHeight: '1.6'
        }}>
          🎁 New users who register via your link get a <strong style={{ color: '#00e096' }}>$5 bonus</strong>. You earn {revShareRate} RevShare from their transactions.
        </div>
      </div>

      {/* Referral List */}
      <div style={{
        backgroundColor: theme.colors.backgroundCard,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: '16px', padding: '24px', backdropFilter: 'blur(20px)'
      }}>
        <h2 style={{ margin: '0 0 20px', fontSize: '18px', fontWeight: '700' }}>
          Your Referrals ({referrals?.length ?? 0})
        </h2>
        {(referrals?.length ?? 0) === 0 ? (
          <div style={{ padding: '30px', textAlign: 'center', color: theme.colors.textSecondary, fontSize: '14px' }}>
            No referrals yet. Share your link to get started!
          </div>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                {['Email', 'Status', 'Commission', 'Joined'].map(h => (
                  <th key={h} style={{
                    textAlign: 'left', padding: '12px 8px', fontSize: '11px',
                    color: theme.colors.textSecondary, fontWeight: '600',
                    textTransform: 'uppercase', letterSpacing: '0.5px',
                    borderBottom: `1px solid ${theme.colors.border}`
                  }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {(referrals ?? []).map(r => (
                <tr key={r.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                  <td style={{ padding: '12px 8px', fontSize: '13px', fontWeight: '600' }}>{r.email}</td>
                  <td style={{ padding: '12px 8px' }}>
                    <span style={{
                      fontSize: '10px', padding: '3px 8px', borderRadius: '4px', fontWeight: '700',
                      backgroundColor: r.status === 'ACTIVE' ? 'rgba(0,224,150,0.15)' : 'rgba(255,107,107,0.15)',
                      color: r.status === 'ACTIVE' ? '#00e096' : '#ff6b6b'
                    }}>{r.status}</span>
                  </td>
                  <td style={{ padding: '12px 8px', fontSize: '13px', color: '#00e096', fontWeight: '600' }}>
                    ${parseFloat(r.commission || '0').toFixed(2)}
                  </td>
                  <td style={{ padding: '12px 8px', fontSize: '12px', color: theme.colors.textSecondary }}>
                    {r.created_at || '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* How it works */}
      <div style={{
        backgroundColor: theme.colors.backgroundCard,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: '16px', padding: '28px', marginTop: '30px', backdropFilter: 'blur(20px)'
      }}>
        <h2 style={{ margin: '0 0 20px', fontSize: '18px', fontWeight: '700' }}>How it works</h2>
        <div style={{ display: 'flex', gap: '24px' }}>
          {[
            { step: '1', title: 'Share your link', desc: 'Send your unique referral link to friends or post it on social media.' },
            { step: '2', title: 'Friend registers', desc: 'They sign up via your link and instantly receive a $5 welcome bonus.' },
            { step: '3', title: 'Earn commission', desc: `You earn ${revShareRate} from all their card transactions — forever.` },
          ].map(s => (
            <div key={s.step} style={{ flex: 1, display: 'flex', gap: '12px', alignItems: 'flex-start' }}>
              <div style={{
                width: '32px', height: '32px', borderRadius: '50%',
                backgroundColor: 'rgba(0, 224, 150, 0.15)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0, fontSize: '14px', fontWeight: '700', color: '#00e096'
              }}>{s.step}</div>
              <div>
                <div style={{ fontWeight: '600', marginBottom: '4px', fontSize: '14px' }}>{s.title}</div>
                <div style={{ fontSize: '13px', color: theme.colors.textSecondary, lineHeight: '1.5' }}>{s.desc}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Referrals;
