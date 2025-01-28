package application

import (
	"context"
	"fmt"
	"muzz-homework/internal/explore/domain"
)

type decisionProviderRepository interface {
	GetLikers(ctx context.Context, recipientID string, paginationToken *string, excludeMutual bool) ([]domain.LikerInfo, string, error)
	GetLikersCount(ctx context.Context, recipientID string) (uint64, error)
}

type DecisionProvider struct {
	repo decisionProviderRepository
}

func NewDecisionProvider(decisionRepo decisionProviderRepository) *DecisionProvider {
	return &DecisionProvider{
		repo: decisionRepo,
	}
}

func (p *DecisionProvider) ListLikedYou(ctx context.Context, recipientID string, paginationToken *string) ([]domain.LikerInfo, string, error) {
	if recipientID == "" {
		return nil, "", domain.ErrInvalidInput
	}

	likers, nextToken, err := p.repo.GetLikers(ctx, recipientID, paginationToken, false)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list likers: %w", err)
	}

	return likers, nextToken, nil
}

func (p *DecisionProvider) ListNewLikedYou(ctx context.Context, recipientID string, paginationToken *string) ([]domain.LikerInfo, string, error) {
	if recipientID == "" {
		return nil, "", domain.ErrInvalidInput
	}

	likers, nextToken, err := p.repo.GetLikers(ctx, recipientID, paginationToken, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list new likers: %w", err)
	}

	return likers, nextToken, nil
}

func (p *DecisionProvider) CountLikedYou(ctx context.Context, recipientID string) (uint64, error) {
	if recipientID == "" {
		return 0, domain.ErrInvalidInput
	}

	count, err := p.repo.GetLikersCount(ctx, recipientID)
	if err != nil {
		return 0, fmt.Errorf("failed to count likers: %w", err)
	}

	return count, nil
}
