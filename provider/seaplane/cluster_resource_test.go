// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"

	generated "github.com/terraform-redhat/terraform-provider-rhcs/provider/seaplane/generated"
)

// ── regionFromAZ ──────────────────────────────────────────────────────────────

func TestRegionFromAZ(t *testing.T) {
	cases := []struct {
		az   string
		want string
	}{
		{"us-east-1a", "us-east-1"},
		{"us-east-1b", "us-east-1"},
		{"eu-west-2a", "eu-west-2"},
		{"us-gov-east-1a", "us-gov-east-1"},
		{"us-gov-west-1b", "us-gov-west-1"},
		// Local Zone: strip the suffix after the base region
		{"us-east-1-bos-1a", "us-east-1"},
		// Wavelength Zone
		{"us-east-1-wl1-bos-wlz-1", "us-east-1"},
		// No match
		{"", ""},
		{"invalid", ""},
	}
	for _, tc := range cases {
		got := regionFromAZ(tc.az)
		if got != tc.want {
			t.Errorf("regionFromAZ(%q) = %q, want %q", tc.az, got, tc.want)
		}
	}
}

// ── computeRolesRef ───────────────────────────────────────────────────────────

func TestComputeRolesRef(t *testing.T) {
	ref := computeRolesRef("my-prefix", "123456789012", "aws")
	cases := []struct {
		name string
		arn  string
	}{
		{"ingress", ref.IngressARN},
		{"cloud-controller-manager", ref.KubeCloudControllerARN},
		{"ebs-csi", ref.StorageARN},
		{"image-registry", ref.ImageRegistryARN},
		{"network-config", ref.NetworkARN},
		{"control-plane-operator", ref.ControlPlaneOperatorARN},
		{"node-pool-management", ref.NodePoolManagementARN},
	}
	for _, tc := range cases {
		want := "arn:aws:iam::123456789012:role/my-prefix-" + tc.name
		if tc.arn != want {
			t.Errorf("computeRolesRef %s = %q, want %q", tc.name, tc.arn, want)
		}
	}
}

func TestComputeRolesRef_GovCloud(t *testing.T) {
	ref := computeRolesRef("pfx", "111111111111", "aws-us-gov")
	if ref.IngressARN != "arn:aws-us-gov:iam::111111111111:role/pfx-ingress" {
		t.Errorf("unexpected GovCloud ARN: %s", ref.IngressARN)
	}
}

// ── prefixAndPartitionFromRolesRef ────────────────────────────────────────────

func TestPrefixAndPartitionFromRolesRef_Roundtrip(t *testing.T) {
	cases := []struct {
		prefix    string
		accountID string
		partition string
	}{
		{"my-cluster", "123456789012", "aws"},
		{"sp-e2e-sanity", "999999999999", "aws-us-gov"},
		{"pfx", "000000000001", "aws-cn"},
	}
	for _, tc := range cases {
		ref := computeRolesRef(tc.prefix, tc.accountID, tc.partition)
		gotPrefix, gotPartition := prefixAndPartitionFromRolesRef(ref)
		if gotPrefix != tc.prefix {
			t.Errorf("roundtrip prefix: got %q, want %q", gotPrefix, tc.prefix)
		}
		if gotPartition != tc.partition {
			t.Errorf("roundtrip partition: got %q, want %q", gotPartition, tc.partition)
		}
	}
}

func TestPrefixAndPartitionFromRolesRef_Empty(t *testing.T) {
	prefix, partition := prefixAndPartitionFromRolesRef(hypershiftv1beta1.AWSRolesRef{})
	if prefix != "" || partition != "aws" {
		t.Errorf("empty rolesRef: got prefix=%q partition=%q", prefix, partition)
	}
}

