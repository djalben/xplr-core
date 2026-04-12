interface IconProps {
  className?: string;
}

export function SteamIcon({ className = "w-8 h-8" }: IconProps) {
  return (
    <svg viewBox="0 0 256 256" className={className} fill="none">
      <defs>
        <linearGradient id="steam-grad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#111D2E" />
          <stop offset="21%" stopColor="#051839" />
          <stop offset="41%" stopColor="#0A1B48" />
          <stop offset="58%" stopColor="#132E62" />
          <stop offset="74%" stopColor="#144B7E" />
          <stop offset="87%" stopColor="#136497" />
          <stop offset="100%" stopColor="#1387B8" />
        </linearGradient>
      </defs>
      <circle cx="128" cy="128" r="128" fill="url(#steam-grad)" />
      <path
        d="M128.079 57.604c28.299 0 51.364 22.544 52.253 50.664l.033 1.58-30.474 17.747a37.406 37.406 0 0 0-14.6-2.958c-1.238 0-2.46.063-3.667.182l-20.378-29.473-.002-.514c0-20.672 16.78-37.228 37.43-37.228h-.595Zm22.263 56.444 16.68-9.72c-2.74-14.14-15.21-24.858-30.148-24.858-16.899 0-30.6 13.632-30.728 30.474l20.607 14.73a26.468 26.468 0 0 1 10.198-2.041c4.84 0 9.335 1.286 13.391 4.415ZM91.51 151.773l-14.682-6.065c2.657 7.853 8.788 14.306 16.573 17.498 16.346 6.695 35.093-1.171 41.811-17.546a30.49 30.49 0 0 0 1.758-15.395 30.567 30.567 0 0 0-6.298-14.44l-16.396 9.596a19.478 19.478 0 0 1-7.25 21.516 19.39 19.39 0 0 1-15.516 4.836Z"
        fill="white"
      />
    </svg>
  );
}

export function PlayStationIcon({ className = "w-8 h-8" }: IconProps) {
  return (
    <svg viewBox="0 0 256 256" className={className}>
      <rect width="256" height="256" rx="40" fill="#003791" />
      <path
        d="M99.8 166.4V85.6l36.4-12v119.2l-36.4 12.8V166.4Zm77.6-22.8c-6-3.6-13.6-5.2-24-3.6l-17.2 4.8v29.6l12-3.6c6.4-2 12-1.2 12 4.8s-5.6 10.8-12 12.8l-12 4V209l11.6-4c13.6-4.8 24.8-11.6 30.4-20.8 6-9.6 5.6-22-.8-30l-.8.4h.8ZM72.6 194.8l36.4-12.8v-20.4l-22 7.6c-6.4 2-12 1.2-12-4.8s5.6-10.8 12-12.8l22-7.6v-18l-12 4c-13.6 4.8-24.8 11.6-30.4 20.8-5.6 10-5.2 22.4 1.2 30.8l4.8 2.8v10.4Z"
        fill="white"
      />
    </svg>
  );
}

export function XboxIcon({ className = "w-8 h-8" }: IconProps) {
  return (
    <svg viewBox="0 0 256 256" className={className}>
      <rect width="256" height="256" rx="40" fill="#107C10" />
      <path
        d="M128 44c-8.4 0-16.4 1.2-24 3.4 8-2 19.2 4.4 24 8.4 4.8-4 16-10.4 24-8.4-7.6-2.2-15.6-3.4-24-3.4Zm-50.4 20.8C65.2 76.4 56 94.8 52.4 114c-3.2 17.2-1.6 34.4 4 48.8 12-20 28.8-40 47.2-56.8-8.4-12-18.4-28-26-41.2h.4Zm100.8 0c-7.6 13.2-17.6 29.2-26 41.2 18.4 16.8 35.2 36.8 47.2 56.8 5.6-14.4 7.2-31.6 4-48.8-3.6-19.2-12.8-37.6-25.2-49.2Zm-50.4 52.8c-20.4 18.8-38.8 41.2-50 62.4 14.4 14.8 34 24.4 56 25.6V208c-2 0-3.6-.4-5.6-.4-17.6-1.6-34-10.4-44.8-24.4 12.8-20.4 28.4-41.6 44.4-56.8v-8.8Zm0 0c16 15.2 31.6 36.4 44.4 56.8-10.8 14-27.2 22.8-44.8 24.4-2 0-3.6.4-5.6.4v-2.4c22-1.2 41.6-10.8 56-25.6-11.2-21.2-29.6-43.6-50-53.6Z"
        fill="white"
      />
    </svg>
  );
}

export function NintendoIcon({ className = "w-8 h-8" }: IconProps) {
  return (
    <svg viewBox="0 0 256 256" className={className}>
      <rect width="256" height="256" rx="40" fill="#E60012" />
      <path
        d="M92 60h-16c-17.6 0-32 14.4-32 32v72c0 17.6 14.4 32 32 32h16c4.4 0 8-3.6 8-8V68c0-4.4-3.6-8-8-8Zm-16 112c-8.8 0-16-7.2-16-16s7.2-16 16-16 16 7.2 16 16-7.2 16-16 16Zm104-112h-16c-4.4 0-8 3.6-8 8v120c0 4.4 3.6 8 8 8h16c17.6 0 32-14.4 32-32V92c0-17.6-14.4-32-32-32Z"
        fill="white"
      />
    </svg>
  );
}

export function SpotifyIcon({ className = "w-8 h-8" }: IconProps) {
  return (
    <svg viewBox="0 0 256 256" className={className}>
      <circle cx="128" cy="128" r="128" fill="#1DB954" />
      <path
        d="M186.8 118c-23.2-13.8-61.4-15-83.6-8.3-3.6 1.1-7.3-.9-8.4-4.5s.9-7.3 4.5-8.4c25.5-7.7 67.9-6.2 94.7 9.6 3.2 1.9 4.2 6 2.3 9.2-1.9 3.1-6 4.2-9.2 2.3l-.3.1Zm-2-21.7c-27.4-16.3-68.5-17.7-98.6-9.8-4.2 1.1-8.5-1.3-9.7-5.5-1.1-4.2 1.3-8.5 5.5-9.7 34.4-9 86.3-7.3 117.6 11.4 3.7 2.2 5 7.1 2.8 10.8-2.2 3.7-7 5-10.8 2.8h1.2Zm-10.7 41.2c-19-11.3-50.4-12.3-68.5-6.8-2.9.9-6-.7-6.9-3.7-.9-2.9.7-6 3.7-6.9 20.8-6.3 55.5-5.1 77.3 7.9 2.6 1.6 3.5 5 1.9 7.6-1.6 2.6-5 3.4-7.5 1.9Z"
        fill="white"
      />
    </svg>
  );
}

export function NetflixIcon({ className = "w-8 h-8" }: IconProps) {
  return (
    <svg viewBox="0 0 256 256" className={className}>
      <rect width="256" height="256" rx="40" fill="#E50914" />
      <path
        d="M88 60v136l36-60V60H88Zm44 0v136l36-60V60h-36Zm0 76v60l36-60v-60l-36 60Z"
        fill="white"
      />
    </svg>
  );
}

export const brandIcons: Record<string, React.FC<IconProps>> = {
  steam: SteamIcon,
  playstation: PlayStationIcon,
  xbox: XboxIcon,
  nintendo: NintendoIcon,
  spotify: SpotifyIcon,
  netflix: NetflixIcon,
};
