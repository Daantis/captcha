package server

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// interceptors
func streamCloseInterceptor(c incomingStreamCloser) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if c == nil {
			return status.Errorf(codes.Internal, "sever configured inproperly")
		}

		if c.closed() {
			return status.Error(codes.Unavailable, "closing stream")
		}

		return handler(srv, ss)
	}
}

func unaryCloseInterceptor(c incomingStreamCloser) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if c == nil {
			return nil, status.Errorf(codes.Internal, "sever configured inproperly")
		}

		if c.closed() {
			return nil, status.Errorf(codes.Unavailable, "closing unary")
		}

		return handler(ctx, req)
	}
}
