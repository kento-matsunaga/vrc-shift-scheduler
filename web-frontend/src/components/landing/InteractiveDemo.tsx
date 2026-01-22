import { useState, useEffect, useRef } from 'react';

type AttendanceStatus = 'available' | 'maybe' | 'unavailable';

interface ShiftSlot {
  id: string;
  name: string;
  time: string;
  required: number;
  assigned: number;
}

interface Member {
  id: string;
  name: string;
  avatar: string;
  status: AttendanceStatus;
}

// デモ用のモックデータ
const mockShiftSlots: ShiftSlot[] = [
  { id: '1', name: '受付', time: '21:00〜21:30', required: 2, assigned: 1 },
  { id: '2', name: 'MC', time: '21:30〜22:00', required: 1, assigned: 0 },
  { id: '3', name: 'フロア案内', time: '21:00〜23:00', required: 3, assigned: 2 },
];

const mockMembers: Member[] = [
  { id: '1', name: 'Haruka', avatar: 'H', status: 'available' },
  { id: '2', name: 'Yuki', avatar: 'Y', status: 'available' },
  { id: '3', name: 'Sora', avatar: 'S', status: 'maybe' },
  { id: '4', name: 'Ren', avatar: 'R', status: 'unavailable' },
];

const statusColors: Record<AttendanceStatus, { bg: string; text: string; label: string }> = {
  available: { bg: 'bg-emerald-500/20', text: 'text-emerald-400', label: '○' },
  maybe: { bg: 'bg-amber-500/20', text: 'text-amber-400', label: '△' },
  unavailable: { bg: 'bg-red-500/20', text: 'text-red-400', label: '×' },
};

const ANIMATION_INTERVAL_MS = 5000;

