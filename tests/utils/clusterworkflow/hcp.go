// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

type hcpBackend struct {
	profile profilehandler.ProfileHandler
	token   string
}

type hcpDestroyer struct {
	profile profilehandler.ProfileHandler
	token   string
}

func (d hcpDestroyer) Destroy(_ context.Context) error {
	return d.profile.DestroyRHCSClusterResources(d.token)
}

func newHCPBackend(profile profilehandler.ProfileHandler, token string) *Backend {
	backend := &hcpBackend{profile: profile, token: token}
	return &Backend{
		Lifecycle: backend,
		Fetcher:   backend,
	}
}

func (b *hcpBackend) FetchClusterID(_ context.Context) (string, error) {
	if clusterID := os.Getenv(config.EnvClusterID); clusterID != "" {
		if _, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID); err != nil {
			return "", err
		}
		return clusterID, nil
	}
	return b.profile.RetrieveClusterID()
}

func (b *hcpBackend) Create() (string, error) {
	return b.profile.CreateRHCSClusterByProfile(b.token)
}

func (b *hcpBackend) Update(spec exec.ClusterSpec) error {
	if spec.SchedulerProfile != nil {
		return fmt.Errorf("scheduler profile is not supported by the HCP cluster workflow")
	}

	clusterService, err := b.profile.Services().GetClusterService()
	if err != nil {
		return err
	}
	clusterOutput, err := clusterService.Output()
	if err != nil {
		return err
	}

	if spec.Autoscaler != nil {
		autoscalerService, err := b.profile.Services().GetClusterAutoscalerService()
		if err != nil {
			return err
		}
		args, err := autoscalerService.ReadTFVars()
		if err != nil {
			return err
		}
		if args.Cluster == nil {
			args.Cluster = &clusterOutput.ClusterID
		}
		if value := spec.Autoscaler.MaxNodeProvisionTime; value != nil {
			args.MaxNodeProvisionTime = value
		}
		if value := spec.Autoscaler.MaxNodesTotal; value != nil {
			converted := int(*value)
			if args.ResourceLimits == nil {
				args.ResourceLimits = &exec.ResourceLimits{}
			}
			args.ResourceLimits.MaxNodesTotal = &converted
		}
		if value := spec.Autoscaler.MaxPodGracePeriod; value != nil {
			converted := int(*value)
			args.MaxPodGracePeriod = &converted
		}
		if value := spec.Autoscaler.PodPriorityThreshold; value != nil {
			converted := int(*value)
			args.PodPriorityThreshold = &converted
		}
		if _, err := autoscalerService.Apply(args); err != nil {
			return err
		}
	}

	if spec.ComponentRoutes != nil {
		ingressService, err := b.profile.Services().GetIngressService()
		if err != nil {
			return err
		}
		args, err := ingressService.ReadTFVars()
		if err != nil {
			return err
		}
		routes, err := parseComponentRoutes(*spec.ComponentRoutes)
		if err != nil {
			return err
		}
		args.Cluster = &clusterOutput.ClusterID
		args.ComponentRoutes = &routes
		if _, err := ingressService.Apply(args); err != nil {
			return err
		}
	}

	return nil
}

func (b *hcpBackend) Read(_ context.Context) (exec.ClusterState, error) {
	clusterService, err := b.profile.Services().GetClusterService()
	if err != nil {
		return exec.ClusterState{}, err
	}
	clusterOutput, err := clusterService.Output()
	if err != nil {
		return exec.ClusterState{}, err
	}
	clusterDetail, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterOutput.ClusterID)
	if err != nil {
		return exec.ClusterState{}, err
	}

	state := exec.ClusterState{
		ClusterID:   clusterOutput.ClusterID,
		ClusterName: clusterOutput.ClusterName,
		Phase:       string(clusterDetail.Body().State()),
	}

	autoscalerService, err := b.profile.Services().GetClusterAutoscalerService()
	if err != nil {
		return exec.ClusterState{}, err
	}
	if autoscalerOutput, outputErr := autoscalerService.Output(); outputErr == nil {
		state.Autoscaler = &exec.ClusterAutoscalerState{
			MaxNodeProvisionTime: autoscalerOutput.MaxNodeProvisionTime,
			MaxNodesTotal:        int64(autoscalerOutput.MaxNodesTotal),
			MaxPodGracePeriod:    int64(autoscalerOutput.MaxPodGracePeriod),
			PodPriorityThreshold: int64(autoscalerOutput.PodPriorityThreshold),
		}
	}

	ingress, err := cms.RetrieveClusterIngress(cms.RHCSConnection, clusterOutput.ClusterID)
	if err != nil {
		return exec.ClusterState{}, err
	}
	state.ComponentRoutes, err = marshalComponentRoutes(ingress)
	if err != nil {
		return exec.ClusterState{}, err
	}
	return state, nil
}

type componentRouteInput struct {
	Hostname                 string `json:"hostname"`
	TLSSecretRef             string `json:"tls_secret_ref"`
	ServingCertKeyPairSecret struct {
		Name string `json:"name"`
	} `json:"servingCertKeyPairSecret"`
	Name string `json:"name"`
}

func parseComponentRoutes(raw string) (map[string]*exec.IngressComponentRoute, error) {
	var input map[string]componentRouteInput
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		var routes []componentRouteInput
		if arrayErr := json.Unmarshal([]byte(raw), &routes); arrayErr != nil {
			return nil, fmt.Errorf("parsing component routes: %w", err)
		}
		input = make(map[string]componentRouteInput, len(routes))
		for index, route := range routes {
			name := route.Name
			if name == "" {
				name = fmt.Sprintf("route-%d", index)
			}
			input[name] = route
		}
	}
	routes := make(map[string]*exec.IngressComponentRoute, len(input))
	for key, route := range input {
		name := key
		if route.Name != "" {
			name = route.Name
		}
		secret := route.TLSSecretRef
		if secret == "" {
			secret = route.ServingCertKeyPairSecret.Name
		}
		routes[name] = exec.NewIngressComponentRoute(&route.Hostname, &secret)
	}
	return routes, nil
}

func marshalComponentRoutes(ingress *cmv1.Ingress) (string, error) {
	routes := map[string]map[string]string{}
	for name, route := range ingress.ComponentRoutes() {
		routes[name] = map[string]string{
			"hostname":       route.Hostname(),
			"tls_secret_ref": route.TlsSecretRef(),
		}
	}
	encoded, err := json.Marshal(routes)
	if err != nil {
		return "", fmt.Errorf("marshalling component routes: %w", err)
	}
	return string(encoded), nil
}

func (b *hcpBackend) WaitReady(_ context.Context, clusterID string) error {
	if !config.IsWaitForOperators() || b.profile.Profile().IsPrivate() {
		return nil
	}
	return openshift.WaitForOperatorsToBeReady(cms.RHCSConnection, clusterID, 60)
}

var _ exec.ClusterReadinessLifecycle = (*hcpBackend)(nil)
var _ exec.ClusterLifecycle = (*hcpBackend)(nil)
