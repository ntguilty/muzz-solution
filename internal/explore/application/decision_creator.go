package application

import (
	"context"
	"fmt"
	"muzz-homework/internal/explore/domain"
)

type decisionCreatorRepository interface {
	InsertDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error)
}

type DecisionCreator struct {
	repo decisionCreatorRepository
}

func NewDecisionCreator(decisionRepo decisionCreatorRepository) *DecisionCreator {
	return &DecisionCreator{
		repo: decisionRepo,
	}
}
func (c *DecisionCreator) SaveDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
	if actorID == "" || recipientID == "" {
		return false, domain.ErrInvalidInput
	}

	mutualLike, err := c.repo.InsertDecision(ctx, actorID, recipientID, liked)
	if err != nil {
		return false, fmt.Errorf("failed to save decision: %w", err)
	}

	return mutualLike, nil
}
