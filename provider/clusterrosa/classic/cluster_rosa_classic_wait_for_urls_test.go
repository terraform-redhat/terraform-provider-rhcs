// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package classic

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// nullObjectValue returns an object value of the given type whose attributes are all null.
func nullObjectValue(ctx context.Context, objectType attr.Type) tftypes.Value {
	tfType := objectType.TerraformType(ctx).(tftypes.Object)
	attrs := map[string]tftypes.Value{}
	for name, attrType := range tfType.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	return tftypes.NewValue(tfType, attrs)
}

var _ = Describe("Rosa Classic wait_for_urls", func() {
	ctx := context.Background()

	resourceSchema := func() schema.Schema {
		resp := &resource.SchemaResponse{}
		New().Schema(ctx, resource.SchemaRequest{}, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
		return resp.Schema
	}

	datasourceSchema := func() dsschema.Schema {
		resp := &datasource.SchemaResponse{}
		NewDataSource().Schema(ctx, datasource.SchemaRequest{}, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
		return resp.Schema
	}

	It("is an optional resource attribute without default, computed value or replacement", func() {
		attribute, ok := resourceSchema().Attributes["wait_for_urls"].(schema.BoolAttribute)
		Expect(ok).To(BeTrue())
		Expect(attribute.Optional).To(BeTrue())
		Expect(attribute.Required).To(BeFalse())
		// Not computed and no default: configurations and states written before the
		// attribute existed keep a null value, so upgrading the provider shows no diff.
		Expect(attribute.Computed).To(BeFalse())
		Expect(attribute.Default).To(BeNil())
		// No plan modifiers, so changing the value never plans a replacement.
		Expect(attribute.PlanModifiers).To(BeEmpty())
	})

	It("is a computed-only data source attribute", func() {
		attribute, ok := datasourceSchema().Attributes["wait_for_urls"].(dsschema.BoolAttribute)
		Expect(ok).To(BeTrue())
		Expect(attribute.Computed).To(BeTrue())
		Expect(attribute.Optional).To(BeFalse())
		Expect(attribute.Required).To(BeFalse())
	})

	It("keeps the shared state struct aligned with the resource schema", func() {
		s := resourceSchema()
		state := tfsdk.State{Schema: s, Raw: nullObjectValue(ctx, s.Type())}
		model := &ClusterRosaClassicState{}
		diags := state.Get(ctx, model)
		Expect(diags.HasError()).To(BeFalse(), "%v", diags)
		// A null value (for example from a state written before the attribute existed)
		// is treated as disabled.
		Expect(model.WaitForURLs.IsNull()).To(BeTrue())
	})

	It("keeps the shared state struct aligned with the data source schema", func() {
		s := datasourceSchema()
		config := tfsdk.Config{Schema: s, Raw: nullObjectValue(ctx, s.Type())}
		model := &ClusterRosaClassicState{}
		diags := config.Get(ctx, model)
		Expect(diags.HasError()).To(BeFalse(), "%v", diags)
		Expect(model.WaitForURLs.IsNull()).To(BeTrue())
	})

	It("is preserved by state population, which reports missing URLs as empty", func() {
		clusterJson := generateBasicRosaClassicClusterJson()
		delete(clusterJson, "api")
		delete(clusterJson, "console")
		clusterJsonString, err := json.Marshal(clusterJson)
		Expect(err).ToNot(HaveOccurred())
		clusterObject, err := cmv1.UnmarshalCluster(clusterJsonString)
		Expect(err).ToNot(HaveOccurred())

		state := &ClusterRosaClassicState{
			WaitForURLs: types.BoolValue(true),
			APIURL:      types.StringValue("https://stale-api.example.com"),
			ConsoleURL:  types.StringValue("https://stale-console.example.com"),
		}
		Expect(populateRosaClassicClusterState(ctx, clusterObject, state, mockHttpClient)).To(Succeed())

		Expect(state.WaitForURLs).To(Equal(types.BoolValue(true)))
		Expect(state.APIURL).To(Equal(types.StringValue("")))
		Expect(state.ConsoleURL).To(Equal(types.StringValue("")))
	})
})
