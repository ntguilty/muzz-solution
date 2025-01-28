package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	pb "muzz-homework/pkg/proto"
	"net"
)

type decisionProvider interface {
	ListLikedYou(ctx context.Context, recipientID string, paginationToken *string) ([]*pb.ListLikedYouResponse_Liker, string, error)
	ListNewLikedYou(ctx context.Context, recipientID string, paginationToken *string) ([]*pb.ListLikedYouResponse_Liker, string, error)
	CountLikedYou(ctx context.Context, recipientID string) (uint64, error)
	PutDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error)
}

type grpcServer struct {
	pb.UnimplementedExploreServiceServer
	engine  *grpc.Server
	port    string
	service decisionProvider
}

func NewGRPCServer(port string, service decisionProvider) *grpcServer {
	return &grpcServer{
		port:    port,
		engine:  grpc.NewServer(),
		service: service,
	}
}

func (s *grpcServer) Register() {
	pb.RegisterExploreServiceServer(s.engine, s)
	grpc_health_v1.RegisterHealthServer(s.engine, health.NewServer())
}

func (s *grpcServer) Run() error {
	s.Register()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	return s.engine.Serve(lis)
}

func (s *grpcServer) ListLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient user ID is required")
	}

	users, nextToken, err := s.service.ListLikedYou(ctx, req.RecipientUserId, req.PaginationToken)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list users: %v", err))
	}

	return &pb.ListLikedYouResponse{
		Likers:              users,
		NextPaginationToken: &nextToken,
	}, nil
}

func (s *grpcServer) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient user ID is required")
	}

	likers, nextToken, err := s.service.ListNewLikedYou(ctx, req.RecipientUserId, req.PaginationToken)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list new likes: %v", err))
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextToken,
	}, nil
}

func (s *grpcServer) CountLikedYou(ctx context.Context, req *pb.CountLikedYouRequest) (*pb.CountLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient user ID is required")
	}

	count, err := s.service.CountLikedYou(ctx, req.RecipientUserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to count likes: %v", err))
	}

	return &pb.CountLikedYouResponse{
		Count: count,
	}, nil
}

func (s *grpcServer) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	if req.ActorUserId == "" || req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "both actor and recipient user IDs are required")
	}

	mutualLikes, err := s.service.PutDecision(ctx, req.ActorUserId, req.RecipientUserId, req.LikedRecipient)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to put decision: %v", err))
	}

	return &pb.PutDecisionResponse{
		MutualLikes: mutualLikes,
	}, nil
}

func (s *grpcServer) GracefulStop() {
	s.engine.GracefulStop()
}

func (s *grpcServer) Stop() {
	s.engine.Stop()
}
