# PR #237 コードレビュー解説書

**PR タイトル**: 日程調整・出欠確認の編集機能を追加（更新API/編集UI/候補日削除の確認対応）

---

## 1. 概要

### 1.1 PRの目的

作成済みの日程調整（Schedule）・出欠確認（Attendance Collection）を後から編集できる機能を追加する。

**主な機能**:
- 日程調整（Schedule）の編集API・UI
- 出欠確認（Attendance）の編集API・UI
- 候補日削除時の確認ダイアログ

**背景・課題**:
現状、日程調整や出欠確認を作成した後に内容を修正する手段がありませんでした。
タイトルの誤字や候補日の追加漏れなどに気づいた場合、削除して最初から作り直す必要があり、
既に回答済みのデータも失われてしまう課題がありました。

**編集可能な項目**:
- タイトルの修正
- 説明文の修正
- 締切日の変更
- 候補日の追加
- 候補日の削除（※回答済みデータの扱いに注意）

### 1.2 変更ファイル一覧

PR #237 では 18 ファイルが変更されました（+1238行 / -178行）。

| レイヤー | ファイル | 役割 | 行数変更 |
|---------|---------|------|----------|
| **App層** | `backend/internal/app/attendance/dto.go` | 出欠確認更新のDTO | +20 |
| **App層** | `backend/internal/app/attendance/update_collection_usecase.go` | 出欠確認更新ユースケース | +57 |
| **App層** | `backend/internal/app/schedule/dto.go` | 日程調整更新のDTO | +23 |
| **App層** | `backend/internal/app/schedule/update_schedule_usecase.go` | 日程調整更新ユースケース | +181 |
| **Domain層** | `backend/internal/domain/attendance/collection.go` | 出欠確認ドメインモデル | +34 |
| **Domain層** | `backend/internal/domain/attendance/errors.go` | 出欠確認エラー定義 | +3 |
| **Domain層** | `backend/internal/domain/schedule/errors.go` | 日程調整エラー定義 | +3 |
| **Domain層** | `backend/internal/domain/schedule/schedule.go` | 日程調整ドメインモデル | +38 |
| **Infra層** | `backend/internal/infra/db/schedule_repository.go` | 日程調整リポジトリ | +16 |
| **Interface層** | `backend/internal/interface/rest/attendance_handler.go` | 出欠確認HTTPハンドラー | +80 |
| **Interface層** | `backend/internal/interface/rest/router.go` | ルーティング定義 | +6 |
| **Interface層** | `backend/internal/interface/rest/schedule_handler.go` | 日程調整HTTPハンドラー | +132/-22 |
| **Frontend API** | `web-frontend/src/lib/api/attendanceApi.ts` | 出欠確認APIクライアント | +23 |
| **Frontend API** | `web-frontend/src/lib/api/scheduleApi.ts` | 日程調整APIクライアント | +25 |
| **Frontend API** | `web-frontend/src/lib/api/timeUtils.ts` | 時刻ユーティリティ | +166 |
| **Frontend API** | `web-frontend/src/lib/apiClient.ts` | APIクライアント基盤 | +1 |
| **Frontend Pages** | `web-frontend/src/pages/AttendanceList.tsx` | 出欠確認一覧ページ | +137/-63 |
| **Frontend Pages** | `web-frontend/src/pages/ScheduleList.tsx` | 日程調整一覧ページ | +293/-93 |

### 1.3 アーキテクチャ概要

このプロジェクトは**DDD（Domain-Driven Design）**のレイヤードアーキテクチャを採用しています。

```
┌─────────────────────────────────────────────────────────────────┐
│                    Interface層（REST Handler）                   │
│   HTTPリクエストの受付、入力バリデーション、レスポンス生成        │
│   例: schedule_handler.go, attendance_handler.go                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Application層（Usecase）                      │
│   ビジネスロジックのオーケストレーション、トランザクション管理    │
│   例: update_schedule_usecase.go, update_collection_usecase.go  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Domain層（Entity, Value Object）              │
│   ビジネスルール、不変条件の保護、ドメインエラー                 │
│   例: schedule.go, collection.go, errors.go                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure層（Repository）                │
│   データベースアクセス、外部サービス連携                         │
│   例: schedule_repository.go                                    │
└─────────────────────────────────────────────────────────────────┘
```

