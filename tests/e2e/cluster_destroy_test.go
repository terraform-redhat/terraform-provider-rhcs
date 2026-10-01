// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Delete cluster", func() {
	It("DestroyClusterByProfile", ci.Destroy,
		func() {
			// Destroy kubeconfig folder
			if _, err := os.Stat(config.GetKubeConfigDir()); err == nil {
				os.RemoveAll(config.GetKubeConfigDir())
			}

			// Generate/build cluster by profile selected
			profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
			Expect(err).ToNot(HaveOccurred())
			destroyer, err := clusterworkflow.NewDestroyer(profileHandler, token, profileHandler.Profile().GetName())
			Expect(err).ToNot(HaveOccurred())
			err = destroyer.Destroy(context.Background())
			Expect(err).ToNot(HaveOccurred())
		})
})
