package wifconfig_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/wifconfig"
)

func TestWifConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WifConfig Suite")
}

var _ = Describe("WifConfigResource", func() {
	It("satisfies the framework resource interfaces", func() {
		var _ resource.Resource = &wifconfig.WifConfigResource{}
		var _ resource.ResourceWithConfigure = &wifconfig.WifConfigResource{}
		var _ resource.ResourceWithImportState = &wifconfig.WifConfigResource{}
		var _ resource.ResourceWithModifyPlan = &wifconfig.WifConfigResource{}
	})

	It("constructs a non-nil instance", func() {
		r := wifconfig.New()
		Expect(r).ToNot(BeNil())
	})

	It("sets the rhcs_wif_config TypeName", func() {
		r := wifconfig.New()
		req := resource.MetadataRequest{ProviderTypeName: "rhcs"}
		resp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), req, resp)
		Expect(resp.TypeName).To(Equal("rhcs_wif_config"))
	})

	It("publishes a schema with no diagnostics", func() {
		r := wifconfig.New()
		req := resource.SchemaRequest{}
		resp := &resource.SchemaResponse{}
		r.Schema(context.Background(), req, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
		Expect(resp.Schema.Attributes).To(HaveKey("display_name"))
		Expect(resp.Schema.Attributes).To(HaveKey("gcp"))
		Expect(resp.Schema.Attributes).To(HaveKey("id"))
	})
})

var _ = Describe("WifConfigDataSource", func() {
	It("satisfies the framework data source interfaces", func() {
		var _ datasource.DataSource = &wifconfig.WifConfigDataSource{}
		var _ datasource.DataSourceWithConfigure = &wifconfig.WifConfigDataSource{}
	})

	It("constructs a non-nil instance", func() {
		d := wifconfig.NewDataSource()
		Expect(d).ToNot(BeNil())
	})

	It("sets the rhcs_wif_config TypeName", func() {
		d := wifconfig.NewDataSource()
		req := datasource.MetadataRequest{ProviderTypeName: "rhcs"}
		resp := &datasource.MetadataResponse{}
		d.Metadata(context.Background(), req, resp)
		Expect(resp.TypeName).To(Equal("rhcs_wif_config"))
	})

	It("publishes a schema with no diagnostics and exposes lookup keys", func() {
		d := wifconfig.NewDataSource()
		req := datasource.SchemaRequest{}
		resp := &datasource.SchemaResponse{}
		d.Schema(context.Background(), req, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
		Expect(resp.Schema.Attributes).To(HaveKey("display_name"))
		Expect(resp.Schema.Attributes).To(HaveKey("id"))
		Expect(resp.Schema.Attributes).To(HaveKey("gcp"))
	})
})
