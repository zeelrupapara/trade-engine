package handlers

import (
	"github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type TradeEngineCore struct {
	logger *zap.Logger
	db     *goqu.Database
}

func NewGrpcTradeEngineService(grpc *grpc.Server, db *goqu.Database, logger *zap.Logger) error {
	// init news services

	// tradeEngineCore := &TradeEngineCore{
	// 	logger: logger,
	// 	db:     db,
	// }

	// register the OrderServiceServer
	// engine.RegisterNewsServiceServer(grpc, tradeEngineCore)
	return nil
}
