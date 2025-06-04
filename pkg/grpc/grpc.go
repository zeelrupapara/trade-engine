package grpc_server

import (
	"net"

	"github.com/doug-martin/goqu/v9"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	Addr   string
	Logger *zap.Logger
	DB     *goqu.Database
	Nats   *nats.Conn
}

// NewGRPCServer creates new grpc server
func NewGRPCServer(addr string, logger *zap.Logger, db *goqu.Database, nats *nats.Conn) *GRPCServer {
	return &GRPCServer{Addr: addr, Logger: logger, DB: db, Nats: nats}
}

// Run starts grpc server
func (s *GRPCServer) StartRPC() error {
	// Start net server for grpc
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		s.Logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	s.Logger.Info("Starting gRPC server", zap.String("addr", s.Addr))
	return grpcServer.Serve(lis)
}
