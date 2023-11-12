// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

var _ = Describe("ParseStringArrayMap", func() {
	var Fail = HaveOccurred
	testcase := func(input string, expected map[string][]string, opts ...parser.Option) Assertion {
		var err error
		dispenser := NewTestDispenser(input)
		var dest map[string][]string
		err = parser.ExpectString(dispenser, parser.Match("test"))
		Expect(err).NotTo(HaveOccurred())
		err = parser.ParseStringArrayMap(dispenser, &dest, opts...)
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
		It("should parse string array map", func() {
			testcase(`
				test {
					key1 string2
				}`,
				map[string][]string{"key1": {"string2"}},
			).Should(Succeed())
		})
		It("should parse a multi-line string array map", func() {
			testcase(`
				test {
					key1 string2 string3
					key2 string4	
				}`,
				map[string][]string{"key1": {"string2", "string3"}, "key2": {"string4"}},
			).Should(Succeed())
		})
		It("should parse duplicate keys as an array", func() {
			testcase(`
				test {
					key1 string2
					key1 string3
				}`,
				map[string][]string{"key1": {"string2", "string3"}},
			).Should(Succeed())
		})
		It("should return an error when empty", func() {
			testcase("test", nil).Should(Fail())
		})
		It("should return an error when empty value for a key", func() {
			testcase(`
				test {
					key1
				}`,
				nil,
			).Should(Fail())
		})
		It("should return an error when not at the beginning of a block", func() {
			testcase(`
				test value1 value2
			`, nil).Should(Fail())
		})
		It("should skip empty strings", func() {
			testcase(`test {
				key1 value1 "" value2
			}`,
				map[string][]string{"key1": {"value1", "value2"}},
			).Should(Succeed())
		})
		It("should return an error if only empty strings", func() {
			testcase(`test {
				key1 "" ""
			}`,
				nil,
			).Should(Fail())
		})
	})
	Context("With AllowEmpty() option", func() {
		It("should allow empty values", func() {
			testcase("test", nil, parser.AllowEmpty()).Should(Succeed())
		})
		It("should not allow key with empty values", func() {
			testcase(`
				test {
					key1
				}`,
				nil,
				parser.AllowEmpty(),
			).Should(Fail())
		})
		It("should return an error if only empty strings", func() {
			testcase(`test {
				key1 "" ""
			}`,
				nil,
			).Should(Fail())
		})
	})
	Context("With AllowEmptyValues() option", func() {
		It("should allow key with empty values", func() {
			testcase(`
				test {
					key1
				}`,
				map[string][]string{"key1": {}},
				parser.AllowEmptyValues(),
			).Should(Succeed())
		})
		It("should allow empty strings as values", func() {
			testcase(`test {
				key1 ""
			}`,
				map[string][]string{"key1": {""}},
				parser.AllowEmptyValues(),
			).Should(Succeed())
		})
	})
	Context("With Default() option", func() {
		It("should use default value when value is empty", func() {
			testcase("test",
				map[string][]string{"key1": {"string2"}},
				parser.Default(map[string][]string{"key1": {"string2"}}),
			).Should(Succeed())
		})
		It("should ignore default value when value is not empty", func() {
			testcase(`
				test {
					key1 string2
				}`,
				map[string][]string{"key1": {"string2"}},
				parser.Default(map[string][]string{"key1": {"string3"}}),
			).Should(Succeed())
		})
	})
	Context("With Inline() option", func() {
		It("should parse inline string array map with single key and singe value", func() {
			testcase(`test key1=string2`,
				map[string][]string{"key1": {"string2"}},
				parser.Inline(parser.Separator("=", ",")),
			).Should(Succeed())
		})
		It("should return an error when epmty", func() {
			testcase(`test`,
				nil,
				parser.Inline(parser.Separator("=", ",")),
			).Should(Fail())
		})
		It("should return an error when only empty strings", func() {
			testcase(`test key1=,,,,`,
				nil,
				parser.Inline(parser.Separator("=", ",")),
			).Should(Fail())
		})
		It("should skip empty string values", func() {
			testcase(`test key1=,string2,string3`,
				map[string][]string{"key1": {"string2", "string3"}},
				parser.Inline(parser.Separator("=", ",")),
			).Should(Succeed())
		})
		It("should parse inline string array map with single key and multiple values", func() {
			testcase(`test key1=string2,string3`,
				map[string][]string{"key1": {"string2", "string3"}},
				parser.Inline(parser.Separator("=", ",")),
			).Should(Succeed())
		})
		It("should parse inline string array map with multiple keys and multiple values", func() {
			testcase(`test key1=string2,string3 key2=string4,string5`,
				map[string][]string{"key1": {"string2", "string3"}, "key2": {"string4", "string5"}},
				parser.Inline(parser.Separator("=", ",")),
			).Should(Succeed())
		})
		Context("With AllowEmpty() option", func() {
			It("should allow empty values", func() {
				testcase(`test`,
					nil,
					parser.AllowEmpty(),
					parser.Inline(parser.Separator("=", ",")),
				).Should(Succeed())
			})
			It("should not allow key with empty values", func() {
				testcase(`test key1=FOO,BAR,BAZ key2=`,
					map[string][]string{"key1": {"FOO", "BAR", "BAZ"}},
					parser.AllowEmpty(),
					parser.Inline(parser.Separator("=", ",")),
				).Should(Fail())
			})
		})
		Context("WithAllowEmptyValues() option", func() {
			It("should not allow empty values", func() {
				testcase(`test`,
					nil,
					parser.AllowEmptyValues(),
					parser.Inline(parser.Separator("=", ",")),
				).Should(Fail())
			})
			It("should allow key with empty values", func() {
				testcase(`test key`,
					map[string][]string{"key": {}},
					parser.AllowEmptyValues(),
					parser.Inline(parser.Separator("=", ",")),
				).Should(Succeed())
			})
			It("should allow empty strings as values", func() {
				testcase(`test key1=,2,3`,
					map[string][]string{"key1": {"", "2", "3"}},
					parser.Inline(parser.Separator("=", ",")),
					parser.AllowEmptyValues(),
				).Should(Succeed())
			})
		})
	})
})
