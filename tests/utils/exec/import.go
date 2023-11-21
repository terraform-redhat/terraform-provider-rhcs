package exec

import (
	"context"
	"fmt"

	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type ImportArgs struct {
	URL          string `json:"url,omitempty"`
	Token        string `json:"token,omitempty"`
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
	importTF.ManifestDir = con.ImportResourceDir
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
	importArgs.URL = con.GateWayURL
	importService.CreationArgs = importArgs

	args := combineImportStructArgs(importArgs)
	args = append(args, extraArgs...)

	_, err := runTerraformImportWithArgs(importService.Context, importService.ManifestDir, args)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}
	return nil
}
