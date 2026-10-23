package clusterosdgcp_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterosdgcp"
)

func TestClusterOsdGcp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterOsdGcp Suite")
}

var _ = Describe("ClusterOsdGcpResource", func() {
	It("satisfies the framework resource interfaces", func() {
		var _ resource.Resource = &clusterosdgcp.ClusterOsdGcpResource{}
		var _ resource.ResourceWithConfigure = &clusterosdgcp.ClusterOsdGcpResource{}
		var _ resource.ResourceWithImportState = &clusterosdgcp.ClusterOsdGcpResource{}
		var _ resource.ResourceWithConfigValidators = &clusterosdgcp.ClusterOsdGcpResource{}
	})

	It("constructs a non-nil instance", func() {
		r := clusterosdgcp.New()
		Expect(r).ToNot(BeNil())
	})

	It("sets the rhcs_cluster_osd_gcp TypeName", func() {
		r := clusterosdgcp.New()
		req := resource.MetadataRequest{ProviderTypeName: "rhcs"}
		resp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), req, resp)
		Expect(resp.TypeName).To(Equal("rhcs_cluster_osd_gcp"))
	})

	It("publishes a schema exposing the four GCP feature attributes", func() {
		r := clusterosdgcp.New()
		req := resource.SchemaRequest{}
		resp := &resource.SchemaResponse{}
		r.Schema(context.Background(), req, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
		for _, key := range []string{
			"wif_config_id",           // WIF
			"gcp_encryption_key",      // CMEK
			"private_service_connect", // PSC
			"gcp_network",             // Shared VPC
			"private",
			"security",
			"admin_credentials",
			"create_admin_user",
		} {
			Expect(resp.Schema.Attributes).To(HaveKey(key), "missing schema attribute %q", key)
		}
	})

	It("returns the availability-zones cross-attribute validator", func() {
		r, ok := clusterosdgcp.New().(resource.ResourceWithConfigValidators)
		Expect(ok).To(BeTrue())
		validators := r.ConfigValidators(context.Background())
		Expect(validators).ToNot(BeEmpty())
	})
})
