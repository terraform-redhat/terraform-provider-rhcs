package ci

***REMOVED***
***REMOVED***
	"os"
	"regexp"
	"strings"

***REMOVED***
	client "github.com/openshift-online/ocm-sdk-go"

	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	HELPER "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
***REMOVED***

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

func PrepareVPC(region string, privateLink bool, multiZone bool, azIDs []string, name ...string***REMOVED*** (*EXE.VPCOutput, error***REMOVED*** {

	vpcService := EXE.NewVPCService(***REMOVED***
	vpcArgs := &EXE.VPCArgs{
		AWSRegion: region,
		MultiAZ:   multiZone,
		VPCCIDR:   CON.DefaultVPCCIDR,
	}

	if len(azIDs***REMOVED*** != 0 {
		turnedZoneIDs := []string{}
		for _, zone := range azIDs {
			if strings.Contains(zone, region***REMOVED*** {
				turnedZoneIDs = append(turnedZoneIDs, zone***REMOVED***
	***REMOVED*** else {
				turnedZoneIDs = append(turnedZoneIDs, region+zone***REMOVED***
	***REMOVED***
***REMOVED***
		vpcArgs.AZIDs = turnedZoneIDs
	}
	if len(name***REMOVED*** == 1 {
		vpcArgs.Name = name[0]
	}
	err := vpcService.Create(vpcArgs***REMOVED***
	if err != nil {
		vpcService.Destroy(***REMOVED***
	}
	output, err := vpcService.Output(***REMOVED***

	if err != nil {
		vpcService.Destroy(***REMOVED***
		return nil, err
	}
	return output, err
}

func PrepareAccountRoles(token string, accountRolePrefix string, awsRegion string, openshiftVersion string, channelGroup string***REMOVED*** (
	*EXE.AccountRolesOutput, error***REMOVED*** {
	accService, err := EXE.NewAccountRoleService(***REMOVED***
	if err != nil {
		return nil, err
	}
	args := &EXE.AccountRolesArgs{
		AccountRolePrefix: accountRolePrefix,
		OpenshiftVersion:  openshiftVersion,
		Token:             token,
		ChannelGroup:      channelGroup,
	}
	accRoleOutput, err := accService.Create(args***REMOVED***
	if err != nil {
		accService.Destroy(***REMOVED***
	}
	return accRoleOutput, err
}

func PrepareOIDCProviderAndOperatorRoles(token string, oidcConfigType string, operatorRolePrefix string, accountRolePrefix string, awsRegion string***REMOVED*** (
	*EXE.OIDCProviderOperatorRolesOutput, error***REMOVED*** {
	oidcOpService, err := EXE.NewOIDCProviderOperatorRolesService(***REMOVED***
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
	oidcOpOutput, err := oidcOpService.Create(args***REMOVED***
	if err != nil {
		oidcOpService.Destroy(***REMOVED***
	}
	return oidcOpOutput, err

}

// PrepareVersion supports below types
// version with a openshift version like 4.13.12
// version with latest
// verion with x-1, it means the version will choose one with x-1 version which can be used for x stream upgrade
// version with y-1, it means the version will choose one with y-1 version which can be used for y stream upgrade
func PrepareVersion(connection *client.Connection, versionTag string, channelGroup string***REMOVED*** string {
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\-*[\s\S]*$`***REMOVED***
	// Check that the version is matching openshift version regexp
	if versionRegex.MatchString(versionTag***REMOVED*** {
		return versionTag
	}
	var vResult string
	switch versionTag {
	case "", "latest":
		versions := cms.EnabledVersions(connection, channelGroup, "", true***REMOVED***
		versions = cms.SortVersions(versions***REMOVED***
		vResult = versions[len(versions***REMOVED***-1].RawID
	case "y-1":
		versions, _ := cms.GetVersionsWithUpgrades(connection, channelGroup, CON.Y, true, false, 1***REMOVED***
		vResult = versions[len(versions***REMOVED***-1].RawID
	case "z-1":
		versions, _ := cms.GetVersionsWithUpgrades(connection, channelGroup, CON.Z, true, false, 1***REMOVED***
		vResult = versions[len(versions***REMOVED***-1].RawID
	case "eol":
		vResult = ""
	}
	return vResult
}
func PrepareProxy(***REMOVED*** {}

func PrepareKMSKey(***REMOVED*** {}

func PrepareRoute53(***REMOVED*** {}

func GenerateClusterCreationArgsByProfile(token string, profile *Profile***REMOVED*** (clusterArgs *EXE.ClusterCreationArgs, manifestsDir string, err error***REMOVED*** {
	profile.Version = PrepareVersion(RHCSConnection, profile.Version, profile.ChannelGroup***REMOVED***

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
		clusterArgs.ClusterName = HELPER.GenerateClusterName(profile.Name***REMOVED***
	}
	if profile.AdminEnabled {
		// placeholder for admin enabled automation
	}
	if profile.Region != "" {
		clusterArgs.AWSRegion = profile.Region
	} else {
		clusterArgs.AWSRegion = CON.DefaultAWSRegion
	}

	if profile.STS {
		accountRolesOutput, err := PrepareAccountRoles(token, clusterArgs.ClusterName, clusterArgs.AWSRegion, profile.Version, profile.ChannelGroup***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		clusterArgs.AccountRolePrefix = accountRolesOutput.AccountRolePrefix

		oidcOutput, err := PrepareOIDCProviderAndOperatorRoles(token, profile.OIDCConfig, clusterArgs.ClusterName, accountRolesOutput.AccountRolePrefix, clusterArgs.AWSRegion***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
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
		// Supports ENV set passed to make cluster provision more flexy in prow
		// Export the subnetIDs via env variable if you have existing ones export SubnetIDs=<subnet1>,<subnet2>,<subnet3>
		// Export the availability zones via env variable export AvailabilitiZones=<az1>,<az2>,<az3>
		if os.Getenv("SubnetIDs"***REMOVED*** != "" && os.Getenv("AvailabilitiZones"***REMOVED*** != "" {
			subnetIDs := strings.Split(os.Getenv("SubnetIDs"***REMOVED***, ","***REMOVED***
			azs := strings.Split(os.Getenv("AvailabilitiZones"***REMOVED***, ","***REMOVED***
			clusterArgs.AWSAvailabilityZones = azs
			clusterArgs.AWSSubnetIDs = subnetIDs
***REMOVED*** else {
			if profile.Zones != "" {
				zones = strings.Split(profile.Zones, ","***REMOVED***
	***REMOVED***

			vpcOutput, err = PrepareVPC(profile.Region, profile.PrivateLink, profile.MultiAZ, zones, clusterArgs.ClusterName***REMOVED***
			if err != nil {
				return
	***REMOVED***
			clusterArgs.AWSAvailabilityZones = vpcOutput.AZs
			if vpcOutput.ClusterPrivateSubnets == nil {
				err = fmt.Errorf("error when creating the vpc, check the previous log. The created resources had been destroyed"***REMOVED***
				return
	***REMOVED***
			if profile.PrivateLink {
				clusterArgs.AWSSubnetIDs = vpcOutput.ClusterPrivateSubnets
	***REMOVED*** else {
				clusterArgs.AWSSubnetIDs = append(vpcOutput.ClusterPrivateSubnets, vpcOutput.ClusterPublicSubnets...***REMOVED***
	***REMOVED***
			clusterArgs.MachineCIDR = vpcOutput.VPCCIDR
***REMOVED***
	}

	if profile.ChannelGroup != "" {
		clusterArgs.ChannelGroup = profile.ChannelGroup
	}

	if profile.Tagging {

	}
	if profile.Labeling {
		clusterArgs.DefaultMPLabels = map[string]string{
			"test1": "testdata1",
***REMOVED***
	}
	if profile.NetWorkingSet {
		clusterArgs.MachineCIDR = CON.DefaultVPCCIDR
	}

	if profile.Proxy {
	}

	return clusterArgs, profile.ManifestsDIR, err
}

func LoadProfileYamlFile(profileName string***REMOVED*** *Profile {
	filename := GetYAMLProfileFile(CON.TFYAMLProfile***REMOVED***
	p := HELPER.GetProfile(profileName, filename***REMOVED***
	Logger.Infof("Loaded cluster profile configuration from profile %s : %v", profileName, p.Cluster***REMOVED***
	profile := Profile{
		Name: profileName,
	}
	err := HELPER.MapStructure(p.Cluster, &profile***REMOVED***
	if profile.ManifestsDIR == "" {
		profile.ManifestsDIR = CON.ROSAClassic
	}
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	return &profile
}

func LoadProfileYamlFileByENV(***REMOVED*** *Profile {
	profileEnv := os.Getenv(CON.RhcsClusterProfileENV***REMOVED***
	if profileEnv == "" {
		panic(fmt.Errorf("ENV Variable CLUSTER_PROFILE is empty, please make sure you set the env value"***REMOVED******REMOVED***
	}
	profile := LoadProfileYamlFile(profileEnv***REMOVED***
	// Supporting global env setting to overrite profile settings
	if os.Getenv("CHANNEL_GROUP"***REMOVED*** != "" {
		Logger.Infof("Got global env settings for CHANNEL_GROUP, overwritten the profile setting with value %s", os.Getenv("CHANNEL_GROUP"***REMOVED******REMOVED***
		profile.ChannelGroup = os.Getenv("CHANNEL_GROUP"***REMOVED***
	}
	if os.Getenv("VERSION"***REMOVED*** != "" {
		Logger.Infof("Got global env settings for VERSION, overwritten the profile setting with value %s", os.Getenv("VERSION"***REMOVED******REMOVED***
		profile.Version = os.Getenv("VERSION"***REMOVED***
	}
	if os.Getenv("REGION"***REMOVED*** != "" {
		Logger.Infof("Got global env settings for REGION, overwritten the profile setting with value %s", os.Getenv("REGION"***REMOVED******REMOVED***
		profile.Region = os.Getenv("REGION"***REMOVED***
	}
	return profile
}

func CreateRHCSClusterByProfile(token string, profile *Profile***REMOVED*** (string, error***REMOVED*** {
	creationArgs, _, err := GenerateClusterCreationArgsByProfile(token, profile***REMOVED***
	if err != nil {
		defer DestroyRHCSClusterByProfile(token, profile***REMOVED***
	}
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR***REMOVED***
	if err != nil {
		defer DestroyRHCSClusterByProfile(token, profile***REMOVED***
	}
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	err = clusterService.Create(creationArgs***REMOVED***
	if err != nil {
		clusterService.Destroy(creationArgs***REMOVED***
		return "", err
	}
	clusterID, err := clusterService.Output(***REMOVED***
	return clusterID, err
}

func DestroyRHCSClusterByProfile(token string, profile *Profile***REMOVED*** error {

	// Destroy cluster
	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	clusterArgs := &EXE.ClusterCreationArgs{
		Token:              token,
		AWSRegion:          profile.Region,
		AccountRolePrefix:  "",
		OperatorRolePrefix: "",
	}
	err = clusterService.Destroy(clusterArgs***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

	// Destroy VPC
	if profile.BYOVPC {
		vpcService := EXE.NewVPCService(***REMOVED***
		vpcArgs := &EXE.VPCArgs{
			AWSRegion: profile.Region,
***REMOVED***
		err := vpcService.Destroy(vpcArgs***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}
	if profile.STS {
		// Destroy oidc and operator roles
		oidcOpService, err := EXE.NewOIDCProviderOperatorRolesService(***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		args := &EXE.OIDCProviderOperatorRolesArgs{
			Token:      token,
			OIDCConfig: profile.OIDCConfig,
			AWSRegion:  profile.Region,
***REMOVED***
		err = oidcOpService.Destroy(args***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

		//  Destroy Account roles
		accService, err := EXE.NewAccountRoleService(***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		accargs := &EXE.AccountRolesArgs{
			Token: token,
***REMOVED***
		err = accService.Destroy(accargs***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

	}
	return nil
}

// PrepareRHCSClusterByProfileENV will be used for all day2 tests.
// Do not need to create a cluster, it needs an existing cluster
// Two ways:
//   - If you created a cluster by other way, you can Export CLUSTER_ID=<cluster id>
//   - If you are using this CI created the cluster, just need to Export CLUSTER_PROFILE=<profile name>
func PrepareRHCSClusterByProfileENV(***REMOVED*** string {
	// Support the cluster ID to set to ENV in case somebody created cluster by other way
	// Export CLUSTER_ID=<cluster id>
	if os.Getenv(CON.ClusterIDEnv***REMOVED*** != "" {
		return os.Getenv(CON.ClusterIDEnv***REMOVED***
	}
	if os.Getenv(CON.RhcsClusterProfileENV***REMOVED*** == "" {
		Logger.Warnf("Either env variables %s and %s set. Will return an empty string.", CON.ClusterIDEnv, CON.RhcsClusterProfileENV***REMOVED***
		return ""
	}
	profile := LoadProfileYamlFileByENV(***REMOVED***
	if profile.ManifestsDIR == "" {
		profile.ManifestsDIR = CON.ROSAClassic
	}
	clusterService, err := EXE.NewClusterService(profile.ManifestsDIR***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	clusterID, err := clusterService.Output(***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	return clusterID
}
