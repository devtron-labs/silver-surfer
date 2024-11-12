package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/silver-surfer/app/api"
	"github.com/devtron-labs/silver-surfer/app/constants"
	grpc2 "github.com/devtron-labs/silver-surfer/app/grpc"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"runtime/debug"
	"time"
)

type App struct {
	grpcServer    *grpc.Server
	StartupConfig *StartupConfig
	logger        *zap.SugaredLogger
	GrpcHandler   *api.GrpcHandlerImpl
}

func NewApp(
	logger *zap.SugaredLogger,
	GrpcHandler *api.GrpcHandlerImpl,
) *App {
	return &App{
		logger:      logger,
		GrpcHandler: GrpcHandler,
	}
}

type StartupConfig struct {
	GrpcPort           int `env:"SERVER_GRPC_PORT" envDefault:"8111"`
	GrpcMaxRecvMsgSize int `env:"GRPC_MAX_RECEIVE_MSG_SIZE" envDefault:"20"` // In mb
	GrpcMaxSendMsgSize int `env:"GRPC_MAX_SEND_MSG_SIZE" envDefault:"4"`     // In mb
}

func (app *App) Start() {
	// Parse config
	app.StartupConfig = &StartupConfig{}
	err := env.Parse(app.StartupConfig)
	if err != nil {
		app.logger.Errorw("failed to parse configuration")
		os.Exit(2)
	}

	// Start gRPC server
	err = app.initGrpcServer(app.StartupConfig.GrpcPort)
	if err != nil {
		app.logger.Errorw("error starting grpc server", "err", err)
		os.Exit(2)
	}
}

func (app *App) initGrpcServer(port int) error {
	app.logger.Infow("gRPC server starting", "port", app.StartupConfig.GrpcPort)

	//listen on the port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to start grpcServer %v", err)
		return err
	}

	grpcPanicRecoveryHandler := func(p any) (err error) {
		app.logger.Error(constants.PanicLogIdentifier, "recovered from panic", "panic", p, "stack", string(debug.Stack()))
		return status.Errorf(codes.Internal, "%s", p)
	}
	recoveryOption := recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(app.StartupConfig.GrpcMaxRecvMsgSize * 1024 * 1024), // GRPC Request size
		grpc.MaxSendMsgSize(app.StartupConfig.GrpcMaxSendMsgSize * 1024 * 1024), // GRPC Response size
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: 10 * time.Second,
		}),
		grpc.ChainStreamInterceptor(
			grpcPrometheus.StreamServerInterceptor,
			recovery.StreamServerInterceptor(recoveryOption)), // panic interceptor, should be at last
		grpc.ChainUnaryInterceptor(
			grpcPrometheus.UnaryServerInterceptor,
			recovery.UnaryServerInterceptor(recoveryOption)), // panic interceptor, should be at last
	}
	// create a new gRPC grpcServer
	app.grpcServer = grpc.NewServer(opts...)

	// register Silver Surfer service
	grpc2.RegisterSilverSurferServiceServer(app.grpcServer, app.GrpcHandler)
	grpcPrometheus.EnableHandlingTimeHistogram()
	grpcPrometheus.Register(app.grpcServer)
	// start listening on address
	if err = app.grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start: %v", err)
		return err
	}
	return nil
}

func (app *App) Stop() {

	app.logger.Infow("silver-surfer shutdown initiating")

	// Gracefully stop the gRPC server
	app.logger.Info("Stopping gRPC server...")
	app.grpcServer.GracefulStop()

	app.logger.Infow("housekeeping done. exiting now")
}
