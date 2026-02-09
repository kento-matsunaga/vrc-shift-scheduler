package attendance

import (
	"context"
	"fmt"
	"log"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

type UpdateCollectionUsecase struct {
	repo      attendance.AttendanceCollectionRepository
	txManager services.TxManager
	clock     services.Clock
}

func NewUpdateCollectionUsecase(repo attendance.AttendanceCollectionRepository, txManager services.TxManager, clk services.Clock) *UpdateCollectionUsecase {
	return &UpdateCollectionUsecase{repo: repo, txManager: txManager, clock: clk}
}

func (u *UpdateCollectionUsecase) Execute(ctx context.Context, input UpdateCollectionInput) (*UpdateCollectionOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant ID のパースに失敗: %w", err)
	}

	collectionID, err := common.ParseCollectionID(input.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection ID のパースに失敗: %w", err)
	}

	collection, err := u.repo.FindByID(ctx, tenantID, collectionID)
	if err != nil {
		return nil, fmt.Errorf("出欠確認の取得に失敗: %w", err)
	}

	now := u.clock.Now()
	if err := collection.Update(now, input.Title, input.Description, input.Deadline); err != nil {
		return nil, fmt.Errorf("出欠確認の更新に失敗: %w", err)
	}

	// 対象日の更新がある場合はトランザクション内で処理
	if input.TargetDates != nil {
		err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
			if err := u.repo.Save(txCtx, collection); err != nil {
				return fmt.Errorf("出欠確認の保存に失敗: %w", err)
			}

			// 既存の対象日を取得（IDマッチング用）
			existingDates, err := u.repo.FindTargetDatesByCollectionID(txCtx, collectionID)
			if err != nil {
				return fmt.Errorf("既存対象日の取得に失敗: %w", err)
			}
			existingMap := make(map[string]*attendance.TargetDate)
			for _, ed := range existingDates {
				existingMap[ed.TargetDateID().String()] = ed
			}

			// 入力から TargetDate エンティティを構築
			// 既存 → ドメインメソッド UpdateFields で更新、新規 → NewTargetDate で作成
			allDates := make([]*attendance.TargetDate, 0, len(input.TargetDates))
			for i, td := range input.TargetDates {
				if td.TargetDateID != "" {
					// 既存対象日の更新 → ドメインメソッドでフィールドを更新
					existingTD, ok := existingMap[td.TargetDateID]
					if !ok {
						return fmt.Errorf("対象日ID %s が見つかりません", td.TargetDateID)
					}
					if err := existingTD.UpdateFields(td.TargetDate, td.StartTime, td.EndTime, i); err != nil {
						return fmt.Errorf("対象日の更新に失敗: %w", err)
					}
					allDates = append(allDates, existingTD)
				} else {
					// 新規対象日
					newTD, err := attendance.NewTargetDate(now, collectionID, td.TargetDate, td.StartTime, td.EndTime, i)
					if err != nil {
						return fmt.Errorf("対象日の作成に失敗: %w", err)
					}
					allDates = append(allDates, newTD)
				}
			}

			// 差分更新（既存IDの回答を保持、削除された対象日はCASCADEで回答も削除）
			if err := u.repo.ReplaceTargetDates(txCtx, collectionID, allDates); err != nil {
				return fmt.Errorf("対象日の更新に失敗: %w", err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// 対象日の更新なし → コレクション本体のみ保存
		if err := u.repo.Save(ctx, collection); err != nil {
			return nil, fmt.Errorf("出欠確認の保存に失敗: %w", err)
		}
	}

	log.Printf("[AUDIT] UpdateCollection: tenant=%s collection=%s", collection.TenantID().String(), collection.CollectionID().String())

	return &UpdateCollectionOutput{
		CollectionID: collection.CollectionID().String(),
		TenantID:     collection.TenantID().String(),
		Title:        collection.Title(),
		Description:  collection.Description(),
		Status:       collection.Status().String(),
		Deadline:     collection.Deadline(),
		UpdatedAt:    collection.UpdatedAt(),
	}, nil
}
