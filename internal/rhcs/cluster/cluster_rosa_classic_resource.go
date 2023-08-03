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
package cluster

import (
	"context"
	"errors"
	"fmt"
	clusterschema2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/cluster/clusterschema"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/upgrade"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	semver "github.com/hashicorp/go-version"
	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocm_errors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift/rosa/pkg/ocm"
	"github.com/openshift/rosa/pkg/properties"

	"github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
)

var kmsArnRE = regexp.MustCompile(
	`^arn:aws[\w-]*:kms:[\w-]+:\d{12}:key\/mrk-[0-9a-f]{32}$|[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

func ResourceClusterRosaClassic() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterRosaClassicCreate,
		ReadContext:   resourceClusterRosaClassicRead,
		UpdateContext: resourceClusterRosaClassicUpdate,
		DeleteContext: resourceClusterRosaClassicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: clusterschema2.ClusterRosaClassicFields(),
	}
}

func resourceClusterRosaClassicCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	// Get the cluster collection:
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	// Get the version collection:
	versionCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Versions()
	err := validateAttributes(resourceData)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Failed in validation",
				Detail:   err.Error(),
			}}
	}

	clusterRosaState := clusterRosaFromResourceData(resourceData)

	summary := "Can't build cluster"
	if clusterRosaState.Version != nil && strings.HasPrefix(*clusterRosaState.Version, "openshift-v") {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  summary,
				Detail:   "Openshift version must be provided without the \"openshift-v\" prefix",
			}}
	}

	version, err := getAndValidateVersionInChannelGroup(ctx, clusterRosaState, versionCollection)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"Can't build cluster with name '%s': %v",
					clusterRosaState.Name, err),
			}}
	}

	err = validateHttpTokensVersion(ctx, clusterRosaState, version)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"Can't build cluster with name '%s': %v",
					clusterRosaState.Name, err),
			}}
	}

	object, err := createClassicRosaClusterObject(ctx, clusterRosaState, diags)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"Can't build cluster with name '%s': %v",
					clusterRosaState.Name, err),
			}}
	}

	add, err := clusterCollection.Add().Body(object).SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"Can't create cluster with name '%s': %v",
					clusterRosaState.Name, err),
			}}
	}
	object = add.Body()
	// Save the state:
	clusterRosaToResourceData(ctx, object, resourceData)
	return
}

func validateAttributes(resourceData *schema.ResourceData) error {
	if proxy, ok := resourceData.GetOk("proxy"); ok {
		err := common2.ValidAllDiag(common2.ListOfMapValidator, clusterschema2.ProxyValidators)(proxy)
		if err != nil {
			return err
		}
	}
	if adminCreds, ok := resourceData.GetOk("admin_credentials"); ok {
		err := common2.ValidAllDiag(common2.ListOfMapValidator, clusterschema2.AdminCredsValidators)(adminCreds)
		if err != nil {
			return err
		}
	}
	return nil
}

func clusterRosaToResourceData(ctx context.Context, object *cmv1.Cluster, resourceData *schema.ResourceData) {
	resourceData.SetId(object.ID())

	resourceData.Set("external_id", object.ExternalID())
	resourceData.Set("name", object.Name())
	resourceData.Set("state", object.State())
	resourceData.Set("cloud_region", object.Region().ID())
	resourceData.Set("multi_az", object.MultiAZ())
	resourceData.Set("api_url", object.API().URL())
	resourceData.Set("console_url", object.Console().URL())
	resourceData.Set("domain", fmt.Sprintf("%s.%s", object.Name(), object.DNS().BaseDomain()))
	resourceData.Set("compute_machine_type", object.Nodes().ComputeMachineType().ID())
	resourceData.Set("ccs_enabled", object.CCS().Enabled())
	resourceData.Set("etcd_encryption", object.EtcdEncryption())
	resourceData.Set("proxy", clusterschema2.FlatProxy(object, resourceData))
	resourceData.Set("sts", clusterschema2.FlatSts(object))
	resourceData.Set("admin_credentials", clusterschema2.FlatAdminCredentials(object))

	if props, ok := object.GetProperties(); ok {
		propertiesMap := map[string]string{}
		ocmPropertiesMap := map[string]string{}
		for k, v := range props {
			ocmPropertiesMap[k] = v
			if _, isDefault := clusterschema2.OCMProperties[k]; !isDefault {
				propertiesMap[k] = v
			}
		}

		resourceData.Set("properties", propertiesMap)
		resourceData.Set("ocm_properties", ocmPropertiesMap)
	}

	// Note: The API does not currently return account id, but we try to get it
	// anyway. Failing that, we fetch the creator ARN from the properties like
	// rosa cli does.
	awsAccountID, ok := object.AWS().GetAccountID()
	if ok {
		resourceData.Set("aws_account_id", awsAccountID)
	} else {
		// rosa cli gets it from the properties, so we do the same
		if creatorARN, ok := object.Properties()[properties.CreatorARN]; ok {
			if arn, err := arn.Parse(creatorARN); err == nil {
				resourceData.Set("aws_account_id", arn.AccountID)
			}
		}

	}

	disableUserWorkload, ok := object.GetDisableUserWorkloadMonitoring()
	if ok && disableUserWorkload {
		resourceData.Set("disable_workload_monitoring", true)
	}

	isFips, ok := object.GetFIPS()
	if ok && isFips {
		resourceData.Set("fips", true)
	}

	disableSCPChecks, ok := object.CCS().GetDisableSCPChecks()
	if ok && disableSCPChecks {
		resourceData.Set("disable_scp_checks", true)
	}

	awsPrivateLink, ok := object.AWS().GetPrivateLink()
	if ok {
		resourceData.Set("aws_private_link", awsPrivateLink)
	}

	listeningMethod, ok := object.API().GetListening()
	if ok {
		resourceData.Set("private", listeningMethod == cmv1.ListeningMethodInternal)
	}

	kmsKeyArn, ok := object.AWS().GetKMSKeyArn()
	if ok {
		resourceData.Set("kms_key_arn", kmsKeyArn)
	}

	httpTokensState, ok := object.AWS().GetEc2MetadataHttpTokens()
	if ok && httpTokensState != "" {
		resourceData.Set("ec2_metadata_http_tokens", string(httpTokensState))
	}

	machineCIDR, ok := object.Network().GetMachineCIDR()
	if ok {
		resourceData.Set("machine_cidr", machineCIDR)
	}

	serviceCIDR, ok := object.Network().GetServiceCIDR()
	if ok {
		resourceData.Set("service_cidr", serviceCIDR)
	}

	podCIDR, ok := object.Network().GetPodCIDR()
	if ok {
		resourceData.Set("pod_cidr", podCIDR)
	}

	hostPrefix, ok := object.Network().GetHostPrefix()
	if ok {
		resourceData.Set("host_prefix", hostPrefix)
	}

	channelGroup, ok := object.Version().GetChannelGroup()
	if ok {
		resourceData.Set("channel_group", channelGroup)
	}

	version, ok := object.Version().GetID()
	if ok {
		// If we're using a non-default channel group, it will have been appended to
		// the version ID. Remove it before saving state.
		version = strings.TrimSuffix(version, fmt.Sprintf("-%s", channelGroup))
		version = strings.TrimPrefix(version, "openshift-v")
		tflog.Debug(ctx, fmt.Sprintf("actual cluster version: %v", version))
		resourceData.Set("current_version", version)

	}

	if subnetIds, ok := object.AWS().GetSubnetIDs(); ok && len(subnetIds) > 0 {
		resourceData.Set("aws_subnet_ids", subnetIds)
	}

	if azs, ok := object.Nodes().GetAvailabilityZones(); ok && len(azs) > 0 {
		resourceData.Set("availability_zones", azs)
	}

	if labels, ok := object.Nodes().GetComputeLabels(); ok {
		resourceData.Set("default_mp_labels", labels)
	}
}

func clusterRosaFromResourceData(resourceData *schema.ResourceData) *clusterschema2.ClusterRosaClassicState {
	result := &clusterschema2.ClusterRosaClassicState{
		Name:         resourceData.Get("name").(string),
		CloudRegion:  resourceData.Get("cloud_region").(string),
		AWSAccountID: resourceData.Get("aws_account_id").(string),
		ID:           resourceData.Id(),
	}

	// optional attributes
	result.ExternalID = common2.GetOptionalString(resourceData, "external_id")
	result.MultiAZ = common2.GetOptionalBool(resourceData, "multi_az")
	result.DisableSCPChecks = common2.GetOptionalBool(resourceData, "disable_scp_checks")
	result.DisableWorkloadMonitoring = common2.GetOptionalBool(resourceData, "disable_workload_monitoring")
	result.Sts = clusterschema2.ExpandStsFromResourceData(resourceData)
	result.Properties = common2.GetOptionalMapStringFromResourceData(resourceData, "properties")
	result.Tags = common2.GetOptionalMapStringFromResourceData(resourceData, "tags")
	result.Replicas = common2.GetOptionalInt(resourceData, "replicas")
	result.ComputeMachineType = common2.GetOptionalString(resourceData, "compute_machine_type")
	result.EtcdEncryption = common2.GetOptionalBool(resourceData, "etcd_encryption")
	result.AutoScalingEnabled = common2.GetOptionalBool(resourceData, "autoscaling_enabled")
	result.MinReplicas = common2.GetOptionalInt(resourceData, "min_replicas")
	result.MaxReplicas = common2.GetOptionalInt(resourceData, "max_replicas")
	result.AWSSubnetIDs = common2.GetOptionalListOfValueStringsFromResourceData(resourceData, "aws_subnet_ids")
	result.AWSPrivateLink = common2.GetOptionalBool(resourceData, "aws_private_link")
	result.Private = common2.GetOptionalBool(resourceData, "private")
	result.AvailabilityZones = common2.GetOptionalListOfValueStringsFromResourceData(resourceData, "availability_zones")
	result.MachineCIDR = common2.GetOptionalString(resourceData, "machine_cidr")
	result.Proxy = clusterschema2.ExpandProxyFromResourceData(resourceData)
	result.ServiceCIDR = common2.GetOptionalString(resourceData, "service_cidr")
	result.PodCIDR = common2.GetOptionalString(resourceData, "pod_cidr")
	result.HostPrefix = common2.GetOptionalInt(resourceData, "host_prefix")
	result.Version = common2.GetOptionalString(resourceData, "version")
	result.DefaultMPLabels = common2.GetOptionalMapStringFromResourceData(resourceData, "default_mp_labels")
	result.KMSKeyArn = common2.GetOptionalString(resourceData, "kms_key_arn")
	result.FIPS = common2.GetOptionalBool(resourceData, "fips")
	result.ChannelGroup = common2.GetOptionalString(resourceData, "channel_group")
	result.DisableWaitingInDestroy = common2.GetOptionalBool(resourceData, "disable_waiting_in_destroy")
	result.DestroyTimeout = common2.GetOptionalInt(resourceData, "destroy_timeout")
	result.Ec2MetadataHttpTokens = common2.GetOptionalString(resourceData, "ec2_metadata_http_tokens")
	result.UpgradeAcksFor = common2.GetOptionalString(resourceData, "upgrade_acknowledgements_for")
	result.AdminCredentials = clusterschema2.ExpandAdminCredentialsFromResourceData(resourceData)

	if state := common2.GetOptionalString(resourceData, "state"); state != nil {
		result.State = *state
	}
	if ccsEnabled := common2.GetOptionalBool(resourceData, "ccs_enabled"); ccsEnabled != nil {
		result.CCSEnabled = *ccsEnabled
	}
	if currentVersion := common2.GetOptionalString(resourceData, "current_version"); currentVersion != nil {
		result.CurrentVersion = *currentVersion
	}
	if apiURL := common2.GetOptionalString(resourceData, "api_url"); apiURL != nil {
		result.APIURL = *apiURL
	}
	if consoleURL := common2.GetOptionalString(resourceData, "console_url"); consoleURL != nil {
		result.ConsoleURL = *consoleURL
	}
	if domain := common2.GetOptionalString(resourceData, "domain"); domain != nil {
		result.Domain = *domain
	}
	result.OCMProperties = common2.GetOptionalMapStringFromResourceData(resourceData, "ocm_properties")

	return result
}

func resourceClusterRosaClassicRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Read()")

	// Get the cluster collection:
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()

	// Find the cluster:
	get, err := clusterCollection.Cluster(resourceData.Id()).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		resourceID := resourceData.Id()
		summary := fmt.Sprintf("cluster (%s) not found, removing from state", resourceID)
		tflog.Warn(ctx, summary)
		resourceData.SetId("")
		return []diag.Diagnostic{
			{
				Severity: diag.Warning,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"cluster (%s) not found, removing from state",
					resourceID),
			}}
	} else if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't find cluster",
				Detail: fmt.Sprintf(
					"Can't find cluster with identifier '%s': %v",
					resourceData.Id(), err),
			}}
	}
	object := get.Body()

	// Save the state:
	clusterRosaToResourceData(ctx, object, resourceData)
	return
}

func resourceClusterRosaClassicUpdate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin update()")

	// Get the cluster collection:
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	// Get the version collection:
	versionCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Versions()
	err := validateAttributes(resourceData)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Failed in validation",
				Detail:   err.Error(),
			}}
	}

	clusterRosaState := clusterRosaFromResourceData(resourceData)
	// Schedule a cluster upgrade if a newer version is requested
	if err := upgradeClusterIfNeeded(ctx, resourceData, clusterRosaState, clusterCollection, versionCollection); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't upgrade cluster",
				Detail: fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, %v",
					resourceData.Id(), err),
			}}
	}

	clusterBuilder := cmv1.NewCluster()

	clusterBuilder, _, err = updateNodes(resourceData, clusterRosaState, clusterBuilder)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't update cluster",
				Detail: fmt.Sprintf("Can't update cluster nodes for cluster with identifier: `%s`, %v",
					resourceData.Id(), err),
			}}
	}

	clusterBuilder, _, err = updateProxy(resourceData, clusterRosaState, clusterBuilder)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't update cluster",
				Detail: fmt.Sprintf("Can't update proxy's configuration for cluster with identifier: `%s`, %v",
					resourceData.Id(), err),
			}}
	}

	if resourceData.HasChange("disable_workload_monitoring") {
		_, newV := resourceData.GetChange("disable_workload_monitoring")
		newValue := common2.GetOptionalInterfaceBool(newV)
		clusterBuilder.DisableUserWorkloadMonitoring(newValue)
	}

	if resourceData.HasChange("properties") {
		propertiesMap := make(map[string]string)
		_, newV := resourceData.GetChange("properties")
		if newV != nil && len(newV.(map[string]interface{})) > 0 {
			propertiesMap = common2.ExpandStringValueMap(newV.(map[string]interface{}))
		}

		// set the default properties anyway
		for k, v := range clusterschema2.OCMProperties {
			propertiesMap[k] = v
		}
		clusterBuilder.Properties(propertiesMap)
	}

	clusterSpec, err := clusterBuilder.Build()
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build cluster patch",
				Detail: fmt.Sprintf("Can't build patch for cluster with identifier '%s': %v",
					resourceData.Id(), err),
			}}
	}

	update, err := clusterCollection.Cluster(resourceData.Id()).Update().
		Body(clusterSpec).
		SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't update cluster",
				Detail: fmt.Sprintf("Can't update cluster with identifier '%s': %v",
					resourceData.Id(), err),
			}}
	}

	object := update.Body()

	// Update the state:
	clusterRosaToResourceData(ctx, object, resourceData)
	return
}
func resourceClusterRosaClassicDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin delete()")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()

	// Send the request to delete the cluster:
	resource := clusterCollection.Cluster(resourceData.Id())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		return diag.Errorf(
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				resourceData.Id(), err,
			))
	}
	clusterRosaState := clusterRosaFromResourceData(resourceData)
	if clusterRosaState.DisableWaitingInDestroy != nil && *clusterRosaState.DisableWaitingInDestroy {
		tflog.Info(ctx, "Waiting for destroy to be completed, is disabled")
	} else {
		timeout := clusterschema2.DefaultTimeoutInMinutes
		if clusterRosaState.DestroyTimeout != nil {
			if *clusterRosaState.DestroyTimeout <= 0 {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  NonPositiveTimeoutSummary,
					Detail:   fmt.Sprintf(NonPositiveTimeoutFormat, resourceData.Id()),
				})
			} else {
				timeout = *clusterRosaState.DestroyTimeout
			}
		}
		isNotFound, err := retryClusterNotFoundWithTimeout(3, 1*time.Minute, ctx, timeout, resource)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Can't poll cluster state",
				Detail:   fmt.Sprintf("Can't poll state of cluster with identifier '%s': %v", resourceData.Id(), err),
			})
			return
		}

		if !isNotFound {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Cluster wasn't deleted yet",
				Detail:   fmt.Sprintf("The cluster with identifier '%s' is not deleted yet, but the polling finisehd due to a timeout", resourceData.Id()),
			})
		}
	}

	resourceData.SetId("")
	return
}

func createClassicRosaClusterObject(ctx context.Context,
	state *clusterschema2.ClusterRosaClassicState, diags diag.Diagnostics) (*cmv1.Cluster, error) {

	ocmClusterResource := resource.NewCluster()
	builder := ocmClusterResource.GetClusterBuilder()
	clusterName := state.Name
	if len(clusterName) > clusterschema2.MaxClusterNameLength {
		errDescription := fmt.Sprintf("Expected a valid value for 'name' maximum of 15 characters in length. Provided Cluster name '%s' is of length '%d'",
			clusterName, len(clusterName),
		)
		tflog.Error(ctx, errDescription)
		return nil, errors.New(clusterschema2.ErrHeadline + "\n" + errDescription)
	}

	builder.Name(state.Name)
	builder.CloudProvider(cmv1.NewCloudProvider().ID(clusterschema2.AwsCloudProvider))
	builder.Product(cmv1.NewProduct().ID(clusterschema2.RosaProduct))
	builder.Region(cmv1.NewCloudRegion().ID(state.CloudRegion))

	multiAZ := common2.Bool(state.MultiAZ)
	builder.MultiAZ(multiAZ)

	propertiesMap := make(map[string]string)
	if state.Properties != nil {
		propertiesMap = state.Properties
	}
	// Set default properties
	for k, v := range clusterschema2.OCMProperties {
		propertiesMap[k] = v
	}
	builder.Properties(propertiesMap)

	if state.EtcdEncryption != nil {
		builder.EtcdEncryption(*state.EtcdEncryption)
	}

	if state.ExternalID != nil {
		builder.ExternalID(*state.ExternalID)
	}

	if state.DisableWorkloadMonitoring != nil {
		builder.DisableUserWorkloadMonitoring(*state.DisableWorkloadMonitoring)
	}

	autoScalingEnabled := common2.Bool(state.AutoScalingEnabled)

	if err := ocmClusterResource.CreateNodes(autoScalingEnabled, state.Replicas, state.MinReplicas, state.MaxReplicas,
		state.ComputeMachineType, state.DefaultMPLabels, state.AvailabilityZones, multiAZ); err != nil {
		return nil, err
	}

	// ccs should be enabled in ocm rosa clusters
	ccs := cmv1.NewCCS()
	ccs.Enabled(true)

	if state.DisableSCPChecks != nil && *state.DisableSCPChecks {
		ccs.DisableSCPChecks(true)
	}
	builder.CCS(ccs)

	isPrivateLink := common2.Bool(state.AWSPrivateLink)
	isPrivate := common2.Bool(state.Private)
	var stsBuilder *cmv1.STSBuilder
	if state.Sts != nil {
		stsBuilder = resource.CreateSTS(state.Sts.RoleARN, state.Sts.SupportRoleArn,
			state.Sts.InstanceIAMRoles.MasterRoleARN, state.Sts.InstanceIAMRoles.WorkerRoleARN,
			state.Sts.OperatorRolePrefix, state.Sts.OIDCConfigID)
	}

	if err := ocmClusterResource.CreateAWSBuilder(state.Tags, state.Ec2MetadataHttpTokens, state.KMSKeyArn,
		isPrivateLink, state.AWSAccountID, stsBuilder, state.AWSSubnetIDs); err != nil {
		return nil, err
	}

	if err := ocmClusterResource.SetAPIPrivacy(isPrivate, isPrivateLink, stsBuilder != nil); err != nil {
		return nil, err
	}

	if state.FIPS != nil && *state.FIPS {
		builder.FIPS(true)
	}

	network := cmv1.NewNetwork()
	if state.MachineCIDR != nil {
		network.MachineCIDR(*state.MachineCIDR)
	}
	if state.ServiceCIDR != nil {
		network.ServiceCIDR(*state.ServiceCIDR)
	}
	if state.PodCIDR != nil {
		network.PodCIDR(*state.PodCIDR)
	}
	if state.HostPrefix != nil {
		network.HostPrefix(*state.HostPrefix)
	}
	if !network.Empty() {
		builder.Network(network)
	}

	channelGroup := ocm.DefaultChannelGroup
	if state.ChannelGroup != nil {
		channelGroup = *state.ChannelGroup
	}

	if state.Version != nil {
		// TODO: update it to support all cluster versions
		isSupported, err := common2.IsGreaterThanOrEqual(*state.Version, clusterschema2.MinVersion)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error validating required cluster version %s", err))
			errDescription := fmt.Sprintf(
				"Can't check if cluster version is supported '%s': %v",
				*state.Version, err,
			)
			return nil, errors.New(clusterschema2.ErrHeadline + "\n" + errDescription)
		}
		if !isSupported {
			description := fmt.Sprintf("Cluster version %s is not supported (minimal supported version is %s)", *state.Version, clusterschema2.MinVersion)
			tflog.Error(ctx, description)
			return nil, errors.New(clusterschema2.ErrHeadline + "\n" + description)
		}
		vBuilder := cmv1.NewVersion()
		versionID := fmt.Sprintf("openshift-v%s", *state.Version)
		// When using a channel group other than the default, the channel name
		// must be appended to the version ID or the API server will return an
		// error stating unexpected channel group.
		if channelGroup != ocm.DefaultChannelGroup {
			versionID = versionID + "-" + channelGroup
		}
		vBuilder.ID(versionID)
		vBuilder.ChannelGroup(channelGroup)
		builder.Version(vBuilder)
	}

	if state.AdminCredentials != nil {
		htpasswdUsers := []*cmv1.HTPasswdUserBuilder{}
		htpasswdUsers = append(htpasswdUsers, cmv1.NewHTPasswdUser().
			Username(state.AdminCredentials.Username).Password(state.AdminCredentials.Password))
		htpassUserList := cmv1.NewHTPasswdUserList().Items(htpasswdUsers...)
		htPasswdIDP := cmv1.NewHTPasswdIdentityProvider().Users(htpassUserList)
		builder.Htpasswd(htPasswdIDP)
	}

	builder, err := buildProxy(state, builder)
	if err != nil {
		tflog.Error(ctx, "Failed to build the Proxy's attributes")
		return nil, err
	}

	object, err := builder.Build()
	return object, err
}

func buildProxy(state *clusterschema2.ClusterRosaClassicState, builder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, error) {
	proxy := cmv1.NewProxy()
	if state.Proxy != nil {
		httpsProxy := ""
		httpProxy := ""
		additionalTrustBundle := ""

		if !common2.IsStringAttributeEmpty(state.Proxy.HttpProxy) {
			httpProxy = *state.Proxy.HttpProxy
			proxy.HTTPProxy(httpProxy)
		}
		if !common2.IsStringAttributeEmpty(state.Proxy.HttpsProxy) {
			httpsProxy = *state.Proxy.HttpsProxy
			proxy.HTTPSProxy(httpsProxy)
		}
		if !common2.IsStringAttributeEmpty(state.Proxy.NoProxy) {
			proxy.NoProxy(*state.Proxy.NoProxy)
		}

		if !common2.IsStringAttributeEmpty(state.Proxy.AdditionalTrustBundle) {
			additionalTrustBundle = *state.Proxy.AdditionalTrustBundle
			builder.AdditionalTrustBundle(additionalTrustBundle)
		}

		builder.Proxy(proxy)
	}

	return builder, nil
}

// getAndValidateVersionInChannelGroup ensures that the cluster version is
// available in the channel group
func getAndValidateVersionInChannelGroup(ctx context.Context, state *clusterschema2.ClusterRosaClassicState, versionCollection *cmv1.VersionsClient) (string, error) {
	channelGroup := ocm.DefaultChannelGroup
	if state.ChannelGroup != nil {
		channelGroup = *state.ChannelGroup
	}

	versionList, err := getVersionList(ctx, channelGroup, versionCollection)
	if err != nil {
		return "", err
	}

	version := versionList[0]
	if state.Version != nil && *state.Version != "" {
		version = *state.Version
	}

	tflog.Debug(ctx, fmt.Sprintf("Validating if cluster version %s is in the list of supported versions: %v", version, versionList))
	for _, v := range versionList {
		if v == version {
			return version, nil
		}
	}

	return "", fmt.Errorf("version %s is not in the list of supported versions: %v", version, versionList)
}

func validateHttpTokensVersion(ctx context.Context, state *clusterschema2.ClusterRosaClassicState, version string) error {
	if common2.IsStringAttributeEmpty(state.Ec2MetadataHttpTokens) ||
		cmv1.Ec2MetadataHttpTokens(*state.Ec2MetadataHttpTokens) == cmv1.Ec2MetadataHttpTokensOptional {
		return nil
	}

	greater, err := common2.IsGreaterThanOrEqual(version, clusterschema2.LowestHttpTokensVer)
	if err != nil {
		return fmt.Errorf("version '%s' is not supported: %v", version, err)
	}
	if !greater {
		msg := fmt.Sprintf("version '%s' is not supported with ec2_metadata_http_tokens, "+
			"minimum supported version is %s", version, clusterschema2.LowestHttpTokensVer)
		tflog.Error(ctx, msg)
		return fmt.Errorf(msg)
	}
	return nil
}

func getOcmVersionMinor(ver string) string {
	version, err := semver.NewVersion(ver)
	if err != nil {
		segments := strings.Split(ver, ".")
		return fmt.Sprintf("%s.%s", segments[0], segments[1])
	}
	segments := version.Segments()
	return fmt.Sprintf("%d.%d", segments[0], segments[1])
}

// getVersionList returns a list of versions for the given channel group, sorted by
// descending semver
func getVersionList(ctx context.Context, channelGroup string, versionCollection *cmv1.VersionsClient) (versionList []string, err error) {
	vs, err := getVersions(ctx, channelGroup, versionCollection)
	if err != nil {
		err = fmt.Errorf("Failed to retrieve versions: %s", err)
		return
	}

	for _, v := range vs {
		versionList = append(versionList, v.RawID())
	}

	if len(versionList) == 0 {
		err = fmt.Errorf("Could not find versions")
		return
	}

	return
}

func getVersions(ctx context.Context, channelGroup string, versionCollection *cmv1.VersionsClient) (versions []*cmv1.Version, err error) {
	page := 1
	size := 100
	filter := strings.Join([]string{
		"enabled = 'true'",
		"rosa_enabled = 'true'",
		fmt.Sprintf("channel_group = '%s'", channelGroup),
	}, " AND ")
	for {
		var response *cmv1.VersionsListResponse
		response, err = versionCollection.List().
			Search(filter).
			Order("default desc, id desc").
			Page(page).
			Size(size).
			Send()
		if err != nil {
			tflog.Debug(ctx, err.Error())
			return nil, err
		}
		versions = append(versions, response.Items().Slice()...)
		if response.Size() < size {
			break
		}
		page++
	}

	// Sort list in descending order
	sort.Slice(versions, func(i, j int) bool {
		a, erra := ver.NewVersion(versions[i].RawID())
		b, errb := ver.NewVersion(versions[j].RawID())
		if erra != nil || errb != nil {
			return false
		}
		return a.GreaterThan(b)
	})

	return
}

// Upgrades the cluster if the desired (plan) version is greater than the
// current version
func upgradeClusterIfNeeded(ctx context.Context, resourceData *schema.ResourceData,
	clusterRosaState *clusterschema2.ClusterRosaClassicState, clusterCollection *cmv1.ClustersClient,
	versionCollection *cmv1.VersionsClient) error {

	requestedVersionChanged := true
	if !resourceData.HasChange("version") {
		requestedVersionChanged = false
	}
	oldV, newV := resourceData.GetChange("version")
	oldVersion := common2.GetOptionalStringValue(oldV)
	newVersion := common2.GetOptionalStringValue(newV)

	if newVersion == "" || clusterRosaState.CurrentVersion == "" {
		// No version information, nothing to do
		tflog.Debug(ctx, "Insufficient cluster version information to determine if upgrade should be performed.")
		return nil
	}

	logMap := make(map[string]interface{}, 3)
	logMap["current_version"] = clusterRosaState.CurrentVersion
	logMap["old-version"] = oldVersion
	logMap["new-version"] = newVersion
	tflog.Debug(ctx, "Cluster versions", logMap)

	requestedVersionChanged = oldVersion != newVersion

	// Check the versions to see if we need to upgrade
	currentVersion, err := semver.NewVersion(clusterRosaState.CurrentVersion)
	if err != nil {
		return fmt.Errorf("failed to parse current cluster version: %v", err)
	}
	// For backward compatibility
	// In case version format with "openshift-v" was already used
	// remove the prefix to adapt the right format and avoid failure
	fixedVersion := strings.TrimPrefix(newVersion, "openshift-v")
	desiredVersion, err := semver.NewVersion(fixedVersion)
	if err != nil {
		return fmt.Errorf("failed to parse desired cluster version: %v", err)
	}
	if currentVersion.GreaterThan(desiredVersion) {
		tflog.Debug(ctx, "No cluster version upgrade needed.")
		if requestedVersionChanged {
			// User changed the version they want, but actual is higher. We
			// don't support downgrades.
			return fmt.Errorf("cluster version is already above the requested version")
		}
		return nil
	}
	cancelingUpgradeOnly := desiredVersion.Equal(currentVersion)

	if !cancelingUpgradeOnly {
		if err = validateUpgrade(ctx, newVersion, clusterRosaState, versionCollection); err != nil {
			return err
		}
	}

	// Fetch existing upgrade policies
	upgrades, err := upgrade.GetScheduledUpgrades(ctx, clusterCollection, clusterRosaState.ID)
	if err != nil {
		return fmt.Errorf("failed to get upgrade policies: %v", err)
	}
	// Stop if an upgrade is already in progress
	correctUpgradePending, err := upgrade.CheckAndCancelUpgrades(ctx, clusterCollection, upgrades, desiredVersion)
	if err != nil {
		return err
	}
	// Schedule a new upgrade
	if !correctUpgradePending && !cancelingUpgradeOnly {
		ackString := ""
		if clusterRosaState.UpgradeAcksFor != nil {
			ackString = *clusterRosaState.UpgradeAcksFor
		}
		if err = scheduleUpgrade(ctx, clusterCollection, clusterRosaState.ID, desiredVersion, ackString); err != nil {
			return err
		}
	}
	return nil
}

func validateUpgrade(ctx context.Context, newVersion string, clusterRosaState *clusterschema2.ClusterRosaClassicState,
	versionCollection *cmv1.VersionsClient) error {
	// Make sure the desired version is available
	versionId := fmt.Sprintf("openshift-v%s", clusterRosaState.CurrentVersion)
	if clusterRosaState.ChannelGroup != nil && *clusterRosaState.ChannelGroup != ocm.DefaultChannelGroup {
		versionId += "-" + *clusterRosaState.ChannelGroup
	}
	availableVersions, err := upgrade.GetAvailableUpgradeVersions(ctx, versionCollection, versionId)
	if err != nil {
		return fmt.Errorf("failed to get available upgrades: %v", err)
	}
	trimmedDesiredVersion := strings.TrimPrefix(newVersion, "openshift-v")
	desiredVersion, err := semver.NewVersion(trimmedDesiredVersion)
	if err != nil {
		return fmt.Errorf("failed to parse desired version: %v", err)
	}
	found := false
	for _, v := range availableVersions {
		sem, err := semver.NewVersion(v.RawID())
		if err != nil {
			return fmt.Errorf("failed to parse available upgrade version: %v", err)
		}
		if desiredVersion.Equal(sem) {
			found = true
			break
		}
	}
	if !found {
		avail := []string{}
		for _, v := range availableVersions {
			avail = append(avail, v.RawID())
		}
		return fmt.Errorf("desired version (%s) is not in the list of available upgrades (%v)", desiredVersion, avail)
	}

	return nil
}

// Ensure user has acked upgrade gates and schedule the upgrade
func scheduleUpgrade(ctx context.Context, client *cmv1.ClustersClient, clusterID string, desiredVersion *semver.Version, userAckString string) error {
	// Gate agreements are checked when the upgrade is scheduled, resulting
	// in an error return. ROSA cli does this by scheduling once w/ dryRun
	// to look for un-acked agreements.
	clusterClient := client.Cluster(clusterID)
	upgradePoliciesClient := clusterClient.UpgradePolicies()
	gates, description, err := upgrade.CheckMissingAgreements(desiredVersion.String(), clusterID, upgradePoliciesClient)
	if err != nil {
		return fmt.Errorf("failed to check for missing upgrade agreements: %v", err)
	}
	// User ack is required if we have any non-STS-only gates
	userAckRequired := false
	for _, gate := range gates {
		if !gate.STSOnly() {
			userAckRequired = true
		}
	}
	targetMinorVersion := getOcmVersionMinor(desiredVersion.String())
	if userAckRequired && userAckString != targetMinorVersion { // User has not acknowledged mandatory gates, stop here.
		return fmt.Errorf("%s\nTo acknowledge these items, please add \"upgrade_acknowledgements_for = %s\""+
			" and re-apply the changes", description, targetMinorVersion)
	}

	// Ack all gates to OCM
	for _, gate := range gates {
		gateID := gate.ID()
		tflog.Debug(ctx, fmt.Sprintf("Acknowledging version gate %s", gateID))
		gateAgreementsClient := clusterClient.GateAgreements()
		err := upgrade.AckVersionGate(gateAgreementsClient, gateID)
		if err != nil {
			return fmt.Errorf("failed to acknowledge version gate '%s' for cluster '%s': %v",
				gateID, clusterID, err)
		}
	}

	// Schedule an upgrade
	tenMinFromNow := time.Now().UTC().Add(10 * time.Minute)
	newPolicy, err := cmv1.NewUpgradePolicy().
		ScheduleType("manual").
		Version(desiredVersion.String()).
		NextRun(tenMinFromNow).
		Build()
	if err != nil {
		return fmt.Errorf("failed to create upgrade policy: %v", err)
	}
	_, err = clusterClient.UpgradePolicies().
		Add().
		Body(newPolicy).
		SendContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to schedule upgrade: %v", err)
	}
	return nil
}

func updateProxy(resourceData *schema.ResourceData, clusterRosaState *clusterschema2.ClusterRosaClassicState,
	clusterBuilder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, bool, error) {
	if !resourceData.HasChange("proxy") {
		return clusterBuilder, false, nil
	}
	clusterBuilder, err := buildProxy(clusterRosaState, clusterBuilder)
	if err != nil {
		return nil, false, err
	}

	return clusterBuilder, true, nil
}

func updateNodes(resourceData *schema.ResourceData, clusterRosaState *clusterschema2.ClusterRosaClassicState,
	clusterBuilder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, bool, error) {
	// Send request to update the cluster:
	shouldUpdateNodes := false
	clusterNodesBuilder := cmv1.NewClusterNodes()
	if resourceData.HasChange("replicas") {
		shouldUpdateNodes = true
		_, newV := resourceData.GetChange("replicas")

		clusterNodesBuilder = clusterNodesBuilder.Compute(newV.(int))
		shouldUpdateNodes = true

	}

	if clusterRosaState.AutoScalingEnabled != nil && *clusterRosaState.AutoScalingEnabled {
		// autoscaling enabled
		autoscaling := cmv1.NewMachinePoolAutoscaling()

		if clusterRosaState.MaxReplicas != nil {
			autoscaling = autoscaling.MaxReplicas(*clusterRosaState.MaxReplicas)
		}
		if clusterRosaState.MinReplicas != nil {
			autoscaling = autoscaling.MinReplicas(*clusterRosaState.MinReplicas)
		}

		clusterNodesBuilder = clusterNodesBuilder.AutoscaleCompute(autoscaling)
		shouldUpdateNodes = true

	} else {
		if clusterRosaState.MaxReplicas != nil || clusterRosaState.MinReplicas != nil {
			return nil, false, fmt.Errorf("Can't update MaxReplica and/or MinReplica of cluster when autoscaling is not enabled")
		}
	}

	// MP labels update
	if resourceData.HasChange("default_mp_labels") {
		clusterNodesBuilder.ComputeLabels(clusterRosaState.DefaultMPLabels)
		shouldUpdateNodes = true
	}

	if shouldUpdateNodes {
		clusterBuilder = clusterBuilder.Nodes(clusterNodesBuilder)
	}

	return clusterBuilder, shouldUpdateNodes, nil
}

func retryClusterNotFoundWithTimeout(attempts int, sleep time.Duration, ctx context.Context, timeout int,
	resource *cmv1.ClusterClient) (bool, error) {
	isNotFound, err := waitTillClusterIsNotFoundWithTimeout(ctx, timeout, resource)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return retryClusterNotFoundWithTimeout(attempts, 2*sleep, ctx, timeout, resource)
		}
		return isNotFound, err
	}

	return isNotFound, nil
}

func waitTillClusterIsNotFoundWithTimeout(ctx context.Context, timeout int,
	resource *cmv1.ClusterClient) (bool, error) {
	timeoutInMinutes := time.Duration(timeout) * time.Minute
	pollCtx, cancel := context.WithTimeout(ctx, timeoutInMinutes)
	defer cancel()
	_, err := resource.Poll().
		Interval(PollingIntervalInMinutes * time.Minute).
		Status(http.StatusNotFound).
		StartContext(pollCtx)
	sdkErr, ok := err.(*ocm_errors.Error)
	if ok && sdkErr.Status() == http.StatusNotFound {
		tflog.Info(ctx, "Cluster was removed")
		return true, nil
	}
	if err != nil {
		tflog.Error(ctx, "Can't poll cluster deletion")
		return false, err
	}

	return false, nil
}
