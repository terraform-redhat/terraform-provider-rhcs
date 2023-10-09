package cms

import (
	"fmt"

	client "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// RetrieveClusterDetail will retrieve cluster detailed information based on the clusterID
func RetrieveClusterDetail(connection *client.Connection, clusterID string) (*cmv1.ClusterGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().Send()
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

// RetrieveMachineTypeDetail will return the retrieve result of machine type detailed information
func RetrieveMachineTypeDetail(connection *client.Connection, machineTypeID string) (*client.Response, error) {
	return connection.Get().Path(fmt.Sprintf(machineTypeIDURL, machineTypeID)).Send()
}

func RetrieveIDP(connection *client.Connection, clusterID string, idpID string) (*cmv1.IdentityProviderGetResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().IdentityProvider(idpID).Get().Send()
}

func ListIDPs(connection *client.Connection, clusterID string) (*cmv1.IdentityProvidersListResponse, error) {
	return connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).IdentityProviders().List().Send()
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

func RetrieveUpgradePolicies(connection *client.Connection, clusterID string, upgradepolicyID string) (*cmv1.UpgradePolicyGetResponse, error) {
	resp, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().UpgradePolicy(upgradepolicyID).Get().Send()
	return resp, err
}
