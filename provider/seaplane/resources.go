// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	generated "github.com/terraform-redhat/terraform-provider-rhcs/provider/seaplane/generated"
)

func NewCluster() resource.Resource {
	return &clusterResourceWithSchemaFixes{
		ClusterResource: generated.NewCluster(NewClusterHandlerImpl).(*generated.ClusterResource),
	}
}

func NewNodePool() resource.Resource {
	return generated.NewNodePool(NewNodePoolHandlerImpl)
}

func NewOidcConfig() resource.Resource {
	return &oidcConfigResourceWithSchemaFixes{
		OidcConfigResource: generated.NewOidcConfig(NewOidcConfigHandlerImpl).(*generated.OidcConfigResource),
	}
}

func NewDNSDomain() resource.Resource {
	return &DNSDomainSeaplaneResource{}
}

// The current pathbind generator emits RequiresReplace before
// UseStateForUnknown for immutable computed attributes. Terraform evaluates
// these modifiers in order, so the unknown value can trigger a replacement
// before the prior state is restored. It also drops Optional on the
// Optional+Computed no_cni attribute. Apply the intended schema semantics
// after generation so regenerating the pathbind resources keeps this fix.
type clusterResourceWithSchemaFixes struct {
	*generated.ClusterResource
}

var _ resource.Resource = (*clusterResourceWithSchemaFixes)(nil)
var _ resource.ResourceWithConfigure = (*clusterResourceWithSchemaFixes)(nil)
var _ resource.ResourceWithImportState = (*clusterResourceWithSchemaFixes)(nil)

func (r *clusterResourceWithSchemaFixes) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	r.ClusterResource.Schema(ctx, req, resp)
	if err := fixClusterResourceSchema(&resp.Schema); err != nil {
		resp.Diagnostics.AddError("Failed to fix generated Seaplane cluster schema", err.Error())
	}
}

type oidcConfigResourceWithSchemaFixes struct {
	*generated.OidcConfigResource
}

var _ resource.Resource = (*oidcConfigResourceWithSchemaFixes)(nil)
var _ resource.ResourceWithConfigure = (*oidcConfigResourceWithSchemaFixes)(nil)
var _ resource.ResourceWithImportState = (*oidcConfigResourceWithSchemaFixes)(nil)

func (r *oidcConfigResourceWithSchemaFixes) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	r.OidcConfigResource.Schema(ctx, req, resp)
	if err := useStateBeforeReplaceString(resp.Schema.Attributes, "issuer_url"); err != nil {
		resp.Diagnostics.AddError("Failed to fix generated Seaplane OIDC config schema", err.Error())
	}
}

func fixClusterResourceSchema(s *schema.Schema) error {
	if err := useStateBeforeReplaceString(s.Attributes, "cloud_region"); err != nil {
		return err
	}

	noCNI, ok := s.Attributes["no_cni"].(schema.BoolAttribute)
	if !ok {
		return fmt.Errorf("no_cni is not a boolean attribute")
	}
	noCNI.Optional = true
	noCNI.Computed = true
	noCNI.PlanModifiers = []planmodifier.Bool{
		boolplanmodifier.UseStateForUnknown(),
		boolplanmodifier.RequiresReplace(),
	}
	s.Attributes["no_cni"] = noCNI

	aws, ok := s.Attributes["aws"].(schema.SingleNestedAttribute)
	if !ok {
		return fmt.Errorf("aws is not a single nested attribute")
	}
	if err := useStateBeforeReplaceString(aws.Attributes, "aws_partition"); err != nil {
		return err
	}
	s.Attributes["aws"] = aws

	return nil
}

func useStateBeforeReplaceString(attributes map[string]schema.Attribute, name string) error {
	attribute, ok := attributes[name].(schema.StringAttribute)
	if !ok {
		return fmt.Errorf("%s is not a string attribute", name)
	}
	attribute.PlanModifiers = []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
		stringplanmodifier.RequiresReplace(),
	}
	attributes[name] = attribute
	return nil
}
