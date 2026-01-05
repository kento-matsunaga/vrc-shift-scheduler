# Issue #74 実装計画: お知らせ機能とチュートリアル

## ブランチ
`feature/announcements-and-tutorials`

---

## Phase 1: データベース・ドメイン層

### 1.1 マイグレーション作成
- [ ] `032_create_announcements.up.sql`
  ```sql
  CREATE TABLE announcements (
    id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26), -- NULL = 全テナント向け
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
  );
  CREATE INDEX idx_announcements_tenant ON announcements(tenant_id);
  CREATE INDEX idx_announcements_published ON announcements(published_at);
  CREATE INDEX idx_announcements_deleted ON announcements(deleted_at);

  CREATE TABLE announcement_reads (
    id VARCHAR(26) PRIMARY KEY,
    announcement_id VARCHAR(26) NOT NULL,
    admin_id VARCHAR(26) NOT NULL,
    read_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (announcement_id) REFERENCES announcements(id),
    FOREIGN KEY (admin_id) REFERENCES admins(id),
    UNIQUE (announcement_id, admin_id)
  );
  CREATE INDEX idx_announcement_reads_admin ON announcement_reads(admin_id);
  ```

- [ ] `033_create_tutorials.up.sql`
  ```sql
  CREATE TABLE tutorials (
    id VARCHAR(26) PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
  );
  CREATE INDEX idx_tutorials_category ON tutorials(category);
  CREATE INDEX idx_tutorials_published ON tutorials(is_published);
  CREATE INDEX idx_tutorials_order ON tutorials(display_order);
  ```

### 1.2 ドメインエンティティ作成

#### backend/internal/domain/announcement/
- [ ] `announcement.go` - Announcementエンティティ
  - ID, TenantID, Title, Body, PublishedAt, CreatedAt, UpdatedAt, DeletedAt
  - NewAnnouncement(), Update(), Delete(), IsPublished()
- [ ] `announcement_test.go` - テスト
- [ ] `repository.go` - リポジトリインターフェース
- [ ] `errors.go` - ドメインエラー

#### backend/internal/domain/tutorial/
- [ ] `tutorial.go` - Tutorialエンティティ
  - ID, Category, Title, Body, DisplayOrder, IsPublished, CreatedAt, UpdatedAt
  - NewTutorial(), Update(), Delete(), Publish(), Unpublish()
- [ ] `tutorial_test.go` - テスト
- [ ] `repository.go` - リポジトリインターフェース

---

## Phase 2: インフラ層（リポジトリ実装）

### 2.1 お知らせリポジトリ
- [ ] `backend/internal/infra/db/announcement_repository.go`
  - Save(), FindByID(), FindAll(), FindByTenantID()
  - FindUnreadByAdminID(), MarkAsRead(), MarkAllAsRead()
  - GetUnreadCount()

### 2.2 チュートリアルリポジトリ
- [ ] `backend/internal/infra/db/tutorial_repository.go`
  - Save(), FindByID(), FindAll(), FindPublished()
  - FindByCategory()

---

## Phase 3: アプリケーション層（ユースケース）

### 3.1 お知らせユースケース
#### backend/internal/app/announcement/
- [ ] `list_announcements_usecase.go` - 一覧取得（公開済み、テナント対象）
- [ ] `get_announcement_usecase.go` - 詳細取得
- [ ] `mark_as_read_usecase.go` - 既読にする
- [ ] `mark_all_as_read_usecase.go` - すべて既読
- [ ] `get_unread_count_usecase.go` - 未読件数取得
- [ ] `create_announcement_usecase.go` - 作成（admin用）
- [ ] `update_announcement_usecase.go` - 更新（admin用）
- [ ] `delete_announcement_usecase.go` - 削除（admin用）
- [ ] `dto.go` - 入出力DTO

### 3.2 チュートリアルユースケース
#### backend/internal/app/tutorial/
- [ ] `list_tutorials_usecase.go` - 一覧取得（公開済み）
- [ ] `get_tutorial_usecase.go` - 詳細取得
- [ ] `create_tutorial_usecase.go` - 作成（admin用）
- [ ] `update_tutorial_usecase.go` - 更新（admin用）
- [ ] `delete_tutorial_usecase.go` - 削除（admin用）
- [ ] `dto.go` - 入出力DTO

---

## Phase 4: インターフェース層（REST API）

