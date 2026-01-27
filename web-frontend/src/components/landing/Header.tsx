import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useReleaseStatus } from '../../hooks/useReleaseStatus';

export function Header() {
  const [scrolled, setScrolled] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const { released, isLoading } = useReleaseStatus();

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ${scrolled ? 'py-3' : 'py-5'}`}
      style={{
        background: scrolled ? 'rgba(10, 10, 15, 0.85)' : 'transparent',
        backdropFilter: scrolled ? 'blur(20px) saturate(180%)' : 'none',
        WebkitBackdropFilter: scrolled ? 'blur(20px) saturate(180%)' : 'none',
        borderBottom: scrolled ? '1px solid rgba(139, 92, 246, 0.1)' : 'none',
        paddingTop: 'env(safe-area-inset-top, 0)',
      }}
    >
      <div className="max-w-6xl mx-auto px-6 flex items-center justify-between">
        <a href="#top" className="flex items-center gap-3 group">
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center transition-transform duration-300 group-hover:scale-110"
            style={{
              background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
              boxShadow: '0 4px 20px rgba(79, 70, 229, 0.4)',
            }}
          >
            <span className="text-xl">📅</span>
          </div>
          <span className="font-bold text-lg tracking-tight text-white">VRC Shift Scheduler</span>
        </a>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex items-center gap-8">
          <a href="#features" className="text-gray-400 hover:text-white transition-colors text-sm">
            機能
          </a>
          <a href="#pricing" className="text-gray-400 hover:text-white transition-colors text-sm">
            料金
          </a>
          <Link to="/admin/login" className="text-gray-400 hover:text-white transition-colors text-sm">
            ログイン
          </Link>
          {released ? (
            <Link
              to="/subscribe"
              className="px-5 py-2.5 rounded-full text-sm font-medium transition-all duration-300 hover:scale-105 text-white"
              style={{
                background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
                boxShadow: '0 4px 20px rgba(79, 70, 229, 0.4)',
              }}
            >
              今すぐ始める
            </Link>
          ) : (
            <span
              className="px-5 py-2.5 rounded-full text-sm font-medium text-white cursor-not-allowed opacity-75"
              style={{
                background: 'linear-gradient(135deg, #6b7280 0%, #9ca3af 100%)',
              }}
            >
              {isLoading ? '...' : 'リリース前'}
            </span>
          )}
        </nav>

        {/* Mobile Menu Button */}
        <button
          className="md:hidden p-3 -mr-2 min-w-[44px] min-h-[44px] flex items-center justify-center"
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          aria-label="メニューを開く"
        >
          <div className="w-6 h-5 flex flex-col justify-between">
            <span
              className={`h-0.5 bg-white transition-all duration-300 ${mobileMenuOpen ? 'rotate-45 translate-y-2' : ''}`}
            />
            <span className={`h-0.5 bg-white transition-all duration-300 ${mobileMenuOpen ? 'opacity-0' : ''}`} />
            <span
              className={`h-0.5 bg-white transition-all duration-300 ${mobileMenuOpen ? '-rotate-45 -translate-y-2' : ''}`}
            />
          </div>
        </button>
      </div>

      {/* Mobile Menu */}
      <div
        className={`md:hidden absolute top-full left-0 right-0 transition-all duration-300 ${mobileMenuOpen ? 'opacity-100 visible' : 'opacity-0 invisible'}`}
        style={{
          background: 'rgba(10, 10, 15, 0.95)',
          backdropFilter: 'blur(20px)',
          WebkitBackdropFilter: 'blur(20px)',
          paddingBottom: 'env(safe-area-inset-bottom, 0)',
        }}
      >
        <nav className="flex flex-col p-6 gap-2">
          <a
            href="#features"
            className="text-gray-300 hover:text-white py-3 min-h-[44px] flex items-center"
            onClick={() => setMobileMenuOpen(false)}
          >
            機能
          </a>
          <a
            href="#pricing"
            className="text-gray-300 hover:text-white py-3 min-h-[44px] flex items-center"
            onClick={() => setMobileMenuOpen(false)}
          >
            料金
          </a>
          <Link
            to="/admin/login"
            className="text-gray-300 hover:text-white py-3 min-h-[44px] flex items-center"
            onClick={() => setMobileMenuOpen(false)}
          >
            ログイン
          </Link>
          {released ? (
            <Link
              to="/subscribe"
              className="px-5 py-3 rounded-full text-center font-medium mt-4 text-white min-h-[44px] flex items-center justify-center"
              style={{
                background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
              }}
              onClick={() => setMobileMenuOpen(false)}
            >
              今すぐ始める
            </Link>
          ) : (
            <span
              className="px-5 py-3 rounded-full text-center font-medium mt-4 text-white min-h-[44px] flex items-center justify-center cursor-not-allowed opacity-75"
              style={{
                background: 'linear-gradient(135deg, #6b7280 0%, #9ca3af 100%)',
              }}
            >
              {isLoading ? '...' : 'リリース前です'}
            </span>
          )}
        </nav>
      </div>
    </header>
  );
}
