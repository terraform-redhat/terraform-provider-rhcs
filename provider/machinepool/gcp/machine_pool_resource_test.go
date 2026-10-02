package gcp_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint

	machinepoolgcp "github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool/gcp"
)

func TestGcpMachinePool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GcpMachinePool Suite")
}

var _ = Describe("MachinePoolResource (gcp variant)", func() {
	It("satisfies the framework resource interfaces", func() {
		var _ resource.Resource = &machinepoolgcp.MachinePoolResource{}
		var _ resource.ResourceWithConfigure = &machinepoolgcp.MachinePoolResource{}
		var _ resource.ResourceWithImportState = &machinepoolgcp.MachinePoolResource{}
		var _ resource.ResourceWithConfigValidators = &machinepoolgcp.MachinePoolResource{}
	})

	It("constructs a non-nil instance", func() {
		r := machinepoolgcp.New()
		Expect(r).ToNot(BeNil())
	})

	It("uses rhcs_gcp_machine_pool as TypeName so it does not collide with the Classic resource", func() {
		r := machinepoolgcp.New()
		req := resource.MetadataRequest{ProviderTypeName: "rhcs"}
		resp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), req, resp)
		Expect(resp.TypeName).To(Equal("rhcs_gcp_machine_pool"))
	})

	It("publishes a schema exposing the GCP-specific attributes", func() {
		r := machinepoolgcp.New()
		req := resource.SchemaRequest{}
		resp := &resource.SchemaResponse{}
		r.Schema(context.Background(), req, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
		for _, key := range []string{
			"cluster_id",
			"instance_type",
			"name",
			"gcp",
			"autoscaling",
			"replicas",
		} {
			Expect(resp.Schema.Attributes).To(HaveKey(key), "missing schema attribute %q", key)
		}
	})

	It("returns config validators (autoscaling vs replicas + secure_boot bare-metal)", func() {
		r, ok := machinepoolgcp.New().(resource.ResourceWithConfigValidators)
		Expect(ok).To(BeTrue())
		validators := r.ConfigValidators(context.Background())
		Expect(len(validators)).To(BeNumerically(">=", 3))
	})
})
