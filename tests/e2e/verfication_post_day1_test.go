package e2e

import (
	// nolint

	"fmt"
	"net/http"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/rosa/pkg/aws"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	H "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("TF Test", func() {
	Describe("Verfication/Post day 1 tests", func() {
		var err error
		var profile *ci.Profile
		var cluster *cmv1.Cluster
		var importService = *exe.NewImportService(con.ImportResourceDir) // init new import service

		BeforeEach(func() {
			profile = ci.LoadProfileYamlFileByENV()
			Expect(err).ToNot(HaveOccurred())
			getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			cluster = getResp.Body()

		})
		AfterEach(func() {
		})

		Context("Author:smiron-High-OCP-63134 @OCP-63134 @smiron", func() {
			It("Verify cluster install was successful post cluster deployment", ci.Day1Post, ci.High, func() {
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
		})
		Context("Author:smiron-Medium-OCP-64906 @OCP-64906 @smiron", func() {
			It("Verify custom properties is set post cluster creation", ci.Day1Post, ci.Medium, func() {
				By("Check custom_property field is present under cluster's properties")
				getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(getResp.Body().Properties()["custom_property"]).To(Equal("test"))

				By("Check rosa_tf_commit is present under cluster's properties")
				Expect(getResp.Body().Properties()["rosa_tf_commit"]).ShouldNot(BeEmpty())

				By("Check rosa_tf_version is present under cluster's properties")
				Expect(getResp.Body().Properties()["rosa_tf_version"]).ShouldNot(BeEmpty())
			})
		})

		Context("Author:smiron-High-OCP-63140 @OCP-63140 @smiron", func() {
			It("Verify fips is enabled/disabled post cluster creation", ci.Day1Post, ci.High, func() {
				Expect(cluster.FIPS()).To(Equal(profile.FIPS))
			})
		})
		Context("Author:smiron-High-OCP-63133 @OCP-63133 @smiron", func() {
			It("Verify private_link is enabled/disabled post cluster creation", ci.Day1Post, ci.High, func() {
				Expect(cluster.AWS().PrivateLink()).To(Equal(profile.PrivateLink))
			})
		})
		Context("Author:smiron-High-OCP-63143 @OCP-63143 @smiron", func() {
			It("Verify etcd-encryption is enabled/disabled post cluster creation", ci.Day1Post, ci.High, func() {
				Expect(cluster.EtcdEncryption()).To(Equal(profile.Etcd))
			})
		})
		Context("Author:smiron-Medium-OCP-64023 @OCP-64023 @smiron", func() {
			It("Verify compute_machine_type value is set post cluster creation", ci.Day1Post, ci.Medium, func() {
				if profile.ComputeMachineType != "" {
					Expect(cluster.Nodes().ComputeMachineType().ID()).To(Equal(profile.ComputeMachineType))
				}
			})
		})
		Context("Author:smiron-Medium-OCP-63141 @OCP-63141 @smiron", func() {
			It("Verify availability zones and multi-az is set post cluster creation", ci.Day1Post, ci.Medium, func() {
				vpcService := exe.NewVPCService()
				zonesArray := strings.Split(profile.Zones, ",")
				clusterAvailZones := cluster.Nodes().AvailabilityZones()
				Expect(err).ToNot(HaveOccurred())
				Expect(cluster.MultiAZ()).To(Equal(profile.MultiAZ))
				if profile.Zones != "" {
					Expect(clusterAvailZones).
						To(Equal(H.JoinStringWithArray(profile.Region, zonesArray)))
				} else {
					vpcOut, _ := vpcService.Output()
					Expect(clusterAvailZones).To(Equal(vpcOut.AZs))
				}
			})
		})
		Context("Author:smiron-High-OCP-68423 @OCP-68423 @smiron", func() {
			It("Verify compute_labels are set post cluster creation", ci.Day1Post, ci.High, func() {
				if profile.Labeling {
					Expect(cluster.Nodes().ComputeLabels()).To(Equal(con.DefaultMPLabels))
				} else {
					Expect(cluster.Nodes().ComputeLabels()).To(Equal(con.NilMap))
				}
			})
		})
		Context("Author:smiron-High-OCP-63777 @OCP-63777 @smiron", func() {
			It("Verify AWS tags are set post cluster deployment", ci.Day1Post, ci.High, func() {

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
		})

		Context("Author:smiron-Medium-OCP-65684 @OCP-65684 @smiron", func() {
			It("OCP-65684 - cluster_rosa_classic resource can be import by the terraform import command",
				ci.Day2, ci.Medium, ci.FeatureImport, func() {

					By("Run the command to import rosa_classic resource")
					importParam := &exe.ImportArgs{
						Token:        token,
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
						Token:        token,
						ClusterID:    unknownClusterID,
						ResourceKind: "rhcs_cluster_rosa_classic",
						ResourceName: "rosa_import_no_cluster_id",
					}

					err = importService.Import(importParam)
					Expect(err.Error()).To(ContainSubstring("Cannot import non-existent remote object"))

					// skip due to bug :: OCM-5246
					// Expect(output).To(ContainSubstring(profile.ComputeMachineType))
				})
		})

		Context("Author:amalykhi-High-OCP-65928 @OCP-65928 @amalykhi", func() {
			It("Cluster admin during deployment - confirm user created ONLY during cluster creation operation",
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
		})
		Context("Author:xueli-Critical-OCP-69145 @OCP-69145 @xueli", func() {
			It("Create sts cluster with additional security group set will work via terraform provider",
				ci.Day1Post, ci.Critical,
				func() {
					By("Check the profile settings")
					if profile.AdditionalSGNumber == 0 {
						Expect(cluster.AWS().AdditionalComputeSecurityGroupIds()).To(BeEmpty())
						Expect(cluster.AWS().AdditionalControlPlaneSecurityGroupIds()).To(BeEmpty())
						Expect(cluster.AWS().AdditionalInfraSecurityGroupIds()).To(BeEmpty())
					} else {
						By("Verify CMS are using the correct configuration")
						clusterService, err := exe.NewClusterService(con.ROSAClassic)
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
			XIt("Apply to change security group will be forbidden", func() {
				clusterService, err := exe.NewClusterService(con.ROSAClassic)
				Expect(err).ToNot(HaveOccurred())
				outPut, err := clusterService.Output()
				Expect(err).ToNot(HaveOccurred())
				args := map[string]*exe.ClusterCreationArgs{
					"aws_additional_compute_security_group_ids": {
						Token:                                token,
						AdditionalComputeSecurityGroups:      outPut.AdditionalComputeSecurityGroups[0:1],
						AdditionalInfraSecurityGroups:        outPut.AdditionalInfraSecurityGroups,
						AdditionalControlPlaneSecurityGroups: outPut.AdditionalControlPlaneSecurityGroups,
						AWSRegion:                            profile.Region,
					},
					"aws_additional_infra_security_group_ids": {
						Token:                                token,
						AdditionalInfraSecurityGroups:        outPut.AdditionalInfraSecurityGroups[0:1],
						AdditionalComputeSecurityGroups:      outPut.AdditionalComputeSecurityGroups,
						AdditionalControlPlaneSecurityGroups: outPut.AdditionalControlPlaneSecurityGroups,
						AWSRegion:                            profile.Region,
					},
					"aws_additional_control_plane_security_group_ids": {
						Token:                                token,
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

		})
		Context("Author:xueli-Critical-OCP-69143 @OCP-69143 @xueli", func() {
			It("Create cluster with worker disk size will work via terraform provider",
				ci.Day1Post, ci.Critical,
				func() {
					switch profile.WorkerDiskSize {
					case 0:
						Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(300))
					default:
						Expect(cluster.Nodes().ComputeRootVolume().AWS().Size()).To(Equal(profile.WorkerDiskSize))
					}
				})
		})
		Context("Author:amalykhi-Medium-OCP-63138 @OCP-63138 @amalykhi", func() {
			It("Create sts cluster with account roles/policies unified path", ci.Day1Post, ci.Medium, func() {
				unifiedPath, err := aws.GetPathFromAccountRole(cluster, aws.AccountRoles[aws.InstallerAccountRole].Name)
				Expect(err).ToNot(HaveOccurred())
				Expect(profile.UnifiedAccRolesPath).Should(ContainSubstring(unifiedPath))
			})
		})
	})
})
