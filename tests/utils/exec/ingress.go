package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type IngressArgs struct {
	Cluster *string `hcl:"cluster"`

	// Classic
	ID                            *string                            `hcl:"id"`
	RouteNamespaceOwnershipPolicy *string                            `hcl:"route_namespace_ownership_policy"`
	RouteWildcardPolicy           *string                            `hcl:"route_wildcard_policy"`
	ClusterRoutesHostename        *string                            `hcl:"cluster_routes_hostname"`
	ClusterRoutestlsSecretRef     *string                            `hcl:"cluster_routes_tls_secret_ref"`
	LoadBalancerType              *string                            `hcl:"load_balancer_type"`
	ExcludedNamespaces            *[]string                          `hcl:"excluded_namespaces"`
	RouteSelectors                *map[string]string                 `hcl:"route_selectors"`
	ComponentRoutes               *map[string]*IngressComponentRoute `hcl:"component_routes"`

	// HCP
	ListeningMethod *string `hcl:"listening_method"`
}

type IngressComponentRoute struct {
	Hostname     *string `cty:"hostname"`
	TlsSecretRef *string `cty:"tls_secret_ref"`
}

type IngressOutput struct {
	ID string `json:"id,omitempty"`
}

type IngressService interface {
	Init() error
	Plan(args *IngressArgs) (string, error)
	Apply(args *IngressArgs) (string, error)
	Output() (*IngressOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*IngressArgs, error)
	DeleteTFVars() error
}

type ingressService struct {
	tfExecutor TerraformExecutor
}

func NewIngressService(manifestsDirs ...string) (IngressService, error) {
	manifestsDir := constants.ClassicIngressDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &ingressService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *ingressService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *ingressService) Plan(args *IngressArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *ingressService) Apply(args *IngressArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *ingressService) Output() (*IngressOutput, error) {
	var output IngressOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *ingressService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *ingressService) ReadTFVars() (*IngressArgs, error) {
	args := &IngressArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *ingressService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}

func NewIngressComponentRoute(hostname, tlsSecretRef *string) *IngressComponentRoute {
	return &IngressComponentRoute{
		Hostname:     hostname,
		TlsSecretRef: tlsSecretRef,
	}
}
