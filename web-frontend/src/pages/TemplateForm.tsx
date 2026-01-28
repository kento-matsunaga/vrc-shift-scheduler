import { useState, useEffect, useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { SEO } from '../components/seo';
import { getTemplate, createTemplate, updateTemplate } from '../lib/api/templateApi';
import { listInstances, type Instance } from '../lib/api/instanceApi';
import type { TemplateItem, CreateTemplateRequest } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

// インスタンスごとに役職をグループ化した構造
interface InstanceGroup {
  instanceName: string;
  isNew: boolean; // 既存のInstanceか新規作成か
  roles: RoleItem[];
}

interface RoleItem {
  slotName: string;
  startTime: string;
  endTime: string;
  requiredCount: number;
  priority: number;
}

export default function TemplateForm() {
  const { eventId, templateId } = useParams<{ eventId: string; templateId?: string }>();
  const navigate = useNavigate();
  const isEditMode = !!templateId;

  const [templateName, setTemplateName] = useState('');
  const [description, setDescription] = useState('');
  const [instanceGroups, setInstanceGroups] = useState<InstanceGroup[]>([]);
  const [existingInstances, setExistingInstances] = useState<Instance[]>([]);
  const [expandedInstances, setExpandedInstances] = useState<Set<number>>(new Set([0]));
  const [loading, setLoading] = useState(false);
  const [loadingData, setLoadingData] = useState(true);
  const [error, setError] = useState('');

  // 新規インスタンス追加用のモーダル状態
  const [showInstanceModal, setShowInstanceModal] = useState(false);
  const [newInstanceName, setNewInstanceName] = useState('');
  const [selectedExistingInstance, setSelectedExistingInstance] = useState('');

  useEffect(() => {
    loadData();
  }, [eventId, templateId, isEditMode]);

  const loadData = async () => {
    if (!eventId) return;

    try {
      setLoadingData(true);

      // 既存のインスタンスを取得
      const instances = await listInstances(eventId);
      setExistingInstances(instances);

      // 編集モードの場合はテンプレートも取得
      if (isEditMode && templateId) {
        const template = await getTemplate(eventId, templateId);
        setTemplateName(template.template_name);
        setDescription(template.description);

        // テンプレートアイテムをインスタンスグループに変換
        const groups = convertItemsToGroups(template.items || [], instances);
        setInstanceGroups(groups);
        // 全て展開
        setExpandedInstances(new Set(groups.map((_, i) => i)));
      }
    } catch (err) {
      console.error('Failed to load data:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('データの読み込みに失敗しました');
      }
    } finally {
      setLoadingData(false);
    }
  };

  // TemplateItemをInstanceGroupに変換
  const convertItemsToGroups = (items: TemplateItem[], instances: Instance[]): InstanceGroup[] => {
    const groupMap = new Map<string, InstanceGroup>();
    const instanceNameSet = new Set(instances.map((i) => i.name));

    items.forEach((item) => {
      if (!groupMap.has(item.instance_name)) {
        groupMap.set(item.instance_name, {
          instanceName: item.instance_name,
          isNew: !instanceNameSet.has(item.instance_name),
          roles: [],
        });
      }
      groupMap.get(item.instance_name)!.roles.push({
        slotName: item.slot_name,
        startTime: item.start_time,
        endTime: item.end_time,
        requiredCount: item.required_count,
        priority: item.priority,
      });
    });

    // 役職をpriority順にソート
    groupMap.forEach((group) => {
      group.roles.sort((a, b) => a.priority - b.priority);
    });

    return Array.from(groupMap.values());
  };

  // InstanceGroupをTemplateItemに変換
  const convertGroupsToItems = (): TemplateItem[] => {
    const items: TemplateItem[] = [];
    instanceGroups.forEach((group) => {
      group.roles.forEach((role) => {
        items.push({
          slot_name: role.slotName,
          instance_name: group.instanceName,
          start_time: role.startTime,
          end_time: role.endTime,
          required_count: role.requiredCount,
          priority: role.priority,
        });
      });
    });
    return items;
  };

  // 使用可能なインスタンス名（まだ追加されていないもの）
  const availableInstances = useMemo(() => {
    const usedNames = new Set(instanceGroups.map((g) => g.instanceName));
    return existingInstances.filter((inst) => !usedNames.has(inst.name));
  }, [existingInstances, instanceGroups]);

  const toggleInstanceExpand = (index: number) => {
    setExpandedInstances((prev) => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  };

  const addInstance = () => {
    setShowInstanceModal(true);
    setNewInstanceName('');
    setSelectedExistingInstance('');
  };

  const confirmAddInstance = () => {
    const instanceName = selectedExistingInstance || newInstanceName.trim();
    if (!instanceName) {
      return;
    }

    // 既に存在するかチェック
    if (instanceGroups.some((g) => g.instanceName === instanceName)) {
      setError('同じ名前のインスタンスが既に追加されています');
      return;
    }

    const isNew = !existingInstances.some((i) => i.name === instanceName);
    const newGroup: InstanceGroup = {
      instanceName,
      isNew,
      roles: [
        {
          slotName: '',
          startTime: '21:00:00',
          endTime: '23:00:00',
          requiredCount: 1,
          priority: 1,
        },
      ],
    };

    setInstanceGroups([...instanceGroups, newGroup]);
    setExpandedInstances((prev) => new Set([...prev, instanceGroups.length]));
    setShowInstanceModal(false);
    setError('');
  };

  const removeInstance = (index: number) => {
    setInstanceGroups(instanceGroups.filter((_, i) => i !== index));
    setExpandedInstances((prev) => {
      const next = new Set<number>();
      prev.forEach((i) => {
        if (i < index) next.add(i);
        else if (i > index) next.add(i - 1);
      });
      return next;
    });
  };

  const addRole = (instanceIndex: number) => {
    const newGroups = [...instanceGroups];
    const maxPriority = Math.max(0, ...newGroups[instanceIndex].roles.map((r) => r.priority));
    newGroups[instanceIndex].roles.push({
      slotName: '',
      startTime: '21:00:00',
      endTime: '23:00:00',
      requiredCount: 1,
      priority: maxPriority + 1,
    });
    setInstanceGroups(newGroups);
  };

  const removeRole = (instanceIndex: number, roleIndex: number) => {
    const newGroups = [...instanceGroups];
    newGroups[instanceIndex].roles = newGroups[instanceIndex].roles.filter((_, i) => i !== roleIndex);
    setInstanceGroups(newGroups);
  };

  const updateRole = (
    instanceIndex: number,
    roleIndex: number,
    field: keyof RoleItem,
    value: string | number
  ) => {
    const newGroups = [...instanceGroups];
    newGroups[instanceIndex].roles[roleIndex] = {
      ...newGroups[instanceIndex].roles[roleIndex],
      [field]: value,
    };
    setInstanceGroups(newGroups);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!eventId) {
      setError('イベントIDが指定されていません');
      return;
    }

    if (!templateName.trim()) {
      setError('テンプレート名を入力してください');
      return;
    }

    if (instanceGroups.length === 0) {
      setError('少なくとも1つのインスタンスを追加してください');
      return;
    }

    // バリデーション
    for (let i = 0; i < instanceGroups.length; i++) {
      const group = instanceGroups[i];
      if (group.roles.length === 0) {
        setError(`インスタンス「${group.instanceName}」に役職を追加してください`);
        return;
      }
      for (let j = 0; j < group.roles.length; j++) {
        const role = group.roles[j];
        if (!role.slotName.trim()) {
          setError(`インスタンス「${group.instanceName}」の役職 ${j + 1}: 役職名を入力してください`);
          return;
        }
        if (role.requiredCount < 1) {
          setError(`インスタンス「${group.instanceName}」の役職「${role.slotName}」: 必要人数は1以上にしてください`);
          return;
        }
      }
    }

    setLoading(true);

    try {
      const items = convertGroupsToItems();
      const data: CreateTemplateRequest = {
        template_name: templateName.trim(),
        description: description.trim(),
        items: items.map((item) => ({
          slot_name: item.slot_name.trim(),
          instance_name: item.instance_name.trim(),
          start_time: item.start_time,
          end_time: item.end_time,
          required_count: item.required_count,
          priority: item.priority,
        })),
      };

      if (isEditMode && templateId) {
        await updateTemplate(eventId, templateId, data);
      } else {
        await createTemplate(eventId, data);
      }

      navigate(`/events/${eventId}/templates`);
    } catch (err) {
      console.error('Failed to save template:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError(isEditMode ? 'テンプレートの更新に失敗しました' : 'テンプレートの作成に失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  if (loadingData) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
        <p className="mt-4 text-gray-600">読み込み中...</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <SEO noindex={true} />
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">
          {isEditMode ? 'テンプレートを編集' : '新規テンプレート作成'}
        </h2>
        <p className="text-sm text-gray-600 mt-1">
          営業日作成時に使用するシフト枠のテンプレートを{isEditMode ? '編集' : '作成'}します
        </p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">基本情報</h3>

          <div className="mb-4">
            <label htmlFor="templateName" className="label">
              テンプレート名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="templateName"
              value={templateName}
              onChange={(e) => setTemplateName(e.target.value)}
              className="input-field"
              disabled={loading}
              placeholder="例: 土曜日テンプレート、通常営業日テンプレート"
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label htmlFor="description" className="label">
              説明
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="input-field"
              disabled={loading}
              rows={3}
              placeholder="このテンプレートの用途や特徴を入力してください"
            />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex justify-between items-center mb-4">
            <div>
              <h3 className="text-lg font-semibold text-gray-900">インスタンス・役職設定</h3>
              <p className="text-sm text-gray-500 mt-1">
                インスタンスを追加し、各インスタンス内に役職（シフト枠）を設定します
              </p>
            </div>
            <button
              type="button"
              onClick={addInstance}
              disabled={loading}
              className="bg-accent hover:bg-accent-dark text-white px-4 py-2 rounded-lg text-sm flex items-center"
            >
              <svg
                className="w-5 h-5 mr-1"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 4v16m8-8H4"
                />
              </svg>
              インスタンスを追加
            </button>
          </div>

          {instanceGroups.length === 0 ? (
            <div className="text-center py-8 bg-gray-50 rounded-lg">
              <p className="text-gray-600 mb-4">まだインスタンスが追加されていません</p>
              <button
                type="button"
                onClick={addInstance}
                disabled={loading}
                className="bg-accent hover:bg-accent-dark text-white px-4 py-2 rounded-lg text-sm"
              >
                最初のインスタンスを追加
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {instanceGroups.map((group, instanceIndex) => (
                <div
                  key={instanceIndex}
                  className="border border-gray-200 rounded-lg overflow-hidden"
                >
                  {/* インスタンスヘッダー */}
                  <div
                    className="bg-gray-100 px-4 py-3 flex justify-between items-center cursor-pointer"
                    onClick={() => toggleInstanceExpand(instanceIndex)}
                  >
                    <div className="flex items-center gap-3">
                      <svg
                        className={`w-5 h-5 text-gray-500 transition-transform ${
                          expandedInstances.has(instanceIndex) ? 'rotate-90' : ''
                        }`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 5l7 7-7 7"
                        />
                      </svg>
                      <span className="font-semibold text-gray-900">{group.instanceName}</span>
                      {group.isNew && (
                        <span className="bg-blue-100 text-blue-800 text-xs px-2 py-0.5 rounded">
                          新規
                        </span>
                      )}
                      <span className="text-sm text-gray-500">
                        ({group.roles.length}役職)
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        removeInstance(instanceIndex);
                      }}
                      disabled={loading}
                      className="text-red-600 hover:text-red-800 text-sm px-2 py-1"
                    >
                      削除
                    </button>
                  </div>

                  {/* 役職リスト */}
                  {expandedInstances.has(instanceIndex) && (
                    <div className="p-4 bg-white">
                      {group.roles.length === 0 ? (
                        <div className="text-center py-4 text-gray-500">
                          役職がありません
                        </div>
                      ) : (
                        <div className="space-y-3">
                          {group.roles.map((role, roleIndex) => (
                            <div
                              key={roleIndex}
                              className="border border-gray-200 rounded-lg p-4 bg-gray-50"
                            >
                              <div className="flex justify-between items-start mb-3">
                                <h5 className="font-medium text-gray-700">
                                  役職 {roleIndex + 1}
                                </h5>
                                <button
                                  type="button"
                                  onClick={() => removeRole(instanceIndex, roleIndex)}
                                  disabled={loading}
                                  className="text-red-600 hover:text-red-800 text-sm"
                                >
                                  削除
                                </button>
                              </div>

                              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                <div className="md:col-span-2 lg:col-span-1">
                                  <label className="label text-sm">
                                    役職名 <span className="text-red-500">*</span>
                                  </label>
                                  <input
                                    type="text"
                                    value={role.slotName}
                                    onChange={(e) =>
                                      updateRole(instanceIndex, roleIndex, 'slotName', e.target.value)
                                    }
                                    className="input-field text-sm"
                                    disabled={loading}
                                    placeholder="例: 受付、案内、配信、DJ"
                                  />
                                </div>

                                <div>
                                  <label className="label text-sm">
                                    開始時刻 <span className="text-red-500">*</span>
                                  </label>
                                  <input
                                    type="time"
                                    value={role.startTime.substring(0, 5)}
                                    onChange={(e) =>
                                      updateRole(
                                        instanceIndex,
                                        roleIndex,
                                        'startTime',
                                        e.target.value + ':00'
                                      )
                                    }
                                    className="input-field text-sm"
                                    disabled={loading}
                                  />
                                </div>

                                <div>
                                  <label className="label text-sm">
                                    終了時刻 <span className="text-red-500">*</span>
                                  </label>
                                  <input
                                    type="time"
                                    value={role.endTime.substring(0, 5)}
                                    onChange={(e) =>
                                      updateRole(
                                        instanceIndex,
                                        roleIndex,
                                        'endTime',
                                        e.target.value + ':00'
                                      )
                                    }
                                    className="input-field text-sm"
                                    disabled={loading}
                                  />
                                </div>

                                <div>
                                  <label className="label text-sm">
                                    必要人数 <span className="text-red-500">*</span>
                                  </label>
                                  <input
                                    type="number"
                                    value={role.requiredCount}
                                    onChange={(e) =>
                                      updateRole(
                                        instanceIndex,
                                        roleIndex,
                                        'requiredCount',
                                        Number(e.target.value)
                                      )
                                    }
                                    className="input-field text-sm"
                                    disabled={loading}
                                    min="1"
                                  />
                                </div>

                                <div>
                                  <label className="label text-sm">優先度</label>
                                  <input
                                    type="number"
                                    value={role.priority}
                                    onChange={(e) =>
                                      updateRole(
                                        instanceIndex,
                                        roleIndex,
                                        'priority',
                                        Number(e.target.value)
                                      )
                                    }
                                    className="input-field text-sm"
                                    disabled={loading}
                                    min="1"
                                  />
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      )}

                      <button
                        type="button"
                        onClick={() => addRole(instanceIndex)}
                        disabled={loading}
                        className="mt-3 w-full py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-accent hover:text-accent transition-colors text-sm"
                      >
                        + 役職を追加
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex gap-4">
          <button
            type="button"
            onClick={() => navigate(`/events/${eventId}/templates`)}
            disabled={loading}
            className="flex-1 bg-gray-200 hover:bg-gray-300 text-gray-700 px-6 py-3 rounded-lg font-medium"
          >
            キャンセル
          </button>
          <button
            type="submit"
            disabled={loading || instanceGroups.length === 0}
            className="flex-1 bg-accent hover:bg-accent-dark text-white px-6 py-3 rounded-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '保存中...' : isEditMode ? '更新する' : '作成する'}
          </button>
        </div>
      </form>

      {/* インスタンス追加モーダル */}
      {showInstanceModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">インスタンスを追加</h3>

            {availableInstances.length > 0 && (
              <div className="mb-4">
                <label className="label">既存のインスタンスから選択</label>
                <select
                  value={selectedExistingInstance}
                  onChange={(e) => {
                    setSelectedExistingInstance(e.target.value);
                    if (e.target.value) {
                      setNewInstanceName('');
                    }
                  }}
                  className="input-field"
                >
                  <option value="">選択してください</option>
                  {availableInstances.map((inst) => (
                    <option key={inst.instance_id} value={inst.name}>
                      {inst.name}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <div className="mb-4">
              <label className="label">
                {availableInstances.length > 0 ? 'または新しいインスタンス名を入力' : '新しいインスタンス名'}
              </label>
              <input
                type="text"
                value={newInstanceName}
                onChange={(e) => {
                  setNewInstanceName(e.target.value);
                  if (e.target.value) {
                    setSelectedExistingInstance('');
                  }
                }}
                className="input-field"
                placeholder="例: 受付1、配信1、メインホール"
              />
            </div>

            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => setShowInstanceModal(false)}
                className="flex-1 bg-gray-200 hover:bg-gray-300 text-gray-700 px-4 py-2 rounded-lg"
              >
                キャンセル
              </button>
              <button
                type="button"
                onClick={confirmAddInstance}
                disabled={!selectedExistingInstance && !newInstanceName.trim()}
                className="flex-1 bg-accent hover:bg-accent-dark text-white px-4 py-2 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
              >
                追加
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
