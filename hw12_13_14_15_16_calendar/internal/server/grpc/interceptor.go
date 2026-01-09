package grpcserver

import (
	"context"
	"fmt"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/contracts"
	"google.golang.org/grpc"
)

func UnaryLogRequestInterceptor(logger contracts.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		duration := time.Since(start)

		timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
		method := info.FullMethod
		latency := duration.Milliseconds()

		logLine := fmt.Sprintf(
			"[%s] %s %d",
			timestamp,
			method,
			latency,
		)

		logger.Info(logLine)

		return resp, err
	}
}
