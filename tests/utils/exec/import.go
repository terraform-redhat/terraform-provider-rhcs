package exec

import (
	"context"
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type ImportArgs struct {
	ResourceKind string `json:"resource_kind,omitempty"`
	ResourceName string `json:"resource_name,omitempty"`
	ClusterID    string `json:"cluster_id,omitempty"`
	ObjectName   string `json:"obj_name,omitempty"`
}

type ImportService struct {
	CreationArgs *ImportArgs
	ManifestDir  string
	Context      context.Context
}

func (importTF *ImportService) InitImport(manifestDirs ...string) error {
	importTF.ManifestDir = constants.ImportResourceDir
	if len(manifestDirs) > 0 {
		importTF.ManifestDir = manifestDirs[0]
	}

	importTF.Context = context.TODO()
	return runTerraformInit(importTF.Context, importTF.ManifestDir)
}

func NewImportService(manifestDir ...string) *ImportService {
	importTF := &ImportService{}
	if err := importTF.InitImport(manifestDir...); err != nil {
		// handle the error as needed
		panic(fmt.Sprintf("Failed to initialize ImportService: %v", err))
	}
	return importTF
}

func combineImportStructArgs(importObj *ImportArgs) []string {
	baseArgs := fmt.Sprintf("%s.%s", importObj.ResourceKind, importObj.ResourceName)

	if importObj.ResourceKind != "rhcs_cluster_rosa_classic" {
		return []string{baseArgs, fmt.Sprintf("%s,%s", importObj.ClusterID, importObj.ObjectName)}
	}

	return []string{baseArgs, importObj.ClusterID}
}

func (importService *ImportService) Import(importArgs *ImportArgs, extraArgs ...string) error {
	importService.CreationArgs = importArgs

	args := combineImportStructArgs(importArgs)
	args = append(args, extraArgs...)

	_, err := runTerraformImport(importService.Context, importService.ManifestDir, args...)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}
	return nil
}

func (importService *ImportService) ShowState(importArgs *ImportArgs) (string, error) {
	args := fmt.Sprintf("%s.%s", importArgs.ResourceKind, importArgs.ResourceName)
	output, err := runTerraformState(importService.ManifestDir, "show", args)
	return output, err
}

func (importService *ImportService) RemoveState(importArgs *ImportArgs) (string, error) {
	args := fmt.Sprintf("%s.%s", importArgs.ResourceKind, importArgs.ResourceName)
	output, err := runTerraformState(importService.ManifestDir, "rm", args)
	return output, err
}

func (importService *ImportService) Destroy(createArgs ...*ImportArgs) (output string, err error) {
	if importService.CreationArgs == nil && len(createArgs) == 0 {
		return "", fmt.Errorf("unset destroy args, set them in the object or pass as parameters")
	}
	return runTerraformDestroy(importService.Context, importService.ManifestDir)
}
