package common

***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
***REMOVED***             // nolint
***REMOVED***

var _ = Describe("Helper function tests", func(***REMOVED*** {
	Context("ShouldPatchStringMap", func(***REMOVED*** {
		map0 := types.Map{}
		map1 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key1": types.String{
					Value: "val1",
		***REMOVED***,
	***REMOVED***,
***REMOVED***
		map2 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key": types.String{
					Value: "val1",
		***REMOVED***,
	***REMOVED***,
***REMOVED***
		map3 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key1": types.String{
					Value: "val",
		***REMOVED***,
	***REMOVED***,
***REMOVED***
		map4 := types.Map{
			ElemType: types.StringType,
			Elems: map[string]attr.Value{
				"key1": types.String{
					Value: "val",
		***REMOVED***,
				"key2": types.String{
					Value: "val2",
		***REMOVED***,
	***REMOVED***,
***REMOVED***

		It("Should return true when maps are different", func(***REMOVED*** {
			r, ok := ShouldPatchMap(map0, map1***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
			Expect(r***REMOVED***.To(Equal(map1***REMOVED******REMOVED***

			r, ok = ShouldPatchMap(map1, map0***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
			Expect(r***REMOVED***.To(Equal(map0***REMOVED******REMOVED***

			r, ok = ShouldPatchMap(map1, map2***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
			Expect(r***REMOVED***.To(Equal(map2***REMOVED******REMOVED***

			r, ok = ShouldPatchMap(map1, map3***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
			Expect(r***REMOVED***.To(Equal(map3***REMOVED******REMOVED***

			r, ok = ShouldPatchMap(map1, map4***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED******REMOVED***
			Expect(r***REMOVED***.To(Equal(map4***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Should return false when maps are identical", func(***REMOVED*** {
			_, ok := ShouldPatchMap(map1, map1***REMOVED***
			Expect(ok***REMOVED***.ToNot(BeTrue(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
