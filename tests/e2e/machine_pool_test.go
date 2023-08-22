package e2e

import (

	// nolint

	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-online/ocm-sdk-go/logging"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	conn "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/connection"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var logger logging.Logger

var _ = Describe("TF Test", func() {
	Describe("Create MachinePool test", func() {
		var cluster_id string
		var mpService *exe.MachinePoolService
		BeforeEach(func() {
			cluster_id = os.Getenv("CLUSTER_ID")
			mpService = exe.NewMachinePoolService(con.MachinePoolDir)
		})
		AfterEach(func() {
			err := mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
			return
		})
		It("MachinePoolExampleNegative", func() {

			mpParam := &exe.MachinePoolArgs{
				Token:       token,
				Cluster:     cluster_id,
				Replicas:    3,
				MachineType: "invalid",
				Name:        "testmp",
			}

			_, err := exe.CreateMyTFMachinePool(mpParam, "-auto-approve", "-no-color")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("is not supported for cloud provider 'aws'"))
		})
		Context("Author:amalykhi-High-OCP-64757 @OCP-64757 @amalykhi", func() {
			It("Create a second machine pool", func() {
				token := os.Getenv(con.TokenENVName)
				replicas := 3
				machine_type := "r5.xlarge"
				name := "testmp14"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     cluster_id,
					Replicas:    replicas,
					MachineType: machine_type,
					Name:        name,
				}

				err := mpService.Create(MachinePoolArgs)
				mp_out, err := mpService.Output()
				Expect(err).ToNot(HaveOccurred())
				mp_response, err := cms.RetrieveClusterMachinePool(conn.RHCSConnection, cluster_id, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mp_response.Status()).To(Equal(http.StatusOK))
				respBody := mp_response.Body()
				Expect(respBody.Replicas()).To(Equal(mp_out.Replicas))
				Expect(respBody.InstanceType()).To(Equal(mp_out.MachineType))
				Expect(respBody.ID()).To(Equal(mp_out.Name))
			})
		})
	})
})
