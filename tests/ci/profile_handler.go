package ci

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"

	client "github.com/openshift-online/ocm-sdk-go"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

// Profile Provides profile struct for cluster creation be matrix
type Profile struct {
	Name                  string `ini:"name,omitempty" json:"name,omitempty"`
	ClusterName           string `ini:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	ClusterType           string `ini:"cluster_type,omitempty" json:"cluster_type,omitempty"`
	ProductID             string `ini:"product_id,omitempty" json:"product_id,omitempty"`
	MajorVersion          string `ini:"major_version,omitempty" json:"major_version,omitempty"`
	Version               string `ini:"version,omitempty" json:"version,omitempty"`                 //Specific OCP version to be specified
	VersionPattern        string `ini:"version_pattern,omitempty" json:"version_pattern,omitempty"` //Version supports indicated version started with openshift-v or major-1 (y-1) or minor-1 (z-1)
	ChannelGroup          string `ini:"channel_group,omitempty" json:"channel_group,omitempty"`
	CloudProvider         string `ini:"cloud_provider,omitempty" json:"cloud_provider,omitempty"`
	Region                string `ini:"region,omitempty" json:"region,omitempty"`
	InstanceType          string `ini:"instance_type,omitempty" json:"instance_type,omitempty"`
	Zones                 string `ini:"zones,omitempty" json:"zones,omitempty"`           // zones should be like a,b,c,d
	StorageLB             bool   `ini:"storage_lb,omitempty" json:"storage_lb,omitempty"` // the unit is GIB, don't support unit set
	Tagging               bool   `ini:"tagging,omitempty" json:"tagging,omitempty"`
	Labeling              bool   `ini:"labeling,omitempty" json:"labeling,omitempty"`
	Etcd                  bool   `ini:"etcd_encryption,omitempty" json:"etcd_encryption,omitempty"`
	FIPS                  bool   `ini:"fips,omitempty" json:"fips,omitempty"`
	CCS                   bool   `ini:"ccs,omitempty" json:"ccs,omitempty"`
	STS                   bool   `ini:"sts,omitempty" json:"sts,omitempty"`
	Autoscale             bool   `ini:"autoscaling_enabled,omitempty" json:"autoscaling_enabled,omitempty"`
	MultiAZ               bool   `ini:"multi_az,omitempty" json:"multi_az,omitempty"`
	BYOVPC                bool   `ini:"byovpc,omitempty" json:"byovpc,omitempty"`
	PrivateLink           bool   `ini:"private_link,omitempty" json:"private_link,omitempty"`
	Private               bool   `ini:"private,omitempty" json:"private,omitempty"`
	BYOK                  bool   `ini:"byok,omitempty" json:"byok,omitempty"`
	KMSKey                bool   `ini:"kms_key_arn,omitempty" json:"kms_key_arn,omitempty"`
	NetWorkingSet         bool   `ini:"networking_set,omitempty" json:"networking_set,omitempty"`
	Proxy                 bool   `ini:"proxy,omitempty" json:"proxy,omitempty"`
	OIDCConfig            string `ini:"oidc_config,omitempty" json:"oidc_config,omitempty"`
	ProvisionShard        string `ini:"provisionShard,omitempty" json:"provisionShard,omitempty"`
	Ec2MetadataHttpTokens string `ini:"ec2_metadata_http_tokens,omitempty" json:"ec2_metadata_http_tokens,omitempty"`
	ComputeMachineType    string `ini:"compute_machine_type,omitempty" json:"compute_machine_type,omitempty"`
	AuditLogForward       bool   `ini:"auditlog_forward,omitempty" json:"auditlog_forward,omitempty"`
	AdminEnabled          bool   `ini:"admin_enabled,omitempty" json:"admin_enabled,omitempty"`
	ManagedPolicies       bool   `ini:"managed_policies,omitempty" json:"managed_policies,omitempty"`
	WorkerDiskSize        int    `ini:"worker_disk_size,omitempty" json:"worker_disk_size,omitempty"`
	AdditionalSGNumber    int    `ini:"additional_sg_number,omitempty" json:"additional_sg_number,omitempty"`
	UnifiedAccRolesPath   string `ini:"unified_acc_role_path,omitempty" json:"unified_acc_role_path,omitempty"`
	SharedVpc             bool   `ini:"shared_vpc,omitempty" json:"shared_vpc,omitempty"`
}

func GetTerraformExecHelperForProfile(profile *Profile) (*exec.TerraformExecHelper, error) {
	return exec.NewTerraformExecHelperWithWorkspaceName(profile.GetClusterType(), profile.Name)
}

func GenerateNewTerraformWorkspaceFromProfile(profile *Profile) string {
	return helper.GenerateRandomName(profile.Name, 7)
}

// PrepareVersion supports below types
// version with a openshift version like 4.13.12
// version with latest
// verion with x-1, it means the version will choose one with x-1 version which can be used for x stream upgrade
// version with y-1, it means the version will choose one with y-1 version which can be used for y stream upgrade
func PrepareVersion(connection *client.Connection, versionTag string, channelGroup string, profile *Profile) string {
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\-*[\s\S]*$`)
	// Check that the version is matching openshift version regexp
	if versionRegex.MatchString(versionTag) {
		return versionTag
	}
	var vResult string
	switch versionTag {
	case "", "latest":
		versions := cms.EnabledVersions(connection, channelGroup, profile.MajorVersion, true)
		versions = cms.SortVersions(versions)
		vResult = versions[len(versions)-1].RawID
	case "y-1":
		versions, _ := cms.GetVersionsWithUpgrades(connection, channelGroup, constants.Y, true, false, 1)
		vResult = versions[len(versions)-1].RawID
	case "z-1":
		versions, _ := cms.GetVersionsWithUpgrades(connection, channelGroup, constants.Z, true, false, 1)
		vResult = versions[len(versions)-1].RawID
	case "eol":
		vResult = ""
	}
	Logger.Infof("Cluster OCP latest version is set to %s", vResult)
	return vResult
}

