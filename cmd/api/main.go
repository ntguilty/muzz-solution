package main

import (
	"context"
	"errors"
	"github.com/labstack/gommon/log"
	goredis "github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"muzz-homework/internal/explore/adapters/grpc"
	"muzz-homework/internal/explore/application"
	infraPostgre "muzz-homework/internal/explore/infrastructure/postgres"
	infraRedis "muzz-homework/internal/explore/infrastructure/redis"
	"muzz-homework/pkg/postgres"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const port = "8000"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	sqlDB, err := postgres.NewSQLDB()
	if err != nil {
		log.Fatalf("failed to create db err: %v", err)
		return
	}

	ttlStr := getEnvOrDefault("REDIS_TTL_SECONDS", "900")
	ttlSeconds, err := strconv.Atoi(ttlStr)
	if err != nil {
		log.Warnf("invalid REDIS_TTL_SECONDS value, using default: %v", err)
		ttlSeconds = 900
	}
	ttl := time.Duration(ttlSeconds) * time.Second

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     getEnvOrDefault("REDIS_ADDR", "redis:6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
		return
	}

	decisionRepo := infraPostgre.NewDecisionRepository(sqlDB)
	redisCache := infraRedis.NewRedisCache(redisClient, infraRedis.RedisConfig{
		Prefix: getEnvOrDefault("REDIS_PREFIX", "muzz"),
		TTL:    ttl,
	})

	decisionProvider := application.NewDecisionProvider(decisionRepo, redisCache)
	decisionCreator := application.NewDecisionCreator(decisionRepo)

	grpcServer := grpc.NewGRPCServer(port, decisionProvider, decisionCreator, logger)

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		log.Infof("starting grpcServer on: %v", port)
		return grpcServer.Run()
	})

	group.Go(func() error {
		<-ctx.Done()
		log.Infof("shutting down gRPC server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-shutdownCtx.Done():
			log.Infof("timeout during graceful shutdown, forcing exit")
			grpcServer.Stop()
		case <-done:
			log.Infof("server stopped gracefully")
		}

		return nil
	})

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Errorf("failed to wait for group: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
