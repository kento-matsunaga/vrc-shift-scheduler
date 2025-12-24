import type { ReactNode } from 'react';

interface MobileCardProps {
  children: ReactNode;
  onClick?: () => void;
  className?: string;
}

// モバイル用カードコンポーネント
export function MobileCard({ children, onClick, className = '' }: MobileCardProps) {
  const baseClass = 'bg-white rounded-lg shadow-sm border border-gray-200 p-4';
  const clickableClass = onClick ? 'cursor-pointer hover:shadow-md transition-shadow' : '';

  return (
    <div
      className={`${baseClass} ${clickableClass} ${className}`}
      onClick={onClick}
    >
      {children}
    </div>
  );
}

interface CardFieldProps {
  label: string;
  value: ReactNode;
  className?: string;
}

// カード内のフィールド表示
export function CardField({ label, value, className = '' }: CardFieldProps) {
  return (
    <div className={`flex justify-between items-start py-1 ${className}`}>
      <span className="text-sm text-gray-500 flex-shrink-0">{label}</span>
      <span className="text-sm text-gray-900 text-right ml-2">{value}</span>
    </div>
  );
}

interface CardHeaderProps {
  title: string;
  subtitle?: string;
  badge?: ReactNode;
  actions?: ReactNode;
}

// カードヘッダー
export function CardHeader({ title, subtitle, badge, actions }: CardHeaderProps) {
  return (
    <div className="flex items-start justify-between mb-3">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <h3 className="font-medium text-gray-900 truncate">{title}</h3>
          {badge}
        </div>
        {subtitle && (
          <p className="text-sm text-gray-500 mt-0.5 truncate">{subtitle}</p>
        )}
      </div>
      {actions && (
        <div className="flex-shrink-0 ml-2">
          {actions}
        </div>
      )}
    </div>
  );
}

interface CardActionsProps {
  children: ReactNode;
}

// カードアクション（ボタン群）
export function CardActions({ children }: CardActionsProps) {
  return (
    <div className="flex gap-2 mt-3 pt-3 border-t border-gray-100">
      {children}
    </div>
  );
}
