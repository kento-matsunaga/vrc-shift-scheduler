import { useState, useEffect, useCallback } from 'react';
import {
  generateLicenseKeys,
  listLicenseKeys,
  revokeLicenseKey,
  type LicenseKey,
  type GeneratedKey,
} from '../lib/api';

export default function LicenseKeys() {
  const [licenseKeys, setLicenseKeys] = useState<LicenseKey[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [generatedKeys, setGeneratedKeys] = useState<GeneratedKey[]>([]);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [totalCount, setTotalCount] = useState(0);

  const fetchLicenseKeys = useCallback(async () => {
    try {
      setIsLoading(true);
      const params: { status?: string; limit?: number } = { limit: 100 };
      if (statusFilter) params.status = statusFilter;
      const response = await listLicenseKeys(params);
      setLicenseKeys(response.data.keys);
      setTotalCount(response.data.total_count);
    } catch (err) {
      setError('ライセンスキーの取得に失敗しました');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    fetchLicenseKeys();
  }, [fetchLicenseKeys]);

  const handleGenerate = async (count: number, memo: string) => {
    try {
      const response = await generateLicenseKeys({ count, memo });
      setGeneratedKeys(response.data.keys);
      await fetchLicenseKeys();
    } catch (err) {
      setError('ライセンスキーの生成に失敗しました');
      console.error(err);
    }
  };

  const handleRevoke = async (keyId: string) => {
    if (!confirm('このライセンスキーを失効させますか？この操作は取り消せません。')) return;
    try {
      await revokeLicenseKey(keyId);
      await fetchLicenseKeys();
    } catch (err) {
      setError('ライセンスキーの失効に失敗しました');
      console.error(err);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const copyAllKeys = () => {
    const keys = generatedKeys.map((k) => k.key).join('\n');
    navigator.clipboard.writeText(keys);
  };

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      unused: 'bg-green-100 text-green-800',
      used: 'bg-blue-100 text-blue-800',
      revoked: 'bg-red-100 text-red-800',
    };
    const labels: Record<string, string> = {
      unused: '未使用',
      used: '使用済',
      revoked: '失効',
    };
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${colors[status] || 'bg-gray-100'}`}>
        {labels[status] || status}
      </span>
    );
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">ライセンスキー管理</h2>
          <p className="text-sm text-gray-500 mt-1">
            BOOTH販売用のライセンスキーを発行・管理します
          </p>
        </div>
        <button
          onClick={() => setShowGenerateModal(true)}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
        >
          キーを発行
        </button>
      </div>

      {/* 生成されたキーの表示 */}
      {generatedKeys.length > 0 && (
        <div className="mb-6 rounded-lg bg-green-50 border border-green-200 p-4">
          <div className="flex justify-between items-start mb-3">
            <h4 className="text-sm font-medium text-green-800">
              生成されたキー（{generatedKeys.length}件）
            </h4>
            <div className="flex space-x-2">
              <button
                onClick={copyAllKeys}
                className="text-sm text-green-600 hover:text-green-500 underline"
              >
                すべてコピー
              </button>
              <button
                onClick={() => setGeneratedKeys([])}
                className="text-sm text-gray-600 hover:text-gray-500"
              >
                閉じる
              </button>
            </div>
          </div>
          <div className="space-y-2 max-h-60 overflow-y-auto">
            {generatedKeys.map((key) => (
              <div
                key={key.key_id}
                className="flex items-center justify-between bg-white rounded px-3 py-2 font-mono text-sm"
              >
                <span>{key.key}</span>
                <button
                  onClick={() => copyToClipboard(key.key)}
                  className="text-indigo-600 hover:text-indigo-500 text-xs"
                >
                  コピー
                </button>
              </div>
            ))}
          </div>
          <p className="mt-3 text-xs text-green-600">
            ※ これらのキーは再表示できません。必ずコピーして保存してください。
          </p>
        </div>
      )}

      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4">
          <div className="text-sm text-red-700">{error}</div>
        </div>
      )}

      {/* フィルター */}
      <div className="mb-4 flex items-center space-x-4">
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="block rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        >
          <option value="">すべてのステータス</option>
          <option value="unused">未使用</option>
          <option value="used">使用済</option>
          <option value="revoked">失効</option>
        </select>
        <span className="text-sm text-gray-500">
          全 {totalCount} 件
        </span>
      </div>

      {/* テーブル */}
      {isLoading ? (
        <div className="text-center py-8">読み込み中...</div>
      ) : (
        <div className="bg-white shadow overflow-hidden rounded-lg">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  キーID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  ステータス
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  メモ
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  使用者
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  作成日
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {licenseKeys.map((key) => (
                <tr key={key.key_id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                    {key.key_id.slice(0, 12)}...
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getStatusBadge(key.status)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {key.memo || '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {key.claimed_by ? (
                      <span className="font-mono text-xs">{key.claimed_by.slice(0, 12)}...</span>
                    ) : (
                      '-'
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(key.created_at).toLocaleDateString('ja-JP')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {key.status === 'unused' && (
                      <button
                        onClick={() => handleRevoke(key.key_id)}
                        className="text-red-600 hover:text-red-900"
                      >
                        失効
                      </button>
                    )}
                  </td>
                </tr>
              ))}
              {licenseKeys.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                    ライセンスキーがありません
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* 生成モーダル */}
      {showGenerateModal && (
        <GenerateKeysModal
          onClose={() => setShowGenerateModal(false)}
          onGenerate={handleGenerate}
        />
      )}
    </div>
  );
}

function GenerateKeysModal({
  onClose,
  onGenerate,
}: {
  onClose: () => void;
  onGenerate: (count: number, memo: string) => Promise<void>;
}) {
  const [count, setCount] = useState(10);
  const [memo, setMemo] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      await onGenerate(count, memo);
      onClose();
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75" onClick={onClose} />
        <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">ライセンスキーを発行</h3>
          <form onSubmit={handleSubmit}>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">発行数</label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={count}
                  onChange={(e) => setCount(parseInt(e.target.value) || 1)}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                />
                <p className="mt-1 text-xs text-gray-500">1〜100まで指定できます</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">メモ</label>
                <input
                  type="text"
                  value={memo}
                  onChange={(e) => setMemo(e.target.value)}
                  placeholder="例: BOOTH 2025年12月バッチ"
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                />
              </div>
            </div>
            <div className="mt-6 flex justify-end space-x-3">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
              >
                キャンセル
              </button>
              <button
                type="submit"
                disabled={isLoading}
                className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md hover:bg-indigo-700 disabled:opacity-50"
              >
                {isLoading ? '発行中...' : '発行する'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
