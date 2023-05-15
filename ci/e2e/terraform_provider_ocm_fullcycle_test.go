***REMOVED***

***REMOVED***
	"context"
	"flag"
***REMOVED***
	"io"
***REMOVED***
	"os"
	"os/exec"
***REMOVED***
	"path/filepath"
	"strings"
	"testing"
	"time"

***REMOVED***
***REMOVED***
	client "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/segmentio/ksuid"
	"github.com/terraform-redhat/terraform-provider-ocm/ci/helper"
	"k8s.io/apimachinery/pkg/util/rand"
***REMOVED***

const (
	accountRolesFilesDir         = "account_roles_files"
	terraformProviderOCMFilesDir = "terraform_provider_ocm_files"
***REMOVED***

var (
	randSuffix          string
	providerTempDir     string
	accountRolesTempDir string
	operatorRolePrefix  string
	accountRolePrefix   string
	ocmEnvironment      string
	openshiftVersion    string
	tokenFilter         string
	gatewayFilter       string
	clusterName         string
	ctx                 context.Context
	TestID              = ksuid.New(***REMOVED***
	connection          *client.Connection
	logger              logging.Logger
***REMOVED***

var URLAliases = map[string]string{
	"https://api.openshift.com":             "production",
	"https://api.stage.openshift.com":       "staging",
	"https://api.integration.openshift.com": "integration",
	"http://localhost:8000":                 "local",
	"http://localhost:9000":                 "local",
}

var args struct {
	tokenURL         string
	gatewayURL       string
	offlineToken     string
	clientID         string
	clientSecret     string
	openshiftVersion string
}

func init(***REMOVED*** {
	flag.StringVar(
		&args.tokenURL,
		"token-url",
		"https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token",
		"Token URL.",
	***REMOVED***
	flag.StringVar(
		&args.gatewayURL,
		"gateway-url",
		"http://localhost:8001",
		"Gateway URL.",
	***REMOVED***
	flag.StringVar(
		&args.clientID,
		"client-id",
		"cloud-services",
		"OpenID client identifier.",
	***REMOVED***
	flag.StringVar(
		&args.clientSecret,
		"client-secret",
		"",
		"OpenID client secret.",
	***REMOVED***
	flag.StringVar(
		&args.offlineToken,
		"offline-token",
		"",
		"Offline token for authentication.",
	***REMOVED***
	flag.StringVar(
		&args.openshiftVersion,
		"openshift-version",
		"4.12",
		"Openshift version.",
	***REMOVED***
}

func TestE2E(t *testing.T***REMOVED*** {
	// Create the context:
	ctx = context.Background(***REMOVED***
	// logger is the global logger used by the tests.
	logger = helper.GetLogger(***REMOVED***

	// Parse the command line flags:
	flag.Parse(***REMOVED***

	// Run the tests:
	logger.Info(ctx, "Test ID: %s", TestID***REMOVED***
	validateArgs(***REMOVED***
	RegisterFailHandler(Fail***REMOVED***
	RunSpecs(t, "Full Cycle Test Suite"***REMOVED***

}

func validateArgs(***REMOVED*** {
	helper.CheckEmpty(args.gatewayURL, "gateway-url"***REMOVED***
	helper.CheckEmpty(args.clientID, "client-id"***REMOVED***
	helper.CheckEmpty(args.openshiftVersion, "openshift-version"***REMOVED***
	helper.CheckEmpty(args.offlineToken, "offline-token"***REMOVED***
}

var _ = BeforeEach(func(***REMOVED*** {
	var err error

	// Replace the global logger with one specific for this test that writes to the Ginkgo
	// streams, that way the log messages will only be displayed if the test fails:
	logger, err = logging.NewStdLoggerBuilder(***REMOVED***.
		Streams(GinkgoWriter, GinkgoWriter***REMOVED***.
		Build(***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
}***REMOVED***

var _ = BeforeSuite(func(***REMOVED*** {
	connection = helper.CreateConnectionWithToken(
		args.offlineToken,
		args.tokenURL,
		args.gatewayURL,
		args.clientID,
		args.clientSecret***REMOVED***
	helper.WaitForBackendToBeReady(ctx, connection***REMOVED***
}***REMOVED***

func runTerraformInit(ctx context.Context, dir string***REMOVED*** {
	logger.Info(ctx, "Running terraform init against the dir %s", dir***REMOVED***
	terraformInitCmd := exec.Command("terraform", "init"***REMOVED***
	terraformInitCmd.Dir = dir
	err := terraformInitCmd.Run(***REMOVED***
	helper.CheckError(err***REMOVED***
}

func runTerraformApplyWithArgs(ctx context.Context, dir string, terraformArgs []string***REMOVED*** {
	applyArgs := append([]string{"apply"}, terraformArgs...***REMOVED***
	logger.Info(ctx, "Running terraform apply against the dir: %s", dir***REMOVED***
	terraformApply := exec.Command("terraform", applyArgs...***REMOVED***
	terraformApply.Dir = dir
	terraformApply.Stdout = os.Stdout
	terraformApply.Stderr = os.Stderr
	err := terraformApply.Run(***REMOVED***
	helper.CheckError(err***REMOVED***
}
func runTerraformDestroyWithArgs(ctx context.Context, dir string, terraformArgs []string***REMOVED*** {
	destroyArgs := append([]string{"destroy"}, terraformArgs...***REMOVED***
	logger.Info(ctx, "Running terraform destroy against the dir: %s", dir***REMOVED***
	terraformApply := exec.Command("terraform", destroyArgs...***REMOVED***
	terraformApply.Dir = dir
	terraformApply.Stdout = os.Stdout
	terraformApply.Stderr = os.Stderr
	err := terraformApply.Run(***REMOVED***
	helper.CheckError(err***REMOVED***
}

func createAccountRolesUsingTerraformAWSRosaStsModule(ctx context.Context***REMOVED*** {
	runTerraformInit(ctx, accountRolesTempDir***REMOVED***

	runTerraformApplyWithArgs(ctx, accountRolesTempDir, []string{
		"-var", accountRolePrefix,
		"-var", ocmEnvironment,
		"-var", openshiftVersion,
		"-var", tokenFilter,
		"-var", gatewayFilter,
		"-auto-approve"}***REMOVED***
}

func createClusterUsingTerraformProviderOCM(ctx context.Context***REMOVED*** string {
	runTerraformInit(ctx, providerTempDir***REMOVED***

	runTerraformApplyWithArgs(ctx, providerTempDir, []string{
		"-var", tokenFilter,
		"-var", operatorRolePrefix,
		"-var", accountRolePrefix,
		"-var", gatewayFilter,
		"-var", clusterName,
		"-auto-approve"}***REMOVED***

	getClusterIdCmd := exec.Command("terraform", "output", "-json", "cluster_id"***REMOVED***
	getClusterIdCmd.Dir = providerTempDir
	output, err := getClusterIdCmd.Output(***REMOVED***
	helper.CheckError(err***REMOVED***

	splitOutput := strings.Split(string(output***REMOVED***, "\""***REMOVED***
	Expect(len(splitOutput***REMOVED******REMOVED***.To(BeNumerically(">", 1***REMOVED******REMOVED***

	return splitOutput[1]
}

func defineVariablesValues(***REMOVED*** {
	randSuffix = rand.String(4***REMOVED***
	logger.Info(ctx, "The random suffix that was chosen is %s", randSuffix***REMOVED***

	providerTempDir = fmt.Sprintf("%s_%s", terraformProviderOCMFilesDir, randSuffix***REMOVED***
	logger.Info(ctx, "The temp directory that was chosen is %s", providerTempDir***REMOVED***

	accountRolesTempDir = fmt.Sprintf("%s_%s", accountRolesFilesDir, randSuffix***REMOVED***
	logger.Info(ctx, "The temp directory that was chosen is %s", accountRolesTempDir***REMOVED***

	clusterName = fmt.Sprintf("cluster_name=terr-ocm-%s", randSuffix***REMOVED***
	logger.Info(ctx, "The cluster name that was chosen is %s", clusterName***REMOVED***

	operatorRolePrefix = fmt.Sprintf("operator_role_prefix=terr-operator-%s", randSuffix***REMOVED***
	logger.Info(ctx, "The operator IAM role prefix that was chose is %s", operatorRolePrefix***REMOVED***

	accountRolePrefix = fmt.Sprintf("account_role_prefix=terr-account-%s", randSuffix***REMOVED***
	logger.Info(ctx, "The account IAM role prefix that was chose is %s", accountRolePrefix***REMOVED***

	ocmEnvironment = fmt.Sprintf("ocm_environment=%s", URLAliases[args.gatewayURL]***REMOVED***
	logger.Info(ctx, "The ocm environment that was chose is %s", ocmEnvironment***REMOVED***

	openshiftVersion = fmt.Sprintf("openshift_version=%s", args.openshiftVersion***REMOVED***
	logger.Info(ctx, "The cluster version that was chose is %s", openshiftVersion***REMOVED***

	tokenFilter = fmt.Sprintf("token=%s", args.offlineToken***REMOVED***

	gatewayFilter = fmt.Sprintf("url=%s", args.gatewayURL***REMOVED***
	logger.Info(ctx, "The gateway url filter is %s", gatewayFilter***REMOVED***

}

var _ = FDescribe("Terraform provider OCM test", Ordered, func(***REMOVED*** {
	var terraformProviderOCMClusterID string

	BeforeAll(func(***REMOVED*** {
		defineVariablesValues(***REMOVED***

		//create temp dirs
		err := Unpack(providerTempDir, terraformProviderOCMFilesDir***REMOVED***
		helper.CheckError(err***REMOVED***

		err = Unpack(accountRolesTempDir, accountRolesFilesDir***REMOVED***
		helper.CheckError(err***REMOVED***

		// prepareDirectory
		createAccountRolesUsingTerraformAWSRosaStsModule(ctx***REMOVED***
		time.Sleep(5 * time.Second***REMOVED***
		terraformProviderOCMClusterID = createClusterUsingTerraformProviderOCM(ctx***REMOVED***
	}***REMOVED***

	Context("Cluster creation", func(***REMOVED*** {
		It("creates a cluster using terraform-provider-ocm", func(***REMOVED*** {
			resp, err := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(terraformProviderOCMClusterID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
			helper.CheckResponse(resp, err, http.StatusOK***REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	AfterAll(func(***REMOVED*** {
		// destroy cluster
		runTerraformDestroyWithArgs(ctx, providerTempDir, []string{
			"-var", tokenFilter,
			"-var", operatorRolePrefix,
			"-var", accountRolePrefix,
			"-var", gatewayFilter,
			"-var", clusterName,
			"-auto-approve"}***REMOVED***

		// destroy account roles
		runTerraformDestroyWithArgs(ctx, accountRolesTempDir, []string{
			"-var", accountRolePrefix,
			"-var", ocmEnvironment,
			"-var", openshiftVersion,
			"-var", gatewayFilter,
			"-auto-approve"}***REMOVED***

		// remove temporary directories
		os.RemoveAll(providerTempDir***REMOVED***
		os.RemoveAll(accountRolesTempDir***REMOVED***
	}***REMOVED***
}***REMOVED***

// Unpack unpacks the terraform tf files from this package into a target directory.
func Unpack(targetDir string, sourceDir string***REMOVED*** (err error***REMOVED*** {
	dir := ""
	Assets := http.Dir(dir***REMOVED***

	file, err := Assets.Open(sourceDir***REMOVED***
	if err != nil {
		return err
	}
	defer file.Close(***REMOVED***

	info, err := file.Stat(***REMOVED***
	if err != nil {
		return err
	}

	if info.IsDir(***REMOVED*** {
		os.Mkdir(targetDir, 0777***REMOVED***
		children, err := file.Readdir(0***REMOVED***
		if err != nil {
			return err
***REMOVED***
		file.Close(***REMOVED***

		for _, childInfo := range children {
			name := childInfo.Name(***REMOVED***
			err = Unpack(filepath.Join(targetDir, name***REMOVED***, path.Join(sourceDir, name***REMOVED******REMOVED***
			if err != nil {
				return err
	***REMOVED***
***REMOVED***
		return nil
	}

	out, err := os.OpenFile(targetDir, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666***REMOVED***
	if err != nil {
		return err
	}
	defer out.Close(***REMOVED***

	_, err = io.Copy(out, file***REMOVED***
	return err
}
