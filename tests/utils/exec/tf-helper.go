package exec

import (
	"path"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

type TerraformExecHelper struct {
	tfExec      TerraformExec
	clusterType constants.ClusterType
}

func NewTerraformExecHelper(clusterType constants.ClusterType, tfExec TerraformExec) *TerraformExecHelper {
	return &TerraformExecHelper{
		tfExec:      tfExec,
		clusterType: clusterType,
	}
}

func NewTerraformExecHelperWithWorkspaceName(clusterType constants.ClusterType, workspaceName string) (*TerraformExecHelper, error) {
	tfExec, err := NewTerraformExec(workspaceName)
	if err != nil {
		return nil, err
	}
	return NewTerraformExecHelper(clusterType, tfExec), nil
}

func (th *TerraformExecHelper) GetTerraformExec() TerraformExec {
	return th.tfExec
}

func (th *TerraformExecHelper) GetAccountRoleService() (*AccountRoleService, error) {
	return newAccountRoleService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetClusterService() (*ClusterService, error) {
	return newClusterService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetClusterAutoscalerService() (*ClusterAutoscalerService, error) {
	return newClusterAutoscalerService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetDnsDomainService() (*DnsService, error) {
	return newDnsDomainService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetIDPService(idpType constants.IDPType) (*IDPService, error) {
	return newIDPService(idpType, th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetImportService() (*ImportService, error) {
	return newImportService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetKMSService() (*KMSService, error) {
	return newKMSService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetKubeletConfigService() (*KubeletConfigService, error) {
	return newKubeletConfigService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetMachinePoolService() (*MachinePoolService, error) {
	return newMachinePoolService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetOIDCProviderOperatorRoleService() (*OIDCProviderOperatorRolesService, error) {
	return newOIDCProviderOperatorRoleService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetProxyService() (*ProxyService, error) {
	return newProxyService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetRhcsInfoService() (*RhcsInfoService, error) {
	return newRhcsInfoService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetSecurityGroupService() (*SecurityGroupService, error) {
	return newSecurityGroupService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetSharedVpcPolicyAndHostedZoneService() (*SharedVpcPolicyAndHostedZoneService, error) {
	return newSharedVPCPolicyAndHostedZoneService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetVPCTagService() (*VPCTagService, error) {
	return newVPCTagService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) GetVPCService() (*VPCService, error) {
	return newVPCService(th.clusterType, th.tfExec)
}

func (th *TerraformExecHelper) PrepareVPC(region string, multiZone bool, azIDs []string, name string, sharedVpcAWSSharedCredentialsFile string) (*VPCOutput, error) {
	vpcService, err := th.GetVPCService()
	if err != nil {
		return nil, err
	}
	vpcArgs := &VPCArgs{
		AWSRegion: region,
		MultiAZ:   multiZone,
		VPCCIDR:   constants.DefaultVPCCIDR,
		HCP:       th.clusterType.HCP,
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
	if name != "" {
		vpcArgs.Name = name
	}
	if sharedVpcAWSSharedCredentialsFile != "" {
		vpcArgs.AWSSharedCredentialsFiles = []string{sharedVpcAWSSharedCredentialsFile}
	}
	_, err = vpcService.Apply(vpcArgs)
	if err != nil {
		vpcService.Destroy(false)
		return nil, err
	}
	output, err := vpcService.Output()

	if err != nil {
		vpcService.Destroy(false)
		return nil, err
	}
	return output, err
}

func (th *TerraformExecHelper) PrepareAdditionalSecurityGroups(region string, vpcID string, sgNumbers int) ([]string, error) {
	sgService, err := th.GetSecurityGroupService()
	if err != nil {
		return nil, err
	}
	sgArgs := &SecurityGroupArgs{
		AWSRegion:  region,
		VPCID:      vpcID,
		SGNumber:   sgNumbers,
		NamePrefix: "rhcs-ci",
	}
	_, err = sgService.Apply(sgArgs)
	if err != nil {
		sgService.Destroy(false)
		return nil, err
	}
	output, err := sgService.Output()
	if err != nil {
		sgService.Destroy(false)
		return nil, err
	}
	return output.SGIDs, err
}

func (th *TerraformExecHelper) PrepareAccountRoles(token string, accountRolePrefix string, accountRolesPath string, awsRegion string, openshiftVersion string, channelGroup string, sharedVpcRoleArn string) (
	*AccountRolesOutput, error) {
	accService, err := th.GetAccountRoleService()
	if err != nil {
		return nil, err
	}
	args := &AccountRolesArgs{
		AccountRolePrefix:   accountRolePrefix,
		OpenshiftVersion:    openshiftVersion,
		ChannelGroup:        channelGroup,
		UnifiedAccRolesPath: accountRolesPath,
	}
	if sharedVpcRoleArn != "" {
		args.SharedVpcRoleArn = sharedVpcRoleArn
	}
	_, err = accService.Apply(args, true)
	if err != nil {
		accService.Destroy(false)
		return nil, err
	}

	output, err := accService.Output()
	if err != nil {
		accService.Destroy(false)
		return nil, err
	}
	return output, err
}

func (th *TerraformExecHelper) PrepareOIDCProviderAndOperatorRoles(token string, oidcConfigType string, operatorRolePrefix string, accountRolePrefix string, accountRolesPath string, awsRegion string) (
	*OIDCProviderOperatorRolesOutput, error) {
	oidcOpService, err := th.GetOIDCProviderOperatorRoleService()
	if err != nil {
		return nil, err
	}
	args := &OIDCProviderOperatorRolesArgs{
		AccountRolePrefix:   accountRolePrefix,
		OperatorRolePrefix:  operatorRolePrefix,
		OIDCConfig:          oidcConfigType,
		AWSRegion:           awsRegion,
		UnifiedAccRolesPath: accountRolesPath,
	}
	_, err = oidcOpService.Apply(args)
	if err != nil {
		oidcOpService.Destroy(false)
		return nil, err
	}
	output, err := oidcOpService.Output()
	if err != nil {
		oidcOpService.Destroy(false)
		return nil, err
	}
	return output, err
}

func (th *TerraformExecHelper) PrepareProxy(region string, VPCID string, subnetPublicID string) (*ProxyOutput, error) {
	proxyService, err := th.GetProxyService()
	if err != nil {
		return nil, err
	}
	proxyArgs := &ProxyArgs{
		Region:              region,
		VPCID:               VPCID,
		PublicSubnetID:      subnetPublicID,
		TrustBundleFilePath: path.Join(constants.RHCS.RhcsOutputDir, "ca.cert"),
	}

	_, err = proxyService.Apply(proxyArgs)
	if err != nil {
		proxyService.Destroy(false)
		return nil, err
	}
	proxyOutput, err := proxyService.Output()
	if err != nil {
		proxyService.Destroy(false)
		return nil, err
	}

	return proxyOutput, err
}

func (th *TerraformExecHelper) PrepareKMSKey(region string, kmsName string, accountRolePrefix string, accountRolePath string) (string, error) {
	kmsService, err := th.GetKMSService()
	if err != nil {
		return "", err
	}
	kmsArgs := &KMSArgs{
		KMSName:           kmsName,
		AWSRegion:         region,
		AccountRolePrefix: accountRolePrefix,
		AccountRolePath:   accountRolePath,
		TagKey:            "Purpose",
		TagValue:          "RHCS automation test",
		TagDescription:    "BYOK Test Key for API automation",
		HCP:               th.clusterType.HCP,
	}

	_, err = kmsService.Apply(kmsArgs)
	if err != nil {
		kmsService.Destroy(false)
		return "", err
	}
	kmsOutput, err := kmsService.Output()
	if err != nil {
		kmsService.Destroy(false)
		return "", err
	}

	return kmsOutput.KeyARN, err
}

func (th *TerraformExecHelper) PrepareRoute53() (string, error) {
	dnsService, err := th.GetDnsDomainService()
	if err != nil {
		return "", err
	}
	dnsArgs := &DnsDomainArgs{}

	_, err = dnsService.Apply(dnsArgs)
	if err != nil {
		dnsService.Destroy(false)
		return "", err
	}
	output, err := dnsService.Output()
	if err != nil {
		dnsService.Destroy(false)
		return "", err
	}

	return output.DnsDomainId, err
}

func (th *TerraformExecHelper) PrepareSharedVpcPolicyAndHostedZone(region string,
	shared_vpc_aws_shared_credentials_file string,
	cluster_name string,
	dns_domain_id string,
	ingress_operator_role_arn string,
	installer_role_arn string,
	cluster_aws_account string,
	vpc_id string,
	subnets []string) (*SharedVpcPolicyAndHostedZoneOutput, error) {

	sharedVPCService, err := th.GetSharedVpcPolicyAndHostedZoneService()
	if err != nil {
		return nil, err
	}

	sharedVPCArgs := &SharedVpcPolicyAndHostedZoneArgs{
		SharedVpcAWSSharedCredentialsFiles: []string{shared_vpc_aws_shared_credentials_file},
		Region:                             region,
		ClusterName:                        cluster_name,
		DnsDomainId:                        dns_domain_id,
		IngressOperatorRoleArn:             ingress_operator_role_arn,
		InstallerRoleArn:                   installer_role_arn,
		ClusterAWSAccount:                  cluster_aws_account,
		VpcId:                              vpc_id,
		Subnets:                            subnets,
	}

	_, err = sharedVPCService.Apply(sharedVPCArgs)
	if err != nil {
		sharedVPCService.Destroy(false)
		return nil, err
	}
	output, err := sharedVPCService.Output()
	if err != nil {
		sharedVPCService.Destroy(false)
		return nil, err
	}

	return output, err
}
