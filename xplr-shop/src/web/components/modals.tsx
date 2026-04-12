import { motion, AnimatePresence } from "motion/react";
import type { Product } from "../data/products";

interface WarningModalProps {
  open: boolean;
  onClose: () => void;
}

export function WarningModal({ open, onClose }: WarningModalProps) {
  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          onClick={onClose}
        >
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 10 }}
            transition={{ duration: 0.25, ease: "easeOut" }}
            onClick={(e) => e.stopPropagation()}
            className="relative z-10 w-full max-w-sm rounded-2xl bg-[#0F1629] border border-white/[0.08] p-6 sm:p-8"
          >
            {/* Warning icon */}
            <div className="flex items-center justify-center w-14 h-14 rounded-2xl bg-amber-500/10 mb-5 mx-auto">
              <svg
                width="28"
                height="28"
                viewBox="0 0 24 24"
                fill="none"
                className="text-amber-400"
              >
                <path
                  d="M12 9V13M12 17H12.01M10.29 3.86L1.82 18A2 2 0 003.54 21H20.46A2 2 0 0022.18 18L13.71 3.86A2 2 0 0010.29 3.86Z"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </div>

            <h3 className="text-lg font-bold text-white text-center mb-2">
              Требуется карта XPLR
            </h3>
            <p className="text-sm text-white/50 text-center leading-relaxed mb-6">
              Для покупки необходима активная карта XPLR. Пожалуйста, откройте и
              пополните карту для подписок.
            </p>

            <button
              onClick={onClose}
              className="w-full py-3 rounded-xl font-semibold text-sm transition-all duration-200 bg-gradient-to-r from-[#38BDF8] to-[#A78BFA] text-white hover:opacity-90 active:scale-[0.98]"
            >
              Открыть карту
            </button>

            <button
              onClick={onClose}
              className="w-full py-2.5 mt-2 text-sm text-white/40 hover:text-white/60 transition-colors"
            >
              Отмена
            </button>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}

interface ConfirmModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  product: Product | null;
  cardSystem: string;
  cardLastFour: string;
}

export function ConfirmModal({
  open,
  onClose,
  onConfirm,
  product,
  cardSystem,
  cardLastFour,
}: ConfirmModalProps) {
  if (!product) return null;

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          onClick={onClose}
        >
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 10 }}
            transition={{ duration: 0.25, ease: "easeOut" }}
            onClick={(e) => e.stopPropagation()}
            className="relative z-10 w-full max-w-sm rounded-2xl bg-[#0F1629] border border-white/[0.08] p-6 sm:p-8"
          >
            {/* Success icon */}
            <div className="flex items-center justify-center w-14 h-14 rounded-2xl bg-emerald-500/10 mb-5 mx-auto">
              <svg
                width="28"
                height="28"
                viewBox="0 0 24 24"
                fill="none"
                className="text-emerald-400"
              >
                <path
                  d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </div>

            <h3 className="text-lg font-bold text-white text-center mb-4">
              Подтверждение покупки
            </h3>

            {/* Product info */}
            <div className="rounded-xl bg-white/[0.04] border border-white/[0.05] p-4 mb-4 space-y-2.5">
              <div className="flex justify-between items-center">
                <span className="text-sm text-white/40">Товар</span>
                <span className="text-sm text-white font-medium">
                  {product.brand} — ${product.originalPrice}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-white/40">Карта</span>
                <span className="text-sm text-white font-medium">
                  {cardSystem} •••• {cardLastFour}
                </span>
              </div>
              <div className="border-t border-white/[0.05] pt-2.5 flex justify-between items-center">
                <span className="text-sm text-white/60 font-medium">Итого</span>
                <span className="text-lg font-bold text-gradient">
                  ${product.price.toFixed(2)}
                </span>
              </div>
            </div>

            <button
              onClick={onConfirm}
              className="w-full py-3 rounded-xl font-semibold text-sm transition-all duration-200 bg-gradient-to-r from-[#38BDF8] to-[#A78BFA] text-white hover:opacity-90 active:scale-[0.98]"
            >
              Подтвердить оплату
            </button>

            <button
              onClick={onClose}
              className="w-full py-2.5 mt-2 text-sm text-white/40 hover:text-white/60 transition-colors"
            >
              Отмена
            </button>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
