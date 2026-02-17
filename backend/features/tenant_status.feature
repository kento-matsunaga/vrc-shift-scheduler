# language: ja
@tenant @billing
Feature: テナント課金ステータス遷移
  テナントの課金ステータスは以下の4状態を持ち、
  ビジネスルールに基づいた遷移のみが許可される。

  状態一覧:
    - active:          有効（全機能利用可）
    - grace:           猶予期間（読み取りのみ、14日以内に再決済で復帰可）
    - suspended:       停止（読み取りのみ、再決済で復帰可）
    - pending_payment: 決済待ち（新規登録 or 再決済の決済完了待ち）

  # =====================================================
  # 正常な状態遷移
  # =====================================================

  Scenario: サブスク解約後、猶予期間に入る
    Given テナント「VRChat Japan」のステータスが「active」である
    When サブスクリプション期間が終了する（期間終了日: 2026-01-31）
    Then テナントのステータスが「grace」に遷移する
    And 猶予期限が「2026-02-14」に設定される（期間終了日 + 14日）
    And テナントのデータは読み取り可能である
    And テナントへの書き込みは拒否される

  Scenario: 猶予期間中に再決済して復帰する
    Given テナント「VRChat Japan」のステータスが「grace」である
    When 管理者が再決済を完了する
    Then テナントのステータスが「active」に遷移する
    And 猶予期限がクリアされる
    And テナントの全機能が利用可能になる

  Scenario: 猶予期間が切れてサービス停止になる
    Given テナント「VRChat Japan」のステータスが「grace」である
    When 猶予期間（14日間）が終了する
    Then テナントのステータスが「suspended」に遷移する
    And 猶予期限がクリアされる

  Scenario: 停止状態から再決済を開始する
    Given テナント「VRChat Japan」のステータスが「suspended」である
    When 管理者が再決済を開始する（Stripe Session: cs_test_xxx）
    Then テナントのステータスが「pending_payment」に遷移する
    And 決済セッションIDが保存される
    And 決済有効期限が設定される

  Scenario: 決済完了でアクティブに復帰する
    Given テナント「VRChat Japan」のステータスが「pending_payment」である
    When Stripe決済が完了する
    Then テナントのステータスが「active」に遷移する
    And 決済セッション情報がクリアされる
    And テナントの全機能が利用可能になる

  # =====================================================
  # 禁止された状態遷移
  # =====================================================

  Scenario: active から直接 pending_payment には遷移できない
    Given テナント「VRChat Japan」のステータスが「active」である
    When ステータスを「pending_payment」に変更しようとする
    Then エラー「invalid status transition from active to pending_payment」が返される
    And ステータスは「active」のまま変更されない

  Scenario: grace から pending_payment には遷移できない
    Given テナント「VRChat Japan」のステータスが「grace」である
    When ステータスを「pending_payment」に変更しようとする
    Then エラーが返される
    And ステータスは「grace」のまま変更されない

  Scenario: suspended から直接 grace には遷移できない
    Given テナント「VRChat Japan」のステータスが「suspended」である
    When ステータスを「grace」に変更しようとする
    Then エラーが返される
    And ステータスは「suspended」のまま変更されない

  # =====================================================
  # アクセス制御
  # =====================================================

  Scenario Outline: ステータスごとのアクセス制御
    Given テナントのステータスが「<status>」である
    Then データの読み取りは「<can_read>」である
    And データの書き込みは「<can_write>」である

    Examples:
      | status          | can_read | can_write |
      | active          | 可能     | 可能      |
      | grace           | 可能     | 不可      |
      | suspended       | 可能     | 不可      |
      | pending_payment | 不可     | 不可      |

  # =====================================================
  # ソフトデリート
  # =====================================================

  Scenario: 削除されたテナントは全操作が不可
    Given テナント「VRChat Japan」のステータスが「active」である
    When テナントが削除される（ソフトデリート）
    Then データの読み取りは不可である
    And データの書き込みは不可である
    And deleted_at が記録される
