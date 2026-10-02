// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"os"

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

func (b *hcpBackend) WaitReady(_ context.Context, clusterID string) error {
	if !config.IsWaitForOperators() || b.profile.Profile().IsPrivate() {
		return nil
	}
	return openshift.WaitForOperatorsToBeReady(cms.RHCSConnection, clusterID, 60)
}

var _ exec.ClusterReadinessLifecycle = (*hcpBackend)(nil)
