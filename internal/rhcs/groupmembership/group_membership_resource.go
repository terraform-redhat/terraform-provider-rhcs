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

package groupmembership

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
)

func ResourceGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupMembershipCreate,
		ReadContext:   resourceGroupMembershipRead,
		DeleteContext: resourceGroupMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: GroupMembershipFields(),
	}
}

func resourceGroupMembershipCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	// Get the cluster collection:
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()

	groupMembershipState := groupMembershipFromResourceData(resourceData)

	// Wait till the cluster is ready:
	if err := common.WaitTillClusterIsReadyOrFail(ctx, clusterCollection, groupMembershipState.Cluster); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't poll cluster state",
				Detail: fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					groupMembershipState.Cluster, err),
			}}
	}

	// Create the membership:
	builder := cmv1.NewUser()
	builder.ID(groupMembershipState.User)
	object, err := builder.Build()
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build group membership",
				Detail: fmt.Sprintf(
					"Can't build group membership for cluster '%s' and group '%s': %v",
					groupMembershipState.Cluster, groupMembershipState.Group, err),
			}}
	}

	groupCollection := clusterCollection.
		Cluster(groupMembershipState.Cluster).
		Groups().
		Group(groupMembershipState.Group).
		Users()
	add, err := groupCollection.Add().Body(object).SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't create group membership",
				Detail: fmt.Sprintf(
					"Can't create group membership for cluster '%s' and group '%s': %v",
					groupMembershipState.Cluster, groupMembershipState.Group, err),
			}}
	}

	object = add.Body()

	// Save the state:
	groupMembershipToResourceData(object, resourceData, groupMembershipState.Cluster, groupMembershipState.Group)
	return
}

func resourceGroupMembershipRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Read()")
	groupMembershipState := groupMembershipFromResourceData(resourceData)
	resource := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters().Cluster(groupMembershipState.Cluster).
		Groups().Group(groupMembershipState.Group).
		Users().User(groupMembershipState.User)

	get, err := resource.Get().SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't find group membership",
				Detail: fmt.Sprintf(
					"Can't find user group membership identifier '%s' for cluster '%s' and group '%s': %v",
					groupMembershipState.ID, groupMembershipState.Cluster, groupMembershipState.Group, err),
			}}
	}
	object := get.Body()

	// Save the state:
	groupMembershipToResourceData(object, resourceData, groupMembershipState.Cluster, groupMembershipState.Group)
	return
}

func resourceGroupMembershipDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin delete()")
	groupMembershipState := groupMembershipFromResourceData(resourceData)
	resource := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters().Cluster(groupMembershipState.Cluster).
		Groups().Group(groupMembershipState.Group).
		Users().User(groupMembershipState.User)

	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't delete group membership",
				Detail: fmt.Sprintf(
					"Can't delete group membership with identifier '%s' for cluster '%s' and group '%s': %v",
					groupMembershipState.ID, groupMembershipState.Cluster, groupMembershipState.Group, err),
			}}
	}

	// Remove the state:
	resourceData.SetId("")
	return
}

func groupMembershipFromResourceData(resourceData *schema.ResourceData) *GroupMembershipState {
	return &GroupMembershipState{
		Cluster: resourceData.Get("cluster").(string),
		Group:   resourceData.Get("group").(string),
		User:    resourceData.Get("user").(string),
		ID:      resourceData.Id(),
	}
}

func groupMembershipToResourceData(object *cmv1.User, resourceData *schema.ResourceData, cluster, group string) {
	resourceData.SetId(object.ID())
	resourceData.Set("user", object.ID())
	resourceData.Set("cluster", cluster)
	resourceData.Set("group", group)

}