func GenerateClusterCreationArgsByProfile(token string, profile *Profile) (clusterArgs *exec.ClusterCreationArgs, err error) {
	tfExecHelper, err := GetTerraformExecHelperForProfile(profile)
	Expect(err).ToNot(HaveOccurred())

	profile.Version = PrepareVersion(RHCSConnection, profile.VersionPattern, profile.ChannelGroup, profile)

	clusterArgs = &exec.ClusterCreationArgs{
		OpenshiftVersion: profile.Version,
	}

	// Init cluster's args by profile's attributes

	// For Shared VPC
	var cluster_aws_account string
	var installer_role_arn string
	var ingress_role_arn string

	if profile.FIPS {
		clusterArgs.Fips = profile.FIPS
	}

	if profile.Etcd {
		clusterArgs.Etcd = profile.Etcd
	}

	if profile.MultiAZ {
		clusterArgs.MultiAZ = profile.MultiAZ
	}

	if profile.NetWorkingSet {
		clusterArgs.MachineCIDR = constants.DefaultVPCCIDR
	}

	if profile.Autoscale {
		clusterArgs.Autoscale = profile.Autoscale
	}

	if profile.ComputeMachineType != "" {
		clusterArgs.ComputeMachineType = profile.ComputeMachineType
	}

	if profile.ChannelGroup != "" {
		clusterArgs.ChannelGroup = profile.ChannelGroup
	}

	if profile.Ec2MetadataHttpTokens != "" {
		clusterArgs.Ec2MetadataHttpTokens = profile.Ec2MetadataHttpTokens
	}

	if profile.Region == "" {
		profile.Region = constants.DefaultAWSRegion
	}

	if profile.Labeling {
		clusterArgs.DefaultMPLabels = constants.DefaultMPLabels
	}

	if profile.Tagging {
		clusterArgs.Tagging = constants.Tags
	}

	if profile.AdminEnabled {
		userName := constants.ClusterAdminUser
		password := helper.GenerateRandomStringWithSymbols(14)
		adminPasswdMap := map[string]string{"username": userName, "password": password}
		clusterArgs.AdminCredentials = adminPasswdMap
		pass := []byte(password)
		err = os.WriteFile(path.Join(constants.GetRHCSOutputDir(), constants.ClusterAdminUser), pass, 0644)
		if err != nil {
			return
		}
		Logger.Infof("Admin password is written to the output directory")

	}

	if profile.AuditLogForward {
		// ToDo
	}

	if profile.ClusterName != "" {
		clusterArgs.ClusterName = profile.ClusterName
	} else {
		// Generate random chars later cluster name with profile name
		clusterArgs.ClusterName = helper.GenerateClusterName(profile.Name)
	}

	if profile.Region != "" {
		clusterArgs.AWSRegion = profile.Region
	} else {
		clusterArgs.AWSRegion = constants.DefaultAWSRegion
	}

	if profile.STS {
		majorVersion := helper.GetMajorVersion(profile.Version)
		shared_vpc_role_arn := ""
		if profile.SharedVpc {
			// FIXME:
			//	To create Shared-VPC compatible policies, we need to pass a role arn to create_account_roles module.
			//  But we got an chicken-egg prolems here:
			//		* The Shared-VPC compatible policie requries installer role
			//		* The install role (account roles) require Shared-VPC ARN.
			//  Use hardcode as a temporary solution.
			shared_vpc_role_arn = fmt.Sprintf("arn:aws:iam::641733028092:role/%s-shared-vpc-role", clusterArgs.ClusterName)
		}
		accountRolesOutput, err := tfExecHelper.PrepareAccountRoles(token, clusterArgs.ClusterName, profile.UnifiedAccRolesPath, clusterArgs.AWSRegion, majorVersion, profile.ChannelGroup, shared_vpc_role_arn)
		Expect(err).ToNot(HaveOccurred())
		clusterArgs.AccountRolePrefix = accountRolesOutput.AccountRolePrefix
		clusterArgs.UnifiedAccRolesPath = profile.UnifiedAccRolesPath
		Logger.Infof("Created account roles with prefix %s", accountRolesOutput.AccountRolePrefix)

		Logger.Infof("Sleep for 10 sec to let aws account role async creation finished")
		time.Sleep(10 * time.Second)

		oidcOutput, err := tfExecHelper.PrepareOIDCProviderAndOperatorRoles(token, profile.OIDCConfig, clusterArgs.ClusterName, accountRolesOutput.AccountRolePrefix, profile.UnifiedAccRolesPath, clusterArgs.AWSRegion)
		Expect(err).ToNot(HaveOccurred())
		clusterArgs.OIDCConfigID = oidcOutput.OIDCConfigID
		clusterArgs.OperatorRolePrefix = oidcOutput.OperatorRolePrefix

		cluster_aws_account = accountRolesOutput.AWSAccountId
		installer_role_arn = accountRolesOutput.InstallerRoleArn
		ingress_role_arn = oidcOutput.IngressOperatorRoleArn
	}

	if profile.BYOVPC {
		var zones []string
		var vpcOutput *exec.VPCOutput
		var sgIDs []string

		// Supports ENV set passed to make cluster provision more flexy in prow
		// Export the subnetIDs via env variable if you have existing ones export SubnetIDs=<subnet1>,<subnet2>,<subnet3>
		// Export the availability zones via env variable export AvailabilitiZones=<az1>,<az2>,<az3>
		if os.Getenv("SubnetIDs") != "" && os.Getenv("AvailabilitiZones") != "" {
			subnetIDs := strings.Split(os.Getenv("SubnetIDs"), ",")
			azs := strings.Split(os.Getenv("AvailabilitiZones"), ",")
			clusterArgs.AWSAvailabilityZones = azs
			clusterArgs.AWSSubnetIDs = subnetIDs
		} else {
			if profile.Zones != "" {
				zones = strings.Split(profile.Zones, ",")
			}

			shared_vpc_aws_shared_credentials_file := ""

			if profile.SharedVpc {
				if constants.SharedVpcAWSSharedCredentialsFileENV == "" {
					panic(fmt.Errorf("SHARED_VPC_AWS_SHARED_CREDENTIALS_FILE env is not set or empty, it's requried by Shared-VPC cluster"))
				}

				shared_vpc_aws_shared_credentials_file = constants.SharedVpcAWSSharedCredentialsFileENV
			}
			vpcOutput, err = tfExecHelper.PrepareVPC(profile.Region, profile.MultiAZ, zones, clusterArgs.ClusterName, shared_vpc_aws_shared_credentials_file)
			if err != nil {
				return
			}

			if vpcOutput.ClusterPrivateSubnets == nil {
				err = fmt.Errorf("error when creating the vpc, check the previous log. The created resources had been destroyed")
				return
			}
			if profile.Private {
				clusterArgs.Private = profile.Private
				clusterArgs.PrivateLink = profile.PrivateLink
				if profile.IsPrivateLink() {
					clusterArgs.AWSSubnetIDs = vpcOutput.ClusterPrivateSubnets
				}
			} else {
				clusterArgs.AWSSubnetIDs = append(vpcOutput.ClusterPrivateSubnets, vpcOutput.ClusterPublicSubnets...)
			}

			if profile.SharedVpc {
				// Base domain
				var base_dns_domain string
				base_dns_domain, err = tfExecHelper.PrepareRoute53()
				if err != nil {
					return
				}

				// Resources for Shared-VPC
				var sharedVpcPolicyAndHostedZoneOutput *exec.SharedVpcPolicyAndHostedZoneOutput
				sharedVpcPolicyAndHostedZoneOutput, err = tfExecHelper.PrepareSharedVpcPolicyAndHostedZone(
					profile.Region,
					constants.SharedVpcAWSSharedCredentialsFileENV,
					clusterArgs.ClusterName,
					base_dns_domain,
					ingress_role_arn,
					installer_role_arn,
					cluster_aws_account,
					vpcOutput.VPCID,
					clusterArgs.AWSSubnetIDs)
				if err != nil {
					return
				}

				clusterArgs.BaseDnsDomain = base_dns_domain
				private_hosted_zone := exec.PrivateHostedZone{
					ID:      sharedVpcPolicyAndHostedZoneOutput.HostedZoneId,
					RoleArn: sharedVpcPolicyAndHostedZoneOutput.SharedRole,
				}
				clusterArgs.PrivateHostedZone = &private_hosted_zone
				/*
					The AZ us-east-1a for VPC-account might not have the same location as us-east-1a for Cluster-account.
					For AZs which will be used in cluster configuration, the values should be the ones in Cluster-account.
				*/
				clusterArgs.AWSAvailabilityZones = sharedVpcPolicyAndHostedZoneOutput.AZs
			} else {
				clusterArgs.AWSAvailabilityZones = vpcOutput.AZs
			}

			clusterArgs.MachineCIDR = vpcOutput.VPCCIDR
			if profile.AdditionalSGNumber != 0 {
				// Prepare profile.AdditionalSGNumber+5 security groups for negative testing
				sgIDs, err = tfExecHelper.PrepareAdditionalSecurityGroups(profile.Region, vpcOutput.VPCID, profile.AdditionalSGNumber+5)
				if err != nil {
					return
				}
				clusterArgs.AdditionalComputeSecurityGroups = sgIDs[0:profile.AdditionalSGNumber]
				clusterArgs.AdditionalInfraSecurityGroups = sgIDs[0:profile.AdditionalSGNumber]
				clusterArgs.AdditionalControlPlaneSecurityGroups = sgIDs[0:profile.AdditionalSGNumber]
			}

			// in case Proxy is enabled
			if profile.Proxy {
				var proxyOutput *exec.ProxyOutput
				proxyOutput, err = tfExecHelper.PrepareProxy(profile.Region, vpcOutput.VPCID, vpcOutput.ClusterPublicSubnets[0])
				if err != nil {
					return
				}
				proxy := exec.Proxy{
					AdditionalTrustBundle: proxyOutput.AdditionalTrustBundle,
					HTTPSProxy:            proxyOutput.HttpsProxy,
					HTTPProxy:             proxyOutput.HttpProxy,
					NoProxy:               proxyOutput.NoProxy,
				}
				clusterArgs.Proxy = &proxy
			}
		}
	}

	if profile.KMSKey {
		var kmskey string
		kmskey, err = tfExecHelper.PrepareKMSKey(profile.Region, clusterArgs.ClusterName, clusterArgs.AccountRolePrefix, profile.UnifiedAccRolesPath)
		if err != nil {
			return
		}
		clusterArgs.KmsKeyARN = kmskey

	}

	if profile.WorkerDiskSize != 0 {
		clusterArgs.WorkerDiskSize = profile.WorkerDiskSize
	}
	clusterArgs.UnifiedAccRolesPath = profile.UnifiedAccRolesPath
	clusterArgs.CustomProperties = constants.CustomProperties

	return clusterArgs, err
}

