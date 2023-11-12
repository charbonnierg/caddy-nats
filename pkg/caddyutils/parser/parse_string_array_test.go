// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseStringArray", func() {
	Context("Parse string array", func() {
		var Fail = HaveOccurred
		testcase := func(input string, expected []string, opts ...parser.Option) Assertion {
			var err error
			dispenser := NewTestDispenser(input)
			var dest []string
			err = parser.ExpectString(dispenser, parser.Match("test"))
			Expect(err).NotTo(HaveOccurred())
			err = parser.ParseStringArray(dispenser, &dest, opts...)
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
			It("should parse a string array", func() {
				testcase("test a b c", []string{"a", "b", "c"}).Should(Succeed())
				testcase("test a", []string{"a"}).Should(Succeed())
			})
			It("should skip empty tokens", func() {
				testcase("test a ", []string{"a"}).Should(Succeed())
			})
			It("shoud accept spaces when tokens are wrapped with quotes", func() {
				testcase(`test "a b c"`, []string{"a b c"}).Should(Succeed())
			})
			It("should return an error if empty", func() {
				testcase("test", nil).Should(Fail())
			})
			It("should skip empty strings", func() {
				testcase(`test 1 "" 2`, []string{"1", "2"}).Should(Succeed())
			})
		})
		Context("With AllowEmpty option", func() {
			It("should return nil if empty", func() {
				testcase("test", nil, parser.AllowEmpty()).Should(Succeed())
			})
			It("should skip empty strings", func() {
				testcase(`test 1 "" 2`, []string{"1", "2"}).Should(Succeed())
			})
		})
		Context("With AllowEmptyValues() option", func() {
			It("should return array with single empty string", func() {
				testcase(`test ""`, []string{""}, parser.AllowEmptyValues()).Should(Succeed())
			})
			It("should preserve empty strings among other values", func() {
				testcase(`test 1 "" 2`, []string{"1", "", "2"}, parser.AllowEmptyValues()).Should(Succeed())
			})
		})
		Context("With Default option", func() {
			It("should return default value if empty", func() {
				testcase("test", []string{"a", "b", "c"}, parser.Default([]string{"a", "b", "c"})).Should(Succeed())
			})
			It("should ignore default value if not empty", func() {
				testcase("test a", []string{"a"}, parser.Default([]string{"b", "c"})).Should(Succeed())
			})
		})
		Context("With Separator option", func() {
			It("should return an error if empty", func() {
				testcase("test", nil, parser.Separator("/")).Should(Fail())
			})
			It("should return an error if only empty string", func() {
				testcase(`test ""`, nil, parser.Separator("/")).Should(Fail())
			})
			It("should parse a string array", func() {
				testcase(`test a/b/c`, []string{"a", "b", "c"}, parser.Separator("/")).Should(Succeed())
				testcase(`test a/`, []string{"a"}, parser.Separator("/")).Should(Succeed())
			})
			It("should should only consume first token", func() {
				testcase(`test a/b/c d e f`, []string{"a", "b", "c"}, parser.Separator("/")).Should(Succeed())
			})
			It("should skip empty strings", func() {
				testcase(`test a//b`, []string{"a", "b"}, parser.Separator("/")).Should(Succeed())
			})
			It("should return an error when several separators are used", func() {
				testcase(`test a/b,c`, nil, parser.Separator("/", ",")).Should(Fail())
			})
			Context("With AllowEmptyValues() option", func() {
				It("should return array with single empty string", func() {
					testcase(`test ""`, []string{""}, parser.AllowEmptyValues(), parser.Separator("/")).Should(Succeed())
				})
				It("should only consume a single empty string", func() {
					testcase(`test "" ""`, []string{""}, parser.AllowEmptyValues(), parser.Separator("/")).Should(Succeed())
				})
				It("should preserve empty strings among other values", func() {
					testcase(`test a//b`, []string{"a", "", "b"}, parser.AllowEmptyValues(), parser.Separator("/")).Should(Succeed())
				})
			})
		})
		Context("With Reverse option", func() {
			It("should parse a string array in reverse order", func() {
				testcase(`test a b c`, []string{"c", "b", "a"}, parser.Reverse()).Should(Succeed())
			})
			Context("With Separator option", func() {
				It("should parse a string array in reverse order", func() {
					testcase(`test a,b,c`, []string{"c", "b", "a"}, parser.Reverse(), parser.Separator(",")).Should(Succeed())
				})
			})
		})
	})

})
