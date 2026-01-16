import { useState, useEffect, useCallback } from 'react';
import { generateShiftText, copyToClipboard, type MemberSeparator, type InstanceData } from '../lib/shiftTextExport';

interface ShiftTextPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  instanceData: InstanceData[];
}

export default function ShiftTextPreviewModal({
  isOpen,
  onClose,
  instanceData,
}: ShiftTextPreviewModalProps) {
  const [separator, setSeparator] = useState<MemberSeparator>('newline');
  const [text, setText] = useState('');
  const [copied, setCopied] = useState(false);
  const [copyError, setCopyError] = useState(false);

  // instanceDataまたはseparatorが変更されたらテキストを再生成
  useEffect(() => {
    if (isOpen) {
      const generatedText = generateShiftText(instanceData, separator);
      setText(generatedText);
      setCopyError(false);
    }
  }, [instanceData, separator, isOpen]);

  // Escキーでモーダルを閉じる
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, onClose]);

  const handleCopy = async () => {
    setCopyError(false);
    const success = await copyToClipboard(text);
    if (success) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } else {
      setCopyError(true);
      setTimeout(() => setCopyError(false), 3000);
    }
  };

  const handleRegenerate = () => {
    const generatedText = generateShiftText(instanceData, separator);
    setText(generatedText);
  };

  // 背景クリックでモーダルを閉じる
  const handleBackdropClick = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  }, [onClose]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={handleBackdropClick}
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div className="bg-white rounded-lg max-w-2xl w-full p-6 max-h-[90vh] flex flex-col">
        <div className="flex justify-between items-center mb-4">
          <h3 id="modal-title" className="text-xl font-bold text-gray-900">インスタンス表プレビュー</h3>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700"
            aria-label="閉じる"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* 区切り文字選択と再生成ボタン */}
        <div className="flex items-center gap-3 mb-4">
          <label className="text-sm text-gray-600">メンバー区切り:</label>
          <select
            value={separator}
            onChange={(e) => setSeparator(e.target.value as MemberSeparator)}
            className="px-3 py-1.5 border border-gray-300 rounded-lg text-sm bg-white"
          >
            <option value="newline">行区切り</option>
            <option value="comma">カンマ区切り</option>
          </select>
          <button
            onClick={handleRegenerate}
            className="text-sm text-accent hover:text-accent-dark underline"
          >
            再生成
          </button>
        </div>

        {/* 編集可能なテキストエリア */}
        <div className="flex-1 min-h-0 mb-4">
          {text.trim() === '' ? (
            <div className="w-full h-full min-h-[300px] p-3 border border-gray-300 rounded-lg flex items-center justify-center text-gray-500">
              データがありません
            </div>
          ) : (
            <textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              className="w-full h-full min-h-[300px] p-3 border border-gray-300 rounded-lg font-mono text-sm resize-none focus:outline-none focus:ring-2 focus:ring-accent"
              placeholder="シフト配置データがここに表示されます"
            />
          )}
        </div>

        <p className="text-xs text-gray-500 mb-4">
          テキストは自由に編集できます。編集内容は保存されません。
        </p>

        {/* エラーメッセージ */}
        {copyError && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
            クリップボードへのコピーに失敗しました。ブラウザの設定を確認してください。
          </div>
        )}

        {/* アクションボタン */}
        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg"
          >
            閉じる
          </button>
          <button
            onClick={handleCopy}
            disabled={text.trim() === ''}
            className="px-4 py-2 bg-accent text-white hover:bg-accent-dark rounded-lg flex items-center disabled:bg-gray-400 disabled:cursor-not-allowed"
            aria-label="インスタンス表をクリップボードにコピー"
          >
            <svg
              className="w-5 h-5 mr-2"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
              />
            </svg>
            {copied ? 'コピーしました!' : 'コピー'}
          </button>
        </div>
      </div>
    </div>
  );
}
