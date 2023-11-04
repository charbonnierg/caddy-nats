package parser_test

import (
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParsePermissions", func() {
	var Fail = HaveOccurred
	testcase := func(input string, expected fs.FileMode, opts ...parser.Option) Assertion {
		var err error
		dispenser := NewTestDispenser(input)
		var dest fs.FileMode
		err = parser.ExpectString(dispenser)
		Expect(err).NotTo(HaveOccurred())
		err = parser.ParsePermissions(dispenser, &dest, opts...)
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
		It("should parse permissions", func() {
			// octal
			testcase("chmod 0755", 0755).Should(Succeed())
		})
		It("should return an error when empty", func() {
			testcase("chmod", 0).Should(Fail())
		})
		It("should return an error when empty string", func() {
			testcase(`chmod ""`, 0).Should(Fail())
		})
		It("should return an error when invalid permissions", func() {
			testcase("chmod invalid", 0).Should(Fail())
			testcase("chmod 999", 0).Should(Fail())
		})
	})
	Context("With Default() option", func() {
		It("should return default when empty", func() {
			testcase("chmod", 0755, parser.Default(fs.FileMode(0755))).Should(Succeed())
		})
		It("should return default when empty string", func() {
			testcase(`chmod ""`, 0777, parser.Default(fs.FileMode(0777))).Should(Succeed())
		})
		It("should ignore default when not empty", func() {
			testcase("chmod 0644", 0644, parser.Default(fs.FileMode(0755))).Should(Succeed())
		})
		It("should return an error when invalid", func() {
			testcase("chmod invalid", 0, parser.Default(fs.FileMode(0644))).Should(Fail())
		})
	})
	Context("With Inplace() option", func() {
		It("should parse permissions", func() {
			testcase("0755", 0755, parser.Inplace()).Should(Succeed())
		})
		It("should return an error when empty string", func() {
			testcase(`""`, 0, parser.Inplace()).Should(Fail())
		})
		It("should return an error when invalid permissions", func() {
			testcase("invalid", 0, parser.Inplace()).Should(Fail())
			testcase("999", 0, parser.Inplace()).Should(Fail())
		})
	})
})
