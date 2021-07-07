/*
Copyright (c) 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
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

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Creates an OpenShift managed cluster.",
		Schema: map[string]*schema.Schema{
			nameKey: {
				Description: "Name of the cluster.",
				Type:        schema.TypeString,
				Required:    true,
			},
			cloudProviderKey: {
				Description: "Cloud provider identifier, for example 'aws'.",
				Type:        schema.TypeString,
				Required:    true,
			},
			cloudRegionKey: {
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Type:        schema.TypeString,
				Required:    true,
			},
			propertiesKey: {
				Description:      "User defined properties.",
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: resourceClusterValidateProperties,
			},
			stateKey: {
				Description: "State of the cluster.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
			waitKey: {
				Description: "Wait till the cluster is ready.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Hour),
		},
	}
}

func resourceClusterCreate(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	// Get the connection:
	connection := config.(*sdk.Connection)

	// Check if the cluster already exists. If it does exist then we don't need to do anything
	// else.
	var cluster *cmv1.Cluster
	cluster, result = resourceClusterLookup(ctx, connection, data)
	if result.HasError() {
		return
	}

	// If the cluster doesn't exist yet then try to create it:
	if cluster == nil {
		cluster, result = resourceClusterRender(data)
		if result.HasError() {
			return
		}
		clustersResource := connection.ClustersMgmt().V1().Clusters()
		addResponse, err := clustersResource.Add().
			Body(cluster).
			SendContext(ctx)
		if err != nil {
			result = diag.FromErr(err)
			return
		}
		cluster = addResponse.Body()
		result = resourceClusterParse(cluster, data)
		if result.HasError() {
			return
		}
	}

	// Wait till the cluster is ready:
	wait := data.Get(waitKey).(bool)
	if wait && cluster.State() != cmv1.ClusterStateReady {
		clustersResource := connection.ClustersMgmt().V1().Clusters()
		clusterResource := clustersResource.Cluster(cluster.ID())
		clusterResource.Poll().
			Interval(1 * time.Minute).
			Predicate(func(getResponse *cmv1.ClusterGetResponse) bool {
				cluster = getResponse.Body()
				return cluster.State() == cmv1.ClusterStateReady
			}).
			StartContext(ctx)
	}

	// Copy the cluster data:
	result = resourceClusterParse(cluster, data)
	if result.HasError() {
		return
	}

	return
}

func resourceClusterRead(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	// Get the connection:
	connection := config.(*sdk.Connection)

	// Try to find the cluster:
	var cluster *cmv1.Cluster
	cluster, result = resourceClusterLookup(ctx, connection, data)
	if result.HasError() {
		return
	}

	// If there is no matching cluster the mark it for creation:
	if cluster == nil {
		data.SetId("")
		return
	}

	// Parse the cluster data:
	result = resourceClusterParse(cluster, data)
	if result.HasError() {
		return
	}

	return
}

func resourceClusterUpdate(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	return
}

func resourceClusterDelete(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	// Get the connection:
	connection := config.(*sdk.Connection)

	// Try to find the cluster. If it doesn't exist then we don't need to do anything else.
	cluster, result := resourceClusterLookup(ctx, connection, data)
	if result.HasError() || cluster == nil {
		return
	}

	// Send the request to delete the cluster:
	clusterID := cluster.ID()
	clusterResource := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID)
	deleteResponse, err := clusterResource.Delete().SendContext(ctx)
	if deleteResponse != nil && deleteResponse.Status() == http.StatusNotFound {
		return
	}
	if err != nil {
		result = append(result, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("can't delete cluster '%s'", clusterID),
			Detail:   err.Error(),
		})
		return
	}

	// Wait till the cluster doesn't exist:
	wait := data.Get(waitKey).(bool)
	if wait {
		clusterResource.Poll().
			Interval(1 * time.Minute).
			Predicate(func(getResponse *cmv1.ClusterGetResponse) bool {
				return getResponse.Status() == http.StatusNotFound
			}).
			StartContext(ctx)
	}

	return
}

// resourceClusterRender converts the internal representation of a cluster into corresponding SDK
// cluster object.
func resourceClusterRender(data *schema.ResourceData) (cluster *cmv1.Cluster,
	result diag.Diagnostics) {
	builder := cmv1.NewCluster()
	var value interface{}
	var ok bool
	value, ok = data.GetOk(nameKey)
	if ok {
		builder.Name(value.(string))
	}
	value, ok = data.GetOk(cloudProviderKey)
	if ok {
		builder.CloudProvider(cmv1.NewCloudProvider().ID(value.(string)))
	}
	value, ok = data.GetOk(cloudRegionKey)
	if ok {
		builder.Region(cmv1.NewCloudRegion().ID(value.(string)))
	}
	value, ok = data.GetOk(propertiesKey)
	if ok {
		builder.Properties(resourceClusterConvertProperties(value))
	}
	value, ok = data.GetOk(stateKey)
	if ok {
		builder.State(cmv1.ClusterState(value.(string)))
	}
	cluster, err := builder.Build()
	if err != nil {
		result = diag.FromErr(err)
	}
	return
}

// resourceClusterParse converts a SDK cluster into the internal representation.
func resourceClusterParse(cluster *cmv1.Cluster,
	data *schema.ResourceData) (result diag.Diagnostics) {
	data.SetId(cluster.ID())
	data.Set(nameKey, cluster.Name())
	data.Set(cloudProviderKey, cluster.CloudProvider().ID())
	data.Set(cloudRegionKey, cluster.Region().ID())
	data.Set(propertiesKey, cluster.Properties())
	data.Set(stateKey, string(cluster.State()))
	return
}

// resourceClusterValidateProperties checks that the given value is valid for the `properties`
// attribute of a cluster.
func resourceClusterValidateProperties(value interface{}, path cty.Path) diag.Diagnostics {
	var result diag.Diagnostics
	values, ok := value.(map[string]interface{})
	if !ok {
		result = append(result, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "properties should be a map of strings",
		})
		return result
	}
	for k, v := range values {
		_, ok := v.(string)
		if !ok {
			result = append(result, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "property value should be a string",
				AttributePath: cty.IndexStringPath(k),
			})
		}
	}
	return result
}

// resourceClusterConvertProperties converts the given value to the type require by the `properties`
// attribute of a cluster. This assumes that the value has already been validated and will panic if
// the value isn't compatible.
func resourceClusterConvertProperties(value interface{}) map[string]string {
	if value == nil {
		return nil
	}
	values := value.(map[string]interface{})
	result := map[string]string{}
	for k, v := range values {
		result[k] = v.(string)
	}
	return result
}

// resourceClusterLookup tries to find a cluster that matches the given identifier or name. Returns
// nil if no such cluster exists.
func resourceClusterLookup(ctx context.Context, connection *sdk.Connection,
	data *schema.ResourceData) (cluster *cmv1.Cluster, result diag.Diagnostics) {
	// First try to locate the cluster using the identifier:
	clusterID := data.Id()
	clustersResource := connection.ClustersMgmt().V1().Clusters()
	if clusterID != "" {
		clusterResource := clustersResource.Cluster(clusterID)
		getResponse, err := clusterResource.Get().SendContext(ctx)
		if err == nil {
			cluster = getResponse.Body()
			return
		}
		if getResponse.Status() != http.StatusNotFound {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't fetch cluster with identifier '%s'",
					clusterID,
				),
				Detail: err.Error(),
			})
			return
		}
	}

	// Try to locate the cluster using the name:
	clusterName := data.Get("name").(string)
	if clusterName != "" {
		listResponse, err := clustersResource.List().
			Search(fmt.Sprintf("name = '%s'", clusterName)).
			Size(1).
			SendContext(ctx)
		if err != nil {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't fetch clusters with name '%s'",
					clusterName,
				),
				Detail: err.Error(),
			})
			return
		}
		listTotal := listResponse.Total()
		listItems := listResponse.Items().Slice()
		if listTotal == 0 {
			return
		}
		if listTotal > 1 {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"cluster name '%s' is ambiguous, there are %d clusters "+
						"with that name",
					clusterName, listTotal,
				),
			})
		}
		cluster = listItems[0]
	}

	return
}
