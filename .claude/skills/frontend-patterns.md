---
description: React/TypeScript/Tailwind CSS開発パターンとベストプラクティス
---

# Frontend Patterns

VRC Shift Scheduler のフロントエンド開発パターン。

---

## 技術スタック

- **React 19** + **TypeScript 5.9**
- **Vite 7** (ビルドツール)
- **Tailwind CSS 4** (スタイリング)
- **React Router** (ルーティング)

---

## ディレクトリ構成

```
web-frontend/
├── src/
│   ├── components/      # 再利用可能なUIコンポーネント
│   │   ├── common/      # 汎用コンポーネント（Button, Modal等）
│   │   └── features/    # 機能別コンポーネント
│   ├── hooks/           # カスタムフック
│   ├── pages/           # ページコンポーネント
│   ├── services/        # API通信
│   ├── types/           # TypeScript型定義
│   ├── utils/           # ユーティリティ関数
│   └── contexts/        # React Context
├── public/              # 静的ファイル
└── tests/               # テストファイル

admin-frontend/          # 管理者向けフロントエンド（同様の構成）
```

---

## コンポーネント設計

### ファイル命名規則

```
PascalCase.tsx     # コンポーネント
camelCase.ts       # ユーティリティ、フック
types.ts           # 型定義
index.ts           # エクスポート
```

### コンポーネント構成

```tsx
// components/features/EventList/EventList.tsx
import { useState, useEffect } from 'react';
import type { Event } from '@/types/event';

interface EventListProps {
  tenantId: string;
  onEventSelect?: (event: Event) => void;
}

export function EventList({ tenantId, onEventSelect }: EventListProps) {
  const [events, setEvents] = useState<Event[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // ...実装
}
```

### Props設計

```tsx
// 必須プロパティと任意プロパティを明確に
interface ButtonProps {
  // 必須
  children: React.ReactNode;
  onClick: () => void;

  // 任意（デフォルト値あり）
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  isLoading?: boolean;
}
```

---

## 状態管理パターン

### ローカル状態（useState）

単一コンポーネント内の状態:

```tsx
const [isOpen, setIsOpen] = useState(false);
const [formData, setFormData] = useState<FormData>({ name: '', email: '' });
```

### 複雑な状態（useReducer）

複数の関連する状態変更:

```tsx
type Action =
  | { type: 'SET_LOADING' }
  | { type: 'SET_DATA'; payload: Event[] }
  | { type: 'SET_ERROR'; payload: string };

function eventReducer(state: State, action: Action): State {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: true, error: null };
    case 'SET_DATA':
      return { ...state, isLoading: false, events: action.payload };
    case 'SET_ERROR':
      return { ...state, isLoading: false, error: action.payload };
  }
}
```

### グローバル状態（Context）

認証情報など複数コンポーネントで共有する状態:

```tsx
// contexts/AuthContext.tsx
interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
```

---

## カスタムフック

### データフェッチフック

```tsx
// hooks/useEvents.ts
export function useEvents(tenantId: string) {
  const [events, setEvents] = useState<Event[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        setIsLoading(true);
        const response = await eventService.getEvents(tenantId);
        setEvents(response.data);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'));
      } finally {
        setIsLoading(false);
      }
    };

    fetchEvents();
  }, [tenantId]);

  return { events, isLoading, error, refetch: () => { /* ... */ } };
}
```

### フォームフック

```tsx
// hooks/useForm.ts
export function useForm<T>(initialValues: T) {
  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({});

  const handleChange = (field: keyof T, value: T[keyof T]) => {
    setValues(prev => ({ ...prev, [field]: value }));
    // エラーをクリア
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const reset = () => {
    setValues(initialValues);
    setErrors({});
  };

  return { values, errors, handleChange, setErrors, reset };
}
```

---

## API通信パターン

### サービス層

```tsx
// services/eventService.ts
import { api } from './api';
import type { Event, CreateEventRequest } from '@/types/event';

export const eventService = {
  getEvents: (tenantId: string) =>
    api.get<{ data: Event[] }>(`/tenants/${tenantId}/events`),

  getEvent: (eventId: string) =>
    api.get<{ data: Event }>(`/events/${eventId}`),

  createEvent: (data: CreateEventRequest) =>
    api.post<{ data: Event }>('/events', data),

  updateEvent: (eventId: string, data: Partial<Event>) =>
    api.put<{ data: Event }>(`/events/${eventId}`, data),

  deleteEvent: (eventId: string) =>
    api.delete(`/events/${eventId}`),
};
```

### APIクライアント

```tsx
// services/api.ts
const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

async function request<T>(url: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token');

  const response = await fetch(`${BASE_URL}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error?.message || 'API Error');
  }

  return response.json();
}

export const api = {
  get: <T>(url: string) => request<T>(url),
  post: <T>(url: string, data: unknown) =>
    request<T>(url, { method: 'POST', body: JSON.stringify(data) }),
  put: <T>(url: string, data: unknown) =>
    request<T>(url, { method: 'PUT', body: JSON.stringify(data) }),
  delete: <T>(url: string) =>
    request<T>(url, { method: 'DELETE' }),
};
```

---

## Tailwind CSS パターン

### コンポーネントスタイル

```tsx
// クラス名は読みやすく整理
<button
  className={`
    px-4 py-2 rounded-md font-medium
    ${variant === 'primary'
      ? 'bg-blue-600 text-white hover:bg-blue-700'
      : 'bg-gray-200 text-gray-800 hover:bg-gray-300'
    }
    ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
    transition-colors duration-200
  `}
>
  {children}
</button>
```

### レスポンシブデザイン

```tsx
<div className="
  grid grid-cols-1
  md:grid-cols-2
  lg:grid-cols-3
  gap-4 p-4
">
  {/* コンテンツ */}
</div>
```

### ダークモード対応

```tsx
<div className="
  bg-white dark:bg-gray-800
  text-gray-900 dark:text-gray-100
">
  {/* コンテンツ */}
</div>
```

---

## エラーハンドリング

### ErrorBoundary

```tsx
// components/common/ErrorBoundary.tsx
class ErrorBoundary extends React.Component<Props, State> {
  state = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error caught:', error, errorInfo);
    // エラー報告サービスに送信
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />;
    }
    return this.props.children;
  }
}
```

### APIエラー表示

```tsx
function EventList() {
  const { events, isLoading, error } = useEvents(tenantId);

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage message={error.message} />;
  if (events.length === 0) return <EmptyState message="イベントがありません" />;

  return (
    <ul>
      {events.map(event => <EventItem key={event.id} event={event} />)}
    </ul>
  );
}
```

---

## テストパターン

### コンポーネントテスト

```tsx
// components/EventList.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import { EventList } from './EventList';

describe('EventList', () => {
  it('イベント一覧を表示する', async () => {
    render(<EventList tenantId="tenant-1" />);

    await waitFor(() => {
      expect(screen.getByText('テストイベント')).toBeInTheDocument();
    });
  });

  it('ローディング中はスピナーを表示する', () => {
    render(<EventList tenantId="tenant-1" />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });
});
```

---

## 禁止事項

1. **any型の使用禁止** - 適切な型定義を使用
2. **console.logの本番残留禁止** - デバッグ後は削除
3. **インラインスタイル禁止** - Tailwind CSSを使用
4. **直接DOM操作禁止** - Reactの仮想DOMを使用
5. **useEffectでのデータ変更禁止** - イベントハンドラで処理

---

## 関連ファイル

- `web-frontend/src/` - テナント向けフロントエンド
- `admin-frontend/src/` - 管理者向けフロントエンド
- `docs/requirements/02_機能要件.md` - 機能仕様