func TestClusterPostFlattenRefreshesAdditionalAllowedPrincipals(t *testing.T) {
	cases := []struct {
		name       string
		principals []string
	}{
		{name: "absent", principals: nil},
		{name: "empty", principals: []string{}},
		{name: "present", principals: []string{"arn:aws:iam::123456789012:role/example-consumer"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			stalePrincipals := stringList(t, []string{
				"arn:aws:iam::123456789012:role/stale-consumer",
			})
			state := ClusterState{
				Aws: awsObjectWithPrincipals(t, stalePrincipals),
			}
			response := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					HostedCluster: v1alpha1.HostedClusterSpecPassthrough{
						Platform: v1alpha1.PlatformSpec{
							AWS: &v1alpha1.AWSPlatformSpec{
								AdditionalAllowedPrincipals: testCase.principals,
							},
						},
					},
				},
			}

			(&ClusterHandlerImpl{}).PostFlatten(context.Background(), &state, response)

			refreshed, ok := state.Aws.Attributes()["additional_allowed_principals"].(types.List)
			if !ok {
				t.Fatalf("additional_allowed_principals is %T, want types.List", state.Aws.Attributes()["additional_allowed_principals"])
			}
			want := types.ListNull(types.StringType)
			if len(testCase.principals) > 0 {
				want = stringList(t, testCase.principals)
			}
			if !refreshed.Equal(want) {
				t.Fatalf("additional_allowed_principals = %s, want %s", refreshed, want)
			}
		})
	}
}

func stringList(t *testing.T, values []string) types.List {
	t.Helper()
	elements := make([]attr.Value, len(values))
	for index, value := range values {
		elements[index] = types.StringValue(value)
	}
	list, diags := types.ListValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("building string list: %v", diags)
	}
	return list
}

func awsObjectWithPrincipals(t *testing.T, principals types.List) types.Object {
	t.Helper()
	object, diags := types.ObjectValue(
		map[string]attr.Type{"additional_allowed_principals": types.ListType{ElemType: types.StringType}},
		map[string]attr.Value{"additional_allowed_principals": principals},
	)
	if diags.HasError() {
		t.Fatalf("building AWS object: %v", diags)
	}
	return object
}

// ── validateClusterTags ───────────────────────────────────────────────────────

// tagsMap builds the types.Map the generated schema produces for the `tags`
// attribute.
func tagsMap(t *testing.T, values map[string]string) types.Map {
	t.Helper()
	elements := make(map[string]attr.Value, len(values))
	for key, value := range values {
		elements[key] = types.StringValue(value)
	}
	tags, diags := types.MapValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("building tags map: %v", diags)
	}
	return tags
}

func TestValidateClusterTags_Valid(t *testing.T) {
	cases := []struct {
		name string
		tags types.Map
	}{
		{"null", types.MapNull(types.StringType)},
		{"unknown", types.MapUnknown(types.StringType)},
		{"empty", tagsMap(t, nil)},
		{"typical", tagsMap(t, map[string]string{"cost-center": "cc-1234", "environment": "production"})},
		// Keys and values may contain the extra characters AWS allows.
		{"allowed characters", tagsMap(t, map[string]string{"kubernetes.io/role": "worker+1@eu=west"})},
	}
	for _, tc := range cases {
		if diags := validateClusterTags(context.Background(), tc.tags); diags.HasError() {
			t.Errorf("%s: unexpected error: %v", tc.name, diags.Errors())
		}
	}
}

func TestValidateClusterTags_Invalid(t *testing.T) {
	cases := []struct {
		name   string
		tags   types.Map
		detail string
	}{
		{
			name:   "empty value",
			tags:   tagsMap(t, map[string]string{"owner": ""}),
			detail: "tag key or tag value can not be empty",
		},
		{
			name:   "empty key",
			tags:   tagsMap(t, map[string]string{"": "platform"}),
			detail: "tag key or tag value can not be empty",
		},
		{
			name:   "reserved aws prefix",
			tags:   tagsMap(t, map[string]string{"aws:cloudformation:stack-name": "mine"}),
			detail: "reserved for AWS use",
		},
		{
			name:   "invalid key character",
			tags:   tagsMap(t, map[string]string{"owner!": "platform"}),
			detail: "expected a valid user tag key 'owner!'",
		},
		{
			name:   "invalid value character",
			tags:   tagsMap(t, map[string]string{"owner": "platform!"}),
			detail: "expected a valid user tag value for key 'owner'",
		},
	}
	for _, tc := range cases {
		diags := validateClusterTags(context.Background(), tc.tags)
		if !diags.HasError() {
			t.Errorf("%s: expected an error, got none", tc.name)
			continue
		}
		if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, tc.detail) {
			t.Errorf("%s: detail = %q, want it to contain %q", tc.name, detail, tc.detail)
		}
	}
}

