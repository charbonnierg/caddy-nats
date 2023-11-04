package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParsePorts", func() {
	Context("ParsePort", func() {
		var Fail = HaveOccurred
		testcase := func(input string, expected int, opts ...parser.Option) Assertion {
			var err error
			dispenser := NewTestDispenser(input)
			var dest int
			err = parser.ExpectString(dispenser)
			Expect(err).NotTo(HaveOccurred())
			err = parser.ParseNetworkPort(dispenser, &dest, opts...)
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
			It("should parse integer", func() {
				testcase("test 1", 1).Should(Succeed())
				testcase("test 65535", 65535).Should(Succeed())
			})
			It("should return an error when empty", func() {
				testcase("test", 0).Should(Fail())
			})
			It("should consider empty string as an invalid value", func() {
				testcase(`test ""`, 0).Should(Fail())
			})
			It("should return an error when invalid integer", func() {
				testcase("test invalid", 0).Should(Fail())
				testcase("test 65536", 0).Should(Fail())
			})
		})
		Context("With Default() option", func() {
			It("should use default option when empty", func() {
				testcase("test", 1, parser.Default(1)).Should(Succeed())
			})
			It("should ignore default option when value is valid", func() {
				testcase("test 2", 2, parser.Default(1)).Should(Succeed())
			})
			It("should ignore default option when value is invalid empty string", func() {
				testcase(`test ""`, 0, parser.Default(1)).Should(Fail())
			})
			It("should ignore default option when value is invalid", func() {
				testcase("test invalid", 0, parser.Default(1)).Should(Fail())
				testcase("test 1.0", 0, parser.Default(1)).Should(Fail())
				testcase("test 65536", 0).Should(Fail())
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
			})
			It("should return an error when empty", func() {
				testcase(`""`, 0, parser.Inplace()).Should(Fail())
			})
			It("should return an error when invalid integer", func() {
				testcase("invalid", 0, parser.Inplace()).Should(Fail())
				testcase("1.0", 0, parser.Inplace()).Should(Fail())
				testcase("65536", 0, parser.Inplace()).Should(Fail())
			})
			Context("With AllowEmpty() option", func() {
				It("should allow empty value", func() {
					testcase(`""`, 0, parser.Inplace(), parser.AllowEmpty()).Should(Succeed())
				})
			})
			Context("With Default() option", func() {
				It("should use default option when empty", func() {
					testcase(`""`, 1, parser.Inplace(), parser.Default(1)).Should(Succeed())
				})
				It("should ignore default option when value is valid", func() {
					testcase("2", 2, parser.Inplace(), parser.Default(1)).Should(Succeed())
				})
			})
		})
	})
})
