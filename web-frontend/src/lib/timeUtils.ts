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
 * 時刻形式のバリデーション: HH:MM形式かつ有効な時刻かチェック
 *
 * @param time HH:MM形式の時刻文字列
 * @returns 有効な場合true、無効な場合false
 *
 * バリデーション:
 * - 形式: HH:MM（2桁:2桁）
 * - 時間: 0-23
 * - 分: 0-59
 *
 * エッジケース:
 * - "" → false
 * - "21:00" → true
 * - "00:00" → true
 * - "23:59" → true
 * - "24:00" → false (時間が範囲外)
 * - "25:00" → false (時間が範囲外)
 * - "12:60" → false (分が範囲外)
 * - "1:00" → false (形式不正)
 * - "abc" → false (形式不正)
 */
export const isValidTimeFormat = (time: string): boolean => {
  if (!time) return false;

  // HH:MM形式のチェック
  const match = time.match(/^(\d{2}):(\d{2})$/);
  if (!match) return false;

  const hours = parseInt(match[1], 10);
  const minutes = parseInt(match[2], 10);

  // 時間と分の範囲チェック
  return hours >= 0 && hours <= 23 && minutes >= 0 && minutes <= 59;
};

/**
 * 時刻バリデーション: 有効な時間範囲かチェック（深夜営業対応）
 *
 * 深夜営業（オーバーナイト）をサポートするため、終了時刻が開始時刻より
 * 前の場合も有効とする（例: 21:00-02:00）。
 * 同じ時刻の場合のみ無効とする。
 *
 * @param startTime HH:MM形式の開始時刻
 * @param endTime HH:MM形式の終了時刻
 * @returns 有効な場合true、無効な場合false
 *
 * エッジケース:
 * - ("", "") → true (両方未設定は有効)
 * - ("21:00", "") → true (片方のみ設定は有効)
 * - ("", "23:00") → true (片方のみ設定は有効)
 * - ("21:00", "23:00") → true (通常パターン: 開始 < 終了)
 * - ("21:00", "02:00") → true (深夜営業: 開始 > 終了)
 * - ("21:00", "21:00") → false (同じ時間は無効)
 * - ("00:00", "23:59") → true (深夜0時から23:59まで)
 * - ("25:00", "26:00") → false (不正な時刻形式)
 * - ("12:60", "13:00") → false (不正な時刻形式)
 */
export const isValidTimeRange = (startTime: string, endTime: string): boolean => {
  // 両方未設定、または片方のみ設定は有効
  if (!startTime || !endTime) {
    return true;
  }

  // 時刻形式のバリデーション
  if (!isValidTimeFormat(startTime) || !isValidTimeFormat(endTime)) {
    return false;
  }

  const [startH, startM] = startTime.split(':').map(Number);
  const [endH, endM] = endTime.split(':').map(Number);
  const startMinutes = startH * 60 + startM;
  const endMinutes = endH * 60 + endM;

  // 同じ時刻のみ無効（深夜営業パターンは許可）
  return startMinutes !== endMinutes;
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
