import { useState } from 'react';
import { useOnboardingContext } from '../OnboardingContext';

export function StartTutorialButton() {
  const { state, startTour, stopTour } = useOnboardingContext();
  const [showConfirm, setShowConfirm] = useState(false);

  if (state.isActive) {
    return (
      <button
        onClick={() => setShowConfirm(true)}
        className="px-3 py-1.5 text-xs font-medium text-white bg-red-500 hover:bg-red-600 rounded-lg transition-colors"
        title="チュートリアルを中断"
        id="btn-stop-tutorial"
      >
        チュートリアル中断
      </button>
    );
  }

  return (
    <>
      <button
        onClick={() => setShowConfirm(true)}
        className="px-3 py-1.5 text-xs font-medium text-accent hover:text-accent-dark bg-accent/10 hover:bg-accent/20 rounded-lg transition-colors"
        title="インタラクティブチュートリアルを開始"
        id="btn-start-tutorial"
      >
        体験ツアー
      </button>

      {showConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[9999]">
          <div className="bg-white rounded-lg max-w-sm w-full p-6">
            {state.isActive ? (
              <>
                <h3 className="text-lg font-bold text-gray-900 mb-3">チュートリアルを中断しますか？</h3>
                <p className="text-sm text-gray-600 mb-4">
                  進行状況はリセットされます。ダミーデータも削除されます。
                </p>
                <div className="flex gap-3">
                  <button
                    onClick={() => setShowConfirm(false)}
                    className="flex-1 btn-secondary text-sm"
                  >
                    続ける
                  </button>
                  <button
                    onClick={async () => {
                      await stopTour();
                      setShowConfirm(false);
                      window.location.reload();
                    }}
                    className="flex-1 btn-danger text-sm"
                  >
                    中断する
                  </button>
                </div>
              </>
            ) : (
              <>
                <h3 className="text-lg font-bold text-gray-900 mb-3">体験ツアーを開始</h3>
                <p className="text-sm text-gray-600 mb-2">
                  画面上のガイドに従って、全機能を体験できます。
                </p>
                <ul className="text-sm text-gray-600 mb-4 space-y-1">
                  <li>- ダミーデータを使うので実データに影響はありません</li>
                  <li>- 途中で中断・再開できます</li>
                  <li>- 所要時間: 約5-10分</li>
                </ul>
                <div className="flex gap-3">
                  <button
                    onClick={() => setShowConfirm(false)}
                    className="flex-1 btn-secondary text-sm"
                  >
                    キャンセル
                  </button>
                  <button
                    onClick={async () => {
                      setShowConfirm(false);
                      await startTour();
                    }}
                    className="flex-1 btn-primary text-sm"
                  >
                    開始する
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
}
