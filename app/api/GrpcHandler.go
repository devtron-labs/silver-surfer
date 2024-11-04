package api

import (
	"context"
	"github.com/devtron-labs/silver-surfer/app/grpc"
	"go.uber.org/zap"
)

type GrpcHandlerImpl struct {
	grpc.UnimplementedApplicationServiceServer
	logger *zap.SugaredLogger
}

func NewGrpcHandlerImpl(logger *zap.SugaredLogger) *GrpcHandlerImpl {
	return &GrpcHandlerImpl{
		logger: logger,
	}
}

func (impl *GrpcHandlerImpl) GetClusterUpgradeSummaryValidationResult(ctx context.Context, request *grpc.ClusterUpgradeRequest) (*grpc.ClusterUpgradeResponse, error) {

}
