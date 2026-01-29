import type { ReactNode } from 'react';
import { SettingsSidebar } from './SettingsSidebar';
import { SettingsMobileMenu } from './SettingsMobileMenu';

interface SettingsLayoutProps {
  children: ReactNode;
  title: string;
}

export function SettingsLayout({ children, title }: SettingsLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* モバイルヘッダー */}
      <div className="md:hidden">
        <SettingsMobileMenu currentTitle={title} />
      </div>

      <div className="flex">
        {/* デスクトップサイドバー */}
        <aside className="hidden md:block w-60 min-h-screen bg-white border-r">
          <SettingsSidebar />
        </aside>

        {/* コンテンツエリア */}
        <main className="flex-1 p-6">
          <h1 className="text-2xl font-bold mb-6">{title}</h1>
          {children}
        </main>
      </div>
    </div>
  );
}