**各層の責務**:

| 層 | 責務 | この PR での役割 |
|----|------|-----------------|
| Interface | HTTP通信の処理 | PUT エンドポイントの追加 |
| Application | ユースケースの実行 | 更新ロジックの調整 |
| Domain | ビジネスルールの保護 | 更新可能条件のバリデーション |
| Infrastructure | データ永続化 | SQLクエリの実行 |

---

## 2. レビュー結果

### 2.1 発見した問題点の概要

全4層のコードレビューで、以下の問題点が発見されました。

| # | カテゴリ | 深刻度 | ファイル | 概要 |
|---|---------|--------|---------|------|
| 1 | エラーハンドリング | 中 | `update_collection_usecase.go:23-35` | エラーがラップされていない |
| 2 | エラーハンドリング | 中 | `update_schedule_usecase.go:24-37` | エラーがラップされていない |
| 3 | 命名 | 低 | `update_schedule_usecase.go:121-143` | 関数名と戻り値の型が不一致 |
| 4 | ドメイン設計 | 中 | `schedule/errors.go:9-23` | エラー型の不整合 |
| 5 | 一貫性 | 中 | `attendanceApi.ts` | apiClient使用が統一されていない |
| 6 | 一貫性 | 中 | `scheduleApi.ts` | apiClient使用が統一されていない |
| 7 | 型安全性 | 低 | `scheduleApi.ts:162-163` | レスポンス構造の不一致 |
| 8 | パフォーマンス | 中 | `AttendanceList.tsx:258-260` | useMemo未使用 |
| 9 | パフォーマンス | 中 | `ScheduleList.tsx:162-164` | useMemo未使用 |
| 10 | コード品質 | 低 | `ScheduleList.tsx:82` | eslint-disableでルール無効化 |
| 11 | アクセシビリティ | 中 | 両ページコンポーネント | aria-label欠如 |
| 12 | 保守性 | 低 | 両ページコンポーネント | コンポーネントが大きすぎる |

### 2.2 問題点の詳細

---

#### 問題1: エラーがラップされていない（App層・出欠確認）

**ファイル**: `backend/internal/app/attendance/update_collection_usecase.go:23-35`
**カテゴリ**: エラーハンドリング
**深刻度**: 中

##### 何が問題だったか

エラーがそのまま返されており、発生箇所を特定するコンテキスト情報がありません。

##### なぜ問題なのか

- **デバッグの困難さ**: エラーが発生した際、スタックトレースからどの処理で失敗したか分かりにくい
- **運用時の影響**: 本番環境でのトラブルシューティングに時間がかかる
- **Go のベストプラクティス違反**: Go では `fmt.Errorf("context: %w", err)` でエラーをラップすることが推奨される

##### 修正前のコード

```go
// backend/internal/app/attendance/update_collection_usecase.go:23-27
tenantID, err := common.ParseTenantID(input.TenantID)
if err != nil {
    return nil, err  // コンテキスト情報がない
}
```

##### 修正後のコード

```go
// backend/internal/app/attendance/update_collection_usecase.go:23-27
tenantID, err := common.ParseTenantID(input.TenantID)
if err != nil {
    return nil, fmt.Errorf("tenant ID のパースに失敗: %w", err)
}
```

##### 学習ポイント

Go では `%w` を使ってエラーをラップすることで、`errors.Is()` や `errors.As()` でエラーチェーンを辿れるようになります。これにより、呼び出し側で適切なエラーハンドリングが可能になります。

---

#### 問題2: エラーがラップされていない（App層・日程調整）

**ファイル**: `backend/internal/app/schedule/update_schedule_usecase.go:24-37`
**カテゴリ**: エラーハンドリング
**深刻度**: 中

##### 何が問題だったか

