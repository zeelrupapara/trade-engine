package handlers

import (
	"github.com/doug-martin/goqu/v9"
	"gitlab.com/zeelrupapara/news-srv/protobuf/genproto/news"
	"gitlab.com/zeelrupapara/news-srv/services"
	"gitlab.com/zeelrupapara/news-srv/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	newsSrv types.NewsService
	logger  *zap.Logger
	db      *goqu.Database
	news.UnimplementedNewsServiceServer
}

func NewGrpcNewsService(grpc *grpc.Server, db *goqu.Database, logger *zap.Logger) error {
	// init news services
	newsService := services.NewNewsService(db, logger)

	gRPCHandler := &GrpcHandler{
		newsSrv: newsService,
		logger:  logger,
		db:      db,
	}

	// register the OrderServiceServer
	news.RegisterNewsServiceServer(grpc, gRPCHandler)
	return nil
}
