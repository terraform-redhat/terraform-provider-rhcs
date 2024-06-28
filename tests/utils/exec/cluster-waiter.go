package exec

import "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"

type ClusterWaiterArgs struct {
	Cluster      *string `hcl:"cluster_id"`
	TimeoutInMin *int    `hcl:"timeout_in_min"`
}

type ClusterWaiterOutput struct {
	Cluster      string `json:"cluster_id,omitempty"`
	TimeoutInMin int    `json:"timeout_in_min,omitempty"`
}

type ClusterWaiterService interface {
	Init() error
	Plan(args *ClusterWaiterArgs) (string, error)
	Apply(args *ClusterWaiterArgs) (string, error)
	Output() (*ClusterWaiterOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*ClusterWaiterArgs, error)
	DeleteTFVars() error
}

type clusterWaiterService struct {
	tfExecutor TerraformExecutor
}

func NewClusterWaiterService() (ClusterWaiterService, error) {
	svc := &clusterWaiterService{
		tfExecutor: NewTerraformExecutor(constants.ClusterWaiterDir),
	}
	err := svc.Init()
	return svc, err
}

func (svc *clusterWaiterService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *clusterWaiterService) Plan(args *ClusterWaiterArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *clusterWaiterService) Apply(args *ClusterWaiterArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *clusterWaiterService) Output() (*ClusterWaiterOutput, error) {
	var output ClusterWaiterOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *clusterWaiterService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *clusterWaiterService) ReadTFVars() (*ClusterWaiterArgs, error) {
	args := &ClusterWaiterArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *clusterWaiterService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}
