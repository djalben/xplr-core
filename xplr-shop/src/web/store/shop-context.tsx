import { createContext, useContext, useState, type ReactNode } from "react";

type View = "hub" | "esim" | "digital";

interface CardInfo {
  system: string;
  lastFour: string;
}

interface ShopContextType {
  view: View;
  setView: (v: View) => void;
  hasCard: boolean;
  setHasCard: (v: boolean) => void;
  cardInfo: CardInfo | null;
  setCardInfo: (v: CardInfo | null) => void;
}

const ShopContext = createContext<ShopContextType | null>(null);

export function ShopProvider({ children }: { children: ReactNode }) {
  const [view, setView] = useState<View>("hub");
  const [hasCard, setHasCard] = useState(false);
  const [cardInfo, setCardInfo] = useState<CardInfo | null>(null);

  return (
    <ShopContext.Provider
      value={{ view, setView, hasCard, setHasCard, cardInfo, setCardInfo }}
    >
      {children}
    </ShopContext.Provider>
  );
}

export function useShop() {
  const ctx = useContext(ShopContext);
  if (!ctx) throw new Error("useShop must be used within ShopProvider");
  return ctx;
}
