package application

import (
	"context"
	"fmt"
	"muzz-homework/internal/explore/domain"
)

type decisionProviderRepository interface {
	GetLikers(ctx context.Context, recipientID string, paginationToken *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error)
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

func (p *DecisionProvider) ListLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error) {
	timestamp, err := domain.DecodePaginationToken(encodedToken)
	if err != nil {
		return nil, "", fmt.Errorf("invalid pagination token: %w", err)
	}

	likers, nextTimestamp, err := p.repo.GetLikers(ctx, recipientID, timestamp, false)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list likers: %w", err)
	}

	var nextToken string
	if nextTimestamp != nil {
		nextToken = domain.EncodePaginationToken(*nextTimestamp)
	}

	return likers, nextToken, nil
}

func (p *DecisionProvider) ListNewLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error) {
	timestamp, err := domain.DecodePaginationToken(encodedToken)
	if err != nil {
		return nil, "", fmt.Errorf("invalid pagination token: %w", err)
	}

	likers, nextTimestamp, err := p.repo.GetLikers(ctx, recipientID, timestamp, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list new likers: %w", err)
	}

	var nextToken string
	if nextTimestamp != nil {
		nextToken = domain.EncodePaginationToken(*nextTimestamp)
	}

	return likers, nextToken, nil
}

func (p *DecisionProvider) CountLikedYou(ctx context.Context, recipientID string) (uint64, error) {
	count, err := p.repo.GetLikersCount(ctx, recipientID)
	if err != nil {
		return 0, fmt.Errorf("failed to count likers: %w", err)
	}

	return count, nil
}
