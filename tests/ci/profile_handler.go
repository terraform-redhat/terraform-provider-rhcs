package ci

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	client "github.com/openshift-online/ocm-sdk-go"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// Profile Provides profile struct for cluster creation be matrix
type Profile struct {
	Name                    string `ini:"name,omitempty" json:"name,omitempty"`
	ClusterName             string `ini:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	DomainPrefix            string `ini:"domain_prefix,omitempty" json:"domain_prefix,omitempty"`
	ClusterType             string `ini:"cluster_type,omitempty" json:"cluster_type,omitempty"`
	ProductID               string `ini:"product_id,omitempty" json:"product_id,omitempty"`
	MajorVersion            string `ini:"major_version,omitempty" json:"major_version,omitempty"`
	Version                 string `ini:"version,omitempty" json:"version,omitempty"`                 //Specific OCP version to be specified
	VersionPattern          string `ini:"version_pattern,omitempty" json:"version_pattern,omitempty"` //Version supports indicated version started with openshift-v or major-1 (y-1) or minor-1 (z-1)
	ChannelGroup            string `ini:"channel_group,omitempty" json:"channel_group,omitempty"`
	CloudProvider           string `ini:"cloud_provider,omitempty" json:"cloud_provider,omitempty"`
	Region                  string `ini:"region,omitempty" json:"region,omitempty"`
	InstanceType            string `ini:"instance_type,omitempty" json:"instance_type,omitempty"`
	Zones                   string `ini:"zones,omitempty" json:"zones,omitempty"`           // zones should be like a,b,c,d
	StorageLB               bool   `ini:"storage_lb,omitempty" json:"storage_lb,omitempty"` // the unit is GIB, don't support unit set
	Tagging                 bool   `ini:"tagging,omitempty" json:"tagging,omitempty"`
	Labeling                bool   `ini:"labeling,omitempty" json:"labeling,omitempty"`
	Etcd                    bool   `ini:"etcd_encryption,omitempty" json:"etcd_encryption,omitempty"`
	FIPS                    bool   `ini:"fips,omitempty" json:"fips,omitempty"`
	CCS                     bool   `ini:"ccs,omitempty" json:"ccs,omitempty"`
	STS                     bool   `ini:"sts,omitempty" json:"sts,omitempty"`
	Autoscale               bool   `ini:"autoscaling_enabled,omitempty" json:"autoscaling_enabled,omitempty"`
	MultiAZ                 bool   `ini:"multi_az,omitempty" json:"multi_az,omitempty"`
	BYOVPC                  bool   `ini:"byovpc,omitempty" json:"byovpc,omitempty"`
	PrivateLink             bool   `ini:"private_link,omitempty" json:"private_link,omitempty"`
	Private                 bool   `ini:"private,omitempty" json:"private,omitempty"`
	BYOK                    bool   `ini:"byok,omitempty" json:"byok,omitempty"`
	KMSKey                  bool   `ini:"kms_key_arn,omitempty" json:"kms_key_arn,omitempty"`
	DifferentEncryptionKeys bool   `ini:"different_encryption_keys,omitempty" json:"different_encryption_keys,omitempty"`
	NetWorkingSet           bool   `ini:"networking_set,omitempty" json:"networking_set,omitempty"`
	Proxy                   bool   `ini:"proxy,omitempty" json:"proxy,omitempty"`
	OIDCConfig              string `ini:"oidc_config,omitempty" json:"oidc_config,omitempty"`
	ProvisionShard          string `ini:"provisionShard,omitempty" json:"provisionShard,omitempty"`
	Ec2MetadataHttpTokens   string `ini:"ec2_metadata_http_tokens,omitempty" json:"ec2_metadata_http_tokens,omitempty"`
	ComputeReplicas         int    `ini:"compute_replicas,omitempty" json:"compute_replicas,omitempty"`
	ComputeMachineType      string `ini:"compute_machine_type,omitempty" json:"compute_machine_type,omitempty"`
	AuditLogForward         bool   `ini:"auditlog_forward,omitempty" json:"auditlog_forward,omitempty"`
	AdminEnabled            bool   `ini:"admin_enabled,omitempty" json:"admin_enabled,omitempty"`
	ManagedPolicies         bool   `ini:"managed_policies,omitempty" json:"managed_policies,omitempty"`
	WorkerDiskSize          int    `ini:"worker_disk_size,omitempty" json:"worker_disk_size,omitempty"`
	AdditionalSGNumber      int    `ini:"additional_sg_number,omitempty" json:"additional_sg_number,omitempty"`
	UnifiedAccRolesPath     string `ini:"unified_acc_role_path,omitempty" json:"unified_acc_role_path,omitempty"`
	SharedVpc               bool   `ini:"shared_vpc,omitempty" json:"shared_vpc,omitempty"`
	MachineCIDR             string `ini:"machine_cidr,omitempty" json:"machine_cidr,omitempty"`
	ServiceCIDR             string `ini:"service_cidr,omitempty" json:"service_cidr,omitempty"`
	PodCIDR                 string `ini:"pod_cidr,omitempty" json:"pod_cidr,omitempty"`
	HostPrefix              int    `ini:"host_prefix,omitempty" json:"host_prefix,omitempty"`
	FullResources           bool   `ini:"full_resources,omitempty" json:"full_resources,omitempty"`
	DontWaitForCluster      bool   `ini:"no_wait_cluster,omitempty" json:"no_wait_cluster,omitempty"`
}

func PrepareVPC(region string, multiZone bool, azIDs []string, clusterType constants.ClusterType, name string, sharedVpcAWSSharedCredentialsFile string) (*exec.VPCOutput, error) {
	vpcService, err := exec.NewVPCService(constants.GetAWSVPCDefaultManifestDir(clusterType))
	if err != nil {
		return nil, err
	}
	vpcArgs := &exec.VPCArgs{
		AWSRegion: helper.StringPointer(region),
		VPCCIDR:   helper.StringPointer(constants.DefaultVPCCIDR),
	}

	if len(azIDs) != 0 {
		turnedZoneIDs := []string{}
		for _, zone := range azIDs {
			if strings.Contains(zone, region) {
				turnedZoneIDs = append(turnedZoneIDs, zone)
			} else {
				turnedZoneIDs = append(turnedZoneIDs, region+zone)
			}
		}
		vpcArgs.AvailabilityZones = helper.StringSlicePointer(turnedZoneIDs)
	} else {
		azCount := 1
		if multiZone {
			azCount = 3
		}
		vpcArgs.AvailabilityZonesCount = helper.IntPointer(azCount)
	}
	if name != "" {
		vpcArgs.NamePrefix = helper.StringPointer(name)
	}

	if sharedVpcAWSSharedCredentialsFile != "" {
		vpcArgs.AWSSharedCredentialsFiles = helper.StringSlicePointer([]string{sharedVpcAWSSharedCredentialsFile})
	}

	_, err = vpcService.Apply(vpcArgs)
	if err != nil {
		vpcService.Destroy()
		return nil, err
	}
	output, err := vpcService.Output()
	if err != nil {
		vpcService.Destroy()
		return nil, err
	}
	return output, err
}

func PrepareAdditionalSecurityGroups(region string, vpcID string, sgNumbers int) ([]string, error) {
	sgService, err := exec.NewSecurityGroupService()
	if err != nil {
		return nil, err
	}
	sgArgs := &exec.SecurityGroupArgs{
		AWSRegion:  helper.StringPointer(region),
		VPCID:      helper.StringPointer(vpcID),
		SGNumber:   helper.IntPointer(sgNumbers),
		NamePrefix: helper.StringPointer("rhcs-ci"),
	}
	_, err = sgService.Apply(sgArgs)
	if err != nil {
		sgService.Destroy()
		return nil, err
	}
	output, err := sgService.Output()
	if err != nil {
		sgService.Destroy()
		return nil, err
	}
	return output.SGIDs, err
}

func PrepareAccountRoles(token string, accountRolePrefix string, accountRolesPath string, awsRegion string, openshiftVersion string, channelGroup string, clusterType constants.ClusterType, sharedVpcRoleArn string) (
	*exec.AccountRolesOutput, error) {
	accService, err := exec.NewAccountRoleService(constants.GetAccountRoleDefaultManifestDir(clusterType))
	if err != nil {
		return nil, err
	}
	args := &exec.AccountRolesArgs{
		AccountRolePrefix:   helper.StringPointer(accountRolePrefix),
		OpenshiftVersion:    helper.StringPointer(openshiftVersion),
		UnifiedAccRolesPath: helper.StringPointer(accountRolesPath),
	}

	if sharedVpcRoleArn != "" {
		args.SharedVpcRoleArn = helper.StringPointer(sharedVpcRoleArn)
	}

	_, err = accService.Apply(args)
	if err != nil {
		accService.Destroy()
		return nil, err
	}
	return accService.Output()
}

func PrepareOIDCProviderAndOperatorRoles(token string, oidcConfigType string, operatorRolePrefix string, accountRolePrefix string, accountRolesPath string, clusterType constants.ClusterType, awsRegion string) (
	*exec.OIDCProviderOperatorRolesOutput, error) {
	oidcOpService, err := exec.NewOIDCProviderOperatorRolesService(constants.GetOIDCProviderOperatorRolesDefaultManifestDir(clusterType))
	if err != nil {
		return nil, err
	}
	args := &exec.OIDCProviderOperatorRolesArgs{
		AccountRolePrefix:   helper.StringPointer(accountRolePrefix),
		OperatorRolePrefix:  helper.StringPointer(operatorRolePrefix),
		OIDCConfig:          helper.StringPointer(oidcConfigType),
		UnifiedAccRolesPath: helper.StringPointer(accountRolesPath),
	}
	_, err = oidcOpService.Apply(args)
	if err != nil {
		oidcOpService.Destroy()
		return nil, err
	}
	return oidcOpService.Output()

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

func GetMajorVersion(rawVersion string) string {
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+`)
	vResults := versionRegex.FindAllStringSubmatch(rawVersion, 1)
	vResult := ""
	if len(vResults) != 0 {
		vResult = vResults[0][0]
	}
	return vResult
}

