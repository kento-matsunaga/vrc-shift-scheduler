import { useState, useEffect, useRef } from 'react';
import { getAnnouncements, getUnreadCount, markAsRead, markAllAsRead, type Announcement } from '../lib/api/announcementApi';

export function AnnouncementBell() {
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // 未読件数を取得
  useEffect(() => {
    fetchUnreadCount(true); // 初回は通知あり
    const interval = setInterval(() => fetchUnreadCount(false), 60000); // ポーリングは通知なし
    return () => clearInterval(interval);
  }, []);

  // 外側クリックで閉じる
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  function showError(message: string) {
    setError(message);
    setTimeout(() => setError(null), 3000);
  }

  async function fetchUnreadCount(showNotify = false) {
    try {
      const count = await getUnreadCount();
      setUnreadCount(count);
    } catch (err) {
      console.error('Failed to fetch unread count:', err);
      if (showNotify) {
        showError('お知らせの取得に失敗しました');
      }
    }
  }

  async function fetchAnnouncements() {
    setLoading(true);
    setError(null);
    try {
      const data = await getAnnouncements();
      setAnnouncements(data);
    } catch (err) {
      console.error('Failed to fetch announcements:', err);
      showError('お知らせの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }

  async function handleOpen() {
    if (!isOpen) {
      await fetchAnnouncements();
    }
    setIsOpen(!isOpen);
  }

  async function handleMarkAsRead(id: string) {
    try {
      await markAsRead(id);
      setAnnouncements(prev => prev.map(a => a.id === id ? { ...a, is_read: true } : a));
      setUnreadCount(prev => Math.max(0, prev - 1));
    } catch (err) {
      console.error('Failed to mark as read:', err);
      showError('既読にできませんでした');
    }
  }

  async function handleMarkAllAsRead() {
    try {
      await markAllAsRead();
      setAnnouncements(prev => prev.map(a => ({ ...a, is_read: true })));
      setUnreadCount(0);
    } catch (err) {
      console.error('Failed to mark all as read:', err);
      showError('既読にできませんでした');
    }
  }

  function formatDate(dateString: string) {
    const date = new Date(dateString);
    return date.toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' });
  }

  function handleAnnouncementClick(announcement: Announcement) {
    // 展開/折りたたみをトグル
    if (expandedId === announcement.id) {
      setExpandedId(null);
    } else {
      setExpandedId(announcement.id);
      // 未読なら既読にする
      if (!announcement.is_read) {
        handleMarkAsRead(announcement.id);
      }
    }
  }

  function handleKeyDown(event: React.KeyboardEvent, announcement: Announcement) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleAnnouncementClick(announcement);
    }
  }

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={handleOpen}
        className="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-full transition-colors"
        aria-label={unreadCount > 0 ? `お知らせ（${unreadCount}件の未読）` : 'お知らせ'}
        aria-expanded={isOpen}
        aria-haspopup="true"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        {unreadCount > 0 && (
          <span className="absolute top-1 right-1 w-2.5 h-2.5 bg-orange-500 rounded-full" />
        )}
      </button>

      {isOpen && (
        <div
          role="menu"
          aria-label="お知らせ一覧"
          className="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg border border-gray-200 z-50"
        >
          <div className="p-3 border-b border-gray-200 flex justify-between items-center">
            <h3 id="announcement-heading" className="font-semibold text-gray-900">お知らせ</h3>
            {unreadCount > 0 && (
              <button
                onClick={handleMarkAllAsRead}
                className="text-sm text-blue-600 hover:text-blue-800"
                aria-label={`${unreadCount}件のお知らせをすべて既読にする`}
              >
                すべて既読
              </button>
            )}
          </div>

          {error && (
            <div className="px-3 py-2 bg-red-50 border-b border-red-200 text-sm text-red-600">
              {error}
            </div>
          )}

          <div className="max-h-96 overflow-y-auto">
            {loading ? (
              <div className="p-4 text-center text-gray-500">読み込み中...</div>
            ) : announcements.length === 0 ? (
              <div className="p-4 text-center text-gray-500">お知らせはありません</div>
            ) : (
              announcements.map(announcement => {
                const isExpanded = expandedId === announcement.id;
                return (
                  <div
                    key={announcement.id}
                    role="menuitem"
                    tabIndex={0}
                    aria-expanded={isExpanded}
                    aria-label={`${announcement.title}${!announcement.is_read ? '（未読）' : ''}、${formatDate(announcement.published_at)}。${isExpanded ? '折りたたむ' : '展開する'}`}
                    onClick={() => handleAnnouncementClick(announcement)}
                    onKeyDown={(e) => handleKeyDown(e, announcement)}
                    className={`p-3 border-b border-gray-100 cursor-pointer hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset ${
                      !announcement.is_read ? 'bg-blue-50' : ''
                    }`}
                  >
                    <div className="flex items-start gap-2">
                      {!announcement.is_read && (
                        <span className="mt-1.5 w-2 h-2 bg-orange-500 rounded-full flex-shrink-0" />
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between gap-2">
                          <h4 className="font-medium text-gray-900 truncate">{announcement.title}</h4>
                          <div className="flex items-center gap-1 flex-shrink-0">
                            <span className="text-xs text-gray-500">
                              {formatDate(announcement.published_at)}
                            </span>
                            <svg
                              className={`w-4 h-4 text-gray-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                            </svg>
                          </div>
                        </div>
                        {isExpanded && (
                          <p className="mt-2 text-sm text-gray-600 whitespace-pre-wrap">{announcement.body}</p>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
}
