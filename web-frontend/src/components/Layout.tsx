import { useState } from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const adminRole = localStorage.getItem('admin_role') || '';
  const [showGroupSubmenu, setShowGroupSubmenu] = useState(false);

  const handleLogout = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('admin_id');
    localStorage.removeItem('tenant_id');
    localStorage.removeItem('admin_role');
    localStorage.removeItem('member_id');
    localStorage.removeItem('member_name');
    navigate('/admin/login');
  };

  // サイドバーリンクのスタイル
  const linkClass = (path: string) =>
    `flex items-center px-4 py-2.5 text-sm font-medium rounded-lg transition-colors ${
      location.pathname.startsWith(path)
        ? 'bg-indigo-100 text-indigo-700'
        : 'text-gray-700 hover:bg-gray-100'
    }`;

  // グループメニューがアクティブかどうか
  const isGroupActive = location.pathname.startsWith('/groups') || location.pathname.startsWith('/role-groups');

  return (
    <div className="min-h-screen bg-gray-100 flex">
      {/* サイドバー */}
      <aside className="w-56 bg-white shadow-md flex flex-col fixed h-full">
        {/* ロゴ */}
        <div className="p-4 border-b">
          <h1 className="text-lg font-bold text-indigo-900">VRC Shift Scheduler</h1>
          <span className="inline-block mt-1 bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded text-xs font-medium">
            {adminRole === 'owner' ? 'オーナー' : 'マネージャー'}
          </span>
        </div>

        {/* ナビゲーション */}
        <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
          <Link to="/events" className={linkClass('/events')}>
            イベント
          </Link>
          <Link to="/members" className={linkClass('/members')}>
            メンバー
          </Link>
          <Link to="/roles" className={linkClass('/roles')}>
            ロール
          </Link>

          {/* グループサブメニュー */}
          <div>
            <button
              onClick={() => setShowGroupSubmenu(!showGroupSubmenu)}
              className={`w-full flex items-center justify-between px-4 py-2.5 text-sm font-medium rounded-lg transition-colors ${
                isGroupActive
                  ? 'bg-indigo-100 text-indigo-700'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}
            >
              <span>グループ</span>
              <svg
                className={`w-4 h-4 transition-transform ${showGroupSubmenu ? 'rotate-180' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            {showGroupSubmenu && (
              <div className="ml-4 mt-1 space-y-1">
                <Link
                  to="/role-groups"
                  className={`block px-4 py-2 text-sm rounded-lg transition-colors ${
                    location.pathname.startsWith('/role-groups')
                      ? 'bg-indigo-50 text-indigo-700'
                      : 'text-gray-600 hover:bg-gray-50'
                  }`}
                >
                  ロールグループ
                </Link>
                <Link
                  to="/groups"
                  className={`block px-4 py-2 text-sm rounded-lg transition-colors ${
                    location.pathname.startsWith('/groups')
                      ? 'bg-indigo-50 text-indigo-700'
                      : 'text-gray-600 hover:bg-gray-50'
                  }`}
                >
                  メンバーグループ
                </Link>
              </div>
            )}
          </div>

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
        </nav>

        {/* ログアウト */}
        <div className="p-3 border-t">
          <button
            onClick={handleLogout}
            className="w-full px-4 py-2 text-sm font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors text-left"
          >
            ログアウト
          </button>
        </div>
      </aside>

      {/* メインコンテンツ */}
      <div className="flex-1 ml-56">
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
