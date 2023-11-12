// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package main

import "go.uber.org/zap"

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger.Info("Hello, world!")
}
