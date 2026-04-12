import { motion } from "motion/react";
import { useShop } from "../store/shop-context";
import { countries, getFlagUrl } from "../data/countries";

export function EsimView() {
  const { setView } = useShop();

  return (
    <motion.div
      initial={{ opacity: 0, x: 40 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -40 }}
      transition={{ duration: 0.35, ease: "easeOut" }}
      className="p-4 sm:p-6 max-w-2xl mx-auto"
    >
      {/* Header */}
      <div className="flex items-center gap-4 mb-6 sm:mb-8">
        <button
          onClick={() => setView("hub")}
          className="flex items-center justify-center w-10 h-10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] transition-colors"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 20 20"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M12.5 15L7.5 10L12.5 5"
              stroke="white"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </button>
        <div>
          <h1 className="text-xl sm:text-2xl font-bold text-white">
            eSIM и Сим-карты
          </h1>
          <p className="text-sm text-white/40 mt-0.5">Выберите страну</p>
        </div>
      </div>

      {/* Country List */}
      <div className="rounded-2xl bg-white/[0.02] border border-white/[0.05] overflow-hidden">
        {countries.map((country, i) => (
          <motion.button
            key={country.code}
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: i * 0.03 }}
            className={`w-full flex items-center gap-4 px-4 sm:px-5 py-3.5 sm:py-4 hover:bg-white/[0.04] transition-colors text-left ${
              i !== countries.length - 1
                ? "border-b border-white/[0.05]"
                : ""
            }`}
          >
            <img
              src={getFlagUrl(country.code)}
              alt={country.name}
              className="w-8 h-6 sm:w-10 sm:h-7 rounded-[3px] object-cover shadow-sm flex-shrink-0"
              loading="lazy"
            />
            <span className="text-sm sm:text-[15px] font-medium text-white flex-1 truncate">
              {country.name}
            </span>
            {country.planCount !== null ? (
              <span className="text-xs sm:text-sm text-white/30 flex-shrink-0 tabular-nums">
                {country.planCount}{" "}
                {country.planCount === 3 ? "тарифа" : "тарифа"}
              </span>
            ) : null}
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              className="text-white/20 flex-shrink-0"
            >
              <path
                d="M6 12L10 8L6 4"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </motion.button>
        ))}
      </div>
    </motion.div>
  );
}
