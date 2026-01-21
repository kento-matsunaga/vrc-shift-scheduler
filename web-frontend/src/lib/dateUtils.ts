/**
 * 日付ユーティリティ関数
 * 期間選択・曜日フィルター機能で使用
 *
 * 注意: タイムゾーンの影響を避けるため、日付文字列を直接パースして処理しています。
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

/** 日付範囲生成時の最大日数（パフォーマンス保護） */
const MAX_DATE_RANGE_DAYS = 366;

/**
 * YYYY-MM-DD形式の日付文字列をパースしてUTC日付を生成
 * タイムゾーンの影響を避けるためUTCで処理
 * @param dateStr YYYY-MM-DD形式の日付文字列
 * @returns UTCのDateオブジェクト、無効な場合はnull
 */
function parseDateString(dateStr: string): Date | null {
  if (!dateStr) return null;

  const match = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!match) return null;

  const [, yearStr, monthStr, dayStr] = match;
  const year = parseInt(yearStr, 10);
  const month = parseInt(monthStr, 10) - 1; // 0-indexed
  const day = parseInt(dayStr, 10);

  // 日付の妥当性チェック
  const date = new Date(Date.UTC(year, month, day));
  if (
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month ||
    date.getUTCDate() !== day
  ) {
    return null; // 無効な日付（例: 2月30日）
  }

  return date;
}

/**
 * UTC Dateオブジェクトからローカル日付情報を取得するためのヘルパー
 * 注意: 入力のDateオブジェクトはUTCで作成されている前提
 */
function getLocalDateParts(date: Date): { year: number; month: number; day: number; dayOfWeek: number } {
  return {
    year: date.getUTCFullYear(),
    month: date.getUTCMonth(),
    day: date.getUTCDate(),
    dayOfWeek: date.getUTCDay(),
  };
}

/**
 * 期間から日付配列を生成
 * @param startDate 開始日（YYYY-MM-DD形式）
 * @param endDate 終了日（YYYY-MM-DD形式）
 * @param maxDays 最大日数（デフォルト: 366）
 * @returns 日付配列（YYYY-MM-DD形式）
 */
export function generateDateRange(startDate: string, endDate: string, maxDays: number = MAX_DATE_RANGE_DAYS): string[] {
  const start = parseDateString(startDate);
  const end = parseDateString(endDate);

  if (!start || !end) return [];
  if (start > end) return [];

  const dates: string[] = [];
  const current = new Date(start);
  let count = 0;

  while (current <= end && count < maxDays) {
    const { year, month, day } = getLocalDateParts(current);
    dates.push(
      `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    );
    current.setUTCDate(current.getUTCDate() + 1);
    count++;
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
    const date = parseDateString(dateStr);
    if (!date) return false; // 無効な日付をスキップ
    const dayOfWeek = date.getUTCDay();
    return includeDays.includes(dayOfWeek);
  });
}

/**
 * 日付を曜日付きでフォーマット
 * @param date 日付（YYYY-MM-DD形式またはDateオブジェクト）
 * @returns フォーマットされた文字列（例：2026/01/22(水)）
 */
export function formatDateWithWeekday(date: string | Date): string {
  let parsedDate: Date | null;

  if (typeof date === 'string') {
    parsedDate = parseDateString(date);
  } else {
    parsedDate = date;
  }

  if (!parsedDate || isNaN(parsedDate.getTime())) return '';

  // 文字列の場合はUTC、Dateオブジェクトの場合はローカル時刻として扱う
  const isUTC = typeof date === 'string';
  const year = isUTC ? parsedDate.getUTCFullYear() : parsedDate.getFullYear();
  const month = isUTC ? parsedDate.getUTCMonth() + 1 : parsedDate.getMonth() + 1;
  const day = isUTC ? parsedDate.getUTCDate() : parsedDate.getDate();
  const dayOfWeek = isUTC ? parsedDate.getUTCDay() : parsedDate.getDay();
  const weekday = WEEKDAY_NAMES[dayOfWeek];

  return `${year}/${String(month).padStart(2, '0')}/${String(day).padStart(2, '0')}(${weekday})`;
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
  return parseDateString(dateStr) !== null;
}
