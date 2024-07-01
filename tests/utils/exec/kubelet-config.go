package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type KubeletConfigArgs struct {
	Cluster             *string `hcl:"cluster"`
	PodPidsLimit        *int    `hcl:"pod_pids_limit"`
	KubeletConfigNumber *int    `hcl:"kubelet_config_number"`
	NamePrefix          *string `hcl:"name_prefix"`
}

type KubeletConfig struct {
	Cluster      string `json:"cluster,omitempty"`
	PodPidsLimit int    `json:"pod_pids_limit,omitempty"`
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
}
type KubeletConfigs struct {
	KubeConfigs []*KubeletConfig `json:"kubelet_configs,omitempty"`
}

type KubeletConfigService interface {
	Init() error
	Plan(args *KubeletConfigArgs) (string, error)
	Apply(args *KubeletConfigArgs) (string, error)
	Output() ([]*KubeletConfig, error)
	Destroy() (string, error)

	ReadTFVars() (*KubeletConfigArgs, error)
	WriteTFVars(args *KubeletConfigArgs) error
	DeleteTFVars() error
}

type kubeletConfigService struct {
	tfExecutor TerraformExecutor
}

func NewKubeletConfigService(manifestsDirs ...string) (KubeletConfigService, error) {
	manifestsDir := constants.KubeletConfigDir
	if len(manifestsDirs) > 0 {
		manifestsDir = manifestsDirs[0]
	}
	svc := &kubeletConfigService{
		tfExecutor: NewTerraformExecutor(manifestsDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *kubeletConfigService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *kubeletConfigService) Plan(args *KubeletConfigArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *kubeletConfigService) Apply(args *KubeletConfigArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *kubeletConfigService) Output() ([]*KubeletConfig, error) {
	output := &KubeletConfigs{}
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return output.KubeConfigs, nil
}

func (svc *kubeletConfigService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *kubeletConfigService) ReadTFVars() (*KubeletConfigArgs, error) {
	args := &KubeletConfigArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *kubeletConfigService) WriteTFVars(args *KubeletConfigArgs) error {
	err := svc.tfExecutor.WriteTerraformVars(args)
	return err
}

func (svc *kubeletConfigService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
