import { motion } from "motion/react";
import { useShop } from "../store/shop-context";

export function HubView() {
  const { setView } = useShop();

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.4 }}
      className="flex flex-col lg:flex-row gap-4 sm:gap-5 p-4 sm:p-6 min-h-[calc(100dvh-80px)]"
    >
      {/* eSIM Card */}
      <motion.button
        onClick={() => setView("esim")}
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.1 }}
        whileHover={{ scale: 1.01 }}
        whileTap={{ scale: 0.99 }}
        className="relative flex-1 min-h-[240px] sm:min-h-[280px] rounded-2xl overflow-hidden group cursor-pointer"
      >
        <img
          src="/esim-card.png"
          alt="eSIM"
          className="absolute inset-0 w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-[#060B18] via-[#060B18]/60 to-transparent" />
        <div className="relative z-10 flex flex-col justify-end h-full p-6 sm:p-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-2 h-2 rounded-full bg-[#38BDF8] animate-pulse" />
            <span className="text-xs sm:text-sm font-medium tracking-widest uppercase text-[#38BDF8]/80">
              20 стран
            </span>
          </div>
          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-extrabold text-white text-left leading-tight">
            eSIM и Сим-карты
          </h2>
          <p className="text-sm sm:text-base text-white/50 mt-2 text-left max-w-md">
            Мобильный интернет в любой точке мира. Мгновенная активация.
          </p>
        </div>
      </motion.button>

      {/* Digital Products Card */}
      <motion.button
        onClick={() => setView("digital")}
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.25 }}
        whileHover={{ scale: 1.01 }}
        whileTap={{ scale: 0.99 }}
        className="relative flex-1 min-h-[240px] sm:min-h-[280px] rounded-2xl overflow-hidden group cursor-pointer"
      >
        <img
          src="/digital-products.png"
          alt="Digital Products"
          className="absolute inset-0 w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-[#060B18] via-[#060B18]/60 to-transparent" />
        <div className="relative z-10 flex flex-col justify-end h-full p-6 sm:p-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-2 h-2 rounded-full bg-[#A78BFA] animate-pulse" />
            <span className="text-xs sm:text-sm font-medium tracking-widest uppercase text-[#A78BFA]/80">
              6 брендов
            </span>
          </div>
          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-extrabold text-white text-left leading-tight">
            Цифровые товары
          </h2>
          <p className="text-sm sm:text-base text-white/50 mt-2 text-left max-w-md">
            Подарочные карты, подписки и пополнения по лучшим ценам.
          </p>
        </div>
      </motion.button>
    </motion.div>
  );
}
