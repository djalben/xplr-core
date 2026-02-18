import { ReactNode } from 'react';
import { Sidebar } from './sidebar';
import { NeuralBackground } from './neural-background';

interface DashboardLayoutProps {
  children: ReactNode;
}

export const DashboardLayout = ({ children }: DashboardLayoutProps) => {
  return (
    <div className="min-h-[100dvh] bg-gradient-to-br from-[#0a0a0f] via-[#0f0f18] to-[#12121a]">
      <NeuralBackground />
      <Sidebar />
      {/* Desktop: sidebar offset. Mobile: header + bottom nav offset */}
      <main className="lg:ml-64 min-h-[100dvh] relative z-10 gpu-accelerated">
        <div className="p-4 pt-32 pb-28 lg:p-8 lg:pt-8 lg:pb-8">
          {children}
        </div>
      </main>
    </div>
  );
};
