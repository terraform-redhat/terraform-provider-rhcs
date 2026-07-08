// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-online/ocm-common/pkg/aws/aws_client"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Negative Tests", Ordered, ContinueOnFailure, func() {
	defer GinkgoRecover()

	var (
		profileHandler          profilehandler.ProfileHandler
		originalClusterVarsFile string
		clusterService          exec.ClusterService
	)

	retrieveOriginalClusterArgs := func() (clusterArgs *exec.ClusterArgs, err error) {
		if originalClusterVarsFile != "" {
			clusterArgs = &exec.ClusterArgs{}
			err = exec.ReadTerraformVarsFile(originalClusterVarsFile, clusterArgs)
		} else {
			err = fmt.Errorf("No original cluster file was setup prior to this method call. Please set it up")
		}
		return
	}

	validateClusterArgAgainstErrorSubstrings := func(updateFields func(args *exec.ClusterArgs), errSubStrings ...string) {
		clusterArgs, err := retrieveOriginalClusterArgs()
		Expect(err).ToNot(HaveOccurred())
		updateFields(clusterArgs)
		defer clusterService.Destroy()
		_, err = clusterService.Apply(clusterArgs)
		Expect(err).To(HaveOccurred())
		for _, errStr := range errSubStrings {
			helper.ExpectTFErrorContains(err, errStr)
		}
	}

	BeforeAll(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		originalClusterArgs, err := profileHandler.GenerateClusterCreationArgs(token)
		if err != nil {
			defer profileHandler.DestroyRHCSClusterResources(token)
		}
		Expect(err).ToNot(HaveOccurred())

		clusterService, err = profileHandler.Services().GetClusterService()
		if err != nil {
			defer profileHandler.DestroyRHCSClusterResources(token)
		}
		Expect(err).ToNot(HaveOccurred())

		// Save original cluster values before any update
		f, err := os.CreateTemp("", "tfvars-")
		Expect(err).ToNot(HaveOccurred())
		originalClusterVarsFile = f.Name()
		err = exec.WriteTFvarsFile(originalClusterArgs, originalClusterVarsFile)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		if originalClusterVarsFile != "" {
			exec.DeleteTFvarsFile(originalClusterVarsFile)
		}

		profileHandler.DestroyRHCSClusterResources(token)
	})

	Describe("cluster admin", ci.Day1Negative, func() {
		BeforeEach(OncePerOrdered, func() {
			if profileHandler.Profile().IsHCP() {
				Skip("Test can run only on Classic cluster")
			}
			if !profileHandler.Profile().IsAdminEnabled() {
				Skip("The tests configured for cluster admin only")
			}
		})

		It("validate user name policy - [id:65961]", ci.Medium,
			func() {
				By("Edit cluster admin user name to not valid")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					(*args.AdminCredentials)["username"] = "one:two"
				}, "Attribute admin_credentials.username username may not contain the characters: '/:%'")

				By("Edit cluster admin user name to empty")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					(*args.AdminCredentials)["username"] = ""
				}, "Attribute admin_credentials.username username may not be empty/blank string")
			})

		It("validate password policy - [id:65963]", ci.Medium, func() {
			By("Edit cluster admin password  to the short one")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				(*args.AdminCredentials)["password"] = helper.GenerateRandomPassword(13)
			}, "Attribute admin_credentials.password string length must be at least 14")

			By("Edit cluster admin password to empty")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				(*args.AdminCredentials)["password"] = ""
			}, "Attribute admin_credentials.password password should use ASCII-standard")

			By("Edit cluster admin password that lacks a capital letter")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				(*args.AdminCredentials)["password"] = strings.ToLower(helper.GenerateRandomPassword(14))
			}, "Attribute admin_credentials.password password must contain uppercase characters")

			By("Edit cluster admin password that lacks symbol but has digits")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				(*args.AdminCredentials)["password"] = "QwertyPasswordNoDigitsSymbols"
			}, "Attribute admin_credentials.password password must contain numbers or symbols")

			By("Edit cluster admin password that includes Non English chars")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				(*args.AdminCredentials)["password"] = "Qwert12345345@ש"
			}, "Attribute admin_credentials.password password should use ASCII-standard characters only")
		})
	})

	Describe("Create HCP cluster", ci.Day1Negative, func() {
		BeforeEach(OncePerOrdered, func() {
			if !profileHandler.Profile().IsHCP() {
				Skip("Test can run only on Hosted cluster")
			}
		})

		It("validate required fields - [id:72445]", ci.High, func() {
			By("Create cluster with wrong billing account")
			newValue := "012345678912"
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSBillingAccountID = new(newValue)
			}, fmt.Sprintf("billing account %s not linked to organization", newValue))

			By("Create cluster with wrong cloud region")
			newValue = "us-anything-2"
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSRegion = new(newValue)
			}, fmt.Sprintf("Invalid AWS Region: %s", newValue))

			By("Create cluster with cluster name > 54 characters")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ClusterName = new(helper.GenerateRandomName("cluster-72445", 50))
			}, "Attribute name string length must be at most 54")

			By("Create cluster with empty aws account id")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAccountID = helper.EmptyStringPointer
			}, "Attribute aws_account_id aws account ID must be only digits and exactly 12 in length")

			By("Create cluster with empty aws billing account id")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSBillingAccountID = helper.EmptyStringPointer
			}, "Attribute aws_billing_account_id aws billing account ID must be only digits and exactly 12 in length")

			By("Create cluster with empty name")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ClusterName = helper.EmptyStringPointer
			}, "'name' must consist of lower case characters or '-', start with an alphabetic character, and end with an alphanumeric character")

			By("Create cluster with empty cloud region")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSRegion = helper.EmptyStringPointer
			}, "Attribute 'region.id' is mandatory")

			By("Create cluster with empty sts")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OperatorRolePrefix = helper.EmptyStringPointer
				args.OIDCConfigID = helper.EmptyStringPointer
				args.StsInstallerRole = helper.EmptyStringPointer
				args.StsSupportRole = helper.EmptyStringPointer
				args.StsWorkerRole = helper.EmptyStringPointer
			}, "The 'aws.sts.role_arn' parameter is mandatory")

			By("Create cluster with empty subnet Ids")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSSubnetIDs = helper.EmptyStringSlicePointer
			}, "Hosted Control Plane requires at least one subnet")

			By("Create cluster with empty subnet azs")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAvailabilityZones = helper.EmptyStringSlicePointer
			}, "Availability zones must be provided for the subnet IDs provided")

			By("Create cluster without rosa_creator_arn property")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.IncludeCreatorProperty = new(false)
			}, "Expected property 'rosa_creator_arn'")

			By("Create cluster with wrong rosa_creator_arn property")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				(*args.CustomProperties)["rosa_creator_arn"] = "wrong_value"
				args.IncludeCreatorProperty = new(false)
			}, "Property 'rosa_creator_arn' does not have a valid user arn")
		})

		It("validate compute fields - [id:72449]", ci.Medium, func() {
			By("Create cluster with wrong compute machine type")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ComputeMachineType = new("anything")
			}, "Machine type 'anything' is not supported")

			By("Create cluster with replicas < 1")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Replicas = new(1)
			}, "A hosted cluster requires at least 2 replicas")
		})

		It("validate version fields - [id:72477]", ci.Medium, func() {
			By("Create cluster with wrong version")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OpenshiftVersion = new("4.14.2-rc")
			}, "version 4.14.2-rc is not in the list of supported versions")

			By("Create cluster with wrong channel group")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ChannelGroup = new("anything")
			}, "Could not find versions")

			By("Create cluster with version from another channel group")
			versions := cms.HCPEnabledVersions(cms.RHCSConnection, constants.VersionCandidateChannel)
			versions = cms.SortVersions(versions)
			vs := versions[len(versions)-1].RawID
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OpenshiftVersion = new(vs)
				args.ChannelGroup = new(constants.VersionStableChannel)
			}, fmt.Sprintf("version %s is not in the list of supported versions", vs))
		})

		It("validate encryption fields - [id:72486]", ci.Medium, func() {
			By("Create cluster with etcd_encryption=true and no etcd_kms_key_arn")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Etcd = new(true)
				args.EtcdKmsKeyARN = helper.EmptyStringPointer
			}, "When utilizing etcd encryption an etcd kms key arn must also be supplied and vice versa")

			By("Create cluster with etcd_encryption=true and etcd_kms_key_arn wrong format")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Etcd = new(true)
				args.EtcdKmsKeyARN = new("anything")
			}, "expected the kms-key-arn: anything to match")

			By("Create cluster with etcd_encryption=true and etcd_kms_key_arn wrong arn")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Etcd = new(true)
				args.EtcdKmsKeyARN = new("arn:aws:kms:us-west-2:301721915996:key/9f1b5aee-3dc6-43d2-8c6e-793ca48c0c5c")
			}, "Create a new one in the correct region, replace the ARN, and try again")

			By("Create cluster with kms_key_arn wrong format")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.KmsKeyARN = new("anything")
			}, "expected the kms-key-arn: anything to match")

			By("Create cluster with kms_key_arn wrong arn")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.KmsKeyARN = new("arn:aws:kms:us-west-2:301721915996:key/9f1b5aee-3dc6-43d2-8c6e-793ca48c0c5c")
			}, "Create a new one in the correct region, replace the ARN, and try again")
		})

		It("validate proxy fields - [id:72491]", ci.Medium, func() {
			By("Create cluster with invalid http_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy: new("aaavvv"),
				}
			}, "Invalid 'proxy.http_proxy' attribute 'aaavvv'")

			By("Create cluster with http_proxy not starting with http")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy: new("https://aaavvv.test.nohttp.com/"),
				}
			}, "Attribute 'proxy.http_proxy' prefix is not 'http'")

			By("Create cluster with invalid https_proxy")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPSProxy: new("aaavvv"),
				}
			}, "Invalid 'proxy.https_proxy' attribute 'aaavvv'")

			By("Create cluster with invalid additional_trust_bundle")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					AdditionalTrustBundle: new("/home/wrong_path/ca.cert"),
				}
			}, "Failed to parse additional_trust_bundle")

			By("Create cluster with no http/https proxy defined but no-proxy is set")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					NoProxy: new("quay.io"),
				}
			}, "Either 'proxy.http_proxy' or 'proxy.https_proxy' attributes is needed to set 'proxy.no_proxy'")

			By("Create cluster with http proxy set and no-proxy=\"*\"")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Proxy = &exec.Proxy{
					HTTPProxy: new("http://example.com"),
					NoProxy:   new("*"),
				}
			}, "expected a valid user no-proxy value: '*' should match")
		})

		It("validate sts fields - [id:72496]", ci.Medium, func() {
			By("Create cluster with with empty installer role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsInstallerRole = helper.EmptyStringPointer
			}, "The 'aws.sts.role_arn' parameter is mandatory")

			By("Create cluster with with empty support role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsSupportRole = helper.EmptyStringPointer
			}, "The 'aws.sts.support_role_arn' parameter is mandatory")

			By("Create cluster with with empty worker role")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.StsWorkerRole = helper.EmptyStringPointer
			}, "Attribute 'aws.sts.instance_iam_roles.worker_role_arn' is mandatory")

			By("Create cluster with with empty opreator role prefix")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OperatorRolePrefix = helper.EmptyStringPointer
			}, "Expected a 'aws.sts.operator_role_prefix' matching")
		})

		It("validate tags fields - [id:72627]", ci.Medium, func() {
			By("Create cluster with tag wrong key")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Tags = new(map[string]string{
					"~~~": "cluster",
				})
			}, "Attribute key 'aws.tags.~~~' invalid")

			By("Create cluster with tag wrong value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Tags = new(map[string]string{
					"name": "***",
				})
			}, "Attribute value '***' of 'aws.tags.name' invalid")
		})

		It("validate network fields - [id:72468]", ci.Medium, func() {
			By("Retrieve VPC output")
			vpcService, err := profileHandler.Services().GetVPCService()
			Expect(err).ToNot(HaveOccurred())
			vpcOutput, err := vpcService.Output()
			Expect(err).ToNot(HaveOccurred())

			By("Create cluster with wrong AZ name")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAvailabilityZones = new([]string{"us-west-2abhd"})
			}, "Invalid availability zone: [us-west-2abhd]")

			By("Create cluster with AZ not in region name")
			az := "us-west-2a"
			if profileHandler.Profile().GetRegion() == "us-west-2" {
				az = "us-east-1a"
			}
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.AWSAvailabilityZones = new([]string{az})
			}, fmt.Sprintf("Invalid AZ '%s' for region", az))

			By("Create cluster with wrong subnet")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				subnetIDs := []string{"subnet-08f6089e344f3e1f"}
				if !*args.Private {
					subnetIDs = append(subnetIDs, "subnet-08f6089e344f3e1d")
				}
				args.AWSSubnetIDs = new(subnetIDs)
			}, "Failed to find subnet with ID 'subnet-08f6089e344f3e1f'")

			By("Create cluster with subnet from another VPC")
			// To implement with OCM-7807

			By("Create cluster with incorrect machine CIDR")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.MachineCIDR = new("10.0.0.0/24")
			}, "is outside of the machine CIDR range '10.0.0.0/24'")

			By("Create cluster with machine_cidr overlap with service_cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.MachineCIDR = new("10.0.0.0/16")
				args.ServiceCIDR = new("10.0.0.0/20")
			}, "Machine CIDR '10.0.0.0/16' and service CIDR '10.0.0.0/20' overlap")

			By("Create cluster with machine_cidr overlap with pod_cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.MachineCIDR = new("10.0.0.0/16")
				args.PodCIDR = new("10.0.0.0/18")
			}, "Machine CIDR '10.0.0.0/16' and pod CIDR '10.0.0.0/18' overlap")

			By("Create cluster with pod_cidr overlaps with default machine_cidr in AWS")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.PodCIDR = new("10.0.0.0/16")
			}, "Machine CIDR '10.0.0.0/16' and pod CIDR '10.0.0.0/16' overlap")

			By("Create cluster with service_cidr overlap with pod_cidr")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.ServiceCIDR = new("172.0.0.0/16")
				args.PodCIDR = new("172.0.0.0/18")
			}, "Service CIDR '172.0.0.0/16' and pod CIDR '172.0.0.0/18' overlap")

			By("Create cluster  with CIDR without corresponding host prefix")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.MachineCIDR = new("11.19.1.0/15")
				args.PodCIDR = new("11.19.0.0/21")
			}, "network address '11.19.1.0' isn't consistent with network prefix 15")

			By("Create cluster with AZ and subnets not matching")
			if len(vpcOutput.AvailabilityZones) > 1 {
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.AWSAvailabilityZones = new([]string{vpcOutput.AvailabilityZones[0]})
					subnetIDs := []string{vpcOutput.PrivateSubnets[1]}
					if !*args.Private {
						subnetIDs = append(subnetIDs, vpcOutput.PublicSubnets[1])
					}
					args.AWSSubnetIDs = new(subnetIDs)
				}, "does not belong to any of the provided zones. Provide a new subnet ID and try again.")
			} else {
				Logger.Infof("Not enough AZ to test this. Need at least 2 but found only %v", len(vpcOutput.AvailabilityZones))
			}

			By("Create cluster with more AZ than corresponding subnets")
			if len(vpcOutput.AvailabilityZones) > 1 {
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.AWSAvailabilityZones = new([]string{vpcOutput.AvailabilityZones[0], vpcOutput.AvailabilityZones[1]})
					subnetIDs := []string{vpcOutput.PrivateSubnets[1]}
					if !*args.Private {
						subnetIDs = append(subnetIDs, vpcOutput.PublicSubnets[1])
					}
					args.AWSSubnetIDs = new(subnetIDs)
				}, "1 private subnet is required per zone")
			} else {
				Logger.Infof("Not enough AZ to test this. Need at least 2 but found only %v", len(vpcOutput.AvailabilityZones))
			}

			By("Create cluster multiAZ with 3 private subnets and no replicas defined")
			if len(vpcOutput.AvailabilityZones) > 2 {
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Replicas = new(2)
					args.AWSAvailabilityZones = new([]string{vpcOutput.AvailabilityZones[0], vpcOutput.AvailabilityZones[1], vpcOutput.AvailabilityZones[2]})
					subnetIDs := []string{vpcOutput.PrivateSubnets[0], vpcOutput.PrivateSubnets[1], vpcOutput.PrivateSubnets[2]}
					if !*args.Private {
						subnetIDs = append(subnetIDs, vpcOutput.PublicSubnets[1])
					}
					args.AWSSubnetIDs = new(subnetIDs)
				}, "Hosted clusters require that the compute nodes be a multiple of the private subnets 3")
			} else {
				Logger.Infof("Not enough AZ to test this. Need at least 3 but found only %v", len(vpcOutput.AvailabilityZones))
			}

			By("Create service with unsupported host prefix")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.HostPrefix = new(22)
			}, "Invalid Network Host Prefix '22': Subnet length should be between '23' and '26")

			By("Remove Subnets tagging")
			privateSubnetMandatoryTag := "kubernetes.io/role/internal-elb"
			publicSubnetMandatoryTag := "kubernetes.io/role/elb"
			awsClient, err := aws_client.CreateAWSClient("", "")
			Expect(err).To(BeNil())
			originalPublicSubnetsDetails, err := awsClient.ListSubnetDetail(vpcOutput.PublicSubnets...)
			Expect(err).To(BeNil())
			originalPrivateSubnetsDetails, err := awsClient.ListSubnetDetail(vpcOutput.PrivateSubnets...)
			Expect(err).To(BeNil())
			defer func() {
				for _, subnet := range originalPrivateSubnetsDetails {
					for _, tag := range subnet.Tags {
						if tag.Key == &privateSubnetMandatoryTag {
							tagMap := map[string]string{}
							tagMap[*tag.Key] = *tag.Value
							_, err = awsClient.TagResource(*subnet.SubnetId, tagMap)
							Expect(err).ToNot(HaveOccurred())
						}
					}
				}
				for _, subnet := range originalPublicSubnetsDetails {
					for _, tag := range subnet.Tags {
						if tag.Key == &publicSubnetMandatoryTag {
							tagMap := map[string]string{}
							tagMap[*tag.Key] = *tag.Value
							_, err = awsClient.TagResource(*subnet.SubnetId, tagMap)
							Expect(err).ToNot(HaveOccurred())
						}
					}
				}
			}()
			for _, subnet := range originalPrivateSubnetsDetails {
				for _, tag := range subnet.Tags {
					if tag.Key == &privateSubnetMandatoryTag {
						_, err = awsClient.RemoveResourceTag(*subnet.SubnetId, *tag.Key, *tag.Value)
						Expect(err).ToNot(HaveOccurred())
					}
				}
			}
			for _, subnet := range originalPublicSubnetsDetails {
				for _, tag := range subnet.Tags {
					if tag.Key == &publicSubnetMandatoryTag {
						_, err = awsClient.RemoveResourceTag(*subnet.SubnetId, *tag.Key, *tag.Value)
						Expect(err).ToNot(HaveOccurred())
					}
				}
			}

			By("Create public cluster with public subnets without elb tag")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Private = new(false)
				subnetIDs := append(*args.AWSSubnetIDs, vpcOutput.PublicSubnets[0])
				args.AWSSubnetIDs = new(subnetIDs)
			}, "The VPC needs to contain a public subnet with the tag 'kubernetes.io/role/elb'")

			By("Create private cluster with private subnets without internal-elb tag")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Private = new(true)
				args.AWSSubnetIDs = new(vpcOutput.PrivateSubnets)
				args.AWSAvailabilityZones = new(vpcOutput.AvailabilityZones)
			}, "The VPC needs to contain a private subnet with the tag 'kubernetes.io/role/internal-elb'")
		})

		It("validate imdsv2 fields - [id:75392]", ci.Medium, func() {
			By("Create cluster with invalid imdsv2 value")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.Ec2MetadataHttpTokens = new("invalid")
			}, "Expected a valid param. Options are [optional required]. Got invalid.")
		})
	})

	Describe("Create Classic or HCP cluster", ci.Day1Negative, func() {
		It("validate autoscaling setting - [id:88409]", ci.Medium,
			func() {
				By("Setting autoscaling_enable and replicas at the same time")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Autoscaling = new(true)
					args.Replicas = new(3)
				}, "When autoscaling is enabled, replicas should not be configured")

				By("autoscaling_enable=false and set min_replicas and max_replicas at the same time")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Autoscaling = new(false)
					args.MinReplicas = new(3)
					args.MaxReplicas = new(6)
				}, "Autoscaling must be enabled in order to set min and max replicas")
				By("autoscaling_enable=true and set min_replicas greater than max_replicas")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Autoscaling = new(true)
					args.Replicas = nil
					args.MinReplicas = new(6)
					args.MaxReplicas = new(3)
				}, "greater or equal to min-replicas")
				By("min_relicas with negative value")
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Autoscaling = new(true)
					args.Replicas = nil
					args.MinReplicas = new(-3)
					args.MaxReplicas = new(3)
				}, "The value for the number of cluster nodes must be non-negative")
				By("setting max_replicas to 0")
				maxReplicasZeroErr := "must be a integer greater than 0"
				if profileHandler.Profile().IsMultiAZ() {
					maxReplicasZeroErr = "Multi AZ cluster requires at least 3 compute nodes"
				}
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Autoscaling = new(true)
					args.Replicas = nil
					args.MinReplicas = new(0)
					args.MaxReplicas = new(0)
				}, maxReplicasZeroErr)
				By("with a number not a multiple of 3")
				multipleOfThreeErr := "require that the compute nodes be a multiple of the private subnets 3"
				if profileHandler.Profile().IsMultiAZ() {
					multipleOfThreeErr = "Multi AZ clusters require that the number of compute nodes be a multiple of 3"
				}
				validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
					args.Autoscaling = new(true)
					args.Replicas = nil
					args.MinReplicas = new(3)
					args.MaxReplicas = new(7)
				}, multipleOfThreeErr)
			})
		It("validate worker disk size - [id:76344]", ci.Low, func() {
			maxDiskSize := constants.MaxDiskSize
			minDiskSize := constants.MinClassicDiskSize
			if profileHandler.Profile().IsHCP() {
				minDiskSize = constants.MinHCPDiskSize
			}

			By("Create cluster with invalid worker disk size")
			errMsg := fmt.Sprintf("Must be between %d GiB and %d GiB", minDiskSize, maxDiskSize)
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.WorkerDiskSize = new(minDiskSize - 1)
			}, errMsg)
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.WorkerDiskSize = new(maxDiskSize + 1)
			}, errMsg)

			// TODO OCM-11521 terraform plan doesn't have validation
		})
		It("validate BYOVPC and deployment region match - [id:63825]", ci.Low, func() {
			if !profileHandler.Profile().IsBYOVPC() {
				Skip("Test requires BYOVPC")
			}

			By("Get the region of the BYOVPC")
			region := ""
			if profileHandler.Profile().GetRegion() == "us-east-2" {
				region = "us-west-2"
			} else {
				region = "us-east-2"
			}
			Expect(region).ToNot(BeEmpty())

			By("Set the deployment region to something else")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				var azs []string
				for _, az := range *args.AWSAvailabilityZones {
					newAz := strings.ReplaceAll(az, *args.AWSRegion, region)
					azs = append(azs, newAz)
				}
				args.AWSRegion = &region
				args.AWSAvailabilityZones = &azs
			}, "Failed to find subnet")
		})
	})

	Describe("The EOL OCP version validation", ci.Day1Negative, func() {
		It("version validation - [id:64095]", ci.Medium, func() {
			if profileHandler.Profile().GetAdditionalSGNumber() > 0 {
				Skip("Test is not made when security groups is enabled as the message will not be related to EOL support")
			}
			By("create cluster with an EOL OCP version")
			validateClusterArgAgainstErrorSubstrings(func(args *exec.ClusterArgs) {
				args.OpenshiftVersion = new("4.9.59")
			}, "version 4.9.59 is not in the list of supported versions")
		})
	})
})
