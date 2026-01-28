# vitest-react-setup

Vite + React プロジェクトに vitest + testing-library を導入するセットアップスキル。

## 前提条件

- Vite ビルドツール導入済み
- React 18/19
- @vitejs/plugin-react 導入済み

## 実行手順

### 1. 依存パッケージのインストール

```bash
npm install -D vitest @testing-library/react @testing-library/dom jsdom @vitest/ui
```

### 2. vitest.config.ts の作成

プロジェクトルートに以下の内容で作成：

```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{js,ts,jsx,tsx}'],
  },
})
```

### 3. src/test/setup.ts の作成

```typescript
import '@testing-library/dom'
```

### 4. package.json への scripts 追加（任意）

```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest --coverage"
  }
}
```

### 5. テストファイルのテンプレート

`src/components/Example/Example.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { Example } from './Example'

describe('Example', () => {
  it('renders correctly', () => {
    render(<Example />)
    expect(screen.getByText('Example')).toBeInTheDocument()
  })
})
```

## カスタマイズオプション

| オプション | 説明 | デフォルト |
|------------|------|-----------|
| environment | DOM環境 | jsdom |
| setupFiles | セットアップファイル | ./src/test/setup.ts |
| include | テストファイルパターン | src/**/*.{test,spec}.{js,ts,jsx,tsx} |
| globals | グローバル変数（describe, it, expect） | true |

## 依存パッケージ一覧

| パッケージ | バージョン | 用途 |
|------------|-----------|------|
| vitest | ^4.0.0 | テストランナー |
| @testing-library/react | ^16.0.0 | React テストユーティリティ |
| @testing-library/dom | ^10.0.0 | DOM テストユーティリティ |
| jsdom | ^26.0.0 | DOM シミュレート |
| @vitest/ui | ^4.0.0 | UI モード（任意） |
