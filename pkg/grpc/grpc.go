package grpc_server

import (
	"net"

	"github.com/doug-martin/goqu/v9"
	"gitlab.com/zeelrupapara/news-srv/handlers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	addr   string
	logger *zap.Logger
	db     *goqu.Database
}

// NewGRPCServer creates new grpc server
func NewGRPCServer(addr string, logger *zap.Logger, db *goqu.Database) *gRPCServer {
	return &gRPCServer{addr: addr, logger: logger, db: db}
}

// Run starts grpc server
func (s *gRPCServer) Run() error {
	// Start net server for grpc
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	// register our grpc services
	err = handlers.NewGrpcNewsService(grpcServer, s.db, s.logger)
	if err != nil {
		return err
	}

	s.logger.Info("Starting gRPC server", zap.String("addr", s.addr))
	return grpcServer.Serve(lis)
}
