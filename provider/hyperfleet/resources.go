package hyperfleet

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"

	generated "github.com/terraform-redhat/terraform-provider-rhcs/provider/hyperfleet/generated"
)

func NewCluster() resource.Resource {
	return generated.NewCluster(NewClusterHandlerImpl)
}

func NewNodePool() resource.Resource {
	return generated.NewNodePool(NewNodePoolHandlerImpl)
}

func NewOidcConfig() resource.Resource {
	return generated.NewOidcConfig(NewOidcConfigHandlerImpl)
}
