// Package port ...
package port

import (
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

// Server ...
type Server struct {
	grpc   *grpc.Server
	logger *slog.Logger
}

// NewServer ...
func NewServer(logger *slog.Logger) *Server {
	grpcServer := grpc.NewServer()
	return &Server{
		grpc:   grpcServer,
		logger: logger,
	}
}

// GRPC ...
func (s *Server) GRPC() *grpc.Server {
	return s.grpc
}

// Run ...
func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	s.logger.Info("gRPC server started", slog.String("addr", addr))

	return s.grpc.Serve(lis)
}