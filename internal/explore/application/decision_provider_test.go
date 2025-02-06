package application

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"muzz-homework/internal/explore/domain"
	"testing"
)

type mockDecisionProviderRepo struct {
	getLikers      func(ctx context.Context, recipientID string, paginationToken *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error)
	getLikersCount func(ctx context.Context, recipientID string) (uint64, error)
}

func (m *mockDecisionProviderRepo) GetLikers(ctx context.Context, recipientID string, paginationToken *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
	return m.getLikers(ctx, recipientID, paginationToken, excludeMutual)
}

func (m *mockDecisionProviderRepo) GetLikersCount(ctx context.Context, recipientID string) (uint64, error) {
	return m.getLikersCount(ctx, recipientID)
}

type mockCacheRepo struct {
	getLikers      func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error)
	setLikers      func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool, likers []domain.LikerInfo, nextTS *uint64) error
	getLikersCount func(ctx context.Context, recipientID string) (uint64, error)
	setLikersCount func(ctx context.Context, recipientID string, count uint64) error
}

func (m *mockCacheRepo) GetLikers(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
	return m.getLikers(ctx, recipientID, timestamp, excludeMutual)
}

func (m *mockCacheRepo) SetLikers(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool, likers []domain.LikerInfo, nextTS *uint64) error {
	return m.setLikers(ctx, recipientID, timestamp, excludeMutual, likers, nextTS)
}

func (m *mockCacheRepo) GetLikersCount(ctx context.Context, recipientID string) (uint64, error) {
	return m.getLikersCount(ctx, recipientID)
}

func (m *mockCacheRepo) SetLikersCount(ctx context.Context, recipientID string, count uint64) error {
	return m.setLikersCount(ctx, recipientID, count)
}

func TestDecisionProvider_ListLikedYou(t *testing.T) {
	tests := []struct {
		name          string
		recipientID   string
		encodedToken  string
		setCache      bool
		mockBehavior  func(*mockDecisionProviderRepo, *mockCacheRepo)
		wantLikers    []domain.LikerInfo
		wantNextToken string
		wantErr       error
	}{
		{
			name:         "success - from cache",
			recipientID:  "user1",
			encodedToken: "",
			setCache:     true,
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {
				mc.getLikers = func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
					return []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}}, uint64Ptr(123456), nil
				}
			},
			wantLikers:    []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}},
			wantNextToken: "eyJ0IjoxMjM0NTZ9",
			wantErr:       nil,
		},
		{
			name:         "success - from db",
			recipientID:  "user1",
			encodedToken: "",
			setCache:     false,
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {
				mc.getLikers = func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
					return nil, nil, errors.New("cache miss")
				}
				mr.getLikers = func(ctx context.Context, recipientID string, paginationToken *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
					return []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}}, uint64Ptr(123456), nil
				}
				mc.setLikers = func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool, likers []domain.LikerInfo, nextTS *uint64) error {
					return nil
				}
			},
			wantLikers:    []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}},
			wantNextToken: "eyJ0IjoxMjM0NTZ9",
			wantErr:       nil,
		},
		{
			name:         "error - empty recipient ID",
			recipientID:  "",
			encodedToken: "",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {},
			wantLikers:   nil,
			wantErr:      domain.ErrInvalidInput,
		},
		{
			name:         "error - invalid token",
			recipientID:  "user1",
			encodedToken: "invalid-token",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {},
			wantLikers:   nil,
			wantErr:      errors.New("invalid pagination token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDecisionProviderRepo{}
			mockCache := &mockCacheRepo{}
			tt.mockBehavior(mockRepo, mockCache)

			provider := NewDecisionProvider(mockRepo, mockCache)
			gotLikers, gotNextToken, err := provider.ListLikedYou(context.Background(), tt.recipientID, tt.encodedToken)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLikers, gotLikers)
				assert.Equal(t, tt.wantNextToken, gotNextToken)
			}
		})
	}
}

