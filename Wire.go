//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/silver-surfer/app"
	"github.com/devtron-labs/silver-surfer/app/api"
	"github.com/devtron-labs/silver-surfer/app/logger"
	"github.com/google/wire"
)

func InitializeApp() (*app.App, error) {
	wire.Build(
		app.NewApp,
		logger.NewSugaredLogger,
		api.NewGrpcHandlerImpl,
	)
	return &app.App{}, nil
}
