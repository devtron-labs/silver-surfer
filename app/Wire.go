//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/silver-surfer/app/api"
	"github.com/devtron-labs/silver-surfer/app/logger"
	"github.com/devtron-labs/silver-surfer/app/service"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		NewApp,
		logger.NewSugaredLogger,
		api.NewGrpcHandlerImpl,
		service.NewClusterUpgradeReadServiceImpl,
		wire.Bind(new(service.ClusterUpgradeReadService), new(*service.ClusterUpgradeReadServiceImpl)),
		k8s.GetRuntimeConfig,
		k8s.NewK8sUtil,
		wire.Bind(new(k8s.K8sService), new(*k8s.K8sServiceImpl)),
	)
	return &App{}, nil
}
