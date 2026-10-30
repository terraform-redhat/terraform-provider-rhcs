// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmsv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	sprest "github.com/openshift-online/rosa-seaplane-api/clientset/rest"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Edit cluster", ci.Day2, func() {

	var profileHandler profilehandler.ProfileHandler
	var clusterService exec.ClusterService
	var clusterArgs *exec.ClusterArgs
	var originalClusterArgs *exec.ClusterArgs

	retrieveClusterArgs := func() (args *exec.ClusterArgs) {
		var err error
		args, err = clusterService.ReadTFVars()
		Expect(err).ShouldNot(HaveOccurred())
		return args
	}

	BeforeEach(func() {
		var err error
		By("Load profile")
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ShouldNot(HaveOccurred())
		if profileHandler.Profile().GetClusterType() == constants.SEAPLANE {
			Skip("Seaplane cluster edits use Seaplane-specific workflows")
		}

		// Initialize the cluster service
		By("Create cluster service")
		clusterService, err = profileHandler.Services().GetClusterService()
		Expect(err).ShouldNot(HaveOccurred())

		By("Retrieve current cluster args")
		clusterArgs = retrieveClusterArgs()
		originalClusterArgs = retrieveClusterArgs()
	})

	AfterEach(func() {
		if clusterService != nil && originalClusterArgs != nil {
			clusterService.Apply(originalClusterArgs)
		}
	})

	Context("can edit/delete", func() {
		It("delete_protection - [id:66500]", ci.High, ci.FeatureClusterDeleteProtection, func() {
			DeferCleanup(func() {
				By("Ensure delete_protection is disabled for teardown")
				clusterArgs.DeleteProtection = new(false)
				_, err := clusterService.Apply(clusterArgs)
				Expect(err).ShouldNot(HaveOccurred())
			})

			By("Enable delete_protection")
			clusterArgs.DeleteProtection = new(true)
			_, err := clusterService.Apply(clusterArgs)
			Expect(err).ShouldNot(HaveOccurred())

			By("Verify delete_protection is enabled via OCM API")
			enabled, err := cms.RetrieveClusterDeleteProtection(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeTrue())

			By("Verify destroy fails when delete_protection is enabled")
			_, err = clusterService.Destroy()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Can't delete cluster"))

			By("Disable delete_protection")
			clusterArgs.DeleteProtection = new(false)
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ShouldNot(HaveOccurred())

			By("Verify delete_protection is disabled via OCM API")
			enabled, err = cms.RetrieveClusterDeleteProtection(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeFalse())
		})

		It("registry config - [id:76500]", ci.High, ci.FeatureClusterRegistryConfig, func() {
			if !profileHandler.Profile().IsHCP() {
				Skip("Test can run only on Hosted cluster")
			}

			getCMSClusterRegistryConfig := func() *cmsv1.ClusterRegistryConfig {
				resp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				return resp.Body().RegistryConfig()
			}

			By("Edit the cluster with allowed_registries and blocked_registries is empty")
			registries := helper.GetRegistries(8088)
			clusterArgs.RegistryConfig.RegistrySources.AllowedRegistries = new(registries)
			clusterArgs.RegistryConfig.RegistrySources.BlockedRegistries = nil
			_, err := clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
			registryConfig := getCMSClusterRegistryConfig()
			Expect(registryConfig.RegistrySources().AllowedRegistries()).To(Equal(registries))
			Expect(len(registryConfig.RegistrySources().BlockedRegistries())).To(Equal(0))

			By("Edit the cluster with blocked_registries and allowed_registries is empty")
			registries = helper.GetRegistries(8089)
			clusterArgs.RegistryConfig.RegistrySources.AllowedRegistries = nil
			clusterArgs.RegistryConfig.RegistrySources.BlockedRegistries = new(registries)
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
			registryConfig = getCMSClusterRegistryConfig()
			Expect(registryConfig.RegistrySources().BlockedRegistries()).To(Equal(registries))
			Expect(len(registryConfig.RegistrySources().AllowedRegistries())).To(Equal(0))

			By("Edit insecure registries")
			registries = helper.GetRegistries(8090)
			clusterArgs.RegistryConfig.RegistrySources.InsecureRegistries = new(registries)
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
			registryConfig = getCMSClusterRegistryConfig()
			Expect(registryConfig.RegistrySources().InsecureRegistries()).To(Equal(registries))

			By("Edit platform allowlist id")
			alResp, err := cms.ListRegistryAllowlists(cms.RHCSConnection)
			Expect(err).ToNot(HaveOccurred())
			if alResp.Size() > 0 {
				allowListID := alResp.Items().Slice()[0].ID()
				clusterArgs.RegistryConfig.PlatformAllowlistID = new(allowListID)
				_, err = clusterService.Apply(clusterArgs)
				Expect(err).ToNot(HaveOccurred())
				registryConfig = getCMSClusterRegistryConfig()
				Expect(registryConfig.PlatformAllowlist().ID()).To(Equal(allowListID))
			} else {
				Logger.Info("No allowlist platform ID available for testing")
			}

			By("Edit platform additional trust ca")
			registry := "10.0.0.0..8091"
			trustCAs := map[string]string{}
			trustCAs[registry], err = helper.CreatePEMCertificate()
			Expect(err).ToNot(HaveOccurred())
			clusterArgs.RegistryConfig.AdditionalTrustedCA = &trustCAs
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
			registryConfig = getCMSClusterRegistryConfig()
			Expect(registryConfig.AdditionalTrustedCa()[registry]).ToNot(BeEmpty())

			By("Edit allowed registries for import")
			registry = helper.GetRegistry(8092)
			allowedRegistriesForImport := []exec.AllowedRegistryForImport{
				exec.GetAllowedRegistryForImport(registry, true),
			}
			clusterArgs.RegistryConfig.AllowedRegistriesForImport = &allowedRegistriesForImport
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
			registryConfig = getCMSClusterRegistryConfig()
			var resultAllowedRegistriesForImport []exec.AllowedRegistryForImport
			for _, registry := range registryConfig.AllowedRegistriesForImport() {
				resultAllowedRegistriesForImport = append(resultAllowedRegistriesForImport, exec.GetAllowedRegistryForImport(registry.DomainName(), registry.Insecure()))
			}
			Expect(resultAllowedRegistriesForImport).To(Equal(allowedRegistriesForImport))

			By("Remove all")
			clusterArgs.RegistryConfig.RegistrySources.AllowedRegistries = nil
			clusterArgs.RegistryConfig.RegistrySources.BlockedRegistries = nil
			clusterArgs.RegistryConfig.RegistrySources.InsecureRegistries = nil
			clusterArgs.RegistryConfig.AllowedRegistriesForImport = nil
			clusterArgs.RegistryConfig.AdditionalTrustedCA = nil
			clusterArgs.RegistryConfig.PlatformAllowlistID = nil
		})
	})

	Context("validate", func() {
		BeforeEach(func() {
			if !profileHandler.Profile().IsHCP() {
				Skip("Test can run only on Hosted cluster")
			}
		})

		validateClusterArg := func(updateFields func(args *exec.ClusterArgs), validateFunc func(output string, err error)) {
			updateFields(clusterArgs)
			applyOutput, err := clusterService.Apply(clusterArgs)
			validateFunc(applyOutput, err)
			clusterArgs = retrieveClusterArgs()
		}
		validateClusterArgAgainstErrorSubstrings := func(updateFields func(args *exec.ClusterArgs), errSubStrings ...string) {
			validateClusterArg(updateFields, func(output string, err error) {
				GinkgoHelper()
				Expect(err).To(HaveOccurred())
				Expect(output).NotTo(BeEmpty())
				for _, errStr := range errSubStrings {
					helper.ExpectTFErrorContains(err, errStr)
				}
			})
		}

		It("required fields - [id:72452]", ci.Medium, ci.FeatureClusterDefault, func() {
			By("Try to edit aws account with wrong value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAccountID = new("another_account")
			}, "Attribute aws_account_id aws account ID must be only digits and exactly 12")

			By("Try to edit aws account with wrong account")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAccountID = new("000000000000")
			}, "Attribute aws_account_id, cannot be changed from")

			By("Try to edit billing account with wrong value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSBillingAccountID = new("anything")
			}, "Attribute aws_billing_account_id aws billing account ID must be only digits and exactly 12 in length")

			By("Try to edit billing account with wrong account")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSBillingAccountID = new("000000000000")
			}, "billing account 000000000000 not linked to organization")

			By("Try to edit cloud region")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				region := "us-east-1"
				if profileHandler.Profile().GetRegion() == region {
					region = "us-west-2" // make sure we are not in the same region
				}
				args.AWSRegion = new(region)
			}, "Invalid AZ")

			By("Try to edit cloud region and Availability zone(s)")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				region := "us-east-1"
				azs := []string{"us-east-1a"}
				if profileHandler.Profile().GetRegion() == region {
					region = "us-west-2" // make sure we are not in the same region
					azs = []string{"us-west-2b"}
				}
				args.AWSRegion = new(region)
				args.AWSAvailabilityZones = new(azs)
			}, "Attribute cloud_region, cannot be changed from")

			By("Try to edit name")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ClusterName = new("any_name")
			}, "Attribute name, cannot be changed from")
		})

		It("compute fields - [id:72453]", ci.Medium, ci.FeatureClusterCompute, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			By("Try to edit compute machine type")
			machineType := "m5.2xlarge"
			if clusterResp.Body().Nodes().ComputeMachineType().ID() == machineType {
				machineType = "m5.xlarge"
			}
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ComputeMachineType = new(machineType)
			}, "Attribute compute_machine_type, cannot be changed from")

			if profileHandler.Profile().IsAutoscaling() {
				By("Try to set replicas while autoscaling is enabled")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Replicas = new(5)
				}, "When autoscaling is enabled, replicas should not be configured")
			} else {
				By("Edit replicas (create-only: sync state and warn)")
				initialCompute := clusterResp.Body().Nodes().Compute()
				newReplicas := 5
				validateClusterArg(func(args *exec.ClusterArgs) {
					args.Replicas = new(newReplicas)
				}, func(output string, err error) {
					Expect(err).ToNot(HaveOccurred())
					Expect(output).To(ContainSubstring("Warning: Cluster replicas is create-only"))
					if profileHandler.Profile().IsHCP() {
						Expect(output).To(ContainSubstring("rhcs_hcp_machine_pool"))
						resource, stateErr := clusterService.GetStateResource(
							"rhcs_cluster_rosa_hcp", "rosa_hcp_cluster",
						)
						Expect(stateErr).ToNot(HaveOccurred())
						Expect(helper.DigInt(resource, "instances", 0, "attributes", "replicas")).
							To(Equal(newReplicas))
					} else {
						Expect(output).To(ContainSubstring("rhcs_machine_pool"))
						resource, stateErr := clusterService.GetStateResource(
							"rhcs_cluster_rosa_classic", "rosa_sts_cluster",
						)
						Expect(stateErr).ToNot(HaveOccurred())
						Expect(helper.DigInt(resource, "instances", 0, "attributes", "replicas")).
							To(Equal(newReplicas))
					}
					afterResp, afterErr := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
					Expect(afterErr).ToNot(HaveOccurred())
					Expect(afterResp.Body().Nodes().Compute()).To(Equal(initialCompute))
				})

				By("Try to edit with replicas < 2")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Replicas = new(1)
				}, "Attribute replicas value must be at least 2")
			}
		})

		It("properties - [id:72455]", ci.Medium, ci.FeatureClusterMisc, func() {
			By("Retrieve current properties")
			out, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			currentProperties := out.Properties

			By("Try to edit `rosa_creator_arn` property")
			props := helper.CopyStringMap(currentProperties)
			props["rosa_creator_arn"] = "anything"
			validateClusterArg(func(args *exec.ClusterArgs) {
				args.CustomProperties = new(props)
			}, func(output string, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring("Warning: Shouldn't patch cluster properties"))
				clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterResp.Body().Properties()["rosa_creator_arn"]).To(Equal(props["rosa_creator_arn"]))
			})

			By("Try to remove `rosa_creator_arn` property")
			currentProperties = out.Properties
			props = helper.CopyStringMap(currentProperties)
			delete(props, "rosa_creator_arn")
			validateClusterArg(func(args *exec.ClusterArgs) {
				args.CustomProperties = new(props)
				args.IncludeCreatorProperty = new(false)
			}, func(output string, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeEmpty())
				clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterResp.Body().Properties()["rosa_creator_arn"]).To(BeEmpty())
			})

			By("Try to edit `reserved` property")
			props = helper.CopyStringMap(currentProperties)
			props["rosa_tf_version"] = "any"
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.CustomProperties = new(props)
			}, "Can not override reserved properties keys. rosa_tf_version is a reserved")
		})

		It("network fields - [id:72470]", ci.Medium, ci.FeatureClusterNetwork, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			originalAZs := clusterResp.Body().Nodes().AvailabilityZones()

			By("Try to edit availability zones")
			azs := []string{originalAZs[0]}
			// TODO report issue ? This should fail
			// validateClusterArgErrorSubstring(func(args *exec.ClusterArgs) {
			// 	args.AWSAvailabilityZones = helper.StringSlicePointer(azs)
			// }, "Attribute availability_zones, cannot be changed from")
			validateClusterArg(func(args *exec.ClusterArgs) {
				args.AWSAvailabilityZones = new(azs)
			}, func(output string, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeEmpty())
				clusterResp, err = cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterResp.Body().Nodes().AvailabilityZones()).To(Equal(originalAZs))
			})

			By("Try to edit subnet ids")
			azs = []string{"subnet-1", "subnet-2"}
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAvailabilityZones = new(azs)
			}, "Invalid AZ")

			By("Try to edit host prefix")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.HostPrefix = new(25)
			}, "Attribute host_prefix, cannot be changed from")

			By("Try to edit machine cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.MachineCIDR = new("10.0.0.0/17")
			}, "Attribute machine_cidr, cannot be changed from")

			By("Try to edit service cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ServiceCIDR = new("172.50.0.0/20")
			}, "Attribute service_cidr, cannot be changed from")

			By("Try to edit pod cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.PodCIDR = new("10.128.0.0/16")
			}, "Attribute pod_cidr, cannot be changed from")
		})

		It("version fields - [id:72478]", ci.Medium, ci.FeatureClusterNetwork, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			currentVersion := clusterResp.Body().Version().RawID()
			currentChannelGroup := profileHandler.Profile().GetChannelGroup()

			skipChannelGroups := []string{constants.VersionCandidateChannel, constants.VersionNightlyChannel}
			if slices.Contains(skipChannelGroups, currentChannelGroup) {
				Skip(fmt.Sprintf("There are versions in the %s channel that don't exist in other channels", currentChannelGroup))
			}

			By("Get new channel group")
			otherChannelGroup := constants.VersionFastChannel
			if currentChannelGroup == constants.VersionFastChannel {
				otherChannelGroup = constants.VersionCandidateChannel
			}

			By("Retrieve latest version")
			versions := cms.SortVersions(cms.HCPEnabledVersions(cms.RHCSConnection, otherChannelGroup))
			lastVersion := versions[len(versions)-1]

			if lastVersion.RawID != currentVersion && currentChannelGroup != constants.VersionCandidateChannel {
				By("Try to edit version to one from another channel_group")
				errString := fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, desired version (%s) is not in the list of available upgrades", clusterID, lastVersion.RawID)
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.OpenshiftVersion = new(lastVersion.RawID)
				}, errString)
			}

			By("Edit channel_group")
			validateClusterArg(func(args *exec.ClusterArgs) {
				args.ChannelGroup = new(otherChannelGroup)
			}, func(output string, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeEmpty())
				clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterResp.Body().Version().ChannelGroup()).To(Equal(otherChannelGroup))
			})

			By("Try to edit channel_group to a non-existing value")
			otherChannelGroup = "invalidChannelGroup"
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ChannelGroup = new(otherChannelGroup)
			}, fmt.Sprintf("Current cluster version %s is not available in channel group %s", currentVersion, otherChannelGroup))
		})

		It("private fields - [id:72480]", ci.Medium, ci.FeatureClusterPrivate, func() {
			By("Try to edit private")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Private = new(!profileHandler.Profile().IsPrivate())
			}, "Attribute private, cannot be changed from")
		})

		It("imdsv2 fields - [id:75414]", ci.Medium, ci.FeatureClusterIMDSv2, func() {
			By("Try to edit ec2_metadata_http_tokens value")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			currentHttpToken := string(clusterResp.Body().AWS().Ec2MetadataHttpTokens())
			otherHttpToken := constants.RequiredEc2MetadataHttpTokens
			if currentHttpToken == constants.RequiredEc2MetadataHttpTokens {
				otherHttpToken = constants.OptionalEc2MetadataHttpTokens
			}
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Ec2MetadataHttpTokens = new(otherHttpToken)
			})
		})

		It("encryption fields - [id:72487]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit etcd_encryption")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Etcd = new(!profileHandler.Profile().IsEtcd())
			}, "Attribute etcd_encryption, cannot be changed from")

			By("Try to edit etcd_kms_key_arn")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.EtcdKmsKeyARN = new("anything")
			}, "Attribute etcd_kms_key_arn, cannot be changed from")

			By("Try to edit kms_key_arn")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.KmsKeyARN = new("anything")
			}, "Attribute kms_key_arn, cannot be changed from")
		})

		It("sts fields - [id:72497]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit oidc config")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OIDCConfigID = new("2a4rv4o76gljek6c3po16abquaciv0a7")
			}, "Attribute sts.oidc_config_id, cannot be changed from")

			By("Try to edit installer role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsInstallerRole = new("anything")
			}, "Attribute sts.role_arn, cannot be changed from")

			By("Try to edit support role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsSupportRole = new("anything")
			}, "Attribute sts.support_role_arn, cannot be changed from")

			By("Try to edit worker role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsWorkerRole = new("anything")
			}, "Attribute sts.instance_iam_roles.worker_role_arn, cannot be changed from")

			By("Try to edit operator role prefix")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OperatorRolePrefix = new("anything")
			}, "Attribute sts.operator_role_prefix, cannot be changed from")
		})

		It("proxy fields - [id:72490]", ci.Medium, ci.FeatureClusterProxy, func() {
			By("Edit proxy with wrong http_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy: new("aaaaxxxx"),
				}
			}, "Invalid 'proxy.http_proxy' attribute 'aaaaxxxx'")

			By("Edit proxy with http_proxy starts with https")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy: new("https://anything"),
				}
			}, "'proxy.http_proxy' prefix is not 'http'")

			By("Edit proxy with invalid https_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPSProxy: new("aaavvv"),
				}
			},
				"Invalid",
				"'proxy.https_proxy' attribute 'aaavvv'",
			)

			By("Edit proxy with invalid additional_trust_bundle set")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					AdditionalTrustBundle: new("wrong value"),
				}
			}, "Failed to parse")

			By("Edit proxy with http-proxy and http-proxy empty but no-proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy:  helper.EmptyStringPointer,
					HTTPSProxy: helper.EmptyStringPointer,
					NoProxy:    new("test.com"),
				}
			}, "Cannot set 'proxy.no_proxy' attribute 'test.com' while removing 'proxy.http_proxy'")

			By("Edit proxy with with invalid no_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					NoProxy: new("*"),
				}
			}, "no-proxy value: '*' should match the regular expression")
		})

		It("tags - [id:72628]", ci.Medium, ci.FeatureClusterMisc, func() {
			By("Retrieve current tags")
			out, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			currentTags := out.UserTags

			By("Try to edit tags")
			tags := helper.CopyStringMap(currentTags)
			var firstKey string
			// Remove first key
			for k := range tags {
				firstKey = k
				break
			}
			if firstKey != "" {
				delete(tags, firstKey)
			}
			// Edit second key
			for k := range tags {
				firstKey = k
				break
			}
			if firstKey != "" {
				tags[firstKey] = "newValue"
			}
			// Add new key
			tags["newTag"] = "appendTag"

			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Tags = new(tags)
			}, "Attribute tags, cannot be changed from")
		})

		// Skip this tests until OCM-5079 fixed
		It("security groups - [id:69145]",
			ci.Exclude, ci.Day2,
			func() {
				if profileHandler.Profile().IsHCP() {
					Skip("Test can run only on Classic cluster")
				}
				clusterService, err := profileHandler.Services().GetClusterService()
				Expect(err).ToNot(HaveOccurred())
				output, err := clusterService.Output()
				Expect(err).ToNot(HaveOccurred())
				args := map[string]*exec.ClusterArgs{
					"aws_additional_compute_security_group_ids": {
						AdditionalComputeSecurityGroups:      new(output.AdditionalComputeSecurityGroups[0:1]),
						AdditionalInfraSecurityGroups:        new(output.AdditionalInfraSecurityGroups),
						AdditionalControlPlaneSecurityGroups: new(output.AdditionalControlPlaneSecurityGroups),
						AWSRegion:                            new(profileHandler.Profile().GetRegion()),
					},
					"aws_additional_infra_security_group_ids": {
						AdditionalInfraSecurityGroups:        new(output.AdditionalInfraSecurityGroups[0:1]),
						AdditionalComputeSecurityGroups:      new(output.AdditionalComputeSecurityGroups),
						AdditionalControlPlaneSecurityGroups: new(output.AdditionalControlPlaneSecurityGroups),
						AWSRegion:                            new(profileHandler.Profile().GetRegion()),
					},
					"aws_additional_control_plane_security_group_ids": {
						AdditionalControlPlaneSecurityGroups: new(output.AdditionalControlPlaneSecurityGroups[0:1]),
						AdditionalComputeSecurityGroups:      new(output.AdditionalComputeSecurityGroups),
						AdditionalInfraSecurityGroups:        new(output.AdditionalInfraSecurityGroups),
						AWSRegion:                            new(profileHandler.Profile().GetRegion()),
					},
				}
				for keyword, updatingArgs := range args {
					_, err := clusterService.Apply(updatingArgs)
					Expect(err).To(HaveOccurred(), keyword)
					Expect(err.Error()).Should(ContainSubstring(`Attribute value cannot be changed`))
				}
			})

		It("registry config - [id:76501]", ci.Medium, ci.FeatureClusterRegistryConfig, func() {
			if !profileHandler.Profile().IsHCP() {
				Skip("Test can run only on Hosted cluster")
			}
			registry := helper.GetRegistry(8090)
			registries := helper.GetRegistries(8090, 8091)
			duplicatedRegistries := []string{
				registry,
				registry,
			}
			duplicatedRegistriesErrMsg := fmt.Sprintf("duplicated registry '%s'", registry)

			By("Allowed and blocked registries set at the same time")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				registries := helper.GetRegistries(8090, 8091)
				args.RegistryConfig.RegistrySources.AllowedRegistries = &registries
				args.RegistryConfig.RegistrySources.BlockedRegistries = &registries
			}, "Attribute \"registry_config.registry_sources.allowed_registries\" cannot be specified and be not empty "+
				"when \"registry_config.registry_sources.blocked_registries\" is specified")

			By("Blocked registries and insecure registries have same value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.RegistryConfig.RegistrySources.AllowedRegistries = nil
				args.RegistryConfig.RegistrySources.BlockedRegistries = &registries
				args.RegistryConfig.RegistrySources.InsecureRegistries = &registries
			}, "Insecure registries should not include registries already present in blocked registries, found duplicated")

			By("Allowed registries have duplicate value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.RegistryConfig.RegistrySources.AllowedRegistries = &duplicatedRegistries
				args.RegistryConfig.RegistrySources.BlockedRegistries = nil
			}, duplicatedRegistriesErrMsg)

			By("Blocked registries have duplicate value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.RegistryConfig.RegistrySources.AllowedRegistries = nil
				args.RegistryConfig.RegistrySources.BlockedRegistries = &duplicatedRegistries
			}, duplicatedRegistriesErrMsg)

			By("Insecure registries have duplicate value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.RegistryConfig.RegistrySources.InsecureRegistries = &duplicatedRegistries
			}, duplicatedRegistriesErrMsg)

			By("Allowed Registries For Import have duplicate value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.RegistryConfig.AllowedRegistriesForImport = &[]exec.AllowedRegistryForImport{
					exec.GetAllowedRegistryForImport(registry, true),
					exec.GetAllowedRegistryForImport(registry, false),
				}
			}, fmt.Sprintf("Duplicate domain '%s' in AllowedRegistriesForImport", registry))

			By("Platform allowlist does not exist")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.RegistryConfig.PlatformAllowlistID = new("anything")
			}, "Allowlist with id 'anything' not found")

			By("Additional Trust CA is invalid")
			trustedCARegistry := "10.0.0.0..8090"
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				trustedCA := map[string]string{}
				trustedCA[trustedCARegistry] = "invalid"
				args.RegistryConfig.AdditionalTrustedCA = &trustedCA
			}, fmt.Sprintf("failed to parse CA bundle for registry '%s'", trustedCARegistry))
		})
	})

	Context("work for", func() {
		It("autoscaling change - [id:63147]",
			ci.Medium,
			func() {
				By("Update the cluster autoscaling")
				if profileHandler.Profile().IsAutoscaling() {
					By("Change autoscaling_enabled from true to false")
					clusterArgs.Autoscaling = new(false)
					_, err := clusterService.Apply(clusterArgs)
					Expect(err).To(HaveOccurred())
					helper.ExpectTFErrorContains(err, "Attribute autoscaling_enabled, cannot be changed")

					By("Change max_replicas")
					clusterArgs.Autoscaling = new(true)
					clusterArgs.MaxReplicas = new(9)
					_, err = clusterService.Apply(clusterArgs)
					Expect(err).To(HaveOccurred())
					helper.ExpectTFErrorContains(err, " Attribute max_replicas, cannot be changed")

					By("Change min_replicas")
					clusterArgs.Autoscaling = new(true)
					clusterArgs.MaxReplicas = originalClusterArgs.MaxReplicas
					clusterArgs.MinReplicas = new(3)
					_, err = clusterService.Apply(clusterArgs)
					Expect(err).To(HaveOccurred())
					helper.ExpectTFErrorContains(err, " Attribute min_replicas, cannot be changed")
				} else {
					By("Change autoscaling_enabled from failse to true")
					clusterArgs.Autoscaling = new(true)
					_, err := clusterService.Apply(clusterArgs)
					Expect(err).To(HaveOccurred())
					helper.ExpectTFErrorContains(err, "Attribute autoscaling_enabled, cannot be changed")

					By("Change replicas (create-only: sync state and warn)")
					clusterArgs.Autoscaling = originalClusterArgs.Autoscaling
					newReplicas := 9
					clusterArgs.Replicas = new(newReplicas)
					applyOut, err := clusterService.Apply(clusterArgs)
					Expect(err).ToNot(HaveOccurred())
					Expect(applyOut).To(ContainSubstring("Warning: Cluster replicas is create-only"))
					if profileHandler.Profile().IsHCP() {
						Expect(applyOut).To(ContainSubstring("rhcs_hcp_machine_pool"))
						resource, stateErr := clusterService.GetStateResource(
							"rhcs_cluster_rosa_hcp", "rosa_hcp_cluster",
						)
						Expect(stateErr).ToNot(HaveOccurred())
						Expect(helper.DigInt(resource, "instances", 0, "attributes", "replicas")).
							To(Equal(newReplicas))
					} else {
						Expect(applyOut).To(ContainSubstring("rhcs_machine_pool"))
						resource, stateErr := clusterService.GetStateResource(
							"rhcs_cluster_rosa_classic", "rosa_sts_cluster",
						)
						Expect(stateErr).ToNot(HaveOccurred())
						Expect(helper.DigInt(resource, "instances", 0, "attributes", "replicas")).
							To(Equal(newReplicas))
					}
				}
			})
	})

})

var sharedClusterProxyAWSRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

const (
	sharedClusterProxyPollInterval          = 30 * time.Second
	sharedClusterProxyClusterDeletedTimeout = 60 * time.Minute
	sharedClusterProxyTeardownGracePeriod   = 90 * time.Minute
)

var _ = Describe("Shared cluster proxy workflow", ci.Day2, ci.FeatureClusterProxy, func() {
	It("proxy - [id:72489]", ci.High, ci.SeaplaneValidated, func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if clusterID == "" {
			Skip("an existing cluster ID is required for the proxy day-2 workflow")
		}
		if !profile.Profile().IsProxy() {
			Skip("No proxy configured")
		}

		clusterType := profile.Profile().GetClusterType()
		switch clusterType {
		case constants.SEAPLANE:
			runSeaplaneClusterProxyWorkflow(ctx, profile)
		case constants.ROSA_CLASSIC, constants.ROSA_HCP:
			runClassicClusterProxyWorkflow(profile)
		default:
			Skip(fmt.Sprintf("proxy workflow is not supported for cluster type %q", clusterType.Name))
		}
	})
})

func runSeaplaneClusterProxyWorkflow(ctx SpecContext, profile profilehandler.ProfileHandler) {
	seaplaneURL := sharedClusterProxyRequiredEnv("SEAPLANE_URL")
	clusterName := os.Getenv("SEAPLANE_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = config.GetRHCSClusterName()
	}
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	Expect(clusterName).NotTo(BeEmpty(), "a Seaplane cluster name is required")
	awsRegion := config.GetRegion()
	if awsRegion == "" {
		awsRegion = profile.Profile().GetRegion()
	}
	if awsRegion == "" {
		awsRegion = sharedClusterProxyAWSRegionRE.FindString(seaplaneURL)
	}
	Expect(awsRegion).NotTo(BeEmpty(), "cannot determine the AWS region from the profile or SEAPLANE_URL")
	workspace := clusterName

	clusterSvc, err := exec.NewSeaplaneClusterService(workspace)
	Expect(err).NotTo(HaveOccurred())
	clusterOutput, err := clusterSvc.Output()
	Expect(err).NotTo(HaveOccurred())
	Expect(clusterOutput.ClusterID).NotTo(BeEmpty(), fmt.Sprintf("proxy checks require existing cluster Terraform state in workspace %q", workspace))
	Expect(clusterOutput.ClusterID).To(Equal(clusterID), "the cluster Terraform state must match the shared E2E fixture")
	clusterArgs, err := clusterSvc.ReadTFVars()
	Expect(err).NotTo(HaveOccurred())
	originalArgs := *clusterArgs
	Expect(clusterArgs.ClusterName).NotTo(BeNil(), "proxy checks require the cluster name for Platform API lookups")
	Expect(*clusterArgs.ClusterName).To(Equal(clusterName))
	runProxyValidationChecks := strings.EqualFold(os.Getenv("SEAPLANE_PROXY_VALIDATION_TEST"), "true")
	invalidHTTPProxy := "http://proxy.example.com/" + strings.Repeat("x", 2049)

	vpcSvc, err := exec.NewSeaplaneVPCService(workspace)
	Expect(err).NotTo(HaveOccurred())
	vpcOutput, err := vpcSvc.Output()
	Expect(err).NotTo(HaveOccurred())
	Expect(vpcOutput.VPCID).NotTo(BeEmpty())
	Expect(vpcOutput.PublicSubnetID).NotTo(BeEmpty())

	proxyWorkspace := workspace + "-proxy-sanity"
	proxySvc, err := exec.NewProxyService(proxyWorkspace, constants.SEAPLANE)
	Expect(err).NotTo(HaveOccurred())
	proxyTrustBundlePath := filepath.Join(config.GetRHCSOutputDir(), proxyWorkspace+"-ca.cert")
	var proxyAttempted bool
	var clusterProxyChangeAttempted bool

	seaplaneClient := buildSharedProxySeaplaneClient(seaplaneURL, awsRegion)
	clusters := seaplaneClient.SeaplaneV1alpha1().Clusters()
	currentCluster, err := clusters.Get(ctx, clusterName, spplatform.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	originalArgs.HTTPProxy = nil
	originalArgs.HTTPSProxy = nil
	originalArgs.NoProxy = nil
	if currentCluster.Spec.HostedCluster.Configuration != nil && currentCluster.Spec.HostedCluster.Configuration.Proxy != nil {
		httpProxy := currentCluster.Spec.HostedCluster.Configuration.Proxy.HTTPProxy
		httpsProxy := currentCluster.Spec.HostedCluster.Configuration.Proxy.HTTPSProxy
		noProxy := currentCluster.Spec.HostedCluster.Configuration.Proxy.NoProxy
		originalArgs.HTTPProxy = &httpProxy
		originalArgs.HTTPSProxy = &httpsProxy
		originalArgs.NoProxy = &noProxy
	}
	verifyClusterProxy := func(waitCtx context.Context, wantHTTP, wantHTTPS, wantNo string) {
		Eventually(func() bool {
			cluster, getErr := clusters.Get(waitCtx, clusterName, spplatform.GetOptions{})
			if getErr != nil {
				return false
			}
			configuration := cluster.Spec.HostedCluster.Configuration
			if wantHTTP == "" && wantHTTPS == "" && wantNo == "" {
				return configuration == nil || configuration.Proxy == nil
			}
			if configuration == nil || configuration.Proxy == nil {
				return false
			}
			proxy := configuration.Proxy
			return proxy.HTTPProxy == wantHTTP && proxy.HTTPSProxy == wantHTTPS && proxy.NoProxy == wantNo
		}).WithContext(waitCtx).WithTimeout(5 * time.Minute).WithPolling(sharedClusterProxyPollInterval).Should(BeTrue())
	}
	proxyString := func(value *string) string {
		if value == nil {
			return ""
		}
		return *value
	}
	DeferCleanup(func(cleanupCtx SpecContext) {
		if !proxyAttempted {
			return
		}
		if clusterProxyChangeAttempted {
			By("Restoring the retained cluster's original proxy settings")
			if _, restoreErr := clusterSvc.Apply(&originalArgs); restoreErr != nil {
				Fail(fmt.Sprintf("restoring the cluster proxy settings failed; leaving proxy endpoints in place: %v", restoreErr))
				return
			}
			verifyClusterProxy(cleanupCtx, proxyString(originalArgs.HTTPProxy), proxyString(originalArgs.HTTPSProxy), proxyString(originalArgs.NoProxy))
		}
		By("Destroying the proxy endpoints created by this spec")
		_, destroyErr := proxySvc.Destroy()
		Expect(destroyErr).NotTo(HaveOccurred())
	}, GracePeriod(sharedClusterProxyTeardownGracePeriod))

	if runProxyValidationChecks {
		By("[72491] Verifying an over-limit proxy is rejected during cluster creation")
		invalidWorkspace := fmt.Sprintf("%s-invalid-proxy-%d", workspace, time.Now().UnixNano())
		invalidProxyClusterSvc, createErr := exec.NewSeaplaneClusterService(invalidWorkspace)
		Expect(createErr).NotTo(HaveOccurred())
		invalidClusterName := fmt.Sprintf("ip-%012d", time.Now().UnixNano()%1_000_000_000_000)
		invalidExpiration := time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339)
		invalidArgs := originalArgs
		invalidArgs.ClusterName = &invalidClusterName
		invalidArgs.ExpirationTimestamp = &invalidExpiration
		invalidArgs.HTTPProxy = &invalidHTTPProxy
		invalidArgs.HTTPSProxy = nil
		invalidArgs.NoProxy = nil
		// A configured OidcConfig can back only one cluster. Leave it unset here
		// so the API validates the intentionally over-limit proxy value instead
		// of rejecting the request because the retained cluster already uses it.
		invalidArgs.OIDCConfigID = nil
		_, invalidCreateErr := invalidProxyClusterSvc.Apply(&invalidArgs)
		if invalidCreateErr == nil {
			unexpectedCluster, outputErr := invalidProxyClusterSvc.Output()
			_, destroyErr := invalidProxyClusterSvc.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())
			if outputErr == nil && unexpectedCluster.ClusterID != "" {
				deleteErr := clusters.WaitUntil(ctx, invalidClusterName,
					func(cluster *v1alpha1.Cluster) bool { return cluster == nil }, sharedClusterProxyPollInterval, sharedClusterProxyClusterDeletedTimeout)
				Expect(deleteErr).NotTo(HaveOccurred())
			}
			Fail(fmt.Sprintf("Seaplane accepted an HTTP proxy longer than the 2048-character limit (output error: %v)", outputErr))
		}
		Expect(invalidCreateErr).To(HaveOccurred())
		invalidCreateMessage := strings.ToLower(invalidCreateErr.Error())
		Expect(invalidCreateMessage).To(Or(
			ContainSubstring("2048"),
			ContainSubstring("too long"),
			ContainSubstring("length"),
		), "cluster creation should fail because the HTTP proxy exceeds the API limit")
		_, destroyErr := invalidProxyClusterSvc.Destroy()
		Expect(destroyErr).NotTo(HaveOccurred())
	}

	By("Creating two proxy endpoints in the retained cluster VPC")
	proxyCount := 2
	proxyAttempted = true // Ensure partially created AWS resources are cleaned up on failure.
	region := awsRegion
	_, err = proxySvc.Apply(&exec.ProxyArgs{
		ProxyCount:          &proxyCount,
		Region:              &region,
		VPCID:               &vpcOutput.VPCID,
		PublicSubnetID:      &vpcOutput.PublicSubnetID,
		TrustBundleFilePath: &proxyTrustBundlePath,
	})
	Expect(err).NotTo(HaveOccurred())
	proxies, err := proxySvc.Output()
	Expect(err).NotTo(HaveOccurred())
	Expect(proxies.Proxies).To(HaveLen(2))
	initialProxy := proxies.Proxies[0]
	updatedProxy := proxies.Proxies[1]
	Expect(initialProxy.HttpProxy).NotTo(BeEmpty())
	Expect(initialProxy.HttpsProxy).NotTo(BeEmpty())
	Expect(updatedProxy.HttpProxy).NotTo(BeEmpty())
	Expect(updatedProxy.HttpsProxy).NotTo(BeEmpty())

	By("[72489] Applying and verifying the first proxy endpoint")
	updatedArgs := originalArgs
	updatedArgs.HTTPProxy = &initialProxy.HttpProxy
	updatedArgs.HTTPSProxy = &initialProxy.HttpsProxy
	updatedArgs.NoProxy = &initialProxy.NoProxy
	clusterProxyChangeAttempted = true
	_, err = clusterSvc.Apply(&updatedArgs)
	Expect(err).NotTo(HaveOccurred())
	verifyClusterProxy(ctx, initialProxy.HttpProxy, initialProxy.HttpsProxy, initialProxy.NoProxy)

	if runProxyValidationChecks {
		By("[72490] Rejecting an over-limit proxy update without changing stored settings")
		invalidEditArgs := updatedArgs
		invalidEditArgs.HTTPProxy = &invalidHTTPProxy
		_, err = clusterSvc.Apply(&invalidEditArgs)
		Expect(err).To(HaveOccurred())
		verifyClusterProxy(ctx, initialProxy.HttpProxy, initialProxy.HttpsProxy, initialProxy.NoProxy)
	}

	By("[72489] Updating proxy endpoints and no_proxy")
	updatedNoProxy := updatedProxy.NoProxy + ",example.com"
	updatedArgs.HTTPProxy = &updatedProxy.HttpProxy
	updatedArgs.HTTPSProxy = &updatedProxy.HttpsProxy
	updatedArgs.NoProxy = &updatedNoProxy
	_, err = clusterSvc.Apply(&updatedArgs)
	Expect(err).NotTo(HaveOccurred())
	verifyClusterProxy(ctx, updatedProxy.HttpProxy, updatedProxy.HttpsProxy, updatedNoProxy)

	By("Clearing proxy settings from the retained cluster")
	updatedArgs.HTTPProxy = nil
	updatedArgs.HTTPSProxy = nil
	updatedArgs.NoProxy = nil
	_, err = clusterSvc.Apply(&updatedArgs)
	Expect(err).NotTo(HaveOccurred())
	verifyClusterProxy(ctx, "", "", "")
}

