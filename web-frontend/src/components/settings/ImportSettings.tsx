import { Upload } from 'lucide-react';
import BulkImport from '../BulkImport';

export function ImportSettings() {
  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
          <Upload className="w-5 h-5 text-accent" />
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
