package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"muzz-homework/internal/explore/domain"
	pb "muzz-homework/pkg/proto"
	"testing"
)

type mockDecisionProvider struct {
	listLikedYou    func(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error)
	listNewLikedYou func(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error)
	countLikedYou   func(ctx context.Context, recipientID string) (uint64, error)
}

func (m *mockDecisionProvider) ListLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error) {
	return m.listLikedYou(ctx, recipientID, encodedToken)
}

func (m *mockDecisionProvider) ListNewLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error) {
	return m.listNewLikedYou(ctx, recipientID, encodedToken)
}

func (m *mockDecisionProvider) CountLikedYou(ctx context.Context, recipientID string) (uint64, error) {
	return m.countLikedYou(ctx, recipientID)
}

type mockDecisionCreator struct {
	saveDecision func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error)
}

func (m *mockDecisionCreator) SaveDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
	return m.saveDecision(ctx, actorID, recipientID, liked)
}

type mockLogger struct {
	error func(format string, args ...any)
}

func (m *mockLogger) Error(format string, args ...any) {
	m.error(format, args...)
}

func TestServer(t *testing.T) {
	tests := []struct {
		name          string
		req           interface{}
		mockBehavior  func(*mockDecisionProvider, *mockDecisionCreator, *mockLogger)
		expectedResp  interface{}
		expectedError error
	}{
		{
			name: "ListLikedYou - success",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user1",
			},
			mockBehavior: func(mp *mockDecisionProvider, mc *mockDecisionCreator, ml *mockLogger) {
				mp.listLikedYou = func(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error) {
					return []domain.LikerInfo{{
						ActorID:   "user2",
						Timestamp: 1234567890,
					}}, "next_token", nil
				}
			},
			expectedResp: &pb.ListLikedYouResponse{
				Likers: []*pb.ListLikedYouResponse_Liker{{
					ActorId:       "user2",
					UnixTimestamp: 1234567890,
				}},
				NextPaginationToken: stringPtr("next_token"),
			},
			expectedError: nil,
		},
		{
			name: "ListLikedYou - empty recipient ID",
			req:  &pb.ListLikedYouRequest{},
			mockBehavior: func(mp *mockDecisionProvider, mc *mockDecisionCreator, ml *mockLogger) {
			},
			expectedResp:  nil,
			expectedError: status.Error(codes.InvalidArgument, "recipient user ID is required"),
		},
		{
			name: "CountLikedYou - success",
			req: &pb.CountLikedYouRequest{
				RecipientUserId: "user1",
			},
			mockBehavior: func(mp *mockDecisionProvider, mc *mockDecisionCreator, ml *mockLogger) {
				mp.countLikedYou = func(ctx context.Context, recipientID string) (uint64, error) {
					return 42, nil
				}
			},
			expectedResp: &pb.CountLikedYouResponse{
				Count: 42,
			},
			expectedError: nil,
		},
		{
			name: "PutDecision - success",
			req: &pb.PutDecisionRequest{
				ActorUserId:     "user1",
				RecipientUserId: "user2",
				LikedRecipient:  true,
			},
			mockBehavior: func(mp *mockDecisionProvider, mc *mockDecisionCreator, ml *mockLogger) {
				mc.saveDecision = func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
					return true, nil
				}
			},
			expectedResp: &pb.PutDecisionResponse{
				MutualLikes: true,
			},
			expectedError: nil,
		},
		{
			name: "PutDecision - same user",
			req: &pb.PutDecisionRequest{
				ActorUserId:     "user1",
				RecipientUserId: "user1",
				LikedRecipient:  true,
			},
			mockBehavior: func(mp *mockDecisionProvider, mc *mockDecisionCreator, ml *mockLogger) {
			},
			expectedResp:  nil,
			expectedError: status.Error(codes.InvalidArgument, "both actor and recipient user IDs are the same"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &mockDecisionProvider{}
			mockCreator := &mockDecisionCreator{}
			mockLogger := &mockLogger{}

			tt.mockBehavior(mockProvider, mockCreator, mockLogger)

			server := NewGRPCServer("8080", mockProvider, mockCreator, mockLogger)

			var resp interface{}
			var err error

			switch req := tt.req.(type) {
			case *pb.ListLikedYouRequest:
				resp, err = server.ListLikedYou(context.Background(), req)
			case *pb.CountLikedYouRequest:
				resp, err = server.CountLikedYou(context.Background(), req)
			case *pb.PutDecisionRequest:
				resp, err = server.PutDecision(context.Background(), req)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
