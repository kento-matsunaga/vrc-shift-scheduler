import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const adminRole = localStorage.getItem('admin_role') || '';

  const handleLogout = () => {
    // JWT認証関連のデータをクリア
    localStorage.removeItem('auth_token');
    localStorage.removeItem('admin_id');
    localStorage.removeItem('tenant_id');
    localStorage.removeItem('admin_role');
    // 旧形式のデータもクリア（念のため）
    localStorage.removeItem('member_id');
    localStorage.removeItem('member_name');
    navigate('/admin/login');
  };

  // ナビゲーションリンクのスタイル
  const linkClass = (path: string) =>
    `px-4 py-2 rounded-md text-sm font-medium transition-colors ${
      location.pathname.startsWith(path)
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
                {adminRole === 'owner' ? 'オーナー' : 'マネージャー'}
              </span>
            </div>
            <button
              onClick={handleLogout}
              className="px-4 py-2 text-sm font-medium text-indigo-200 hover:text-white transition-colors"
            >
              ログアウト
            </button>
          </div>
        </div>
      </header>

      {/* ナビゲーション */}
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-4 py-3 overflow-x-auto">
            <Link to="/events" className={linkClass('/events')}>
              イベント
            </Link>
            <Link to="/members" className={linkClass('/members')}>
              メンバー
            </Link>
            <Link to="/roles" className={linkClass('/roles')}>
              ロール
            </Link>
            <Link to="/groups" className={linkClass('/groups')}>
              グループ
            </Link>
            <Link to="/attendance" className={linkClass('/attendance')}>
              出欠確認
            </Link>
            <Link to="/schedules" className={linkClass('/schedules')}>
              日程調整
            </Link>
            {(adminRole === 'admin' || adminRole === 'owner') && (
              <Link to="/admin/invite" className={linkClass('/admin/invite')}>
                管理者招待
              </Link>
            )}
            <Link to="/settings" className={linkClass('/settings')}>
              設定
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
            VRC Shift Scheduler
          </p>
        </div>
      </footer>
    </div>
  );
}

