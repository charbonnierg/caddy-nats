// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ExpectString", func() {
	Context("Expect string", func() {
		var Fail = HaveOccurred
		testcase := func(input string, opts ...parser.Option) Assertion {
			dispenser := NewTestDispenser(input)
			// First expect is to discard the first token
			Expect(parser.ExpectString(dispenser)).To(Succeed())
			// Second expect is to test the string
			err := parser.ExpectString(dispenser, opts...)
			return Expect(err)
		}
		Context("Without options", func() {
			It("should parse any string", func() {
				testcase("test astring").Should(Succeed())
			})
			It("should return an error if empty", func() {
				testcase("test").Should(Fail())
			})
		})
		Context("With Match() option", func() {
			It("should parse a string", func() {
				testcase("test astring", parser.Match("astring")).Should(Succeed())
			})
			It("should return an error if string does not match", func() {
				testcase("test astring", parser.Match("anotherstring")).Should(Fail())
			})
		})
		Context("With several Match() option values", func() {
			It("should parse a string", func() {
				testcase("test astring", parser.Match("astring", "anotherstring")).Should(Succeed())
			})
			It("should return an error if string does not match", func() {
				testcase("test astring", parser.Match("anotherstring", "yetanotherstring")).Should(Fail())
			})
		})
		Context("With Inplace() option", func() {
			It("should parse a string", func() {
				testcase("astring", parser.Inplace(), parser.Match("astring")).Should(Succeed())
			})
			It("should return an error if string does not match", func() {
				testcase("anotherstring", parser.Inplace(), parser.Match("astring")).Should(Fail())
			})
		})
	})
})