func TestDecisionProvider_ListNewLikedYou(t *testing.T) {
	tests := []struct {
		name          string
		recipientID   string
		encodedToken  string
		setCache      bool
		mockBehavior  func(*mockDecisionProviderRepo, *mockCacheRepo)
		wantLikers    []domain.LikerInfo
		wantNextToken string
		wantErr       error
	}{
		{
			name:         "success - from cache",
			recipientID:  "user1",
			encodedToken: "",
			setCache:     true,
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {
				mc.getLikers = func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
					assert.True(t, excludeMutual)
					return []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}}, uint64Ptr(123456), nil
				}
			},
			wantLikers:    []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}},
			wantNextToken: "eyJ0IjoxMjM0NTZ9",
			wantErr:       nil,
		},
		{
			name:         "success - from db",
			recipientID:  "user1",
			encodedToken: "",
			setCache:     false,
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {
				mc.getLikers = func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
					return nil, nil, errors.New("cache miss")
				}
				mr.getLikers = func(ctx context.Context, recipientID string, paginationToken *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
					assert.True(t, excludeMutual)
					return []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}}, uint64Ptr(123456), nil
				}
				mc.setLikers = func(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool, likers []domain.LikerInfo, nextTS *uint64) error {
					return nil
				}
			},
			wantLikers:    []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}},
			wantNextToken: "eyJ0IjoxMjM0NTZ9",
			wantErr:       nil,
		},
		{
			name:         "error - empty recipient ID",
			recipientID:  "",
			encodedToken: "",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {},
			wantLikers:   nil,
			wantErr:      domain.ErrInvalidInput,
		},
		{
			name:         "error - invalid token",
			recipientID:  "user1",
			encodedToken: "invalid-token",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {},
			wantLikers:   nil,
			wantErr:      errors.New("invalid pagination token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDecisionProviderRepo{}
			mockCache := &mockCacheRepo{}
			tt.mockBehavior(mockRepo, mockCache)

			provider := NewDecisionProvider(mockRepo, mockCache)
			gotLikers, gotNextToken, err := provider.ListNewLikedYou(context.Background(), tt.recipientID, tt.encodedToken)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLikers, gotLikers)
				assert.Equal(t, tt.wantNextToken, gotNextToken)
			}
		})
	}
}

func TestDecisionProvider_CountLikedYou(t *testing.T) {
	tests := []struct {
		name         string
		recipientID  string
		mockBehavior func(*mockDecisionProviderRepo, *mockCacheRepo)
		wantCount    uint64
		wantErr      error
	}{
		{
			name:        "success - from cache",
			recipientID: "user1",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {
				mc.getLikersCount = func(ctx context.Context, recipientID string) (uint64, error) {
					return 42, nil
				}
			},
			wantCount: 42,
			wantErr:   nil,
		},
		{
			name:        "success - from db",
			recipientID: "user1",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {
				mc.getLikersCount = func(ctx context.Context, recipientID string) (uint64, error) {
					return 0, errors.New("cache miss")
				}
				mr.getLikersCount = func(ctx context.Context, recipientID string) (uint64, error) {
					return 42, nil
				}
				mc.setLikersCount = func(ctx context.Context, recipientID string, count uint64) error {
					return nil
				}
			},
			wantCount: 42,
			wantErr:   nil,
		},
		{
			name:         "error - empty recipient ID",
			recipientID:  "",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {},
			wantCount:    0,
			wantErr:      domain.ErrInvalidInput,
		},
		{
			name:        "error - db error",
			recipientID: "user1",
			mockBehavior: func(mr *mockDecisionProviderRepo, mc *mockCacheRepo) {
				mc.getLikersCount = func(ctx context.Context, recipientID string) (uint64, error) {
					return 0, errors.New("cache miss")
				}
				mr.getLikersCount = func(ctx context.Context, recipientID string) (uint64, error) {
					return 0, errors.New("db error")
				}
			},
			wantCount: 0,
			wantErr:   errors.New("failed to count likers: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDecisionProviderRepo{}
			mockCache := &mockCacheRepo{}
			tt.mockBehavior(mockRepo, mockCache)

			provider := NewDecisionProvider(mockRepo, mockCache)
			gotCount, err := provider.CountLikedYou(context.Background(), tt.recipientID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, gotCount)
			}
		})
	}
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}
