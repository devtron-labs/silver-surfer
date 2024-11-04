package service

import (
	"github.com/devtron-labs/silver-surfer/app/constants"
	"github.com/devtron-labs/silver-surfer/kubedd"
	"github.com/devtron-labs/silver-surfer/pkg"
	"go.uber.org/zap"
)

type ClusterUpgradeReadService interface {
	GetClusterUpgradeSummaryValidationResult(targetK8sVersion string) ([]pkg.SummaryValidationResult, error)
}

type ClusterUpgradeReadServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewClusterUpgradeReadServiceImpl(logger *zap.SugaredLogger) *ClusterUpgradeReadServiceImpl {
	return &ClusterUpgradeReadServiceImpl{
		logger: logger,
	}
}

func (impl *ClusterUpgradeReadServiceImpl) GetClusterUpgradeSummaryValidationResult(targetK8sVersion string) ([]pkg.SummaryValidationResult, error) {
	cluster := pkg.NewClusterViaInClusterConfig()
	results, err := kubedd.ValidateCluster(cluster, &pkg.Config{TargetKubernetesVersion: targetK8sVersion})
	if err != nil {
		impl.logger.Errorw("error in ValidateCluster", "err", err)
		return nil, err
	}
	outputManager := pkg.GetOutputManager(constants.OutputJson, false)
	outputManager.PutBulk(results)
	return outputManager.GetSummaryValidationResultBulk(), nil
}
