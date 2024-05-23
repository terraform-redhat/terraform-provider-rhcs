package e2e

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

var originalCreationArgs *exec.ClusterCreationArgs
var creationArgs *exec.ClusterCreationArgs
var clusterService *exec.ClusterService
var err error

var profile *ci.Profile

var _ = Describe("Negative Tests", Ordered, func() {
	defer GinkgoRecover()

	BeforeAll(func() {
		profile = ci.LoadProfileYamlFileByENV()

		originalCreationArgs, err = ci.GenerateClusterCreationArgsByProfile(token, profile)
		if err != nil {
			defer ci.DestroyRHCSClusterByProfile(token, profile)
		}
		Expect(err).ToNot(HaveOccurred())

		clusterService, err = exec.NewClusterService(profile.GetClusterManifestsDir())
		if err != nil {
			defer ci.DestroyRHCSClusterByProfile(token, profile)
		}
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		err := ci.DestroyRHCSClusterByProfile(token, profile)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("cluster admin", ci.NonHCPCluster, ci.Day1Negative, func() {
		BeforeEach(OncePerOrdered, func() {
			if !profile.AdminEnabled {
				Skip("The tests configured for cluster admin only")
			}

			// Restore cluster args
			creationArgs = originalCreationArgs
		})

		It("validate user name policy - [id:65961]", ci.Medium,
			func() {
				By("Edit cluster admin user name to not valid")
				creationArgs.AdminCredentials["username"] = "one:two"
				err = clusterService.Apply(creationArgs, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.username username may not contain the characters:\n'/:%'"))

				By("Edit cluster admin user name to empty")
				creationArgs.AdminCredentials["username"] = ""
				err = clusterService.Apply(creationArgs, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.username username may not be empty/blank string"))
			})

		It("validate password policy - [id:65963]", ci.Medium, func() {
			By("Edit cluster admin password  to the short one")
			creationArgs.AdminCredentials["password"] = helper.GenerateRandomStringWithSymbols(13)
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password string length must be at least 14"))
			By("Edit cluster admin password to empty")
			creationArgs.AdminCredentials["password"] = ""
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password should use ASCII-standard"))

			By("Edit cluster admin password that lacks a capital letter")
			creationArgs.AdminCredentials["password"] = strings.ToLower(helper.GenerateRandomStringWithSymbols(14))
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password must contain uppercase\ncharacters"))

			By("Edit cluster admin password that lacks symbol but has digits")
			creationArgs.AdminCredentials["password"] = "QwertyPasswordNoDigitsSymbols"
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password must contain numbers or\nsymbols"))

			By("Edit cluster admin password that includes Non English chars")
			creationArgs.AdminCredentials["password"] = "Qwert12345345@ש"
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password should use ASCII-standard\ncharacters only"))
		})
	})

	Describe("Create HCP cluster", ci.NonClassicCluster, ci.Day1Negative, func() {
		BeforeEach(OncePerOrdered, func() {
			creationArgs = originalCreationArgs
		})

		It("validate required fields - [id:72445]", ci.High, func() {
			By("Create cluster with wrong billing account")
			oldBillingValue := creationArgs.AWSBillingAccountID
			restoreValues := func() { creationArgs.AWSBillingAccountID = oldBillingValue }
			newValue := "012345678912"
			creationArgs.AWSBillingAccountID = helper.StringPointer(newValue)
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("billing account %s not linked to organization", newValue))
			restoreValues()

			By("Create cluster with wrong cloud region")
			regionOldValue := creationArgs.AWSRegion
			restoreValues = func() { creationArgs.AWSRegion = regionOldValue }
			newValue = "us-anything-2"
			creationArgs.AWSRegion = helper.StringPointer(newValue)
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("Invalid AWS Region: %s", newValue))
			restoreValues()

			By("Create cluster with cluster name > 54 characters")
			clusterNameOldValue := creationArgs.ClusterName
			restoreValues = func() { creationArgs.ClusterName = clusterNameOldValue }
			creationArgs.ClusterName = helper.StringPointer(helper.GenerateRandomName("cluster-72445", 50))
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute name string length must be at most 54")
			restoreValues()

			By("Create cluster with empty aws account id")
			AwsAccountIDOldValue := creationArgs.AWSAccountID
			restoreValues = func() { creationArgs.AWSAccountID = AwsAccountIDOldValue }
			creationArgs.AWSAccountID = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute aws_account_id aws account ID must be only digits and exactly 12 in length")
			restoreValues()

			By("Create cluster with empty aws billing account id")
			oldBillingValue = creationArgs.AWSBillingAccountID
			restoreValues = func() { creationArgs.AWSBillingAccountID = oldBillingValue }
			creationArgs.AWSBillingAccountID = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute aws_billing_account_id aws billing account ID must be only digits and exactly 12 in length")
			restoreValues()

			By("Create cluster with empty name")
			clusterNameOldValue = creationArgs.ClusterName
			restoreValues = func() { creationArgs.ClusterName = clusterNameOldValue }
			creationArgs.ClusterName = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "'name' must consist of lower case characters or '-', start with an alphabetic character, and end with an alphanumeric character")
			restoreValues()

			By("Create cluster with empty cloud region")
			regionOldValue = creationArgs.AWSRegion
			restoreValues = func() { creationArgs.AWSRegion = regionOldValue }
			creationArgs.AWSRegion = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute 'region.id' is mandatory")
			restoreValues()

			By("Create cluster with empty sts")
			oldOperatorPrefix := creationArgs.OperatorRolePrefix
			oldOIDCConfigID := creationArgs.OIDCConfigID
			oldInstallerRole := creationArgs.StsInstallerRole
			oldSupportRole := creationArgs.StsSupportRole
			oldWorkerRole := creationArgs.StsWorkerRole
			restoreValues = func() {
				creationArgs.OperatorRolePrefix = oldOperatorPrefix
				creationArgs.OIDCConfigID = oldOIDCConfigID
				creationArgs.StsInstallerRole = oldInstallerRole
				creationArgs.StsSupportRole = oldSupportRole
				creationArgs.StsWorkerRole = oldWorkerRole
			}
			creationArgs.OperatorRolePrefix = helper.EmptyStringPointer
			creationArgs.OIDCConfigID = helper.EmptyStringPointer
			creationArgs.StsInstallerRole = helper.EmptyStringPointer
			creationArgs.StsSupportRole = helper.EmptyStringPointer
			creationArgs.StsWorkerRole = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "The 'aws.sts.role_arn' parameter is mandatory")
			restoreValues()

			By("Create cluster with empty subnet Ids")
			oldSubnetIDs := creationArgs.AWSSubnetIDs
			restoreValues = func() { creationArgs.AWSSubnetIDs = oldSubnetIDs }
			creationArgs.AWSSubnetIDs = helper.EmptyStringSlicePointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Hosted Control Plane requires at least one subnet")
			restoreValues()

			By("Create cluster with subnet azs")
			oldAZs := creationArgs.AWSAvailabilityZones
			restoreValues = func() { creationArgs.AWSAvailabilityZones = oldAZs }
			creationArgs.AWSAvailabilityZones = helper.EmptyStringSlicePointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Availability zones must be provided for the subnet IDs provided")
			restoreValues()

			By("Create cluster without rosa_creator_arn property")
			oldIncludeCreatorProp := creationArgs.IncludeCreatorProperty
			restoreValues = func() {
				creationArgs.IncludeCreatorProperty = oldIncludeCreatorProp
			}
			includeProp := false
			creationArgs.IncludeCreatorProperty = &includeProp
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Expected property 'rosa_creator_arn'")
			restoreValues()

			By("Create cluster with wrong rosa_creator_arn property")
			oldProperties := creationArgs.CustomProperties
			oldIncludeCreatorProp = creationArgs.IncludeCreatorProperty
			restoreValues = func() {
				creationArgs.CustomProperties = oldProperties
				creationArgs.IncludeCreatorProperty = oldIncludeCreatorProp
			}
			newProperties := helper.CopyStringMap(oldProperties)
			newProperties["rosa_creator_arn"] = "wrong_value"
			includeProp = false
			creationArgs.CustomProperties = newProperties
			creationArgs.IncludeCreatorProperty = &includeProp
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Property 'rosa_creator_arn' does not have a valid user arn")
			restoreValues()

		})

		It("validate compute fields - [id:72449]", ci.Medium, func() {
			By("Create cluster with wrong compute machine type")
			oldMachineType := creationArgs.ComputeMachineType
			restoreValues := func() {
				creationArgs.ComputeMachineType = oldMachineType
			}
			creationArgs.ComputeMachineType = "anything"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Machine type 'anything' is not supported")
			restoreValues()

			By("Create cluster with replicas < 1")
			oldReplicas := creationArgs.Replicas
			restoreValues = func() {
				creationArgs.Replicas = oldReplicas
			}
			creationArgs.Replicas = 1
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "A hosted cluster requires at least 2 replicas")
			restoreValues()
		})

		It("validate version fields - [id:72477]", ci.Medium, func() {
			By("Create cluster with wrong version")
			oldVersion := creationArgs.OpenshiftVersion
			restoreValues := func() {
				creationArgs.OpenshiftVersion = oldVersion
			}
			creationArgs.OpenshiftVersion = "4.14.2-rc"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "version 4.14.2-rc is not in the list of supported versions")
			restoreValues()

			By("Create cluster with wrong channel group")
			oldChannel := creationArgs.ChannelGroup
			restoreValues = func() {
				creationArgs.ChannelGroup = oldChannel
			}
			creationArgs.ChannelGroup = "anything"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Could not find versions")
			restoreValues()

			By("Create cluster with version from another channel group")
			oldVersion = creationArgs.OpenshiftVersion
			oldChannel = creationArgs.ChannelGroup
			restoreValues = func() {
				creationArgs.OpenshiftVersion = oldVersion
				creationArgs.ChannelGroup = oldChannel
			}
			versions := cms.HCPEnabledVersions(ci.RHCSConnection, exec.CandidateChannel)
			versions = cms.SortVersions(versions)
			vs := versions[len(versions)-1].RawID
			creationArgs.OpenshiftVersion = vs
			creationArgs.ChannelGroup = exec.StableChannel
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("version %s is not in the list of supported versions", vs))
			restoreValues()
		})

		It("validate encryption fields - [id:72486]", ci.Medium, func() {
			By("Create cluster with etcd_encryption=true and no etcd_kms_key_arn")
			oldEtcdEncryption := creationArgs.Etcd
			oldEtcdKmsKeyArn := creationArgs.EtcdKmsKeyARN
			restoreValues := func() {
				creationArgs.Etcd = oldEtcdEncryption
				creationArgs.EtcdKmsKeyARN = oldEtcdKmsKeyArn
			}
			creationArgs.Etcd = helper.BoolPointer(true)
			creationArgs.EtcdKmsKeyARN = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "When utilizing etcd encryption an etcd kms key arn must also be supplied and vice versa")
			restoreValues()

			By("Create cluster with etcd_encryption=true and etcd_kms_key_arn wrong format")
			oldEtcdEncryption = creationArgs.Etcd
			oldEtcdKmsKeyArn = creationArgs.EtcdKmsKeyARN
			restoreValues = func() {
				creationArgs.Etcd = oldEtcdEncryption
				creationArgs.EtcdKmsKeyARN = oldEtcdKmsKeyArn
			}
			creationArgs.Etcd = helper.BoolPointer(true)
			creationArgs.EtcdKmsKeyARN = helper.StringPointer("anything")
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "expected the kms-key-arn: anything to match")
			restoreValues()

			By("Create cluster with etcd_encryption=true and etcd_kms_key_arn wrong arn")
			oldEtcdEncryption = creationArgs.Etcd
			oldEtcdKmsKeyArn = creationArgs.EtcdKmsKeyARN
			restoreValues = func() {
				creationArgs.Etcd = oldEtcdEncryption
				creationArgs.EtcdKmsKeyARN = oldEtcdKmsKeyArn
			}
			creationArgs.Etcd = helper.BoolPointer(true)
			creationArgs.EtcdKmsKeyARN = helper.StringPointer("arn:aws:kms:us-west-2:301721915996:key/9f1b5aee-3dc6-43d2-8c6e-793ca48c0c5c")
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Create a new one in the correct region, replace the ARN, and try again")
			restoreValues()

			By("Create cluster with kms_key_arn wrong format")
			oldKmsKeyArn := creationArgs.KmsKeyARN
			restoreValues = func() {
				creationArgs.KmsKeyARN = oldKmsKeyArn
			}
			creationArgs.KmsKeyARN = helper.StringPointer("anything")
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "expected the kms-key-arn: anything to match")
			restoreValues()

			By("Create cluster with kms_key_arn wrong arn")
			oldKmsKeyArn = creationArgs.KmsKeyARN
			restoreValues = func() {
				creationArgs.KmsKeyARN = oldKmsKeyArn
			}
			creationArgs.KmsKeyARN = helper.StringPointer("arn:aws:kms:us-west-2:301721915996:key/9f1b5aee-3dc6-43d2-8c6e-793ca48c0c5c")
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Create a new one in the correct region, replace the ARN, and try again")
			restoreValues()
		})

		It("validate proxy fields - [id:72491]", ci.Medium, func() {
			By("Create cluster with invalid http_proxy")
			oldProxy := creationArgs.Proxy
			restoreValues := func() {
				creationArgs.Proxy = oldProxy
			}
			creationArgs.Proxy = &exec.Proxy{
				HTTPProxy: helper.StringPointer("aaavvv"),
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Invalid 'proxy.http_proxy' attribute 'aaavvv'")
			restoreValues()

			By("Create cluster with http_proxy not starting with http")
			oldProxy = creationArgs.Proxy
			restoreValues = func() {
				creationArgs.Proxy = oldProxy
			}
			creationArgs.Proxy = &exec.Proxy{
				HTTPProxy: helper.StringPointer("https://aaavvv.test.nohttp.com/"),
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute 'proxy.http_proxy' prefix is not 'http'")
			restoreValues()

			By("Create cluster with invalid https_proxy")
			oldProxy = creationArgs.Proxy
			restoreValues = func() {
				creationArgs.Proxy = oldProxy
			}
			creationArgs.Proxy = &exec.Proxy{
				HTTPSProxy: helper.StringPointer("aaavvv"),
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Invalid 'proxy.https_proxy' attribute 'aaavvv'")
			restoreValues()

			By("Create cluster with invalid additional_trust_bundle")
			oldProxy = creationArgs.Proxy
			restoreValues = func() {
				creationArgs.Proxy = oldProxy
			}
			creationArgs.Proxy = &exec.Proxy{
				AdditionalTrustBundle: helper.StringPointer("/home/wrong_path/ca.cert"),
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Failed to parse additional_trust_bundle")
			restoreValues()

			By("Create cluster with no http/https proxy defined but no-proxy is set")
			oldProxy = creationArgs.Proxy
			restoreValues = func() {
				creationArgs.Proxy = oldProxy
			}
			creationArgs.Proxy = &exec.Proxy{
				NoProxy: helper.StringPointer("quay.io"),
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Either 'proxy.http_proxy' or 'proxy.https_proxy' attributes is needed to set 'proxy.no_proxy'")
			restoreValues()

			By("Create cluster with http proxy set and no-proxy=\"*\"")
			oldProxy = creationArgs.Proxy
			restoreValues = func() {
				creationArgs.Proxy = oldProxy
			}
			creationArgs.Proxy = &exec.Proxy{
				HTTPProxy: helper.StringPointer("http://example.com"),
				NoProxy:   helper.StringPointer("*"),
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "expected a valid user no-proxy value: '*' should match")
			restoreValues()
		})

		It("validate sts fields - [id:72496]", ci.Medium, func() {
			By("Create cluster with with empty installer role")
			oldInstallerRole := creationArgs.StsInstallerRole
			restoreValues := func() {
				creationArgs.StsInstallerRole = oldInstallerRole
			}
			creationArgs.StsInstallerRole = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "The 'aws.sts.role_arn' parameter is mandatory")
			restoreValues()

			By("Create cluster with with empty support role")
			oldSupportRole := creationArgs.StsSupportRole
			restoreValues = func() {
				creationArgs.StsSupportRole = oldSupportRole
			}
			creationArgs.StsSupportRole = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "The 'aws.sts.support_role_arn' parameter is mandatory")
			restoreValues()

			By("Create cluster with with empty worker role")
			oldWorkerRole := creationArgs.StsWorkerRole
			restoreValues = func() {
				creationArgs.StsWorkerRole = oldWorkerRole
			}
			creationArgs.StsWorkerRole = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute 'aws.sts.instance_iam_roles.worker_role_arn' is mandatory")
			restoreValues()

			By("Create cluster with with empty opreator role prefix")
			oldOperatorPrefix := creationArgs.OperatorRolePrefix
			restoreValues = func() {
				creationArgs.OperatorRolePrefix = oldOperatorPrefix
			}
			creationArgs.OperatorRolePrefix = helper.EmptyStringPointer
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Expected a 'aws.sts.operator_role_prefix' matching")
			restoreValues()
		})

		It("validate tags fields - [id:72627]", ci.Medium, func() {
			By("Create cluster with tag wrong key")
			oldTags := creationArgs.Tags
			restoreValues := func() {
				creationArgs.Tags = oldTags
			}
			creationArgs.Tags = map[string]string{
				"~~~": "cluster",
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute key 'aws.tags.~~~' invalid")
			restoreValues()

			By("Create cluster with tag wrong value")
			oldTags = creationArgs.Tags
			restoreValues = func() {
				creationArgs.Tags = oldTags
			}
			creationArgs.Tags = map[string]string{
				"name": "***",
			}
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Attribute value '***' of 'aws.tags.name' invalid")
			restoreValues()
		})

		It("validate network fields - [id:72468]", ci.Medium, func() {
			By("Retrieve VPC output")
			vpcService := exec.NewVPCService(constants.GetAWSVPCDefaultManifestDir(profile.GetClusterType()))
			vpcOutput, err := vpcService.Output()
			Expect(err).ToNot(HaveOccurred())

			By("Create cluster with wrong AZ name")
			oldAZs := creationArgs.AWSAvailabilityZones
			restoreValues := func() {
				creationArgs.AWSAvailabilityZones = oldAZs
			}
			creationArgs.AWSAvailabilityZones = helper.StringSlicePointer([]string{"us-west-2abhd"})
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Invalid availability zone: [us-west-2abhd]")
			restoreValues()

			By("Create cluster with AZ not in region name")
			az := "us-west-2a"
			if profile.Region == "us-west-2" {
				az = "us-east-1a"
			}
			oldAZs = creationArgs.AWSAvailabilityZones
			restoreValues = func() {
				creationArgs.AWSAvailabilityZones = oldAZs
			}
			creationArgs.AWSAvailabilityZones = helper.StringSlicePointer([]string{az})
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, fmt.Sprintf("Invalid AZ '%s' for region", az))
			restoreValues()

			By("Create cluster with wrong subnet")
			oldSubnetIDs := creationArgs.AWSSubnetIDs
			restoreValues = func() {
				creationArgs.AWSSubnetIDs = oldSubnetIDs
			}
			subnetIDs := []string{"subnet-08f6089e344f3e1f"}
			if !*creationArgs.Private {
				subnetIDs = append(subnetIDs, "subnet-08f6089e344f3e1d")
			}
			creationArgs.AWSSubnetIDs = helper.StringSlicePointer(subnetIDs)
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Failed to find subnet with ID 'subnet-08f6089e344f3e1f'")
			restoreValues()

			By("Create cluster with subnet from another VPC")
			// To implement with OCM-7807

			By("Create cluster with incorrect machine CIDR")
			oldMachineCIDR := creationArgs.MachineCIDR
			restoreValues = func() {
				creationArgs.MachineCIDR = oldMachineCIDR
			}
			creationArgs.MachineCIDR = "10.0.0.0/24"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "is outside of the machine CIDR range '10.0.0.0/24'")
			restoreValues()

			By("Create cluster with machine_cidr overlap with service_cidr")
			oldMachineCIDR = creationArgs.MachineCIDR
			oldServiceCIDR := creationArgs.ServiceCIDR
			restoreValues = func() {
				creationArgs.MachineCIDR = oldMachineCIDR
				creationArgs.ServiceCIDR = oldServiceCIDR
			}
			creationArgs.MachineCIDR = "10.0.0.0/16"
			creationArgs.ServiceCIDR = "10.0.0.0/20"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Machine CIDR '10.0.0.0/16' and service CIDR '10.0.0.0/20' overlap")
			restoreValues()

			By("Create cluster with machine_cidr overlap with pod_cidr")
			oldMachineCIDR = creationArgs.MachineCIDR
			oldPodCIDR := creationArgs.PodCIDR
			restoreValues = func() {
				creationArgs.MachineCIDR = oldMachineCIDR
				creationArgs.PodCIDR = oldPodCIDR
			}
			creationArgs.MachineCIDR = "10.0.0.0/16"
			creationArgs.PodCIDR = "10.0.0.0/18"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Machine CIDR '10.0.0.0/16' and pod CIDR '10.0.0.0/18' overlap")
			restoreValues()

			By("Create cluster with pod_cidr overlaps with default machine_cidr in AWS")
			oldPodCIDR = creationArgs.PodCIDR
			restoreValues = func() {
				creationArgs.PodCIDR = oldPodCIDR
			}
			creationArgs.PodCIDR = "10.0.0.0/16"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Machine CIDR '10.0.0.0/16' and pod CIDR '10.0.0.0/16' overlap")
			restoreValues()

			By("Create cluster with service_cidr overlap with pod_cidr")
			oldPodCIDR = creationArgs.PodCIDR
			oldServiceCIDR = creationArgs.ServiceCIDR
			restoreValues = func() {
				creationArgs.ServiceCIDR = oldServiceCIDR
				creationArgs.PodCIDR = oldPodCIDR
			}
			creationArgs.ServiceCIDR = "172.0.0.0/16"
			creationArgs.PodCIDR = "172.0.0.0/18"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Service CIDR '172.0.0.0/16' and pod CIDR '172.0.0.0/18' overlap")
			restoreValues()

			By("Create cluster  with CIDR without corresponding host prefix")
			oldPodCIDR = creationArgs.PodCIDR
			oldMachineCIDR = creationArgs.MachineCIDR
			restoreValues = func() {
				creationArgs.MachineCIDR = oldMachineCIDR
				creationArgs.PodCIDR = oldPodCIDR
			}
			creationArgs.MachineCIDR = "11.19.1.0/15"
			creationArgs.PodCIDR = "11.19.0.0/21"
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "network address '11.19.1.0' isn't consistent with network prefix 15")
			restoreValues()

			By("Create cluster with AZ and subnets not matching")
			if len(vpcOutput.AZs) > 1 {
				oldAZs = creationArgs.AWSAvailabilityZones
				oldSubnetIDs = creationArgs.AWSSubnetIDs
				restoreValues = func() {
					creationArgs.AWSAvailabilityZones = oldAZs
					creationArgs.AWSSubnetIDs = oldSubnetIDs
				}
				creationArgs.AWSAvailabilityZones = helper.StringSlicePointer([]string{vpcOutput.AZs[0]})
				subnetIDs := []string{vpcOutput.ClusterPrivateSubnets[1]}
				if !*creationArgs.Private {
					subnetIDs = append(subnetIDs, vpcOutput.ClusterPublicSubnets[1])
				}
				creationArgs.AWSSubnetIDs = helper.StringSlicePointer(subnetIDs)
				defer restoreValues()
				err = clusterService.Apply(creationArgs, false, false)
				Expect(err).To(HaveOccurred())
				helper.ExpectTFErrorContains(err, "does not belong to any of the provided zones. Provide a new subnet ID and try again.")
				restoreValues()
			} else {
				Logger.Infof("Not enough AZ to test this. Need at least 2 but found only %v", len(vpcOutput.AZs))
			}

			By("Create cluster with more AZ than corresponding subnets")
			if len(vpcOutput.AZs) > 1 {
				oldAZs = creationArgs.AWSAvailabilityZones
				oldSubnetIDs = creationArgs.AWSSubnetIDs
				restoreValues = func() {
					creationArgs.AWSAvailabilityZones = oldAZs
					creationArgs.AWSSubnetIDs = oldSubnetIDs
				}
				creationArgs.AWSAvailabilityZones = helper.StringSlicePointer([]string{vpcOutput.AZs[0], vpcOutput.AZs[1]})
				subnetIDs := []string{vpcOutput.ClusterPrivateSubnets[1]}
				if !*creationArgs.Private {
					subnetIDs = append(subnetIDs, vpcOutput.ClusterPublicSubnets[1])
				}
				creationArgs.AWSSubnetIDs = helper.StringSlicePointer(subnetIDs)
				defer restoreValues()
				err = clusterService.Apply(creationArgs, false, false)
				Expect(err).To(HaveOccurred())
				helper.ExpectTFErrorContains(err, "1 private subnet is required per zone")
				restoreValues()
			} else {
				Logger.Infof("Not enough AZ to test this. Need at least 2 but found only %v", len(vpcOutput.AZs))
			}

			By("Create cluster multiAZ with 3 private subnets and no replicas defined")
			if len(vpcOutput.AZs) > 2 {
				oldAZs = creationArgs.AWSAvailabilityZones
				oldSubnetIDs = creationArgs.AWSSubnetIDs
				oldReplicas := creationArgs.Replicas
				restoreValues = func() {
					creationArgs.AWSAvailabilityZones = oldAZs
					creationArgs.AWSSubnetIDs = oldSubnetIDs
					creationArgs.Replicas = oldReplicas
				}
				creationArgs.Replicas = 2
				creationArgs.AWSAvailabilityZones = helper.StringSlicePointer([]string{vpcOutput.AZs[0], vpcOutput.AZs[1], vpcOutput.AZs[2]})
				subnetIDs := []string{vpcOutput.ClusterPrivateSubnets[0], vpcOutput.ClusterPrivateSubnets[1], vpcOutput.ClusterPrivateSubnets[2]}
				if !*creationArgs.Private {
					subnetIDs = append(subnetIDs, vpcOutput.ClusterPublicSubnets[1])
				}
				creationArgs.AWSSubnetIDs = helper.StringSlicePointer(subnetIDs)
				defer restoreValues()
				err = clusterService.Apply(creationArgs, false, false)
				Expect(err).To(HaveOccurred())
				helper.ExpectTFErrorContains(err, "Hosted clusters require that the compute nodes be a multiple of the private subnets 3")
				restoreValues()
			} else {
				Logger.Infof("Not enough AZ to test this. Need at least 3 but found only %v", len(vpcOutput.AZs))
			}

			By("Create service with unsupported host prefix")
			oldHostPrefix := creationArgs.HostPrefix
			restoreValues = func() {
				creationArgs.HostPrefix = oldHostPrefix
			}
			creationArgs.HostPrefix = 22
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Invalid Network Host Prefix '22': Subnet length should be between '23' and '26")
			restoreValues()

			By("Remove Subnet tagging")
			restoreSubnetTagging := func() {
				By("Restore Subnet tagging")
				vpcArgs := exec.VPCArgs{
					DisableSubnetTagging: false,
				}
				err = vpcService.Apply(&vpcArgs, false)
			}
			defer restoreSubnetTagging()
			vpcArgs := exec.VPCArgs{
				DisableSubnetTagging: true,
			}
			err = vpcService.Apply(&vpcArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Create public cluster with public subnets without elb tag")
			oldSubnetIDs = creationArgs.AWSSubnetIDs
			oldPrivate := creationArgs.Private
			restoreValues = func() {
				creationArgs.AWSSubnetIDs = oldSubnetIDs
				creationArgs.Private = oldPrivate
			}
			creationArgs.Private = helper.BoolPointer(false)
			subnetIDs = *oldSubnetIDs
			subnetIDs = append(subnetIDs, vpcOutput.ClusterPublicSubnets[0])
			creationArgs.AWSSubnetIDs = helper.StringSlicePointer(subnetIDs)
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "The VPC needs to contain a public subnet with the tag 'kubernetes.io/role/elb'")
			restoreValues()

			By("Create private cluster with private subnets without internal-elb tag")
			oldSubnetIDs = creationArgs.AWSSubnetIDs
			oldPrivate = creationArgs.Private
			oldAZs = creationArgs.AWSAvailabilityZones
			restoreValues = func() {
				creationArgs.AWSSubnetIDs = oldSubnetIDs
				creationArgs.Private = oldPrivate
				creationArgs.AWSAvailabilityZones = oldAZs
			}
			creationArgs.Private = helper.BoolPointer(true)
			creationArgs.AWSSubnetIDs = helper.StringSlicePointer(vpcOutput.ClusterPrivateSubnets)
			creationArgs.AWSAvailabilityZones = helper.StringSlicePointer(vpcOutput.AZs)
			defer restoreValues()
			err = clusterService.Apply(creationArgs, false, false)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "The VPC needs to contain a private subnet with the tag 'kubernetes.io/role/internal-elb'")
			restoreValues()

			restoreSubnetTagging()
		})
	})
})