func TestClusterNetworkType(t *testing.T) {
	network, diags := types.ObjectValue(map[string]attr.Type{
		"network_type": types.StringType,
	}, map[string]attr.Value{
		"network_type": types.StringValue("OVNKubernetes"),
	})
	if diags.HasError() {
		t.Fatalf("building network object: %v", diags)
	}

	if got := clusterNetworkType(network); got != "OVNKubernetes" {
		t.Fatalf("clusterNetworkType() = %q, want %q", got, "OVNKubernetes")
	}
	if got := clusterNetworkType(types.ObjectNull(map[string]attr.Type{"network_type": types.StringType})); got != "" {
		t.Fatalf("clusterNetworkType(null) = %q, want empty", got)
	}
}

func TestPreExpandRejectsNoCNIWithNetworkType(t *testing.T) {
	network, diags := types.ObjectValue(map[string]attr.Type{
		"network_type": types.StringType,
	}, map[string]attr.Value{
		"network_type": types.StringValue("OVNKubernetes"),
	})
	if diags.HasError() {
		t.Fatalf("building network object: %v", diags)
	}

	input := &ClusterState{
		No_cni:  types.BoolValue(true),
		Network: network,
	}
	if diags := (&ClusterHandlerImpl{}).PreExpand(context.Background(), input); !diags.HasError() {
		t.Fatal("PreExpand() accepted no_cni together with network.network_type")
	}
}

func TestPreExpandRejectsUnsupportedNetworkType(t *testing.T) {
	network, diags := types.ObjectValue(map[string]attr.Type{
		"network_type": types.StringType,
	}, map[string]attr.Value{
		"network_type": types.StringValue("Calico"),
	})
	if diags.HasError() {
		t.Fatalf("building network object: %v", diags)
	}

	input := &ClusterState{Network: network}
	if diags := (&ClusterHandlerImpl{}).PreExpand(context.Background(), input); !diags.HasError() {
		t.Fatal("PreExpand() accepted unsupported network type")
	}
}

func TestPostExpandSetsOtherNetworkTypeForNoCNI(t *testing.T) {
	aws, diags := types.ObjectValue(map[string]attr.Type{
		"aws_partition":      types.StringType,
		"vpc_id":             types.StringType,
		"aws_subnet_ids":     types.ListType{ElemType: types.StringType},
		"availability_zones": types.ListType{ElemType: types.StringType},
	}, map[string]attr.Value{
		"aws_partition":      types.StringValue("aws"),
		"vpc_id":             types.StringValue("vpc-123"),
		"aws_subnet_ids":     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-123")}),
		"availability_zones": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-east-1a")}),
	})
	if diags.HasError() {
		t.Fatalf("building AWS object: %v", diags)
	}

	input := &ClusterState{
		Operator_roles_prefix: types.StringValue("test"),
		Cloud_region:          types.StringValue("us-east-1"),
		No_cni:                types.BoolValue(true),
		Aws:                   aws,
	}
	obj := &v1alpha1.Cluster{}
	if diags := (&ClusterHandlerImpl{}).PostExpand(context.Background(), input, obj); diags.HasError() {
		t.Fatalf("PostExpand() returned diagnostics: %v", diags)
	}

	if got := obj.Spec.HostedCluster.Networking.NetworkType; got != hypershiftv1beta1.Other {
		t.Fatalf("PostExpand() network type = %q, want %q", got, hypershiftv1beta1.Other)
	}
}

// fullNetworkObject builds the network bundle the way nativeClusterToTerraform
// does on flatten, so tests can mirror real (e.g. imported) state.
func fullNetworkObject(t *testing.T, networkType string) types.Object {
	t.Helper()
	attributes := map[string]attr.Value{
		"api_allowed_cidr_blocks": types.StringNull(),
		"pod_cidr":                types.StringNull(),
		"machine_cidr":            types.StringNull(),
		"network_type":            types.StringNull(),
		"service_cidr":            types.StringNull(),
	}
	if networkType != "" {
		attributes["network_type"] = types.StringValue(networkType)
	}
	network, diags := types.ObjectValue(map[string]attr.Type{
		"api_allowed_cidr_blocks": types.StringType,
		"pod_cidr":                types.StringType,
		"machine_cidr":            types.StringType,
		"network_type":            types.StringType,
		"service_cidr":            types.StringType,
	}, attributes)
	if diags.HasError() {
		t.Fatalf("building network object: %v", diags)
	}
	return network
}

