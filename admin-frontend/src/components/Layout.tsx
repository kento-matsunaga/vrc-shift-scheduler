import { Outlet, Link, useLocation } from 'react-router-dom';

export default function Layout() {
  const location = useLocation();

  const linkClass = (path: string) =>
    `px-4 py-2 rounded-md text-sm font-medium transition-colors ${
      location.pathname === path
        ? 'bg-indigo-100 text-indigo-700'
        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
    }`;

  return (
    <div className="min-h-screen bg-gray-100">
      {/* ヘッダー */}
      <header className="bg-indigo-900 text-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <h1 className="text-xl font-bold">VRC Shift Scheduler</h1>
              <span className="bg-indigo-700 px-2 py-1 rounded text-xs font-medium">
                管理コンソール
              </span>
            </div>
            <div className="text-sm text-indigo-200">
              Cloudflare Access で保護されています
            </div>
          </div>
        </div>
      </header>

      {/* ナビゲーション */}
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-4 py-3">
            <Link to="/" className={linkClass('/')}>
              ライセンスキー
            </Link>
            <Link to="/tenants" className={linkClass('/tenants')}>
              テナント
            </Link>
            <Link to="/announcements" className={linkClass('/announcements')}>
              お知らせ
            </Link>
            <Link to="/tutorials" className={linkClass('/tutorials')}>
              チュートリアル
            </Link>
            <Link to="/audit-logs" className={linkClass('/audit-logs')}>
              監査ログ
            </Link>
          </div>
        </div>
      </nav>

      {/* メインコンテンツ */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>

      {/* フッター */}
      <footer className="bg-white border-t mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-500">
            VRC Shift Scheduler 管理コンソール - 運営専用
          </p>
        </div>
      </footer>
    </div>
  );
}
