package cms

***REMOVED***
***REMOVED***

	client "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

// RetrieveClusterDetail will retrieve cluster detailed information based on the clusterID
func RetrieveClusterDetail(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.ClusterGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

// ListClusters will list the clusters
func ListClusters(connection *client.Connection, parameters ...map[string]interface{}***REMOVED*** (response *cmv1.ClustersListResponse, err error***REMOVED*** {

	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.List(***REMOVED***
	for _, param := range parameters {
		for k, v := range param {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	response, err = request.Send(***REMOVED***
	return
}

// RetrieveClusterCredentials will return the response of cluster credentials
func RetrieveClusterCredentials(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.CredentialsGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Credentials(***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

// ListClusterGroups will return cluster groups
func ListClusterGroups(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.GroupsListResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Groups(***REMOVED***.List(***REMOVED***.Send(***REMOVED***
}

// RetrieveClusterGroupDetail will return cluster specified group information
func RetrieveClusterGroupDetail(connection *client.Connection, clusterID string, groupID string***REMOVED*** (*cmv1.GroupGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Groups(***REMOVED***.Group(groupID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

func ListClusterGroupUsers(connection *client.Connection, clusterID string, groupID string***REMOVED*** (*cmv1.UsersListResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Groups(***REMOVED***.Group(groupID***REMOVED***.Users(***REMOVED***.List(***REMOVED***.Send(***REMOVED***
}

func RetrieveClusterGroupUserDetail(connection *client.Connection, clusterID string, groupID string, userID string***REMOVED*** (*cmv1.UserGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Groups(***REMOVED***.Group(groupID***REMOVED***.Users(***REMOVED***.User(userID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

func ListClusterIDPs(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.IdentityProvidersListResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.IdentityProviders(***REMOVED***.List(***REMOVED***.Send(***REMOVED***
}

func RetrieveClusterIDPDetail(connection *client.Connection, clusterID string, IDPID string***REMOVED*** (*cmv1.IdentityProviderGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.IdentityProviders(***REMOVED***.IdentityProvider(IDPID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

func ListHtpasswdUsers(connection *client.Connection, clusterID string, IDPID string***REMOVED*** (*cmv1.HTPasswdUsersListResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.
		Clusters(***REMOVED***.
		Cluster(clusterID***REMOVED***.
		IdentityProviders(***REMOVED***.
		IdentityProvider(IDPID***REMOVED***.
		HtpasswdUsers(***REMOVED***.
		List(***REMOVED***.
		Send(***REMOVED***
}

// RetrieveClusterLogDetail return the log response based on parameter
func RetrieveClusterInstallLogDetail(connection *client.Connection, clusterID string,
	parameter ...map[string]interface{}***REMOVED*** (*cmv1.LogGetResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Logs(***REMOVED***.Install(***REMOVED***.Get(***REMOVED***
	if len(parameter***REMOVED*** == 1 {
		for paramK, paramV := range parameter[0] {
			request = request.Parameter(paramK, paramV***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}

// RetrieveClusterUninstallLogDetail return the uninstall log response based on parameter
func RetrieveClusterUninstallLogDetail(connection *client.Connection, clusterID string,
	parameter ...map[string]interface{}***REMOVED*** (*cmv1.LogGetResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Logs(***REMOVED***.Uninstall(***REMOVED***.Get(***REMOVED***
	if len(parameter***REMOVED*** == 1 {
		for paramK, paramV := range parameter[0] {
			request = request.Parameter(paramK, paramV***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}

func RetrieveClusterStatus(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.ClusterStatusGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Status(***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

// cloud_providers & regions

func ListCloudProviders(connection *client.Connection, params ...map[string]interface{}***REMOVED*** (*cmv1.CloudProvidersListResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.CloudProviders(***REMOVED***.List(***REMOVED***
	if len(params***REMOVED*** == 1 {
		for k, v := range params[0] {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}

func RetrieveCloudProviderDetail(connection *client.Connection, providerID string***REMOVED*** (*cmv1.CloudProviderGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.CloudProviders(***REMOVED***.CloudProvider(providerID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

// ListRegions list the regions of specified cloud providers
// If params passed, will add parameter to the request
func ListRegions(connection *client.Connection, providerID string, params ...map[string]interface{}***REMOVED*** (*cmv1.CloudRegionsListResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.CloudProviders(***REMOVED***.CloudProvider(providerID***REMOVED***.Regions(***REMOVED***.List(***REMOVED***
	if len(params***REMOVED*** == 1 {
		for k, v := range params[0] {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}

func RetrieveRegionDetail(connection *client.Connection, providerID string, regionID string***REMOVED*** (*cmv1.CloudRegionGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.CloudProviders(***REMOVED***.CloudProvider(providerID***REMOVED***.Regions(***REMOVED***.Region(regionID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

func ListAvailableRegions(connection *client.Connection, providerID string, body *cmv1.AWS***REMOVED*** (
	*cmv1.AvailableRegionsSearchResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.
		V1(***REMOVED***.CloudProviders(***REMOVED***.
		CloudProvider(providerID***REMOVED***.
		AvailableRegions(***REMOVED***.
		Search(***REMOVED***.
		Body(body***REMOVED***.
		Send(***REMOVED***
}

// version

func ListVersions(connection *client.Connection, parameter ...map[string]interface{}***REMOVED*** (resp *cmv1.VersionsListResponse, err error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***.List(***REMOVED***
	if len(parameter***REMOVED*** == 1 {
		for k, v := range parameter[0] {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	resp, err = request.Send(***REMOVED***
	return
}

func RetrieveVersionDetail(connection *client.Connection, versionID string***REMOVED*** (*cmv1.VersionGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***.Version(versionID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

// ListMachineTypes will list the machine types
func ListMachineTypes(connection *client.Connection, params ...map[string]interface{}***REMOVED*** (*cmv1.MachineTypesListResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.MachineTypes(***REMOVED***.List(***REMOVED***
	if len(params***REMOVED*** == 1 {
		for k, v := range params[0] {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}

// RetrieveMachineTypeDetail will return the retrieve result of machine type detailed information
func RetrieveMachineTypeDetail(connection *client.Connection, machineTypeID string***REMOVED*** (*client.Response, error***REMOVED*** {
	return connection.Get(***REMOVED***.Path(fmt.Sprintf(machineTypeIDURL, machineTypeID***REMOVED******REMOVED***.Send(***REMOVED***
}

func RetrieveIDP(connection *client.Connection, clusterID string, idpID string***REMOVED*** (*cmv1.IdentityProviderGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.IdentityProviders(***REMOVED***.IdentityProvider(idpID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

func ListIDPs(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.IdentityProvidersListResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.IdentityProviders(***REMOVED***.List(***REMOVED***.Send(***REMOVED***
}

// RetrieveClusterCPUTotalByNodeRolesOS will return the physical cpu_total of the compute nodes of the cluster
func RetrieveClusterCPUTotalByNodeRolesOS(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.CPUTotalByNodeRolesOSMetricQueryGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.MetricQueries(***REMOVED***.CPUTotalByNodeRolesOS(***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

// RetrieveClusterSocketTotalByNodeRolesOS will return the physical socket_total of the compute nodes of the cluster
func RetrieveClusterSocketTotalByNodeRolesOS(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.SocketTotalByNodeRolesOSMetricQueryGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.MetricQueries(***REMOVED***.SocketTotalByNodeRolesOS(***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

func RetrieveDetailedIngressOfCluster(connection *client.Connection, clusterID string, ingressID string***REMOVED*** (*cmv1.IngressGetResponse, error***REMOVED*** {
	resp, err := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.Ingresses(***REMOVED***.Ingress(ingressID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
	return resp, err
}

// Cluster labels
func ListClusterExternalConfiguration(connection *client.Connection, clusterID string***REMOVED*** (*cmv1.ExternalConfigurationGetResponse, error***REMOVED*** {
	resp, err := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.ExternalConfiguration(***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
	return resp, err
}

func ListClusterLabels(connection *client.Connection, clusterID string, parameter ...map[string]interface{}***REMOVED*** (*cmv1.LabelsListResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.ExternalConfiguration(***REMOVED***.Labels(***REMOVED***.List(***REMOVED***
	for _, param := range parameter {
		for k, v := range param {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}

func RetrieveDetailedLabelOfCluster(connection *client.Connection, clusterID string, labelID string***REMOVED*** (*cmv1.LabelGetResponse, error***REMOVED*** {
	return connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.ExternalConfiguration(***REMOVED***.Labels(***REMOVED***.Label(labelID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
}

// Machine Pool related
func ListMachinePool(connection *client.Connection, clusterID string, params ...map[string]interface{}***REMOVED*** (*cmv1.MachinePoolsListResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.MachinePools(***REMOVED***.List(***REMOVED***
	for _, param := range params {
		for k, v := range param {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}
func RetrieveClusterMachinePool(connection *client.Connection, clusterID string, machinePoolID string***REMOVED*** (*cmv1.MachinePoolGetResponse, error***REMOVED*** {
	resp, err := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.MachinePools(***REMOVED***.MachinePool(machinePoolID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
	return resp, err
}

// Upgrade policies related
func ListUpgradePolicies(connection *client.Connection, clusterID string, params ...map[string]interface{}***REMOVED*** (*cmv1.UpgradePoliciesListResponse, error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.UpgradePolicies(***REMOVED***.List(***REMOVED***
	for _, param := range params {
		for k, v := range param {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	return request.Send(***REMOVED***
}

func RetrieveUpgradePolicies(connection *client.Connection, clusterID string, upgradepolicyID string***REMOVED*** (*cmv1.UpgradePolicyGetResponse, error***REMOVED*** {
	resp, err := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***.UpgradePolicies(***REMOVED***.UpgradePolicy(upgradepolicyID***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
	return resp, err
}