export function InteractiveDemo() {
  const [activeTab, setActiveTab] = useState<'shifts' | 'attendance'>('shifts');
  const [highlightedSlot, setHighlightedSlot] = useState<string | null>(null);
  const [assigningMember, setAssigningMember] = useState<string | null>(null);
  const [slots, setSlots] = useState(mockShiftSlots);
  const [showAssignSuccess, setShowAssignSuccess] = useState(false);
  const slotsRef = useRef(slots);

  // slotsの最新値をrefに保持
  useEffect(() => {
    slotsRef.current = slots;
  }, [slots]);

  // 自動アニメーション
  useEffect(() => {
    const interval = setInterval(() => {
      // シフトタブの場合: スロットをハイライト → メンバー割り当てアニメーション
      if (activeTab === 'shifts') {
        const currentSlots = slotsRef.current;
        const unfilledSlot = currentSlots.find(s => s.assigned < s.required);
        if (unfilledSlot) {
          setHighlightedSlot(unfilledSlot.id);
          setTimeout(() => {
            setAssigningMember('2'); // Yukiを割り当て
            setTimeout(() => {
              setSlots(prev => prev.map(s =>
                s.id === unfilledSlot.id ? { ...s, assigned: Math.min(s.assigned + 1, s.required) } : s
              ));
              setShowAssignSuccess(true);
              setTimeout(() => {
                setShowAssignSuccess(false);
                setAssigningMember(null);
                setHighlightedSlot(null);
              }, 1500);
            }, 800);
          }, 1000);
        } else {
          // リセット
          setSlots(mockShiftSlots);
        }
      }
    }, ANIMATION_INTERVAL_MS);

    return () => clearInterval(interval);
  }, [activeTab]);

  return (
    <div className="relative w-full max-w-sm sm:max-w-md">
      {/* ブラウザ風ウィンドウフレーム */}
      <div
        className="rounded-xl overflow-hidden"
        style={{
          background: 'rgba(17, 24, 39, 0.9)',
          border: '1px solid rgba(139, 92, 246, 0.2)',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5), 0 0 60px rgba(139, 92, 246, 0.1)',
          WebkitBackdropFilter: 'blur(10px)',
          backdropFilter: 'blur(10px)',
        }}
      >
        {/* ウィンドウヘッダー */}
        <div
          className="flex items-center gap-2 px-4 py-3"
          style={{ borderBottom: '1px solid rgba(139, 92, 246, 0.15)' }}
        >
          <div className="flex gap-1.5">
            <div className="w-3 h-3 rounded-full bg-red-500/60" />
            <div className="w-3 h-3 rounded-full bg-yellow-500/60" />
            <div className="w-3 h-3 rounded-full bg-green-500/60" />
          </div>
          <div className="flex-1 text-center">
            <span className="text-xs text-gray-500">VRC Shift Scheduler</span>
          </div>
        </div>

        {/* タブ */}
        <div className="flex border-b" style={{ borderColor: 'rgba(139, 92, 246, 0.15)' }}>
          <button
            onClick={() => setActiveTab('shifts')}
            className={`flex-1 px-3 sm:px-4 py-3 text-xs sm:text-sm font-medium transition-colors min-h-[44px] ${
              activeTab === 'shifts'
                ? 'text-violet-400 border-b-2 border-violet-500'
                : 'text-gray-500 hover:text-gray-400 active:text-gray-300'
            }`}
          >
            シフト枠
          </button>
          <button
            onClick={() => setActiveTab('attendance')}
            className={`flex-1 px-3 sm:px-4 py-3 text-xs sm:text-sm font-medium transition-colors min-h-[44px] ${
              activeTab === 'attendance'
                ? 'text-violet-400 border-b-2 border-violet-500'
                : 'text-gray-500 hover:text-gray-400 active:text-gray-300'
            }`}
          >
            出欠状況
          </button>
        </div>

        {/* コンテンツ */}
        <div className="p-3 sm:p-4 min-h-[260px] sm:min-h-[280px]">
          {activeTab === 'shifts' ? (
            <div className="space-y-2 sm:space-y-3">
              {/* イベント情報 */}
              <div className="flex items-center justify-between mb-3 sm:mb-4">
                <div className="min-w-0 flex-1">
                  <h3 className="text-white font-medium text-sm sm:text-base truncate">VRChat 交流会</h3>
                  <p className="text-[10px] sm:text-xs text-gray-500">2025/01/25 (土) 21:00〜</p>
                </div>
                <span className="px-2 py-1 text-[10px] sm:text-xs font-medium rounded bg-violet-500/20 text-violet-400 flex-shrink-0 ml-2">
                  毎週
                </span>
              </div>

              {/* シフト枠リスト */}
              {slots.map((slot) => (
                <div
                  key={slot.id}
                  className={`p-2.5 sm:p-3 rounded-lg transition-all duration-300 ${
                    highlightedSlot === slot.id
                      ? 'ring-2 ring-violet-500 bg-violet-500/10'
                      : 'bg-gray-800/50'
                  }`}
                  style={{ border: '1px solid rgba(75, 85, 99, 0.3)' }}
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <span className="text-white font-medium text-xs sm:text-sm">{slot.name}</span>
                      <span className="text-gray-500 text-[10px] sm:text-xs ml-1 sm:ml-2">{slot.time}</span>
                    </div>
                    <div className="flex items-center flex-shrink-0">
                      <span
                        className={`text-[10px] sm:text-xs font-medium px-1.5 sm:px-2 py-0.5 rounded ${
                          slot.assigned >= slot.required
                            ? 'bg-emerald-500/20 text-emerald-400'
                            : 'bg-amber-500/20 text-amber-400'
                        }`}
                      >
                        {slot.assigned}/{slot.required}
                      </span>
                    </div>
                  </div>

                  {/* アサインアニメーション */}
                  {highlightedSlot === slot.id && assigningMember && (
                    <div className="mt-2 flex items-center gap-2 animate-pulse">
                      <div className="w-5 h-5 sm:w-6 sm:h-6 rounded-full bg-gradient-to-br from-violet-500 to-indigo-500 flex items-center justify-center text-white text-[10px] sm:text-xs font-bold flex-shrink-0">
                        Y
                      </div>
                      <span className="text-violet-400 text-[10px] sm:text-xs">Yuki を割り当て中...</span>
                    </div>
                  )}
                  {showAssignSuccess && highlightedSlot === slot.id && (
                    <div className="mt-2 flex items-center gap-2 text-emerald-400 text-[10px] sm:text-xs">
                      <svg className="w-3.5 h-3.5 sm:w-4 sm:h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                      割り当て完了
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="space-y-2 sm:space-y-3">
              {/* 出欠収集ヘッダー */}
              <div className="flex items-center justify-between mb-3 sm:mb-4 gap-2">
                <div className="min-w-0">
                  <h3 className="text-white font-medium text-sm sm:text-base">出欠状況</h3>
                  <p className="text-[10px] sm:text-xs text-gray-500">回答: 4/6名</p>
                </div>
                <div className="flex gap-1 flex-shrink-0">
                  <span className="px-1.5 sm:px-2 py-0.5 text-[10px] sm:text-xs rounded bg-emerald-500/20 text-emerald-400">○2</span>
                  <span className="px-1.5 sm:px-2 py-0.5 text-[10px] sm:text-xs rounded bg-amber-500/20 text-amber-400">△1</span>
                  <span className="px-1.5 sm:px-2 py-0.5 text-[10px] sm:text-xs rounded bg-red-500/20 text-red-400">×1</span>
                </div>
              </div>

              {/* メンバーリスト */}
              {mockMembers.map((member) => (
                <div
                  key={member.id}
                  className="flex items-center justify-between p-2.5 sm:p-3 rounded-lg bg-gray-800/50"
                  style={{ border: '1px solid rgba(75, 85, 99, 0.3)' }}
                >
                  <div className="flex items-center gap-2 sm:gap-3 min-w-0">
                    <div
                      className="w-7 h-7 sm:w-8 sm:h-8 rounded-full flex items-center justify-center text-white text-xs sm:text-sm font-bold flex-shrink-0"
                      style={{ background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)' }}
                    >
                      {member.avatar}
                    </div>
                    <span className="text-white text-xs sm:text-sm truncate">{member.name}</span>
                  </div>
                  <span
                    className={`w-7 h-7 sm:w-8 sm:h-8 rounded-lg flex items-center justify-center text-base sm:text-lg font-bold flex-shrink-0 ${
                      statusColors[member.status].bg
                    } ${statusColors[member.status].text}`}
                  >
                    {statusColors[member.status].label}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* 装飾: グロー効果 */}
      <div
        className="absolute -z-10 top-1/2 left-1/2 w-[120%] h-[120%] -translate-x-1/2 -translate-y-1/2"
        style={{
          background: 'radial-gradient(circle, rgba(139, 92, 246, 0.15) 0%, transparent 70%)',
          filter: 'blur(60px)',
        }}
      />
    </div>
  );
}
