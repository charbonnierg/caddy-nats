package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseBool", func() {

	Context("Parse bool", func() {
		var Fail = HaveOccurred
		testcase := func(input string, expected bool, opts ...parser.Option) Assertion {
			var err error
			dispenser := NewTestDispenser(input)
			var dest bool
			err = parser.ExpectString(dispenser, parser.Match("test"))
			Expect(err).NotTo(HaveOccurred())
			err = parser.ParseBool(dispenser, &dest, opts...)
			if err != nil {
				Expect(dest).To(Equal(expected))
				return Expect(err)
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(dest).To(Equal(expected))
				return Expect(nil)
			}
		}

		It("Should return an error if value is invalid", func() {
			testcase("test invalid", false).Should(Fail())
		})
		Context("Without options", func() {
			It("should parse true", func() {
				testcase("test true", true).Should(Succeed())
				testcase("test 1", true).Should(Succeed())
				testcase("test on", true).Should(Succeed())
				testcase("test yes", true).Should(Succeed())
			})
			It("should parse false", func() {
				testcase("test false", false).Should(Succeed())
				testcase("test 0", false).Should(Succeed())
				testcase("test off", false).Should(Succeed())
				testcase("test no", false).Should(Succeed())
			})
			It("should parse empty value as true", func() {
				testcase("test", true).Should(Succeed())
			})
		})
		Context("With reverse option", func() {
			It("should parse false", func() {
				testcase("test true", false, parser.Reverse()).Should(Succeed())
				testcase("test 1", false, parser.Reverse()).Should(Succeed())
				testcase("test on", false, parser.Reverse()).Should(Succeed())
				testcase("test yes", false, parser.Reverse()).Should(Succeed())
				testcase("test", false, parser.Reverse()).Should(Succeed())
			})
			It("should parse true", func() {
				testcase("test false", true, parser.Reverse()).Should(Succeed())
				testcase("test 0", true, parser.Reverse()).Should(Succeed())
				testcase("test off", true, parser.Reverse()).Should(Succeed())
				testcase("test no", true, parser.Reverse()).Should(Succeed())
			})
			It("should parse empty value as false", func() {
				testcase("test", false, parser.Reverse()).Should(Succeed())
			})
		})
		Context("With default option", func() {
			It("should parse the default value", func() {
				testcase("test", true, parser.Default(true)).Should(Succeed())
				testcase("test", false, parser.Default(false)).Should(Succeed())
			})
			It("should not use default value when value is present", func() {
				testcase("test true", true, parser.Default(false)).Should(Succeed())
				testcase("test false", false, parser.Default(true)).Should(Succeed())
			})
			Context("With the reverse option", func() {
				It("Should not reverse the default value", func() {
					testcase("test", true, parser.Default(true), parser.Reverse()).Should(Succeed())
					testcase("test", false, parser.Default(false), parser.Reverse()).Should(Succeed())
				})
			})
		})
		Context("With error if empty option", func() {
			It("Should return an error if empty", func() {
				testcase("test", false, parser.ErrorIfEmpty()).Should(Fail())
			})
		})
	})

})
