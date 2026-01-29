import { useState } from 'react';

interface GenerateBusinessDaysModalProps {
  eventName: string;
  onConfirm: (months: number) => void;
  onCancel: () => void;
  loading?: boolean;
}

const PERIOD_OPTIONS = [
  { label: '1ヶ月', value: 1 },
  { label: '3ヶ月', value: 3 },
  { label: '6ヶ月', value: 6 },
  { label: '12ヶ月', value: 12 },
];

export default function GenerateBusinessDaysModal({
  eventName,
  onConfirm,
  onCancel,
  loading = false,
}: GenerateBusinessDaysModalProps) {
  const [selectedMonths, setSelectedMonths] = useState(3);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <h3 className="text-xl font-bold text-gray-900 mb-4">
          営業日を生成
        </h3>

        <p className="text-gray-600 mb-4">
          <span className="font-medium">{eventName}</span> の営業日を生成します。
        </p>

        <p className="text-sm text-gray-600 mb-3">
          何ヶ月先まで生成しますか？
        </p>

        <div className="grid grid-cols-2 gap-2 mb-6">
          {PERIOD_OPTIONS.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => setSelectedMonths(option.value)}
              className={`px-4 py-3 rounded-lg border-2 transition-colors ${
                selectedMonths === option.value
                  ? 'border-accent bg-accent/10 text-accent-dark font-medium'
                  : 'border-gray-200 hover:border-gray-300 text-gray-700'
              }`}
              disabled={loading}
            >
              {option.label}
            </button>
          ))}
        </div>

        <div className="flex space-x-3">
          <button
            type="button"
            onClick={onCancel}
            className="flex-1 btn-secondary"
            disabled={loading}
          >
            キャンセル
          </button>
          <button
            type="button"
            onClick={() => onConfirm(selectedMonths)}
            className="flex-1 btn-primary"
            disabled={loading}
          >
            {loading ? '生成中...' : '生成'}
          </button>
        </div>
      </div>
    </div>
  );
}
