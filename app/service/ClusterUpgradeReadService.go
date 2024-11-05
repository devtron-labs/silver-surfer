package service

import (
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/silver-surfer/app/adaptors"
	"github.com/devtron-labs/silver-surfer/app/constants"
	"github.com/devtron-labs/silver-surfer/app/grpc"
	"github.com/devtron-labs/silver-surfer/kubedd"
	"github.com/devtron-labs/silver-surfer/pkg"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

type ClusterUpgradeReadService interface {
	GetClusterUpgradeSummaryValidationResult(targetK8sVersion string, clusterConfig *grpc.ClusterConfig) ([]pkg.SummaryValidationResult, error)
}

type ClusterUpgradeReadServiceImpl struct {
	logger  *zap.SugaredLogger
	k8sUtil k8s2.K8sService
}

func NewClusterUpgradeReadServiceImpl(logger *zap.SugaredLogger, k8sUtil k8s2.K8sService) *ClusterUpgradeReadServiceImpl {
	return &ClusterUpgradeReadServiceImpl{
		logger:  logger,
		k8sUtil: k8sUtil,
	}
}

func (impl *ClusterUpgradeReadServiceImpl) GetClusterUpgradeSummaryValidationResult(targetK8sVersion string, clusterConfig *grpc.ClusterConfig) ([]pkg.SummaryValidationResult, error) {
	var restConfig *rest.Config
	var err error
	localClusterConfig := adaptors.ConvertGrpcObjToClusterConfig(clusterConfig)
	if len(localClusterConfig.ClusterName) > 0 {
		impl.logger.Infow("fetching restConfig via GetRestConfigByCluster", "clusterName", localClusterConfig.ClusterName)
		restConfig, err = impl.k8sUtil.GetRestConfigByCluster(localClusterConfig)
		if err != nil {
			impl.logger.Errorw("error in getting rest config by cluster config", "clusterName", localClusterConfig.ClusterName, "err", err)
			return nil, err
		}
	}
	cluster := pkg.NewClusterFromEnvOrConfig(restConfig)
	results, err := kubedd.ValidateCluster(cluster, &pkg.Config{TargetKubernetesVersion: targetK8sVersion})
	if err != nil {
		impl.logger.Errorw("error in ValidateCluster", "err", err)
		return nil, err
	}
	outputManager := pkg.GetOutputManager(constants.OutputJson, false)
	outputManager.PutBulk(results)
	return outputManager.GetSummaryValidationResultBulk(), nil
}
