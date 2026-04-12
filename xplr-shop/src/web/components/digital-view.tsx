import { useState } from "react";
import { motion } from "motion/react";
import { useShop } from "../store/shop-context";
import { products, type Product } from "../data/products";
import { WarningModal } from "./modals";
import { ConfirmModal } from "./modals";
import { brandIcons } from "./brand-icons";

export function DigitalView() {
  const { setView, hasCard, cardInfo } = useShop();
  const [warningOpen, setWarningOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);

  function handleBuy(product: Product) {
    setSelectedProduct(product);
    if (!hasCard) {
      setWarningOpen(true);
    } else {
      setConfirmOpen(true);
    }
  }

  function handleConfirmPurchase() {
    setConfirmOpen(false);
    setSelectedProduct(null);
    // Backend integration will go here
  }

  // Group products by brand
  const grouped = products.reduce<Record<string, Product[]>>((acc, p) => {
    if (!acc[p.brand]) acc[p.brand] = [];
    acc[p.brand].push(p);
    return acc;
  }, {});

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
            Цифровые товары
          </h1>
          <p className="text-sm text-white/40 mt-0.5">
            Подарочные карты и подписки
          </p>
        </div>
      </div>

      {/* Product groups */}
      <div className="space-y-4">
        {Object.entries(grouped).map(([brand, items], groupIdx) => (
          <motion.div
            key={brand}
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.35, delay: groupIdx * 0.06 }}
            className="rounded-2xl bg-white/[0.02] border border-white/[0.05] overflow-hidden"
          >
            {items.map((product, i) => (
              <div
                key={product.id}
                className={`flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-3.5 sm:py-4 ${
                  i !== items.length - 1
                    ? "border-b border-white/[0.05]"
                    : ""
                }`}
              >
                {/* Brand Logo */}
                {(() => {
                  const Icon = brandIcons[product.logoSlug];
                  return Icon ? (
                    <Icon className="w-8 h-8 sm:w-9 sm:h-9 flex-shrink-0" />
                  ) : (
                    <div className="w-8 h-8 sm:w-9 sm:h-9 rounded-lg bg-white/10 flex-shrink-0" />
                  );
                })()}

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <div className="text-sm sm:text-[15px] font-medium text-white truncate">
                    {product.name}
                  </div>
                  <div className="text-xs text-white/35 mt-0.5 truncate">
                    {product.description}
                  </div>
                </div>

                {/* Price + Buy */}
                <div className="flex items-center gap-3 flex-shrink-0">
                  <span className="text-sm sm:text-[15px] font-bold text-gradient tabular-nums">
                    ${product.price.toFixed(1)}
                  </span>
                  <button
                    onClick={() => handleBuy(product)}
                    className="px-3.5 sm:px-4 py-1.5 sm:py-2 rounded-lg text-xs sm:text-sm font-semibold bg-white/[0.06] hover:bg-white/[0.1] text-white transition-all duration-200 active:scale-[0.96]"
                  >
                    Купить
                  </button>
                </div>
              </div>
            ))}
          </motion.div>
        ))}
      </div>

      {/* Modals */}
      <WarningModal open={warningOpen} onClose={() => setWarningOpen(false)} />
      <ConfirmModal
        open={confirmOpen}
        onClose={() => setConfirmOpen(false)}
        onConfirm={handleConfirmPurchase}
        product={selectedProduct}
        cardSystem={cardInfo?.system ?? "Visa"}
        cardLastFour={cardInfo?.lastFour ?? "4242"}
      />
    </motion.div>
  );
}
