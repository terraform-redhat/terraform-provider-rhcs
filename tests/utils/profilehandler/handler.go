package profilehandler

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type ProfileHandler interface {
	Services() ProfileServices
	Prepare() ProfilePrepare
	Profile() ProfileSpec

	CreateRHCSClusterByProfile(token string) (string, error)
	DestroyRHCSClusterResources(token string) error
	GenerateClusterCreationArgs(token string) (*exec.ClusterArgs, error)
	RetrieveClusterID() (string, error)

	// This will create a new profile handler with a different TF workspace (adding `-dup` suffix to current)
	Duplicate() ProfileHandler
	DuplicateRandom() ProfileHandler
}

type ProfilePrepare interface {
	PrepareVPC(multiZone bool, azIDs []string, name string, sharedVpcAWSSharedCredentialsFile string) (*exec.VPCOutput, error)
	PrepareAdditionalSecurityGroups(vpcID string, sgNumbers int) ([]string, error)
	PrepareAccountRoles(token string, accountRolePrefix string, accountRolesPath string, openshiftVersion string, channelGroup string, sharedVpcRoleArn string) (*exec.AccountRolesOutput, error)
	PrepareOIDCProviderAndOperatorRoles(token string, oidcConfigType string, operatorRolePrefix string, accountRolePrefix string, accountRolesPath string) (*exec.OIDCProviderOperatorRolesOutput, error)
	PrepareProxy(VPCID string, subnetPublicID string, keyPairID string) (*exec.ProxyOutput, error)
	PrepareKMSKey(kmsName string, accountRolePrefix string, accountRolePath string) (string, error)
	PrepareRoute53() (string, error)
	PrepareSharedVpcPolicyAndHostedZone(sharedCredentialsFile string, clusterName string, dnsDomainID string, ingressOperatorRoleArn string,
		installerRoleArn string, clusterAwsAccount string, vpcID string, subnets []string, domainPrefix string) (*exec.SharedVpcPolicyAndHostedZoneOutput, error)
	PrepareVersion() string
}

type ProfileServices interface {
	GetAccountRolesService() (exec.AccountRoleService, error)
	GetKMSService() (exec.KMSService, error)
	GetOIDCProviderOperatorRolesService() (exec.OIDCProviderOperatorRolesService, error)
	GetProxyService() (exec.ProxyService, error)
	GetSecurityGroupService() (exec.SecurityGroupService, error)
	GetSharedVPCPolicyAndHostedZoneService() (exec.SharedVpcPolicyAndHostedZoneService, error)
	GetVPCService() (exec.VPCService, error)
	GetVPCTagService() (exec.VPCTagService, error)
	GetClusterService() (exec.ClusterService, error)
	GetClusterAutoscalerService() (exec.ClusterAutoscalerService, error)
	GetClusterWaiterService() (exec.ClusterWaiterService, error)
	GetDnsDomainService() (exec.DnsDomainService, error)
	GetIDPService(idpType constants.IDPType) (exec.IDPService, error)
	GetIngressService() (exec.IngressService, error)
	GetImportService() (exec.ImportService, error)
	GetKubeletConfigService() (exec.KubeletConfigService, error)
	GetMachinePoolsService() (exec.MachinePoolService, error)
	GetRHCSInfoService() (exec.RhcsInfoService, error)
	GetTrustedIPsService() (exec.TrustedIPsService, error)
	GetTuningConfigService() (exec.TuningConfigService, error)
}

type ProfileSpec interface {
	GetClusterType() constants.ClusterType
	GetRegion() string
	GetName() string
	GetChannelGroup() string
	GetVersionPattern() string
	GetMajorVersion() string
	GetComputeMachineType() string
	GetZones() string
	GetEc2MetadataHttpTokens() string
	GetUnifiedAccRolesPath() string
	GetOIDCConfig() string

	GetAdditionalSGNumber() int
	GetComputeReplicas() int
	GetWorkerDiskSize() int
	GetImdsv2() string
	GetMachineCIDR() string
	GetServiceCIDR() string
	GetPodCIDR() string
	GetHostPrefix() int

	IsHCP() bool
	IsPrivateLink() bool
	IsPrivate() bool
	IsMultiAZ() bool
	IsBYOVPC() bool
	IsProxy() bool
	IsEtcd() bool
	IsDifferentEncryptionKeys() bool
	IsAutoscale() bool
	IsAdminEnabled() bool
	IsFIPS() bool
	IsLabeling() bool
	IsTagging() bool
	IsKMSKey() bool
	IsFullResources() bool
}

