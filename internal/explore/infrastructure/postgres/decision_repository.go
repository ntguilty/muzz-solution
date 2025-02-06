package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"muzz-homework/internal/explore/domain"
	"time"
)

const paginationLimit = 20

type decisionRepository struct {
	db *sql.DB
	sq sq.StatementBuilderType
}

func NewDecisionRepository(db *sql.DB) *decisionRepository {
	return &decisionRepository{
		db: db,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *decisionRepository) InsertDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
	timestamp := uint64(time.Now().Unix())

	_, err := r.sq.Insert("user_decisions").
		Columns("actor_user_id", "recipient_user_id", "liked_recipient", "decision_timestamp").
		Values(actorID, recipientID, liked, timestamp).
		Suffix(`
           ON CONFLICT (actor_user_id, recipient_user_id) 
           DO UPDATE SET 
               liked_recipient = EXCLUDED.liked_recipient,
               decision_timestamp = EXCLUDED.decision_timestamp`).
		RunWith(r.db).
		ExecContext(ctx)

	if err != nil {
		return false, fmt.Errorf("inserting decision: %w", err)
	}

	return true, nil
}

func (r *decisionRepository) GetLikers(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
	query := r.sq.Select("actor_user_id", "decision_timestamp").
		From("user_decisions").
		Where(sq.Eq{"recipient_user_id": recipientID, "liked_recipient": true})

	if excludeMutual {
		query = query.Where("NOT EXISTS (SELECT 1 FROM user_decisions ud2 WHERE "+
			"ud2.actor_user_id = ? AND "+
			"ud2.recipient_user_id = user_decisions.actor_user_id AND "+
			"ud2.liked_recipient = true)", recipientID)
	}

	if timestamp != nil {
		query = query.Where("decision_timestamp < ?", *timestamp)
	}

	query = query.OrderBy("decision_timestamp DESC").
		Limit(paginationLimit + 1)

	rows, err := query.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("selecting likers: %w", err)
	}
	defer rows.Close()

	var likers []domain.LikerInfo
	var lastTimestamp uint64
	var hasMore bool

	for rows.Next() {
		var liker domain.LikerInfo
		if err := rows.Scan(&liker.ActorID, &liker.Timestamp); err != nil {
			return nil, nil, fmt.Errorf("scanning liker: %w", err)
		}

		if len(likers) < paginationLimit {
			likers = append(likers, liker)
			lastTimestamp = liker.Timestamp
		} else {
			hasMore = true
			break
		}
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterating over likers: %w", err)
	}

	var nextTimestamp *uint64
	if hasMore {
		nextTimestamp = &lastTimestamp
	}

	return likers, nextTimestamp, nil
}

func (r *decisionRepository) GetLikersCount(ctx context.Context, recipientID string) (uint64, error) {
	var count uint64

	err := r.sq.Select("COUNT(*)").
		From("user_decisions").
		Where(sq.Eq{
			"recipient_user_id": recipientID,
			"liked_recipient":   true,
		}).
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("counting likers: %w", err)
	}

	return count, nil
}
