import BulkImport from '../BulkImport';

// SVG Icon
const UploadIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
  </svg>
);

export function ImportSettings() {
  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
          <UploadIcon className="w-5 h-5 text-accent" />
          データ取込
        </h2>

        <p className="text-sm text-gray-500 mb-6">
          CSVファイルからメンバーデータを一括インポートします。
          テンプレートをダウンロードして、必要な情報を入力してからアップロードしてください。
        </p>

        {/* 既存のBulkImportコンポーネントを使用 */}
        <BulkImport />
      </div>
    </div>
  );
}
