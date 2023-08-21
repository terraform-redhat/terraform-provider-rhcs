package exec

***REMOVED***
	"context"
	"encoding/json"
***REMOVED***
	"os/exec"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***
***REMOVED***

type ClusterCreationArgs struct {
	AccountRolePrefix    string            `json:"account_role_prefix,omitempty"`
	OCMENV               string            `json:"rhcs_environment,omitempty"`
	ClusterName          string            `json:"cluster_name,omitempty"`
	OperatorRolePrefix   string            `json:"operator_role_prefix,omitempty"`
	OpenshiftVersion     string            `json:"openshift_version,omitempty"`
	Token                string            `json:"token,omitempty"`
	URL                  string            `json:"url,omitempty"`
	AWSRegion            string            `json:"aws_region,omitempty"`
	AWSAvailabilityZones []string          `json:"aws_availability_zones,omitempty"`
	Replicas             int               `json:"replicas,omitempty"`
	ChannelGroup         string            `json:"channel_group,omitempty"`
	AWSHttpTokensState   string            `json:"aws_http_tokens_state,omitempty"`
	PrivateLink          string            `json:"private_link,omitempty"`
	Private              string            `json:"private,omitempty"`
	AWSSubnetIDs         []string          `json:"aws_subnet_ids,omitempty"`
	ComputeMachineType   string            `json:"compute_machine_type,omitempty"`
	DefaultMPLabels      map[string]string `json:"default_mp_labels,omitempty"`
	DisableSCPChecks     bool              `json:"disable_scp_checks,omitempty"`
	MultiAZ              bool              `json:"multi_az,omitempty"`
	MachineCIDR          string            `json:"machine_cidr,omitempty"`
	OIDCConfig           string            `json:"oidc_config,omitempty"`
}

// Just a placeholder, not research what to output yet.
type ClusterOutout struct {
	ClusterID string `json:"cluster_id,omitempty"`
}

// *********************** Cluster CMS ***********************************
// func CreateCluster(ctx context.Context,manifestsDir string, args ...string***REMOVED*** (string, error***REMOVED*** {
// 	runTerraformInit(ctx, CON.ClusterDir***REMOVED***

// 	runTerraformApplyWithArgs(ctx, CON.ClusterDir, args***REMOVED***

// 	getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
// 	getClusterIdCmd.Dir = CON.ClusterDir
// 	output, err := getClusterIdCmd.Output(***REMOVED***
// 	if err != nil {
// 		return "", err
// 	}

// 	splitOutput := strings.Split(string(output***REMOVED***, "\""***REMOVED***
// 	if len(splitOutput***REMOVED*** <= 1 {
// 		return "", fmt.Errorf("got no cluster id from the output"***REMOVED***
// 	}

// 	return splitOutput[1], nil
// }

type ClusterService struct {
	CreationArgs *ClusterCreationArgs
	ManifestDir  string
	Context      context.Context
}

func (creator *ClusterService***REMOVED*** Init(manifestDir string***REMOVED*** error {
	creator.ManifestDir = CON.GrantClusterManifestDir(manifestDir***REMOVED***
	ctx := context.TODO(***REMOVED***
	creator.Context = ctx
	err := runTerraformInit(ctx, creator.ManifestDir***REMOVED***
	if err != nil {
		return err
	}
	return nil

}

func (creator *ClusterService***REMOVED*** Create(createArgs *ClusterCreationArgs, extraArgs ...string***REMOVED*** error {
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	_, err := runTerraformApplyWithArgs(creator.Context, creator.ManifestDir, args***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

func (creator *ClusterService***REMOVED*** Output(***REMOVED*** (string, error***REMOVED*** {
	out, err := runTerraformOutput(creator.Context, creator.ManifestDir***REMOVED***
	if err != nil {
		return "", err
	}
	clusterObj := out["cluster_id"]
	clusterID := h.DigString(clusterObj, "value"***REMOVED***
	return clusterID, nil
}

func (creator *ClusterService***REMOVED*** Destroy(createArgs *ClusterCreationArgs, extraArgs ...string***REMOVED*** error {
	args := combineStructArgs(createArgs, extraArgs...***REMOVED***
	err := runTerraformDestroyWithArgs(creator.Context, creator.ManifestDir, args***REMOVED***
	return err
}

func NewClusterService(manifestDir string***REMOVED*** *ClusterService {
	sc := &ClusterService{}
	sc.Init(manifestDir***REMOVED***
	return sc
}

//******************************************************

func CreateTFCluster(ctx context.Context, manifestsDir string,
	varArgs map[string]interface{}, abArgs ...string***REMOVED*** (string, error***REMOVED*** {
	targetDir := CON.GrantClusterManifestDir(manifestsDir***REMOVED***
	err := runTerraformInit(ctx, targetDir***REMOVED***
	if err != nil {
		return "", err
	}

	args := combineArgs(varArgs, abArgs...***REMOVED***

	_, err = runTerraformApplyWithArgs(ctx, targetDir, args***REMOVED***
	if err != nil {
		return "", err
	}

	getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	getClusterIdCmd.Dir = targetDir
	output, err := getClusterIdCmd.Output(***REMOVED***
	if err != nil {
		return "", err
	}

	splitOutput := strings.Split(string(output***REMOVED***, "\""***REMOVED***
	if len(splitOutput***REMOVED*** <= 1 {
		return "", fmt.Errorf("got no cluster id from the output"***REMOVED***
	}

	return splitOutput[1], nil
}

func DestroyTFCluster(ctx context.Context, manifestDir string,
	varArgs map[string]interface{}, abArgs ...string***REMOVED*** error {
	targetDir := CON.GrantClusterManifestDir(manifestDir***REMOVED***
	err := runTerraformInit(ctx, targetDir***REMOVED***
	if err != nil {
		return err
	}

	args := combineArgs(varArgs, abArgs...***REMOVED***
	err = runTerraformDestroyWithArgs(ctx, targetDir, args***REMOVED***
	// if err != nil {
	// 	return err
	// }

	// getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	// getClusterIdCmd.Dir = targetDir
	// _, err = getClusterIdCmd.Output(***REMOVED***

	return err
}

func CreateMyTFCluster(clusterArgs *ClusterCreationArgs, manifestsDir string, arg ...string***REMOVED*** (string, error***REMOVED*** {
	parambytes, _ := json.Marshal(clusterArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	return CreateTFCluster(context.TODO(***REMOVED***, manifestsDir, args, arg...***REMOVED***
}

func DestroyMyTFCluster(clusterArgs *ClusterCreationArgs, manifestsDir string, arg ...string***REMOVED*** error {
	parambytes, _ := json.Marshal(clusterArgs***REMOVED***
	args := map[string]interface{}{}
	json.Unmarshal(parambytes, &args***REMOVED***
	return DestroyTFCluster(context.TODO(***REMOVED***, manifestsDir, args, arg...***REMOVED***
}
