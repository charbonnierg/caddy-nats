// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseString", func() {
	var Fail = HaveOccurred
	testcase := func(input string, expected string, opts ...parser.Option) Assertion {
		var err error
		dispenser := NewTestDispenser(input)
		var dest string
		err = parser.ExpectString(dispenser)
		Expect(err).NotTo(HaveOccurred())
		err = parser.ParseString(dispenser, &dest, opts...)
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
		It("should parse string", func() {
			testcase("test string", "string").Should(Succeed())
		})
		It("should only consume first token", func() {
			testcase("test string other", "string").Should(Succeed())
		})
		It("should return an error when empty", func() {
			testcase("test", "").Should(Fail())
		})
		It("Should return an error when only an empty string", func() {
			testcase(`test ""`, "").Should(Fail())
		})
	})
	Context("With AllowEmpty() option", func() {
		It("should parse empty value", func() {
			testcase("test", "", parser.AllowEmpty()).Should(Succeed())
		})
		It("should parse empty string value", func() {
			testcase(`test ""`, "", parser.AllowEmpty()).Should(Succeed())
		})
	})
	Context("With DefaultValue() option", func() {
		It("should parse empty value", func() {
			testcase("test", "default", parser.Default("default")).Should(Succeed())
		})
		It("should use default rather than empty string", func() {
			testcase(`test ""`, "default", parser.Default("default")).Should(Succeed())
		})
		It("should ignore default value when not empty", func() {
			testcase("test string", "string", parser.Default("default")).Should(Succeed())
		})
		Context("With AllowEmpty() option", func() {
			It("should not use default when empty string is present", func() {
				testcase(`test ""`, "", parser.Default("default"), parser.AllowEmpty()).Should(Succeed())
			})
		})
	})
	Context("With Match() option", func() {
		It("should parse string", func() {
			testcase("test string", "string", parser.Match("string")).Should(Succeed())
		})
		It("should return an error when invalid", func() {
			testcase("test invalid", "", parser.Match("string")).Should(Fail())
		})
		Context("With several match candidates", func() {
			It("should parse string", func() {
				testcase("test string", "string", parser.Match("string", "other")).Should(Succeed())
			})
			It("should parse string", func() {
				testcase("test other", "other", parser.Match("string", "other")).Should(Succeed())
			})
			It("should return an error when invalid", func() {
				testcase("test invalid", "", parser.Match("string", "other")).Should(Fail())
			})
		})
	})
	Context("With Inplace() option", func() {
		It("should parse string", func() {
			testcase("string", "string", parser.Inplace()).Should(Succeed())
		})
		It("should not consume next token", func() {
			testcase("string other", "string", parser.Inplace()).Should(Succeed())
		})
		It("should return an error when empty string", func() {
			testcase(`""`, "", parser.Inplace()).Should(Fail())
		})
		Context("With AllowEmpty() option", func() {
			It("should parse empty string", func() {
				testcase(`""`, "", parser.Inplace(), parser.AllowEmpty()).Should(Succeed())
			})
			It("should use default value if empty string", func() {
				testcase(`""`, "default", parser.Inplace(), parser.AllowEmpty(), parser.Default("default")).Should(Succeed())
			})
		})
		Context("With Match() option", func() {
			It("should parse string", func() {
				testcase("string", "string", parser.Inplace(), parser.Match("string")).Should(Succeed())
			})
			It("should return an error when invalid", func() {
				testcase("invalid", "", parser.Inplace(), parser.Match("string")).Should(Fail())
			})
			Context("With several match candidates", func() {
				It("should parse string", func() {
					testcase("string", "string", parser.Inplace(), parser.Match("string", "other")).Should(Succeed())
				})
				It("should parse string", func() {
					testcase("other", "other", parser.Inplace(), parser.Match("string", "other")).Should(Succeed())
				})
				It("should return an error when invalid", func() {
					testcase("invalid", "", parser.Inplace(), parser.Match("string", "other")).Should(Fail())
				})
			})
		})
	})
})