func LoadProfileYamlFile(profileName string) *Profile {
	p := helper.GetProfile(profileName, GetYAMLProfilesDir())
	Logger.Infof("Loaded cluster profile configuration from profile %s : %v", profileName, p.Cluster)
	profile := Profile{
		Name: profileName,
	}
	err := helper.MapStructure(p.Cluster, &profile)
	Expect(err).ToNot(HaveOccurred())
	return &profile
}

func LoadProfileYamlFileByENV() *Profile {
	profileEnv := os.Getenv(constants.RhcsClusterProfileENV)
	if profileEnv == "" {
		panic(fmt.Errorf("ENV Variable CLUSTER_PROFILE is empty, please make sure you set the env value"))
	}
	profile := LoadProfileYamlFile(profileEnv)

	// Supporting global env setting to overrite profile settings
	if os.Getenv("CHANNEL_GROUP") != "" {
		Logger.Infof("Got global env settings for CHANNEL_GROUP, overwritten the profile setting with value %s", os.Getenv("CHANNEL_GROUP"))
		profile.ChannelGroup = os.Getenv("CHANNEL_GROUP")
	}
	if os.Getenv("VERSION") != "" {
		Logger.Infof("Got global env settings for VERSION, overwritten the profile setting with value %s", os.Getenv("VERSION"))
		profile.Version = os.Getenv("VERSION")
	}
	if os.Getenv("REGION") != "" {
		Logger.Infof("Got global env settings for REGION, overwritten the profile setting with value %s", os.Getenv("REGION"))
		profile.Region = os.Getenv("REGION")
	}
	return profile
}

