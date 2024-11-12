package api

import (
	"context"
	"github.com/devtron-labs/silver-surfer/app/adaptors"
	"github.com/devtron-labs/silver-surfer/app/grpc"
	"github.com/devtron-labs/silver-surfer/app/service"
	"go.uber.org/zap"
)

type GrpcHandlerImpl struct {
	grpc.UnimplementedSilverSurferServiceServer
	logger                    *zap.SugaredLogger
	clusterUpgradeReadService service.ClusterUpgradeReadService
}

func NewGrpcHandlerImpl(
	logger *zap.SugaredLogger,
	clusterUpgradeReadService service.ClusterUpgradeReadService,
) *GrpcHandlerImpl {
	return &GrpcHandlerImpl{
		logger:                    logger,
		clusterUpgradeReadService: clusterUpgradeReadService,
	}
}

func (impl *GrpcHandlerImpl) GetClusterUpgradeSummaryValidationResult(ctx context.Context, request *grpc.ClusterUpgradeRequest) (*grpc.ClusterUpgradeResponse, error) {
	impl.logger.Infow("scan cluster resources compatibility for k8s version upgrade request", "request", request)
	summaryValidationResult, err := impl.clusterUpgradeReadService.GetClusterUpgradeSummaryValidationResult(request.TargetK8SVersion, request.ClusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting cluster upgrade summary validation result", "targetK8sVersion", request.TargetK8SVersion, "err", err)
		return nil, err
	}
	svr := adaptors.ConvertSummaryValidationResultToGrpcObj(summaryValidationResult)
	return &grpc.ClusterUpgradeResponse{Results: svr}, nil
}
