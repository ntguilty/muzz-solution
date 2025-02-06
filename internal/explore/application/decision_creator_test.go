package application

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockDecisionCreatorRepo struct {
	insertDecision func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error)
}

func (m *mockDecisionCreatorRepo) InsertDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
	return m.insertDecision(ctx, actorID, recipientID, liked)
}

func TestDecisionCreator_SaveDecision(t *testing.T) {
	tests := []struct {
		name         string
		actorID      string
		recipientID  string
		liked        bool
		mockBehavior func(*mockDecisionCreatorRepo)
		wantMutual   bool
		wantErr      error
	}{
		{
			name:        "success - mutual like",
			actorID:     "user1",
			recipientID: "user2",
			liked:       true,
			mockBehavior: func(m *mockDecisionCreatorRepo) {
				m.insertDecision = func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
					return true, nil
				}
			},
			wantMutual: true,
			wantErr:    nil,
		},
		{
			name:        "success - no mutual like",
			actorID:     "user1",
			recipientID: "user2",
			liked:       true,
			mockBehavior: func(m *mockDecisionCreatorRepo) {
				m.insertDecision = func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
					return false, nil
				}
			},
			wantMutual: false,
			wantErr:    nil,
		},
		{
			name:        "error - repository error",
			actorID:     "user1",
			recipientID: "user2",
			liked:       true,
			mockBehavior: func(m *mockDecisionCreatorRepo) {
				m.insertDecision = func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
					return false, errors.New("db error")
				}
			},
			wantMutual: false,
			wantErr:    errors.New("failed to save decision: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDecisionCreatorRepo{}
			tt.mockBehavior(mockRepo)

			creator := NewDecisionCreator(mockRepo)
			gotMutual, err := creator.SaveDecision(context.Background(), tt.actorID, tt.recipientID, tt.liked)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMutual, gotMutual)
			}
		})
	}
}