問題1と同様に、エラーがラップされていません。

##### なぜ問題なのか

問題1と同様の理由です。また、同じパッケージ内でエラーハンドリングが一貫していないと、コードの品質にばらつきが生じます。

##### 修正前のコード

```go
// backend/internal/app/schedule/update_schedule_usecase.go:24-28
tenantID, err := common.ParseTenantID(input.TenantID)
if err != nil {
    return nil, err
}
```

##### 修正後のコード

```go
// backend/internal/app/schedule/update_schedule_usecase.go:24-28
tenantID, err := common.ParseTenantID(input.TenantID)
if err != nil {
    return nil, fmt.Errorf("tenant ID のパースに失敗: %w", err)
}
```

---

#### 問題3: 関数名と戻り値の型が不一致

**ファイル**: `backend/internal/app/schedule/update_schedule_usecase.go:121-143`
**カテゴリ**: 命名
**深刻度**: 低

##### 何が問題だったか

`hasResponsesForCandidates` という関数名は bool を返すことを示唆しますが、実際には `*schedule.CandidateDate` を返しています。

##### なぜ問題なのか

- **可読性の低下**: 関数名から戻り値を予測できない
- **認知的負荷**: コードを読む人が混乱する
- **命名規則違反**: `has` や `is` で始まる関数は通常 bool を返す

##### 修正前のコード

```go
// 関数名が bool を返すことを示唆している
func (u *UpdateScheduleUsecase) hasResponsesForCandidates(
    ctx context.Context,
    candidates []schedule.CandidateDateItem,
    existingResponses []*schedule.ScheduleResponse,
) (*schedule.CandidateDate, error)
```

##### 修正後のコード

```go
// 関数名が CandidateDate を返すことを明示
func (u *UpdateScheduleUsecase) findCandidateWithExistingResponses(
    ctx context.Context,
    candidates []schedule.CandidateDateItem,
    existingResponses []*schedule.ScheduleResponse,
) (*schedule.CandidateDate, error)
```

##### 学習ポイント

関数名は以下の規則に従いましょう：
- `is`, `has`, `can`, `should` → bool を返す
- `get`, `find`, `fetch` → オブジェクトを返す
- `create`, `new` → 新規オブジェクトを生成

---

#### 問題4: エラー型の不整合（Domain層）

**ファイル**: `backend/internal/domain/schedule/errors.go:9-23`
**カテゴリ**: ドメイン設計
**深刻度**: 中

##### 何が問題だったか

schedule パッケージのエラー定義で、一部は `errors.New` を使用し、一部は `common.NewInvariantViolationError` を使用していました。

##### なぜ問題なのか

- **型安全性の低下**: エラーの型が統一されていないと、型による判定が困難
- **一貫性の欠如**: attendance パッケージは全て `common.NewInvariantViolationError` を使用しており、パッケージ間で不整合
- **エラーハンドリングの複雑化**: 呼び出し側で複数のエラー型を考慮する必要がある

##### 修正前のコード

```go
// backend/internal/domain/schedule/errors.go
var (
    // 一部は errors.New を使用
    ErrScheduleClosed = errors.New("schedule is closed")
    ErrDeadlinePassed = errors.New("deadline has passed")
    ErrAlreadyClosed = errors.New("schedule is already closed")

    // 一部は common.NewInvariantViolationError を使用
    ErrAlreadyDeleted = common.NewInvariantViolationError("schedule is already deleted")
)
```

##### 修正後のコード

```go
// backend/internal/domain/schedule/errors.go
var (
    // 全て common.NewInvariantViolationError に統一
    ErrScheduleClosed = common.NewInvariantViolationError("schedule is closed")
    ErrDeadlinePassed = common.NewInvariantViolationError("deadline has passed")
    ErrAlreadyClosed = common.NewInvariantViolationError("schedule is already closed")
    ErrAlreadyDeleted = common.NewInvariantViolationError("schedule is already deleted")
)
```

##### 学習ポイント

