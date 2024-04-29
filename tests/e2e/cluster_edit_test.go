package e2e

import (
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
		clusterService, err = exec.NewClusterService(profile.GetClusterManifestsDir())
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("validate", func() {
		It("required fields - [id:72452]", ci.Medium, ci.FeatureClusterDefault, func() {
			By("Try to edit aws account with wrong value")
			clusterArgs = &exec.ClusterCreationArgs{

				AWSAccountID: "another_account",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute aws_account_id aws account ID must be only digits and exactly 12"))

			By("Try to edit aws account with wrong account")
			clusterArgs = &exec.ClusterCreationArgs{

				AWSAccountID: "000000000000",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute aws_account_id, cannot be changed from"))

			// To be activated once issue is solved
			// By("Try to edit billing account")
			// clusterArgs = &exec.ClusterCreationArgs{
			// 	AWSBillingAccountID: "anything",
			// }
			// err = clusterService.Apply(clusterArgs, false, true)
			// Expect(err).To(HaveOccurred())
			// Expect(err.Error()).To(ContainSubstring("Attribute aws_billing_account_id, cannot be changed from"))

			By("Try to edit cloud region")
			region := "us-east-1"
			if profile.Region == region {
				region = "us-west-2" // make sure we are not in the same region
			}
			clusterArgs = &exec.ClusterCreationArgs{
				AWSRegion: region,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid AZ"))

			By("Try to edit name")
			clusterArgs = &exec.ClusterCreationArgs{
				ClusterName: "any_name",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute name, cannot be changed from"))
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
			Expect(err.Error()).To(ContainSubstring("Attribute compute_machine_type, cannot be changed from"))

			By("Try to edit replicas")
			clusterArgs = &exec.ClusterCreationArgs{
				Replicas: 5,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute replicas, cannot be changed from"))

			By("Try to edit with replicas < 2")
			clusterArgs = &exec.ClusterCreationArgs{
				Replicas: 1,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute replicas, cannot be changed from"))
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
			Expect(err.Error()).To(ContainSubstring("Can't patch property 'rosa_creator_arn'"))

			By("Try to edit `reserved` property")
			props = helper.CopyStringMap(currentProperties)
			props["rosa_tf_version"] = "any"
			clusterArgs = &exec.ClusterCreationArgs{
				CustomProperties: props,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Can not override reserved properties keys. rosa_tf_version is a reserved"))
		})

		It("network fields - [id:72470]", ci.Medium, ci.FeatureClusterNetwork, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			originalAZs := clusterResp.Body().Nodes().AvailabilityZones()

			By("Try to edit availability zones")
			clusterArgs = &exec.ClusterCreationArgs{
				AWSAvailabilityZones: []string{originalAZs[0]},
			}
			err = clusterService.Apply(clusterArgs, false, false)
			// TODO report issue ? This should fail
			// Expect(err).To(HaveOccurred())
			// Expect(err.Error()).To(ContainSubstring("Attribute availability_zones, cannot be changed from"))
			Expect(err).ToNot(HaveOccurred())
			clusterResp, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterResp.Body().Nodes().AvailabilityZones()).To(Equal(originalAZs))

			By("Try to edit subnet ids")
			clusterArgs = &exec.ClusterCreationArgs{
				AWSSubnetIDs: []string{"subnet-1", "subnet-2"},
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute aws_subnet_ids, cannot be changed from"))

			By("Try to edit host prefix")
			clusterArgs = &exec.ClusterCreationArgs{
				HostPrefix: 25,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute host_prefix, cannot be changed from"))

			By("Try to edit machine cidr")
			clusterArgs = &exec.ClusterCreationArgs{
				MachineCIDR: "10.0.0.0/17",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute machine_cidr, cannot be changed from"))

			By("Try to edit service cidr")
			clusterArgs = &exec.ClusterCreationArgs{
				ServiceCIDR: "172.50.0.0/20",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute service_cidr, cannot be changed from"))

			By("Try to edit pod cidr")
			clusterArgs = &exec.ClusterCreationArgs{
				PodCIDR: "10.128.0.0/16",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute pod_cidr, cannot be changed from"))
		})

		It("version fields - [id:72478]", ci.Medium, ci.FeatureClusterNetwork, func() {
			By("Retrieve cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			currentVersion := clusterResp.Body().Version().RawID()

			By("Retrieve latest version")
			versions := cms.SortVersions(cms.HCPEnabledVersions(ci.RHCSConnection, profile.ChannelGroup))
			lastVersion := versions[len(versions)-1]

			if lastVersion.RawID != currentVersion {
				By("Try to edit version to one from another channel_group")
				clusterArgs = &exec.ClusterCreationArgs{
					OpenshiftVersion: lastVersion.RawID,
				}
				err = clusterService.Apply(clusterArgs, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't upgrade cluster version with identifier"))
				Expect(err.Error()).To(ContainSubstring("is not in"))
				Expect(err.Error()).To(ContainSubstring("the list of available upgrades"))
			}

			By("Try to edit channel_group")
			otherChannelGroup := "stable"
			if profile.ChannelGroup == exec.StableChannel {
				otherChannelGroup = exec.CandidateChannel
			}
			clusterArgs = &exec.ClusterCreationArgs{
				ChannelGroup: otherChannelGroup,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute channel_group, cannot be changed from"))

		})

		It("private fields - [id:72480]", ci.Medium, ci.FeatureClusterPrivate, func() {
			By("Try to edit private")
			clusterArgs = &exec.ClusterCreationArgs{
				Private: !profile.Private,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute private, cannot be changed from"))
		})

		It("encryption fields - [id:72487]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit etcd_encryption")
			etcd := !profile.Etcd
			clusterArgs = &exec.ClusterCreationArgs{
				Etcd: &etcd,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute etcd_encryption, cannot be changed from"))

			By("Try to edit kms_key_arn")
			clusterArgs = &exec.ClusterCreationArgs{
				KmsKeyARN: "anything",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute kms_key_arn, cannot be changed from"))
		})

		It("sts fields - [id:72497]", ci.Medium, ci.FeatureClusterEncryption, func() {
			By("Try to edit oidc config")
			clusterArgs = &exec.ClusterCreationArgs{
				OIDCConfigID: "2a4rv4o76gljek6c3po16abquaciv0a7",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute sts.oidc_config_id, cannot be changed from"))

			By("Try to edit installer role")
			clusterArgs = &exec.ClusterCreationArgs{
				StsInstallerRole: "anything",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute sts.role_arn, cannot be changed from"))

			By("Try to edit support role")
			clusterArgs = &exec.ClusterCreationArgs{
				StsSupportRole: "anything",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute sts.support_role_arn, cannot be changed from"))

			By("Try to edit worker role")
			clusterArgs = &exec.ClusterCreationArgs{
				StsWorkerRole: "anything",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute sts.instance_iam_roles.worker_role_arn, cannot be changed from"))

			By("Try to edit operator role prefix")
			clusterArgs = &exec.ClusterCreationArgs{
				OperatorRolePrefix: "anything",
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute sts.operator_role_prefix, cannot be changed from"))
		})

		It("proxy fields - [id:72490]", ci.Medium, ci.FeatureClusterProxy, func() {
			clusterArgs = &exec.ClusterCreationArgs{
				Proxy: nil,
			}

			By("Edit proxy with wrong http_proxy")
			value := "aaaaxxxx"
			proxyArgs := exec.Proxy{
				HTTPProxy: &value,
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid"))
			Expect(err.Error()).To(ContainSubstring("'proxy.http_proxy' attribute 'aaaaxxxx'"))

			By("Edit proxy with http_proxy starts with https")
			value = "https://anything"
			proxyArgs = exec.Proxy{
				HTTPProxy: &value,
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("'proxy.http_proxy' prefix is not 'http'"))

			By("Edit proxy with invalid https_proxy")
			value = "aaavvv"
			proxyArgs = exec.Proxy{
				HTTPSProxy: &value,
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid"))
			Expect(err.Error()).To(ContainSubstring("'proxy.https_proxy' attribute 'aaavvv'"))

			By("Edit proxy with invalid additional_trust_bundle set")
			value = "wrong value"
			proxyArgs = exec.Proxy{
				AdditionalTrustBundle: &value,
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to parse"))

			By("Edit proxy with http-proxy and http-proxy empty but no-proxy")
			value = "test.com"
			emptyValue := ""
			proxyArgs = exec.Proxy{
				HTTPProxy:  &emptyValue,
				HTTPSProxy: &emptyValue,
				NoProxy:    &value,
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			if profile.Proxy {
				Expect(err.Error()).To(ContainSubstring("'proxy.no_proxy' attribute 'test.com' while removing 'proxy.http_proxy'"))
			} else {
				Expect(err.Error()).To(ContainSubstring("'proxy.http_proxy' or 'proxy.https_proxy' attributes is needed to set 'proxy.no_proxy' 'test.com'"))
			}

			By("Edit proxy with with invalid no_proxy")
			value = "*"
			proxyArgs = exec.Proxy{
				NoProxy: &value,
			}
			clusterArgs.Proxy = &proxyArgs
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no-proxy value: '*' should match the regular expression"))
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
			Expect(err.Error()).To(ContainSubstring("Attribute tags, cannot be changed from"))
		})
	})

})
