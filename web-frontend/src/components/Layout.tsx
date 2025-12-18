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
    `px-3 py-2 rounded-md text-sm font-medium transition-colors ${
      location.pathname.startsWith(path)
        ? 'bg-blue-100 text-blue-700'
        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
    }`;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-8">
              <h1 className="text-2xl font-bold text-gray-900">VRC Shift Scheduler</h1>
              <nav className="hidden md:flex space-x-2">
                <Link to="/events" className={linkClass('/events')}>
                  イベント
                </Link>
                <Link to="/members" className={linkClass('/members')}>
                  メンバー
                </Link>
                <Link to="/roles" className={linkClass('/roles')}>
                  ロール
                </Link>
                <Link to="/attendance" className={linkClass('/attendance')}>
                  出欠確認
                </Link>
                <Link to="/schedules" className={linkClass('/schedules')}>
                  日程調整
                </Link>
                <Link to="/my-shifts" className={linkClass('/my-shifts')}>
                  自分のシフト
                </Link>
                {(adminRole === 'admin' || adminRole === 'owner') && (
                  <Link to="/admin/invite" className={linkClass('/admin/invite')}>
                    管理者招待
                  </Link>
                )}
              </nav>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">
                {adminRole === 'owner' ? '👑 オーナー' : '👤 マネージャー'}
              </span>
              <button onClick={handleLogout} className="btn-secondary text-sm">
                ログアウト
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* メインコンテンツ */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>

      {/* フッター */}
      <footer className="bg-white border-t mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-500">
            ⚠️ これは α 版のテストです。データは予告なく消える可能性があります。
          </p>
        </div>
      </footer>
    </div>
  );
}

