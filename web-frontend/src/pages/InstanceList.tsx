import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  listInstances,
  createInstance,
  updateInstance,
  deleteInstance,
  type Instance,
  type CreateInstanceInput,
  type UpdateInstanceInput,
} from '../lib/api/instanceApi';
import { getEventDetail } from '../lib/api/eventApi';
import type { Event } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function InstanceList() {
  const { eventId } = useParams<{ eventId: string }>();
  const [instances, setInstances] = useState<Instance[]>([]);
  const [event, setEvent] = useState<Event | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingInstance, setEditingInstance] = useState<Instance | null>(null);

  useEffect(() => {
    if (eventId) {
      loadData();
    }
  }, [eventId]);

  const loadData = async () => {
    if (!eventId) return;
    try {
      setLoading(true);
      const [instancesData, eventData] = await Promise.all([
        listInstances(eventId),
        getEventDetail(eventId),
      ]);
      setInstances(instancesData || []);
      setEvent(eventData);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('インスタンス一覧の取得に失敗しました');
      }
      console.error('Failed to load instances:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    loadData();
  };

  const handleUpdateSuccess = () => {
    setEditingInstance(null);
    loadData();
  };

  const handleDelete = async (instanceId: string) => {
    if (!confirm('このインスタンスを削除してもよろしいですか？\n関連するシフト枠からインスタンス参照が解除されます。')) {
      return;
    }

    try {
      await deleteInstance(instanceId);
      loadData();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('インスタンスの削除に失敗しました');
      }
      console.error('Failed to delete instance:', err);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div>
      {/* パンくず */}
      <nav className="mb-4 text-sm text-gray-600">
        <Link to="/events" className="text-accent hover:underline">
          イベント一覧
        </Link>
        <span className="mx-2">/</span>
        <Link to={`/events/${eventId}/business-days`} className="text-accent hover:underline">
          {event?.event_name || 'イベント'}
        </Link>
        <span className="mx-2">/</span>
        <span className="text-gray-900">インスタンス管理</span>
      </nav>

      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-6">
        <div>
          <h2 className="text-xl sm:text-2xl font-bold text-gray-900">インスタンス管理</h2>
          <p className="text-xs sm:text-sm text-gray-600 mt-1">
            VRChatワールドのインスタンスを管理します。シフト枠と紐付けて使用します。
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn-primary text-sm sm:text-base w-full sm:w-auto"
        >
          + インスタンス追加
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
          <button
            onClick={() => setError('')}
            className="text-xs text-red-600 hover:text-red-800 mt-1"
          >
            閉じる
          </button>
        </div>
      )}

      {instances.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600 mb-4">まだインスタンスがありません</p>
          <p className="text-sm text-gray-500 mb-4">
            テンプレートを適用すると自動的にインスタンスが作成されます。
            <br />
            手動で追加することもできます。
          </p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            最初のインスタンスを追加
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {instances
            .sort((a, b) => a.display_order - b.display_order)
            .map((instance) => (
              <div key={instance.instance_id} className="card hover:shadow-lg transition-shadow">
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <div className="text-lg font-bold text-gray-900">{instance.name}</div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setEditingInstance(instance)}
                      className="text-accent hover:text-accent-dark text-sm"
                    >
                      編集
                    </button>
                    <button
                      onClick={() => handleDelete(instance.instance_id)}
                      className="text-red-600 hover:text-red-800 text-sm"
                    >
                      削除
                    </button>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2 mt-3 text-xs text-gray-500">
                  <span className="bg-gray-100 px-2 py-1 rounded">
                    表示順序: {instance.display_order}
                  </span>
                  {instance.max_members !== null && (
                    <span className="bg-blue-50 text-blue-700 px-2 py-1 rounded">
                      最大人数: {instance.max_members}
                    </span>
                  )}
                </div>
              </div>
            ))}
        </div>
      )}

      {/* インスタンス作成モーダル */}
      {showCreateModal && eventId && (
        <InstanceFormModal
          eventId={eventId}
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}

      {/* インスタンス編集モーダル */}
      {editingInstance && eventId && (
        <InstanceFormModal
          eventId={eventId}
          instance={editingInstance}
          onClose={() => setEditingInstance(null)}
          onSuccess={handleUpdateSuccess}
        />
      )}
    </div>
  );
}

// インスタンス作成・編集モーダルコンポーネント
function InstanceFormModal({
  eventId,
  instance,
  onClose,
  onSuccess,
}: {
  eventId: string;
  instance?: Instance;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [name, setName] = useState(instance?.name || '');
  const [displayOrder, setDisplayOrder] = useState(instance?.display_order || 0);
  const [maxMembers, setMaxMembers] = useState<string>(
    instance?.max_members !== null && instance?.max_members !== undefined
      ? String(instance.max_members)
      : ''
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name.trim()) {
      setError('インスタンス名を入力してください');
      return;
    }

    setLoading(true);

    try {
      const maxMembersValue = maxMembers.trim() === '' ? null : parseInt(maxMembers, 10);

      if (instance) {
        const input: UpdateInstanceInput = {
          name: name.trim(),
          display_order: displayOrder,
          max_members: maxMembersValue,
        };
        await updateInstance(instance.instance_id, input);
      } else {
        const input: CreateInstanceInput = {
          name: name.trim(),
          display_order: displayOrder,
          max_members: maxMembersValue,
        };
        await createInstance(eventId, input);
      }

      onSuccess();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError(instance ? 'インスタンスの更新に失敗しました' : 'インスタンスの作成に失敗しました');
      }
      console.error('Failed to save instance:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-md w-full p-6">
        <h3 className="text-xl font-bold text-gray-900 mb-4">
          {instance ? 'インスタンスを編集' : 'インスタンスを追加'}
        </h3>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="name" className="label">
              インスタンス名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="input-field"
              disabled={loading}
              autoFocus
              placeholder="例: インスタンス1、メイン、サブ"
            />
            <p className="text-xs text-gray-500 mt-1">
              VRChatワールドのインスタンスを識別する名前です
            </p>
          </div>

          <div className="mb-4">
            <label htmlFor="displayOrder" className="label">
              表示順序
            </label>
            <input
              type="number"
              id="displayOrder"
              value={displayOrder}
              onChange={(e) => setDisplayOrder(Number(e.target.value))}
              className="input-field"
              disabled={loading}
              min={0}
              placeholder="0"
            />
            <p className="text-xs text-gray-500 mt-1">小さい数字ほど上に表示されます</p>
          </div>

          <div className="mb-4">
            <label htmlFor="maxMembers" className="label">
              最大人数（任意）
            </label>
            <input
              type="number"
              id="maxMembers"
              value={maxMembers}
              onChange={(e) => setMaxMembers(e.target.value)}
              className="input-field"
              disabled={loading}
              min={1}
              placeholder="未設定"
            />
            <p className="text-xs text-gray-500 mt-1">
              このインスタンスに配置できる最大人数です。未設定の場合は制限なしになります。
            </p>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 btn-secondary"
              disabled={loading}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="flex-1 btn-primary"
              disabled={loading || !name.trim()}
            >
              {loading ? '処理中...' : instance ? '更新' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
