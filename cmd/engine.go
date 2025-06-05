package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/spf13/cobra"
	"gitlab.com/zeelrupapara/trade-engine/config"
	database "gitlab.com/zeelrupapara/trade-engine/database/migrations"
	"gitlab.com/zeelrupapara/trade-engine/handlers"
	"gitlab.com/zeelrupapara/trade-engine/pkg/nats"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	service string = "enginecore"
	version string = "1.0.0"
)

// GetAPICommandDef runs app
func GetEngineCommandDef(cfg config.AppConfig, logger *zap.Logger) cobra.Command {
	apiCommand := cobra.Command{
		Use:   "engine",
		Short: "To start trade engine service server",
		Long:  `To start trade engine service server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start logging
			logger.Sugar().Infof("Logging started for service: ", service+"@"+version)

			// prepare TCP/IP listner for gPRC server
			lis, err := net.Listen("tcp", ":"+cfg.Port)
			if err != nil {
				logger.Sugar().Fatalf("failed to listen: %v", err)
			}

			// prepare gPRC server with recovery middleware as per  https://github.com/grpc-ecosystem/go-grpc-middleware
			s := grpc.NewServer(
				grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
					grpcrecovery.UnaryServerInterceptor(),
				),
				))

			logger.Sugar().Infof("%v", lis.Addr())

			// Create DB connection
			db, err := database.Connect(cfg.DB)
			if err != nil {
				logger.Error(err.Error(), zap.Any("Setup", "Init DB Connection"))
				return err
			}

			// Create Message Broker connection
			nats, err := nats.NewMsgBroker(cfg)
			if err != nil {
				logger.Error(err.Error(), zap.Any("Setup", "Init Nats Connection"))
				return err
			}

			// Init Engine Core
			ec := handlers.NewEngineCore(cfg, logger, db, nats.Nc)
			go ec.StartEngine()

			// Started Helthcheck
			healthcheck := health.NewServer()

			go func() {
				// asynchronously inspect dependencies and toggle serving status as needed
				next := healthpb.HealthCheckResponse_SERVING

				for {
					healthcheck.SetServingStatus(service, next)

					if next == healthpb.HealthCheckResponse_SERVING {
						next = healthpb.HealthCheckResponse_NOT_SERVING
					} else {
						next = healthpb.HealthCheckResponse_SERVING
					}

					time.Sleep(time.Second)
				}
			}()

			// start the gPRC server
			if err := s.Serve(lis); err != nil {
				logger.Sugar().Fatalf("failed to serve: %v", err)

			}

			// when painc receover
			if err := recover(); err != nil {
				logger.Sugar().Errorf("some panic ...:", err)
			}
			// we need nice way to exit will use os package notify
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

			select {
			case v := <-quit:
				fmt.Printf("signal.Notify CTRL+C: %v", v)
				s.GracefulStop()
				ec.StopCh <- os.Interrupt

			case done := <-ctx.Done():
				fmt.Printf("ctx.Done: %v", done)
			}

			// when all work fine then server loop end as we want this line will be called
			s.GracefulStop()
			ec.StopCh <- os.Interrupt
			logger.Info("Server stopped")

			return nil
		},
	}

	return apiCommand
}
