package exec

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
)

type ImportArgs struct {
	ClusterID  string
	Resource   string
	ObjectName string
}

type StateOutput struct {
	Resource string
}

type StatesOutput struct {
	States []StateOutput
}

type ImportService interface {
	Init() error
	Import(importArgs *ImportArgs) (string, error)
	ListStates() (*StatesOutput, error)
	ShowState(resource string) (string, error)
	RemoveState(resource string) (string, error)
	Destroy() (string, error)
}

type importService struct {
	tfExecutor TerraformExecutor
}

func NewImportService(tfWorkspace string, clusterType constants.ClusterType) (ImportService, error) {
	svc := &importService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetImportManifestsDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *importService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *importService) Import(importArgs *ImportArgs) (string, error) {
	args := []string{importArgs.Resource}
	if importArgs.ObjectName != "" {
		args = append(args, fmt.Sprintf("%s,%s", importArgs.ClusterID, importArgs.ObjectName))
	} else {
		args = append(args, importArgs.ClusterID)
	}
	return svc.tfExecutor.RunTerraformImport(args...)
}

func (svc *importService) ListStates() (*StatesOutput, error) {
	output, err := svc.tfExecutor.RunTerraformState("list")
	if err != nil {
		return nil, err
	}
	states := StatesOutput{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		states.States = append(states.States, StateOutput{
			Resource: scanner.Text(),
		})
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return &states, nil
}

func (svc *importService) ShowState(resource string) (string, error) {
	return svc.tfExecutor.RunTerraformState("show", resource)
}

func (svc *importService) RemoveState(resource string) (string, error) {
	return svc.tfExecutor.RunTerraformState("rm", resource)
}

func (svc *importService) Destroy() (string, error) {
	states, err := svc.ListStates()
	if err != nil {
		return "", err
	}
	var errs []error
	var resourceNames []string
	for _, state := range states.States {
		_, err = svc.RemoveState(state.Resource)
		if err != nil {
			errs = append(errs, err)
		} else {
			resourceNames = append(resourceNames, state.Resource)
		}
	}
	if len(errs) > 0 {
		return "", errors.Join(errs...)
	}
	return fmt.Sprintf("Removed resources: %v", resourceNames), nil
}
