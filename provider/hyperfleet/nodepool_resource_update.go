package hyperfleet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hfplatform "github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"

	generated "github.com/terraform-redhat/terraform-provider-rhcs/provider/hyperfleet/generated"
)

var (
	_ resource.Resource                = (*nodePoolResource)(nil)
	_ resource.ResourceWithConfigure   = (*nodePoolResource)(nil)
	_ resource.ResourceWithImportState = (*nodePoolResource)(nil)
)

type nodePoolResource struct {
	*generated.NodePoolResource
}

func newNodePoolResource(inner *generated.NodePoolResource) *nodePoolResource {
	return &nodePoolResource{NodePoolResource: inner}
}

func (resourceImplementation *nodePoolResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan generated.NodePoolState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priorState generated.NodePoolState
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceImplementation.NodePoolResource.Update(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	clearLabels, clearTaints := nodePoolShouldClearCollections(&plan, &priorState)
	if !clearLabels && !clearTaints {
		return
	}

	current, err := resourceImplementation.Client.HyperfleetV1alpha1().NodePools(
		resourceImplementation.Handler.Namespace(ctx, &plan),
	).Get(ctx, plan.Id.ValueString(), hfplatform.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read NodePool for clear update", err.Error())
		return
	}

	updateDocument, err := nodePoolClearUpdateDocument(current, clearLabels, clearTaints)
	if err != nil {
		resp.Diagnostics.AddError("Failed to encode NodePool clear update", err.Error())
		return
	}
	body, err := json.Marshal(updateDocument)
	if err != nil {
		resp.Diagnostics.AddError("Failed to encode NodePool clear update", err.Error())
		return
	}

	var cleared v1alpha1.NodePool
	// Hyperfleet routes NodePool mutations by UID, not metadata.name. This mirrors
	// the generated clientset's Update wrapper, which sets Name to the UID on a
	// deep copy before building the PUT URL.
	err = resourceImplementation.Client.HyperfleetV1alpha1().RESTClient().Put().
		NamespaceIfScoped(resourceImplementation.Handler.Namespace(ctx, &plan), true).
		Resource("nodepools").
		Name(string(current.UID)).
		Body(body).
		Do(ctx).
		Into(&cleared)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to clear NodePool collections",
			fmt.Sprintf("%v", err),
		)
		return
	}

	resp.Diagnostics.Append(resourceImplementation.Handler.PostResponse(ctx, &cleared)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state generated.NodePoolState
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceImplementation.Handler.PostFlatten(ctx, &state, &cleared)
	if clearLabels {
		state.Labels = types.MapNull(types.StringType)
	}
	if clearTaints {
		state.Taints = types.ListNull(plan.Taints.ElementType(ctx))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func nodePoolClearUpdateDocument(
	nodePool *v1alpha1.NodePool,
	clearLabels, clearTaints bool,
) (map[string]any, error) {
	if nodePool == nil {
		return nil, fmt.Errorf("NodePool response must not be nil")
	}

	encoded, err := json.Marshal(nodePool)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		return nil, err
	}
	spec, ok := document["spec"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("NodePool response spec must be an object")
	}

	if clearLabels {
		spec["labels"] = map[string]any{}
	}
	if clearTaints {
		nodePoolSpec, ok := spec["nodePool"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("NodePool response spec.nodePool must be an object")
		}
		nodePoolSpec["taints"] = []any{}
	}
	return document, nil
}

func nodePoolShouldClearCollections(
	plan *generated.NodePoolState,
	state *generated.NodePoolState,
) (bool, bool) {
	clearLabels := plan.Labels.IsNull() && !state.Labels.IsNull()
	clearTaints := plan.Taints.IsNull() && !state.Taints.IsNull()
	return clearLabels, clearTaints
}
