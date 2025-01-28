package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "muzz-homework/pkg/proto"
	"testing"
)

type mockLikeService struct {
	listLikedYou    func(ctx context.Context, recipientID string, paginationToken *string) ([]*pb.ListLikedYouResponse_Liker, string, error)
	listNewLikedYou func(ctx context.Context, recipientID string, paginationToken *string) ([]*pb.ListLikedYouResponse_Liker, string, error)
	countLikedYou   func(ctx context.Context, recipientID string) (uint64, error)
	putDecision     func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error)
}

func (m *mockLikeService) ListLikedYou(ctx context.Context, recipientID string, paginationToken *string) ([]*pb.ListLikedYouResponse_Liker, string, error) {
	return m.listLikedYou(ctx, recipientID, paginationToken)
}

func (m *mockLikeService) ListNewLikedYou(ctx context.Context, recipientID string, paginationToken *string) ([]*pb.ListLikedYouResponse_Liker, string, error) {
	return m.listNewLikedYou(ctx, recipientID, paginationToken)
}

func (m *mockLikeService) CountLikedYou(ctx context.Context, recipientID string) (uint64, error) {
	return m.countLikedYou(ctx, recipientID)
}

func (m *mockLikeService) PutDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
	return m.putDecision(ctx, actorID, recipientID, liked)
}

func TestServer(t *testing.T) {
	tests := []struct {
		name          string
		req           interface{}
		mockBehavior  func(*mockLikeService)
		expectedResp  interface{}
		expectedError error
	}{
		{
			name: "ListLikedYou - success",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user1",
			},
			mockBehavior: func(m *mockLikeService) {
				m.listLikedYou = func(ctx context.Context, recipientID string, paginationToken *string) ([]*pb.ListLikedYouResponse_Liker, string, error) {
					return []*pb.ListLikedYouResponse_Liker{{
						ActorId:       "user2",
						UnixTimestamp: 1234567890,
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
			name:          "ListLikedYou - empty recipient ID",
			req:           &pb.ListLikedYouRequest{},
			mockBehavior:  func(m *mockLikeService) {},
			expectedResp:  nil,
			expectedError: status.Error(codes.InvalidArgument, "recipient user ID is required"),
		},
		{
			name: "CountLikedYou - success",
			req: &pb.CountLikedYouRequest{
				RecipientUserId: "user1",
			},
			mockBehavior: func(m *mockLikeService) {
				m.countLikedYou = func(ctx context.Context, recipientID string) (uint64, error) {
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
			mockBehavior: func(m *mockLikeService) {
				m.putDecision = func(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error) {
					return true, nil
				}
			},
			expectedResp: &pb.PutDecisionResponse{
				MutualLikes: true,
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockLikeService{}
			tt.mockBehavior(mockService)

			server := NewGRPCServer("8080", mockService)

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