func CreateRHCSClusterByProfile(token string, profile *Profile) (string, error) {
	creationArgs, err := GenerateClusterCreationArgsByProfile(token, profile)
	if err != nil {
		defer DestroyRHCSClusterByProfile(token, profile)
	}
	Expect(err).ToNot(HaveOccurred())
	tfExecHelper, err := GetTerraformExecHelperForProfile(profile)
	Expect(err).ToNot(HaveOccurred())
	clusterService, err := tfExecHelper.GetClusterService()
	if err != nil {
		defer DestroyRHCSClusterByProfile(token, profile)
	}
	Expect(err).ToNot(HaveOccurred())
	_, err = clusterService.Apply(creationArgs)
	if err != nil {
		clusterService.Destroy(false)
		return "", err
	}
	clusterOutput, err := clusterService.Output()
	clusterID := clusterOutput.ClusterID

	return clusterID, err
}

func DestroyRHCSClusterByProfile(token string, profile *Profile) error {
	tfExecHelper, err := GetTerraformExecHelperForProfile(profile)
	Expect(err).ToNot(HaveOccurred())

	// Destroy cluster
	clusterService, err := tfExecHelper.GetClusterService()
	Expect(err).ToNot(HaveOccurred())

	// clusterArgs := &exec.ClusterCreationArgs{
	// 	AWSRegion:          profile.Region,
	// 	AccountRolePrefix:  "",
	// 	OperatorRolePrefix: "",
	// }
	// if profile.AdditionalSGNumber != 0 {
	// 	output, err := clusterService.Output()
	// 	Expect(err).ToNot(HaveOccurred())
	// 	clusterArgs.AdditionalComputeSecurityGroups = output.AdditionalComputeSecurityGroups
	// 	clusterArgs.AdditionalInfraSecurityGroups = output.AdditionalInfraSecurityGroups
	// 	clusterArgs.AdditionalControlPlaneSecurityGroups = output.AdditionalControlPlaneSecurityGroups
	// }
	_, err = clusterService.Destroy(true)
	if err != nil {
		return err
	}

	// Destroy VPC
	if profile.BYOVPC {
		if profile.Proxy {
			proxyService, err := tfExecHelper.GetProxyService()
			Expect(err).ToNot(HaveOccurred())
			_, err = proxyService.Destroy(true)
			if err != nil {
				return err
			}
		}
		if profile.AdditionalSGNumber != 0 {
			sgService, err := tfExecHelper.GetSecurityGroupService()
			Expect(err).ToNot(HaveOccurred())
			_, err = sgService.Destroy(true)
			if err != nil {
				return err
			}
		}

		if profile.SharedVpc {

			if constants.SharedVpcAWSSharedCredentialsFileENV == "" {
				panic(fmt.Errorf("SHARED_VPC_AWS_SHARED_CREDENTIALS_FILE env is not set or empty, it's requried by Shared-VPC cluster"))
			}

			sharedVpcPolicyAndHostedZoneService, err := tfExecHelper.GetSharedVpcPolicyAndHostedZoneService()
			if err != nil {
				return err
			}
			_, err = sharedVpcPolicyAndHostedZoneService.Destroy(true)
			if err != nil {
				return err
			}

			// DNS domain
			dnsDomainService, err := tfExecHelper.GetDnsDomainService()
			if err != nil {
				return err
			}
			_, err = dnsDomainService.Destroy(true)
			if err != nil {
				return err
			}
		}

		vpcService, err := tfExecHelper.GetVPCService()
		Expect(err).ToNot(HaveOccurred())
		_, err = vpcService.Destroy(true)
		if err != nil {
			return err
		}

	}
	if profile.STS {
		// Destroy oidc and operator roles
		oidcOpService, err := tfExecHelper.GetOIDCProviderOperatorRoleService()
		Expect(err).ToNot(HaveOccurred())
		_, err = oidcOpService.Destroy(true)
		if err != nil {
			return err
		}

		//  Destroy Account roles
		accService, err := tfExecHelper.GetAccountRoleService()
		Expect(err).ToNot(HaveOccurred())
		_, err = accService.Destroy(true)
		if err != nil {
			return err
		}

	}
	if profile.KMSKey {
		//Destroy KMS Key
		kmsService, err := tfExecHelper.GetKMSService()
		Expect(err).ToNot(HaveOccurred())
		_, err = kmsService.Destroy(true)
		if err != nil {
			return err
		}
	}
	return nil
}