func PrepareProxy(region string, VPCID string, subnetPublicID string, keyPairID string) (*exec.ProxyOutput, error) {
	if VPCID == "" {
		return nil, fmt.Errorf("vpc ID is empty. Cannot prepare proxy")
	}
	proxyService, err := exec.NewProxyService()
	if err != nil {
		return nil, err
	}
	proxyArgs := &exec.ProxyArgs{
		ProxyCount:          helper.IntPointer(1),
		Region:              helper.StringPointer(region),
		VPCID:               helper.StringPointer(VPCID),
		PublicSubnetID:      helper.StringPointer(subnetPublicID),
		TrustBundleFilePath: helper.StringPointer(path.Join(cfg.RhcsOutputDir, "ca.cert")),
		KeyPairID:           helper.StringPointer(keyPairID),
	}

	_, err = proxyService.Apply(proxyArgs)
	if err != nil {
		proxyService.Destroy()
		return nil, err
	}
	proxyOutput, err := proxyService.Output()
	if err != nil {
		proxyService.Destroy()
		return nil, err
	}

	return proxyOutput.Proxies[0], err
}

func PrepareKMSKey(manifestDir string, profile *Profile, kmsName string, accountRolePrefix string, accountRolePath string, clusterType constants.ClusterType) (string, error) {
	kmsService, err := exec.NewKMSService(manifestDir)
	if err != nil {
		return "", err
	}
	kmsArgs := &exec.KMSArgs{
		KMSName:           helper.StringPointer(kmsName),
		AWSRegion:         helper.StringPointer(profile.Region),
		AccountRolePrefix: helper.StringPointer(accountRolePrefix),
		AccountRolePath:   helper.StringPointer(accountRolePath),
		TagKey:            helper.StringPointer("Purpose"),
		TagValue:          helper.StringPointer("RHCS automation test"),
		TagDescription:    helper.StringPointer("BYOK Test Key for API automation"),
		HCP:               helper.BoolPointer(clusterType.HCP),
	}

	_, err = kmsService.Apply(kmsArgs)
	if err != nil {
		kmsService.Destroy()
		return "", err
	}
	kmsOutput, err := kmsService.Output()
	if err != nil {
		kmsService.Destroy()
		return "", err
	}
	return kmsOutput.KeyARN, err
}

