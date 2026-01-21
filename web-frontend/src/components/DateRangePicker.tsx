import { useState, useMemo } from 'react';
import {
  generateDateRange,
  filterByWeekdays,
  formatDateWithWeekday,
  getPresetDateRange,
  WEEKDAY_LABELS,
  type DatePreset,
} from '../lib/dateUtils';

/** 日付入力データ型 */
export interface DateInput {
  date: string;       // YYYY-MM-DD形式
  startTime: string;  // HH:MM形式（任意）
  endTime: string;    // HH:MM形式（任意）
}

interface DateRangePickerProps {
  /** 日付を追加するコールバック */
  onAddDates: (dates: DateInput[]) => void;
  /** 既存の日付（重複チェック用） */
  existingDates?: string[];
  /** 無効状態 */
  disabled?: boolean;
}

/** プリセットボタンの定義 */
const PRESETS: { key: DatePreset; label: string }[] = [
  { key: 'thisWeek', label: '今週' },
  { key: 'nextWeek', label: '来週' },
  { key: 'thisMonth', label: '今月' },
  { key: 'nextMonth', label: '来月' },
];

export function DateRangePicker({
  onAddDates,
  existingDates = [],
  disabled = false,
}: DateRangePickerProps) {
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [includeDays, setIncludeDays] = useState<number[]>([0, 1, 2, 3, 4, 5, 6]); // 全曜日
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');

  // プレビュー用の日付リストを計算
  const previewDates = useMemo(() => {
    if (!startDate || !endDate) return [];

    const allDates = generateDateRange(startDate, endDate);
    const filteredDates = filterByWeekdays(allDates, includeDays);

    // 既存の日付を除外
    const existingSet = new Set(existingDates);
    return filteredDates.filter((d) => !existingSet.has(d));
  }, [startDate, endDate, includeDays, existingDates]);

  // 曜日チェックボックスの切り替え
  const toggleWeekday = (day: number) => {
    setIncludeDays((prev) =>
      prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day]
    );
  };

  // 全選択/全解除
  const toggleAllWeekdays = () => {
    if (includeDays.length === 7) {
      setIncludeDays([]);
    } else {
      setIncludeDays([0, 1, 2, 3, 4, 5, 6]);
    }
  };

  // 平日のみ選択
  const selectWeekdaysOnly = () => {
    setIncludeDays([1, 2, 3, 4, 5]); // 月〜金
  };

  // 週末のみ選択
  const selectWeekendsOnly = () => {
    setIncludeDays([0, 6]); // 日、土
  };

  // プリセット選択
  const handlePresetSelect = (preset: DatePreset) => {
    const range = getPresetDateRange(preset);
    setStartDate(range.start);
    setEndDate(range.end);
  };

  // 日付を追加
  const handleAddDates = () => {
    if (previewDates.length === 0) return;

    const datesToAdd: DateInput[] = previewDates.map((date) => ({
      date,
      startTime,
      endTime,
    }));

    onAddDates(datesToAdd);

    // フォームをリセット
    setStartDate('');
    setEndDate('');
    setStartTime('');
    setEndTime('');
  };

  return (
    <details className="bg-accent/5 border border-accent/20 rounded-lg">
      <summary className="px-4 py-3 cursor-pointer hover:bg-accent/10 transition-colors rounded-lg">
        <span className="font-medium text-gray-700">
          📅 期間から一括追加
        </span>
      </summary>

      <div className="px-4 pb-4 pt-2 space-y-4">
        {/* 期間指定 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            期間を指定
          </label>
          <div className="flex flex-col sm:flex-row gap-2 items-start sm:items-center">
            <div className="flex-1 w-full">
              <label className="block text-xs text-gray-500 mb-1">開始日</label>
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={disabled}
              />
            </div>
            <span className="text-gray-500 hidden sm:block pt-5">〜</span>
            <div className="flex-1 w-full">
              <label className="block text-xs text-gray-500 mb-1">終了日</label>
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={disabled}
              />
            </div>
          </div>
        </div>

        {/* クイック選択 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            クイック選択
          </label>
          <div className="flex flex-wrap gap-2">
            {PRESETS.map((preset) => (
              <button
                key={preset.key}
                type="button"
                onClick={() => handlePresetSelect(preset.key)}
                className="px-3 py-1.5 text-sm bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition-colors"
                disabled={disabled}
              >
                {preset.label}
              </button>
            ))}
          </div>
        </div>

        {/* 曜日選択 */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-medium text-gray-700">
              含める曜日
            </label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={selectWeekdaysOnly}
                className="text-xs text-accent hover:underline"
                disabled={disabled}
              >
                平日のみ
              </button>
              <span className="text-gray-300">|</span>
              <button
                type="button"
                onClick={selectWeekendsOnly}
                className="text-xs text-accent hover:underline"
                disabled={disabled}
              >
                週末のみ
              </button>
              <span className="text-gray-300">|</span>
              <button
                type="button"
                onClick={toggleAllWeekdays}
                className="text-xs text-accent hover:underline"
                disabled={disabled}
              >
                {includeDays.length === 7 ? '全解除' : '全選択'}
              </button>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {WEEKDAY_LABELS.map((day) => (
              <label
                key={day.value}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md border cursor-pointer transition-colors ${
                  includeDays.includes(day.value)
                    ? 'bg-accent text-white border-accent'
                    : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                } ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
              >
                <input
                  type="checkbox"
                  checked={includeDays.includes(day.value)}
                  onChange={() => toggleWeekday(day.value)}
                  className="sr-only"
                  disabled={disabled}
                />
                <span className="text-sm font-medium">{day.label}</span>
              </label>
            ))}
          </div>
        </div>

        {/* 一括時間設定 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            一括設定（任意）
          </label>
          <div className="flex flex-col sm:flex-row gap-2">
            <div className="flex-1">
              <label className="block text-xs text-gray-500 mb-1">開始時間</label>
              <input
                type="time"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={disabled}
              />
            </div>
            <div className="flex-1">
              <label className="block text-xs text-gray-500 mb-1">終了時間</label>
              <input
                type="time"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
                disabled={disabled}
              />
            </div>
          </div>
        </div>

        {/* プレビュー */}
        {startDate && endDate && (
          <div className="border-t border-gray-200 pt-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-700">プレビュー</span>
              {previewDates.length > 0 && (
                <span className="text-sm text-accent font-medium">
                  {previewDates.length}日分が追加されます
                </span>
              )}
            </div>
            {previewDates.length > 0 ? (
              <div className="bg-white border border-gray-200 rounded-md p-3 max-h-40 sm:max-h-48 overflow-y-auto">
                <div className="flex flex-wrap gap-2">
                  {previewDates.slice(0, 20).map((date) => (
                    <span
                      key={date}
                      className="text-xs bg-gray-100 text-gray-700 px-2 py-1 rounded"
                    >
                      {formatDateWithWeekday(date)}
                    </span>
                  ))}
                  {previewDates.length > 20 && (
                    <span className="text-xs text-gray-500">
                      ...他 {previewDates.length - 20} 件
                    </span>
                  )}
                </div>
              </div>
            ) : (
              <p className="text-sm text-gray-500">
                {includeDays.length === 0
                  ? '曜日を選択してください'
                  : '追加できる日付がありません（全て既に追加済み）'}
              </p>
            )}
          </div>
        )}

        {/* 追加ボタン */}
        <button
          type="button"
          onClick={handleAddDates}
          disabled={disabled || previewDates.length === 0}
          className="w-full px-4 py-2 bg-accent text-white rounded-md hover:bg-accent-dark transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          候補日に追加する
        </button>
      </div>
    </details>
  );
}