// PrepareRHCSClusterByProfileENV will be used for all day2 tests.
// Do not need to create a cluster, it needs an existing cluster
// Two ways:
//   - If you created a cluster by other way, you can Export CLUSTER_ID=<cluster id>
//   - If you are using this CI created the cluster, just need to Export CLUSTER_PROFILE=<profile name>
func PrepareRHCSClusterByProfileENV() (string, error) {
	// Support the cluster ID to set to ENV in case somebody created cluster by other way
	if os.Getenv(constants.ClusterIDEnv) != "" {
		return os.Getenv(constants.ClusterIDEnv), nil
	}
	if os.Getenv(constants.RhcsClusterProfileENV) == "" {
		Logger.Warnf("Either env variables %s and %s set. Will return an empty string.", constants.ClusterIDEnv, constants.RhcsClusterProfileENV)
		return "", nil
	}
	profile := LoadProfileYamlFileByENV()
	tfExecHelper, err := GetTerraformExecHelperForProfile(profile)
	if err != nil {
		return "", err
	}
	clusterService, err := tfExecHelper.GetClusterService()
	if err != nil {
		return "", err
	}
	clusterOutput, err := clusterService.Output()
	clusterID := clusterOutput.ClusterID
	return clusterID, err
}

func (profile *Profile) GetClusterType() constants.ClusterType {
	return constants.FindClusterType(profile.ClusterType)
}

func (profile *Profile) IsPrivateLink() bool {
	if profile.GetClusterType().HCP {
		return profile.Private
	} else {
		return profile.PrivateLink
	}
}
