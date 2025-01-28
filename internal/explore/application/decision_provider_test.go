package application

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"muzz-homework/internal/explore/domain"
	"testing"
)

type mockDecisionProviderRepo struct {
	getLikers      func(ctx context.Context, recipientID string, paginationToken *string, excludeMutual bool) ([]domain.LikerInfo, string, error)
	getLikersCount func(ctx context.Context, recipientID string) (uint64, error)
}

func (m *mockDecisionProviderRepo) GetLikers(ctx context.Context, recipientID string, paginationToken *string, excludeMutual bool) ([]domain.LikerInfo, string, error) {
	return m.getLikers(ctx, recipientID, paginationToken, excludeMutual)
}

func (m *mockDecisionProviderRepo) GetLikersCount(ctx context.Context, recipientID string) (uint64, error) {
	return m.getLikersCount(ctx, recipientID)
}

func TestDecisionProvider_ListLikedYou(t *testing.T) {
	tests := []struct {
		name            string
		recipientID     string
		paginationToken *string
		mockBehavior    func(*mockDecisionProviderRepo)
		wantLikers      []domain.LikerInfo
		wantNextToken   string
		wantErr         error
	}{
		{
			name:            "success",
			recipientID:     "user1",
			paginationToken: nil,
			mockBehavior: func(m *mockDecisionProviderRepo) {
				m.getLikers = func(ctx context.Context, recipientID string, paginationToken *string, excludeMutual bool) ([]domain.LikerInfo, string, error) {
					return []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}}, "next_token", nil
				}
			},
			wantLikers:    []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}},
			wantNextToken: "next_token",
			wantErr:       nil,
		},
		{
			name:          "error - empty recipient ID",
			recipientID:   "",
			mockBehavior:  func(m *mockDecisionProviderRepo) {},
			wantLikers:    nil,
			wantNextToken: "",
			wantErr:       domain.ErrInvalidInput,
		},
		{
			name:        "error - repository error",
			recipientID: "user1",
			mockBehavior: func(m *mockDecisionProviderRepo) {
				m.getLikers = func(ctx context.Context, recipientID string, paginationToken *string, excludeMutual bool) ([]domain.LikerInfo, string, error) {
					return nil, "", errors.New("db error")
				}
			},
			wantLikers:    nil,
			wantNextToken: "",
			wantErr:       errors.New("failed to list likers: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDecisionProviderRepo{}
			tt.mockBehavior(mockRepo)

			provider := NewDecisionProvider(mockRepo)
			gotLikers, gotNextToken, err := provider.ListLikedYou(context.Background(), tt.recipientID, tt.paginationToken)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
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
		mockBehavior func(*mockDecisionProviderRepo)
		wantCount    uint64
		wantErr      error
	}{
		{
			name:        "success",
			recipientID: "user1",
			mockBehavior: func(m *mockDecisionProviderRepo) {
				m.getLikersCount = func(ctx context.Context, recipientID string) (uint64, error) {
					return 42, nil
				}
			},
			wantCount: 42,
			wantErr:   nil,
		},
		{
			name:         "error - empty recipient ID",
			recipientID:  "",
			mockBehavior: func(m *mockDecisionProviderRepo) {},
			wantCount:    0,
			wantErr:      domain.ErrInvalidInput,
		},
		{
			name:        "error - repository error",
			recipientID: "user1",
			mockBehavior: func(m *mockDecisionProviderRepo) {
				m.getLikersCount = func(ctx context.Context, recipientID string) (uint64, error) {
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
			tt.mockBehavior(mockRepo)

			provider := NewDecisionProvider(mockRepo)
			gotCount, err := provider.CountLikedYou(context.Background(), tt.recipientID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, gotCount)
			}
		})
	}
}

func TestDecisionProvider_ListNewLikedYou(t *testing.T) {
	tests := []struct {
		name            string
		recipientID     string
		paginationToken *string
		mockBehavior    func(*mockDecisionProviderRepo)
		wantLikers      []domain.LikerInfo
		wantNextToken   string
		wantErr         error
	}{
		{
			name:        "success",
			recipientID: "user1",
			mockBehavior: func(m *mockDecisionProviderRepo) {
				m.getLikers = func(ctx context.Context, recipientID string, paginationToken *string, excludeMutual bool) ([]domain.LikerInfo, string, error) {
					assert.True(t, excludeMutual)
					return []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}}, "next_token", nil
				}
			},
			wantLikers:    []domain.LikerInfo{{ActorID: "user2", Timestamp: 123456}},
			wantNextToken: "next_token",
			wantErr:       nil,
		},
		{
			name:          "error - empty recipient ID",
			recipientID:   "",
			mockBehavior:  func(m *mockDecisionProviderRepo) {},
			wantLikers:    nil,
			wantNextToken: "",
			wantErr:       domain.ErrInvalidInput,
		},
		{
			name:        "error - repository error",
			recipientID: "user1",
			mockBehavior: func(m *mockDecisionProviderRepo) {
				m.getLikers = func(ctx context.Context, recipientID string, paginationToken *string, excludeMutual bool) ([]domain.LikerInfo, string, error) {
					return nil, "", errors.New("db error")
				}
			},
			wantLikers:    nil,
			wantNextToken: "",
			wantErr:       errors.New("failed to list new likers: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDecisionProviderRepo{}
			tt.mockBehavior(mockRepo)

			provider := NewDecisionProvider(mockRepo)
			gotLikers, gotNextToken, err := provider.ListNewLikedYou(context.Background(), tt.recipientID, tt.paginationToken)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLikers, gotLikers)
				assert.Equal(t, tt.wantNextToken, gotNextToken)
			}
		})
	}
}
