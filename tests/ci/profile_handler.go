package ci

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/gomega"
	client "github.com/openshift-online/ocm-sdk-go"

	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	HELPER "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

// Profile Provides profile struct for cluster creation be matrix
type Profile struct {
	Name                  string `ini:"name,omitempty" json:"name,omitempty"`
	ClusterName           string `ini:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	ProductID             string `ini:"product_id,omitempty" json:"product_id,omitempty"`
	Version               string `ini:"version,omitempty" json:"version,omitempty"` //Version supports indicated version started with openshift-v or minor-1
	ChannelGroup          string `ini:"channel_group,omitempty" json:"channel_group,omitempty"`
	CloudProvider         string `ini:"cloud_provider,omitempty" json:"cloud_provider,omitempty"`
	Region                string `ini:"region,omitempty" json:"region,omitempty"`
	InstanceType          string `ini:"instance_type,omitempty" json:"instance_type,omitempty"`
	Zones                 string `ini:"zones,omitempty" json:"zones,omitempty"`           // zones should be like a,b,c,d
	StorageLB             bool   `ini:"storage_lb,omitempty" json:"storage_lb,omitempty"` // the unit is GIB, don't support unit set
	Tagging               bool   `ini:"tagging,omitempty" json:"tagging,omitempty"`
	Labeling              bool   `ini:"labeling,omitempty" json:"labeling,omitempty"`
	Etcd                  bool   `ini:"etcd,omitempty" json:"etcd,omitempty"`
	FIPS                  bool   `ini:"fips,omitempty" json:"fips,omitempty"`
	CCS                   bool   `ini:"ccs,omitempty" json:"ccs,omitempty"`
	STS                   bool   `ini:"sts,omitempty" json:"sts,omitempty"`
	Autoscale             bool   `ini:"autoscale,omitempty" json:"autoscale,omitempty"`
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
	AuditLogForward       bool   `ini:"auditlog_forward,omitempty" json:"auditlog_forward,omitempty"`
	AdminEnabled          bool   `ini:"admin_enabled,omitempty" json:"admin_enabled,omitempty"`
	ManagedPolicies       bool   `ini:"managed_policies,omitempty" json:"managed_policies,omitempty"`
	VolumeSize            int    `ini:"volume_size,omitempty" json:"volume_size,omitempty"`
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
	}
	output, err := vpcService.Output()

	if err != nil {
		vpcService.Destroy()
		return nil, err
	}
	return output, err
}

func PrepareAccountRoles(token string, accountRolePrefix string, awsRegion string, openshiftVersion string, channelGroup string) (
	*EXE.AccountRolesOutput, error) {
	accService, err := EXE.NewAccountRoleService()
	if err != nil {
		return nil, err
	}
	args := &EXE.AccountRolesArgs{
		AccountRolePrefix: accountRolePrefix,
		// OCMENV            string `json:"ocm_environment,omitempty"`
		OpenshiftVersion: openshiftVersion,
		Token:            token,
		// URL               string `json:"url,omitempty"`
		ChannelGroup: channelGroup,
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
		// URL                string `json:"url,omitempty"`
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
func PrepareVersion(connection *client.Connection, versionTag string, channelGroup string) string {
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\-*[\s\S]*$`)
	// Check that the version is matching openshift version regexp
	if versionRegex.MatchString(versionTag) {
		return versionTag
	}
	var vResult string
	switch versionTag {
	case "", "latest":
		versions := cms.EnabledVersions(connection, channelGroup, "", true)
		versions = cms.SortVersions(versions)
		vResult = versions[len(versions)-1].RawID
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
	profile.Version = PrepareVersion(RHCSConnection, profile.Version, profile.ChannelGroup)

	clusterArgs = &EXE.ClusterCreationArgs{
		Token:            token,
		OpenshiftVersion: profile.Version,
	}
	if profile.ManifestsDIR == "" {
		profile.ManifestsDIR = CON.ROSAClassic
	}
	if profile.ClusterName != "" {
		clusterArgs.ClusterName = profile.ClusterName
	} else {
		// Generate random chars later cluster name with profile name
		clusterArgs.ClusterName = HELPER.GenerateClusterName(profile.Name)
	}
	if profile.AdminEnabled {

	}
	if profile.Region != "" {
		clusterArgs.AWSRegion = profile.Region
	} else {
		clusterArgs.AWSRegion = CON.DefaultAWSRegion
	}

	if profile.STS {
		accountRolesOutput, err := PrepareAccountRoles(token, clusterArgs.ClusterName, clusterArgs.AWSRegion, profile.Version, profile.ChannelGroup)
		Expect(err).ToNot(HaveOccurred())
		clusterArgs.AccountRolePrefix = accountRolesOutput.AccountRolePrefix

		oidcOutput, err := PrepareOIDCProviderAndOperatorRoles(token, profile.OIDCConfig, clusterArgs.ClusterName, accountRolesOutput.AccountRolePrefix, clusterArgs.AWSRegion)
		Expect(err).ToNot(HaveOccurred())
		clusterArgs.OIDCConfigID = oidcOutput.OIDCConfigID
		clusterArgs.OperatorRolePrefix = oidcOutput.OperatorRolePrefix
	}
	if profile.Region == "" {
		profile.Region = CON.DefaultAWSRegion
	}

	if profile.AuditLogForward {

	}
	if profile.MultiAZ {
		clusterArgs.MultiAZ = true
	}

	if profile.BYOVPC {
		var zones []string
		var vpcOutput *EXE.VPCOutput
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
		if profile.PrivateLink {
			clusterArgs.AWSSubnetIDs = vpcOutput.ClusterPrivateSubnets
		} else {
			clusterArgs.AWSSubnetIDs = append(vpcOutput.ClusterPrivateSubnets, vpcOutput.ClusterPublicSubnets...)
		}
		clusterArgs.MachineCIDR = vpcOutput.VPCCIDR
	}

	if profile.ChannelGroup != "" {
		clusterArgs.ChannelGroup = profile.ChannelGroup
	}

	if profile.Tagging {

	}
	if profile.Labeling {
		clusterArgs.DefaultMPLabels = map[string]string{
			"test1": "testdata1",
		}
	}
	if profile.NetWorkingSet {
		clusterArgs.MachineCIDR = CON.DefaultVPCCIDR
	}

	if profile.Proxy {
	}

	return clusterArgs, profile.ManifestsDIR, err
}

