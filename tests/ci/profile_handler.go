package ci

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	client "github.com/openshift-online/ocm-sdk-go"

	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	HELPER "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// Profile Provides profile struct for cluster creation be matrix
type Profile struct {
	Name                  string `ini:"name,omitempty" json:"name,omitempty"`
	ClusterName           string `ini:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	ProductID             string `ini:"product_id,omitempty" json:"product_id,omitempty"`
	MajorVersion          string `ini:"major_version,omitempty" json:"major_version,omitempty"`
	Version               string `ini:"version,omitempty" json:"version,omitempty"` //Version supports indicated version started with openshift-v or minor-1
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
	ETCDKMS               bool   `ini:"etcd_kms,omitempty" json:"etcd_kms,omitempty"`
	NetWorkingSet         bool   `ini:"networking_set,omitempty" json:"networking_set,omitempty"`
	Proxy                 bool   `ini:"proxy,omitempty" json:"proxy,omitempty"`
	Hypershift            bool   `ini:"hypershift,omitempty" json:"hypershift,omitempty"`
	OIDCConfig            string `ini:"oidc_config,omitempty" json:"oidc_config,omitempty"`
	ProvisionShard        string `ini:"provisionShard,omitempty" json:"provisionShard,omitempty"`
	Ec2MetadataHttpTokens string `ini:"imdsv2,omitempty" json:"imdsv2,omitempty"`
	ComputeMachineType    string `ini:"compute_machine_type,omitempty" json:"compute_machine_type,omitempty"`
	AuditLogForward       bool   `ini:"auditlog_forward,omitempty" json:"auditlog_forward,omitempty"`
	AdminEnabled          bool   `ini:"admin_enabled,omitempty" json:"admin_enabled,omitempty"`
	ManagedPolicies       bool   `ini:"managed_policies,omitempty" json:"managed_policies,omitempty"`
	WorkerDiskSize        int    `ini:"worker_disk_size,omitempty" json:"worker_disk_size,omitempty"`
	AdditionalSGNumber    int    `ini:"additional_sg_number,omitempty" json:"additional_sg_number,omitempty"`
	ManifestsDIR          string `ini:"manifests_dir,omitempty" json:"manifests_dir,omitempty"`
}

func PrepareVPC(region string, privateLink bool, multiZone bool, azIDs []string, name ...string) (*EXE.VPCOutput, error) {

	vpcService := EXE.NewVPCService()
	vpcArgs := &EXE.VPCArgs{
		AWSRegion: region,
		MultiAZ:   multiZone,
		VPCCIDR:   CON.DefaultVPCCIDR,
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
		vpcArgs.AZIDs = turnedZoneIDs
	}
	if len(name) == 1 {
		vpcArgs.Name = name[0]
	}
	err := vpcService.Create(vpcArgs)
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
	sgService := EXE.NewSecurityGroupService()
	sgArgs := &EXE.SecurityGroupArgs{
		AWSRegion:  region,
		VPCID:      vpcID,
		SGNumber:   sgNumbers,
		NamePrefix: "rhcs-ci",
	}
	err := sgService.Apply(sgArgs)
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

func PrepareAccountRoles(token string, accountRolePrefix string, awsRegion string, openshiftVersion string, channelGroup string) (
	*EXE.AccountRolesOutput, error) {
	accService, err := EXE.NewAccountRoleService()
	if err != nil {
		return nil, err
	}
	args := &EXE.AccountRolesArgs{
		AccountRolePrefix: accountRolePrefix,
		OpenshiftVersion:  openshiftVersion,
		Token:             token,
		ChannelGroup:      channelGroup,
	}
	accRoleOutput, err := accService.Create(args)
	if err != nil {
		accService.Destroy()
	}
	return accRoleOutput, err
}

func PrepareOIDCProviderAndOperatorRoles(token string, oidcConfigType string, operatorRolePrefix string, accountRolePrefix string, awsRegion string) (
	*EXE.OIDCProviderOperatorRolesOutput, error) {
	oidcOpService, err := EXE.NewOIDCProviderOperatorRolesService()
	if err != nil {
		return nil, err
	}
	args := &EXE.OIDCProviderOperatorRolesArgs{
		AccountRolePrefix:  accountRolePrefix,
		OperatorRolePrefix: operatorRolePrefix,
		Token:              token,
		OIDCConfig:         oidcConfigType,
		AWSRegion:          awsRegion,
	}
	oidcOpOutput, err := oidcOpService.Create(args)
	if err != nil {
		oidcOpService.Destroy()
	}
	return oidcOpOutput, err

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
		Logger.Infof("Cluster OCP latest version is set to %s", vResult)
	case "y-1":
		versions, _ := cms.GetVersionsWithUpgrades(connection, channelGroup, CON.Y, true, false, 1)
		vResult = versions[len(versions)-1].RawID
	case "z-1":
		versions, _ := cms.GetVersionsWithUpgrades(connection, channelGroup, CON.Z, true, false, 1)
		vResult = versions[len(versions)-1].RawID
	case "eol":
		vResult = ""
	}
	return vResult
}
func PrepareProxy() {}

func PrepareKMSKey() {}

func PrepareRoute53() {}

func GenerateClusterCreationArgsByProfile(token string, profile *Profile) (clusterArgs *EXE.ClusterCreationArgs, manifestsDir string, err error) {
	profile.Version = PrepareVersion(RHCSConnection, profile.Version, profile.ChannelGroup, profile)

	clusterArgs = &EXE.ClusterCreationArgs{
		Token:            token,
		OpenshiftVersion: profile.Version,
	}

	// Init cluster's args by profile's attributes

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
		clusterArgs.MachineCIDR = CON.DefaultVPCCIDR
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

	if profile.Region == "" {
		profile.Region = CON.DefaultAWSRegion
	}

	if profile.ManifestsDIR == "" {
		profile.ManifestsDIR = CON.ROSAClassic
	}

	if profile.Labeling {
		clusterArgs.DefaultMPLabels = CON.DefaultMPLabels
	}

	if profile.Tagging {
		clusterArgs.Tagging = CON.Tags
	}

	if profile.AdminEnabled {
		userName := CON.ClusterAdminUser
		password := HELPER.GenerateRandomStringWithSymbols(14)
		adminPasswdMap := map[string]string{"username": userName, "password": password}
		clusterArgs.AdminCredentials = adminPasswdMap
		pass := []byte(password)
		err = os.WriteFile(path.Join(CON.GetRHCSOutputDir(), CON.ClusterAdminUser), pass, 0644)
		if err != nil {
			return
		}
		Logger.Infof("Admin password is written to the output directory")

	}

	if profile.AuditLogForward {
		// ToDo
	}

	if profile.Proxy {
		// ToDo
	}

	if profile.ClusterName != "" {
		clusterArgs.ClusterName = profile.ClusterName
	} else {
		// Generate random chars later cluster name with profile name
		clusterArgs.ClusterName = HELPER.GenerateClusterName(profile.Name)
	}

	if profile.Region != "" {
		clusterArgs.AWSRegion = profile.Region
	} else {
		clusterArgs.AWSRegion = CON.DefaultAWSRegion
	}

	if profile.STS {
		accountRolesOutput, err := PrepareAccountRoles(token, clusterArgs.ClusterName, clusterArgs.AWSRegion, profile.MajorVersion, profile.ChannelGroup)
		Expect(err).ToNot(HaveOccurred())
		clusterArgs.AccountRolePrefix = accountRolesOutput.AccountRolePrefix
		Logger.Infof("Created account roles with prefix %s", accountRolesOutput.AccountRolePrefix)

		Logger.Infof("Sleep for 10 sec to let aws account role async creation finished")
		time.Sleep(10 * time.Second)

		oidcOutput, err := PrepareOIDCProviderAndOperatorRoles(token, profile.OIDCConfig, clusterArgs.ClusterName, accountRolesOutput.AccountRolePrefix, clusterArgs.AWSRegion)
		Expect(err).ToNot(HaveOccurred())
		clusterArgs.OIDCConfigID = oidcOutput.OIDCConfigID
		clusterArgs.OperatorRolePrefix = oidcOutput.OperatorRolePrefix
	}

	if profile.BYOVPC {
		var zones []string
		var vpcOutput *EXE.VPCOutput

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

			vpcOutput, err = PrepareVPC(profile.Region, profile.PrivateLink, profile.MultiAZ, zones, clusterArgs.ClusterName)
			if err != nil {
				return
			}
			clusterArgs.AWSAvailabilityZones = vpcOutput.AZs
			if vpcOutput.ClusterPrivateSubnets == nil {
				err = fmt.Errorf("error when creating the vpc, check the previous log. The created resources had been destroyed")
				return
			}
			if profile.Private {
				clusterArgs.Private = profile.Private
				if profile.PrivateLink {
					clusterArgs.PrivateLink = profile.PrivateLink
					clusterArgs.AWSSubnetIDs = vpcOutput.ClusterPrivateSubnets
				}
			} else {
				clusterArgs.AWSSubnetIDs = append(vpcOutput.ClusterPrivateSubnets, vpcOutput.ClusterPublicSubnets...)
			}
			clusterArgs.MachineCIDR = vpcOutput.VPCCIDR
			if profile.AdditionalSGNumber != 0 {
				var sgIDs []string
				// Prepare profile.AdditionalSGNumber+5 security groups for negative testing
				sgIDs, err = PrepareAdditionalSecurityGroups(profile.Region, vpcOutput.VPCID, profile.AdditionalSGNumber+5)
				if err != nil {
					return
				}
				clusterArgs.AdditionalComputeSecurityGroups = sgIDs[0:profile.AdditionalSGNumber]
				clusterArgs.AdditionalInfraSecurityGroups = sgIDs[0:profile.AdditionalSGNumber]
				clusterArgs.AdditionalControlPlaneSecurityGroups = sgIDs[0:profile.AdditionalSGNumber]
			}
		}
	}

	if profile.WorkerDiskSize != 0 {
		clusterArgs.WorkerDiskSize = profile.WorkerDiskSize
	}

	return clusterArgs, profile.ManifestsDIR, err
}

func LoadProfileYamlFile(profileName string) *Profile {
	filename := GetYAMLProfileFile(CON.TFYAMLProfile)
	p := HELPER.GetProfile(profileName, filename)
	Logger.Infof("Loaded cluster profile configuration from profile %s : %v", profileName, p.Cluster)
	profile := Profile{
		Name: profileName,
	}
	err := HELPER.MapStructure(p.Cluster, &profile)
	if profile.ManifestsDIR == "" {
		profile.ManifestsDIR = CON.ROSAClassic
	}
	Expect(err).ToNot(HaveOccurred())
	return &profile
}

func LoadProfileYamlFileByENV() *Profile {
	profileEnv := os.Getenv(CON.RhcsClusterProfileENV)
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
	creationArgs, _, err := GenerateClusterCreationArgsByProfile(token, profile)
	if err != nil {
		defer DestroyRHCSClusterByProfile(token, profile)
	}
	Expect(err).ToNot(HaveOccurred())
	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
	if err != nil {
		defer DestroyRHCSClusterByProfile(token, profile)
	}
	Expect(err).ToNot(HaveOccurred())
	err = clusterService.Create(creationArgs)
	if err != nil {
		clusterService.Destroy(creationArgs)
		return "", err
	}
	clusterOutput, err := clusterService.Output()
	clusterID := clusterOutput.ClusterID
	return clusterID, err
}

func DestroyRHCSClusterByProfile(token string, profile *Profile) error {

	// Destroy cluster
	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
	Expect(err).ToNot(HaveOccurred())
	clusterArgs := &EXE.ClusterCreationArgs{
		Token:              token,
		AWSRegion:          profile.Region,
		AccountRolePrefix:  "",
		OperatorRolePrefix: "",
	}
	if profile.AdditionalSGNumber != 0 {
		output, err := clusterService.Output()
		Expect(err).ToNot(HaveOccurred())
		clusterArgs.AdditionalComputeSecurityGroups = output.AdditionalComputeSecurityGroups
		clusterArgs.AdditionalInfraSecurityGroups = output.AdditionalInfraSecurityGroups
		clusterArgs.AdditionalControlPlaneSecurityGroups = output.AdditionalControlPlaneSecurityGroups
	}
	_, err = clusterService.Destroy(clusterArgs)
	if err != nil {
		return err
	}

	// Destroy VPC
	if profile.BYOVPC {
		if profile.AdditionalSGNumber != 0 {
			sgService := EXE.NewSecurityGroupService()
			sgArgs := &EXE.SecurityGroupArgs{
				AWSRegion: profile.Region,
				SGNumber:  profile.AdditionalSGNumber,
			}
			err := sgService.Destroy(sgArgs)
			if err != nil {
				return err
			}
		}
		vpcService := EXE.NewVPCService()
		vpcArgs := &EXE.VPCArgs{
			AWSRegion: profile.Region,
		}
		err := vpcService.Destroy(vpcArgs)
		if err != nil {
			return err
		}
	}
	if profile.STS {
		// Destroy oidc and operator roles
		oidcOpService, err := EXE.NewOIDCProviderOperatorRolesService()
		if err != nil {
			return err
		}
		args := &EXE.OIDCProviderOperatorRolesArgs{
			Token:      token,
			OIDCConfig: profile.OIDCConfig,
			AWSRegion:  profile.Region,
		}
		err = oidcOpService.Destroy(args)
		if err != nil {
			return err
		}

		//  Destroy Account roles
		accService, err := EXE.NewAccountRoleService()
		if err != nil {
			return err
		}
		accargs := &EXE.AccountRolesArgs{
			Token: token,
		}
		err = accService.Destroy(accargs)
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
	if os.Getenv(CON.ClusterIDEnv) != "" {
		return os.Getenv(CON.ClusterIDEnv), nil
	}
	if os.Getenv(CON.RhcsClusterProfileENV) == "" {
		Logger.Warnf("Either env variables %s and %s set. Will return an empty string.", CON.ClusterIDEnv, CON.RhcsClusterProfileENV)
		return "", nil
	}
	profile := LoadProfileYamlFileByENV()
	if profile.ManifestsDIR == "" {
		profile.ManifestsDIR = CON.ROSAClassic
	}
	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
	if err != nil {
		return "", err
	}
	clusterOutput, err := clusterService.Output()
	clusterID := clusterOutput.ClusterID
	return clusterID, err
}
