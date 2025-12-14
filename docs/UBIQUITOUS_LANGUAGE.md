# ユビキタス言語（Ubiquitous Language）

VRC Shift Scheduler プロジェクトにおけるユビキタス言語の定義書です。
開発チームとドメインエキスパートが共通して使用する用語を定義しています。

---

## 目次

1. [テナント・イベント領域](#1-テナントイベント領域)
2. [メンバー・ロール領域](#2-メンバーロール領域)
3. [シフト枠・ポジション領域](#3-シフト枠ポジション領域)
4. [シフト希望領域](#4-シフト希望領域)
5. [シフト確定・割り当て領域](#5-シフト確定割り当て領域)
6. [出欠確認・日程調整領域](#6-出欠確認日程調整領域)
7. [通知・リマインド領域](#7-通知リマインド領域)
8. [監査・履歴領域](#8-監査履歴領域)
9. [共通概念](#9-共通概念)
10. [状態・区分値](#10-状態区分値)

---

## 1. テナント・イベント領域

### テナント（Tenant）

VRChat内で活動するひとつの団体・店舗・イベント運営チームを表す最上位の境界。

- **英語コード**: `tenant`
- **識別子**: `TenantID`（ULID形式）
- **役割**: 全てのイベント・メンバー・シフト情報の境界として機能し、テナント間でのデータ混在を防ぐ
- **関連**: 通常は1つのDiscordサーバに対応する
- **例**: 「シトロン」「カフェABC」

```go
type Tenant struct {
    tenantID    TenantID
    tenantName  string
    description string
}
```

### イベント（Event）

テナントが運営する営業・イベントの単位。通常営業と特別営業の両方を管理する。

- **英語コード**: `event`
- **識別子**: `EventID`（ULID形式）
- **役割**: 営業日・ポジション・シフト枠の生成と整合性を保証する集約ルート
- **例**: 「シトロン通常営業」「Vket特別営業」

```go
type Event struct {
    eventID     EventID
    tenantID    TenantID
    eventName   string
    eventType   EventType
    description string
    isActive    bool
}
```

### イベント種別（Event Type）

イベントの反復パターンの有無を区別する区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `normal` | 通常営業 | 毎週特定曜日に反復開催されるレギュラー営業 |
| `special` | 特別営業 | 特定日程のみの単発イベント |

### イベント営業日（Event Business Day）

イベントごとの実際の営業日単位。シフトを組む対象となる「1回分の営業日」。

- **英語コード**: `event_business_day`, `business_day`
- **識別子**: `BusinessDayID`（ULID形式）
- **役割**: 通常営業パターンから自動生成された日付と、特別営業日として登録された日付の両方を管理
- **例**: 「2025-02-13 21:30-23:00 シトロン通常営業」

```go
type EventBusinessDay struct {
    businessDayID       BusinessDayID
    tenantID            TenantID
    eventID             EventID
    targetDate          time.Time  // 営業日（DATE）
    startTime           time.Time  // 開始時刻（TIME）
    endTime             time.Time  // 終了時刻（TIME）
    occurrenceType      OccurrenceType
    recurringPatternID  *EventID
    isActive            bool
    validFrom           *time.Time
    validTo             *time.Time
}
```

### 発生種別（Occurrence Type）

営業日の発生パターンを区別する区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `recurring` | 定期営業 | 通常営業パターンから自動生成された営業日 |
| `special` | 特別営業 | 個別に登録された特別営業日 |

### 通常営業パターン（Recurring Pattern）

曜日・時間帯が反復する営業の定義。週次で繰り返される営業のルールを表す。

- **英語コード**: `recurring_pattern`
- **識別子**: `PatternID`（ULID形式）
- **役割**: パターンに基づいて営業日を自動生成する
- **例**: 「毎週木曜 21:30〜23:00」

```go
type RecurringPattern struct {
    patternID   EventID
    tenantID    TenantID
    eventID     EventID
    patternType PatternType
    config      RecurringPatternConfig
}
```

### パターン種別（Pattern Type）

通常営業パターンの繰り返しルールを区別する区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `weekly` | 曜日指定 | 毎週特定曜日に開催 |
| `monthly_date` | 月内日付指定 | 毎月特定日に開催 |
| `custom` | カスタム | 柔軟なカスタムルール |

### 曜日（Day of Week）

週内の曜日を表す区分オブジェクト。

| 値 | 日本語名 |
|---|---|
| `MON` | 月曜日 |
| `TUE` | 火曜日 |
| `WED` | 水曜日 |
| `THU` | 木曜日 |
| `FRI` | 金曜日 |
| `SAT` | 土曜日 |
| `SUN` | 日曜日 |

---

## 2. メンバー・ロール領域

### メンバー（Member）

テナントに所属するキャスト・店長・スタッフを表す。シフトの対象となる人物。

- **英語コード**: `member`
- **識別子**: `MemberID`（ULID形式）
- **役割**: VRChatアカウント情報・ロール・外部アカウントを管理し、シフト確定権限の判定を行う
- **例**: 「アリス（店長）」「ボブ（キャスト）」

```go
type Member struct {
    memberID      MemberID
    tenantID      TenantID
    displayName   string       // 表示名（キャスト名）
    discordUserID string       // Discord連携用ID
    email         string       // メールアドレス
    isActive      bool
}
```

### 表示名（Display Name）

メンバーの公開名称。VRChat内での表示名やキャスト名として使用される。

- **英語コード**: `display_name`
- **役割**: シフト表やUIで表示される名前
- **注意**: 表示名が変更されても、同一人物として追跡できるように内部識別子で管理する

### ロール種別（Role Type）

メンバーの権限レベルを表す区分オブジェクト。

| 値 | 日本語名 | 説明 | シフト確定権限 |
|---|---|---|---|
| `owner` | 店長 | テナントの最終責任者 | ○ |
| `vice_owner` | 副店長 | 店長を補佐し、シフト確定業務を分担する | ○ |
| `cast` | キャスト | お客さまと直接接客する出演者 | × |
| `staff` | スタッフ | 受付・場内案内・撮影補助などの裏方 | × |

### メンバー状態（Member Status）

メンバーがシフト対象として扱われるかどうかの状態。

| 状態 | 日本語名 | 説明 |
|---|---|---|
| 在籍中（Active） | 在籍中 | 通常どおりシフト希望と配置の対象となる |
| 休止中（Inactive） | 休止中 | 一時的にシフト対象から外す |
| 退店（Retired） | 退店 | 過去の実績は保持するが、新規のシフト対象にはならない |

### 外部アカウント（External Account）

メンバーと紐づく外部サービスのアカウント情報。

- **英語コード**: `external_account`
- **役割**: Discord等の連絡手段との紐付け
- **例**: DiscordユーザーID、メールアドレス

---

## 3. シフト枠・ポジション領域

### ポジション（Position）

営業時に必要となる役割を表す。イベントごとに定義される。

- **英語コード**: `position`
- **識別子**: `PositionID`（ULID形式）
- **役割**: シフト枠が参照するポジション定義を管理
- **例**: 「IL（インスタンスリーダー）」「カウンター」「テーブル」「受付」「カメラ」

```go
type Position struct {
    positionID   PositionID
    tenantID     TenantID
    positionName string
    description  string
    displayOrder int
    isActive     bool
}
```

### シフト枠（Shift Slot）

特定のイベント営業日の「時間帯×インスタンス×ポジション」の組み合わせで定義される「1人分の席」。

- **英語コード**: `shift_slot`, `slot`
- **識別子**: `SlotID`（ULID形式）
- **役割**: シフト割り当ての対象となる枠を管理
- **例**: 「2025-11-13 21:30〜23:00 第一インスタンス カウンターA」

```go
type ShiftSlot struct {
    slotID        SlotID
    tenantID      TenantID
    businessDayID BusinessDayID
    positionID    PositionID
    slotName      string
    instanceName  string       // インスタンス名（第一インスタンス等）
    startTime     time.Time    // 開始時刻
    endTime       time.Time    // 終了時刻
    requiredCount int          // 必要人数
    priority      int          // 表示優先度
}
```

### インスタンス名（Instance Name）

VRChatのインスタンス（部屋）を識別する名前。

- **英語コード**: `instance_name`
- **例**: 「第一インスタンス」「第二インスタンス」

### 必要人数（Required Count）

シフト枠に配置が必要な人数。

- **英語コード**: `required_count`
- **制約**: 1以上の正の整数

### 営業枠テンプレート（Shift Template）

イベントごとに定義される、営業時のインスタンス構成とポジション構成のテンプレート。

- **英語コード**: `shift_template`
- **役割**: 営業日のシフト枠を生成する際の雛形
- **例**: 「九龍想定テンプレート（A〜D: カウンター、E〜H: テーブル）」

### 営業日インスタンス（Event Instance）

特定の営業日における、インスタンス単位の営業単位。

- **英語コード**: `event_instance`
- **役割**: 営業日×インスタンスで一意に決まる営業単位

---

## 4. シフト希望領域

### シフト希望（Shift Availability / Availability）

メンバーが申告する「出勤可能な日・時間帯・ポジションの希望」。

- **英語コード**: `shift_availability`, `availability`
- **識別子**: `AvailabilityID`（ULID形式）
- **役割**: シフト確定の入力情報として機能する
- **状態**: 提出済み / 取下げ

```go
type Availability struct {
    availabilityID AvailabilityID
    memberID       MemberID
    businessDayID  BusinessDayID
    status         AvailabilityStatus
    submittedAt    time.Time
}
```

### シフト希望詳細（Availability Detail）

シフト希望の詳細情報。希望ポジション・優先度を含む。

- **英語コード**: `availability_detail`
- **役割**: 特定のシフト枠に対する希望の詳細を管理

### 希望順位（Preference Rank）

ポジションの希望順位。

- **英語コード**: `preference_rank`
- **例**: 第1希望、第2希望、第3希望

### 希望の強さ（Preference Strength）

出勤希望の強さ。

| 値 | 日本語名 | 説明 |
|---|---|---|
| 必ず出たい | 強い希望 | 可能な限り配置してほしい |
| 出られたら出たい | 弱い希望 | 空きがあれば配置してほしい |

### 希望期間（Availability Period）

まとめてシフト希望を提出する単位の期間。従来の「調整さん」の1枚に相当。

- **英語コード**: `availability_period`
- **役割**: シフト希望の提出単位と締切を管理
- **例**: 「12月通常営業分」「Vket特別営業分」

### 提出期限（Submission Deadline）

シフト希望を提出できる期限日時。

- **英語コード**: `submission_deadline`
- **役割**: 期限を過ぎた希望は「期限後提出」としてマークされる

---

## 5. シフト確定・割り当て領域

### シフト案（Shift Draft）

シフト希望と営業枠をもとに作成される暫定的な配置案。

- **英語コード**: `shift_draft`
- **役割**: 運営が微調整する前段階の「叩き台」

### シフト確定（Shift Plan）

運営が作成する最終的な配置計画。

- **英語コード**: `shift_plan`
- **識別子**: `PlanID`（ULID形式）
- **役割**: 特定の営業日に対するシフト割り当て全体を束ね、状態遷移と整合性を保証

```go
type ShiftPlan struct {
    planID        PlanID
    businessDayID BusinessDayID
    status        ShiftPlanStatus
    confirmedAt   *time.Time
}
```

### シフト確定状態（Shift Plan Status）

シフト確定のライフサイクルを表す区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `tentative` | 仮確定 | シフト割り当てが完了したが、まだメンバーへの正式通知前 |
| `confirmed` | 確定 | シフト割り当てが確定し、メンバーに通知済み |

### シフト割り当て（Shift Assignment）

シフト枠に対するメンバーの配置情報。

- **英語コード**: `shift_assignment`, `assignment`
- **識別子**: `AssignmentID`（ULID形式）
- **役割**: 「誰がどの枠に入るか」を管理

```go
type ShiftAssignment struct {
    assignmentID        AssignmentID
    tenantID            TenantID
    planID              PlanID
    slotID              SlotID
    memberID            MemberID
    assignmentStatus    AssignmentStatus
    assignmentMethod    AssignmentMethod
    isOutsidePreference bool           // 希望外配置フラグ
    assignedAt          time.Time
    cancelledAt         *time.Time
}
```

### 割り当て状態（Assignment Status）

シフト割り当ての状態を表す区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `confirmed` | 確定 | 有効な割り当て |
| `cancelled` | キャンセル | メンバーがキャンセルした割り当て |

### 割り当て方法（Assignment Method）

シフト割り当てがどのように行われたかを表す区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `auto` | 自動割り当て | システムが自動で割り当て |
| `manual` | 手動割り当て | 管理者が手動で割り当て |

### 希望外配置（Outside Preference）

メンバーの希望範囲外での配置。運営とメンバーの合意が前提。

- **英語コード**: `is_outside_preference`
- **フラグ**: `true` = 希望外配置

### 充足性（Fulfillment）

シフト枠の必要人数に対する配置済み人数の状態。

| 状態 | 説明 |
|---|---|
| 充足 | 配置済み人数 = 必要人数 |
| 不足 | 配置済み人数 < 必要人数 |
| 過剰配置 | 配置済み人数 > 必要人数 |

---

## 6. 出欠確認・日程調整領域

### 出欠確認（Attendance Collection）

確定した営業日に対するメンバーの出欠を収集する機能単位。「調整さん」のような公開URLを発行し、認証不要で回答できる。

- **英語コード**: `attendance_collection`
- **識別子**: `CollectionID`（ULID形式）
- **役割**: 営業日またはイベントに対する出欠の収集・集計を管理
- **例**: 「1月7日営業 出欠確認」

```go
type AttendanceCollection struct {
    collectionID CollectionID
    tenantID     TenantID
    title        string              // タイトル
    description  string              // 説明文
    targetType   AttendanceTargetType // 対象種別（event/business_day）
    targetID     string              // 対象イベントIDまたは営業日ID
    publicToken  string              // 公開URL用トークン
    status       AttendanceCollectionStatus
    deadline     *time.Time          // 回答締切
    createdAt    time.Time
}
```

### 出欠確認対象種別（Attendance Target Type）

出欠確認の対象を区別する区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `event` | イベント | イベント全体への出欠（複数営業日をまとめて） |
| `business_day` | 営業日 | 特定の営業日への出欠 |

### 出欠確認状態（Attendance Collection Status）

出欠確認のライフサイクルを表す区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `open` | 回答受付中 | 回答を受け付けている状態 |
| `closed` | 締切済み | 回答を締め切った状態 |

### 出欠回答（Attendance Response）

メンバーが提出する出欠の回答。

- **英語コード**: `attendance_response`
- **識別子**: `ResponseID`（ULID形式）
- **役割**: 個別メンバーの出欠回答を管理

```go
type AttendanceResponse struct {
    responseID    ResponseID
    collectionID  CollectionID
    memberID      MemberID
    response      AttendanceResponseType
    note          string
    respondedAt   time.Time
}
```

### 出欠回答種別（Attendance Response Type）

出欠の回答内容を表す区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `attending` | 出席 | 参加できる |
| `absent` | 欠席 | 参加できない |

### 日程調整（Date Schedule）

複数の候補日を提示し、メンバーの参加可否を収集してイベント開催日を決定する機能単位。

- **英語コード**: `date_schedule`
- **識別子**: `ScheduleID`（ULID形式）
- **役割**: 候補日の管理・回答収集・集計・開催日決定を管理
- **例**: 「2月特別イベント 日程調整」

```go
type DateSchedule struct {
    scheduleID   ScheduleID
    tenantID     TenantID
    title        string              // タイトル
    description  string              // 説明文
    eventID      *EventID            // 関連イベントID（任意）
    publicToken  string              // 公開URL用トークン
    status       DateScheduleStatus
    deadline     *time.Time          // 回答締切
    decidedDate  *CandidateDateID    // 決定した候補日ID
    createdAt    time.Time
}
```

### 日程調整状態（Date Schedule Status）

日程調整のライフサイクルを表す区分オブジェクト。

| 値 | 日本語名 | 説明 |
|---|---|---|
| `open` | 回答受付中 | 回答を受け付けている状態 |
| `closed` | 締切済み | 回答を締め切ったが、開催日未決定 |
| `decided` | 決定済み | 開催日が決定した状態 |

### 候補日（Candidate Date）

日程調整における選択肢となる日付・時間帯。

- **英語コード**: `candidate_date`
- **識別子**: `CandidateDateID`（ULID形式）
- **役割**: 日程調整の選択肢を管理

```go
type CandidateDate struct {
    candidateID  CandidateDateID
    scheduleID   ScheduleID
    date         time.Time           // 候補日（DATE）
    startTime    *time.Time          // 開始時刻（任意）
    endTime      *time.Time          // 終了時刻（任意）
    note         string              // 備考
    displayOrder int                 // 表示順
}
```

### 日程調整回答（Date Schedule Response）

メンバーが提出する各候補日への参加可否の回答。

- **英語コード**: `date_schedule_response`
- **識別子**: `DateResponseID`（ULID形式）
- **役割**: 個別メンバーの候補日ごとの回答を管理

```go
type DateScheduleResponse struct {
    dateResponseID DateResponseID
    scheduleID     ScheduleID
    memberID       MemberID
    candidateID    CandidateDateID
    availability   DateAvailabilityType
    respondedAt    time.Time
}
```

### 日程可否種別（Date Availability Type）

候補日への参加可否を表す区分オブジェクト。

| 値 | 日本語名 | 説明 | 表示記号 |
|---|---|---|---|
| `available` | 参加可能 | その日は参加できる | ○ |
| `unavailable` | 参加不可 | その日は参加できない | × |
| `maybe` | 未定・調整中 | 参加できるかもしれない | △ |

### 公開回答ページ（Public Response Page）

出欠確認・日程調整で共通して使用する、認証不要の公開ページ。

- **英語コード**: `public_response_page`
- **役割**: Discord等で共有されるURL先のページとして機能
- **アクセス**: トークン付きURLでアクセス（認証不要）
- **特徴**: 
  - 回答者はプルダウンから自分（登録済みメンバー）を選択
  - モバイルフレンドリー
  - 締切管理

### 公開トークン（Public Token）

公開回答ページへのアクセスに使用する一意の文字列。

- **英語コード**: `public_token`
- **役割**: URLの一部として使用され、認証なしでページを特定
- **形式**: UUID v4（RFC 4122準拠、36文字、ハイフン含む）
- **例**: `550e8400-e29b-41d4-a716-446655440000`
- **DB型**: PostgreSQL UUID型 + UNIQUE制約
- **バリデーション**: UUID形式でないトークンは400 Bad Requestを返す

### 回答締切（Response Deadline）

出欠確認・日程調整の回答を受け付ける期限。

- **英語コード**: `deadline`, `response_deadline`
- **役割**: 締切を過ぎた回答を拒否
- **処理**: 締切後は回答フォームを非表示にし、メッセージのみ表示

---

### 通知（Notification）

シフトに関する出来事をメンバーに伝えるメッセージ。

- **英語コード**: `notification`
- **識別子**: `NotificationID`（ULID形式）

### 通知種別（Notification Type）

| 種別 | 日本語名 | 説明 |
|---|---|---|
| シフト募集開始通知 | 募集開始 | 希望提出の受付開始を通知 |
| シフト提出締切リマインド | 締切リマインド | 提出期限が近づいていることを通知 |
| シフト確定通知 | 確定通知 | シフトが確定したことを通知 |
| 出勤リマインド | 出勤リマインド | 営業日前日/当日の出勤確認 |
| 補欠募集通知 | 緊急ヘルプ | 欠員発生時の補欠募集 |

### 通知対象（Notification Target）

通知を受け取る対象のグループ。

- **英語コード**: `notification_target`
- **例**: テナント全員、特定イベントのキャスト、確定シフトに入っているメンバー

### 通知テンプレート（Notification Template）

通知メッセージの雛形。

- **英語コード**: `notification_template`
- **役割**: テナントごとにカスタマイズ可能な通知文面を管理

### 通知送信履歴（Notification Log）

送信された通知の記録。

- **英語コード**: `notification_log`
- **役割**: 送信状態・送信日時を記録

---

## 8. 監査・履歴領域

### 監査ログ（Audit Log）

シフト管理における重要な操作の履歴。

- **英語コード**: `audit_log`
- **識別子**: `AuditLogID`（ULID形式）
- **保持期間**: 1年（設定可能）

```go
type AuditLog struct {
    auditLogID     AuditLogID
    tenantID       TenantID
    entityType     string         // 操作対象エンティティ種別
    entityID       string         // 操作対象エンティティID
    operationType  string         // 操作種別（CREATE/UPDATE/DELETE）
    actorID        MemberID       // 操作者
    beforeData     *string        // 変更前データ（JSON）
    afterData      *string        // 変更後データ（JSON）
    operatedAt     time.Time      // 操作日時
}
```

### 監査対象（Audit Subject）

監査イベントの対象となる業務オブジェクト。

| 対象 | 説明 |
|---|---|
| シフト希望 | 新規作成・修正・削除 |
| シフト枠 | 作成・変更・削除 |
| シフト確定 | 状態変更（仮確定→確定） |
| シフト割り当て | 作成・キャンセル |
| 営業枠テンプレート | 設定変更 |
| イベント設定 | 設定変更 |

### 操作種別（Operation Type）

| 値 | 日本語名 |
|---|---|
| `CREATE` | 作成 |
| `UPDATE` | 更新 |
| `DELETE` | 削除 |

---

## 9. 共通概念

### ULID

Universally Unique Lexicographically Sortable Identifier。
全てのエンティティの識別子に使用される26文字の一意識別子。

- **特徴**: 時系列でソート可能、衝突しにくい
- **形式**: 26文字の英数字

### テナント境界（Tenant Boundary）

全てのシフト関連データは必ず1つのテナントに属し、テナント間でのデータ参照・変更は禁止される。

### ソフトデリート（Soft Delete）

データを物理削除せず、`deleted_at` タイムスタンプで論理削除する方式。

- **フィールド**: `deleted_at`
- **判定**: `deleted_at IS NOT NULL` で削除済み

### 有効期間（Valid Period）

営業日やシフト枠などの有効期間。

- **フィールド**: `valid_from`, `valid_to`
- **制約**: 両方セットするか、両方NULLにする

### 深夜営業（Overnight）

日付を跨ぐ営業。終了時刻が開始時刻より前になる。

- **判定**: `end_time < start_time`
- **例**: 21:30〜25:00（翌1:00）

---

## 10. 状態・区分値 一覧

### イベント種別（EventType）
- `normal`: 通常営業
- `special`: 特別営業

### 発生種別（OccurrenceType）
- `recurring`: 定期営業
- `special`: 特別営業

### パターン種別（PatternType）
- `weekly`: 曜日指定
- `monthly_date`: 月内日付指定
- `custom`: カスタム

### ロール種別（RoleType）
- `owner`: 店長
- `vice_owner`: 副店長
- `cast`: キャスト
- `staff`: スタッフ

### シフト確定状態（ShiftPlanStatus）
- `tentative`: 仮確定
- `confirmed`: 確定

### 割り当て状態（AssignmentStatus）
- `confirmed`: 確定
- `cancelled`: キャンセル

### 割り当て方法（AssignmentMethod）
- `auto`: 自動割り当て
- `manual`: 手動割り当て

### 曜日（DayOfWeek）
- `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`

### 出欠確認対象種別（AttendanceTargetType）
- `event`: イベント
- `business_day`: 営業日

### 出欠確認状態（AttendanceCollectionStatus）
- `open`: 回答受付中
- `closed`: 締切済み

### 出欠回答種別（AttendanceResponseType）
- `attending`: 出席
- `absent`: 欠席
- `maybe`: 未定・調整中（△）

### 日程調整状態（DateScheduleStatus）
- `open`: 回答受付中
- `closed`: 締切済み
- `decided`: 決定済み

### 日程可否種別（DateAvailabilityType）
- `available`: 参加可能（○）
- `unavailable`: 参加不可（×）
- `maybe`: 未定・調整中（△）

---

## 用語対応表

| 日本語 | 英語（コード） | 説明 |
|---|---|---|
| テナント | Tenant | 団体・店舗単位 |
| イベント | Event | 営業・イベント単位 |
| 通常営業 | Normal Event | 反復パターンの営業 |
| 特別営業 | Special Event | 単発の営業 |
| 営業日 | Business Day | 1回分の営業日 |
| 通常営業パターン | Recurring Pattern | 反復ルール定義 |
| メンバー | Member | 所属する人物 |
| 店長 | Owner | 最終責任者 |
| 副店長 | Vice Owner | 店長補佐 |
| キャスト | Cast | 接客担当 |
| スタッフ | Staff | 裏方担当 |
| ポジション | Position | 役割（カウンター等） |
| シフト枠 | Shift Slot | 1人分の配置枠 |
| インスタンス | Instance | VRChatの部屋単位 |
| シフト希望 | Availability | 出勤可能申告 |
| 希望期間 | Availability Period | 希望提出の単位期間 |
| シフト案 | Shift Draft | 暫定配置案 |
| シフト確定 | Shift Plan | 最終配置計画 |
| シフト割り当て | Shift Assignment | 枠への人員配置 |
| 仮確定 | Tentative | 通知前の確定状態 |
| 確定 | Confirmed | 通知済みの確定状態 |
| 希望外配置 | Outside Preference | 希望範囲外の配置 |
| 出欠確認 | Attendance Collection | 営業日への出欠収集 |
| 出欠回答 | Attendance Response | 出欠への回答 |
| 出席 | Attending | 参加する |
| 欠席 | Absent | 参加しない |
| 日程調整 | Date Schedule | 候補日から開催日を決める |
| 候補日 | Candidate Date | 日程調整の選択肢 |
| 参加可能 | Available | 候補日に参加できる（○） |
| 参加不可 | Unavailable | 候補日に参加できない（×） |
| 未定 | Maybe | 参加できるかもしれない（△） |
| 公開回答ページ | Public Response Page | 認証不要の回答ページ |
| 公開トークン | Public Token | URLアクセス用の一意文字列 |
| 回答締切 | Response Deadline | 回答の受付期限 |
| 通知 | Notification | メンバーへの情報伝達 |
| リマインド | Reminder | 事前の確認通知 |
| 監査ログ | Audit Log | 操作履歴 |
| 必要人数 | Required Count | 枠の必要配置人数 |
| 充足 | Fulfillment | 必要人数を満たす状態 |

---

## 更新履歴

| 日付 | 更新内容 |
|---|---|
| 2025-12-14 | 初版作成 |
| 2025-12-14 | 出欠確認・日程調整領域を追加 |

