// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package testutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/quara-dev/beyond/pkg/testutils"
)

var _ = Describe("Logger", func() {
	Context("WrapLogger", func() {
		It("should wrap logger", func() {
			// Create a logger (or reuse your app's logger)
			logger, err := zap.NewDevelopment()
			Expect(err).NotTo(HaveOccurred())
			// Wrap the logger
			logger, logs := testutils.WrapLogger(logger)
			// Use the logger
			logger.Debug("test1")
			logger.Info("test2")
			// Sync the logger to make sure all logs are flushed
			err = logger.Sync()
			Expect(err).NotTo(HaveOccurred())
			// Check the logs
			allMsgs := logs.All()
			Expect(len(allMsgs)).To(Equal(2))
			msg1 := allMsgs[0]
			Expect(msg1.Level).To(Equal(zapcore.DebugLevel))
			Expect(msg1.Message).To(Equal("test1"))
			msg2 := allMsgs[1]
			Expect(msg2.Message).To(Equal("test2"))
			Expect(msg2.Level).To(Equal(zapcore.InfoLevel))
		})
	})
})
