package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseStringMap", func() {
	var Fail = HaveOccurred
	testcase := func(input string, expected map[string]string, opts ...parser.Option) Assertion {
		var err error
		dispenser := NewTestDispenser(input)
		var dest map[string]string
		err = parser.ExpectString(dispenser, parser.Match("test"))
		Expect(err).NotTo(HaveOccurred())
		err = parser.ParseStringMap(dispenser, &dest, opts...)
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
		It("should parse string map", func() {
			testcase(`
				test {
					string1 string2
				}`,
				map[string]string{"string1": "string2"},
			).Should(Succeed())
		})
		It("should parse a multi-line string map", func() {
			testcase(`
				test {
					string1 string2
					string3 string4	
				}`,
				map[string]string{"string1": "string2", "string3": "string4"},
			).Should(Succeed())
		})
		It("should return an error when empty", func() {
			testcase("test", nil).Should(Fail())
		})
	})
	Context("With default option", func() {
		It("should use default value when value is empty", func() {
			testcase("test",
				map[string]string{"string1": "string2"},
				parser.Default(map[string]string{"string1": "string2"}),
			).Should(Succeed())
		})
		It("should ignore default value when value is not empty", func() {
			testcase(`
				test {
					string1 string2
				}`,
				map[string]string{"string1": "string2"},
				parser.Default(map[string]string{"string1": "string3"}),
			).Should(Succeed())
		})
	})
	Context("With allow empty option", func() {
		It("should parse empty value", func() {
			testcase("test", nil, parser.AllowEmpty()).Should(Succeed())
		})
	})
	Context("With inline option", func() {
		It("should parse string map with default separator", func() {
			testcase(
				`test string1=string2 string3=string4`,
				map[string]string{"string1": "string2", "string3": "string4"},
				parser.Inline(),
			).Should(Succeed())
		})
		It("should split by separator only once per token", func() {
			testcase(
				`test build-arg=FOO=BAR`,
				map[string]string{"build-arg": "FOO=BAR"},
				parser.Inline(),
			).Should(Succeed())
		})
		It("should parse string map with custom separator", func() {
			testcase(
				`test string1:string2 string3:string4`,
				map[string]string{"string1": "string2", "string3": "string4"},
				parser.Inline(parser.Separator(":")),
			).Should(Succeed())
		})
		It("should return an error when empty", func() {
			testcase("test", nil, parser.Inline()).Should(Fail())
		})
	})
})
