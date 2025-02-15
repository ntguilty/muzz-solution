package grpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"muzz-homework/internal/explore/domain"
	pb "muzz-homework/pkg/proto"
	"net"
)

type decisionProvider interface {
	ListLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error)
	ListNewLikedYou(ctx context.Context, recipientID string, encodedToken string) ([]domain.LikerInfo, string, error)
	CountLikedYou(ctx context.Context, recipientID string) (uint64, error)
}

type decisionCreator interface {
	SaveDecision(ctx context.Context, actorID string, recipientID string, liked bool) (bool, error)
}

type logger interface {
	Error(format string, args ...any)
}

type grpcServer struct {
	pb.UnimplementedExploreServiceServer
	engine   *grpc.Server
	port     string
	provider decisionProvider
	creator  decisionCreator
	logger   logger
}

func NewGRPCServer(port string, provider decisionProvider, creator decisionCreator, logger logger) *grpcServer {
	return &grpcServer{
		port:     port,
		engine:   grpc.NewServer(),
		provider: provider,
		creator:  creator,
		logger:   logger,
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

	if req.PaginationToken != nil && !isValidBase64(*req.PaginationToken) {
		return nil, status.Error(codes.InvalidArgument, "invalid pagination token format")
	}

	likers, nextToken, err := s.provider.ListLikedYou(ctx, req.RecipientUserId, req.GetPaginationToken())
	if err != nil {
		s.logger.Error("ListLikedYou failed", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	protoLikers := make([]*pb.ListLikedYouResponse_Liker, len(likers))
	for i, liker := range likers {
		protoLikers[i] = toLikerProto(liker)
	}

	var nextTokenPtr *string
	if nextToken != "" {
		nextTokenPtr = &nextToken
	}

	return &pb.ListLikedYouResponse{
		Likers:              protoLikers,
		NextPaginationToken: nextTokenPtr,
	}, nil
}

func (s *grpcServer) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient user ID is required")
	}

	if req.PaginationToken != nil && !isValidBase64(*req.PaginationToken) {
		return nil, status.Error(codes.InvalidArgument, "invalid pagination token format")
	}

	likers, nextToken, err := s.provider.ListNewLikedYou(ctx, req.RecipientUserId, req.GetPaginationToken())
	if err != nil {
		s.logger.Error("ListNewLikedYou failed", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	protoLikers := make([]*pb.ListLikedYouResponse_Liker, len(likers))
	for i, liker := range likers {
		protoLikers[i] = toLikerProto(liker)
	}

	var nextTokenPtr *string
	if nextToken != "" {
		nextTokenPtr = &nextToken
	}

	return &pb.ListLikedYouResponse{
		Likers:              protoLikers,
		NextPaginationToken: nextTokenPtr,
	}, nil
}

func (s *grpcServer) CountLikedYou(ctx context.Context, req *pb.CountLikedYouRequest) (*pb.CountLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient user ID is required")
	}

	count, err := s.provider.CountLikedYou(ctx, req.RecipientUserId)
	if err != nil {
		s.logger.Error("CountLikedYou failed", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.CountLikedYouResponse{
		Count: count,
	}, nil
}

func (s *grpcServer) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	if req.ActorUserId == "" || req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "both actor and recipient user IDs are required")
	}

	if req.ActorUserId == req.RecipientUserId {
		return nil, status.Error(codes.InvalidArgument, "both actor and recipient user IDs are the same")
	}

	mutualLikes, err := s.creator.SaveDecision(ctx, req.ActorUserId, req.RecipientUserId, req.LikedRecipient)
	if err != nil {
		s.logger.Error("PutDecision failed", err)
		return nil, status.Error(codes.Internal, "internal server error")
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

func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func toLikerProto(info domain.LikerInfo) *pb.ListLikedYouResponse_Liker {
	return &pb.ListLikedYouResponse_Liker{
		ActorId:       info.ActorID,
		UnixTimestamp: info.Timestamp,
	}
}
