import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

interface ClockProps {
  timezone: string;
  label: string;
}

const AnalogClock = ({ timezone, label }: ClockProps) => {
  const [time, setTime] = useState(new Date());

  useEffect(() => {
    const interval = setInterval(() => setTime(new Date()), 1000);
    return () => clearInterval(interval);
  }, []);

  const localTime = new Date(time.toLocaleString('en-US', { timeZone: timezone }));
  const hours = localTime.getHours() % 12;
  const minutes = localTime.getMinutes();
  const seconds = localTime.getSeconds();

  const hourDeg = hours * 30 + minutes * 0.5;
  const minuteDeg = minutes * 6 + seconds * 0.1;
  const secondDeg = seconds * 6;

  const timeStr = localTime.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', hour12: false });

  return (
    <div className="flex flex-col items-center gap-2">
      <div className="relative w-16 h-16 md:w-20 md:h-20">
        {/* Clock face */}
        <svg viewBox="0 0 100 100" className="w-full h-full">
          {/* Outer ring */}
          <circle cx="50" cy="50" r="48" fill="none" stroke="rgba(255,255,255,0.08)" strokeWidth="1" />
          <circle cx="50" cy="50" r="46" fill="rgba(255,255,255,0.02)" stroke="none" />

          {/* Hour markers */}
          {Array.from({ length: 12 }).map((_, i) => {
            const angle = (i * 30 - 90) * (Math.PI / 180);
            const x1 = 50 + 40 * Math.cos(angle);
            const y1 = 50 + 40 * Math.sin(angle);
            const x2 = 50 + (i % 3 === 0 ? 34 : 37) * Math.cos(angle);
            const y2 = 50 + (i % 3 === 0 ? 34 : 37) * Math.sin(angle);
            return (
              <line
                key={i}
                x1={x1} y1={y1} x2={x2} y2={y2}
                stroke={i % 3 === 0 ? 'rgba(255,255,255,0.35)' : 'rgba(255,255,255,0.15)'}
                strokeWidth={i % 3 === 0 ? '1.5' : '0.8'}
                strokeLinecap="round"
              />
            );
          })}

          {/* Hour hand */}
          <line
            x1="50" y1="50"
            x2={50 + 22 * Math.cos((hourDeg - 90) * Math.PI / 180)}
            y2={50 + 22 * Math.sin((hourDeg - 90) * Math.PI / 180)}
            stroke="rgba(255,255,255,0.7)"
            strokeWidth="2.5"
            strokeLinecap="round"
          />

          {/* Minute hand */}
          <line
            x1="50" y1="50"
            x2={50 + 32 * Math.cos((minuteDeg - 90) * Math.PI / 180)}
            y2={50 + 32 * Math.sin((minuteDeg - 90) * Math.PI / 180)}
            stroke="rgba(255,255,255,0.5)"
            strokeWidth="1.5"
            strokeLinecap="round"
          />

          {/* Second hand */}
          <line
            x1="50" y1="50"
            x2={50 + 36 * Math.cos((secondDeg - 90) * Math.PI / 180)}
            y2={50 + 36 * Math.sin((secondDeg - 90) * Math.PI / 180)}
            stroke="rgba(96,165,250,0.6)"
            strokeWidth="0.7"
            strokeLinecap="round"
          />

          {/* Center dot */}
          <circle cx="50" cy="50" r="2" fill="rgba(255,255,255,0.5)" />
        </svg>
      </div>
      <div className="text-center">
        <p className="text-white font-mono text-sm tracking-wider">{timeStr}</p>
        <p className="text-slate-500 text-[10px] uppercase tracking-widest mt-0.5">{label}</p>
      </div>
    </div>
  );
};

export const WorldClocks = () => {
  const { t } = useTranslation();
  return (
    <div className="flex items-center gap-6 md:gap-8">
      <AnalogClock timezone="Europe/Moscow" label={t('clocks.moscow')} />
      <AnalogClock timezone="Europe/London" label={t('clocks.london')} />
      <AnalogClock timezone="America/New_York" label={t('clocks.newYork')} />
    </div>
  );
};