func TestPostFlattenNormalizesNoCNIState(t *testing.T) {
	// Mirror the state ImportState produces for a cluster created with
	// no_cni=true: pure flatten of the API response, so network_type is
	// "Other" and no_cni is unset.
	state := &ClusterState{Network: fullNetworkObject(t, "Other")}
	resp := &v1alpha1.Cluster{}
	resp.Spec.HostedCluster.Networking.NetworkType = hypershiftv1beta1.Other

	(&ClusterHandlerImpl{}).PostFlatten(context.Background(), state, resp)

	if !state.No_cni.ValueBool() {
		t.Fatal("PostFlatten() did not set no_cni for network type Other")
	}
	networkType, ok := state.Network.Attributes()["network_type"].(types.String)
	if !ok || !networkType.IsNull() {
		t.Fatalf("PostFlatten() network_type = %v, want null", state.Network.Attributes()["network_type"])
	}
}

func TestPostFlattenPreservesOVNNetworkType(t *testing.T) {
	// State of a cluster managed with an explicit network_type must not be
	// rewritten by the no-CNI normalization.
	state := &ClusterState{
		Network: fullNetworkObject(t, "OVNKubernetes"),
		No_cni:  types.BoolValue(false),
	}
	resp := &v1alpha1.Cluster{}
	resp.Spec.HostedCluster.Networking.NetworkType = hypershiftv1beta1.OVNKubernetes

	(&ClusterHandlerImpl{}).PostFlatten(context.Background(), state, resp)

	if state.No_cni.ValueBool() {
		t.Fatal("PostFlatten() set no_cni for network type OVNKubernetes")
	}
	networkType, ok := state.Network.Attributes()["network_type"].(types.String)
	if !ok || networkType.ValueString() != "OVNKubernetes" {
		t.Fatalf("PostFlatten() network_type = %v, want OVNKubernetes", state.Network.Attributes()["network_type"])
	}
}

func TestPostFlattenFillsUnknownNoCNI(t *testing.T) {
	// no_cni is Optional+Computed: when the configuration omits it the
	// framework marks the planned value unknown, and PostFlatten must replace
	// that with a known value on every path or the apply fails with an
	// inconsistent-result error.
	cases := []struct {
		networkType hypershiftv1beta1.NetworkType
		want        bool
	}{
		{hypershiftv1beta1.Other, true},
		{hypershiftv1beta1.OVNKubernetes, false},
		{hypershiftv1beta1.NetworkType(""), false},
	}
	for _, tc := range cases {
		state := &ClusterState{
			No_cni:  types.BoolUnknown(),
			Network: fullNetworkObject(t, string(tc.networkType)),
		}
		resp := &v1alpha1.Cluster{}
		resp.Spec.HostedCluster.Networking.NetworkType = tc.networkType

		(&ClusterHandlerImpl{}).PostFlatten(context.Background(), state, resp)

		if state.No_cni.IsUnknown() || state.No_cni.IsNull() {
			t.Fatalf("PostFlatten() left no_cni unknown/null for network type %q", tc.networkType)
		}
		if got := state.No_cni.ValueBool(); got != tc.want {
			t.Fatalf("PostFlatten() no_cni = %v for network type %q, want %v", got, tc.networkType, tc.want)
		}
	}
}

