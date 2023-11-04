package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseNetworkPortRange", func() {
	Context("ParseNetworkPortRange", func() {
		var Fail = HaveOccurred
		testcase := func(input string, expected []int, opts ...parser.Option) Assertion {
			var err error
			dispenser := NewTestDispenser(input)
			var dest []int
			err = parser.ExpectString(dispenser)
			Expect(err).NotTo(HaveOccurred())
			err = parser.ParseNetworkPortRange(dispenser, &dest, opts...)
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
			It("should parse single port", func() {
				testcase("test 1", []int{1}).Should(Succeed())
				testcase("test 65535", []int{65535}).Should(Succeed())
			})
			It("should parse multiple ports", func() {
				testcase("test 1,2,3", []int{1, 2, 3}).Should(Succeed())
			})
			It("should parse port range", func() {
				testcase("test 1-3", []int{1, 2, 3}).Should(Succeed())
				testcase("test 65533-65535", []int{65533, 65534, 65535}).Should(Succeed())
			})
			It("should return an error when empty", func() {
				testcase("test", nil).Should(Fail())
			})
			It("should consider empty string as an invalid value", func() {
				testcase(`test ""`, nil).Should(Fail())
			})
			It("should return an error when invalid integer", func() {
				testcase("test invalid", nil).Should(Fail())
				testcase("test 65536", nil).Should(Fail())
			})
		})
	})
})
