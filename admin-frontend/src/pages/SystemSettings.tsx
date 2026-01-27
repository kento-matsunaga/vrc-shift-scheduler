import { useState, useEffect, useCallback } from 'react';
import { getReleaseStatus, updateReleaseStatus } from '../lib/api';

export default function SystemSettings() {
  const [released, setReleased] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const fetchReleaseStatus = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await getReleaseStatus();
      setReleased(response.data.released);
    } catch (err) {
      setError('リリース状態の取得に失敗しました');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReleaseStatus();
  }, [fetchReleaseStatus]);

  const handleToggle = async () => {
    const newStatus = !released;
    const confirmMessage = newStatus
      ? 'サービスをリリース状態にしますか？\nランディングページのCTAボタンが有効になります。'
      : 'サービスをリリース前状態にしますか？\nランディングページのCTAボタンが無効になり「リリース前です」と表示されます。';

    if (!confirm(confirmMessage)) return;

    try {
      setIsUpdating(true);
      setError(null);
      setSuccessMessage(null);
      await updateReleaseStatus(newStatus);
      setReleased(newStatus);
      setSuccessMessage(
        newStatus
          ? 'リリース状態に変更しました'
          : 'リリース前状態に変更しました'
      );
    } catch (err) {
      setError('リリース状態の更新に失敗しました');
      console.error(err);
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">システム設定</h2>
        <p className="text-sm text-gray-500 mt-1">
          サービス全体に影響する設定を管理します
        </p>
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4">
          <div className="text-sm text-red-700">{error}</div>
        </div>
      )}

      {successMessage && (
        <div className="mb-4 rounded-md bg-green-50 p-4">
          <div className="text-sm text-green-700">{successMessage}</div>
        </div>
      )}

      {isLoading ? (
        <div className="text-center py-8">読み込み中...</div>
      ) : (
        <div className="bg-white shadow rounded-lg">
          <div className="p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">
              リリース状態
            </h3>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">
                  現在の状態:
                  <span
                    className={`ml-2 px-2 py-1 text-xs font-medium rounded-full ${
                      released
                        ? 'bg-green-100 text-green-800'
                        : 'bg-yellow-100 text-yellow-800'
                    }`}
                  >
                    {released ? 'リリース済み' : 'リリース前'}
                  </span>
                </p>
                <p className="text-xs text-gray-500 mt-2">
                  {released
                    ? 'ランディングページのCTAボタンが有効です'
                    : 'ランディングページのCTAボタンは無効になり「リリース前です」と表示されます'}
                </p>
              </div>
              <button
                onClick={handleToggle}
                disabled={isUpdating}
                className={`inline-flex items-center px-4 py-2 border text-sm font-medium rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 ${
                  released
                    ? 'border-yellow-300 text-yellow-700 bg-yellow-50 hover:bg-yellow-100 focus:ring-yellow-500'
                    : 'border-green-300 text-green-700 bg-green-50 hover:bg-green-100 focus:ring-green-500'
                } ${isUpdating ? 'opacity-50 cursor-not-allowed' : ''}`}
              >
                {isUpdating ? (
                  '更新中...'
                ) : released ? (
                  'リリース前に戻す'
                ) : (
                  'リリースする'
                )}
              </button>
            </div>
          </div>

          <div className="border-t border-gray-200 px-6 py-4 bg-gray-50 rounded-b-lg">
            <h4 className="text-sm font-medium text-gray-700 mb-2">
              この設定について
            </h4>
            <ul className="text-xs text-gray-500 space-y-1 list-disc list-inside">
              <li>
                リリース前: ランディングページの「今すぐ始める」ボタンが無効化され、ユーザーは登録できません
              </li>
              <li>
                リリース済み: ランディングページから通常通りサービス登録が可能になります
              </li>
              <li>
                この設定は即時反映されます
              </li>
            </ul>
          </div>
        </div>
      )}
    </div>
  );
}
