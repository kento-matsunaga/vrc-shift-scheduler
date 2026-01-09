import { useState, useRef, useEffect } from 'react';

interface Option {
  value: string;
  label: string;
}

interface SearchableSelectProps {
  options: Option[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

export default function SearchableSelect({
  options,
  value,
  onChange,
  placeholder = '選択してください',
  disabled = false,
  className = '',
}: SearchableSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // 選択中のラベルを取得
  const selectedLabel = options.find((opt) => opt.value === value)?.label || '';

  // 検索でフィルタリング（部分一致、ひらがな・カタカナ対応）
  const normalizeForSearch = (str: string): string => {
    // ひらがなをカタカナに変換
    return str
      .toLowerCase()
      .replace(/[\u3041-\u3096]/g, (match) =>
        String.fromCharCode(match.charCodeAt(0) + 0x60)
      );
  };

  const filteredOptions = options.filter((opt) => {
    if (!searchQuery) return true;
    const normalizedLabel = normalizeForSearch(opt.label);
    const normalizedQuery = normalizeForSearch(searchQuery);
    return normalizedLabel.includes(normalizedQuery);
  });

  // 外側クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setSearchQuery('');
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // 開いたときに入力欄にフォーカス
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  const handleSelect = (optionValue: string) => {
    onChange(optionValue);
    setIsOpen(false);
    setSearchQuery('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setIsOpen(false);
      setSearchQuery('');
    } else if (e.key === 'Enter' && filteredOptions.length === 1) {
      handleSelect(filteredOptions[0].value);
    }
  };

  return (
    <div ref={containerRef} className={`relative ${className}`}>
      {/* 選択表示エリア */}
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={`w-full px-3 py-2 text-left border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent bg-white ${
          disabled ? 'bg-gray-100 cursor-not-allowed' : 'cursor-pointer'
        }`}
      >
        <span className={value ? 'text-gray-900' : 'text-gray-400'}>
          {value ? selectedLabel : placeholder}
        </span>
        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">
          {isOpen ? '▲' : '▼'}
        </span>
      </button>

      {/* ドロップダウン */}
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-hidden">
          {/* 検索入力 */}
          <div className="p-2 border-b border-gray-200">
            <input
              ref={inputRef}
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="名前を検索..."
              className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent"
            />
          </div>

          {/* オプションリスト */}
          <div className="max-h-48 overflow-y-auto">
            {filteredOptions.length === 0 ? (
              <div className="px-3 py-2 text-sm text-gray-500">
                該当するメンバーがいません
              </div>
            ) : (
              filteredOptions.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => handleSelect(option.value)}
                  className={`w-full px-3 py-2 text-left text-sm hover:bg-accent/10 ${
                    option.value === value ? 'bg-accent/20 font-medium' : ''
                  }`}
                >
                  {option.label}
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