DDDにおいて、ドメイン層のエラーは**ドメイン固有のエラー型**を使用することで：
- Interface層で適切なHTTPステータスコードに変換しやすくなる
- エラーの発生源（ドメインルール違反 vs インフラエラー）を明確に区別できる

---

#### 問題5: apiClient使用が統一されていない（Frontend API・出欠確認）

**ファイル**: `web-frontend/src/lib/api/attendanceApi.ts`
**カテゴリ**: 一貫性
**深刻度**: 中

##### 何が問題だったか

一部の関数は `apiClient` を使用し、他の関数は直接 `fetch` を使用していました。

| 関数名 | 使用方法 |
|--------|----------|
| `listAttendanceCollections` | 直接fetch |
| `createAttendanceCollection` | 直接fetch |
| `updateAttendanceCollection` | apiClient |
| `getAttendanceCollection` | 直接fetch |

##### なぜ問題なのか

- **エラーハンドリングの不一致**: apiClient は `ApiClientError` を投げ、直接 fetch は独自の Error を投げる
- **認証処理のコード重複**: 各関数でトークン取得・チェックを繰り返し
- **保守性の低下**: 認証方式を変更する場合、全ての関数を修正する必要がある

##### 修正案

すべての関数を `apiClient` を使用するように統一することで、エラーハンドリングと認証処理を一箇所に集約できます。

---

#### 問題6: apiClient使用が統一されていない（Frontend API・日程調整）

**ファイル**: `web-frontend/src/lib/api/scheduleApi.ts`
**カテゴリ**: 一貫性
**深刻度**: 中

##### 何が問題だったか

問題5と同様に、apiClient の使用が統一されていません。

---

#### 問題7: レスポンス構造の不一致

**ファイル**: `web-frontend/src/lib/api/scheduleApi.ts:162-163`
**カテゴリ**: 型安全性
**深刻度**: 低

##### 何が問題だったか

`getSchedule` は `result` をそのまま返していますが、他の関数は `result.data` を返しています。

##### 修正前のコード

```typescript
// scheduleApi.ts:162-163
const result = await response.json();
return result;  // result 全体を返す
```

##### 他の関数のパターン

```typescript
const result: ApiResponse<Schedule> = await response.json();
return result.data;  // result.data を返す
```

---

#### 問題8: useMemo未使用によるパフォーマンス問題（出欠確認ページ）

**ファイル**: `web-frontend/src/pages/AttendanceList.tsx:258-260`
**カテゴリ**: パフォーマンス
**深刻度**: 中

##### 何が問題だったか

`existingDateStrings` が毎レンダリングで再計算されています。

##### なぜ問題なのか

- **不要な計算**: 依存する値（`targetDates`）が変わっていないのに毎回計算される
- **パフォーマンス低下**: 配列のフィルタリングとマップは計算コストがかかる
- **子コンポーネントへの影響**: 参照が変わるため、メモ化された子コンポーネントも再レンダリングされる可能性

##### 修正前のコード

```tsx
// AttendanceList.tsx:258-260
const existingDateStrings = targetDates
  .filter((d) => d.date.trim() !== '')
  .map((d) => d.date);
```

##### 修正後のコード

```tsx
// AttendanceList.tsx:258-260
const existingDateStrings = useMemo(() =>
  targetDates
    .filter((d) => d.date.trim() !== '')
    .map((d) => d.date),
  [targetDates]
);
```

##### 学習ポイント

React の `useMemo` は：
- 計算コストの高い処理をキャッシュ
- 依存配列の値が変わらない限り、前回の計算結果を再利用
- 不要な再レンダリングを防ぐ

---

#### 問題9: useMemo未使用によるパフォーマンス問題（日程調整ページ）

**ファイル**: `web-frontend/src/pages/ScheduleList.tsx:162-164`
**カテゴリ**: パフォーマンス
**深刻度**: 中

##### 何が問題だったか

問題8と同様に、`existingDateStrings` が毎レンダリングで再計算されています。

---

#### 問題10: eslint-disableでルール無効化

