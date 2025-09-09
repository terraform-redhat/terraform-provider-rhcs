package cms

import (
	"errors"
	"fmt"
	"math"
	"time"

	client "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// RetrieveClusterDetail will retrieve cluster detailed information based on the clusterID
func RetrieveClusterDetail(connection *client.Connection, clusterID string) (*cmv1.ClusterGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().Send()
}

// RetrieveClusterIngress will retrieve default ingress detail information based on the clusterID
func RetrieveClusterIngress(connection *client.Connection, clusterID string) (*cmv1.Ingress, error) {
	ListResp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Ingresses().List().Send()
	if err != nil {
		return nil, err
	}

	return ListResp.Items().Get(0), nil
}

// ListClusters will list the clusters
func ListClusters(connection *client.Connection, parameters ...map[string]interface{}) (response *cmv1.ClustersListResponse, err error) {
	request := connection.ClustersMgmt().V1().Clusters().List()
	for _, param := range parameters {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	response, err = request.Send()
	return
}

// ListClusterResources will retrieve cluster detailed information about its resources based on the clusterID
func ListClusterResources(connection *client.Connection, clusterID string) (*cmv1.ClusterResourcesGetResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Resources().Live().Get()
	return request.Send()
}

// RetrieveClusterCredentials will return the response of cluster credentials
func RetrieveClusterCredentials(connection *client.Connection, clusterID string) (*cmv1.CredentialsGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Credentials().Get().Send()
}

// ListClusterGroups will return cluster groups
func ListClusterGroups(connection *client.Connection, clusterID string) (*cmv1.GroupsListResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Groups().List().Send()
}

// RetrieveClusterGroupDetail will return cluster specified group information
func RetrieveClusterGroupDetail(connection *client.Connection, clusterID string, groupID string) (*cmv1.GroupGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Groups().Group(groupID).Get().Send()
}

func ListClusterGroupUsers(connection *client.Connection, clusterID string, groupID string) (*cmv1.UsersListResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Groups().Group(groupID).Users().List().Send()
}

func RetrieveClusterGroupUserDetail(connection *client.Connection, clusterID string, groupID string, userID string) (*cmv1.UserGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Groups().Group(groupID).Users().User(userID).Get().Send()
}

func ListClusterIDPs(connection *client.Connection, clusterID string) (*cmv1.IdentityProvidersListResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().List().Send()
}

func RetrieveClusterIDPDetail(connection *client.Connection, clusterID string, IDPID string) (*cmv1.IdentityProviderGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().IdentityProvider(IDPID).Get().Send()
}

func ListHtpasswdUsers(connection *client.Connection, clusterID string, IDPID string) (*cmv1.HTPasswdUsersListResponse, error) {
	return connection.ClustersMgmt().V1().
		Clusters().
		Cluster(clusterID).
		IdentityProviders().
		IdentityProvider(IDPID).
		HtpasswdUsers().
		List().
		Send()
}

// RetrieveClusterLogDetail return the log response based on parameter
func RetrieveClusterInstallLogDetail(connection *client.Connection, clusterID string,
	parameter ...map[string]interface{}) (*cmv1.LogGetResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Logs().Install().Get()
	if len(parameter) == 1 {
		for paramK, paramV := range parameter[0] {
			request = request.Parameter(paramK, paramV)
		}
	}
	return request.Send()
}

// RetrieveClusterUninstallLogDetail return the uninstall log response based on parameter
func RetrieveClusterUninstallLogDetail(connection *client.Connection, clusterID string,
	parameter ...map[string]interface{}) (*cmv1.LogGetResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Logs().Uninstall().Get()
	if len(parameter) == 1 {
		for paramK, paramV := range parameter[0] {
			request = request.Parameter(paramK, paramV)
		}
	}
	return request.Send()
}

func RetrieveClusterStatus(connection *client.Connection, clusterID string) (*cmv1.ClusterStatusGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Status().Get().Send()
}

// cloud_providers & regions

func ListCloudProviders(connection *client.Connection, params ...map[string]interface{}) (*cmv1.CloudProvidersListResponse, error) {
	request := connection.ClustersMgmt().V1().CloudProviders().List()
	if len(params) == 1 {
		for k, v := range params[0] {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

func RetrieveCloudProviderDetail(connection *client.Connection, providerID string) (*cmv1.CloudProviderGetResponse, error) {
	return connection.ClustersMgmt().V1().CloudProviders().CloudProvider(providerID).Get().Send()
}

// ListRegions list the regions of specified cloud providers
// If params passed, will add parameter to the request
func ListRegions(connection *client.Connection, providerID string, params ...map[string]interface{}) (*cmv1.CloudRegionsListResponse, error) {
	request := connection.ClustersMgmt().V1().CloudProviders().CloudProvider(providerID).Regions().List()
	if len(params) == 1 {
		for k, v := range params[0] {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

func RetrieveRegionDetail(connection *client.Connection, providerID string, regionID string) (*cmv1.CloudRegionGetResponse, error) {
	return connection.ClustersMgmt().V1().CloudProviders().CloudProvider(providerID).Regions().Region(regionID).Get().Send()
}

func ListAvailableRegions(connection *client.Connection, providerID string, body *cmv1.AWS) (
	*cmv1.AvailableRegionsSearchResponse, error) {
	return connection.ClustersMgmt().
		V1().CloudProviders().
		CloudProvider(providerID).
		AvailableRegions().
		Search().
		Body(body).
		Send()
}

// version
func ListVersions(connection *client.Connection, parameter ...map[string]interface{}) (resp *cmv1.VersionsListResponse, err error) {
	request := connection.ClustersMgmt().V1().Versions().List()
	if len(parameter) == 1 {
		for k, v := range parameter[0] {
			request = request.Parameter(k, v)
		}
	}
	resp, err = request.Send()
	return
}

func RetrieveVersionDetail(connection *client.Connection, versionID string) (*cmv1.VersionGetResponse, error) {
	return connection.ClustersMgmt().V1().Versions().Version(versionID).Get().Send()
}

// ListMachineTypes will list the machine types
func ListMachineTypes(connection *client.Connection, params ...map[string]interface{}) (*cmv1.MachineTypesListResponse, error) {
	request := connection.ClustersMgmt().V1().MachineTypes().List()
	if len(params) == 1 {
		for k, v := range params[0] {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

func DeleteMachinePool(connection *client.Connection, clusterID string, mpID string) (*cmv1.MachinePoolDeleteResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).MachinePools().MachinePool(mpID).Delete().Send()
}

func RetrieveIDP(connection *client.Connection, clusterID string, idpID string) (*cmv1.IdentityProviderGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().IdentityProvider(idpID).Get().Send()
}

func ListIDPs(connection *client.Connection, clusterID string) (*cmv1.IdentityProvidersListResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().List().Send()
}

func DeleteIDP(connection *client.Connection, clusterID string, idpID string) (*cmv1.IdentityProviderDeleteResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().IdentityProvider(idpID).Delete().Send()
}

func PatchIDP(connection *client.Connection, clusterID string, idpID string, patchBody *cmv1.IdentityProvider) (*cmv1.IdentityProviderUpdateResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().IdentityProvider(idpID).Update().Body(patchBody).Send()
}

func CreateClusterIDP(connection *client.Connection, clusterID string, body *cmv1.IdentityProvider) (*cmv1.IdentityProvidersAddResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().Add().Body(body).Send()
}

// RetrieveClusterCPUTotalByNodeRolesOS will return the physical cpu_total of the compute nodes of the cluster
func RetrieveClusterCPUTotalByNodeRolesOS(connection *client.Connection, clusterID string) (*cmv1.CPUTotalByNodeRolesOSMetricQueryGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).MetricQueries().CPUTotalByNodeRolesOS().Get().Send()
}

// RetrieveClusterSocketTotalByNodeRolesOS will return the physical socket_total of the compute nodes of the cluster
func RetrieveClusterSocketTotalByNodeRolesOS(connection *client.Connection, clusterID string) (*cmv1.SocketTotalByNodeRolesOSMetricQueryGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).MetricQueries().SocketTotalByNodeRolesOS().Get().Send()
}

func RetrieveDetailedIngressOfCluster(connection *client.Connection, clusterID string, ingressID string) (*cmv1.IngressGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Ingresses().Ingress(ingressID).Get().Send()
	return resp, err
}

// Cluster labels
func ListClusterExternalConfiguration(connection *client.Connection, clusterID string) (*cmv1.ExternalConfigurationGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).ExternalConfiguration().Get().Send()
	return resp, err
}

func ListClusterLabels(connection *client.Connection, clusterID string, parameter ...map[string]interface{}) (*cmv1.LabelsListResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).ExternalConfiguration().Labels().List()
	for _, param := range parameter {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

func RetrieveDetailedLabelOfCluster(connection *client.Connection, clusterID string, labelID string) (*cmv1.LabelGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).ExternalConfiguration().Labels().Label(labelID).Get().Send()
}

// Machine Pool related
func ListMachinePool(connection *client.Connection, clusterID string, params ...map[string]interface{}) (*cmv1.MachinePoolsListResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).MachinePools().List()
	for _, param := range params {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}
func RetrieveClusterMachinePool(connection *client.Connection, clusterID string, machinePoolID string) (*cmv1.MachinePool, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).MachinePools().MachinePool(machinePoolID).Get().Send()
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
func RetrieveClusterNodePool(connection *client.Connection, clusterID string, machinePoolID string) (*cmv1.NodePool, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).NodePools().NodePool(machinePoolID).Get().Send()
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
func CreateClusterAutoscaler(connection *client.Connection, clusterID string, body *cmv1.ClusterAutoscaler) (*cmv1.AutoscalerPostResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Autoscaler().Post().Request(body).Send()
	return resp, err
}

func PatchClusterAutoscaler(connection *client.Connection, clusterID string, body *cmv1.ClusterAutoscaler) (*cmv1.AutoscalerUpdateResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Autoscaler().Update().Body(body).Send()
	return resp, err
}

func DeleteClusterAutoscaler(connection *client.Connection, clusterID string) (*cmv1.AutoscalerDeleteResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Autoscaler().Delete().Send()
	return resp, err
}

func RetrieveClusterAutoscaler(connection *client.Connection, clusterID string) (*cmv1.AutoscalerGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Autoscaler().Get().Send()
	return resp, err
}

// Upgrade policies related
func ListUpgradePolicies(connection *client.Connection, clusterID string, params ...map[string]interface{}) (*cmv1.UpgradePoliciesListResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().List()
	for _, param := range params {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

func GetUpgradePolicyState(connection *client.Connection, clusterID string, upgradepolicyID string) (*cmv1.UpgradePolicyStateGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().UpgradePolicy(upgradepolicyID).State().Get().Send()
	return resp, err
}

func RetrieveUpgradePolicies(connection *client.Connection, clusterID string, upgradepolicyID string) (*cmv1.UpgradePolicyGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().UpgradePolicy(upgradepolicyID).Get().Send()
	return resp, err
}

// ControlPlane Upgrade policies related
func ListControlPlaneUpgradePolicies(connection *client.Connection, clusterID string, params ...map[string]interface{}) (*cmv1.ControlPlaneUpgradePoliciesListResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).ControlPlane().UpgradePolicies().List()
	for _, param := range params {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

func RetrieveControlPlaneUpgradePolicy(connection *client.Connection, clusterID string, upgradepolicyID string) (*cmv1.ControlPlaneUpgradePolicyGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).ControlPlane().UpgradePolicies().ControlPlaneUpgradePolicy(upgradepolicyID).Get().Send()
	return resp, err
}

// NodePool Upgrade policies related
func ListNodePoolUpgradePolicies(connection *client.Connection, clusterID string, npID string, params ...map[string]interface{}) (*cmv1.NodePoolUpgradePoliciesListResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).NodePools().NodePool(npID).UpgradePolicies().List()
	for _, param := range params {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

func RetrieveNodePoolUpgradePolicy(connection *client.Connection, clusterID string, npID string, upgradepolicyID string) (*cmv1.NodePoolUpgradePolicyGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).NodePools().NodePool(npID).UpgradePolicies().NodePoolUpgradePolicy(upgradepolicyID).Get().Send()
	return resp, err
}

// RetrieveCurrentAccount return the response of retrieve current account
func RetrieveCurrentAccount(connection *client.Connection, params ...map[string]interface{}) (resp *v1.CurrentAccountGetResponse, err error) {
	if len(params) > 1 {
		return nil, errors.New("only one parameter map is allowed")
	}
	resp, err = connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	return resp, err
}

// RetrieveKubeletConfig returns the kubeletconfig
func RetrieveKubeletConfig(connection *client.Connection, clusterID string) (*cmv1.KubeletConfig, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).KubeletConfig().Get().Send()
	return resp.Body(), err
}

// ListHCPKubeletConfig returns the kubeletconfig
func ListHCPKubeletConfigs(connection *client.Connection, clusterID string) ([]*cmv1.KubeletConfig, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).KubeletConfigs().List().Send()
	return resp.Items().Slice(), err
}

// RetrieveHCPKubeletConfig returns the kubeletconfig
func RetrieveHCPKubeletConfig(connection *client.Connection, clusterID string, kubeConfigID string) (*cmv1.KubeletConfig, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).KubeletConfigs().KubeletConfig(kubeConfigID).Get().Send()
	return resp.Body(), err
}

// RetrieveNodePool returns the nodePool detail of HCP
func RetrieveNodePool(connection *client.Connection, clusterID string, npID string, parameter ...map[string]interface{}) (*cmv1.NodePool, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).NodePools().NodePool(npID).Get()
	for _, param := range parameter {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	resp, err := request.Send()
	return resp.Body(), err
}

// RetrieveNodePool returns the nodePool detail of HCP
func ListNodePools(connection *client.Connection, clusterID string, parameter ...map[string]interface{}) ([]*cmv1.NodePool, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).NodePools().List()
	for _, param := range parameter {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	resp, err := request.Send()
	return resp.Items().Slice(), err
}

// Delete cluster
func DeleteCluster(connection *client.Connection, clusterID string, params ...map[string]interface{}) (*cmv1.ClusterDeleteResponse, error) {
	request := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Delete()
	for _, param := range params {
		for k, v := range param {
			request = request.Parameter(k, v)
		}
	}
	return request.Send()
}

// Wait for cluster deleted via OCM
func WaitClusterDeleted(connection *client.Connection, clusterID string, timeoutMinute ...int) error {
	timeout := 30 * time.Minute
	if len(timeoutMinute) == 1 {
		timeout = time.Duration(timeoutMinute[0]) * time.Minute
	}
	start := time.Now()
	for time.Since(start) < timeout {
		Logger.Infof("Waiting for the cluster %s deleted. Timeout after %d mins\n",
			clusterID, int(math.Ceil(timeout.Minutes()-time.Since(start).Minutes())))
		resp, _ := RetrieveClusterDetail(connection, clusterID)

		if resp.Status() != CON.HTTPOK && resp.Status() != CON.HTTPNotFound {
			err := fmt.Errorf(">>> [Error] Getting the cluster information meets error: %s", resp.Error().Reason())
			return err
		}

		if resp.Status() == CON.HTTPNotFound {
			Logger.Infof("OOH! The cluster  %s is deleted.\n", clusterID)
			return nil
		}

		time.Sleep(90 * time.Second)
	}

	err := fmt.Errorf(">>> [Error] Met timeout( %s minites) when wait for cluster deleted via OCM", timeout.String())
	return err
}

func RetrieveTuningConfig(connection *client.Connection, clusterID string, tcName string) (*cmv1.TuningConfigGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).TuningConfigs().TuningConfig(tcName).Get().Send()
}

func ListTuningConfigs(connection *client.Connection, clusterID string) (*cmv1.TuningConfigsListResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).TuningConfigs().List().Send()
}

func ListRegistryAllowlists(connection *client.Connection) (*cmv1.RegistryAllowlistsListResponse, error) {
	return connection.ClustersMgmt().V1().RegistryAllowlists().List().Send()
}

func RetrieveRegistryAllowlist(connection *client.Connection, allowlistID string) (*cmv1.RegistryAllowlistGetResponse, error) {
	return connection.ClustersMgmt().V1().RegistryAllowlists().RegistryAllowlist(allowlistID).Get().Send()
}

func RetrieveClusterImageMirror(connection *client.Connection, clusterID string, imageMirrorID string) (*cmv1.ImageMirror, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).ImageMirrors().ImageMirror(imageMirrorID).Get().Send()
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
