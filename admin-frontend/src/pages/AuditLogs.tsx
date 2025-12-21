import { useState, useEffect, useCallback } from 'react';
import { listAuditLogs, type AuditLogItem } from '../lib/api';

export default function AuditLogs() {
  const [logs, setLogs] = useState<AuditLogItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actorFilter, setActorFilter] = useState<string>('');
  const [actionFilter, setActionFilter] = useState<string>('');
  const [totalCount, setTotalCount] = useState(0);
  const [selectedLog, setSelectedLog] = useState<AuditLogItem | null>(null);

  const fetchLogs = useCallback(async () => {
    try {
      setIsLoading(true);
      const params: { actor_type?: string; action?: string; limit?: number } = { limit: 100 };
      if (actorFilter) params.actor_type = actorFilter;
      if (actionFilter) params.action = actionFilter;
      const response = await listAuditLogs(params);
      setLogs(response.data.logs);
      setTotalCount(response.data.total_count);
    } catch (err) {
      setError('監査ログの取得に失敗しました');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, [actorFilter, actionFilter]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const getActorBadge = (actorType: string) => {
    const colors: Record<string, string> = {
      admin: 'bg-blue-100 text-blue-800',
      system: 'bg-gray-100 text-gray-800',
      stripe: 'bg-purple-100 text-purple-800',
    };
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${colors[actorType] || 'bg-gray-100'}`}>
        {actorType}
      </span>
    );
  };

  const getActionLabel = (action: string) => {
    const labels: Record<string, string> = {
      license_generated: 'ライセンス発行',
      license_claimed: 'ライセンス使用',
      license_revoked: 'ライセンス失効',
      entitlement_created: 'エンタイトルメント作成',
      entitlement_revoked: 'エンタイトルメント失効',
      tenant_status_changed: 'テナントステータス変更',
      tenant_suspended: 'テナント停止',
    };
    return labels[action] || action;
  };

  const formatJSON = (jsonStr: string | null) => {
    if (!jsonStr) return null;
    try {
      return JSON.stringify(JSON.parse(jsonStr), null, 2);
    } catch {
      return jsonStr;
    }
  };

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">監査ログ</h2>
        <p className="text-sm text-gray-500 mt-1">
          課金に関するすべての操作履歴を確認できます
        </p>
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4">
          <div className="text-sm text-red-700">{error}</div>
        </div>
      )}

      {/* フィルター */}
      <div className="mb-4 flex flex-wrap items-center gap-4">
        <select
          value={actorFilter}
          onChange={(e) => setActorFilter(e.target.value)}
          className="block rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        >
          <option value="">すべてのアクター</option>
          <option value="admin">admin</option>
          <option value="system">system</option>
          <option value="stripe">stripe</option>
        </select>
        <select
          value={actionFilter}
          onChange={(e) => setActionFilter(e.target.value)}
          className="block rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        >
          <option value="">すべてのアクション</option>
          <option value="license_generated">ライセンス発行</option>
          <option value="license_claimed">ライセンス使用</option>
          <option value="license_revoked">ライセンス失効</option>
          <option value="entitlement_created">エンタイトルメント作成</option>
          <option value="tenant_status_changed">テナントステータス変更</option>
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
                  日時
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  アクター
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  アクション
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  対象
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  詳細
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {logs.map((log) => (
                <tr key={log.log_id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(log.created_at).toLocaleString('ja-JP')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex flex-col">
                      {getActorBadge(log.actor_type)}
                      {log.actor_id && (
                        <span className="text-xs text-gray-500 font-mono mt-1">
                          {log.actor_id.slice(0, 12)}...
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {getActionLabel(log.action)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {log.target_type && (
                      <div>
                        <span className="text-gray-700">{log.target_type}</span>
                        {log.target_id && (
                          <span className="text-xs font-mono block">
                            {log.target_id.slice(0, 12)}...
                          </span>
                        )}
                      </div>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {(log.before_json || log.after_json) && (
                      <button
                        onClick={() => setSelectedLog(log)}
                        className="text-indigo-600 hover:text-indigo-900"
                      >
                        表示
                      </button>
                    )}
                  </td>
                </tr>
              ))}
              {logs.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-6 py-8 text-center text-gray-500">
                    監査ログがありません
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* 詳細モーダル */}
      {selectedLog && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-screen items-center justify-center p-4">
            <div
              className="fixed inset-0 bg-gray-500 bg-opacity-75"
              onClick={() => setSelectedLog(null)}
            />
            <div className="relative bg-white rounded-lg shadow-xl max-w-2xl w-full p-6 max-h-[80vh] overflow-y-auto">
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                ログ詳細: {getActionLabel(selectedLog.action)}
              </h3>
              <dl className="space-y-4">
                <div>
                  <dt className="text-sm font-medium text-gray-500">日時</dt>
                  <dd className="text-sm text-gray-900">
                    {new Date(selectedLog.created_at).toLocaleString('ja-JP')}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-gray-500">アクター</dt>
                  <dd className="text-sm text-gray-900">
                    {selectedLog.actor_type}
                    {selectedLog.actor_id && ` (${selectedLog.actor_id})`}
                  </dd>
                </div>
                {selectedLog.target_type && (
                  <div>
                    <dt className="text-sm font-medium text-gray-500">対象</dt>
                    <dd className="text-sm text-gray-900">
                      {selectedLog.target_type}: {selectedLog.target_id}
                    </dd>
                  </div>
                )}
                {selectedLog.before_json && (
                  <div>
                    <dt className="text-sm font-medium text-gray-500">変更前</dt>
                    <dd className="mt-1 bg-gray-100 rounded p-3 overflow-x-auto">
                      <pre className="text-xs text-gray-700 font-mono">
                        {formatJSON(selectedLog.before_json)}
                      </pre>
                    </dd>
                  </div>
                )}
                {selectedLog.after_json && (
                  <div>
                    <dt className="text-sm font-medium text-gray-500">変更後</dt>
                    <dd className="mt-1 bg-gray-100 rounded p-3 overflow-x-auto">
                      <pre className="text-xs text-gray-700 font-mono">
                        {formatJSON(selectedLog.after_json)}
                      </pre>
                    </dd>
                  </div>
                )}
              </dl>
              <div className="mt-6 flex justify-end">
                <button
                  onClick={() => setSelectedLog(null)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  閉じる
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