func PrepareRoute53() (string, error) {
	dnsDomainService, err := exec.NewDnsDomainService()
	if err != nil {
		return "", err
	}
	a := &exec.DnsDomainArgs{}

	_, err = dnsDomainService.Apply(a)
	if err != nil {
		dnsDomainService.Destroy()
		return "", err
	}
	output, err := dnsDomainService.Output()
	if err != nil {
		dnsDomainService.Destroy()
		return "", err
	}
	return output.DnsDomainId, err
}

func PrepareSharedVpcPolicyAndHostedZone(region string,
	sharedVpcAWSSharedCredentialsFile string,
	clusterName string,
	dnsDomainID string,
	ingressOperatorRoleArn string,
	installerRoleArn string,
	clusterAwsAccount string,
	vpcID string,
	subnets []string,
	domainPrefix string) (*exec.SharedVpcPolicyAndHostedZoneOutput, error) {

	sharedVPCService, err := exec.NewSharedVpcPolicyAndHostedZoneService()
	if err != nil {
		return nil, err
	}

	hostedZoneArgs := &exec.SharedVpcPolicyAndHostedZoneArgs{
		SharedVpcAWSSharedCredentialsFiles: helper.StringSlicePointer([]string{sharedVpcAWSSharedCredentialsFile}),
		Region:                             helper.StringPointer(region),
		ClusterName:                        helper.StringPointer(clusterName),
		DnsDomainId:                        helper.StringPointer(dnsDomainID),
		IngressOperatorRoleArn:             helper.StringPointer(ingressOperatorRoleArn),
		InstallerRoleArn:                   helper.StringPointer(installerRoleArn),
		ClusterAWSAccount:                  helper.StringPointer(clusterAwsAccount),
		VpcId:                              helper.StringPointer(vpcID),
		Subnets:                            helper.StringSlicePointer(subnets),
	}
	if domainPrefix != "" {
		hostedZoneArgs.DomainPrefix = helper.StringPointer(domainPrefix)
	}

	_, err = sharedVPCService.Apply(hostedZoneArgs)
	if err != nil {
		sharedVPCService.Destroy()
		return nil, err
	}
	output, err := sharedVPCService.Output()
	if err != nil {
		sharedVPCService.Destroy()
		return nil, err
	}
	return output, err
}

