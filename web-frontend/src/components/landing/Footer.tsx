import { Link } from 'react-router-dom';

export function Footer() {
  return (
    <footer
      className="relative py-8 sm:py-12 px-4 sm:px-6"
      style={{
        borderTop: '1px solid rgba(139, 92, 246, 0.1)',
        paddingBottom: 'max(env(safe-area-inset-bottom, 0px), 2rem)',
      }}
    >
      <div className="max-w-6xl mx-auto">
        <div className="flex flex-col items-center gap-6">
          <div className="flex items-center gap-2 sm:gap-3">
            <div
              className="w-7 h-7 sm:w-8 sm:h-8 rounded-lg flex items-center justify-center"
              style={{
                background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
              }}
            >
              <span className="text-xs sm:text-sm">📅</span>
            </div>
            <span className="font-semibold text-white text-sm sm:text-base">VRC Shift Scheduler</span>
          </div>

          <nav className="flex flex-wrap items-center justify-center gap-x-4 sm:gap-x-6 gap-y-2 text-xs sm:text-sm text-gray-400">
            <Link to="/admin/login" className="hover:text-white transition-colors py-1 min-h-[44px] flex items-center">
              ログイン
            </Link>
            <Link to="/terms" className="hover:text-white transition-colors py-1 min-h-[44px] flex items-center">
              利用規約
            </Link>
            <Link to="/privacy" className="hover:text-white transition-colors py-1 min-h-[44px] flex items-center">
              プライバシーポリシー
            </Link>
            <a href="mailto:support@vrcshift.com" className="hover:text-white transition-colors py-1 min-h-[44px] flex items-center">
              お問い合わせ
            </a>
          </nav>

          <p className="text-gray-500 text-xs sm:text-sm">© 2025 VRC Shift Scheduler</p>
        </div>
      </div>
    </footer>
  );
}
