package e2e

import (
	// nolint

	"fmt"
	"net/http"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-online/ocm-common/pkg/rosa/accountroles"
	. "github.com/openshift-online/ocm-sdk-go/testing"

	cmsv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("Verify cluster", func() {
	defer GinkgoRecover()

	var err error
	var profile *ci.Profile
	var cluster *cmsv1.Cluster

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
		Expect(err).ToNot(HaveOccurred())
		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		cluster = getResp.Body()

	})

	It("proxy is correctly set - [id:67607]", ci.Day1Post, ci.High, func() {
		if !profile.Proxy {
			Skip("No proxy is configured for the cluster. skipping the test.")
		}
		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(getResp.Body().Proxy().HTTPProxy()).To(ContainSubstring("http://"))
		Expect(getResp.Body().Proxy().HTTPSProxy()).To(ContainSubstring("https://"))
		Expect(getResp.Body().Proxy().NoProxy()).To(Equal("quay.io"))
		Expect(getResp.Body().AdditionalTrustBundle()).To(Equal("REDACTED"))
	})

	It("is successfully installed - [id:63134]", ci.Day1Post, ci.High, func() {
		getResp, err := cms.RetrieveClusterStatus(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(getResp.Body().State())).To(Equal("ready"))
		Expect(getResp.Status()).To(Equal(http.StatusOK))

		dnsReady, hasValue := getResp.Body().GetDNSReady()
		Expect(dnsReady).To(Equal(true))
		Expect(hasValue).To(Equal(true))

		oidcReady, hasValue := getResp.Body().GetOIDCReady()
		Expect(oidcReady).To(Equal(true))
		Expect(hasValue).To(Equal(true))
	})

	It("custom properties is correctly set - [id:64906]", ci.Day1Post, ci.Medium, func() {
		By("Check custom_property field is present under cluster's properties")
		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(getResp.Body().Properties()["custom_property"]).To(Equal("test"))

		By("Check rosa_tf_commit is present under cluster's properties")
		Expect(getResp.Body().Properties()["rosa_tf_commit"]).ShouldNot(BeEmpty())

		By("Check rosa_tf_version is present under cluster's properties")
		Expect(getResp.Body().Properties()["rosa_tf_version"]).ShouldNot(BeEmpty())
	})

	It("fips is correctly enabled/disabled - [id:63140]", ci.Day1Post, ci.High, func() {
		Expect(cluster.FIPS()).To(Equal(profile.FIPS))
	})

	It("private_link is correctly enabled/disabled - [id:63133]", ci.Day1Post, ci.High, func() {
		Expect(cluster.AWS().PrivateLink()).To(Equal(profile.IsPrivateLink()))
	})

	It("etcd-encryption is correctly enabled/disabled - [id:63143]", ci.Day1Post, ci.High, func() {
		Expect(cluster.EtcdEncryption()).To(Equal(profile.Etcd))
	})

	It("compute_machine_type is correctly set - [id:64023]", ci.Day1Post, ci.Medium, func() {
		if profile.ComputeMachineType == "" {
			Skip("No compute_machine_type is configured for the cluster. skipping the test.")
		}
		Expect(cluster.Nodes().ComputeMachineType().ID()).To(Equal(profile.ComputeMachineType))
	})

	It("compute_replicas is correctly set - [id:73153]", ci.Day1Post, ci.Medium, func() {
		if profile.ComputeReplicas <= 0 {
			Skip("No compute_replicas is configured for the cluster. skipping the test.")
		}
		Expect(cluster.Nodes().Compute()).To(Equal(profile.ComputeReplicas))
	})

	It("availability zones and multi-az are correctly set - [id:63141]", ci.Day1Post, ci.Medium, func() {
		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		zonesArray := strings.Split(profile.Zones, ",")
		clusterAvailZones := getResp.Body().Nodes().AvailabilityZones()
		Expect(err).ToNot(HaveOccurred())
		Expect(getResp.Body().MultiAZ()).To(Equal(profile.MultiAZ))

		if profile.Zones != "" {
			Expect(clusterAvailZones).To(Equal(helper.JoinStringWithArray(profile.Region, zonesArray)))
		} else {
			// the default zone for each region
			Expect(clusterAvailZones[0]).To(Equal(fmt.Sprintf("%sa", profile.Region)))
		}
	})

	It("compute_labels are correctly set - [id:68423]", ci.Day1Post, ci.High, func() {
		if profile.Labeling {
			Expect(cluster.Nodes().ComputeLabels()).To(Equal(constants.DefaultMPLabels))
		} else {
			Expect(cluster.Nodes().ComputeLabels()).To(Equal(constants.NilMap))
		}
	})

	It("AWS tags are set - [id:63777]", ci.Day1Post, ci.High, func() {
		buildInTags := map[string]string{
			"red-hat-clustertype": "rosa",
			"red-hat-managed":     "true",
		}

		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		if profile.Tagging {
			allClusterTags := helper.MergeMaps(buildInTags, constants.Tags)
			Expect(len(getResp.Body().AWS().Tags())).To(Equal(len(allClusterTags)))

			// compare cluster tags to the expected tags to appear
			for k, v := range getResp.Body().AWS().Tags() {
				Expect(v).To(Equal(allClusterTags[k]))
			}
		} else {
			Expect(len(getResp.Body().AWS().Tags())).To(Equal(len(buildInTags)))
		}
	})
	It("can be imported - [id:65684]",
		ci.Day2, ci.Medium, ci.FeatureImport, func() {

			if profile.GetClusterType().HCP {
				Skip("Test can run only on Classic cluster")
			}

			importService, err := exec.NewImportService(constants.ImportResourceDir) // init new import service
			Expect(err).ToNot(HaveOccurred())

			By("Run the command to import rosa_classic resource")
			defer func() {
				importService.Destroy()
			}()
			importParam := &exec.ImportArgs{
				ClusterID: clusterID,
				Resource:  "rhcs_cluster_rosa_classic.rosa_sts_cluster_import",
			}
			_, err = importService.Import(importParam)
			Expect(err).ToNot(HaveOccurred())

			By("Check resource state - import command succeeded")
			output, err := importService.ShowState(importParam.Resource)
			Expect(err).ToNot(HaveOccurred())

			// validate import was successful by checking samples fields
			Expect(output).To(ContainSubstring(profile.ClusterName))
			Expect(output).To(ContainSubstring(profile.Region))
			Expect(output).To(ContainSubstring(profile.ChannelGroup))

			By("Validate terraform import with no clusterID returns error")
			var unknownClusterID = helper.GenerateRandomStringWithSymbols(20)
			importParam = &exec.ImportArgs{
				ClusterID: unknownClusterID,
				Resource:  "rhcs_cluster_rosa_classic.rosa_import_no_cluster_id",
			}

			_, err = importService.Import(importParam)
			Expect(err.Error()).To(ContainSubstring("Cannot import non-existent remote object"))
		})

	It("confirm cluster admin user created ONLY during cluster creation operation - [id:65928]",
		ci.Day1Post, ci.High,
		ci.Exclude,
		func() {
			if !profile.AdminEnabled {
				Skip("The test configured only for cluster admin profile")
			}

			By("List existing Htpasswd IDP")
			idpList, _ := cms.ListClusterIDPs(ci.RHCSConnection, clusterID)
			Expect(idpList.Items().Len()).To(Equal(1))
			Expect(idpList.Items().Slice()[0].Name()).To(Equal("cluster-admin"))

			By("List existing HtpasswdUsers and compare to the created one")
			idpID := idpList.Items().Slice()[0].ID()
			htpasswdUsersList, _ := cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpID)
			Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
			respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
			Expect(respUserName).To(Equal("rhcs-clusteradmin"))

			By("Check resource state file is updated")
			resource, err := helper.GetResource(profile.GetClusterManifestsDir(), "rhcs_cluster_rosa_classic", "rosa_sts_cluster")
			Expect(err).ToNot(HaveOccurred())
			passwordInState, _ := JQ(`.instances[0].attributes.admin_credentials.password`, resource)
			Expect(passwordInState).NotTo(BeEmpty())
			Expect(resource).To(MatchJQ(`.instances[0].attributes.admin_credentials.username`, "rhcs-clusteradmin"))

			By("Login with created cluster admin password")
			getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			server := getResp.Body().API().URL()

			username := constants.ClusterAdminUser
			password := helper.GetClusterAdminPassword()
			Expect(password).ToNot(BeEmpty())

			ocAtter := &openshift.OcAttributes{
				Server:    server,
				Username:  username,
				Password:  password,
				ClusterID: clusterID,
				AdditioanlFlags: []string{
					"--insecure-skip-tls-verify",
					fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, username))),
				},
				Timeout: 10,
			}
			_, err = openshift.OcLogin(*ocAtter)
			Expect(err).ToNot(HaveOccurred())

		})

	It("additional security group are correctly set - [id:69145]",
		ci.Day1Post, ci.Critical,
		func() {
			if profile.GetClusterType().HCP {
				Skip("Test can run only on Classic cluster")
			}
			By("Check the profile settings")
			if profile.AdditionalSGNumber == 0 {
				Expect(cluster.AWS().AdditionalComputeSecurityGroupIds()).To(BeEmpty())
				Expect(cluster.AWS().AdditionalControlPlaneSecurityGroupIds()).To(BeEmpty())
				Expect(cluster.AWS().AdditionalInfraSecurityGroupIds()).To(BeEmpty())
			} else {
				By("Verify CMS are using the correct configuration")
				clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
				Expect(err).ToNot(HaveOccurred())
				output, err := clusterService.Output()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(output.AdditionalComputeSecurityGroups)).To(Equal(len(cluster.AWS().AdditionalComputeSecurityGroupIds())))
				Expect(len(output.AdditionalInfraSecurityGroups)).To(Equal(len(cluster.AWS().AdditionalInfraSecurityGroupIds())))
				Expect(len(output.AdditionalControlPlaneSecurityGroups)).To(Equal(len(cluster.AWS().AdditionalControlPlaneSecurityGroupIds())))
				for _, sg := range output.AdditionalComputeSecurityGroups {
					Expect(sg).To(BeElementOf(cluster.AWS().AdditionalComputeSecurityGroupIds()))
				}

			}

		})

	It("worker disk size is set correctly - [id:69143]",
		ci.Day1Post, ci.Critical,
		func() {
			if profile.GetClusterType().HCP {
				Skip("Test can run only on Classic cluster")
			}

			switch profile.WorkerDiskSize {
			case 0:
				Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(300))
			default:
				Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(profile.WorkerDiskSize))
			}
		})

	It("account roles/policies unified path is correctly set - [id:63138]", ci.Day1Post, ci.Medium, func() {
		unifiedPath, err := accountroles.GetPathFromAccountRole(cluster, accountroles.AccountRoles[accountroles.InstallerAccountRole].Name)
		Expect(err).ToNot(HaveOccurred())
		Expect(profile.UnifiedAccRolesPath).Should(ContainSubstring(unifiedPath))
	})

	It("customer-managed KMS key is set correctly - [id:63334]", ci.Day1Post, ci.Medium, func() {
		if !profile.KMSKey {
			Skip("The test is configured only for the profile containing the KMS key")
		}
		By("Check the kmsKeyARN")
		listRSresp, err := cms.ListClusterResources(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		kmsService, err := exec.NewKMSService()
		Expect(err).ToNot(HaveOccurred())
		kmsOutput, _ := kmsService.Output()
		expectedKeyArn := kmsOutput.KeyARN
		if profile.GetClusterType().HCP {
			for key, value := range listRSresp.Body().Resources() {
				if strings.Contains(key, "workers") {
					Expect(value).Should(ContainSubstring(`"encryptionKey":"` + expectedKeyArn + `"`))
				}
			}
		} else {
			awsAccountClaim := listRSresp.Body().Resources()["aws_account_claim"]
			Expect(awsAccountClaim).Should(ContainSubstring(`"kmsKeyId":"` + expectedKeyArn + `"`))
		}
	})

	It("imdsv2 value is set correctly - [id:63950]", ci.Day1Post, ci.Critical, func() {
		if profile.GetClusterType().HCP {
			Skip("Test can run only on Classic cluster")
		}
		if profile.Ec2MetadataHttpTokens != "" {
			Expect(string(cluster.AWS().Ec2MetadataHttpTokens())).To(Equal(profile.Ec2MetadataHttpTokens))
		} else {
			Expect(cluster.AWS().Ec2MetadataHttpTokens()).To(Equal(cmsv1.Ec2MetadataHttpTokensOptional))
		}
	})
})
