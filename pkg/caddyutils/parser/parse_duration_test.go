package parser_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseDuration", func() {
	Context("ParseDuration", func() {
		var Fail = HaveOccurred
		testcase := func(input string, expected time.Duration, opts ...parser.Option) Assertion {
			var err error
			dispenser := NewTestDispenser(input)
			var dest time.Duration
			err = parser.ExpectString(dispenser)
			Expect(err).NotTo(HaveOccurred())
			err = parser.ParseDuration(dispenser, &dest, opts...)
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
			It("should parse duration", func() {
				testcase("test 1ns", 1).Should(Succeed())
				testcase("test 1µs", 1000).Should(Succeed())
				testcase("test 1ms", 1000000).Should(Succeed())
				testcase("test 1s", 1000000000).Should(Succeed())
				testcase("test 1m", 60000000000).Should(Succeed())
			})
			It("should return an error when empty", func() {
				testcase("test", 0).Should(Fail())
			})
			It("should return an error when unit is missing", func() {
				testcase("test 1", 0).Should(Fail())
			})
			It("should consider empty string as an invalid value", func() {
				testcase(`test ""`, 0).Should(Fail())
			})
			It("should return an error when invalid duration", func() {
				testcase("test invalid", 0).Should(Fail())
			})
		})
		Context("With Default() option", func() {
			It("should use default option when empty", func() {
				testcase("test", 1, parser.Default(time.Duration(1))).Should(Succeed())
			})
			It("should ignore default option when value is valid", func() {
				testcase("test 1µs", 1000, parser.Default(time.Duration(1))).Should(Succeed())
				testcase("test 1ms", 1000000, parser.Default(time.Duration(1))).Should(Succeed())
				testcase("test 1s", 1000000000, parser.Default(time.Duration(1))).Should(Succeed())
				testcase("test 1m", 60000000000, parser.Default(time.Duration(1))).Should(Succeed())
			})
			It("should ignore default option when value is invalid empty string", func() {
				testcase(`test ""`, 0, parser.Default(time.Duration(1))).Should(Fail())
			})
			It("should ignore default option when value is invalid", func() {
				testcase("test invalid", 0, parser.Default(time.Duration(1))).Should(Fail())
			})
		})
		Context("With AllowEmpty() option", func() {
			It("should allow missing value", func() {
				testcase("test", 0, parser.AllowEmpty()).Should(Succeed())
			})
		})
		Context("With Inplace() option", func() {
			It("should parse duration", func() {
				testcase("1h", 3600000000000, parser.Inplace()).Should(Succeed())
			})
			It("should return an error when empty", func() {
				testcase(`""`, 0, parser.Inplace()).Should(Fail())
			})
			It("should return an error when invalid duration", func() {
				testcase("invalid", 0, parser.Inplace()).Should(Fail())
			})
			Context("With AllowEmpty() option", func() {
				It("should allow empty value", func() {
					testcase(`""`, 0, parser.Inplace(), parser.AllowEmpty()).Should(Succeed())
				})
			})
			Context("With Default() option", func() {
				It("should use default option when empty", func() {
					testcase(`""`, 1, parser.Inplace(), parser.Default(time.Duration(1))).Should(Succeed())
				})
				It("should ignore default option when value is valid", func() {
					testcase("2s", 2000000000, parser.Inplace(), parser.Default(time.Duration(1))).Should(Succeed())
				})
			})
		})
	})
})
