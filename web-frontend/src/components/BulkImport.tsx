import { useState, useRef, useCallback } from 'react';
import {
  importMembersFromCSV,
  getImportJobs,
  getImportResult,
  downloadCSVTemplate,
  type ImportMembersResponse,
  type ImportStatusResponse,
  type ImportResultResponse,
  type ImportError,
} from '../lib/api/importApi';
import { ApiClientError } from '../lib/apiClient';

type ImportStep = 'upload' | 'importing' | 'result';

interface ImportHistory {
  jobs: ImportStatusResponse[];
  total_count: number;
}

export default function BulkImport() {
  // File upload state
  const [step, setStep] = useState<ImportStep>('upload');
  const [file, setFile] = useState<File | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Options state
  const [skipExisting, setSkipExisting] = useState(true);
  const [updateExisting, setUpdateExisting] = useState(false);
  const [fuzzyMatch, setFuzzyMatch] = useState(false);

  // Import state
  const [importing, setImporting] = useState(false);
  const [importResult, setImportResult] = useState<ImportMembersResponse | null>(null);
  const [error, setError] = useState('');

  // History state
  const [history, setHistory] = useState<ImportHistory | null>(null);
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [selectedJob, setSelectedJob] = useState<ImportResultResponse | null>(null);
  const [loadingJobDetail, setLoadingJobDetail] = useState(false);

  // Drag and drop handlers
  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const droppedFile = e.dataTransfer.files[0];
      if (droppedFile.type === 'text/csv' || droppedFile.name.endsWith('.csv')) {
        setFile(droppedFile);
        setError('');
      } else {
        setError('CSVファイルを選択してください');
      }
    }
  }, []);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const selectedFile = e.target.files[0];
      if (selectedFile.type === 'text/csv' || selectedFile.name.endsWith('.csv')) {
        setFile(selectedFile);
        setError('');
      } else {
        setError('CSVファイルを選択してください');
      }
    }
  };

  const handleSelectFile = () => {
    fileInputRef.current?.click();
  };

  const handleRemoveFile = () => {
    setFile(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Import handler
  const handleImport = async () => {
    if (!file) return;

    setImporting(true);
    setError('');
    setStep('importing');

    try {
      const result = await importMembersFromCSV(file, {
        skipExisting,
        updateExisting,
        fuzzyMatch,
      });
      setImportResult(result);
      setStep('result');
    } catch (err) {
      console.error('Import error:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      } else {
        setError('インポートに失敗しました');
      }
      setStep('upload');
    } finally {
      setImporting(false);
    }
  };

  // Reset to upload step
  const handleReset = () => {
    setStep('upload');
    setFile(null);
    setImportResult(null);
    setError('');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Load import history
  const handleLoadHistory = async () => {
    setLoadingHistory(true);
    try {
      const data = await getImportJobs(10, 0);
      setHistory(data);
    } catch (err) {
      console.error('Failed to load history:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      }
    } finally {
      setLoadingHistory(false);
    }
  };

  // Load job detail
  const handleViewJobDetail = async (jobId: string) => {
    setLoadingJobDetail(true);
    try {
      const result = await getImportResult(jobId);
      setSelectedJob(result);
    } catch (err) {
      console.error('Failed to load job detail:', err);
      if (err instanceof ApiClientError) {
        setError(err.getUserMessage());
      }
    } finally {
      setLoadingJobDetail(false);
    }
  };

  // Format date
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Status badge
  const StatusBadge = ({ status }: { status: string }) => {
    const styles: Record<string, string> = {
      pending: 'bg-gray-100 text-gray-700',
      processing: 'bg-blue-100 text-blue-700',
      completed: 'bg-green-100 text-green-700',
      failed: 'bg-red-100 text-red-700',
    };
    const labels: Record<string, string> = {
      pending: '待機中',
      processing: '処理中',
      completed: '完了',
      failed: '失敗',
    };
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${styles[status] || styles.pending}`}>
        {labels[status] || status}
      </span>
    );
  };

  // Error list component
  const ErrorList = ({ errors }: { errors: ImportError[] }) => {
    if (!errors || errors.length === 0) return null;
    return (
      <div className="mt-4">
        <h4 className="text-sm font-medium text-gray-700 mb-2">エラー詳細</h4>
        <div className="bg-red-50 border border-red-200 rounded-lg overflow-hidden">
          <div className="max-h-48 overflow-y-auto">
            {errors.map((err, idx) => (
              <div key={idx} className="px-4 py-2 border-b border-red-100 last:border-0">
                <span className="text-xs font-medium text-red-800">行 {err.row}:</span>
                <span className="text-sm text-red-700 ml-2">{err.message}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="space-y-6">
      {/* Upload Section */}
      <div className="card">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
          </svg>
          メンバー一括取り込み
        </h3>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {step === 'upload' && (
          <>
            {/* File Drop Zone */}
            <div
              className={`relative border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                dragActive
                  ? 'border-accent bg-accent/5'
                  : file
                  ? 'border-green-400 bg-green-50'
                  : 'border-gray-300 hover:border-gray-400'
              }`}
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                onChange={handleFileChange}
                className="hidden"
              />

              {file ? (
                <div className="space-y-3">
                  <div className="flex items-center justify-center gap-2 text-green-700">
                    <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span className="font-medium">{file.name}</span>
                  </div>
                  <p className="text-sm text-gray-500">
                    {(file.size / 1024).toFixed(1)} KB
                  </p>
                  <button
                    onClick={handleRemoveFile}
                    className="text-sm text-red-600 hover:text-red-800"
                  >
                    ファイルを削除
                  </button>
                </div>
              ) : (
                <div className="space-y-3">
                  <svg className="w-12 h-12 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                  </svg>
                  <p className="text-gray-600">
                    CSVファイルをドラッグ＆ドロップ
                  </p>
                  <p className="text-sm text-gray-500">または</p>
                  <button
                    onClick={handleSelectFile}
                    className="btn-secondary"
                  >
                    ファイルを選択
                  </button>
                </div>
              )}
            </div>

            {/* Options */}
            <div className="mt-6 space-y-3">
              <h4 className="text-sm font-medium text-gray-700">オプション</h4>
              <label className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={skipExisting}
                  onChange={(e) => {
                    setSkipExisting(e.target.checked);
                    if (e.target.checked) setUpdateExisting(false);
                  }}
                  className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                />
                <div>
                  <span className="text-sm text-gray-700">既存メンバーをスキップ</span>
                  <p className="text-xs text-gray-500">同じ表示名のメンバーが存在する場合はスキップします</p>
                </div>
              </label>
              <label className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={updateExisting}
                  onChange={(e) => {
                    setUpdateExisting(e.target.checked);
                    if (e.target.checked) setSkipExisting(false);
                  }}
                  className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                />
                <div>
                  <span className="text-sm text-gray-700">既存メンバーを更新</span>
                  <p className="text-xs text-gray-500">同じ表示名のメンバーが存在する場合はスキップせずに成功としてカウントします</p>
                </div>
              </label>
              <label className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={fuzzyMatch}
                  onChange={(e) => setFuzzyMatch(e.target.checked)}
                  className="w-4 h-4 text-accent rounded border-gray-300 focus:ring-accent"
                />
                <div>
                  <span className="text-sm text-gray-700">曖昧一致を有効化</span>
                  <p className="text-xs text-gray-500">カタカナ⇔ひらがな、全角⇔半角を同一視して重複チェックを行います</p>
                </div>
              </label>
            </div>

            {/* Template Download */}
            <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <div className="flex items-start gap-3">
                <svg className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div className="flex-1">
                  <p className="text-sm text-blue-800 mb-2">
                    CSVファイルの形式: <code className="bg-blue-100 px-1 rounded">name,display_name,note</code>
                  </p>
                  <ul className="text-xs text-blue-700 space-y-1 mb-3">
                    <li>- <code>name</code>: 本名（任意）</li>
                    <li>- <code>display_name</code>: 表示名（必須、重複チェックに使用）</li>
                    <li>- <code>note</code>: メモ（任意、例: 期生など）</li>
                  </ul>
                  <button
                    onClick={downloadCSVTemplate}
                    className="text-sm text-blue-600 hover:text-blue-800 underline"
                  >
                    テンプレートをダウンロード
                  </button>
                </div>
              </div>
            </div>

            {/* Import Button */}
            <div className="mt-6">
              <button
                onClick={handleImport}
                disabled={!file || importing}
                className="btn-primary w-full"
              >
                インポートを実行
              </button>
            </div>
          </>
        )}

        {step === 'importing' && (
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent mx-auto"></div>
            <p className="mt-4 text-gray-600">インポート中...</p>
            <p className="text-sm text-gray-500 mt-2">{file?.name}</p>
          </div>
        )}

        {step === 'result' && importResult && (
          <div className="space-y-4">
            {/* Result Summary */}
            <div className={`p-4 rounded-lg ${
              importResult.error_count > 0
                ? 'bg-amber-50 border border-amber-200'
                : 'bg-green-50 border border-green-200'
            }`}>
              <div className="flex items-center gap-2 mb-2">
                {importResult.error_count > 0 ? (
                  <svg className="w-6 h-6 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                ) : (
                  <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                )}
                <span className={`font-medium ${
                  importResult.error_count > 0 ? 'text-amber-800' : 'text-green-800'
                }`}>
                  {importResult.error_count > 0 ? '一部エラーあり' : 'インポート完了'}
                </span>
              </div>
              <div className="grid grid-cols-3 gap-4 mt-4">
                <div className="text-center">
                  <p className="text-2xl font-bold text-gray-900">{importResult.total_rows}</p>
                  <p className="text-sm text-gray-600">総行数</p>
                </div>
                <div className="text-center">
                  <p className="text-2xl font-bold text-green-600">{importResult.success_count}</p>
                  <p className="text-sm text-gray-600">成功</p>
                </div>
                <div className="text-center">
                  <p className="text-2xl font-bold text-red-600">{importResult.error_count}</p>
                  <p className="text-sm text-gray-600">エラー</p>
                </div>
              </div>
            </div>

            {/* Error Details */}
            {importResult.errors && importResult.errors.length > 0 && (
              <ErrorList errors={importResult.errors} />
            )}

            {/* Actions */}
            <div className="flex gap-3">
              <button onClick={handleReset} className="btn-primary flex-1">
                新しいインポート
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Import History Section */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
            <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            インポート履歴
          </h3>
          <button
            onClick={handleLoadHistory}
            disabled={loadingHistory}
            className="btn-secondary text-sm"
          >
            {loadingHistory ? '読み込み中...' : '履歴を表示'}
          </button>
        </div>

        {history && (
          <div className="space-y-3">
            {history.jobs.length === 0 ? (
              <p className="text-gray-500 text-center py-4">インポート履歴がありません</p>
            ) : (
              <>
                {history.jobs.map((job) => (
                  <div
                    key={job.import_job_id}
                    className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border border-gray-200"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-gray-900">{job.file_name}</span>
                        <StatusBadge status={job.status} />
                      </div>
                      <p className="text-sm text-gray-500">
                        {formatDate(job.created_at)} -
                        成功: {job.success_count} / エラー: {job.error_count} / 全{job.total_rows}件
                      </p>
                    </div>
                    <button
                      onClick={() => handleViewJobDetail(job.import_job_id)}
                      disabled={loadingJobDetail}
                      className="text-sm text-accent hover:text-accent/80"
                    >
                      詳細
                    </button>
                  </div>
                ))}
              </>
            )}
          </div>
        )}

        {/* Job Detail Modal */}
        {selectedJob && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg max-w-lg w-full p-6 max-h-[80vh] overflow-y-auto">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-gray-900">インポート詳細</h3>
                <button
                  onClick={() => setSelectedJob(null)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <StatusBadge status={selectedJob.status} />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-gray-500">総行数</p>
                    <p className="text-xl font-bold">{selectedJob.total_rows}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-500">成功</p>
                    <p className="text-xl font-bold text-green-600">{selectedJob.success_count}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-500">スキップ</p>
                    <p className="text-xl font-bold text-gray-600">{selectedJob.skipped_count}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-500">エラー</p>
                    <p className="text-xl font-bold text-red-600">{selectedJob.error_count}</p>
                  </div>
                </div>

                {selectedJob.errors && selectedJob.errors.length > 0 && (
                  <ErrorList errors={selectedJob.errors} />
                )}
              </div>

              <div className="mt-6">
                <button
                  onClick={() => setSelectedJob(null)}
                  className="btn-secondary w-full"
                >
                  閉じる
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
