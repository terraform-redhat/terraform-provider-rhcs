/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ocm

***REMOVED***
	"context"
***REMOVED***
***REMOVED***
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

func resourceCluster(***REMOVED*** *schema.Resource {
	return &schema.Resource{
		Description: "Creates an OpenShift managed cluster.",
		Schema: map[string]*schema.Schema{
			nameKey: {
				Description: "Name of the cluster.",
				Type:        schema.TypeString,
				Required:    true,
	***REMOVED***,
			cloudProviderKey: {
				Description: "Cloud provider identifier, for example 'aws'.",
				Type:        schema.TypeString,
				Required:    true,
	***REMOVED***,
			cloudRegionKey: {
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Type:        schema.TypeString,
				Required:    true,
	***REMOVED***,
			propertiesKey: {
				Description:      "User defined properties.",
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: resourceClusterValidateProperties,
	***REMOVED***,
			stateKey: {
				Description: "State of the cluster.",
				Type:        schema.TypeMap,
				Computed:    true,
	***REMOVED***,
			waitKey: {
				Description: "Wait till the cluster is ready.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
	***REMOVED***,
***REMOVED***,
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Hour***REMOVED***,
***REMOVED***,
	}
}

func resourceClusterCreate(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Get the connection:
	connection := config.(*sdk.Connection***REMOVED***

	// Check if the cluster already exists. If it does exist then we don't need to do anything
	// else.
	var cluster *cmv1.Cluster
	cluster, result = resourceClusterLookup(ctx, connection, data***REMOVED***
	if result.HasError(***REMOVED*** {
		return
	}

	// If the cluster doesn't exist yet then try to create it:
	if cluster == nil {
		cluster, result = resourceClusterRender(data***REMOVED***
		if result.HasError(***REMOVED*** {
			return
***REMOVED***
		clustersResource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
		addResponse, err := clustersResource.Add(***REMOVED***.
			Body(cluster***REMOVED***.
			SendContext(ctx***REMOVED***
		if err != nil {
			result = diag.FromErr(err***REMOVED***
			return
***REMOVED***
		cluster = addResponse.Body(***REMOVED***
	}

	// Wait till the cluster is ready:
	wait := data.Get(waitKey***REMOVED***.(bool***REMOVED***
	if wait && cluster.State(***REMOVED*** != cmv1.ClusterStateReady {
		clustersResource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
		clusterResource := clustersResource.Cluster(cluster.ID(***REMOVED******REMOVED***
		clusterResource.Poll(***REMOVED***.
			Interval(1 * time.Minute***REMOVED***.
			Predicate(func(getResponse *cmv1.ClusterGetResponse***REMOVED*** bool {
				cluster = getResponse.Body(***REMOVED***
				return cluster.State(***REMOVED*** == cmv1.ClusterStateReady
	***REMOVED******REMOVED***.
			StartContext(ctx***REMOVED***
	}

	// Copy the cluster data:
	result = resourceClusterParse(cluster, data***REMOVED***
	return
}

func resourceClusterRead(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Get the connection:
	connection := config.(*sdk.Connection***REMOVED***

	// Try to find the cluster:
	var cluster *cmv1.Cluster
	cluster, result = resourceClusterLookup(ctx, connection, data***REMOVED***
	if result.HasError(***REMOVED*** {
		return
	}

	// If there is no matching cluster the mark it for creation:
	if cluster == nil {
		data.SetId(""***REMOVED***
		return
	}

	// Parse the cluster data:
	result = resourceClusterParse(cluster, data***REMOVED***
	if result.HasError(***REMOVED*** {
		return
	}

	return
}

func resourceClusterUpdate(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	return
}

func resourceClusterDelete(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Get the connection:
	connection := config.(*sdk.Connection***REMOVED***

	// Try to find the cluster. If it doesn't exist then we don't need to do anything else.
	cluster, result := resourceClusterLookup(ctx, connection, data***REMOVED***
	if result.HasError(***REMOVED*** || cluster == nil {
		return
	}

	// Send the request to delete the cluster:
	clusterID := cluster.ID(***REMOVED***
	clusterResource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***
	deleteResponse, err := clusterResource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if deleteResponse != nil && deleteResponse.Status(***REMOVED*** == http.StatusNotFound {
		return
	}
	if err != nil {
		result = append(result, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("can't delete cluster '%s'", clusterID***REMOVED***,
			Detail:   err.Error(***REMOVED***,
***REMOVED******REMOVED***
		return
	}

	// Wait till the cluster doesn't exist:
	wait := data.Get(waitKey***REMOVED***.(bool***REMOVED***
	if wait {
		clusterResource.Poll(***REMOVED***.
			Interval(1 * time.Minute***REMOVED***.
			Predicate(func(getResponse *cmv1.ClusterGetResponse***REMOVED*** bool {
				return getResponse.Status(***REMOVED*** == http.StatusNotFound
	***REMOVED******REMOVED***.
			StartContext(ctx***REMOVED***
	}

	return
}

// resourceClusterRender converts the internal representation of a cluster into corresponding SDK
// cluster object.
func resourceClusterRender(data *schema.ResourceData***REMOVED*** (cluster *cmv1.Cluster,
	result diag.Diagnostics***REMOVED*** {
	builder := cmv1.NewCluster(***REMOVED***
	var value interface{}
	var ok bool
	value, ok = data.GetOk(nameKey***REMOVED***
	if ok {
		builder.Name(value.(string***REMOVED******REMOVED***
	}
	value, ok = data.GetOk(cloudProviderKey***REMOVED***
	if ok {
		builder.CloudProvider(cmv1.NewCloudProvider(***REMOVED***.ID(value.(string***REMOVED******REMOVED******REMOVED***
	}
	value, ok = data.GetOk(cloudRegionKey***REMOVED***
	if ok {
		builder.Region(cmv1.NewCloudRegion(***REMOVED***.ID(value.(string***REMOVED******REMOVED******REMOVED***
	}
	value, ok = data.GetOk(propertiesKey***REMOVED***
	if ok {
		builder.Properties(resourceClusterConvertProperties(value***REMOVED******REMOVED***
	}
	value, ok = data.GetOk(stateKey***REMOVED***
	if ok {
		builder.State(cmv1.ClusterState(value.(string***REMOVED******REMOVED******REMOVED***
	}
	cluster, err := builder.Build(***REMOVED***
	if err != nil {
		result = diag.FromErr(err***REMOVED***
	}
	return
}

// resourceClusterParse converts a SDK cluster into the internal representation.
func resourceClusterParse(cluster *cmv1.Cluster,
	data *schema.ResourceData***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	data.SetId(cluster.ID(***REMOVED******REMOVED***
	data.Set(nameKey, cluster.Name(***REMOVED******REMOVED***
	data.Set(cloudProviderKey, cluster.CloudProvider(***REMOVED***.ID(***REMOVED******REMOVED***
	data.Set(cloudRegionKey, cluster.Region(***REMOVED***.ID(***REMOVED******REMOVED***
	data.Set(propertiesKey, cluster.Properties(***REMOVED******REMOVED***
	data.Set(stateKey, string(cluster.State(***REMOVED******REMOVED******REMOVED***
	return
}

// resourceClusterValidateProperties checks that the given value is valid for the `properties`
// attribute of a cluster.
func resourceClusterValidateProperties(value interface{}, path cty.Path***REMOVED*** diag.Diagnostics {
	var result diag.Diagnostics
	values, ok := value.(map[string]interface{}***REMOVED***
	if !ok {
		result = append(result, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "properties should be a map of strings",
***REMOVED******REMOVED***
		return result
	}
	for k, v := range values {
		_, ok := v.(string***REMOVED***
		if !ok {
			result = append(result, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "property value should be a string",
				AttributePath: cty.IndexStringPath(k***REMOVED***,
	***REMOVED******REMOVED***
***REMOVED***
	}
	return result
}

// resourceClusterConvertProperties converts the given value to the type require by the `properties`
// attribute of a cluster. This assumes that the value has already been validated and will panic if
// the value isn't compatible.
func resourceClusterConvertProperties(value interface{}***REMOVED*** map[string]string {
	if value == nil {
		return nil
	}
	values := value.(map[string]interface{}***REMOVED***
	result := map[string]string{}
	for k, v := range values {
		result[k] = v.(string***REMOVED***
	}
	return result
}

// resourceClusterLookup tries to find a cluster that matches the given identifier or name. Returns
// nil if no such cluster exists.
func resourceClusterLookup(ctx context.Context, connection *sdk.Connection,
	data *schema.ResourceData***REMOVED*** (cluster *cmv1.Cluster, result diag.Diagnostics***REMOVED*** {
	// First try to locate the cluster using the identifier:
	clusterID := data.Id(***REMOVED***
	clustersResource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
	if clusterID != "" {
		clusterResource := clustersResource.Cluster(clusterID***REMOVED***
		getResponse, err := clusterResource.Get(***REMOVED***.SendContext(ctx***REMOVED***
		if err == nil {
			cluster = getResponse.Body(***REMOVED***
			return
***REMOVED***
		if getResponse.Status(***REMOVED*** != http.StatusNotFound {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't fetch cluster with identifier '%s'",
					clusterID,
				***REMOVED***,
				Detail: err.Error(***REMOVED***,
	***REMOVED******REMOVED***
			return
***REMOVED***
	}

	// Try to locate the cluster using the name:
	value, ok := data.GetOk(nameKey***REMOVED***
	if ok {
		clusterName := value.(string***REMOVED***
		listResponse, err := clustersResource.List(***REMOVED***.
			Search(fmt.Sprintf("name = '%s'", clusterName***REMOVED******REMOVED***.
			Size(1***REMOVED***.
			SendContext(ctx***REMOVED***
		if err != nil {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't find clusters with name '%s'",
					clusterName,
				***REMOVED***,
				Detail: err.Error(***REMOVED***,
	***REMOVED******REMOVED***
			return
***REMOVED***
		listTotal := listResponse.Total(***REMOVED***
		listItems := listResponse.Items(***REMOVED***.Slice(***REMOVED***
		if listTotal == 0 {
			return
***REMOVED***
		if listTotal > 1 {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"cluster name '%s' is ambiguous, there are %d clusters "+
						"with that name",
					clusterName, listTotal,
				***REMOVED***,
	***REMOVED******REMOVED***
***REMOVED***
		cluster = listItems[0]
	}

	return
}
