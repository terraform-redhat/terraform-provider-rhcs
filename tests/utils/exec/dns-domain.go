// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type DnsDomainArgs struct {
	Name        *string `hcl:"name"`
	ID          *string `hcl:"id"`
	ClusterArch *string `hcl:"cluster_arch"`
	SeaplaneURL *string `hcl:"seaplane_url"`
	AWSRegion   *string `hcl:"aws_region"`
}

type DnsDomainOutput struct {
	DnsDomainId string `json:"dns_domain_id,omitempty"`
	ClusterArch string `json:"dns_domain_cluster_arch,omitempty"`
	BaseDomain  string `json:"base_domain,omitempty"`
	Phase       string `json:"phase,omitempty"`
}

type DnsDomainService interface {
	Init() error
	Plan(args *DnsDomainArgs) (string, error)
	Apply(args *DnsDomainArgs) (string, error)
	Output() (*DnsDomainOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*DnsDomainArgs, error)
	DeleteTFVars() error
}

type dnsDomainService struct {
	tfExecutor  TerraformExecutor
	clusterType constants.ClusterType
}

func NewDnsDomainService(tfWorkspace string, clusterType constants.ClusterType) (DnsDomainService, error) {
	svc := &dnsDomainService{
		tfExecutor:  NewTerraformExecutor(tfWorkspace, manifests.GetDnsDomainManifestsDir(clusterType)),
		clusterType: clusterType,
	}
	err := svc.Init()
	return svc, err
}

func (svc *dnsDomainService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *dnsDomainService) Plan(args *DnsDomainArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(svc.withDefaults(args))
}

func (svc *dnsDomainService) Apply(args *DnsDomainArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(svc.withDefaults(args))
}

func (svc *dnsDomainService) Output() (*DnsDomainOutput, error) {
	var output DnsDomainOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *dnsDomainService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *dnsDomainService) ReadTFVars() (*DnsDomainArgs, error) {
	args := &DnsDomainArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *dnsDomainService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}

func (svc *dnsDomainService) withDefaults(args *DnsDomainArgs) *DnsDomainArgs {
	prepared := &DnsDomainArgs{}
	if args != nil {
		*prepared = *args
	}
	if svc.clusterType != constants.SEAPLANE && prepared.ClusterArch == nil {
		clusterArch := "classic"
		if svc.clusterType.HCP {
			clusterArch = "hcp"
		}
		prepared.ClusterArch = &clusterArch
	}
	return prepared
}
