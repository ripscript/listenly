package grpcserver

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic recovered in grpc handler",
				"method", info.FullMethod,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
