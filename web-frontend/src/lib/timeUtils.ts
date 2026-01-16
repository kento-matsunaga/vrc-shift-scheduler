/**
 * 時刻関連のユーティリティ関数
 *
 * 時刻データの仕様:
 * - バックエンドはGoのtime.Time型を使用（PostgreSQLのTIME型にマッピング）
 * - フロントエンドからはISO 8601形式でダミー日付（0001-01-01）と共に送信
 * - タイムゾーンはUTC（Z）で統一し、表示時にタイムゾーン変換しない
 */

/**
 * 時刻文字列をHH:MM形式に変換
 *
 * 重要: new Date()でパースするとブラウザがローカルタイムゾーンに変換してしまうため、
 * 正規表現で時刻部分を直接抽出する
 *
 * 対応フォーマット:
 * - ISO 8601形式: "0001-01-01T21:00:00Z" → "21:00"
 * - HH:MM:SS形式: "21:00:00" → "21:00"
 * - HH:MM形式: "21:00" → "21:00"
 *
 * エッジケース:
 * - undefined/null → ""
 * - 空文字 → ""
 * - "0001-01-01T00:00:00Z" → "00:00" (深夜0時)
 * - "0001-01-01T23:59:00Z" → "23:59" (23:59)
 */
export const formatTime = (timeStr?: string): string => {
  if (!timeStr) return '';

  // ISO 8601形式からHH:MM形式を抽出（タイムゾーン変換なし）
  const match = timeStr.match(/T(\d{2}:\d{2})/);
  if (match) {
    return match[1];
  }

  // フォールバック: HH:MM:SS形式の場合、先頭5文字を取得
  return timeStr.substring(0, 5);
};

/**
 * 時刻文字列をAPI送信用のISO 8601形式に変換
 *
 * @param time HH:MM形式の時刻文字列
 * @returns ISO 8601形式（例: "0001-01-01T21:00:00Z"）、空の場合はundefined
 *
 * エッジケース:
 * - "" → undefined
 * - "21:00" → "0001-01-01T21:00:00Z"
 * - "00:00" → "0001-01-01T00:00:00Z"
 */
export const toApiTimeFormat = (time: string): string | undefined => {
  if (!time) return undefined;
  return `0001-01-01T${time}:00Z`;
};

/**
 * 時刻バリデーション: 開始時間が終了時間より前かチェック
 *
 * 数値パース後の比較でより堅牢なバリデーションを実行
 *
 * @param startTime HH:MM形式の開始時刻
 * @param endTime HH:MM形式の終了時刻
 * @returns 有効な場合true、無効な場合false
 *
 * エッジケース:
 * - ("", "") → true (両方未設定は有効)
 * - ("21:00", "") → true (片方のみ設定は有効)
 * - ("", "23:00") → true (片方のみ設定は有効)
 * - ("21:00", "23:00") → true (開始 < 終了)
 * - ("23:00", "21:00") → false (開始 > 終了)
 * - ("21:00", "21:00") → false (同じ時間は無効)
 * - ("00:00", "23:59") → true (深夜0時から23:59まで)
 */
export const isValidTimeRange = (startTime: string, endTime: string): boolean => {
  // 両方未設定、または片方のみ設定は有効
  if (!startTime || !endTime) {
    return true;
  }

  const [startH, startM] = startTime.split(':').map(Number);
  const [endH, endM] = endTime.split(':').map(Number);
  const startMinutes = startH * 60 + startM;
  const endMinutes = endH * 60 + endM;

  return startMinutes < endMinutes;
};

/**
 * 時間表示用のフォーマット（片方のみ設定も対応）
 *
 * @param startTime 開始時刻（ISO 8601形式またはHH:MM形式）
 * @param endTime 終了時刻（ISO 8601形式またはHH:MM形式）
 * @param separator 区切り文字（デフォルト: " 〜 "）
 * @returns フォーマットされた時間範囲文字列
 *
 * 表示パターン:
 * - 両方設定: "21:00 〜 23:00"
 * - 開始のみ: "21:00 〜"
 * - 終了のみ: "〜 23:00"
 * - 両方未設定: ""
 */
export const formatTimeRange = (
  startTime?: string,
  endTime?: string,
  separator = ' 〜 '
): string => {
  const formattedStart = formatTime(startTime);
  const formattedEnd = formatTime(endTime);

  if (formattedStart && formattedEnd) {
    return `${formattedStart}${separator}${formattedEnd}`;
  }
  if (formattedStart) {
    return `${formattedStart}${separator.trim()}`;
  }
  if (formattedEnd) {
    return `${separator.trim()} ${formattedEnd}`;
  }
  return '';
};
