import { AnimatePresence } from "motion/react";
import { ShopProvider, useShop } from "../store/shop-context";
import { ShopHeader } from "../components/shop-header";
import { HubView } from "../components/hub-view";
import { EsimView } from "../components/esim-view";
import { DigitalView } from "../components/digital-view";

function ShopContent() {
  const { view } = useShop();

  return (
    <div className="min-h-dvh bg-[#060B18] flex flex-col">
      <ShopHeader />
      <main className="flex-1 overflow-y-auto">
        <AnimatePresence mode="wait">
          {view === "hub" && <HubView key="hub" />}
          {view === "esim" && <EsimView key="esim" />}
          {view === "digital" && <DigitalView key="digital" />}
        </AnimatePresence>
      </main>
    </div>
  );
}

export default function Index() {
  return (
    <ShopProvider>
      <ShopContent />
    </ShopProvider>
  );
}
