import { useState, useEffect } from 'react';
import { getAssignments } from '../lib/api';
import type { ShiftAssignment } from '../types/api';
import { ApiClientError } from '../lib/apiClient';

export default function MyShifts() {
  const [assignments, setAssignments] = useState<ShiftAssignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<'upcoming' | 'past'>('upcoming');

  const memberId = localStorage.getItem('member_id');

  useEffect(() => {
    if (memberId) {
      loadAssignments();
    }
  }, [memberId, filter]);

  const loadAssignments = async () => {
    if (!memberId) return;

    try {
      setLoading(true);
      const today = new Date().toISOString().split('T')[0];
      
      const data = await getAssignments({
        member_id: memberId,
        assignment_status: 'confirmed',
        ...(filter === 'upcoming' ? { start_date: today } : { end_date: today }),
      });
      
      setAssignments(data.assignments);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('シフト一覧の取得に失敗しました');
      }
      console.error('Failed to load assignments:', err);
    } finally {
      setLoading(false);
    }
  };

  if (!memberId) {
    return (
      <div className="card text-center py-12">
        <p className="text-gray-600">ログインしてください</p>
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">自分のシフト一覧</h2>
        <div className="flex space-x-2">
          <button
            onClick={() => setFilter('upcoming')}
            className={`px-4 py-2 rounded-lg transition-colors ${
              filter === 'upcoming'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-200 text-gray-800 hover:bg-gray-300'
            }`}
          >
            今後のシフト
          </button>
          <button
            onClick={() => setFilter('past')}
            className={`px-4 py-2 rounded-lg transition-colors ${
              filter === 'past'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-200 text-gray-800 hover:bg-gray-300'
            }`}
          >
            過去のシフト
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {loading ? (
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">読み込み中...</p>
        </div>
      ) : assignments.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600">
            {filter === 'upcoming' ? '今後のシフトはありません' : '過去のシフトはありません'}
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {assignments.map((assignment) => (
            <div key={assignment.assignment_id} className="card">
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 mb-2">
                    <h3 className="text-lg font-bold text-gray-900">
                      {assignment.slot_name || '役職未設定'}
                    </h3>
                    <span
                      className={`inline-block px-2 py-1 text-xs font-semibold rounded ${
                        assignment.assignment_method === 'auto'
                          ? 'bg-purple-100 text-purple-800'
                          : 'bg-blue-100 text-blue-800'
                      }`}
                    >
                      {assignment.assignment_method === 'auto' ? '自動割当' : '手動割当'}
                    </span>
                  </div>
                  <div className="text-sm text-gray-600 space-y-1">
                    <p>
                      📅 {assignment.target_date ? new Date(assignment.target_date).toLocaleDateString('ja-JP', {
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric',
                        weekday: 'short',
                      }) : '日付不明'}
                    </p>
                    <p>
                      🕐 {assignment.start_time ? assignment.start_time.slice(0, 5) : '??:??'} 〜{' '}
                      {assignment.end_time ? assignment.end_time.slice(0, 5) : '??:??'}
                    </p>
                    <p className="text-xs text-gray-500">
                      確定日時: {new Date(assignment.assigned_at).toLocaleString('ja-JP')}
                    </p>
                  </div>
                </div>
                <span className="inline-block px-3 py-1 text-sm font-semibold rounded bg-green-100 text-green-800">
                  確定済み
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

