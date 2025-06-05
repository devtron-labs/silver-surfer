/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logger

import (
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level int `env:"LOG_LEVEL" envDefault:"-1"` // default debug
}

func NewSugaredLogger() *zap.SugaredLogger {
	logConfig := &LogConfig{}
	err := env.Parse(logConfig)
	if err != nil {
		panic("failed to parse env config: " + err.Error())
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(logConfig.Level))

	log, err := config.Build()
	if err != nil {
		panic("failed to create the logger: " + err.Error())
	}
	return log.Sugar()
}
