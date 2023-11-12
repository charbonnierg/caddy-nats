// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseInt64ByteSize", func() {
	Context("ParseInt64ByteSize", func() {
		var Fail = HaveOccurred
		testcase := func(input string, expected int64, opts ...parser.Option) Assertion {
			var err error
			dispenser := NewTestDispenser(input)
			var dest int64
			err = parser.ExpectString(dispenser)
			Expect(err).NotTo(HaveOccurred())
			err = parser.ParseInt64ByteSize(dispenser, &dest, opts...)
			if err != nil {
				Expect(dest).To(Equal(expected))
				return Expect(err)
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(dest).To(Equal(expected))
				return Expect(nil)
			}
		}
		Context("Without options", func() {
			It("should parse byte size", func() {
				testcase("test 2KiB", 2048).Should(Succeed())
				testcase("test 2KB", 2000).Should(Succeed())
				testcase("test 2K", 2000).Should(Succeed())
				testcase("test 2", 2).Should(Succeed())
			})
			It("should return an error when empty", func() {
				testcase("test", 0).Should(Fail())
			})
			It("should consider empty string as an invalid value", func() {
				testcase(`test ""`, 0).Should(Fail())
			})
			It("should return an error when invalid byte size", func() {
				testcase("test invalid", 0).Should(Fail())
				testcase("test 200HH", 0).Should(Fail())
			})
		})
		Context("With Default() option", func() {
			It("should use default option when empty", func() {
				testcase("test", 1, parser.Default(int64(1))).Should(Succeed())
			})
			It("should ignore default option when value is valid", func() {
				testcase("test 2KiB", 2048, parser.Default(int64(1))).Should(Succeed())
				testcase("test 2KB", 2000, parser.Default(int64(1))).Should(Succeed())
				testcase("test 2K", 2000, parser.Default(int64(1))).Should(Succeed())
				testcase("test 2", 2, parser.Default(int64(1))).Should(Succeed())
			})
			It("should ignore default option when value is invalid empty string", func() {
				testcase(`test ""`, 0, parser.Default(int64(1))).Should(Fail())
			})
			It("should ignore default option when value is invalid", func() {
				testcase("test invalid", 0, parser.Default(int64(1))).Should(Fail())
				testcase("test 3.4.0", 0, parser.Default(int64(1))).Should(Fail())
			})
		})
		Context("With AllowEmpty() option", func() {
			It("should allow missing value", func() {
				testcase("test", 0, parser.AllowEmpty()).Should(Succeed())
			})
		})
		Context("With Inplace() option", func() {
			It("should parse integer", func() {
				testcase("1", 1, parser.Inplace()).Should(Succeed())
				testcase("1KiB", 1024, parser.Inplace()).Should(Succeed())
				testcase("1KB", 1000, parser.Inplace()).Should(Succeed())
				testcase("1K", 1000, parser.Inplace()).Should(Succeed())
			})
			It("should return an error when empty", func() {
				testcase(`""`, 0, parser.Inplace()).Should(Fail())
			})
			It("should return an error when invalid integer", func() {
				testcase("invalid", 0, parser.Inplace()).Should(Fail())
				testcase("1.0.0", 0, parser.Inplace()).Should(Fail())
			})
			Context("With Allow() option", func() {
				It("should allow empty value", func() {
					testcase(`""`, 0, parser.Inplace(), parser.AllowEmpty()).Should(Succeed())
				})
			})
			Context("With Default() option", func() {
				It("should use default option when empty", func() {
					testcase(`""`, 1, parser.Inplace(), parser.Default(int64(1))).Should(Succeed())
				})
				It("should ignore default option when value is valid", func() {
					testcase("2MiB", 2097152, parser.Inplace(), parser.Default(int64(1))).Should(Succeed())
				})
			})
		})
	})
})
