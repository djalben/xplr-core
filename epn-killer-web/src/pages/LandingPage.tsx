import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const LandingPage: React.FC = () => {
  const navigate = useNavigate();
  const [scrolled, setScrolled] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const token = localStorage.getItem('token');

  /* ─── Cross-browser blur helper ─── */
  const blur = (px: number): React.CSSProperties => ({
    WebkitBackdropFilter: `blur(${px}px)`,
    backdropFilter: `blur(${px}px)`,
    MozBackdropFilter: `blur(${px}px)`,
  } as React.CSSProperties);

  /* ─── Styles ─── */
  const css = {
    page: {
      minHeight: '100vh',
      background: '#000',
      color: '#f5f5f7',
      fontFamily: '-apple-system, BlinkMacSystemFont, "SF Pro Display", "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
      WebkitFontSmoothing: 'antialiased' as const,
      MozOsxFontSmoothing: 'grayscale' as const,
      overflowX: 'hidden' as const,
    },
    /* ─── HEADER ─── */
    header: {
      position: 'fixed' as const,
      top: 0, left: 0, right: 0,
      zIndex: 100,
      padding: scrolled ? '12px 0' : '18px 0',
      background: scrolled ? 'rgba(0,0,0,0.72)' : 'transparent',
      ...blur(scrolled ? 20 : 0),
      borderBottom: scrolled ? '1px solid rgba(255,255,255,0.08)' : 'none',
      transition: 'all 0.35s cubic-bezier(.4,0,.2,1)',
    },
    headerInner: {
      maxWidth: '1120px',
      margin: '0 auto',
      padding: '0 24px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    logo: {
      fontSize: '22px',
      fontWeight: 700,
      letterSpacing: '-0.5px',
      color: '#f5f5f7',
      cursor: 'pointer',
    },
    nav: {
      display: 'flex' as const,
      gap: '32px',
      alignItems: 'center',
    },
    navLink: {
      color: 'rgba(245,245,247,0.7)',
      fontSize: '14px',
      fontWeight: 400,
      textDecoration: 'none',
      cursor: 'pointer',
      transition: 'color 0.2s',
    },
    ctaSmall: {
      padding: '8px 20px',
      borderRadius: '980px',
      background: '#0071e3',
      color: '#fff',
      fontSize: '14px',
      fontWeight: 500,
      border: 'none',
      cursor: 'pointer',
      transition: 'background 0.2s',
    },
    /* ─── HERO ─── */
    hero: {
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column' as const,
      alignItems: 'center',
      justifyContent: 'center',
      textAlign: 'center' as const,
      padding: '120px 24px 80px',
      position: 'relative' as const,
      overflow: 'hidden' as const,
    },
    heroGlow: {
      position: 'absolute' as const,
      width: '800px', height: '800px',
      borderRadius: '50%',
      background: 'radial-gradient(circle, rgba(0,113,227,0.15) 0%, transparent 70%)',
      top: '50%', left: '50%',
      transform: 'translate(-50%, -50%)',
      pointerEvents: 'none' as const,
      filter: 'blur(60px)',
      WebkitFilter: 'blur(60px)',
    },
    heroTag: {
      fontSize: '16px',
      fontWeight: 500,
      color: '#0071e3',
      letterSpacing: '0.02em',
      marginBottom: '16px',
    },
    heroTitle: {
      fontSize: 'clamp(40px, 8vw, 80px)',
      fontWeight: 700,
      letterSpacing: '-0.03em',
      lineHeight: 1.05,
      margin: '0 0 24px',
      maxWidth: '800px',
      background: 'linear-gradient(180deg, #f5f5f7 30%, rgba(245,245,247,0.5))',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      backgroundClip: 'text',
    } as React.CSSProperties,
    heroSub: {
      fontSize: 'clamp(17px, 2.5vw, 21px)',
      fontWeight: 400,
      lineHeight: 1.5,
      color: 'rgba(245,245,247,0.6)',
      maxWidth: '540px',
      margin: '0 0 40px',
    },
    heroCta: {
      display: 'inline-flex',
      gap: '16px',
      flexWrap: 'wrap' as const,
      justifyContent: 'center',
    },
    btnPrimary: {
      padding: '16px 36px',
      borderRadius: '980px',
      background: '#0071e3',
      color: '#fff',
      fontSize: '17px',
      fontWeight: 500,
      border: 'none',
      cursor: 'pointer',
      transition: 'all 0.25s',
      letterSpacing: '-0.01em',
    },
    btnSecondary: {
      padding: '16px 36px',
      borderRadius: '980px',
      background: 'transparent',
      color: '#0071e3',
      fontSize: '17px',
      fontWeight: 500,
      border: '1px solid rgba(0,113,227,0.4)',
      cursor: 'pointer',
      transition: 'all 0.25s',
      letterSpacing: '-0.01em',
    },
    /* ─── SECTION (shared) ─── */
    section: {
      maxWidth: '1120px',
      margin: '0 auto',
      padding: '100px 24px',
    },
    sectionTitle: {
      fontSize: 'clamp(28px, 5vw, 48px)',
      fontWeight: 700,
      letterSpacing: '-0.025em',
      lineHeight: 1.1,
      textAlign: 'center' as const,
      marginBottom: '16px',
    },
    sectionSub: {
      fontSize: 'clamp(16px, 2vw, 19px)',
      color: 'rgba(245,245,247,0.55)',
      textAlign: 'center' as const,
      maxWidth: '580px',
      margin: '0 auto 60px',
      lineHeight: 1.5,
    },
    /* ─── FEATURE GRID ─── */
    featureGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
      gap: '20px',
    },
    featureCard: {
      background: 'rgba(255,255,255,0.04)',
      border: '1px solid rgba(255,255,255,0.08)',
      borderRadius: '20px',
      padding: '40px 32px',
      transition: 'all 0.3s cubic-bezier(.4,0,.2,1)',
      cursor: 'default',
      ...blur(12),
    },
    featureIcon: {
      width: '48px', height: '48px',
      borderRadius: '12px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '24px',
      marginBottom: '20px',
    },
    featureTitle: {
      fontSize: '20px',
      fontWeight: 600,
      marginBottom: '10px',
      letterSpacing: '-0.01em',
    },
    featureDesc: {
      fontSize: '15px',
      lineHeight: 1.55,
      color: 'rgba(245,245,247,0.55)',
    },
    /* ─── STATS ─── */
    statsGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
      gap: '20px',
      marginTop: '60px',
    },
    statCard: {
      textAlign: 'center' as const,
      padding: '32px 16px',
      background: 'rgba(255,255,255,0.03)',
      borderRadius: '16px',
      border: '1px solid rgba(255,255,255,0.06)',
    },
    statValue: {
      fontSize: 'clamp(32px, 5vw, 48px)',
      fontWeight: 700,
      letterSpacing: '-0.03em',
      background: 'linear-gradient(135deg, #0071e3, #00e096)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      backgroundClip: 'text',
    } as React.CSSProperties,
    statLabel: {
      fontSize: '14px',
      color: 'rgba(245,245,247,0.5)',
      marginTop: '8px',
    },
    /* ─── CTA SECTION ─── */
    ctaSection: {
      textAlign: 'center' as const,
      padding: '100px 24px 120px',
    },
    ctaTitle: {
      fontSize: 'clamp(28px, 5vw, 48px)',
      fontWeight: 700,
      letterSpacing: '-0.025em',
      lineHeight: 1.1,
      marginBottom: '24px',
    },
    /* ─── FOOTER ─── */
    footer: {
      borderTop: '1px solid rgba(255,255,255,0.06)',
      padding: '32px 24px',
      textAlign: 'center' as const,
      fontSize: '13px',
      color: 'rgba(245,245,247,0.35)',
    },
    /* ─── MOBILE MENU ─── */
    burger: {
      display: 'none',
      background: 'none', border: 'none',
      color: '#f5f5f7', fontSize: '24px', cursor: 'pointer',
      padding: '4px',
    } as React.CSSProperties,
    mobileOverlay: {
      position: 'fixed' as const,
      inset: 0, zIndex: 99,
      background: 'rgba(0,0,0,0.92)',
      ...blur(20),
      display: 'flex',
      flexDirection: 'column' as const,
      alignItems: 'center',
      justifyContent: 'center',
      gap: '24px',
    },
    mobileLink: {
      fontSize: '22px',
      fontWeight: 500,
      color: '#f5f5f7',
      cursor: 'pointer',
      background: 'none', border: 'none',
    },
  };

  const features = [
    { icon: '🛍', bg: 'rgba(255,59,48,0.12)', title: 'Services', desc: 'Virtual cards for subscriptions, SaaS, and online services with instant issuance and smart limits.' },
    { icon: '✈️', bg: 'rgba(0,113,227,0.12)', title: 'Travel', desc: 'Optimized for travel bookings. Multi-currency support with competitive exchange rates.' },
    { icon: '⚡', bg: 'rgba(0,224,150,0.12)', title: 'Arbitrage', desc: 'High-limit cards for media buying and traffic arbitrage. BIN rotation and auto-replenishment.' },
  ];

  const stats = [
    { value: '50K+', label: 'Cards Issued' },
    { value: '$2M+', label: 'Monthly Volume' },
    { value: '99.8%', label: 'Uptime' },
    { value: '< 3s', label: 'Card Issuance' },
  ];

  return (
    <div style={css.page}>
      {/* ─── HEADER ─── */}
      <header style={css.header}>
        <div style={css.headerInner}>
          <div style={css.logo}>XPLR</div>

          {/* Desktop nav */}
          <nav style={css.nav} className="landing-desktop-nav">
            <span style={css.navLink} onClick={() => document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })}>Features</span>
            <span style={css.navLink} onClick={() => document.getElementById('stats')?.scrollIntoView({ behavior: 'smooth' })}>Stats</span>
            <span style={css.navLink} onClick={() => document.getElementById('pricing')?.scrollIntoView({ behavior: 'smooth' })}>Pricing</span>
            {token ? (
              <button style={css.ctaSmall} onClick={() => navigate('/dashboard')}>Dashboard</button>
            ) : (
              <button style={css.ctaSmall} onClick={() => navigate('/login')}>Sign In</button>
            )}
          </nav>

          {/* Mobile burger */}
          <button
            style={css.burger}
            className="landing-burger"
            onClick={() => setMobileMenu(!mobileMenu)}
            aria-label="Menu"
          >
            {mobileMenu ? '✕' : '☰'}
          </button>
        </div>
      </header>

      {/* Mobile overlay */}
      {mobileMenu && (
        <div style={css.mobileOverlay} onClick={() => setMobileMenu(false)}>
          <button style={css.mobileLink} onClick={() => { setMobileMenu(false); document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' }); }}>Features</button>
          <button style={css.mobileLink} onClick={() => { setMobileMenu(false); document.getElementById('stats')?.scrollIntoView({ behavior: 'smooth' }); }}>Stats</button>
          <button style={css.mobileLink} onClick={() => { setMobileMenu(false); document.getElementById('pricing')?.scrollIntoView({ behavior: 'smooth' }); }}>Pricing</button>
          <button style={{ ...css.btnPrimary, marginTop: '16px' }} onClick={() => { setMobileMenu(false); navigate(token ? '/dashboard' : '/register'); }}>
            {token ? 'Dashboard' : 'Get Started'}
          </button>
        </div>
      )}

      {/* ─── HERO ─── */}
      <section style={css.hero}>
        <div style={css.heroGlow} />
        <div style={css.heroTag}>Virtual Cards Platform</div>
        <h1 style={css.heroTitle}>Cards without limits.</h1>
        <p style={css.heroSub}>
          Issue virtual Visa & Mastercard instantly. Built for arbitrage, travel,
          and online services with real-time controls.
        </p>
        <div style={css.heroCta}>
          <button
            style={css.btnPrimary}
            onClick={() => navigate(token ? '/dashboard' : '/register')}
            onMouseEnter={e => { (e.target as HTMLElement).style.background = '#0077ED'; }}
            onMouseLeave={e => { (e.target as HTMLElement).style.background = '#0071e3'; }}
          >
            {token ? 'Open Dashboard' : 'Get Started Free'}
          </button>
          <button
            style={css.btnSecondary}
            onClick={() => document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })}
            onMouseEnter={e => { (e.target as HTMLElement).style.borderColor = '#0071e3'; }}
            onMouseLeave={e => { (e.target as HTMLElement).style.borderColor = 'rgba(0,113,227,0.4)'; }}
          >
            Learn More
          </button>
        </div>
      </section>

      {/* ─── FEATURES ─── */}
      <section id="features" style={css.section}>
        <h2 style={css.sectionTitle}>Built for every use case.</h2>
        <p style={css.sectionSub}>
          Whether you manage ad spend, book flights, or pay for SaaS — we've got the right card for you.
        </p>
        <div style={css.featureGrid}>
          {features.map((f, i) => (
            <div
              key={i}
              style={css.featureCard}
              onMouseEnter={e => {
                (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.07)';
                (e.currentTarget as HTMLElement).style.borderColor = 'rgba(255,255,255,0.15)';
                (e.currentTarget as HTMLElement).style.transform = 'translateY(-4px)';
              }}
              onMouseLeave={e => {
                (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.04)';
                (e.currentTarget as HTMLElement).style.borderColor = 'rgba(255,255,255,0.08)';
                (e.currentTarget as HTMLElement).style.transform = 'translateY(0)';
              }}
            >
              <div style={{ ...css.featureIcon, background: f.bg }}>{f.icon}</div>
              <div style={css.featureTitle}>{f.title}</div>
              <div style={css.featureDesc}>{f.desc}</div>
            </div>
          ))}
        </div>
      </section>

      {/* ─── STATS ─── */}
      <section id="stats" style={{ ...css.section, paddingTop: '40px' }}>
        <h2 style={css.sectionTitle}>Trusted at scale.</h2>
        <p style={css.sectionSub}>
          Real numbers from our platform. Reliable infrastructure for professionals.
        </p>
        <div style={css.statsGrid}>
          {stats.map((s, i) => (
            <div key={i} style={css.statCard}>
              <div style={css.statValue}>{s.value}</div>
              <div style={css.statLabel}>{s.label}</div>
            </div>
          ))}
        </div>
      </section>

      {/* ─── PRICING TEASER ─── */}
      <section id="pricing" style={css.section}>
        <h2 style={css.sectionTitle}>Simple, transparent pricing.</h2>
        <p style={css.sectionSub}>
          No hidden fees. Pay only for what you use. Volume discounts start at $10K/month.
        </p>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
          gap: '20px', maxWidth: '840px', margin: '0 auto',
        }}>
          {[
            { name: 'Starter', price: 'Free', sub: 'Up to 5 cards', feats: ['5 virtual cards', 'Basic analytics', 'Email support'] },
            { name: 'Professional', price: '$49/mo', sub: 'Unlimited cards', feats: ['Unlimited cards', 'Advanced analytics', 'Priority support', 'Auto-replenish', 'Team management'], highlight: true },
          ].map((p, i) => (
            <div key={i} style={{
              padding: '40px 32px',
              borderRadius: '20px',
              border: p.highlight ? '1px solid rgba(0,113,227,0.5)' : '1px solid rgba(255,255,255,0.08)',
              background: p.highlight ? 'rgba(0,113,227,0.06)' : 'rgba(255,255,255,0.03)',
              ...blur(12),
            }}>
              <div style={{ fontSize: '14px', color: p.highlight ? '#0071e3' : 'rgba(245,245,247,0.5)', fontWeight: 600, marginBottom: '8px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{p.name}</div>
              <div style={{ fontSize: '36px', fontWeight: 700, letterSpacing: '-0.03em', marginBottom: '4px' }}>{p.price}</div>
              <div style={{ fontSize: '14px', color: 'rgba(245,245,247,0.45)', marginBottom: '24px' }}>{p.sub}</div>
              <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 28px' }}>
                {p.feats.map((f, j) => (
                  <li key={j} style={{ padding: '8px 0', fontSize: '15px', color: 'rgba(245,245,247,0.7)', borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                    <span style={{ color: '#00e096', marginRight: '10px' }}>&#10003;</span>{f}
                  </li>
                ))}
              </ul>
              <button
                style={{ ...css.btnPrimary, width: '100%', background: p.highlight ? '#0071e3' : 'rgba(255,255,255,0.08)', fontSize: '15px', padding: '14px 0' }}
                onClick={() => navigate(token ? '/dashboard' : '/register')}
              >
                {p.highlight ? 'Start Free Trial' : 'Get Started'}
              </button>
            </div>
          ))}
        </div>
      </section>

      {/* ─── FINAL CTA ─── */}
      <section style={css.ctaSection}>
        <h2 style={css.ctaTitle}>Ready to scale your payments?</h2>
        <p style={{ ...css.sectionSub, marginBottom: '40px' }}>
          Join thousands of professionals who trust XPLR for their card infrastructure.
        </p>
        <button
          style={{ ...css.btnPrimary, padding: '18px 44px', fontSize: '18px' }}
          onClick={() => navigate(token ? '/dashboard' : '/register')}
          onMouseEnter={e => { (e.target as HTMLElement).style.background = '#0077ED'; }}
          onMouseLeave={e => { (e.target as HTMLElement).style.background = '#0071e3'; }}
        >
          {token ? 'Go to Dashboard' : 'Create Free Account'}
        </button>
      </section>

      {/* ─── FOOTER ─── */}
      <footer style={css.footer}>
        &copy; {new Date().getFullYear()} XPLR. All rights reserved.
      </footer>

      {/* ─── RESPONSIVE CSS ─── */}
      <style>{`
        /* Mobile: hide desktop nav, show burger */
        @media (max-width: 768px) {
          .landing-desktop-nav { display: none !important; }
          .landing-burger { display: block !important; }
        }
        @media (min-width: 769px) {
          .landing-burger { display: none !important; }
        }
        /* Ensure minimum 16px on inputs (iOS zoom prevention) */
        input, select, textarea { font-size: 16px !important; }
        /* Smooth scroll */
        html { scroll-behavior: smooth; }
        /* Selection color */
        ::selection { background: rgba(0,113,227,0.3); }
        ::-moz-selection { background: rgba(0,113,227,0.3); }
      `}</style>
    </div>
  );
};

export default LandingPage;
