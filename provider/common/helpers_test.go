package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
)

var _ = Describe("Helper function tests", func() {
	Context("ShouldPatchStringMap", func() {
		map0 := types.Map{}
		map1 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key1": types.String{
					Value: "val1",
				},
			},
		}
		map2 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key": types.String{
					Value: "val1",
				},
			},
		}
		map3 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key1": types.String{
					Value: "val",
				},
			},
		}
		map4 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key1": types.String{
					Value: "val",
				},
				"key2": types.String{
					Value: "val2",
				},
			},
		}

		It("Should return true when maps are different", func() {
			r, ok := ShouldPatchMap(map0, map1)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map1))

			r, ok = ShouldPatchMap(map1, map0)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map0))

			r, ok = ShouldPatchMap(map1, map2)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map2))

			r, ok = ShouldPatchMap(map1, map3)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map3))

			r, ok = ShouldPatchMap(map1, map4)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map4))
		})

		It("Should return false when maps are identical", func() {
			_, ok := ShouldPatchMap(map1, map1)
			Expect(ok).ToNot(BeTrue())
		})
	})
})