### 4.1 お知らせハンドラー
- [ ] `backend/internal/interface/rest/announcement_handler.go`
  ```
  GET  /api/v1/announcements           - 一覧取得
  GET  /api/v1/announcements/:id       - 詳細取得
  POST /api/v1/announcements/:id/read  - 既読にする
  POST /api/v1/announcements/read-all  - すべて既読
  GET  /api/v1/announcements/unread-count - 未読件数
  ```

- [ ] `backend/internal/interface/rest/admin_announcement_handler.go`
  ```
  GET    /api/v1/admin/announcements     - 管理用一覧
  POST   /api/v1/admin/announcements     - 作成
  PUT    /api/v1/admin/announcements/:id - 更新
  DELETE /api/v1/admin/announcements/:id - 削除
  ```

### 4.2 チュートリアルハンドラー
- [ ] `backend/internal/interface/rest/tutorial_handler.go`
  ```
  GET /api/v1/tutorials      - 一覧取得
  GET /api/v1/tutorials/:id  - 詳細取得
  ```

- [ ] `backend/internal/interface/rest/admin_tutorial_handler.go`
  ```
  GET    /api/v1/admin/tutorials     - 管理用一覧
  POST   /api/v1/admin/tutorials     - 作成
  PUT    /api/v1/admin/tutorials/:id - 更新
  DELETE /api/v1/admin/tutorials/:id - 削除
  ```

### 4.3 ルーター更新
- [ ] `backend/internal/interface/rest/router.go` にエンドポイント追加

---

## Phase 5: フロントエンド（web-frontend）

### 5.1 API クライアント
- [ ] `web-frontend/src/lib/api/announcementApi.ts`
- [ ] `web-frontend/src/lib/api/tutorialApi.ts`

### 5.2 型定義
- [ ] `web-frontend/src/types/api.ts` に追加
  - Announcement, AnnouncementListResponse
  - Tutorial, TutorialListResponse

### 5.3 コンポーネント
- [ ] `web-frontend/src/components/AnnouncementBell.tsx`
  - ベルアイコン + 未読バッジ（オレンジ）
  - ドロップダウンでお知らせ一覧表示
  - クリックで既読化

- [ ] `web-frontend/src/components/TutorialButton.tsx`
  - 「？」アイコン
  - クリックでモーダル/ページ遷移

- [ ] `web-frontend/src/components/TutorialModal.tsx`
  - カテゴリ別チュートリアル表示
  - Markdown レンダリング

### 5.4 Layout更新
- [ ] `web-frontend/src/components/Layout.tsx`
  - ヘッダー右側に AnnouncementBell と TutorialButton 追加
  ```tsx
  <header>
    ...
    <div className="flex items-center gap-2">
      <TutorialButton />
      <AnnouncementBell />
      <UserMenu />
    </div>
  </header>
  ```

---

## Phase 6: フロントエンド（admin-frontend）

### 6.1 API クライアント
- [ ] `admin-frontend/src/lib/api.ts` に追加

### 6.2 ページ
- [ ] `admin-frontend/src/pages/Announcements.tsx`
  - お知らせ一覧・作成・編集・削除
  - 対象テナント選択（全体 or 特定）
  - 公開日時設定

- [ ] `admin-frontend/src/pages/Tutorials.tsx`
  - チュートリアル一覧・作成・編集・削除
  - カテゴリ管理
  - 表示順設定
  - 公開/非公開切り替え

### 6.3 ナビゲーション更新
- [ ] サイドバーに「お知らせ管理」「チュートリアル管理」追加

---

## Phase 7: テスト・仕上げ

- [ ] ユースケーステスト追加
- [ ] リポジトリ統合テスト追加
- [ ] E2Eテスト（手動確認）
- [ ] ビルド確認
- [ ] PR作成

---

## 実装順序（推奨）

1. **DB マイグレーション** → ドメイン層の基盤
2. **ドメインエンティティ** → ビジネスロジック定義
3. **リポジトリ実装** → データアクセス
4. **ユースケース** → アプリケーションロジック
5. **RESTハンドラー** → API公開
6. **web-frontend** → ユーザー向けUI
7. **admin-frontend** → 管理者向けUI

---

## 注意点

- お知らせは `tenant_id = NULL` で全テナント向け、指定ありで特定テナント向け
- 既読管理は `admin_id` 単位（ログインユーザー）
- チュートリアルはテナント共通（全システム共通コンテンツ）
- Markdown対応は `react-markdown` 等を使用
- 未読バッジの色: `#F97316` (Tailwind orange-500)
