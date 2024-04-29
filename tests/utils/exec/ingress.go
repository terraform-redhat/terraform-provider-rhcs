package exec

import (
	"context"
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type IngressArgs struct {
	ID      string  `json:"id,omitempty"`
	Cluster *string `json:"cluster,omitempty"`

	// Classic
	RouteNamespaceOwnershipPolicy string            `json:"route_namespace_ownership_policy,omitempty"`
	RouteWildcardPolicy           string            `json:"route_wildcard_policy,omitempty"`
	ClusterRoutesHostename        string            `json:"cluster_routes_hostname,omitempty"`
	ClusterRoutestlsSecretRef     string            `json:"cluster_routes_tls_secret_ref,omitempty"`
	LoadBalancerType              string            `json:"load_balancer_type,omitempty"`
	ExcludedNamespaces            []string          `json:"excluded_namespaces,omitempty"`
	RouteSelectors                map[string]string `json:"route_selectors,omitempty"`

	// HCP
	ListeningMethod *string `json:"listening_method,omitempty"`
}

type IngressOutput struct {
	ID string `json:"id,omitempty"`
}

type IngressService struct {
	CreationArgs *IngressArgs
	ManifestDir  string
	Context      context.Context
}

func (ing *IngressService) Init(manifestDirs ...string) error {
	ing.ManifestDir = constants.ClassicIngressDir
	if len(manifestDirs) != 0 {
		ing.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	ing.Context = ctx
	err := runTerraformInit(ctx, ing.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (ing *IngressService) Apply(createArgs *IngressArgs, extraArgs ...string) error {
	ing.CreationArgs = createArgs
	args, _ := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(ing.Context, ing.ManifestDir, args...)
	if err != nil {
		return err
	}
	return nil
}

func (ing *IngressService) Output() (IngressOutput, error) {
	ingressDir := constants.ClassicIngressDir
	if ing.ManifestDir != "" {
		ingressDir = ing.ManifestDir
	}
	var ingressOut IngressOutput
	out, err := runTerraformOutput(context.TODO(), ingressDir)
	if err != nil {
		return ingressOut, err
	}
	ingressOut = IngressOutput{
		ID: helper.DigString(out["id"], "value"),
	}

	return ingressOut, nil
}

func (ing *IngressService) Destroy(createArgs ...*IngressArgs) error {
	if ing.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := ing.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(ing.Context, ing.ManifestDir, args...)

	return err
}

func NewIngressService(manifestDir ...string) (*IngressService, error) {
	ing := &IngressService{}
	err := ing.Init(manifestDir...)
	return ing, err
}
