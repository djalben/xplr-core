import { useShop } from "../store/shop-context";

export function ShopHeader() {
  const { hasCard, setHasCard, setCardInfo, view } = useShop();

  function toggleCard() {
    if (hasCard) {
      setHasCard(false);
      setCardInfo(null);
    } else {
      setHasCard(true);
      setCardInfo({ system: "Visa", lastFour: "4242" });
    }
  }

  return (
    <header className="flex items-center justify-between px-4 sm:px-6 py-4 border-b border-white/[0.05]">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[#38BDF8] to-[#A78BFA] flex items-center justify-center">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            className="text-white"
          >
            <path
              d="M6 2L3 6V20C3 20.5304 3.21071 21.0391 3.58579 21.4142C3.96086 21.7893 4.46957 22 5 22H19C19.5304 22 20.0391 21.7893 20.4142 21.4142C20.7893 21.0391 21 20.5304 21 20V6L18 2H6Z"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M3 6H21"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M16 10C16 11.0609 15.5786 12.0783 14.8284 12.8284C14.0783 13.5786 13.0609 14 12 14C10.9391 14 9.92172 13.5786 9.17157 12.8284C8.42143 12.0783 8 11.0609 8 10"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
        <span className="text-base sm:text-lg font-bold text-white tracking-tight">
          Магазин
        </span>
        {view !== "hub" && (
          <span className="hidden sm:inline text-xs text-white/25 font-medium uppercase tracking-wider ml-1">
            XPLR
          </span>
        )}
      </div>

      {/* Card toggle (demo) */}
      <button
        onClick={toggleCard}
        className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 ${
          hasCard
            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
            : "bg-white/[0.04] text-white/40 border border-white/[0.06] hover:bg-white/[0.06]"
        }`}
      >
        <div
          className={`w-1.5 h-1.5 rounded-full ${
            hasCard ? "bg-emerald-400" : "bg-white/30"
          }`}
        />
        {hasCard ? "Карта активна" : "Нет карты"}
      </button>
    </header>
  );
}