type profileContext struct {
	profile         *Profile
	tfWorkspaceName string
}

func NewProfileHandlerFromYamlFile() (handler ProfileHandler, err error) {
	profile, err := LoadProfileYamlFileByENV()
	if err != nil {
		return
	}
	handler = newProfileHandler(profile, profile.Name)
	return
}

func NewRandomProfileHandler(clusterTypes ...constants.ClusterType) (handler ProfileHandler, err error) {
	profile, err := getRandomProfile(clusterTypes...)
	if err != nil {
		return
	}
	handler = newProfileHandler(profile, helper.GenerateRandomName(profile.Name, 3))
	return
}

func newProfileHandler(profile *Profile, tfWorkspaceName string) ProfileHandler {
	if profile.Region == "" {
		profile.Region = constants.DefaultAWSRegion
	}
	return &profileContext{
		profile:         profile,
		tfWorkspaceName: tfWorkspaceName,
	}
}

func (ctx *profileContext) Duplicate() ProfileHandler {
	return newProfileHandler(ctx.profile, fmt.Sprintf("%s-dup", ctx.tfWorkspaceName))
}

func (ctx *profileContext) DuplicateRandom() ProfileHandler {
	return newProfileHandler(ctx.profile, helper.GenerateRandomName(ctx.tfWorkspaceName, 3))
}

func (ctx *profileContext) Services() ProfileServices {
	return ctx
}

func (ctx *profileContext) Prepare() ProfilePrepare {
	return ctx
}

func (ctx *profileContext) Profile() ProfileSpec {
	return ctx
}

func (ctx *profileContext) GetName() string {
	return ctx.profile.Name
}

func (ctx *profileContext) GetRegion() string {
	return ctx.profile.Region
}

func (ctx *profileContext) GetChannelGroup() string {
	return ctx.profile.ChannelGroup
}

func (ctx *profileContext) GetVersionPattern() string {
	return ctx.profile.VersionPattern
}

func (ctx *profileContext) GetMajorVersion() string {
	return ctx.profile.MajorVersion
}

func (ctx *profileContext) GetComputeMachineType() string {
	return ctx.profile.ComputeMachineType
}

func (ctx *profileContext) GetZones() string {
	return ctx.profile.Zones
}

func (ctx *profileContext) GetEc2MetadataHttpTokens() string {
	return ctx.profile.Ec2MetadataHttpTokens
}

func (ctx *profileContext) GetUnifiedAccRolesPath() string {
	return ctx.profile.UnifiedAccRolesPath
}

func (ctx *profileContext) GetOIDCConfig() string {
	return ctx.profile.OIDCConfig
}

func (ctx *profileContext) GetAdditionalSGNumber() int {
	return ctx.profile.AdditionalSGNumber
}

func (ctx *profileContext) GetComputeReplicas() int {
	return ctx.profile.ComputeReplicas
}

func (ctx *profileContext) GetWorkerDiskSize() int {
	return ctx.profile.WorkerDiskSize
}

func (ctx *profileContext) GetClusterType() constants.ClusterType {
	return constants.FindClusterType(ctx.profile.ClusterType)
}

func (ctx *profileContext) GetTFWorkspace() string {
	return ctx.tfWorkspaceName
}

func (ctx *profileContext) IsHCP() bool {
	return ctx.GetClusterType().HCP
}

func (ctx *profileContext) IsPrivateLink() bool {
	if ctx.GetClusterType().HCP {
		return ctx.profile.Private
	} else {
		return ctx.profile.PrivateLink
	}
}

func (ctx *profileContext) IsPrivate() bool {
	return ctx.profile.Private
}

func (ctx *profileContext) IsMultiAZ() bool {
	return ctx.profile.MultiAZ
}

func (ctx *profileContext) IsBYOVPC() bool {
	return ctx.profile.BYOVPC
}

func (ctx *profileContext) IsProxy() bool {
	return ctx.profile.Proxy
}

func (ctx *profileContext) IsEtcd() bool {
	return ctx.profile.Etcd
}

func (ctx *profileContext) IsDifferentEncryptionKeys() bool {
	return ctx.profile.DifferentEncryptionKeys
}

func (ctx *profileContext) IsAutoscale() bool {
	return ctx.profile.Autoscale
}

