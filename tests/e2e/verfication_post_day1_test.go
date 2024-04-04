package e2e

import (
	// nolint

	"fmt"
	"net/http"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rosaAccountRoles "github.com/openshift-online/ocm-common/pkg/rosa/accountroles"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	H "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("Verify cluster", func() {

	var err error
	var profile *ci.Profile
	var cluster *cmv1.Cluster
	var importService *exe.ImportService

	BeforeEach(func() {
		importService = exe.NewImportService(con.ImportResourceDir) // init new import service
		profile = ci.LoadProfileYamlFileByENV()
		Expect(err).ToNot(HaveOccurred())
		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		cluster = getResp.Body()

	})
	AfterEach(func() {
		// For future implementation
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

	It("availability zones and multi-az are correctly set - [id:63141]", ci.Day1Post, ci.Medium, func() {
		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		zonesArray := strings.Split(profile.Zones, ",")
		clusterAvailZones := getResp.Body().Nodes().AvailabilityZones()
		Expect(err).ToNot(HaveOccurred())
		Expect(getResp.Body().MultiAZ()).To(Equal(profile.MultiAZ))

		if profile.Zones != "" {
			Expect(clusterAvailZones).To(Equal(H.JoinStringWithArray(profile.Region, zonesArray)))
		} else {
			// the default zone for each region
			Expect(clusterAvailZones[0]).To(Equal(fmt.Sprintf("%sa", profile.Region)))
		}
	})

	It("compute_labels are correctly set - [id:68423]", ci.Day1Post, ci.High, func() {
		if profile.Labeling {
			Expect(cluster.Nodes().ComputeLabels()).To(Equal(con.DefaultMPLabels))
		} else {
			Expect(cluster.Nodes().ComputeLabels()).To(Equal(con.NilMap))
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
			allClusterTags := H.MergeMaps(buildInTags, con.Tags)
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
		ci.Day2, ci.Medium, ci.NonHCPCluster, ci.FeatureImport, func() {

			By("Run the command to import rosa_classic resource")
			importParam := &exe.ImportArgs{
				ClusterID:    clusterID,
				ResourceKind: "rhcs_cluster_rosa_classic",
				ResourceName: "rosa_sts_cluster_import",
			}
			Expect(importService.Import(importParam)).To(Succeed())

			By("Check resource state - import command succeeded")
			output, err := importService.ShowState(importParam)
			Expect(err).ToNot(HaveOccurred())

			// validate import was successful by checking samples fields
			Expect(output).To(ContainSubstring(profile.ClusterName))
			Expect(output).To(ContainSubstring(profile.Region))
			Expect(output).To(ContainSubstring(profile.ChannelGroup))

			By("Validate terraform import with no clusterID returns error")
			var unknownClusterID = h.GenerateRandomStringWithSymbols(20)
			importParam = &exe.ImportArgs{
				ClusterID:    unknownClusterID,
				ResourceKind: "rhcs_cluster_rosa_classic",
				ResourceName: "rosa_import_no_cluster_id",
			}

			err = importService.Import(importParam)
			Expect(err.Error()).To(ContainSubstring("Cannot import non-existent remote object"))

			By("clean .tfstate file to revert test changes")
			defer h.CleanManifestsStateFile(con.ImportResourceDir)

		})

	It("confirm cluster admin user created ONLY during cluster creation operation - [id:65928]",
		ci.Day1Post, ci.High,
		ci.Exclude,
		func() {
			if !profile.AdminEnabled {
				Skip("The test configured only for cluster admin profile")
			}
			By("Login with created cluster admin password")
			getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			server := getResp.Body().API().URL()

			username := con.ClusterAdminUser
			password := H.GetClusterAdminPassword()
			Expect(password).ToNot(BeEmpty())

			ocAtter := &openshift.OcAttributes{
				Server:    server,
				Username:  username,
				Password:  password,
				ClusterID: clusterID,
				AdditioanlFlags: []string{
					"--insecure-skip-tls-verify",
					fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, username))),
				},
				Timeout: 10,
			}
			_, err = openshift.OcLogin(*ocAtter)
			Expect(err).ToNot(HaveOccurred())

		})

	It("additional security group are correctly set - [id:69145]",
		ci.Day1Post, ci.Critical, ci.NonHCPCluster,
		func() {
			By("Check the profile settings")
			if profile.AdditionalSGNumber == 0 {
				Expect(cluster.AWS().AdditionalComputeSecurityGroupIds()).To(BeEmpty())
				Expect(cluster.AWS().AdditionalControlPlaneSecurityGroupIds()).To(BeEmpty())
				Expect(cluster.AWS().AdditionalInfraSecurityGroupIds()).To(BeEmpty())
			} else {
				By("Verify CMS are using the correct configuration")
				clusterService, err := exe.NewClusterService(profile.GetClusterManifestsDir())
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

	// Skip this tests until OCM-5079 fixed
	XIt("can not apply changes to security group - [id:69145]", ci.NonHCPCluster, func() {
		clusterService, err := exe.NewClusterService(profile.GetClusterManifestsDir())
		Expect(err).ToNot(HaveOccurred())
		outPut, err := clusterService.Output()
		Expect(err).ToNot(HaveOccurred())
		args := map[string]*exe.ClusterCreationArgs{
			"aws_additional_compute_security_group_ids": {
				AdditionalComputeSecurityGroups:      outPut.AdditionalComputeSecurityGroups[0:1],
				AdditionalInfraSecurityGroups:        outPut.AdditionalInfraSecurityGroups,
				AdditionalControlPlaneSecurityGroups: outPut.AdditionalControlPlaneSecurityGroups,
				AWSRegion:                            profile.Region,
			},
			"aws_additional_infra_security_group_ids": {
				AdditionalInfraSecurityGroups:        outPut.AdditionalInfraSecurityGroups[0:1],
				AdditionalComputeSecurityGroups:      outPut.AdditionalComputeSecurityGroups,
				AdditionalControlPlaneSecurityGroups: outPut.AdditionalControlPlaneSecurityGroups,
				AWSRegion:                            profile.Region,
			},
			"aws_additional_control_plane_security_group_ids": {
				AdditionalControlPlaneSecurityGroups: outPut.AdditionalControlPlaneSecurityGroups[0:1],
				AdditionalComputeSecurityGroups:      outPut.AdditionalComputeSecurityGroups,
				AdditionalInfraSecurityGroups:        outPut.AdditionalInfraSecurityGroups,
				AWSRegion:                            profile.Region,
			},
		}
		for keyword, updatingArgs := range args {
			output, err := clusterService.Plan(updatingArgs)
			Expect(err).To(HaveOccurred(), keyword)
			Expect(output).Should(ContainSubstring(`attribute "%s" must have a known value and may not be changed.`, keyword))
		}

	})

	It("worker disk size is set correctly - [id:69143]",
		ci.Day1Post, ci.Critical, ci.NonHCPCluster,
		func() {
			switch profile.WorkerDiskSize {
			case 0:
				Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(300))
			default:
				Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(profile.WorkerDiskSize))
			}
		})

	It("account roles/policies unified path is correctly set - [id:63138]", ci.Day1Post, ci.Medium, func() {
		unifiedPath, err := rosaAccountRoles.GetPathFromAccountRole(cluster, rosaAccountRoles.AccountRoles[rosaAccountRoles.InstallerAccountRole].Name)
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

		kmsService, err := exe.NewKMSService()
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

	It("imdsv2 value is set correctly - [id:63950]", ci.Day1Post, ci.Critical, ci.NonHCPCluster, func() {
		if profile.Ec2MetadataHttpTokens != "" {
			Expect(string(cluster.AWS().Ec2MetadataHttpTokens())).To(Equal(profile.Ec2MetadataHttpTokens))
		} else {
			Expect(cluster.AWS().Ec2MetadataHttpTokens()).To(Equal(cmv1.Ec2MetadataHttpTokensOptional))
		}
	})
})
