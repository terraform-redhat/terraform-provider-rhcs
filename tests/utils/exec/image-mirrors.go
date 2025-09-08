package exec

import (
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec/manifests"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type ImageMirrorArgs struct {
	Cluster *string   `hcl:"cluster"`
	Type    *string   `hcl:"type"`
	Source  *string   `hcl:"source_registry"`
	Mirrors *[]string `hcl:"mirrors"`
}

type ImageMirrorOutput struct {
	ID                   string   `json:"id,omitempty"`
	ClusterID            string   `json:"cluster_id,omitempty"`
	Type                 string   `json:"type,omitempty"`
	Source               string   `json:"source,omitempty"`
	Mirrors              []string `json:"mirrors,omitempty"`
	CreationTimestamp    string   `json:"creation_timestamp,omitempty"`
	LastUpdateTimestamp  string   `json:"last_update_timestamp,omitempty"`
}

type ImageMirrorService interface {
	Init() error
	Plan(args *ImageMirrorArgs) (string, error)
	Apply(args *ImageMirrorArgs) (string, error)
	Output() (*ImageMirrorOutput, error)
	Destroy() (string, error)

	ReadTFVars() (*ImageMirrorArgs, error)
	DeleteTFVars() error
}

type imageMirrorService struct {
	tfExecutor TerraformExecutor
}

func NewImageMirrorService(tfWorkspace string, clusterType constants.ClusterType) (ImageMirrorService, error) {
	svc := &imageMirrorService{
		tfExecutor: NewTerraformExecutor(tfWorkspace, manifests.GetImageMirrorManifestsDir(clusterType)),
	}
	err := svc.Init()
	return svc, err
}

func (svc *imageMirrorService) Init() (err error) {
	_, err = svc.tfExecutor.RunTerraformInit()
	return
}

func (svc *imageMirrorService) Plan(args *ImageMirrorArgs) (string, error) {
	return svc.tfExecutor.RunTerraformPlan(args)
}

func (svc *imageMirrorService) Apply(args *ImageMirrorArgs) (string, error) {
	return svc.tfExecutor.RunTerraformApply(args)
}

func (svc *imageMirrorService) Output() (*ImageMirrorOutput, error) {
	var output ImageMirrorOutput
	err := svc.tfExecutor.RunTerraformOutputIntoObject(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (svc *imageMirrorService) Destroy() (string, error) {
	return svc.tfExecutor.RunTerraformDestroy()
}

func (svc *imageMirrorService) ReadTFVars() (*ImageMirrorArgs, error) {
	args := &ImageMirrorArgs{}
	err := svc.tfExecutor.ReadTerraformVars(args)
	return args, err
}

func (svc *imageMirrorService) DeleteTFVars() error {
	return svc.tfExecutor.DeleteTerraformVars()
}

func NewImageMirrorArgs(cluster, source string, mirrors []string) *ImageMirrorArgs {
	return &ImageMirrorArgs{
		Cluster: helper.StringPointer(cluster),
		Type:    helper.StringPointer("digest"),
		Source:  helper.StringPointer(source),
		Mirrors: &mirrors,
	}
}