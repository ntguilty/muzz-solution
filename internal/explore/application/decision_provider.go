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

type cacheRepository interface {
	GetLikers(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error)
	SetLikers(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool, likers []domain.LikerInfo, nextTS *uint64) error
	GetLikersCount(ctx context.Context, recipientID string) (uint64, error)
	SetLikersCount(ctx context.Context, recipientID string, count uint64) error
}

type DecisionProvider struct {
	repo  decisionProviderRepository
	cache cacheRepository
}

func NewDecisionProvider(repo decisionProviderRepository, cache cacheRepository) *DecisionProvider {
	return &DecisionProvider{
		repo:  repo,
		cache: cache,
	}
}

func (p *DecisionProvider) ListLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error) {
	if recipientID == "" {
		return nil, "", domain.ErrInvalidInput
	}

	timestamp, err := domain.DecodePaginationToken(encodedToken)
	if err != nil {
		return nil, "", fmt.Errorf("invalid pagination token: %w", err)
	}

	likers, nextTimestamp, err := p.cache.GetLikers(ctx, recipientID, timestamp, false)
	if err == nil {
		var nextToken string
		if nextTimestamp != nil {
			nextToken = domain.EncodePaginationToken(*nextTimestamp)
		}
		return likers, nextToken, nil
	}

	likers, nextTimestamp, err = p.repo.GetLikers(ctx, recipientID, timestamp, false)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list likers: %w", err)
	}

	p.cache.SetLikers(ctx, recipientID, timestamp, false, likers, nextTimestamp)

	var nextToken string
	if nextTimestamp != nil {
		nextToken = domain.EncodePaginationToken(*nextTimestamp)
	}

	return likers, nextToken, nil
}

func (p *DecisionProvider) ListNewLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error) {
	if recipientID == "" {
		return nil, "", domain.ErrInvalidInput
	}

	timestamp, err := domain.DecodePaginationToken(encodedToken)
	if err != nil {
		return nil, "", fmt.Errorf("invalid pagination token: %w", err)
	}

	likers, nextTimestamp, err := p.cache.GetLikers(ctx, recipientID, timestamp, true)
	if err == nil {
		var nextToken string
		if nextTimestamp != nil {
			nextToken = domain.EncodePaginationToken(*nextTimestamp)
		}
		return likers, nextToken, nil
	}

	likers, nextTimestamp, err = p.repo.GetLikers(ctx, recipientID, timestamp, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list new likers: %w", err)
	}

	p.cache.SetLikers(ctx, recipientID, timestamp, true, likers, nextTimestamp)

	var nextToken string
	if nextTimestamp != nil {
		nextToken = domain.EncodePaginationToken(*nextTimestamp)
	}

	return likers, nextToken, nil
}

func (p *DecisionProvider) CountLikedYou(ctx context.Context, recipientID string) (uint64, error) {
	if recipientID == "" {
		return 0, domain.ErrInvalidInput
	}

	count, err := p.cache.GetLikersCount(ctx, recipientID)
	if err == nil {
		return count, nil
	}

	count, err = p.repo.GetLikersCount(ctx, recipientID)
	if err != nil {
		return 0, fmt.Errorf("failed to count likers: %w", err)
	}

	p.cache.SetLikersCount(ctx, recipientID, count)

	return count, nil
}
