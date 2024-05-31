package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("Edit cluster", ci.Day2, ci.NonClassicCluster, func() {

	var profile *ci.Profile
	var clusterService *exec.ClusterService
	var clusterArgs *exec.ClusterCreationArgs

	BeforeEach(func() {
		By("Load profile")
		profile = ci.LoadProfileYamlFileByENV()

		// Initialize the cluster service
		By("Create cluster service")
		var err error
		clusterService, err = exec.NewClusterService(profile.GetClusterManifestsDir())
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("validate", func() {
		It("required fields - [id:72452]", ci.Medium, ci.FeatureClusterDefault, func() {
			// By("Try to edit aws account with wrong value")
			// clusterArgs = &exec.ClusterCreationArgs{
			// 	AWSAccountID: helper.StringPointer("another_account"),
			// }
			// err := clusterService.Apply(clusterArgs, false, false)
			// Expect(err).To(HaveOccurred())
			// helper.ExpectTFErrorContains(err, "Attribute aws_account_id aws account ID must be only digits and exactly 12")

			// By("Try to edit aws account with wrong account")
			// clusterArgs = &exec.ClusterCreationArgs{

			// 	AWSAccountID: helper.StringPointer("000000000000"),
			// }
			// err = clusterService.Apply(clusterArgs, false, false)
			// Expect(err).To(HaveOccurred())
			// helper.ExpectTFErrorContains(err, "Attribute aws_account_id, cannot be changed from")

			// // To be activated once issue is solved
			// By("Try to edit billing account with wrong value")
			// clusterArgs = &exec.ClusterCreationArgs{
			// 	AWSBillingAccountID: helper.StringPointer("anything"),
			// }
			// err = clusterService.Apply(clusterArgs, false, false)
			// Expect(err).To(HaveOccurred())
			// helper.ExpectTFErrorContains(err, "Attribute aws_billing_account_id aws billing account ID must be only digits and exactly 12 in length")

			// By("Try to edit billing account with wrong account")
			// clusterArgs = &exec.ClusterCreationArgs{
			// 	AWSBillingAccountID: helper.StringPointer("000000000000"),
			// }
			// err = clusterService.Apply(clusterArgs, false, false)
			// Expect(err).To(HaveOccurred())
			// helper.ExpectTFErrorContains(err, "billing account 000000000000 not linked to organization")

			By("Try to edit cloud region")
			region := "us-east-1"
			if profile.Region == region {
				region = "us-west-2" // make sure we are not in the same region
			}
			clusterArgs = &exec.ClusterCreationArgs{
				AWSRegion: helper.StringPointer(region),
			}
			err := clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Invalid AZ")

			By("Try to edit cloud region and Availability zone(s)")
			region = "us-east-1"
			azs := []string{"us-east-1a"}
			if profile.Region == region {
				region = "us-west-2" // make sure we are not in the same region
				azs = []string{"us-west-2b"}
			}
			clusterArgs = &exec.ClusterCreationArgs{
				AWSRegion:            helper.StringPointer(region),
				AWSAvailabilityZones: helper.StringSlicePointer(azs),
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute cloud_region, cannot be changed from")

			// By("Try to edit name")
			// clusterArgs = &exec.ClusterCreationArgs{
			// 	ClusterName: helper.StringPointer("any_name"),
			// }
			// err = clusterService.Apply(clusterArgs, false, false)
			// Expect(err).To(HaveOccurred())
			// helper.ExpectTFErrorContains(err, "Attribute name, cannot be changed from")
		})

		It("compute fields - [id:72453]", ci.Medium, ci.FeatureClusterCompute, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			By("Try to edit compute machine type")
			machineType := "m5.2xlarge"
			if clusterResp.Body().Nodes().ComputeMachineType().ID() == machineType {
				machineType = "m5.xlarge"
			}
			clusterArgs = &exec.ClusterCreationArgs{
				ComputeMachineType: machineType,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute compute_machine_type, cannot be changed from")

			By("Try to edit replicas")
			clusterArgs = &exec.ClusterCreationArgs{
				Replicas: 5,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute replicas, cannot be changed from")

			By("Try to edit with replicas < 2")
			clusterArgs = &exec.ClusterCreationArgs{
				Replicas: 1,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute replicas, cannot be changed from")
		})

		It("properties - [id:72455]", ci.Medium, ci.FeatureClusterMisc, func() {
			By("Retrieve current properties")
			out, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			currentProperties := out.Properties

			By("Try to remove `rosa_creator_arn` property")
			props := helper.CopyStringMap(currentProperties)
			delete(props, "rosa_creator_arn")
			clusterArgs = &exec.ClusterCreationArgs{
				CustomProperties: props,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).ToNot(HaveOccurred())
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterResp.Body().Properties()["rosa_creator_arn"]).To(Equal(currentProperties["rosa_creator_arn"]))

			By("Try to edit `rosa_creator_arn` property")
			props = helper.CopyStringMap(currentProperties)
			props["rosa_creator_arn"] = "anything"
			clusterArgs = &exec.ClusterCreationArgs{
				CustomProperties: props,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Can't patch property 'rosa_creator_arn'")

			By("Try to edit `reserved` property")
			props = helper.CopyStringMap(currentProperties)
			props["rosa_tf_version"] = "any"
			clusterArgs = &exec.ClusterCreationArgs{
				CustomProperties: props,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Can not override reserved properties keys. rosa_tf_version is a reserved")
		})

		It("network fields - [id:72470]", ci.Medium, ci.FeatureClusterNetwork, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			originalAZs := clusterResp.Body().Nodes().AvailabilityZones()

			By("Try to edit availability zones")
			azs := []string{originalAZs[0]}
			clusterArgs = &exec.ClusterCreationArgs{
				AWSAvailabilityZones: &azs,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			// TODO report issue ? This should fail
			// Expect(err).To(HaveOccurred())
			// helper.ExpectTFErrorContains(err, "Attribute availability_zones, cannot be changed from")
			Expect(err).ToNot(HaveOccurred())
			clusterResp, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterResp.Body().Nodes().AvailabilityZones()).To(Equal(originalAZs))

			By("Try to edit subnet ids")
			azs = []string{"subnet-1", "subnet-2"}
			clusterArgs = &exec.ClusterCreationArgs{
				AWSSubnetIDs: &azs,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute aws_subnet_ids, cannot be changed from")

			By("Try to edit host prefix")
			clusterArgs = &exec.ClusterCreationArgs{
				HostPrefix: 25,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute host_prefix, cannot be changed from")

			By("Try to edit machine cidr")
			clusterArgs = &exec.ClusterCreationArgs{
				MachineCIDR: "10.0.0.0/17",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute machine_cidr, cannot be changed from")

			By("Try to edit service cidr")
			clusterArgs = &exec.ClusterCreationArgs{
				ServiceCIDR: "172.50.0.0/20",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute service_cidr, cannot be changed from")

			By("Try to edit pod cidr")
			clusterArgs = &exec.ClusterCreationArgs{
				PodCIDR: "10.128.0.0/16",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute pod_cidr, cannot be changed from")
		})

		It("version fields - [id:72478]", ci.Medium, ci.FeatureClusterNetwork, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			currentVersion := clusterResp.Body().Version().RawID()

			By("Get new channel group")
			otherChannelGroup := exec.NightlyChannel
			if profile.ChannelGroup == exec.NightlyChannel {
				otherChannelGroup = exec.CandidateChannel
			}

			By("Retrieve latest version")
			versions := cms.SortVersions(cms.HCPEnabledVersions(ci.RHCSConnection, otherChannelGroup))
			lastVersion := versions[len(versions)-1]

			if lastVersion.RawID != currentVersion {
				By("Try to edit version to one from another channel_group")
				clusterArgs = &exec.ClusterCreationArgs{
					OpenshiftVersion: lastVersion.RawID,
				}
				err = clusterService.Apply(clusterArgs, false, false)
				Expect(err).To(HaveOccurred())
				helper.ExpectTFErrorContains(err, fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, desired version (%s) is not in the list of available upgrades", clusterID, lastVersion.RawID))
			}

			By("Try to edit channel_group")
			otherChannelGroup = "stable"
			if profile.ChannelGroup == exec.StableChannel {
				otherChannelGroup = exec.CandidateChannel
			}
			clusterArgs = &exec.ClusterCreationArgs{
				ChannelGroup: otherChannelGroup,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute channel_group, cannot be changed from")

		})

		It("private fields - [id:72480]", ci.Medium, ci.FeatureClusterPrivate, func() {
			By("Try to edit private")
			clusterArgs = &exec.ClusterCreationArgs{
				Private: helper.BoolPointer(!profile.Private),
			}
			err := clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute private, cannot be changed from")
		})

		It("encryption fields - [id:72487]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit etcd_encryption")
			etcd := !profile.Etcd
			clusterArgs = &exec.ClusterCreationArgs{
				Etcd: &etcd,
			}
			err := clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute etcd_encryption, cannot be changed from")

			By("Try to edit etcd_kms_key_arn")
			clusterArgs = &exec.ClusterCreationArgs{
				EtcdKmsKeyARN: helper.StringPointer("anything"),
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute etcd_kms_key_arn, cannot be changed from")

			By("Try to edit kms_key_arn")
			clusterArgs = &exec.ClusterCreationArgs{
				KmsKeyARN: helper.StringPointer("anything"),
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute kms_key_arn, cannot be changed from")
		})

		It("sts fields - [id:72497]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit oidc config")
			clusterArgs = &exec.ClusterCreationArgs{
				OIDCConfigID: helper.StringPointer("2a4rv4o76gljek6c3po16abquaciv0a7"),
			}
			err := clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute sts.oidc_config_id, cannot be changed from")

			By("Try to edit installer role")
			clusterArgs = &exec.ClusterCreationArgs{
				StsInstallerRole: helper.StringPointer("anything"),
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute sts.role_arn, cannot be changed from")

			By("Try to edit support role")
			clusterArgs = &exec.ClusterCreationArgs{
				StsSupportRole: helper.StringPointer("anything"),
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute sts.support_role_arn, cannot be changed from")

			By("Try to edit worker role")
			clusterArgs = &exec.ClusterCreationArgs{
				StsWorkerRole: helper.StringPointer("anything"),
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute sts.instance_iam_roles.worker_role_arn, cannot be changed from")

			By("Try to edit operator role prefix")
			clusterArgs = &exec.ClusterCreationArgs{
				OperatorRolePrefix: helper.StringPointer("anything"),
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute sts.operator_role_prefix, cannot be changed from")
		})

		It("proxy fields - [id:72490]", ci.Medium, ci.FeatureClusterProxy, func() {
			clusterArgs = &exec.ClusterCreationArgs{
				Proxy: nil,
			}

			By("Edit proxy with wrong http_proxy")
			proxyArgs := exec.Proxy{
				HTTPProxy: helper.StringPointer("aaaaxxxx"),
			}
			clusterArgs.Proxy = &proxyArgs
			err := clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Invalid 'proxy.http_proxy' attribute 'aaaaxxxx'")

			By("Edit proxy with http_proxy starts with https")
			proxyArgs = exec.Proxy{
				HTTPProxy: helper.StringPointer("https://anything"),
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "'proxy.http_proxy' prefix is not 'http'")

			By("Edit proxy with invalid https_proxy")
			proxyArgs = exec.Proxy{
				HTTPSProxy: helper.StringPointer("aaavvv"),
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Invalid")
			helper.ExpectTFErrorContains(err, "'proxy.https_proxy' attribute 'aaavvv'")

			By("Edit proxy with invalid additional_trust_bundle set")
			proxyArgs = exec.Proxy{
				AdditionalTrustBundle: helper.StringPointer("wrong value"),
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Failed to parse")

			By("Edit proxy with http-proxy and http-proxy empty but no-proxy")
			proxyArgs = exec.Proxy{
				HTTPProxy:  helper.EmptyStringPointer,
				HTTPSProxy: helper.EmptyStringPointer,
				NoProxy:    helper.StringPointer("test.com"),
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Cannot set 'proxy.no_proxy' attribute 'test.com' while removing 'proxy.http_proxy'")

			By("Edit proxy with with invalid no_proxy")
			proxyArgs = exec.Proxy{
				NoProxy: helper.StringPointer("*"),
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "no-proxy value: '*' should match the regular expression")
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

			clusterArgs = &exec.ClusterCreationArgs{
				Tags: tags,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute tags, cannot be changed from")
		})
	})

})