func runClassicClusterProxyWorkflow(profile profilehandler.ProfileHandler) {
	clusterService, err := profile.Services().GetClusterService()
	Expect(err).NotTo(HaveOccurred())
	clusterArgs, err := clusterService.ReadTFVars()
	Expect(err).NotTo(HaveOccurred())
	originalClusterArgs := *clusterArgs

	proxyService, err := profile.Services().GetProxyService()
	Expect(err).NotTo(HaveOccurred())
	proxyArgs, err := proxyService.ReadTFVars()
	Expect(err).NotTo(HaveOccurred())
	proxyArgs.ProxyCount = new(2)
	var proxyAttempted bool
	var clusterProxyChangeAttempted bool
	DeferCleanup(func() {
		if !proxyAttempted {
			return
		}
		if clusterProxyChangeAttempted {
			By("Restore the original cluster settings")
			_, applyErr := clusterService.Apply(&originalClusterArgs)
			Expect(applyErr).NotTo(HaveOccurred())
		}
		By("Restore the original proxy endpoint")
		proxyArgs.ProxyCount = new(1)
		_, applyErr := proxyService.Apply(proxyArgs)
		Expect(applyErr).NotTo(HaveOccurred())
	})
	proxyAttempted = true
	_, err = proxyService.Apply(proxyArgs)
	Expect(err).NotTo(HaveOccurred())

	proxiesOutput, err := proxyService.Output()
	Expect(err).NotTo(HaveOccurred())
	Expect(proxiesOutput.Proxies).To(HaveLen(2))
	newProxyOutput := proxiesOutput.Proxies[1]

	By("Apply the new proxy endpoint")
	clusterArgs.Proxy = &exec.Proxy{
		AdditionalTrustBundle: new(newProxyOutput.AdditionalTrustBundle),
		HTTPSProxy:            new(newProxyOutput.HttpsProxy),
		HTTPProxy:             new(newProxyOutput.HttpProxy),
		NoProxy:               new(newProxyOutput.NoProxy),
	}
	clusterProxyChangeAttempted = true
	_, err = clusterService.Apply(clusterArgs)
	Expect(err).NotTo(HaveOccurred())

	By("Verify the new proxy settings")
	clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
	Expect(err).NotTo(HaveOccurred())
	clusterProxy := clusterResp.Body().Proxy()
	Expect(clusterProxy.HTTPProxy()).To(Equal(newProxyOutput.HttpProxy))
	Expect(clusterProxy.HTTPSProxy()).To(Equal(newProxyOutput.HttpsProxy))
	Expect(clusterProxy.NoProxy()).To(Equal(newProxyOutput.NoProxy))

	By("Remove the proxy settings")
	clusterArgs.Proxy = &exec.Proxy{
		AdditionalTrustBundle: helper.EmptyStringPointer,
		HTTPSProxy:            helper.EmptyStringPointer,
		HTTPProxy:             helper.EmptyStringPointer,
		NoProxy:               helper.EmptyStringPointer,
	}
	_, err = clusterService.Apply(clusterArgs)
	Expect(err).NotTo(HaveOccurred())

	By("Verify the proxy settings are removed")
	clusterResp, err = cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
	Expect(err).NotTo(HaveOccurred())
	clusterProxy = clusterResp.Body().Proxy()
	Expect(clusterProxy.HTTPProxy()).To(BeEmpty())
	Expect(clusterProxy.HTTPSProxy()).To(BeEmpty())
	Expect(clusterProxy.NoProxy()).To(BeEmpty())
}

func buildSharedProxySeaplaneClient(seaplaneURL, region string) *seaplane.Clientset {
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background())
	Expect(err).NotTo(HaveOccurred())
	identity, err := sts.NewFromConfig(awsConfig).GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	Expect(err).NotTo(HaveOccurred())
	client, err := seaplane.NewForConfig(&sprest.Config{
		Host:      seaplaneURL,
		Region:    region,
		AccountID: *identity.Account,
		CallerARN: *identity.Arn,
		AWSConfig: awsConfig,
	})
	Expect(err).NotTo(HaveOccurred())
	return client
}

func sharedClusterProxyRequiredEnv(key string) string {
	value := os.Getenv(key)
	Expect(value).NotTo(BeEmpty(), fmt.Sprintf("required env var %s is not set", key))
	return value
}