**ファイル**: `web-frontend/src/pages/ScheduleList.tsx:82`
**カテゴリ**: コード品質
**深刻度**: 低

##### 何が問題だったか

`eslint-disable-next-line react-hooks/exhaustive-deps` でリントルールを無効化しています。

##### なぜ問題なのか

- **将来的なバグの原因**: 依存配列の問題を隠蔽している
- **リントの意味がなくなる**: ルールを無効化すると、静的解析の恩恵を受けられない

##### 修正前のコード

```tsx
useEffect(() => {
  loadSchedules();
  loadMemberGroups();
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, []);
```

##### 修正後のコード

```tsx
const loadSchedules = useCallback(async () => { ... }, []);
const loadMemberGroups = useCallback(async () => { ... }, []);

useEffect(() => {
  loadSchedules();
  loadMemberGroups();
}, [loadSchedules, loadMemberGroups]);
```

##### 学習ポイント

`eslint-disable` を使う前に、まず根本原因を解決しましょう：
- 関数を `useCallback` でメモ化する
- 依存配列を正しく設定する
- どうしても必要な場合のみ、理由をコメントで明記

---

#### 問題11: aria-label欠如によるアクセシビリティ問題

**ファイル**: `AttendanceList.tsx:597-604`, `ScheduleList.tsx:480-488`
**カテゴリ**: アクセシビリティ
**深刻度**: 中

##### 何が問題だったか

削除ボタンに `aria-label` がありません。

##### なぜ問題なのか

- **スクリーンリーダーユーザーへの影響**: 何を削除するボタンなのか分からない
- **WCAG違反**: Web Content Accessibility Guidelines に反する
- **法的リスク**: アクセシビリティ要件を満たさない場合、訴訟リスクがある国も

##### 修正前のコード

```tsx
<button onClick={() => handleRemoveDate(index)}>削除</button>
```

##### 修正後のコード

```tsx
<button
  onClick={() => handleRemoveDate(index)}
  aria-label={`日程${index + 1}を削除`}
>
  削除
</button>
```

---

#### 問題12: コンポーネントが大きすぎる

**ファイル**: 両ページコンポーネント
**カテゴリ**: 保守性
**深刻度**: 低（別PRで対応）

##### 何が問題だったか

- `AttendanceList.tsx`: 1065行
- `ScheduleList.tsx`: 807行

##### なぜ問題なのか

- **可読性の低下**: 一目でコンポーネントの全体像を把握できない
- **テストの困難さ**: 大きなコンポーネントは単体テストが難しい
- **再利用性の低下**: フォーム部分を他で使い回せない

##### 推奨する対応（別PRで実施）

- フォーム部分を `AttendanceCreateForm` / `ScheduleCreateForm` として分離
- 一覧表示部分を `AttendanceTable` / `ScheduleTable` として分離
- 各コンポーネントを 200-300行程度に収める

---

## 2.3 良い点（参考）

レビューでは問題点だけでなく、良い実装も確認しました。

### Backend

| カテゴリ | 内容 |
|---------|------|
| **DDD パターン** | DTO とドメインオブジェクトが適切に分離されている |
| **単一責任原則** | 各ユースケースが明確な責務を持っている |
| **監査ログ** | 更新操作に対して適切にログを記録している |
| **Clock インターフェース** | テスタビリティを考慮した設計 |
| **不変条件の保護** | validate() メソッドで不変条件を適切に検証 |
| **SQLインジェクション対策** | パラメータ化クエリ（$1, $2...）を使用 |
| **ソフトデリート** | deletedAt フィールドによる論理削除 |

### Frontend

| カテゴリ | 内容 |
|---------|------|
| **timeUtils.ts** | 明確なJSDocコメント、エッジケースが網羅的にドキュメント化 |
| **apiClient.ts** | 適切なエラーハンドリング、204 No Content の処理 |
| **ユーザーフレンドリー** | getUserMessage() による分かりやすいエラーメッセージ |

---

## 修正実施状況サマリー

