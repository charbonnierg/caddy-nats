// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package datatypes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/datatypes"
)

var _ = Describe("Set", func() {
	Context("an empty set", func() {
		s := datatypes.NewSet[string]()
		It("should be considered empty", func() {
			Expect(s.IsEmpty()).To(BeTrue())
		})
		It("should be equal to itself", func() {
			Expect(s.IsEqual(s)).To(BeTrue())
		})
		It("should be equal to another empty set", func() {
			s2 := datatypes.NewSet[string]()
			Expect(s.IsEqual(s2)).To(BeTrue())
		})
		It("should return an empty slice", func() {
			Expect(s.Slice()).To(BeEmpty())
		})
	})
	Context("a set with a single element", func() {
		s := datatypes.NewSet[string]("a")
		It("should return a set with one element", func() {
			Expect(s.IsEmpty()).To(BeFalse())
			Expect(s.Contains("a")).To(BeTrue())
		})
		It("should be equal to itself", func() {
			Expect(s.IsEqual(s)).To(BeTrue())
		})
		It("should be equal to another set with the same element", func() {
			s2 := datatypes.NewSet[string]("a")
			Expect(s.IsEqual(s2)).To(BeTrue())
		})
		It("should return a slice with one element", func() {
			Expect(s.Slice()).To(Equal([]string{"a"}))
		})
	})
	Context("with several unique arguments", func() {
		s := datatypes.NewSet[string]("a", "b", "c")
		It("should return a set with all elements", func() {
			Expect(s.IsEmpty()).To(BeFalse())
			Expect(s.Contains("a")).To(BeTrue())
			Expect(s.Contains("b")).To(BeTrue())
			Expect(s.Contains("c")).To(BeTrue())
		})
		It("should be equal to itself", func() {
			Expect(s.IsEqual(s)).To(BeTrue())
		})
		It("should be equal to another set with the same elements", func() {
			s2 := datatypes.NewSet[string]("a", "b", "c")
			Expect(s.IsEqual(s2)).To(BeTrue())
		})
		It("should not be equal to another set with less element", func() {
			s2 := datatypes.NewSet[string]("a", "b")
			Expect(s.IsEqual(s2)).To(BeFalse())
		})
		It("should not be equal to another set with more element", func() {
			s2 := datatypes.NewSet[string]("a", "b", "c", "d")
			Expect(s.IsEqual(s2)).To(BeFalse())
		})
		It("should not be equal to another set with different elements", func() {
			s2 := datatypes.NewSet[string]("a", "b", "d")
			Expect(s.IsEqual(s2)).To(BeFalse())
		})
		It("should return a slice with all elements in any order", func() {
			Expect(s.Slice()).To(ContainElements("a", "b", "c"))
		})
	})
})

var _ = Describe("Delete", func() {
	Context("from an empty set", func() {
		s := datatypes.NewSet[string]()
		It("does nothing", func() {
			Expect(s.IsEmpty()).To(BeTrue())
			s.Remove("a")
			Expect(s.IsEmpty()).To(BeTrue())
		})
	})
	Context("from a set without the value", func() {
		s := datatypes.NewSet[string]("a")
		It("does nothing", func() {
			s.Remove("b")
			Expect(s.Contains("a")).To(BeTrue())
		})
	})
	Context("from a set with the value", func() {
		s := datatypes.NewSet[string]("a")
		It("removes the value", func() {
			s.Remove("a")
			Expect(s.Contains("a")).To(BeFalse())
			Expect(s.IsEmpty()).To(BeTrue())
		})
	})
})
