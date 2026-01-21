import { useEffect } from 'react';

const APP_NAME = 'VRCシフト管理';

/**
 * ページタイトルを設定するカスタムフック
 * @param title - ページ固有のタイトル
 * @param options - オプション設定
 */
export function useDocumentTitle(
  title: string,
  options: { suffix?: boolean } = { suffix: true }
) {
  useEffect(() => {
    const previousTitle = document.title;

    if (options.suffix && title) {
      document.title = `${title} | ${APP_NAME}`;
    } else if (title) {
      document.title = title;
    } else {
      document.title = APP_NAME;
    }

    return () => {
      document.title = previousTitle;
    };
  }, [title, options.suffix]);
}

/**
 * パスからページタイトルを取得するマッピング
 */
export const routeTitleMap: Record<string, string> = {
  '/events': 'イベント一覧',
  '/members': 'メンバー',
  '/roles': 'ロール',
  '/role-groups': 'ロールグループ',
  '/groups': 'メンバーグループ',
  '/attendance': '出欠確認',
  '/schedules': '日程調整',
  '/settings': '設定',
  '/admin/invite': '管理者招待',
};

/**
 * パスパターンからタイトルを取得
 * 動的ルート（:id等）にも対応
 */
export function getTitleFromPath(pathname: string): string {
  // 完全一致を先にチェック
  if (routeTitleMap[pathname]) {
    return routeTitleMap[pathname];
  }

  // パターンマッチング（動的ルート対応）
  if (pathname.match(/^\/events\/[^/]+\/business-days$/)) {
    return '営業日';
  }
  if (pathname.match(/^\/events\/[^/]+\/templates$/)) {
    return 'テンプレート';
  }
  if (pathname.match(/^\/events\/[^/]+\/templates\/new$/)) {
    return 'テンプレート作成';
  }
  if (pathname.match(/^\/events\/[^/]+\/templates\/[^/]+$/)) {
    return 'テンプレート詳細';
  }
  if (pathname.match(/^\/events\/[^/]+\/templates\/[^/]+\/edit$/)) {
    return 'テンプレート編集';
  }
  if (pathname.match(/^\/events\/[^/]+\/instances$/)) {
    return 'インスタンス';
  }
  if (pathname.match(/^\/business-days\/[^/]+\/shift-slots$/)) {
    return 'シフト枠';
  }
  if (pathname.match(/^\/shift-slots\/[^/]+\/assign$/)) {
    return 'シフト割当';
  }
  if (pathname.match(/^\/attendance\/[^/]+\/shift-adjustment$/)) {
    return 'シフト調整';
  }
  if (pathname.match(/^\/attendance\/[^/]+$/)) {
    return '出欠確認詳細';
  }
  if (pathname.match(/^\/schedules\/[^/]+$/)) {
    return '日程調整詳細';
  }

  // デフォルト
  return '';
}
