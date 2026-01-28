import { useState, useEffect } from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { AnnouncementBell } from './AnnouncementBell';
import { TutorialButton } from './TutorialButton';
import { useDocumentTitle, getTitleFromPath } from '../hooks/useDocumentTitle';

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const adminRole = localStorage.getItem('admin_role') || '';
  const [showGroupSubmenu, setShowGroupSubmenu] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // ページタイトルを設定
  const pageTitle = getTitleFromPath(location.pathname);
  useDocumentTitle(pageTitle || 'VRCシフト管理');

  // サイドバー開閉時のbodyスクロール制御とESCキー対応
  useEffect(() => {
    if (sidebarOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && sidebarOpen) {
        setSidebarOpen(false);
      }
    };
    document.addEventListener('keydown', handleEscape);

    return () => {
      document.body.style.overflow = 'unset';
      document.removeEventListener('keydown', handleEscape);
    };
  }, [sidebarOpen]);

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
        ? 'bg-accent/10 text-accent-dark border-l-4 border-accent'
        : 'text-gray-700 hover:bg-gray-100'
    }`;

  // グループメニューがアクティブかどうか
  const isGroupActive = location.pathname.startsWith('/groups') || location.pathname.startsWith('/role-groups');

  // ナビゲーションリンクをクリックしたらサイドバーを閉じる（モバイル）
  const handleNavClick = () => {
    setSidebarOpen(false);
  };

  return (
    <div className="min-h-screen bg-gray-100 flex">
      {/* モバイル用オーバーレイ */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* サイドバー */}
      <aside
        className={`
          fixed inset-y-0 left-0 z-50 w-56 bg-white shadow-md flex flex-col
          transform transition-transform duration-300 ease-in-out
          ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
          md:translate-x-0 md:static md:z-auto
        `}
      >
        {/* ロゴ */}
        <div className="p-4 border-b border-gray-200 bg-vrc-dark flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-white">VRC Shift Scheduler</h1>
            <span className="inline-block mt-1 bg-accent text-white px-2 py-0.5 rounded text-xs font-medium">
              {adminRole === 'owner' ? 'オーナー' : 'マネージャー'}
            </span>
          </div>
          {/* モバイル用閉じるボタン */}
          <button
            onClick={() => setSidebarOpen(false)}
            className="md:hidden p-1 text-white hover:bg-white/10 rounded"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* ナビゲーション */}
        <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
          <Link to="/events" className={linkClass('/events')} onClick={handleNavClick}>
            イベント
          </Link>
          <Link to="/members" className={linkClass('/members')} onClick={handleNavClick}>
            メンバー
          </Link>
          <Link to="/roles" className={linkClass('/roles')} onClick={handleNavClick}>
            ロール
          </Link>

          {/* グループサブメニュー */}
          <div>
            <button
              onClick={() => setShowGroupSubmenu(!showGroupSubmenu)}
              className={`w-full flex items-center justify-between px-4 py-2.5 text-sm font-medium rounded-lg transition-colors ${
                isGroupActive
                  ? 'bg-accent/10 text-accent-dark border-l-4 border-accent'
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
                      ? 'bg-accent/10 text-accent-dark'
                      : 'text-gray-600 hover:bg-gray-50'
                  }`}
                  onClick={handleNavClick}
                >
                  ロールグループ
                </Link>
                <Link
                  to="/groups"
                  className={`block px-4 py-2 text-sm rounded-lg transition-colors ${
                    location.pathname.startsWith('/groups')
                      ? 'bg-accent/10 text-accent-dark'
                      : 'text-gray-600 hover:bg-gray-50'
                  }`}
                  onClick={handleNavClick}
                >
                  メンバーグループ
                </Link>
              </div>
            )}
          </div>

          <Link to="/attendance" className={linkClass('/attendance')} onClick={handleNavClick}>
            出欠確認
          </Link>
          <Link to="/schedules" className={linkClass('/schedules')} onClick={handleNavClick}>
            日程調整
          </Link>
          <Link to="/calendars" className={linkClass('/calendars')} onClick={handleNavClick}>
            カレンダー
          </Link>

          {(adminRole === 'admin' || adminRole === 'owner') && (
            <Link to="/admin/invite" className={linkClass('/admin/invite')} onClick={handleNavClick}>
              管理者招待
            </Link>
          )}

          <Link to="/settings" className={linkClass('/settings')} onClick={handleNavClick}>
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
      <div className="flex-1 md:ml-0 min-w-0">
        {/* モバイル用ヘッダー */}
        <header className="md:hidden sticky top-0 z-30 bg-vrc-dark shadow-md">
          <div className="flex items-center justify-between px-4 py-3">
            <button
              onClick={() => setSidebarOpen(true)}
              className="p-2 text-white hover:bg-white/10 rounded-lg"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <h1 className="text-lg font-bold text-white">VRC Shift</h1>
            <div className="flex items-center gap-1">
              <div className="text-white">
                <TutorialButton />
              </div>
              <div className="text-white">
                <AnnouncementBell />
              </div>
            </div>
          </div>
        </header>

        {/* デスクトップ用ヘッダー */}
        <header className="hidden md:flex sticky top-0 z-30 bg-white shadow-sm border-b border-gray-200">
          <div className="flex items-center justify-end w-full px-6 py-3">
            <div className="flex items-center gap-2">
              <TutorialButton />
              <AnnouncementBell />
            </div>
          </div>
        </header>

        <main className="p-4 md:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
