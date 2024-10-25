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
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Verify cluster", func() {
	defer GinkgoRecover()

	var (
		profileHandler profilehandler.ProfileHandler
		profile        profilehandler.ProfileSpec
		cluster        *cmsv1.Cluster
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())
		profile = profileHandler.Profile()

		getResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		cluster = getResp.Body()
	})

	It("proxy is correctly set - [id:67607]", ci.Day1Post, ci.High, func() {
		if !profile.IsProxy() {
			Skip("No proxy is configured for the cluster. skipping the test.")
		}
		getResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(getResp.Body().Proxy().HTTPProxy()).To(ContainSubstring("http://"))
		Expect(getResp.Body().Proxy().HTTPSProxy()).To(ContainSubstring("https://"))
		Expect(getResp.Body().Proxy().NoProxy()).To(Equal("quay.io"))
		Expect(getResp.Body().AdditionalTrustBundle()).To(Equal("REDACTED"))
	})

	It("is successfully installed - [id:63134]", ci.Day1Post, ci.High, func() {
		getResp, err := cms.RetrieveClusterStatus(cms.RHCSConnection, clusterID)
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
		getResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(getResp.Body().Properties()["custom_property"]).To(Equal("test"))

		By("Check rosa_tf_commit is present under cluster's properties")
		Expect(getResp.Body().Properties()["rosa_tf_commit"]).ShouldNot(BeEmpty())

		By("Check rosa_tf_version is present under cluster's properties")
		Expect(getResp.Body().Properties()["rosa_tf_version"]).ShouldNot(BeEmpty())
	})

	It("fips is correctly enabled/disabled - [id:63140]", ci.Day1Post, ci.High, func() {
		Expect(cluster.FIPS()).To(Equal(profile.IsFIPS()))
	})

	It("private_link is correctly enabled/disabled - [id:63133]", ci.Day1Post, ci.High, func() {
		Expect(cluster.AWS().PrivateLink()).To(Equal(profile.IsPrivateLink()))
	})

	It("etcd-encryption is correctly enabled/disabled - [id:63143]", ci.Day1Post, ci.High, func() {
		Expect(cluster.EtcdEncryption()).To(Equal(profile.IsEtcd()))
	})

	It("etcd-encryption key is correctly set - [id:72485]", ci.Day1Post, ci.High, func() {
		if !profile.IsHCP() {
			Skip("This test is for Hosted cluster")
		}
		if !profile.IsEtcd() {
			Skip("Etcd is not activated. skipping the test.")
		}

		By("Retrieve etcd defined key")
		var kmsService exec.KMSService
		var err error
		if profile.IsDifferentEncryptionKeys() {
			kmsService, err = profileHandler.Duplicate().Services().GetKMSService()
			Expect(err).ToNot(HaveOccurred())
		} else {
			kmsService, err = profileHandler.Services().GetKMSService()
			Expect(err).ToNot(HaveOccurred())
		}
		kmsOutput, err := kmsService.Output()
		Expect(err).ToNot(HaveOccurred())
		etcdKey := kmsOutput.KeyARN

		By("Check etcd Encryption key")
		clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(clusterResp.Body().AWS().EtcdEncryption().KMSKeyARN()).To(Equal(etcdKey))
	})

	It("compute_machine_type is correctly set - [id:64023]", ci.Day1Post, ci.Medium, func() {
		computeMachineType := profile.GetComputeMachineType()
		if computeMachineType == "" {
			Skip("No compute_machine_type is configured for the cluster. skipping the test.")
		}
		Expect(cluster.Nodes().ComputeMachineType().ID()).To(Equal(computeMachineType))
	})

	It("compute_replicas is correctly set - [id:73153]", ci.Day1Post, ci.Medium, func() {
		if profile.GetComputeReplicas() <= 0 {
			Skip("No compute_replicas is configured for the cluster. skipping the test.")
		}
		Expect(cluster.Nodes().Compute()).To(Equal(profile.GetComputeReplicas()))
	})

	It("availability zones and multi-az are correctly set - [id:63141]", ci.Day1Post, ci.Medium, func() {
		getResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		zonesArray := strings.Split(profile.GetZones(), ",")
		clusterAvailZones := getResp.Body().Nodes().AvailabilityZones()
		Expect(err).ToNot(HaveOccurred())
		Expect(getResp.Body().MultiAZ()).To(Equal(profile.IsMultiAZ()))

		if profile.GetZones() != "" {
			Expect(clusterAvailZones).To(Equal(helper.JoinStringWithArray(profile.GetRegion(), zonesArray)))
		} else {
			// the default zone for each region
			Expect(clusterAvailZones[0]).To(Equal(fmt.Sprintf("%sa", profile.GetRegion())))
		}
	})

	It("compute_labels are correctly set - [id:68423]", ci.Day1Post, ci.High, func() {
		if profile.IsLabeling() {
			Expect(cluster.Nodes().ComputeLabels()).To(Equal(profilehandler.DefaultMPLabels))
		} else {
			Expect(cluster.Nodes().ComputeLabels()).To(Equal(helper.NilMap))
		}
	})

	It("AWS tags are set - [id:63777]", ci.Day1Post, ci.High, func() {
		buildInTags := map[string]string{
			"red-hat-clustertype": "rosa",
			"red-hat-managed":     "true",
		}

		getResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		if profile.IsTagging() {
			allClusterTags := helper.MergeMaps(buildInTags, profilehandler.Tags)
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
			if profile.IsHCP() {
				Skip("Test can run only on Classic cluster")
			}

			importService, err := profileHandler.Services().GetImportService()
			Expect(err).ToNot(HaveOccurred())

			By("Run the command to import cluster resource")
			defer func() {
				importService.Destroy()
			}()
			resource := "rhcs_cluster_rosa_classic.rosa_sts_cluster_import"
			if profile.IsHCP() {
				resource = "rhcs_cluster_rosa_hcp.rosa_hcp_cluster_import"
			}
			importParam := &exec.ImportArgs{
				ClusterID: clusterID,
				Resource:  resource,
			}
			_, err = importService.Import(importParam)
			Expect(err).ToNot(HaveOccurred())

			By("Check resource state - import command succeeded")
			output, err := importService.ShowState(importParam.Resource)
			Expect(err).ToNot(HaveOccurred())

			// validate import was successful by checking samples fields
			Expect(output).To(ContainSubstring(cluster.Name()))
			Expect(output).To(ContainSubstring(profile.GetRegion()))
			Expect(output).To(ContainSubstring(cluster.Version().ChannelGroup()))

			By("Remove state")
			_, err = importService.RemoveState(resource)
			Expect(err).ToNot(HaveOccurred())

			By("Validate terraform import with no clusterID returns error")
			var unknownClusterID = helper.GenerateRandomStringWithSymbols(20)
			importParam = &exec.ImportArgs{
				ClusterID: unknownClusterID,
				Resource:  resource,
			}

			_, err = importService.Import(importParam)
			Expect(err.Error()).To(ContainSubstring("Cannot import non-existent remote object"))
		})

	It("confirm cluster admin user created ONLY during cluster creation operation - [id:65928]",
		ci.Day1Post, ci.High,
		ci.Exclude,
		func() {
			if !profile.IsAdminEnabled() {
				Skip("The test configured only for cluster admin profile")
			}

			By("List existing Htpasswd IDP")
			idpList, _ := cms.ListClusterIDPs(cms.RHCSConnection, clusterID)
			Expect(idpList.Items().Len()).To(Equal(1))
			Expect(idpList.Items().Slice()[0].Name()).To(Equal("cluster-admin"))

			By("List existing HtpasswdUsers and compare to the created one")
			idpID := idpList.Items().Slice()[0].ID()
			htpasswdUsersList, _ := cms.ListHtpasswdUsers(cms.RHCSConnection, clusterID, idpID)
			Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
			respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
			Expect(respUserName).To(Equal("rhcs-clusteradmin"))

			By("Check resource state file is updated")
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			resource, err := clusterService.GetStateResource("rhcs_cluster_rosa_classic", "rosa_sts_cluster")
			Expect(err).ToNot(HaveOccurred())
			passwordInState, _ := JQ(`.instances[0].attributes.admin_credentials.password`, resource)
			Expect(passwordInState).NotTo(BeEmpty())
			Expect(resource).To(MatchJQ(`.instances[0].attributes.admin_credentials.username`, "rhcs-clusteradmin"))

			By("Login with created cluster admin password")
			getResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			server := getResp.Body().API().URL()

			username := profilehandler.ClusterAdminUser
			password, _ := helper.GetClusterAdminPassword()
			Expect(password).ToNot(BeEmpty())

			ocAtter := &openshift.OcAttributes{
				Server:    server,
				Username:  username,
				Password:  password,
				ClusterID: clusterID,
				AdditionalFlags: []string{
					"--insecure-skip-tls-verify",
					fmt.Sprintf("--kubeconfig %s", path.Join(config.GetKubeConfigDir(), fmt.Sprintf("%s.%s", clusterID, username))),
				},
				Timeout: 10,
			}
			_, err = openshift.OcLogin(*ocAtter)
			Expect(err).ToNot(HaveOccurred())

		})

	It("additional security group are correctly set - [id:69145]",
		ci.Day1Post, ci.Critical,
		func() {
			if profile.IsHCP() {
				Skip("Test can run only on Classic cluster")
			}
			By("Check the profile settings")
			if profile.GetAdditionalSGNumber() == 0 {
				Expect(cluster.AWS().AdditionalComputeSecurityGroupIds()).To(BeEmpty())
				Expect(cluster.AWS().AdditionalControlPlaneSecurityGroupIds()).To(BeEmpty())
				Expect(cluster.AWS().AdditionalInfraSecurityGroupIds()).To(BeEmpty())
			} else {
				By("Verify CMS are using the correct configuration")
				clusterService, err := profileHandler.Services().GetClusterService()
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
	It("additional security group are correctly set - [id:77065]",
		ci.Day1Post, ci.Critical,
		func() {
			if !profile.IsHCP() {
				Skip("Test can run only on Hosted-CP cluster")
			}
			By("Check the profile settings")
			if profile.GetAdditionalSGNumber() == 0 {
				Expect(cluster.AWS().AdditionalComputeSecurityGroupIds()).To(BeEmpty())
			} else {
				By("Verify CMS are using the correct configuration")
				clusterService, err := profileHandler.Services().GetClusterService()
				Expect(err).ToNot(HaveOccurred())
				output, err := clusterService.Output()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(output.AdditionalComputeSecurityGroups)).To(Equal(len(cluster.AWS().AdditionalComputeSecurityGroupIds())))
				for _, sg := range output.AdditionalComputeSecurityGroups {
					Expect(sg).To(BeElementOf(cluster.AWS().AdditionalComputeSecurityGroupIds()))
				}

			}
		})

	It("worker disk size is set correctly - [id:69143]",
		ci.Day1Post, ci.Critical,
		func() {
			switch profile.GetWorkerDiskSize() {
			case 0:
				Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(300))
			default:
				Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(profile.GetWorkerDiskSize()))
			}
		})

	It("account roles/policies unified path is correctly set - [id:63138]", ci.Day1Post, ci.Medium, func() {
		unifiedPath, err := accountroles.GetPathFromAccountRole(cluster, accountroles.AccountRoles[accountroles.InstallerAccountRole].Name)
		Expect(err).ToNot(HaveOccurred())
		Expect(profile.GetUnifiedAccRolesPath()).Should(ContainSubstring(unifiedPath))
	})

	It("customer-managed KMS key is set correctly - [id:63334]", ci.Day1Post, ci.Medium, func() {
		if !profile.IsKMSKey() {
			Skip("The test is configured only for the profile containing the KMS key")
		}
		By("Check the kmsKeyARN")
		listRSresp, err := cms.ListClusterResources(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		kmsService, err := profileHandler.Services().GetKMSService()
		Expect(err).ToNot(HaveOccurred())
		kmsOutput, _ := kmsService.Output()
		expectedKeyArn := kmsOutput.KeyARN
		if profile.IsHCP() {
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
		if profile.IsHCP() {
			Skip("Test can run only on Classic cluster")
		}
		if profile.GetEc2MetadataHttpTokens() != "" {
			Expect(string(cluster.AWS().Ec2MetadataHttpTokens())).To(Equal(profile.GetEc2MetadataHttpTokens()))
		} else {
			Expect(cluster.AWS().Ec2MetadataHttpTokens()).To(Equal(cmsv1.Ec2MetadataHttpTokensOptional))
		}
	})

	It("host prefix and cidrs are set correctly - [id:72466]", ci.Day1Post, ci.Medium, func() {
		By("Retrieve profile information")
		machineCIDR := profile.GetMachineCIDR()
		serviceCIDR := profile.GetServiceCIDR()
		podCIDR := profile.GetPodCIDR()
		hostPrefix := profile.GetHostPrefix()

		if machineCIDR == "" && serviceCIDR == "" && podCIDR == "" && hostPrefix <= 0 {
			Skip("The test is configured only for the profile containing the cidrs and host prefix")
		}

		By("Retrieve cluster detail")
		clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		networkDetail := clusterResp.Body().Network()

		By("Check cluster information")
		Expect(networkDetail.MachineCIDR()).To(Equal(machineCIDR))
		Expect(networkDetail.ServiceCIDR()).To(Equal(serviceCIDR))
		Expect(networkDetail.PodCIDR()).To(Equal(podCIDR))
		Expect(networkDetail.HostPrefix()).To(Equal(hostPrefix))
	})

	It("resources will wait for cluster ready - [id:74096]", ci.Day1Post, ci.Critical,
		func() {
			By("Check if cluster is full resources, if not skip")
			if !profile.IsFullResources() {
				Skip("This only work for full resources testing")
			}

			By("Check the created IDP")
			resp, err := cms.ListClusterIDPs(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Items().Len()).To(Equal(1))
			Expect(resp.Items().Get(0).Name()).To(Equal("full-resource"))
			idpID := resp.Items().Get(0).ID()

			By("Check the group membership")
			userResp, err := cms.ListHtpasswdUsers(cms.RHCSConnection, clusterID, idpID)
			Expect(err).ToNot(HaveOccurred())
			Expect(userResp.Items().Len()).To(Equal(1))
			Expect(userResp.Items().Get(0).Username()).To(Equal("full-resource"))

			By("Check the kubeletconfig")
			if cluster.Hypershift().Enabled() {

				kubeletConfigs, err := cms.ListHCPKubeletConfigs(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(kubeletConfigs)).To(Equal(1))
				Expect(kubeletConfigs[0].PodPidsLimit()).To(Equal(4097))

			} else {
				kubeConfig, err := cms.RetrieveKubeletConfig(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(kubeConfig.PodPidsLimit()).To(Equal(4097))

			}

			By("Check the default ingress")
			ingress, err := cms.RetrieveClusterIngress(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			if !cluster.Hypershift().Enabled() {
				Expect(ingress.ExcludedNamespaces()).To(ContainElement("full-resource"))
			} else {
				Expect(string(ingress.Listening())).To(Equal("internal"))
			}

			By("Check the created autoscaler")
			if !cluster.Hypershift().Enabled() {
				autoscaler, err := cms.RetrieveClusterAutoscaler(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(autoscaler.Body().MaxPodGracePeriod()).To(Equal(1000))
			}

			By("Check the created machinepool")
			mpName := "full-resource"
			if !cluster.Hypershift().Enabled() {
				mp, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, mpName)
				Expect(err).ToNot(HaveOccurred())
				Expect(mp).ToNot(BeNil())
			} else {
				np, err := cms.RetrieveNodePool(cms.RHCSConnection, clusterID, mpName)
				Expect(err).ToNot(HaveOccurred())
				Expect(np).ToNot(BeNil())
			}

			By("Check the created tuningconfig")
			tunningConfigName := "full-resource"
			if cluster.Hypershift().Enabled() {
				tunnings, err := cms.ListTuningConfigs(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(tunnings.Total()).To(Equal(1))
				Expect(tunnings.Items().Get(0).Name()).To(Equal(tunningConfigName))
			}

		})

	It("multiarch is set correctly - [id:75107]", ci.Day1Post, ci.High, func() {
		By("Verify cluster configuration")
		clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		if profile.IsHCP() {
			Expect(clusterResp.Body().MultiArchEnabled()).To(BeTrue())
		} else {
			Expect(clusterResp.Body().MultiArchEnabled()).To(BeFalse())
		}
	})

	It("imdsv2 is set correctly - [id:75372]", ci.Day1Post, ci.Critical, func() {
		if !profile.IsHCP() {
			Skip("Test can run only on HCP cluster")
		}

		By("Check the cluster description value to match cluster profile configuration")
		if profile.GetEc2MetadataHttpTokens() != "" {
			Expect(string(cluster.AWS().Ec2MetadataHttpTokens())).
				To(Equal(profile.GetEc2MetadataHttpTokens()))
		} else {
			Expect(string(cluster.AWS().Ec2MetadataHttpTokens())).
				To(Equal(constants.DefaultEc2MetadataHttpTokens))
		}

		By("Get the default workers machinepool details")
		npList, err := cms.ListNodePools(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(npList).ToNot(BeEmpty())

		for _, np := range npList {
			Expect(np.ID()).ToNot(BeNil())
			if strings.HasPrefix(np.ID(), constants.DefaultNodePoolName) {
				By("Get the details of the nodepool")
				npRespBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, np.ID())
				Expect(err).ToNot(HaveOccurred())

				By("Check the default workers machinepool value to match cluster level spec attribute")
				Expect(string(npRespBody.AWSNodePool().Ec2MetadataHttpTokens())).
					To(Equal(string(cluster.AWS().Ec2MetadataHttpTokens())))
			}
		}
	})

	It("registry config is set correctly - [id:76499]",
		ci.Day1Post, ci.Critical, ci.FeatureClusterRegistryConfig,
		func() {
			if !profile.IsHCP() {
				Skip("Test can run only on Hosted cluster")
			}

			if !profile.IsUseRegistryConfig() {
				Skip("Registry Config is not configured on this clusters")
			}

			clusterRegistryConfig := cluster.RegistryConfig()
			dftRegistryConfig := exec.GetDefaultRegistryConfig()

			By("Check registry sources")
			Expect(clusterRegistryConfig.RegistrySources()).ToNot(BeNil())
			registries := profile.GetAllowedRegistries()
			if registries == nil {
				registries = []string{}
			}
			Expect(clusterRegistryConfig.RegistrySources().AllowedRegistries()).To(Equal(registries))
			registries = profile.GetBlockedRegistries()
			if registries == nil {
				registries = []string{}
			}
			Expect(clusterRegistryConfig.RegistrySources().BlockedRegistries()).To(Equal(registries))
			registries = *dftRegistryConfig.RegistrySources.InsecureRegistries
			if registries == nil {
				registries = []string{}
			}
			Expect(clusterRegistryConfig.RegistrySources().InsecureRegistries()).To(Equal(registries))

			By("Check allowed registries for import")
			var resultAllowedRegistriesForImport []exec.AllowedRegistryForImport
			for _, registry := range clusterRegistryConfig.AllowedRegistriesForImport() {
				resultAllowedRegistriesForImport = append(resultAllowedRegistriesForImport, exec.GetAllowedRegistryForImport(registry.DomainName(), registry.Insecure()))
			}
			Expect(resultAllowedRegistriesForImport).To(Equal(*dftRegistryConfig.AllowedRegistriesForImport))

			By("Check additional trusted CA")
			var ca map[string]string
			if dftRegistryConfig.AdditionalTrustedCA != nil {
				ca = *dftRegistryConfig.AdditionalTrustedCA
			}
			Expect(clusterRegistryConfig.AdditionalTrustedCa()).To(Equal(ca))

			if dftRegistryConfig.PlatformAllowlistID != nil {
				By("Check Platform Allowlist")
				Expect(clusterRegistryConfig.PlatformAllowlist().Registries()).To(Equal(dftRegistryConfig.PlatformAllowlistID))
			}
		})
})