| # | 問題 | 状況 |
|---|------|------|
| 1 | エラーラップ（出欠確認） | 修正済み |
| 2 | エラーラップ（日程調整） | 修正済み |
| 3 | 関数名修正 | 修正済み |
| 4 | エラー型統一 | 修正済み |
| 5 | apiClient統一（出欠確認） | 修正済み |
| 6 | apiClient統一（日程調整） | 修正済み |
| 7 | レスポンス構造修正 | 修正済み |
| 8 | useMemo追加（出欠確認） | 修正済み |
| 9 | useMemo追加（日程調整） | 修正済み |
| 10 | eslint-disable削除 | 修正済み |
| 11 | aria-label追加 | 修正済み |
| 12 | コンポーネント分離 | 別PRで対応 |

---

## 3. 修正内容の解説

このセクションでは、レビューで発見された各問題に対してどのような修正が行われたか、その背景と理由を詳しく解説します。

### 3.1 Backend 修正

#### 3.1.1 エラーラッピングの追加

**対象ファイル**:
- `backend/internal/app/attendance/update_collection_usecase.go`
- `backend/internal/app/schedule/update_schedule_usecase.go`

**修正の背景**:

Go言語では、エラーが発生した際に「どこで」「なぜ」エラーが起きたかを追跡するために、エラーをラップ（wrap）することが推奨されています。

```go
// ❌ 悪い例：エラーをそのまま返す
if err != nil {
    return nil, err
}

// ✅ 良い例：コンテキスト情報を付けてラップ
if err != nil {
    return nil, fmt.Errorf("tenant ID のパースに失敗: %w", err)
}
```

**なぜ `%w` を使うのか**:

Go 1.13 以降、`fmt.Errorf` で `%w` を使うと、元のエラーを「ラップ」できます。これにより：

1. **エラーチェーンの追跡**: `errors.Unwrap()` で元のエラーを取得可能
2. **型による判定**: `errors.Is()` や `errors.As()` でエラーの種類を判定可能
3. **スタックトレースの改善**: どの処理で失敗したかが明確

```go
// 呼び出し側でのエラーチェック例
if errors.Is(err, common.ErrNotFound) {
    // NotFound エラーとして処理
}
```

#### 3.1.2 関数名の修正

**対象ファイル**: `backend/internal/app/schedule/update_schedule_usecase.go`

**修正内容**:
```go
// 修正前
func (u *UpdateScheduleUsecase) hasResponsesForCandidates(...) (*schedule.CandidateDate, error)

// 修正後
func (u *UpdateScheduleUsecase) findCandidateWithExistingResponses(...) (*schedule.CandidateDate, error)
```

**命名規則の重要性**:

関数名はコードの「ドキュメント」です。適切な命名により：

| プレフィックス | 期待される戻り値 | 例 |
|---------------|-----------------|-----|
| `is`, `has`, `can` | `bool` | `isValid()`, `hasPermission()` |
| `get`, `find`, `fetch` | オブジェクト | `getUser()`, `findById()` |
| `create`, `new` | 新規オブジェクト | `createOrder()`, `newInstance()` |
| `update`, `set` | なし or 更新後オブジェクト | `updateProfile()` |
| `delete`, `remove` | なし or `bool` | `deleteItem()` |

#### 3.1.3 エラー型の統一

**対象ファイル**: `backend/internal/domain/schedule/errors.go`

**修正内容**:
```go
// 修正前：型がバラバラ
var (
    ErrScheduleClosed = errors.New("schedule is closed")           // 標準errors
    ErrAlreadyDeleted = common.NewInvariantViolationError("...")   // カスタム型
)

// 修正後：カスタム型に統一
var (
    ErrScheduleClosed = common.NewInvariantViolationError("schedule is closed")
    ErrAlreadyDeleted = common.NewInvariantViolationError("schedule is already deleted")
)
```

**なぜカスタムエラー型を使うのか**:

DDDにおいて、ドメイン層のエラーは**ビジネスルール違反**を表します。カスタム型を使うことで：