func LoadProfileYamlFile(profileName string) *Profile {
	filename := GetYAMLProfileFile(CON.TFYAMLProfile)
	p := HELPER.GetProfile(profileName, filename)
	fmt.Println(p.Cluster)
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
	profileEnv := os.Getenv("CLUSTER_PROFILE")
	if profileEnv == "" {
		panic(fmt.Errorf("ENV Variable CLUSTER_PROFILE is empty, please make sure you set the env value"))
	}
	return LoadProfileYamlFile(profileEnv)
}

func CreateRHCSClusterByProfile(token string, profile *Profile) (string, error) {

	creationArgs, _, err := GenerateClusterCreationArgsByProfile(token, profile)

	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
	err = clusterService.Create(creationArgs)
	if err != nil {
		clusterService.Destroy(creationArgs)
		return "", err
	}
	clusterID, err := clusterService.Output()
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
	err = clusterService.Destroy(clusterArgs)
	Expect(err).ToNot(HaveOccurred())

	// Destroy VPC
	if profile.BYOVPC {
		vpcService := EXE.NewVPCService()
		vpcArgs := &EXE.VPCArgs{
			AWSRegion: profile.Region,
		}
		err := vpcService.Destroy(vpcArgs)
		Expect(err).ToNot(HaveOccurred())
	}
	if profile.STS {
		// Destroy oidc and operator roles
		oidcOpService, err := EXE.NewOIDCProviderOperatorRolesService()
		Expect(err).ToNot(HaveOccurred())
		args := &EXE.OIDCProviderOperatorRolesArgs{
			Token:      token,
			OIDCConfig: profile.OIDCConfig,
			AWSRegion:  profile.Region,
			// URL                string `json:"url,omitempty"`
		}
		err = oidcOpService.Destroy(args)
		Expect(err).ToNot(HaveOccurred())

		//  Destroy Account roles
		accService, err := EXE.NewAccountRoleService()
		Expect(err).ToNot(HaveOccurred())
		accargs := &EXE.AccountRolesArgs{
			Token: token,
		}
		err = accService.Destroy(accargs)
		Expect(err).ToNot(HaveOccurred())

	}
	return nil
}
func PrepareRHCSClusterByProfileENV() string {
	profile := LoadProfileYamlFileByENV()
	if profile.ManifestsDIR == "" {
		profile.ManifestsDIR = CON.ROSAClassic
	}
	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
	Expect(err).ToNot(HaveOccurred())
	clusterID, err := clusterService.Output()
	Expect(err).ToNot(HaveOccurred())
	return clusterID
}
