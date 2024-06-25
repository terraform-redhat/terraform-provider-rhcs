package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
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

		// Initialize the cluster service
		By("Create cluster service")
		clusterService, err = profileHandler.Services().GetClusterService()
		Expect(err).ShouldNot(HaveOccurred())

		clusterArgs = retrieveClusterArgs()
		originalClusterArgs = retrieveClusterArgs()
	})

	AfterEach(func() {
		clusterService.Apply(originalClusterArgs)
	})

	Context("can edit/delete", func() {
		It("proxy - [id:72489]", ci.High, ci.FeatureClusterProxy, func() {
			if !profileHandler.Profile().IsProxy() {
				Skip("No proxy configured")
			}

			By("Retrieve Proxy service")
			proxyService, err := profileHandler.Services().GetProxyService()
			Expect(err).ShouldNot(HaveOccurred())

			readProxyArgs := func() (*exec.ProxyArgs, error) {
				return proxyService.ReadTFVars()
			}

			By("Create new proxy")
			proxyArgs, err := readProxyArgs()
			Expect(err).ShouldNot(HaveOccurred())
			proxyArgs.ProxyCount = helper.IntPointer(2)
			_, err = proxyService.Apply(proxyArgs)
			Expect(err).ShouldNot(HaveOccurred())
			defer func() {
				By("Delete created proxy")
				proxyArgs.ProxyCount = helper.IntPointer(1)
				_, err = proxyService.Apply(proxyArgs)
				Expect(err).ShouldNot(HaveOccurred())
			}()
			proxiesOutput, err := proxyService.Output()
			Expect(err).ShouldNot(HaveOccurred())
			newProxyOutput := proxiesOutput.Proxies[1]

			By("Edit cluster proxy with new proxy information")
			clusterArgs.Proxy = &exec.Proxy{
				AdditionalTrustBundle: helper.StringPointer(newProxyOutput.AdditionalTrustBundle),
				HTTPSProxy:            helper.StringPointer(newProxyOutput.HttpsProxy),
				HTTPProxy:             helper.StringPointer(newProxyOutput.HttpProxy),
				NoProxy:               helper.StringPointer(newProxyOutput.NoProxy),
			}
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ShouldNot(HaveOccurred())

			By("Verify new proxy information are set")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			clusterProxy := clusterResp.Body().Proxy()
			Expect(clusterProxy.HTTPProxy()).To(Equal(newProxyOutput.HttpProxy))
			Expect(clusterProxy.HTTPSProxy()).To(Equal(newProxyOutput.HttpsProxy))
			Expect(clusterProxy.NoProxy()).To(Equal(newProxyOutput.NoProxy))

			// Remove proxy completely
			By("Remove proxy from cluster")
			clusterArgs.Proxy = &exec.Proxy{
				AdditionalTrustBundle: helper.EmptyStringPointer,
				HTTPSProxy:            helper.EmptyStringPointer,
				HTTPProxy:             helper.EmptyStringPointer,
				NoProxy:               helper.EmptyStringPointer,
			}
			clusterService.Apply(clusterArgs)
			Expect(err).ShouldNot(HaveOccurred())

			By("Check proxy is removed")
			clusterResp, err = cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			clusterProxy = clusterResp.Body().Proxy()
			Expect(clusterProxy.HTTPProxy()).To(BeEmpty())
			Expect(clusterProxy.HTTPSProxy()).To(BeEmpty())
			Expect(clusterProxy.NoProxy()).To(BeEmpty())
		})
	})

	Context("validate", func() {
		BeforeEach(func() {
			if !profileHandler.Profile().IsHCP() {
				Skip("Test can run only on Hosted cluster")
			}
		})

		validateClusterArg := func(updateFields func(args *exec.ClusterArgs), validateErrFunc func(err error)) {
			updateFields(clusterArgs)
			_, err := clusterService.Apply(clusterArgs)
			validateErrFunc(err)
			clusterArgs = retrieveClusterArgs()
		}
		validateClusterArgAgainstErrorSubstrings := func(updateFields func(args *exec.ClusterArgs), errSubStrings ...string) {
			validateClusterArg(updateFields, func(err error) {
				Expect(err).To(HaveOccurred())
				for _, errStr := range errSubStrings {
					helper.ExpectTFErrorContains(err, errStr)
				}
			})
		}

		It("required fields - [id:72452]", ci.Medium, ci.FeatureClusterDefault, func() {
			By("Try to edit aws account with wrong value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAccountID = helper.StringPointer("another_account")
			}, "Attribute aws_account_id aws account ID must be only digits and exactly 12")

			By("Try to edit aws account with wrong account")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAccountID = helper.StringPointer("000000000000")
			}, "Attribute aws_account_id, cannot be changed from")

			By("Try to edit billing account with wrong value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSBillingAccountID = helper.StringPointer("anything")
			}, "Attribute aws_billing_account_id aws billing account ID must be only digits and exactly 12 in length")

			By("Try to edit billing account with wrong account")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSBillingAccountID = helper.StringPointer("000000000000")
			}, "billing account 000000000000 not linked to organization")

			By("Try to edit cloud region")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				region := "us-east-1"
				if profileHandler.Profile().GetRegion() == region {
					region = "us-west-2" // make sure we are not in the same region
				}
				args.AWSRegion = helper.StringPointer(region)
			}, "Invalid AZ")

			By("Try to edit cloud region and Availability zone(s)")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				region := "us-east-1"
				azs := []string{"us-east-1a"}
				if profileHandler.Profile().GetRegion() == region {
					region = "us-west-2" // make sure we are not in the same region
					azs = []string{"us-west-2b"}
				}
				args.AWSRegion = helper.StringPointer(region)
				args.AWSAvailabilityZones = helper.StringSlicePointer(azs)
			}, "Attribute cloud_region, cannot be changed from")

			By("Try to edit name")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ClusterName = helper.StringPointer("any_name")
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
				args.ComputeMachineType = helper.StringPointer(machineType)
			}, "Attribute compute_machine_type, cannot be changed from")

			By("Try to edit replicas")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Replicas = helper.IntPointer(5)
			}, "Attribute replicas, cannot be changed from")

			By("Try to edit with replicas < 2")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Replicas = helper.IntPointer(1)
			}, "Attribute replicas, cannot be changed from")
		})

		It("properties - [id:72455]", ci.Medium, ci.FeatureClusterMisc, func() {
			By("Retrieve current properties")
			out, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			currentProperties := out.Properties

			By("Try to remove `rosa_creator_arn` property")
			props := helper.CopyStringMap(currentProperties)
			delete(props, "rosa_creator_arn")
			validateClusterArg(func(args *exec.ClusterArgs) {
				args.CustomProperties = helper.StringMapPointer(props)
			}, func(err error) {
				Expect(err).ToNot(HaveOccurred())
				clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterResp.Body().Properties()["rosa_creator_arn"]).To(Equal(currentProperties["rosa_creator_arn"]))
			})

			By("Try to edit `rosa_creator_arn` property")
			props = helper.CopyStringMap(currentProperties)
			props["rosa_creator_arn"] = "anything"
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.CustomProperties = helper.StringMapPointer(props)
			}, "Can't patch property 'rosa_creator_arn'")

			By("Try to edit `reserved` property")
			props = helper.CopyStringMap(currentProperties)
			props["rosa_tf_version"] = "any"
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.CustomProperties = helper.StringMapPointer(props)
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
				args.AWSAvailabilityZones = helper.StringSlicePointer(azs)
			}, func(err error) {
				Expect(err).ToNot(HaveOccurred())
				clusterResp, err = cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterResp.Body().Nodes().AvailabilityZones()).To(Equal(originalAZs))
			})

			By("Try to edit subnet ids")
			azs = []string{"subnet-1", "subnet-2"}
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAvailabilityZones = helper.StringSlicePointer(azs)
			}, "Invalid AZ")

			By("Try to edit host prefix")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.HostPrefix = helper.IntPointer(25)
			}, "Attribute host_prefix, cannot be changed from")

			By("Try to edit machine cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.MachineCIDR = helper.StringPointer("10.0.0.0/17")
			}, "Attribute machine_cidr, cannot be changed from")

			By("Try to edit service cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ServiceCIDR = helper.StringPointer("172.50.0.0/20")
			}, "Attribute service_cidr, cannot be changed from")

			By("Try to edit pod cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.PodCIDR = helper.StringPointer("10.128.0.0/16")
			}, "Attribute pod_cidr, cannot be changed from")
		})

		It("version fields - [id:72478]", ci.Medium, ci.FeatureClusterNetwork, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			currentVersion := clusterResp.Body().Version().RawID()

			By("Get new channel group")
			otherChannelGroup := constants.VersionNightlyChannel
			if profileHandler.Profile().GetChannelGroup() == constants.VersionNightlyChannel {
				otherChannelGroup = constants.VersionCandidateChannel
			}

			By("Retrieve latest version")
			versions := cms.SortVersions(cms.HCPEnabledVersions(cms.RHCSConnection, otherChannelGroup))
			lastVersion := versions[len(versions)-1]

			if lastVersion.RawID != currentVersion {
				By("Try to edit version to one from another channel_group")
				errString := fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, desired version (%s) is not in the list of available upgrades", clusterID, lastVersion.RawID)
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.OpenshiftVersion = helper.StringPointer(lastVersion.RawID)
				}, errString)
			}

			By("Try to edit channel_group")
			otherChannelGroup = constants.VersionStableChannel
			if profileHandler.Profile().GetChannelGroup() == constants.VersionStableChannel {
				otherChannelGroup = constants.VersionCandidateChannel
			}
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ChannelGroup = helper.StringPointer(otherChannelGroup)
			}, "Attribute channel_group, cannot be changed from")
		})

		It("private fields - [id:72480]", ci.Medium, ci.FeatureClusterPrivate, func() {
			By("Try to edit private")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Private = helper.BoolPointer(!profileHandler.Profile().IsPrivate())
			}, "Attribute private, cannot be changed from")
		})

		It("imdsv2 fields - [id:75414]", ci.Medium, ci.FeatureClusterIMDSv2, func() {
			By("Try to edit ec2_metadata_http_tokens value")
			otherHttpToken := profileHandler.Profile().GetImdsv2()
			if otherHttpToken == constants.RequiredEc2MetadataHttpTokens {
				otherHttpToken = constants.OptionalEc2MetadataHttpTokens
			} else {
				otherHttpToken = constants.RequiredEc2MetadataHttpTokens
			}
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Ec2MetadataHttpTokens = helper.StringPointer(otherHttpToken)
			})
		})

		It("encryption fields - [id:72487]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit etcd_encryption")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Etcd = helper.BoolPointer(!profileHandler.Profile().IsEtcd())
			}, "Attribute etcd_encryption, cannot be changed from")

			By("Try to edit etcd_kms_key_arn")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.EtcdKmsKeyARN = helper.StringPointer("anything")
			}, "Attribute etcd_kms_key_arn, cannot be changed from")

			By("Try to edit kms_key_arn")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.KmsKeyARN = helper.StringPointer("anything")
			}, "Attribute kms_key_arn, cannot be changed from")
		})

		It("sts fields - [id:72497]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit oidc config")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OIDCConfigID = helper.StringPointer("2a4rv4o76gljek6c3po16abquaciv0a7")
			}, "Attribute sts.oidc_config_id, cannot be changed from")

			By("Try to edit installer role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsInstallerRole = helper.StringPointer("anything")
			}, "Attribute sts.role_arn, cannot be changed from")

			By("Try to edit support role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsSupportRole = helper.StringPointer("anything")
			}, "Attribute sts.support_role_arn, cannot be changed from")

			By("Try to edit worker role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsWorkerRole = helper.StringPointer("anything")
			}, "Attribute sts.instance_iam_roles.worker_role_arn, cannot be changed from")

			By("Try to edit operator role prefix")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OperatorRolePrefix = helper.StringPointer("anything")
			}, "Attribute sts.operator_role_prefix, cannot be changed from")
		})

		It("proxy fields - [id:72490]", ci.Medium, ci.FeatureClusterProxy, func() {
			By("Edit proxy with wrong http_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy: helper.StringPointer("aaaaxxxx"),
				}
			}, "Invalid 'proxy.http_proxy' attribute 'aaaaxxxx'")

			By("Edit proxy with http_proxy starts with https")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy: helper.StringPointer("https://anything"),
				}
			}, "'proxy.http_proxy' prefix is not 'http'")

			By("Edit proxy with invalid https_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPSProxy: helper.StringPointer("aaavvv"),
				}
			},
				"Invalid",
				"'proxy.https_proxy' attribute 'aaavvv'",
			)

			By("Edit proxy with invalid additional_trust_bundle set")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					AdditionalTrustBundle: helper.StringPointer("wrong value"),
				}
			}, "Failed to parse")

			By("Edit proxy with http-proxy and http-proxy empty but no-proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy:  helper.EmptyStringPointer,
					HTTPSProxy: helper.EmptyStringPointer,
					NoProxy:    helper.StringPointer("test.com"),
				}
			}, "Cannot set 'proxy.no_proxy' attribute 'test.com' while removing 'proxy.http_proxy'")

			By("Edit proxy with with invalid no_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					NoProxy: helper.StringPointer("*"),
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
				args.Tags = helper.StringMapPointer(tags)
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
						AdditionalComputeSecurityGroups:      helper.StringSlicePointer(output.AdditionalComputeSecurityGroups[0:1]),
						AdditionalInfraSecurityGroups:        helper.StringSlicePointer(output.AdditionalInfraSecurityGroups),
						AdditionalControlPlaneSecurityGroups: helper.StringSlicePointer(output.AdditionalControlPlaneSecurityGroups),
						AWSRegion:                            helper.StringPointer(profileHandler.Profile().GetRegion()),
					},
					"aws_additional_infra_security_group_ids": {
						AdditionalInfraSecurityGroups:        helper.StringSlicePointer(output.AdditionalInfraSecurityGroups[0:1]),
						AdditionalComputeSecurityGroups:      helper.StringSlicePointer(output.AdditionalComputeSecurityGroups),
						AdditionalControlPlaneSecurityGroups: helper.StringSlicePointer(output.AdditionalControlPlaneSecurityGroups),
						AWSRegion:                            helper.StringPointer(profileHandler.Profile().GetRegion()),
					},
					"aws_additional_control_plane_security_group_ids": {
						AdditionalControlPlaneSecurityGroups: helper.StringSlicePointer(output.AdditionalControlPlaneSecurityGroups[0:1]),
						AdditionalComputeSecurityGroups:      helper.StringSlicePointer(output.AdditionalComputeSecurityGroups),
						AdditionalInfraSecurityGroups:        helper.StringSlicePointer(output.AdditionalInfraSecurityGroups),
						AWSRegion:                            helper.StringPointer(profileHandler.Profile().GetRegion()),
					},
				}
				for keyword, updatingArgs := range args {
					_, err := clusterService.Apply(updatingArgs)
					Expect(err).To(HaveOccurred(), keyword)
					Expect(err.Error()).Should(ContainSubstring(`Attribute value cannot be changed`))
				}
			})
	})

	Context("work for", func() {
		It("autoscaling change - [id:63147]",
			ci.Medium,
			func() {
				if profileHandler.Profile().IsHCP() {
					Skip("This case only works for classic now")
				}

				By("Update the cluster to autoscaling")
				clusterArgs.Autoscaling = &exec.Autoscaling{
					AutoscalingEnabled: helper.BoolPointer(true),
					MinReplicas:        helper.IntPointer(3),
					MaxReplicas:        helper.IntPointer(6),
				}

				if profileHandler.Profile().IsAutoscale() {
					clusterArgs.Autoscaling = &exec.Autoscaling{
						AutoscalingEnabled: helper.BoolPointer(false),
					}
					clusterArgs.Replicas = helper.IntPointer(3)
				}

				_, err := clusterService.Apply(clusterArgs)
				Expect(err).To(HaveOccurred())
				helper.ExpectTFErrorContains(err, "Attribute max_replicas, cannot be changed from")
			})
	})

})
