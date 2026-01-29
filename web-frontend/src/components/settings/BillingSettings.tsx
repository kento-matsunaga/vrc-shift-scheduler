import { useState, useEffect } from 'react';
import { createBillingPortalSession, getBillingStatus } from '../../lib/api';
import type { BillingStatus } from '../../lib/api/billingApi';
import { ApiClientError } from '../../lib/apiClient';

export function BillingSettings() {
  const [billingStatus, setBillingStatus] = useState<BillingStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [openingPortal, setOpeningPortal] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    loadBillingStatus();
  }, []);

  const loadBillingStatus = async () => {
    try {
      setLoading(true);
      const status = await getBillingStatus();
      setBillingStatus(status);
      setError('');
    } catch (err) {
      console.error('Failed to load billing status:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('課金情報の取得に失敗しました');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleOpenBillingPortal = async () => {
    setOpeningPortal(true);
    setError('');

    try {
      const result = await createBillingPortalSession();
      window.location.href = result.portal_url;
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('課金管理ページを開けませんでした');
      }
      console.error('Failed to open billing portal:', err);
      setOpeningPortal(false);
    }
  };

  if (loading) {
    return (
      <div className="card">
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-accent"></div>
          <span className="ml-2 text-gray-500">読み込み中...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="card">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
          </svg>
          課金管理
        </h3>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {/* 現在のプラン表示 */}
        {billingStatus && (
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">現在のプラン</p>
                <p className="text-lg font-semibold text-gray-900">{billingStatus.plan_name}</p>
              </div>
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                billingStatus.status === 'active'
                  ? 'bg-green-100 text-green-800'
                  : billingStatus.status === 'canceled'
                  ? 'bg-gray-100 text-gray-800'
                  : 'bg-yellow-100 text-yellow-800'
              }`}>
                {billingStatus.status === 'active' ? '有効' :
                 billingStatus.status === 'canceled' ? 'キャンセル済み' :
                 billingStatus.status}
              </span>
            </div>
            {billingStatus.plan_type === 'subscription' && billingStatus.current_period_end && (
              <p className="text-sm text-gray-500 mt-2">
                次回更新日: {billingStatus.current_period_end}
              </p>
            )}
            {billingStatus.plan_type === 'lifetime' && (
              <p className="text-sm text-gray-500 mt-2">
                買い切りプランのため、更新は不要です
              </p>
            )}
          </div>
        )}

        {/* サブスクリプションの場合のみポータルボタンを表示 */}
        {billingStatus?.plan_type === 'subscription' && (
          <>
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
              <div className="flex items-start gap-3">
                <svg className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p className="text-sm text-blue-800">
                  Stripeの課金管理ページでは、お支払い方法の変更、請求履歴の確認、サブスクリプションの解約などが行えます。
                </p>
              </div>
            </div>

            <button
              onClick={handleOpenBillingPortal}
              disabled={openingPortal}
              className="btn-primary flex items-center gap-2"
            >
              {openingPortal ? (
                <>
                  <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  読み込み中...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                  </svg>
                  課金管理ページを開く
                </>
              )}
            </button>
          </>
        )}

        {/* 買い切りプランの場合 */}
        {billingStatus?.plan_type === 'lifetime' && (
          <div className="bg-green-50 border border-green-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <svg className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p className="text-sm text-green-800">
                買い切りプランをご利用中です。追加のお支払いは不要です。
              </p>
            </div>
          </div>
        )}

        {/* プラン情報がない場合 */}
        {!billingStatus && !error && (
          <p className="text-gray-500">課金情報を取得できませんでした</p>
        )}
      </div>
    </div>
  );
}