// clusterSchema returns the generated Cluster resource schema.
func clusterSchema(t *testing.T) schema.Schema {
	t.Helper()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	(&generated.ClusterResource{}).Schema(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("cluster schema diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

// TestNoCniPlanModifiersResolveUnknownWithoutReplacement pins the plan
// modifier behavior for no_cni: with Optional+Computed and the configuration
// omitting the field, the framework marks the planned value unknown on
// updates, and the modifiers must resolve it to the state value without
// latching a replacement. This holds only when UseStateForUnknown runs
// before RequiresReplace.
func TestNoCniPlanModifiersResolveUnknownWithoutReplacement(t *testing.T) {
	s := clusterSchema(t)
	boolAttr, ok := s.Attributes["no_cni"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("no_cni is not a BoolAttribute")
	}
	if !boolAttr.Optional || !boolAttr.Computed {
		t.Fatalf("no_cni must be Optional+Computed, got optional=%v computed=%v", boolAttr.Optional, boolAttr.Computed)
	}

	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"no_cni": tftypes.Bool}}
	wrapperSchema := schema.Schema{Attributes: map[string]schema.Attribute{"no_cni": boolAttr}}
	obj := func(v tftypes.Value) tftypes.Value {
		return tftypes.NewValue(objType, map[string]tftypes.Value{"no_cni": v})
	}
	modReq := planmodifier.BoolRequest{
		Config:      tfsdk.Config{Raw: obj(tftypes.NewValue(tftypes.Bool, nil)), Schema: wrapperSchema},
		Plan:        tfsdk.Plan{Raw: obj(tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)), Schema: wrapperSchema},
		State:       tfsdk.State{Raw: obj(tftypes.NewValue(tftypes.Bool, true)), Schema: wrapperSchema},
		ConfigValue: types.BoolNull(),
		PlanValue:   types.BoolUnknown(),
		StateValue:  types.BoolValue(true),
	}
	modResp := &planmodifier.BoolResponse{PlanValue: modReq.PlanValue}
	for _, m := range boolAttr.PlanModifiers {
		m.PlanModifyBool(context.Background(), modReq, modResp)
		modReq.PlanValue = modResp.PlanValue
	}
	if modResp.RequiresReplace {
		t.Fatal("no_cni plan modifiers required replacement for an unconfigured computed value")
	}
	if !modResp.PlanValue.Equal(types.BoolValue(true)) {
		t.Fatalf("no_cni plan value = %v, want the state value true", modResp.PlanValue)
	}
}

// TestCloudRegionPlanModifiersResolveUnknownWithoutReplacement pins the
// same ordering guarantee for cloud_region, which is Computed and was
// previously generated with RequiresReplace before UseStateForUnknown — an
// order that latches a spurious replacement on every in-place update.
func TestCloudRegionPlanModifiersResolveUnknownWithoutReplacement(t *testing.T) {
	s := clusterSchema(t)
	stringAttr, ok := s.Attributes["cloud_region"].(schema.StringAttribute)
	if !ok {
		t.Fatal("cloud_region is not a StringAttribute")
	}
	if !stringAttr.Computed {
		t.Fatal("cloud_region must be Computed")
	}

	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"cloud_region": tftypes.String}}
	wrapperSchema := schema.Schema{Attributes: map[string]schema.Attribute{"cloud_region": stringAttr}}
	obj := func(v tftypes.Value) tftypes.Value {
		return tftypes.NewValue(objType, map[string]tftypes.Value{"cloud_region": v})
	}
	modReq := planmodifier.StringRequest{
		Config: tfsdk.Config{Raw: obj(tftypes.NewValue(tftypes.String, nil)), Schema: wrapperSchema},
		Plan: tfsdk.Plan{
			Raw:    obj(tftypes.NewValue(tftypes.String, tftypes.UnknownValue)),
			Schema: wrapperSchema,
		},
		State:       tfsdk.State{Raw: obj(tftypes.NewValue(tftypes.String, "us-east-1")), Schema: wrapperSchema},
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		StateValue:  types.StringValue("us-east-1"),
	}
	modResp := &planmodifier.StringResponse{PlanValue: modReq.PlanValue}
	for _, m := range stringAttr.PlanModifiers {
		m.PlanModifyString(context.Background(), modReq, modResp)
		modReq.PlanValue = modResp.PlanValue
	}
	if modResp.RequiresReplace {
		t.Fatal("cloud_region plan modifiers required replacement for an unconfigured computed value")
	}
	if !modResp.PlanValue.Equal(types.StringValue("us-east-1")) {
		t.Fatalf("cloud_region plan value = %v, want the state value us-east-1", modResp.PlanValue)
	}
}