1. **HTTP ステータスコードへの変換が容易**:
   - `InvariantViolationError` → 400 Bad Request
   - `NotFoundError` → 404 Not Found
   - 標準 `error` → 500 Internal Server Error

2. **エラーの発生源が明確**:
   - ドメインエラー vs インフラエラー vs アプリケーションエラー

### 3.2 Frontend 修正

#### 3.2.1 apiClient への統一

**対象ファイル**:
- `web-frontend/src/lib/api/attendanceApi.ts`
- `web-frontend/src/lib/api/scheduleApi.ts`

**修正前の問題点**:

```typescript
// ❌ 直接 fetch を使う場合：毎回ボイラープレートが必要
export async function listAttendanceCollections(...) {
    const session = await getSession();
    if (!session?.access_token) {
        throw new Error('認証が必要です');
    }
    const response = await fetch(url, {
        headers: {
            'Authorization': `Bearer ${session.access_token}`,
            'Content-Type': 'application/json',
        },
    });
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
}

// ✅ apiClient を使う場合：シンプル
export async function listAttendanceCollections(...) {
    return apiClient.get<AttendanceCollection[]>(url);
}
```

**統一による恩恵**:

| 観点 | 修正前 | 修正後 |
|------|--------|--------|
| 認証処理 | 各関数で重複 | apiClient で一元管理 |
| エラーハンドリング | バラバラ | 統一された `ApiClientError` |
| コード行数 | 約15行/関数 | 約1行/関数 |
| 保守性 | 低（変更箇所が多い） | 高（変更は1箇所） |

#### 3.2.2 useMemo によるパフォーマンス最適化

**対象ファイル**:
- `web-frontend/src/pages/AttendanceList.tsx`
- `web-frontend/src/pages/ScheduleList.tsx`

**修正内容**:

```tsx
// ❌ 修正前：毎レンダリングで再計算
const existingDateStrings = targetDates
  .filter((d) => d.date.trim() !== '')
  .map((d) => d.date);

// ✅ 修正後：依存配列が変わった時のみ再計算
const existingDateStrings = useMemo(() =>
  targetDates
    .filter((d) => d.date.trim() !== '')
    .map((d) => d.date),
  [targetDates]
);
```

**useMemo を使うべき場面**:

```
┌─────────────────────────────────────────────────────────┐
│  useMemo の判断フローチャート                            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  計算コストは高い？                                      │
│       │                                                 │
│       ├── Yes → useMemo を検討                         │
│       │                                                 │
│       └── No  → useMemo 不要                           │
│                （過剰な最適化は複雑性を増す）            │
│                                                         │
│  「計算コストが高い」の目安：                            │
│   - 配列の filter/map/reduce（要素数が多い場合）        │
│   - オブジェクトの深い走査                              │
│   - 正規表現のコンパイル                                │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### 3.2.3 useCallback による関数の安定化

**対象ファイル**: `web-frontend/src/pages/ScheduleList.tsx`

**修正内容**:

```tsx
// ❌ 修正前：eslint-disable でルールを無効化
useEffect(() => {
  loadSchedules();
  loadMemberGroups();
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, []);

// ✅ 修正後：useCallback で関数を安定化
const loadSchedules = useCallback(async () => {
  // ...
}, []);

const loadMemberGroups = useCallback(async () => {
  // ...
}, []);

useEffect(() => {
  loadSchedules();
  loadMemberGroups();
}, [loadSchedules, loadMemberGroups]);
```

**eslint-disable を避けるべき理由**:

1. **将来のバグを隠蔽**: 依存配列の問題が検出されなくなる
2. **リントの意味がなくなる**: 静的解析の恩恵を放棄
3. **コードレビューでの信頼低下**: 「なぜ無効化したのか」の説明が必要

#### 3.2.4 アクセシビリティの向上

**対象ファイル**: 両ページコンポーネント

**修正内容**:

```tsx
// ❌ 修正前：aria-label なし
<button onClick={() => handleRemoveDate(index)}>
  削除
</button>

// ✅ 修正後：aria-label あり
<button
  onClick={() => handleRemoveDate(index)}
  aria-label={`日程${index + 1}を削除`}
>
  削除
</button>
```

**アクセシビリティが重要な理由**:

1. **スクリーンリーダーユーザー**: 視覚障害者がボタンの目的を理解できる
2. **WCAG 準拠**: Web Content Accessibility Guidelines への対応
3. **SEO への好影響**: 検索エンジンもアクセシビリティを評価
4. **法的要件**: 一部の国・地域ではアクセシビリティ対応が法的に必要

---

## 4. まとめ

### 4.1 この PR から学べること

#### 設計パターン

| パターン | 適用例 | 学習ポイント |
|---------|--------|-------------|
| **DDD レイヤードアーキテクチャ** | 全体構成 | 責務の分離、テスタビリティ向上 |
| **リポジトリパターン** | `schedule_repository.go` | データアクセスの抽象化 |
| **DTO パターン** | `dto.go` | 層間のデータ受け渡し |
| **不変条件の保護** | `schedule.go` の validate() | ドメインオブジェクトの整合性保証 |

#### コーディングベストプラクティス

| 言語/FW | プラクティス | 具体例 |
|---------|-------------|--------|
| **Go** | エラーラッピング | `fmt.Errorf("context: %w", err)` |
| **Go** | カスタムエラー型 | `common.NewInvariantViolationError()` |
| **Go** | 命名規則 | `find*`, `has*`, `is*` の使い分け |
| **React** | `useMemo` | 計算コストの高い処理のキャッシュ |
| **React** | `useCallback` | 関数の参照安定化 |
| **React** | アクセシビリティ | `aria-label` の適切な設定 |

### 4.2 よくある間違いと対策

| よくある間違い | 対策 |
|---------------|------|
| エラーをそのまま return | `fmt.Errorf("コンテキスト: %w", err)` でラップ |
| 関数名と戻り値の不一致 | 命名規則に従う（has→bool, find→object） |
| エラー型の混在 | プロジェクト内で統一されたエラー型を使用 |
| fetch の直接使用 | apiClient などの共通クライアントを使用 |
| eslint-disable の多用 | 根本原因を解決（useCallback 等） |
| useMemo の不使用 | 計算コストが高い処理には useMemo |
| aria-label の欠如 | インタラクティブ要素には必ず追加 |

### 4.3 参考資料

#### DDD・アーキテクチャ

- [Domain-Driven Design Reference（公式）](https://www.domainlanguage.com/ddd/reference/)
- [Clean Architecture（Uncle Bob）](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Implementing Domain-Driven Design（書籍）](https://www.amazon.co.jp/dp/B00BCLEBN8)

#### Go 言語

- [Effective Go（公式）](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Error Handling（公式 Blog）](https://go.dev/blog/go1.13-errors)

#### React / TypeScript

- [React 公式ドキュメント](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/)
- [React Hooks API Reference](https://react.dev/reference/react)
- [useMemo 公式ドキュメント](https://react.dev/reference/react/useMemo)

#### アクセシビリティ

- [WCAG 2.1 Guidelines](https://www.w3.org/TR/WCAG21/)
- [WAI-ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [A11y Project](https://www.a11yproject.com/)

#### プロジェクト内のルール

このプロジェクトでは以下のルールファイルを参照してください：

| ファイル | 内容 |
|---------|------|
| `.claude/rules/ddd-patterns.md` | DDD パターンの適用ルール |
| `.claude/rules/go-coding-style.md` | Go コーディング規約 |
| `.claude/rules/testing.md` | テストの書き方 |
| `.claude/rules/security.md` | セキュリティガイドライン |

### 4.4 次のステップ

この PR のレビューを通じて学んだことを活かすために：

1. **実際にコードを読む**: 修正前後のコードを GitHub で確認
2. **類似のコードを探す**: プロジェクト内で同じパターンが使われている箇所を探索
3. **小さな改善を試す**: 学んだベストプラクティスを他の箇所に適用
4. **チームで共有**: 学んだことをチームメンバーと共有

