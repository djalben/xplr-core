import { ReactNode } from 'react';

export const AdminLayout = ({ children }: { children: ReactNode }) => {
  return (
    <div className="min-h-[100dvh] bg-transparent relative z-2">
      <main className="min-h-[100dvh] relative z-10 overflow-x-hidden">
        <div className="p-4 pb-10 lg:p-10 lg:pb-10">
          {children}
        </div>
      </main>
    </div>
  );
};

