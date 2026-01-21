/**
 * 日付ユーティリティ関数
 * 期間選択・曜日フィルター機能で使用
 */

/** 曜日の日本語表記 */
const WEEKDAY_NAMES = ['日', '月', '火', '水', '木', '金', '土'] as const;

/** 曜日のラベル（チェックボックス用） */
export const WEEKDAY_LABELS = [
  { value: 1, label: '月' },
  { value: 2, label: '火' },
  { value: 3, label: '水' },
  { value: 4, label: '木' },
  { value: 5, label: '金' },
  { value: 6, label: '土' },
  { value: 0, label: '日' },
] as const;

/** プリセットの種類 */
export type DatePreset = 'thisWeek' | 'nextWeek' | 'thisMonth' | 'nextMonth';

/**
 * 期間から日付配列を生成
 * @param startDate 開始日（YYYY-MM-DD形式）
 * @param endDate 終了日（YYYY-MM-DD形式）
 * @returns 日付配列（YYYY-MM-DD形式）
 */
export function generateDateRange(startDate: string, endDate: string): string[] {
  if (!startDate || !endDate) return [];

  const start = new Date(startDate);
  const end = new Date(endDate);

  if (isNaN(start.getTime()) || isNaN(end.getTime())) return [];
  if (start > end) return [];

  const dates: string[] = [];
  const current = new Date(start);

  while (current <= end) {
    dates.push(formatDateToYYYYMMDD(current));
    current.setDate(current.getDate() + 1);
  }

  return dates;
}

/**
 * 曜日でフィルタリング
 * @param dates 日付配列（YYYY-MM-DD形式）
 * @param includeDays 含める曜日の配列（0=日曜, 1=月曜, ..., 6=土曜）
 * @returns フィルタリングされた日付配列
 */
export function filterByWeekdays(dates: string[], includeDays: number[]): string[] {
  if (includeDays.length === 0) return [];
  if (includeDays.length === 7) return dates;

  return dates.filter((dateStr) => {
    const date = new Date(dateStr);
    const dayOfWeek = date.getDay();
    return includeDays.includes(dayOfWeek);
  });
}

/**
 * 日付を曜日付きでフォーマット
 * @param date 日付（YYYY-MM-DD形式またはDateオブジェクト）
 * @returns フォーマットされた文字列（例：2026/01/22(水)）
 */
export function formatDateWithWeekday(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  if (isNaN(d.getTime())) return '';

  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  const weekday = WEEKDAY_NAMES[d.getDay()];

  return `${year}/${month}/${day}(${weekday})`;
}

/**
 * 日付をYYYY-MM-DD形式にフォーマット
 * @param date Dateオブジェクト
 * @returns YYYY-MM-DD形式の文字列
 */
export function formatDateToYYYYMMDD(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * クイック選択用のプリセット期間を取得
 * @param preset プリセットの種類
 * @returns 開始日と終了日（YYYY-MM-DD形式）
 */
export function getPresetDateRange(preset: DatePreset): { start: string; end: string } {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const dayOfWeek = today.getDay(); // 0=日曜, 1=月曜, ..., 6=土曜

  switch (preset) {
    case 'thisWeek': {
      // 今日から今週の日曜日まで
      const start = new Date(today);
      const end = new Date(today);
      const daysUntilSunday = dayOfWeek === 0 ? 0 : 7 - dayOfWeek;
      end.setDate(today.getDate() + daysUntilSunday);
      return {
        start: formatDateToYYYYMMDD(start),
        end: formatDateToYYYYMMDD(end),
      };
    }

    case 'nextWeek': {
      // 来週の月曜日から日曜日まで
      const start = new Date(today);
      const daysUntilNextMonday = dayOfWeek === 0 ? 1 : 8 - dayOfWeek;
      start.setDate(today.getDate() + daysUntilNextMonday);
      const end = new Date(start);
      end.setDate(start.getDate() + 6);
      return {
        start: formatDateToYYYYMMDD(start),
        end: formatDateToYYYYMMDD(end),
      };
    }

    case 'thisMonth': {
      // 今日から今月末まで
      const start = new Date(today);
      const end = new Date(today.getFullYear(), today.getMonth() + 1, 0); // 月末
      return {
        start: formatDateToYYYYMMDD(start),
        end: formatDateToYYYYMMDD(end),
      };
    }

    case 'nextMonth': {
      // 来月1日から月末まで
      const start = new Date(today.getFullYear(), today.getMonth() + 1, 1);
      const end = new Date(today.getFullYear(), today.getMonth() + 2, 0); // 来月末
      return {
        start: formatDateToYYYYMMDD(start),
        end: formatDateToYYYYMMDD(end),
      };
    }

    default:
      return { start: '', end: '' };
  }
}

/**
 * 重複を除いて日付をマージ
 * @param existingDates 既存の日付配列（YYYY-MM-DD形式）
 * @param newDates 新しい日付配列（YYYY-MM-DD形式）
 * @returns マージされた日付配列（重複なし、日付順にソート）
 */
export function mergeDatesUnique(existingDates: string[], newDates: string[]): string[] {
  const dateSet = new Set([...existingDates, ...newDates]);
  return Array.from(dateSet).sort();
}

/**
 * 日付が有効かどうかをチェック
 * @param dateStr 日付文字列（YYYY-MM-DD形式）
 * @returns 有効な場合true
 */
export function isValidDate(dateStr: string): boolean {
  if (!dateStr) return false;
  const date = new Date(dateStr);
  return !isNaN(date.getTime());
}