func GenerateClusterCreationArgsByProfile(token string, profile *Profile) (clusterArgs *exec.ClusterArgs, err error) {
	profile.Version = PrepareVersion(RHCSConnection, profile.VersionPattern, profile.ChannelGroup, profile)

	clusterArgs = &exec.ClusterArgs{
		OpenshiftVersion: helper.StringPointer(profile.Version),
	}

	// Init cluster's args by profile's attributes

	// For Shared VPC
	var cluster_aws_account string
	var installer_role_arn string
	var ingress_role_arn string

	if profile.FIPS {
		clusterArgs.Fips = helper.BoolPointer(profile.FIPS)
	}

	if profile.MultiAZ {
		clusterArgs.MultiAZ = helper.BoolPointer(profile.MultiAZ)
	}

	if profile.NetWorkingSet {
		clusterArgs.MachineCIDR = helper.StringPointer(constants.DefaultVPCCIDR)
	}

	if profile.Autoscale {
		clusterArgs.Autoscaling = &exec.Autoscaling{
			AutoscalingEnabled: helper.BoolPointer(true),
			MinReplicas:        helper.IntPointer(3),
			MaxReplicas:        helper.IntPointer(6),
		}
	}

	if constants.RHCS.ComputeMachineType != "" {
		clusterArgs.ComputeMachineType = helper.StringPointer(constants.RHCS.ComputeMachineType)
	} else if profile.ComputeMachineType != "" {
		clusterArgs.ComputeMachineType = helper.StringPointer(profile.ComputeMachineType)
	}

	if profile.ComputeReplicas > 0 {
		clusterArgs.Replicas = helper.IntPointer(profile.ComputeReplicas)
	}

	if profile.ChannelGroup != "" {
		clusterArgs.ChannelGroup = helper.StringPointer(profile.ChannelGroup)
	}

	if profile.Ec2MetadataHttpTokens != "" {
		clusterArgs.Ec2MetadataHttpTokens = helper.StringPointer(profile.Ec2MetadataHttpTokens)
	}

	if profile.Region == "" {
		profile.Region = constants.DefaultAWSRegion
	}

	if profile.Labeling {
		clusterArgs.DefaultMPLabels = helper.StringMapPointer(constants.DefaultMPLabels)
	}

	if profile.Tagging {
		clusterArgs.Tags = helper.StringMapPointer(constants.Tags)
	}

	if profile.AdminEnabled {
		userName := constants.ClusterAdminUser
		password := helper.GenerateRandomPassword(14)
		adminPasswdMap := map[string]string{"username": userName, "password": password}
		clusterArgs.AdminCredentials = helper.StringMapPointer(adminPasswdMap)
		pass := []byte(password)
		err = os.WriteFile(path.Join(cfg.RhcsOutputDir, constants.ClusterAdminUser), pass, 0644)
		if err != nil {
			Logger.Error("Error happens when try to record the admin password")
			return
		}
		Logger.Info("Admin password is written to the output directory")

	}

	if profile.AuditLogForward {
		// ToDo
	}
	var clusterName string
	if constants.RHCS.RHCSClusterName != "" {
		clusterName = constants.RHCS.RHCSClusterName
	} else if profile.ClusterName != "" {
		clusterName = profile.ClusterName
	} else {
		// Generate random chars later cluster name with profile name
		name := helper.GenerateClusterName(profile.Name)
		clusterName = name
	}

	// There are some problem for cluster created with name length
	// longer than 15 chars with auto generated domain prefix
	if profile.DomainPrefix == "" && profile.SharedVpc && len(clusterName) > 15 {
		profile.DomainPrefix = helper.GenerateRandomName("shared-vpc", 4)
	}

	clusterArgs.ClusterName = &clusterName
	err = os.WriteFile(cfg.ClusterNameFile, []byte(clusterName), 0644)
	if err != nil {
		Logger.Errorf("Error happens when try to record the cluster name file: %s ",
			err.Error())
		return
	}
	Logger.Infof("Recorded cluster name file: %s with name %s",
		cfg.ClusterNameFile, clusterName)
	// short and re-generate the clusterName when it is longer than 15 chars
	if profile.DomainPrefix != "" {
		clusterArgs.DomainPrefix = &profile.DomainPrefix
	}
	if profile.Region != "" {
		clusterArgs.AWSRegion = &profile.Region
	} else {
		clusterArgs.AWSRegion = &constants.DefaultAWSRegion
	}

	if profile.STS {
		majorVersion := GetMajorVersion(profile.Version)
		var accountRolesOutput *exec.AccountRolesOutput
		var oidcOutput *exec.OIDCProviderOperatorRolesOutput

		shared_vpc_role_arn := ""
		if profile.SharedVpc {
			// FIXME:
			//	To create Shared-VPC compatible policies, we need to pass a role arn to create_account_roles module.
			//  But we got an chicken-egg prolems here:
			//		* The Shared-VPC compatible policie requries installer role
			//		* The install role (account roles) require Shared-VPC ARN.
			//  Use hardcode as a temporary solution.
			shared_vpc_role_arn = fmt.Sprintf("arn:aws:iam::641733028092:role/%s-shared-vpc-role", *clusterArgs.ClusterName)
		}
		accountRolesOutput, err = PrepareAccountRoles(token, *clusterArgs.ClusterName, profile.UnifiedAccRolesPath, *clusterArgs.AWSRegion, majorVersion, profile.ChannelGroup, profile.GetClusterType(), shared_vpc_role_arn)
		if err != nil {
			return
		}
		clusterArgs.AccountRolePrefix = helper.StringPointer(accountRolesOutput.AccountRolePrefix)
		clusterArgs.UnifiedAccRolesPath = helper.StringPointer(profile.UnifiedAccRolesPath)
		Logger.Infof("Created account roles with prefix %s", accountRolesOutput.AccountRolePrefix)

		Logger.Infof("Sleep for 10 sec to let aws account role async creation finished")
		time.Sleep(10 * time.Second)

		oidcOutput, err = PrepareOIDCProviderAndOperatorRoles(token, profile.OIDCConfig, *clusterArgs.ClusterName, accountRolesOutput.AccountRolePrefix, profile.UnifiedAccRolesPath, profile.GetClusterType(), *clusterArgs.AWSRegion)
		if err != nil {
			return
		}
		clusterArgs.OIDCConfigID = &oidcOutput.OIDCConfigID
		clusterArgs.OperatorRolePrefix = &oidcOutput.OperatorRolePrefix

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
			clusterArgs.AWSAvailabilityZones = &azs
			clusterArgs.AWSSubnetIDs = &subnetIDs
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
			vpcOutput, err = PrepareVPC(profile.Region, profile.MultiAZ, zones, profile.GetClusterType(), *clusterArgs.ClusterName, shared_vpc_aws_shared_credentials_file)
			if err != nil {
				return
			}

			if vpcOutput.PrivateSubnets == nil {
				err = fmt.Errorf("error when creating the vpc, check the previous log. The created resources had been destroyed")
				return
			}
			if profile.Private {
				clusterArgs.Private = helper.BoolPointer(profile.Private)
				clusterArgs.PrivateLink = helper.BoolPointer(profile.PrivateLink)
				if profile.IsPrivateLink() {
					clusterArgs.AWSSubnetIDs = &vpcOutput.PrivateSubnets
				}
			} else {
				subnetIDs := vpcOutput.PrivateSubnets
				subnetIDs = append(subnetIDs, vpcOutput.PublicSubnets...)
				clusterArgs.AWSSubnetIDs = &subnetIDs
			}

			if profile.SharedVpc {
				// Base domain
				var base_dns_domain string
				base_dns_domain, err = PrepareRoute53()
				if err != nil {
					return
				}

				// Resources for Shared-VPC
				var sharedVpcPolicyAndHostedZoneOutput *exec.SharedVpcPolicyAndHostedZoneOutput
				sharedVpcPolicyAndHostedZoneOutput, err = PrepareSharedVpcPolicyAndHostedZone(
					profile.Region,
					constants.SharedVpcAWSSharedCredentialsFileENV,
					*clusterArgs.ClusterName,
					base_dns_domain,
					ingress_role_arn,
					installer_role_arn,
					cluster_aws_account,
					vpcOutput.VPCID,
					*clusterArgs.AWSSubnetIDs,
					profile.DomainPrefix)
				if err != nil {
					return
				}

				clusterArgs.BaseDnsDomain = helper.StringPointer(base_dns_domain)
				private_hosted_zone := exec.PrivateHostedZone{
					ID:      sharedVpcPolicyAndHostedZoneOutput.HostedZoneId,
					RoleArn: sharedVpcPolicyAndHostedZoneOutput.SharedRole,
				}
				clusterArgs.PrivateHostedZone = &private_hosted_zone
				/*
					The AZ us-east-1a for VPC-account might not have the same location as us-east-1a for Cluster-account.
					For AZs which will be used in cluster configuration, the values should be the ones in Cluster-account.
				*/
				clusterArgs.AWSAvailabilityZones = &sharedVpcPolicyAndHostedZoneOutput.AvailabilityZones
			} else {
				clusterArgs.AWSAvailabilityZones = &vpcOutput.AvailabilityZones
			}

			clusterArgs.MachineCIDR = helper.StringPointer(vpcOutput.VPCCIDR)
			if profile.AdditionalSGNumber != 0 {
				// Prepare profile.AdditionalSGNumber+5 security groups for negative testing
				sgIDs, err = PrepareAdditionalSecurityGroups(profile.Region, vpcOutput.VPCID, profile.AdditionalSGNumber+5)
				if err != nil {
					return
				}
				clusterArgs.AdditionalComputeSecurityGroups = helper.StringSlicePointer(sgIDs[0:profile.AdditionalSGNumber])
				clusterArgs.AdditionalInfraSecurityGroups = helper.StringSlicePointer(sgIDs[0:profile.AdditionalSGNumber])
				clusterArgs.AdditionalControlPlaneSecurityGroups = helper.StringSlicePointer(sgIDs[0:profile.AdditionalSGNumber])
			}

			// in case Proxy is enabled
			if profile.Proxy {
				var proxyOutput *exec.ProxyOutput
				proxyOutput, err = PrepareProxy(profile.Region, vpcOutput.VPCID, vpcOutput.PublicSubnets[0], *clusterArgs.ClusterName)
				if err != nil {
					return
				}
				proxy := exec.Proxy{
					AdditionalTrustBundle: &proxyOutput.AdditionalTrustBundle,
					HTTPSProxy:            &proxyOutput.HttpsProxy,
					HTTPProxy:             &proxyOutput.HttpProxy,
					NoProxy:               &proxyOutput.NoProxy,
				}
				clusterArgs.Proxy = &proxy
			}
		}
	}

	// Prepare KMS key if needed
	if profile.Etcd || profile.KMSKey {
		var kmskey string
		kmskey, err = PrepareKMSKey(constants.KMSDir, profile, *clusterArgs.ClusterName, *clusterArgs.AccountRolePrefix, profile.UnifiedAccRolesPath, profile.GetClusterType())
		if err != nil {
			return
		}

		if profile.Etcd {
			clusterArgs.Etcd = &profile.Etcd
			clusterArgs.EtcdKmsKeyARN = helper.StringPointer(kmskey)
		}
		if profile.KMSKey {
			clusterArgs.KmsKeyARN = &kmskey
			if profile.GetClusterType().HCP {
				if profile.DifferentEncryptionKeys {
					var etcdKMSKeyArn string
					etcdKMSKeyArn, err = PrepareKMSKey(constants.KMSSecondDir, profile, fmt.Sprintf("%s-2", *clusterArgs.ClusterName), *clusterArgs.AccountRolePrefix, profile.UnifiedAccRolesPath, profile.GetClusterType())
					if err != nil {
						return
					}
					clusterArgs.EtcdKmsKeyARN = helper.StringPointer(etcdKMSKeyArn)
				}
			}

		}
	}

	if profile.MachineCIDR != "" {
		clusterArgs.MachineCIDR = helper.StringPointer(profile.MachineCIDR)
	}
	if profile.ServiceCIDR != "" {
		clusterArgs.ServiceCIDR = helper.StringPointer(profile.ServiceCIDR)
	}
	if profile.PodCIDR != "" {
		clusterArgs.PodCIDR = helper.StringPointer(profile.PodCIDR)
	}
	if profile.HostPrefix > 0 {
		clusterArgs.HostPrefix = helper.IntPointer(profile.HostPrefix)
	}

	if profile.WorkerDiskSize != 0 {
		clusterArgs.WorkerDiskSize = helper.IntPointer(profile.WorkerDiskSize)
	}
	clusterArgs.UnifiedAccRolesPath = helper.StringPointer(profile.UnifiedAccRolesPath)
	clusterArgs.CustomProperties = helper.StringMapPointer(constants.CustomProperties) // id:72450

	if profile.FullResources {
		clusterArgs.FullResources = helper.BoolPointer(true)
	}
	if profile.DontWaitForCluster {
		clusterArgs.WaitForCluster = helper.BoolPointer(false)
	}
	return clusterArgs, err
}