func (ctx *profileContext) IsAdminEnabled() bool {
	return ctx.profile.Autoscale
}

func (ctx *profileContext) IsFIPS() bool {
	return ctx.profile.FIPS
}

func (ctx *profileContext) IsLabeling() bool {
	return ctx.profile.Labeling
}

func (ctx *profileContext) IsTagging() bool {
	return ctx.profile.Tagging
}

func (ctx *profileContext) IsKMSKey() bool {
	return ctx.profile.KMSKey
}

func (ctx *profileContext) IsFullResources() bool {
	return ctx.profile.FullResources
}

func (ctx *profileContext) GetImdsv2() string {
	return ctx.profile.Ec2MetadataHttpTokens
}

func (ctx *profileContext) GetMachineCIDR() string {
	return ctx.profile.MachineCIDR
}

func (ctx *profileContext) GetServiceCIDR() string {
	return ctx.profile.ServiceCIDR
}

func (ctx *profileContext) GetPodCIDR() string {
	return ctx.profile.PodCIDR
}

func (ctx *profileContext) GetHostPrefix() int {
	return ctx.profile.HostPrefix
}

func (ctx *profileContext) PrepareVPC(multiZone bool, azIDs []string, name string, sharedVpcAWSSharedCredentialsFile string) (*exec.VPCOutput, error) {
	region := ctx.profile.Region
	vpcService, err := ctx.Services().GetVPCService()
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

func (ctx *profileContext) PrepareAdditionalSecurityGroups(vpcID string, sgNumbers int) ([]string, error) {
	sgService, err := ctx.Services().GetSecurityGroupService()
	if err != nil {
		return nil, err
	}
	sgArgs := &exec.SecurityGroupArgs{
		AWSRegion:  helper.StringPointer(ctx.profile.Region),
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

func (ctx *profileContext) PrepareAccountRoles(token string, accountRolePrefix string, accountRolesPath string, openshiftVersion string, channelGroup string, sharedVpcRoleArn string) (
	*exec.AccountRolesOutput, error) {
	accService, err := ctx.Services().GetAccountRolesService()
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

func (ctx *profileContext) PrepareOIDCProviderAndOperatorRoles(token string, oidcConfigType string, operatorRolePrefix string, accountRolePrefix string, accountRolesPath string) (
	*exec.OIDCProviderOperatorRolesOutput, error) {
	oidcOpService, err := ctx.Services().GetOIDCProviderOperatorRolesService()
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

func (ctx *profileContext) PrepareProxy(VPCID string, subnetPublicID string, keyPairID string) (*exec.ProxyOutput, error) {
	if VPCID == "" {
		return nil, fmt.Errorf("vpc ID is empty. Cannot prepare proxy")
	}
	proxyService, err := ctx.Services().GetProxyService()
	if err != nil {
		return nil, err
	}
	proxyArgs := &exec.ProxyArgs{
		ProxyCount:          helper.IntPointer(1),
		Region:              helper.StringPointer(ctx.profile.Region),
		VPCID:               helper.StringPointer(VPCID),
		PublicSubnetID:      helper.StringPointer(subnetPublicID),
		TrustBundleFilePath: helper.StringPointer(path.Join(constants.RHCS.RhcsOutputDir, "ca.cert")),
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

func (ctx *profileContext) PrepareKMSKey(kmsName string, accountRolePrefix string, accountRolePath string) (string, error) {
	kmsService, err := ctx.Services().GetKMSService()
	if err != nil {
		return "", err
	}
	kmsArgs := &exec.KMSArgs{
		KMSName:           helper.StringPointer(kmsName),
		AWSRegion:         helper.StringPointer(ctx.profile.Region),
		AccountRolePrefix: helper.StringPointer(accountRolePrefix),
		AccountRolePath:   helper.StringPointer(accountRolePath),
		TagKey:            helper.StringPointer("Purpose"),
		TagValue:          helper.StringPointer("RHCS automation test"),
		TagDescription:    helper.StringPointer("BYOK Test Key for API automation"),
		HCP:               helper.BoolPointer(ctx.GetClusterType().HCP),
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

func (ctx *profileContext) PrepareRoute53() (string, error) {
	dnsDomainService, err := ctx.Services().GetDnsDomainService()
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

func (ctx *profileContext) PrepareSharedVpcPolicyAndHostedZone(sharedCredentialsFile string, clusterName string, dnsDomainID string, ingressOperatorRoleArn string,
	installerRoleArn string, clusterAwsAccount string, vpcID string, subnets []string, domainPrefix string) (*exec.SharedVpcPolicyAndHostedZoneOutput, error) {

	sharedVPCService, err := ctx.Services().GetSharedVPCPolicyAndHostedZoneService()
	if err != nil {
		return nil, err
	}

	hostedZoneArgs := &exec.SharedVpcPolicyAndHostedZoneArgs{
		SharedVpcAWSSharedCredentialsFiles: helper.StringSlicePointer([]string{sharedCredentialsFile}),
		Region:                             helper.StringPointer(ctx.profile.Region),
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

// PrepareVersion supports below types
// version with a openshift version like 4.13.12
// version with latest
// verion with x-1, it means the version will choose one with x-1 version which can be used for x stream upgrade
// version with y-1, it means the version will choose one with y-1 version which can be used for y stream upgrade
func (ctx *profileContext) PrepareVersion() string {
	version := ctx.profile.Version
	versionPattern := ctx.profile.VersionPattern
	channelGroup := ctx.profile.ChannelGroup
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\-*[\s\S]*$`)
	if version != "" {
		if versionRegex.MatchString(version) {
			return version
		}
		versionPattern = version // Version has precedence over version pattern
	}
	// Check that the version is matching openshift version regexp
	if versionRegex.MatchString(versionPattern) {
		return versionPattern
	}
	var vResult string
	switch versionPattern {
	case "", "latest":
		versions := cms.EnabledVersions(cms.RHCSConnection, channelGroup, ctx.profile.MajorVersion, true)
		versions = cms.SortVersions(versions)
		vResult = versions[len(versions)-1].RawID
	case "y-1":
		versions, _ := cms.GetVersionsWithUpgrades(cms.RHCSConnection, channelGroup, constants.Y, true, false, 1)
		vResult = versions[len(versions)-1].RawID
	case "z-1":
		versions, _ := cms.GetVersionsWithUpgrades(cms.RHCSConnection, channelGroup, constants.Z, true, false, 1)
		vResult = versions[len(versions)-1].RawID
	case "eol":
		vResult = ""
	}
	Logger.Infof("Cluster OCP latest version is set to %s", vResult)
	return vResult
}

func (ctx *profileContext) GenerateClusterCreationArgs(token string) (clusterArgs *exec.ClusterArgs, err error) {
	// For Shared VPC
	var clusterAwsAccount string
	var installerRoleArn string
	var ingressRoleArn string

	version := ctx.PrepareVersion()

	clusterArgs = &exec.ClusterArgs{
		OpenshiftVersion: helper.StringPointer(version),
	}

	// Init cluster's args by profile's attributes

	clusterArgs.Fips = helper.BoolPointer(ctx.profile.FIPS)
	clusterArgs.MultiAZ = helper.BoolPointer(ctx.profile.MultiAZ)

	if ctx.profile.NetWorkingSet {
		clusterArgs.MachineCIDR = helper.StringPointer(constants.DefaultVPCCIDR)
	}

	if ctx.profile.Autoscale {
		clusterArgs.Autoscaling = &exec.Autoscaling{
			AutoscalingEnabled: helper.BoolPointer(true),
			MinReplicas:        helper.IntPointer(3),
			MaxReplicas:        helper.IntPointer(6),
		}
	}

	if constants.RHCS.ComputeMachineType != "" {
		clusterArgs.ComputeMachineType = helper.StringPointer(constants.RHCS.ComputeMachineType)
	} else if ctx.profile.ComputeMachineType != "" {
		clusterArgs.ComputeMachineType = helper.StringPointer(ctx.profile.ComputeMachineType)
	}

	if ctx.profile.ComputeReplicas > 0 {
		clusterArgs.Replicas = helper.IntPointer(ctx.profile.ComputeReplicas)
	}

	if ctx.profile.ChannelGroup != "" {
		clusterArgs.ChannelGroup = helper.StringPointer(ctx.profile.ChannelGroup)
	}

	if ctx.profile.Ec2MetadataHttpTokens != "" {
		clusterArgs.Ec2MetadataHttpTokens = helper.StringPointer(ctx.profile.Ec2MetadataHttpTokens)
	}

	if ctx.profile.Labeling {
		clusterArgs.DefaultMPLabels = helper.StringMapPointer(constants.DefaultMPLabels)
	}

	if ctx.profile.Tagging {
		clusterArgs.Tags = helper.StringMapPointer(constants.Tags)
	}

	if ctx.profile.AdminEnabled {
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

	if ctx.profile.AuditLogForward {
		// ToDo
	}

	var clusterName string
	if constants.RHCS.RHCSClusterName != "" {
		clusterName = constants.RHCS.RHCSClusterName
	} else if ctx.profile.ClusterName != "" {
		clusterName = ctx.profile.ClusterName
	} else {
		// Generate random chars later cluster name with profile name
		name := helper.GenerateClusterName(ctx.profile.Name)
		clusterName = name
	}
	clusterArgs.ClusterName = helper.StringPointer(clusterName)

	// There are some problem for cluster created with name length
	// longer than 15 chars with auto generated domain prefix
	if ctx.profile.DomainPrefix == "" && ctx.profile.SharedVpc && len(clusterName) > 15 {
		ctx.profile.DomainPrefix = helper.GenerateRandomName("shared-vpc", 4)
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
	if ctx.profile.DomainPrefix != "" {
		clusterArgs.DomainPrefix = &ctx.profile.DomainPrefix
	}
	if ctx.profile.Region != "" {
		clusterArgs.AWSRegion = &ctx.profile.Region
	} else {
		clusterArgs.AWSRegion = &constants.DefaultAWSRegion
	}

	if ctx.profile.STS {
		majorVersion := helper.GetMajorVersion(version)
		var accountRolesOutput *exec.AccountRolesOutput
		var oidcOutput *exec.OIDCProviderOperatorRolesOutput

		sharedVPCRoleArn := ""
		if ctx.profile.SharedVpc {
			// FIXME:
			//	To create Shared-VPC compatible policies, we need to pass a role arn to create_account_roles module.
			//  But we got an chicken-egg prolems here:
			//		* The Shared-VPC compatible policie requries installer role
			//		* The install role (account roles) require Shared-VPC ARN.
			//  Use hardcode as a temporary solution.
			sharedVPCRoleArn = fmt.Sprintf("arn:aws:iam::641733028092:role/%s-shared-vpc-role", clusterName)
		}
		accountRolesOutput, err = ctx.PrepareAccountRoles(token, clusterName, ctx.profile.UnifiedAccRolesPath, majorVersion, ctx.profile.ChannelGroup, sharedVPCRoleArn)
		if err != nil {
			return
		}
		clusterArgs.AccountRolePrefix = helper.StringPointer(accountRolesOutput.AccountRolePrefix)
		clusterArgs.UnifiedAccRolesPath = helper.StringPointer(ctx.profile.UnifiedAccRolesPath)
		Logger.Infof("Created account roles with prefix %s", accountRolesOutput.AccountRolePrefix)

		Logger.Infof("Sleep for 10 sec to let aws account role async creation finished")
		time.Sleep(10 * time.Second)

		oidcOutput, err = ctx.PrepareOIDCProviderAndOperatorRoles(token, ctx.profile.OIDCConfig, clusterName, accountRolesOutput.AccountRolePrefix, ctx.profile.UnifiedAccRolesPath)
		if err != nil {
			return
		}
		clusterArgs.OIDCConfigID = &oidcOutput.OIDCConfigID
		clusterArgs.OperatorRolePrefix = &oidcOutput.OperatorRolePrefix

		clusterAwsAccount = accountRolesOutput.AWSAccountId
		installerRoleArn = accountRolesOutput.InstallerRoleArn
		ingressRoleArn = oidcOutput.IngressOperatorRoleArn
	}

	if ctx.profile.BYOVPC {
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
			if ctx.profile.Zones != "" {
				zones = strings.Split(ctx.profile.Zones, ",")
			}

			sharedVPCAWSSharedCredentialsFile := ""

			if ctx.profile.SharedVpc {
				if constants.SharedVpcAWSSharedCredentialsFileENV == "" {
					panic(fmt.Errorf("SHARED_VPC_AWS_SHARED_CREDENTIALS_FILE env is not set or empty, it's requried by Shared-VPC cluster"))
				}

				sharedVPCAWSSharedCredentialsFile = constants.SharedVpcAWSSharedCredentialsFileENV
			}
			vpcOutput, err = ctx.PrepareVPC(ctx.profile.MultiAZ, zones, clusterName, sharedVPCAWSSharedCredentialsFile)
			if err != nil {
				return
			}

			if vpcOutput.PrivateSubnets == nil {
				err = fmt.Errorf("error when creating the vpc, check the previous log. The created resources had been destroyed")
				return
			}
			if ctx.profile.Private {
				clusterArgs.Private = helper.BoolPointer(ctx.profile.Private)
				clusterArgs.PrivateLink = helper.BoolPointer(ctx.profile.PrivateLink)
				if ctx.IsPrivateLink() {
					clusterArgs.AWSSubnetIDs = &vpcOutput.PrivateSubnets
				}
			} else {
				subnetIDs := vpcOutput.PrivateSubnets
				subnetIDs = append(subnetIDs, vpcOutput.PublicSubnets...)
				clusterArgs.AWSSubnetIDs = &subnetIDs
			}

			if ctx.profile.SharedVpc {
				// Base domain
				var baseDnsDomain string
				baseDnsDomain, err = ctx.PrepareRoute53()
				if err != nil {
					return
				}

				// Resources for Shared-VPC
				var sharedVpcPolicyAndHostedZoneOutput *exec.SharedVpcPolicyAndHostedZoneOutput
				sharedVpcPolicyAndHostedZoneOutput, err = ctx.PrepareSharedVpcPolicyAndHostedZone(
					constants.SharedVpcAWSSharedCredentialsFileENV,
					clusterName,
					baseDnsDomain,
					ingressRoleArn,
					installerRoleArn,
					clusterAwsAccount,
					vpcOutput.VPCID,
					*clusterArgs.AWSSubnetIDs,
					ctx.profile.DomainPrefix)
				if err != nil {
					return
				}

				clusterArgs.BaseDnsDomain = helper.StringPointer(baseDnsDomain)
				privateHostedZone := exec.PrivateHostedZone{
					ID:      sharedVpcPolicyAndHostedZoneOutput.HostedZoneId,
					RoleArn: sharedVpcPolicyAndHostedZoneOutput.SharedRole,
				}
				clusterArgs.PrivateHostedZone = &privateHostedZone
				/*
					The AZ us-east-1a for VPC-account might not have the same location as us-east-1a for Cluster-account.
					For AZs which will be used in cluster configuration, the values should be the ones in Cluster-account.
				*/
				clusterArgs.AWSAvailabilityZones = &sharedVpcPolicyAndHostedZoneOutput.AvailabilityZones
			} else {
				clusterArgs.AWSAvailabilityZones = &vpcOutput.AvailabilityZones
			}

			clusterArgs.MachineCIDR = helper.StringPointer(vpcOutput.VPCCIDR)
			if ctx.profile.AdditionalSGNumber != 0 {
				// Prepare profile.AdditionalSGNumber+5 security groups for negative testing
				sgIDs, err = ctx.PrepareAdditionalSecurityGroups(vpcOutput.VPCID, ctx.profile.AdditionalSGNumber+5)
				if err != nil {
					return
				}
				clusterArgs.AdditionalComputeSecurityGroups = helper.StringSlicePointer(sgIDs[0:ctx.profile.AdditionalSGNumber])
				clusterArgs.AdditionalInfraSecurityGroups = helper.StringSlicePointer(sgIDs[0:ctx.profile.AdditionalSGNumber])
				clusterArgs.AdditionalControlPlaneSecurityGroups = helper.StringSlicePointer(sgIDs[0:ctx.profile.AdditionalSGNumber])
			}

			// in case Proxy is enabled
			if ctx.profile.Proxy {
				var proxyOutput *exec.ProxyOutput
				proxyOutput, err = ctx.PrepareProxy(vpcOutput.VPCID, vpcOutput.PublicSubnets[0], clusterName)
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
	if ctx.profile.Etcd || ctx.profile.KMSKey {
		var kmskey string
		kmskey, err = ctx.PrepareKMSKey(*clusterArgs.ClusterName, *clusterArgs.AccountRolePrefix, ctx.profile.UnifiedAccRolesPath)
		if err != nil {
			return
		}

		if ctx.profile.Etcd {
			clusterArgs.Etcd = &ctx.profile.Etcd
			clusterArgs.EtcdKmsKeyARN = helper.StringPointer(kmskey)
		}
		if ctx.profile.KMSKey {
			clusterArgs.KmsKeyARN = &kmskey
			if ctx.Profile().IsHCP() {
				if ctx.profile.DifferentEncryptionKeys {
					kmsName := fmt.Sprintf("%s-2", *clusterArgs.ClusterName)
					var etcdKMSKeyArn string
					etcdKMSKeyArn, err = ctx.Duplicate().Prepare().
						PrepareKMSKey(kmsName, *clusterArgs.AccountRolePrefix, ctx.profile.UnifiedAccRolesPath)
					if err != nil {
						return
					}
					clusterArgs.EtcdKmsKeyARN = helper.StringPointer(etcdKMSKeyArn)
				}
			}

		}
	}

	if ctx.profile.MachineCIDR != "" {
		clusterArgs.MachineCIDR = helper.StringPointer(ctx.profile.MachineCIDR)
	}
	if ctx.profile.ServiceCIDR != "" {
		clusterArgs.ServiceCIDR = helper.StringPointer(ctx.profile.ServiceCIDR)
	}
	if ctx.profile.PodCIDR != "" {
		clusterArgs.PodCIDR = helper.StringPointer(ctx.profile.PodCIDR)
	}
	if ctx.profile.HostPrefix > 0 {
		clusterArgs.HostPrefix = helper.IntPointer(ctx.profile.HostPrefix)
	}

	if ctx.profile.WorkerDiskSize != 0 {
		clusterArgs.WorkerDiskSize = helper.IntPointer(ctx.profile.WorkerDiskSize)
	}
	clusterArgs.UnifiedAccRolesPath = helper.StringPointer(ctx.profile.UnifiedAccRolesPath)
	clusterArgs.CustomProperties = helper.StringMapPointer(constants.CustomProperties) // id:72450

	if ctx.profile.FullResources {
		clusterArgs.FullResources = helper.BoolPointer(true)
	}
	if ctx.profile.DontWaitForCluster {
		clusterArgs.WaitForCluster = helper.BoolPointer(false)
	}

	return clusterArgs, err
}

func (ctx *profileContext) CreateRHCSClusterByProfile(token string) (string, error) {
	creationArgs, err := ctx.GenerateClusterCreationArgs(token)
	if err != nil {
		defer ctx.DestroyRHCSClusterResources(token)
		return "", err
	}
	clusterService, err := ctx.Services().GetClusterService()
	if err != nil {
		defer ctx.DestroyRHCSClusterResources(token)
		return "", err
	}
	_, err = clusterService.Apply(creationArgs)
	if err != nil {
		clusterService.WriteTFVars(creationArgs)
		defer ctx.DestroyRHCSClusterResources(token)
		return "", err
	}
	clusterOutput, err := clusterService.Output()
	if err != nil {
		clusterService.WriteTFVars(creationArgs)
		defer ctx.DestroyRHCSClusterResources(token)
		return "", err
	}
	clusterID := clusterOutput.ClusterID
	return clusterID, err
}

func (ctx *profileContext) DestroyRHCSClusterResources(token string) error {
	if os.Getenv("NO_CLUSTER_DESTROY") == "true" {
		Logger.Warn("`NO_CLUSTER_DESTROY` is configured, thus no destroy of resources will happen")
		return nil
	}

	// Destroy cluster
	var errs []error
	clusterService, err := ctx.GetClusterService()
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
		resp, err := cms.ListClusters(cms.RHCSConnection, parameter)
		if err != nil {
			errs = append(errs, err)
		} else {
			if resp.Size() != 0 {
				Logger.Infof("Got the matched cluster with name %s, deleting via connection directly",
					clusterName)
				_, err = cms.DeleteCluster(cms.RHCSConnection, resp.Items().Get(0).ID())
				if err != nil {
					errs = append(errs, err)
				} else {
					err = cms.WaitClusterDeleted(cms.RHCSConnection, resp.Items().Get(0).ID())
					if err != nil {
						errs = append(errs, err)
					}
				}

			}
		}
	}

	// Destroy VPC
	if ctx.profile.BYOVPC {
		if ctx.profile.Proxy {
			proxyService, err := ctx.GetProxyService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = proxyService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
		if ctx.profile.AdditionalSGNumber != 0 {
			sgService, err := ctx.GetSecurityGroupService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = sgService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}

		if ctx.profile.SharedVpc {
			sharedVpcPolicyAndHostedZoneService, err := ctx.GetSharedVPCPolicyAndHostedZoneService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = sharedVpcPolicyAndHostedZoneService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}

			// DNS domain
			dnsDomainService, err := ctx.GetDnsDomainService()
			if err != nil {
				errs = append(errs, err)
			} else {
				_, err = dnsDomainService.Destroy()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}

		vpcService, _ := ctx.GetVPCService()
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = vpcService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if ctx.profile.STS {
		// Destroy oidc and operator roles
		oidcOpService, err := ctx.GetOIDCProviderOperatorRolesService()
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = oidcOpService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}

		//  Destroy Account roles
		accService, err := ctx.GetAccountRolesService()
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = accService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}

	}
	if ctx.profile.KMSKey || ctx.profile.Etcd {
		//Destroy KMS Key
		kmsService, err := ctx.GetKMSService()
		if err != nil {
			errs = append(errs, err)
		} else {
			_, err = kmsService.Destroy()
			if err != nil {
				errs = append(errs, err)
			}
		}

		if ctx.profile.DifferentEncryptionKeys {
			kmsService, err = ctx.Duplicate().Services().GetKMSService()
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

// RetrieveClusterID will be used for all day2 tests. It needs an existing cluster.
// Two ways:
//   - If you created a cluster by other way, you can Export CLUSTER_ID=<cluster id>
//   - If you are using this CI created the cluster, just need to Export CLUSTER_PROFILE=<profile name>
func (ctx *profileContext) RetrieveClusterID() (string, error) {
	// Support the cluster ID to set to ENV in case somebody created cluster by other way
	if os.Getenv(constants.ClusterIDEnv) != "" {
		return os.Getenv(constants.ClusterIDEnv), nil
	}
	if os.Getenv(constants.RhcsClusterProfileENV) == "" {
		Logger.Warnf("Either env variables %s and %s set. Will return an empty string.", constants.ClusterIDEnv, constants.RhcsClusterProfileENV)
		return "", nil
	}
	clusterService, err := ctx.Services().GetClusterService()
	if err != nil {
		return "", err
	}
	clusterOutput, err := clusterService.Output()
	clusterID := clusterOutput.ClusterID
	return clusterID, err
}

/////////////////////////////////////////////////////////////////////////////////////////
// Services Interface for easy retrieval of TF exec services from a profile

func (ctx *profileContext) GetAccountRolesService() (exec.AccountRoleService, error) {
	return exec.NewAccountRoleService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetKMSService() (exec.KMSService, error) {
	return exec.NewKMSService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetOIDCProviderOperatorRolesService() (exec.OIDCProviderOperatorRolesService, error) {
	return exec.NewOIDCProviderOperatorRolesService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetProxyService() (exec.ProxyService, error) {
	return exec.NewProxyService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetSecurityGroupService() (exec.SecurityGroupService, error) {
	return exec.NewSecurityGroupService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetSharedVPCPolicyAndHostedZoneService() (exec.SharedVpcPolicyAndHostedZoneService, error) {
	return exec.NewSharedVpcPolicyAndHostedZoneService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetVPCService() (exec.VPCService, error) {
	return exec.NewVPCService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetVPCTagService() (exec.VPCTagService, error) {
	return exec.NewVPCTagService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

// RHCS provider dirs
func (ctx *profileContext) GetClusterService() (exec.ClusterService, error) {
	return exec.NewClusterService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetClusterAutoscalerService() (exec.ClusterAutoscalerService, error) {
	return exec.NewClusterAutoscalerService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetClusterWaiterService() (exec.ClusterWaiterService, error) {
	return exec.NewClusterWaiterService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetDnsDomainService() (exec.DnsDomainService, error) {
	return exec.NewDnsDomainService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetIDPService(idpType constants.IDPType) (exec.IDPService, error) {
	return exec.NewIDPService(ctx.GetTFWorkspace(), ctx.GetClusterType(), idpType)
}

func (ctx *profileContext) GetIngressService() (exec.IngressService, error) {
	return exec.NewIngressService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetImportService() (exec.ImportService, error) {
	return exec.NewImportService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetKubeletConfigService() (exec.KubeletConfigService, error) {
	return exec.NewKubeletConfigService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetMachinePoolsService() (exec.MachinePoolService, error) {
	return exec.NewMachinePoolService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetRHCSInfoService() (exec.RhcsInfoService, error) {
	return exec.NewRhcsInfoService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetTrustedIPsService() (exec.TrustedIPsService, error) {
	return exec.NewTrustedIPsService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}

func (ctx *profileContext) GetTuningConfigService() (exec.TuningConfigService, error) {
	return exec.NewTuningConfigService(ctx.GetTFWorkspace(), ctx.GetClusterType())
}