func LoadProfileYamlFile(profileName string) *Profile {
	p, err := helper.GetProfile(profileName, GetYAMLProfilesDir())
	Expect(err).ToNot(HaveOccurred())
	Logger.Infof("Loaded cluster profile configuration from profile %s : %v", profileName, p.Cluster)
	profile := Profile{
		Name: profileName,
	}
	err = helper.MapStructure(p.Cluster, &profile)
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
		defer DestroyRHCSClusterResourcesByProfile(token, profile)
	}
	Expect(err).ToNot(HaveOccurred())
	clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
	if err != nil {
		defer DestroyRHCSClusterResourcesByProfile(token, profile)
	}
	Expect(err).ToNot(HaveOccurred())
	_, err = clusterService.Apply(creationArgs)
	if err != nil {
		clusterService.WriteTFVars(creationArgs)
		defer DestroyRHCSClusterResourcesByProfile(token, profile)
		return "", err
	}
	clusterOutput, err := clusterService.Output()
	if err != nil {
		clusterService.WriteTFVars(creationArgs)
		defer DestroyRHCSClusterResourcesByProfile(token, profile)
		return "", err
	}
	clusterID := clusterOutput.ClusterID
	return clusterID, err
}

func DestroyRHCSClusterResourcesByProfile(token string, profile *Profile) error {
	if os.Getenv("NO_CLUSTER_DESTROY") == "true" {
		Logger.Warn("`NO_CLUSTER_DESTROY` is configured, thus no destroy of resources will happen")
		return nil
	}

	// Destroy cluster
	var errs []error
	clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
	if err != nil {
		errs = append(errs, err)
	} else {
		_, err = clusterService.Destroy()
		if err != nil {
			errs = append(errs, err)
		}
	}
	// Get the cluster name from backend to double check cluster deleted
	clusterName, _ := helper.ReadFile(cfg.ClusterNameFile)
	Logger.Infof("Double checking with the cluster name %s", clusterName)
	if clusterName != "" {
		parameter := map[string]interface{}{
			"search": fmt.Sprintf("name is '%s'", clusterName),
		}
		resp, err := cms.ListClusters(RHCSConnection, parameter)
		if err != nil {
			errs = append(errs, err)
		} else {
			if resp.Size() != 0 {
				Logger.Infof("Got the matched cluster with name %s, deleting via connection directly",
					clusterName)
				_, err = cms.DeleteCluster(RHCSConnection, resp.Items().Get(0).ID())
				if err != nil {
					errs = append(errs, err)
				} else {
					err = cms.WaitClusterDeleted(RHCSConnection, resp.Items().Get(0).ID())
					if err != nil {
						errs = append(errs, err)
					}
				}

			}
		}
	}

	// Destroy VPC
	if profile.BYOVPC {
		if profile.Proxy {
			proxyService, err := exec.NewProxyService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = proxyService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
		if profile.AdditionalSGNumber != 0 {
			sgService, err := exec.NewSecurityGroupService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = sgService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}

		if profile.SharedVpc {
			sharedVpcPolicyAndHostedZoneService, err := exec.NewSharedVpcPolicyAndHostedZoneService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = sharedVpcPolicyAndHostedZoneService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}

			// DNS domain
			dnsDomainService, err := exec.NewDnsDomainService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = dnsDomainService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}

		vpcService, _ := exec.NewVPCService(constants.GetAWSVPCDefaultManifestDir(profile.GetClusterType()))
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = vpcService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if profile.STS {
		// Destroy oidc and operator roles
		oidcOpService, err := exec.NewOIDCProviderOperatorRolesService(constants.GetOIDCProviderOperatorRolesDefaultManifestDir(profile.GetClusterType()))
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = oidcOpService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}

		//  Destroy Account roles
		accService, err := exec.NewAccountRoleService(constants.GetAccountRoleDefaultManifestDir(profile.GetClusterType()))
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = accService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}

	}
	if profile.KMSKey || profile.Etcd {
		//Destroy KMS Key
		kmsService, err := exec.NewKMSService(constants.KMSDir)
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = kmsService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}

		if profile.DifferentEncryptionKeys {
			kmsService, err = exec.NewKMSService(constants.KMSSecondDir)
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = kmsService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
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
	clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
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

func (profile *Profile) GetClusterManifestsDir() string {
	manifestsDir := constants.GetClusterManifestsDir(profile.GetClusterType())
	return manifestsDir
}

func (profile *Profile) IsPrivateLink() bool {
	if profile.GetClusterType().HCP {
		return profile.Private
	} else {
		return profile.PrivateLink
	}
}

func GetRandomProfile(clusterTypes ...constants.ClusterType) (profile *Profile, err error) {
	if len(clusterTypes) > 0 {
		Logger.Infof("Get random profile for cluster types: %v", clusterTypes)
	} else {
		Logger.Info("Get random profile from all profiles")
	}

	profilesMap, err := helper.ParseProfiles(GetYAMLProfilesDir())
	if err != nil {
		return
	}
	profilesNames := make([]string, 0, len(profilesMap))
	for k, v := range profilesMap {
		clusterType := constants.FindClusterType(fmt.Sprintf("%v", v.Cluster["cluster_type"]))
		if !v.NeedSpecificConfig {
			if len(clusterTypes) <= 0 || slices.Contains(clusterTypes, clusterType) {
				profilesNames = append(profilesNames, k)
			}
		}
	}
	Logger.Debugf("Got profile names %v", profilesNames)
	profileName := profilesMap[profilesNames[helper.RandomInt(len(profilesNames))]].Name
	profile = LoadProfileYamlFile(profileName)
	Logger.Debugf("Choose profile: %s", profile.Name)
	return profile, err
}
