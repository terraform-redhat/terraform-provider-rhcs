// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterworkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	sprest "github.com/openshift-online/rosa-seaplane-api/clientset/rest"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

const seaplaneConditionSynced = "Synced"

var seaplaneAWSRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

type seaplaneBackend struct {
	service  exec.SeaplaneClusterService
	proxy    exec.ProxyService
	client   *seaplane.Clientset
	args     *exec.SeaplaneClusterArgs
	oidcName string
}

type seaplaneFetcher struct {
	client      *seaplane.Clientset
	service     exec.SeaplaneClusterService
	clusterName string
}

type seaplaneDestroyer struct {
	cluster     exec.SeaplaneClusterService
	iam         exec.SeaplaneIAMService
	oidc        exec.SeaplaneOidcConfigService
	proxy       exec.ProxyService
	vpc         exec.SeaplaneVPCService
	client      *seaplane.Clientset
	clusterName string
	ownsVPC     bool
	region      string
}

func newSeaplaneDestroyer(profile profilehandler.ProfileHandler, workspace string) (*seaplaneDestroyer, error) {
	clusterName := config.GetRHCSClusterName()
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	if workspace == "" {
		workspace = clusterName + "-seaplane"
	}
	if workspace != clusterName && workspace != clusterName+"-seaplane" {
		effectiveName := workspace
		for _, suffix := range []string{"-vpc", "-oidc", "-iam", "-proxy"} {
			if strings.HasSuffix(effectiveName, suffix) {
				effectiveName = strings.TrimSuffix(effectiveName, suffix)
				break
			}
		}
		if len(effectiveName) > 18 {
			hashSuffix := fmt.Sprintf("%03d", crc32.ChecksumIEEE([]byte(effectiveName))%1000)
			prefix := helper.TruncateString(effectiveName, 14)
			effectiveName = prefix + "-" + hashSuffix
		}
		clusterName = effectiveName
	}
	seaplaneURL, err := requiredEnv("SEAPLANE_URL")
	if err != nil {
		return nil, err
	}
	region := config.GetRegion()
	if region == "" {
		region = seaplaneAWSRegionRE.FindString(seaplaneURL)
	}
	if region == "" {
		return nil, fmt.Errorf(
			"AWS region is not set in REGION and cannot be derived from SEAPLANE_URL %q",
			seaplaneURL,
		)
	}
	clusterService, err := exec.NewSeaplaneClusterService(workspace)
	if err != nil {
		return nil, err
	}
	iamService, err := exec.NewSeaplaneIAMService(workspace + "-iam")
	if err != nil {
		return nil, err
	}
	oidcService, err := exec.NewSeaplaneOidcConfigService(workspace + "-oidc")
	if err != nil {
		return nil, err
	}
	var proxyService exec.ProxyService
	if profile.Profile().IsProxy() {
		proxyService, err = exec.NewProxyService(workspace+"-proxy", profile.Profile().GetClusterType())
		if err != nil {
			return nil, err
		}
	}
	vpcService, err := exec.NewSeaplaneVPCService(workspace + "-vpc")
	if err != nil {
		return nil, err
	}
	client, err := buildSeaplaneClient(seaplaneURL, region)
	if err != nil {
		return nil, err
	}
	// For secondary workspaces the computed clusterName changed over time
	// (truncate-18 -> hash suffix). If Terraform state already exists for
	// this workspace with the old name, reuse it to avoid a forced
	// replacement of the OIDC config that is blocked by DELETE-003 when
	// the cluster still references the old OIDC.
	if existingOut, err := clusterService.Output(); err == nil && existingOut != nil && existingOut.ClusterName != "" {
		if existingOut.ClusterName != clusterName {
			log.Logger.Infof("Workspace %q has existing cluster %q in TF state, reusing instead of computed %q", workspace, existingOut.ClusterName, clusterName)
			clusterName = existingOut.ClusterName
		}
	} else if existingOidcOut, err := oidcService.Output(); err == nil && existingOidcOut != nil && existingOidcOut.Name != "" {
		if trimmed := strings.TrimSuffix(existingOidcOut.Name, ".oidc"); trimmed != "" && trimmed != clusterName {
			log.Logger.Infof("Workspace %q has existing OIDC %q in TF state, reusing cluster name %q instead of computed %q", workspace, existingOidcOut.Name, trimmed, clusterName)
			clusterName = trimmed
		}
	}
	clusterID := os.Getenv(config.EnvClusterID)
	if clusterID == "" {
		output, outputErr := clusterService.Output()
		if outputErr != nil {
			return nil, outputErr
		}
		clusterID = output.ClusterID
	}
	// For secondary test clusters (e.g. DNS reservation), the destroyer is
	// created before the cluster exists. Don't require clusterID at this
	// point – Destroy will handle whatever Terraform state exists (including
	// none) and the Seaplane API deletion will be a no-op if the cluster
	// was never created.
	if clusterID == "" {
		log.Logger.Infof("ClusterID not found for workspace %q (pre-create), creating destroyer anyway", workspace)
	}
	return &seaplaneDestroyer{
		cluster:     clusterService,
		iam:         iamService,
		oidc:        oidcService,
		proxy:       proxyService,
		vpc:         vpcService,
		client:      client,
		clusterName: clusterName,
		ownsVPC:     len(config.GetSubnetIDList()) == 0 || len(config.GetAvailabilityZoneList()) == 0,
		region:      region,
	}, nil
}

func (d *seaplaneDestroyer) Destroy(ctx context.Context) error {
	if config.IsNoClusterDestroy() {
		return nil
	}
	if _, err := d.cluster.Destroy(); err != nil {
		return err
	}
	if err := waitForSeaplaneClusterDeletion(ctx, d.client, d.clusterName); err != nil {
		return err
	}
	if _, err := d.iam.Destroy(); err != nil {
		return err
	}
	if _, err := d.oidc.Destroy(); err != nil {
		return err
	}
	if d.proxy != nil {
		if _, err := d.proxy.ReadTFVars(); err == nil {
			if _, err := d.proxy.Destroy(); err != nil {
				return err
			}
		}
	}
	if d.ownsVPC {
		vpcOutput, err := d.vpc.Output()
		if err != nil {
			return fmt.Errorf("reading Seaplane VPC outputs before destroy: %w", err)
		}
		return DestroySeaplaneVPC(ctx, d.vpc, vpcOutput, d.region)
	}
	return nil
}

// DestroySeaplaneVPC removes operator-created dependencies before destroying
// the Terraform-managed VPC resources, retrying while asynchronous AWS cleanup completes.
func DestroySeaplaneVPC(
	ctx context.Context,
	service exec.SeaplaneVPCService,
	output *exec.SeaplaneVPCOutput,
	region string,
) error {
	var lastErr error
	err := wait.PollUntilContextTimeout(ctx, 2*time.Minute, 30*time.Minute, true, func(context.Context) (bool, error) {
		// The cluster operator releases VPC resources asynchronously after the
		// cluster object is deleted, and it leaks resources Terraform does not
		// track (classic ELBs, VPC endpoints, security groups, hosted-zone
		// records) that block terraform destroy. Re-prune them on every retry —
		// not just once — so transient states (e.g. a vpce-private-router
		// security group still pinned by an endpoint ENI that has not finished
		// releasing) resolve within the retry window instead of failing the
		// whole teardown on the first attempt.
		if output != nil {
			if cleanupErr := helper.DeleteClassicLoadBalancers(region, output.VPCID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] classic ELB cleanup failed (will retry): %v", cleanupErr)
				return false, nil
			}
			// Delete VPC endpoints before security groups: their managed ENIs
			// pin the <infra-id>-vpce-private-router security group, so the SG
			// cannot be removed until the endpoints are gone and their ENIs
			// have drained.
			if cleanupErr := helper.DeleteVPCEndpoints(region, output.VPCID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] VPC endpoint cleanup failed (will retry): %v", cleanupErr)
				return false, nil
			}
			if cleanupErr := helper.DeleteNonDefaultSecurityGroups(region, output.VPCID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] security group cleanup failed (will retry): %v", cleanupErr)
				return false, nil
			}
			// The cluster operator writes CNAME records (api.*, *.apps.*) into
			// the private <name>.hypershift.local hosted zone. terraform destroy
			// of aws_route53_zone.seaplane fails with HostedZoneNotEmpty while
			// those records remain, so purge them before destroying the VPC.
			if cleanupErr := helper.PurgeHostedZoneRecords(region, output.HostedZoneID); cleanupErr != nil {
				lastErr = cleanupErr
				log.Logger.Infof("[teardown] hosted zone purge failed (will retry): %v", cleanupErr)
				return false, nil
			}
		}
		_, lastErr = service.Destroy()
		if lastErr != nil {
			log.Logger.Infof("[teardown] VPC destroy attempt failed (will retry): %v", lastErr)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if lastErr != nil {
			return lastErr
		}
		return err
	}
	return nil
}

func newSeaplaneFetcher(profile profilehandler.ProfileHandler) (*seaplaneFetcher, error) {
	seaplaneURL, err := requiredEnv("SEAPLANE_URL")
	if err != nil {
		return nil, err
	}
	region := seaplaneAWSRegionRE.FindString(seaplaneURL)
	if region == "" {
		return nil, fmt.Errorf("cannot derive AWS region from SEAPLANE_URL %q", seaplaneURL)
	}
	client, err := buildSeaplaneClient(seaplaneURL, region)
	if err != nil {
		return nil, err
	}
	clusterName := config.GetRHCSClusterName()
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	service, err := exec.NewSeaplaneClusterService(clusterName)
	if err != nil {
		return nil, err
	}
	return &seaplaneFetcher{client: client, service: service, clusterName: clusterName}, nil
}

func (f *seaplaneFetcher) FetchClusterID(ctx context.Context) (string, error) {
	clusterID := os.Getenv(config.EnvClusterID)
	source := config.EnvClusterID
	if clusterID == "" {
		output, err := f.service.Output()
		if err != nil {
			// A create-focused run may not have Terraform state yet.
			return "", nil
		}
		clusterID = output.ClusterID
		source = "Terraform state"
	}
	if clusterID == "" {
		return "", nil
	}
	cluster, err := f.client.SeaplaneV1alpha1().Clusters().Get(ctx, f.clusterName, spplatform.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			if source == config.EnvClusterID {
				return "", fmt.Errorf(
					"seaplane cluster %q configured by %s=%q was not found; unset %s or point it at an existing cluster: %w",
					f.clusterName, source, clusterID, config.EnvClusterID, err,
				)
			}
			return "", fmt.Errorf(
				"terraform state for workspace %q references missing Seaplane cluster %q (ID %q); "+
					"recreate the cluster or remove the stale workspace state: %w",
				f.clusterName, f.clusterName, clusterID, err,
			)
		}
		return "", err
	}
	if string(cluster.UID) != clusterID {
		return "", fmt.Errorf(
			"configured Seaplane cluster ID %q does not match cluster %q UID %q",
			clusterID,
			f.clusterName,
			cluster.UID,
		)
	}
	return clusterID, nil
}

func newExistingSeaplaneBackend(profile profilehandler.ProfileHandler) (*Backend, error) {
	seaplaneURL, err := requiredEnv("SEAPLANE_URL")
	if err != nil {
		return nil, err
	}
	region := config.GetRegion()
	if region == "" {
		region = seaplaneAWSRegionRE.FindString(seaplaneURL)
	}
	if region == "" {
		return nil, fmt.Errorf("cannot derive AWS region from SEAPLANE_URL %q", seaplaneURL)
	}
	// The API cluster name may be overridden independently from the Terraform
	// workspace name. Existing profile-managed clusters keep their Terraform
	// state in the profile workspace, even when RHCS_CLUSTER_NAME is set.
	service, err := exec.NewSeaplaneClusterService(profile.Profile().GetName())
	if err != nil {
		return nil, err
	}
	args, err := service.ReadTFVars()
	if err != nil {
		return nil, err
	}
	if args.SeaplaneURL == nil || *args.SeaplaneURL == "" ||
		args.AWSRegion == nil || *args.AWSRegion == "" ||
		args.ClusterName == nil || *args.ClusterName == "" ||
		args.ReleaseImage == nil || *args.ReleaseImage == "" ||
		args.OperatorRolesPrefix == nil || *args.OperatorRolesPrefix == "" ||
		args.SubnetID == nil || *args.SubnetID == "" ||
		args.VPCID == nil || *args.VPCID == "" ||
		args.AvailabilityZone == nil || *args.AvailabilityZone == "" {
		return nil, fmt.Errorf("seaplane cluster Terraform state not found for workspace %q (missing required variables); create the cluster first via 'CLUSTER_PROFILE=%s go run ./hack/ensure-seaplane-cluster.go' or ensure SEAPLANE_URL/SEAPLANE_RELEASE_IMAGE are set and the profile workspace has been provisioned", profile.Profile().GetName(), profile.Profile().GetName())
	}
	client, err := buildSeaplaneClient(seaplaneURL, region)
	if err != nil {
		return nil, err
	}
	backend := &seaplaneBackend{service: service, client: client, args: args}
	clusterName := config.GetRHCSClusterName()
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	return &Backend{
		Lifecycle: backend,
		Fetcher:   &seaplaneFetcher{client: client, service: service, clusterName: clusterName},
	}, nil
}

func newSeaplaneBackend(profile profilehandler.ProfileHandler, workspace string) (*Backend, error) {
	return newSeaplaneBackendWithOptions(profile, workspace, SeaplaneOptions{})
}

func newSeaplaneBackendWithOptions(
	profile profilehandler.ProfileHandler,
	workspace string,
	options SeaplaneOptions,
) (*Backend, error) {
	if !profile.Profile().IsBYOVPC() {
		return nil, fmt.Errorf("seaplane cluster profiles must set byovpc: true")
	}
	releaseImage, err := requiredEnv(config.EnvSeaplaneReleaseImage)
	if err != nil {
		return nil, err
	}
	seaplaneURL, err := requiredEnv("SEAPLANE_URL")
	if err != nil {
		return nil, err
	}
	clusterName := config.GetRHCSClusterName()
	if clusterName == "" {
		clusterName = profile.Profile().GetName()
	}
	operatorRolesPrefix := clusterName
	if workspace == "" {
		workspace = clusterName + "-seaplane"
	}
	if workspace != clusterName && workspace != clusterName+"-seaplane" {
		effectiveName := workspace
		for _, suffix := range []string{"-vpc", "-oidc", "-iam", "-proxy"} {
			if strings.HasSuffix(effectiveName, suffix) {
				effectiveName = strings.TrimSuffix(effectiveName, suffix)
				break
			}
		}
		if len(effectiveName) > 18 {
			hashSuffix := fmt.Sprintf("%03d", crc32.ChecksumIEEE([]byte(effectiveName))%1000)
			prefix := helper.TruncateString(effectiveName, 14)
			effectiveName = prefix + "-" + hashSuffix
		}
		clusterName = effectiveName
		operatorRolesPrefix = effectiveName
		log.Logger.Infof("Secondary seaplane workspace %q mapped to cluster name %q to avoid OIDC 1:1 collision", workspace, clusterName)
	}
	// Preserve backward compatibility for workspaces that already have
	// Terraform state with the old truncate-18 naming (e.g.
	// sp-optional-conf-s). The hash-suffix change would otherwise force a
	// replacement of the OIDC config, which fails with DELETE-003 when a
	// cluster still references it. Check TF state before deciding final
	// clusterName for VPC/IAM/OIDC resources.
	tmpOidcService, tmpOidcErr := exec.NewSeaplaneOidcConfigService(workspace + "-oidc")
	if tmpOidcErr == nil {
		if existingOidcOut, err := tmpOidcService.Output(); err == nil && existingOidcOut != nil && existingOidcOut.Name != "" {
			if trimmed := strings.TrimSuffix(existingOidcOut.Name, ".oidc"); trimmed != "" && trimmed != clusterName {
				log.Logger.Infof("Workspace %q has existing OIDC %q in TF state, reusing cluster name %q instead of computed %q", workspace, existingOidcOut.Name, trimmed, clusterName)
				clusterName = trimmed
				operatorRolesPrefix = trimmed
			}
		}
		// Also check cluster output directly – it is more authoritative when
		// both exist; OIDC check above covers case where cluster not yet created.
		tmpClusterService, tmpClusterErr := exec.NewSeaplaneClusterService(workspace)
		if tmpClusterErr == nil {
			if existingClusterOut, err := tmpClusterService.Output(); err == nil && existingClusterOut != nil && existingClusterOut.ClusterName != "" {
				if existingClusterOut.ClusterName != clusterName {
					log.Logger.Infof("Workspace %q has existing cluster %q in TF state, reusing instead of computed %q", workspace, existingClusterOut.ClusterName, clusterName)
					clusterName = existingClusterOut.ClusterName
					operatorRolesPrefix = existingClusterOut.ClusterName
				}
			}
		}
	}
	region := config.GetRegion()
	if region == "" {
		region = seaplaneAWSRegionRE.FindString(seaplaneURL)
	}
	if region == "" {
		return nil, fmt.Errorf(
			"AWS region is not set in REGION and cannot be derived from SEAPLANE_URL %q",
			seaplaneURL,
		)
	}

	var vpcID, subnetID, publicSubnetID, availabilityZone string
	var vpcService exec.SeaplaneVPCService
	vpcCIDR := envOrDefault("SEAPLANE_VPC_CIDR", "10.0.0.0/16")
	subnetIDs := config.GetSubnetIDList()
	availabilityZones := config.GetAvailabilityZoneList()
	if len(availabilityZones) == 0 {
		availabilityZones = []string{region + "a"}
	}
	availabilityZone = availabilityZones[0]
	if len(subnetIDs) > 0 && len(availabilityZones) > 0 {
		subnetID = subnetIDs[0]
		vpcID, err = vpcIDFromSubnet(subnetID, region)
		if err != nil {
			return nil, err
		}
		if profile.Profile().IsProxy() {
			publicSubnetID, err = publicSubnetIDFromVPC(vpcID, region)
			if err != nil {
				return nil, err
			}
		}
	} else {
		vpcService, err = exec.NewSeaplaneVPCService(workspace + "-vpc")
		if err != nil {
			return nil, err
		}
		if _, err := vpcService.Apply(&exec.SeaplaneVPCArgs{
			AWSRegion:        &region,
			NamePrefix:       &clusterName,
			VPCCIDR:          &vpcCIDR,
			AvailabilityZone: &availabilityZone,
		}); err != nil {
			return nil, err
		}
		output, err := vpcService.Output()
		if err != nil {
			return nil, err
		}
		vpcID = output.VPCID
		subnetID = output.PrivateSubnetID
		publicSubnetID = output.PublicSubnetID
		availabilityZone = output.AvailabilityZone
	}

	service, err := exec.NewSeaplaneClusterService(workspace)
	if err != nil {
		return nil, err
	}
	client, err := buildSeaplaneClient(seaplaneURL, region)
	if err != nil {
		return nil, err
	}
	oidcService, err := exec.NewSeaplaneOidcConfigService(workspace + "-oidc")
	if err != nil {
		return nil, err
	}
	oidcType := profile.Profile().GetOIDCConfig()
	if oidcType == "" {
		oidcType = "managed"
	}
	oidcConfigName := clusterName + ".oidc"
	var oidcOutput *exec.SeaplaneOidcConfigOutput
	if _, err := oidcService.Apply(&exec.SeaplaneOidcConfigArgs{
		SeaplaneURL: &seaplaneURL,
		Name:        &oidcConfigName,
		Type:        &oidcType,
	}); err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "referenced by clusters") || strings.Contains(err.Error(), "DELETE-003") || strings.Contains(err.Error(), "Failed to delete OidcConfig") {
			ctx := context.Background()
			if existing, getErr := client.SeaplaneV1alpha1().OidcConfigs().Get(ctx, oidcConfigName, spplatform.GetOptions{}); getErr == nil && existing != nil {
				log.Logger.Infof("OIDC config %q already exists (not in workspace %q state), deleting orphan and retrying", oidcConfigName, workspace+"-oidc")
				if delErr := client.SeaplaneV1alpha1().OidcConfigs().Delete(ctx, oidcConfigName, spplatform.DeleteOptions{}); delErr != nil {
					if apierrors.IsConflict(delErr) || strings.Contains(delErr.Error(), "referenced by clusters") {
						log.Logger.Infof("OIDC config %q is referenced by existing cluster(s), reusing existing config", oidcConfigName)
						oidcOutput = &exec.SeaplaneOidcConfigOutput{
							OidcConfigID: string(existing.UID),
							Name:         existing.Name,
							IssuerURL:    existing.Spec.IssuerUrl,
							Phase:        string(existing.Status.Phase),
							Thumbprint:   existing.Status.Thumbprint,
						}
					} else {
						return nil, fmt.Errorf("oidc config %q already exists and failed to delete orphan: %w (original: %w)", oidcConfigName, delErr, err)
					}
				} else {
					time.Sleep(5 * time.Second)
					if _, err := oidcService.Apply(&exec.SeaplaneOidcConfigArgs{
						SeaplaneURL: &seaplaneURL,
						Name:        &oidcConfigName,
						Type:        &oidcType,
					}); err != nil {
						return nil, err
					}
					oidcOutput, err = oidcService.Output()
					if err != nil {
						return nil, err
					}
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		oidcOutput, err = oidcService.Output()
		if err != nil {
			return nil, err
		}
	}
	iamService, err := exec.NewSeaplaneIAMService(workspace + "-iam")
	if err != nil {
		return nil, err
	}
	if _, err := iamService.Apply(&exec.SeaplaneIAMArgs{
		AWSRegion:           &region,
		OperatorRolesPrefix: &operatorRolesPrefix,
		OIDCIssuerURL:       &oidcOutput.IssuerURL,
		OIDCThumbprint:      &oidcOutput.Thumbprint,
	}); err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "EntityAlreadyExists") {
			log.Logger.Infof("IAM roles for prefix %q already exist, reusing", operatorRolesPrefix)
		} else {
			return nil, err
		}
	}
	var proxyService exec.ProxyService
	var proxyOutput *exec.ProxyOutput
	if profile.Profile().IsProxy() {
		if publicSubnetID == "" {
			return nil, fmt.Errorf("seaplane proxy profile requires a public subnet in VPC %s", vpcID)
		}
		proxyService, err = exec.NewProxyService(workspace+"-proxy", profile.Profile().GetClusterType())
		if err != nil {
			return nil, err
		}
		proxyCount := 1
		trustBundlePath := filepath.Join(config.GetRHCSOutputDir(), clusterName+"-ca.cert")
		_, err = proxyService.Apply(&exec.ProxyArgs{
			ProxyCount:          &proxyCount,
			Region:              &region,
			VPCID:               &vpcID,
			PublicSubnetID:      &publicSubnetID,
			TrustBundleFilePath: &trustBundlePath,
			KeyPairID:           &clusterName,
		})
		if err != nil {
			return nil, cleanupSeaplaneProxyOnError(proxyService, fmt.Errorf("creating Seaplane proxy endpoint: %w", err))
		}
		proxies, outputErr := proxyService.Output()
		if outputErr != nil {
			return nil, cleanupSeaplaneProxyOnError(
				proxyService, fmt.Errorf("reading Seaplane proxy endpoint: %w", outputErr),
			)
		}
		if proxies == nil || len(proxies.Proxies) == 0 {
			return nil, cleanupSeaplaneProxyOnError(
				proxyService, fmt.Errorf("seaplane proxy Terraform output has no endpoints"),
			)
		}
		proxyOutput = proxies.Proxies[0]
	}
	schedulerProfile := profile.Profile().GetSchedulerProfile()
	var schedulerProfileValue *string
	if schedulerProfile != "" {
		schedulerProfileValue = &schedulerProfile
	}
	var maxProvision *string
	if value := profile.Profile().GetAutoscalerMaxNodeProvisionTime(); value != "" {
		maxProvision = &value
	}
	var maxNodes, maxGrace, podPriority *int64
	if value := profile.Profile().GetAutoscalerMaxNodesTotal(); value != 0 {
		maxNodes = &value
	}
	if value := profile.Profile().GetAutoscalerMaxPodGracePeriod(); value != 0 {
		maxGrace = &value
	}
	if value := profile.Profile().GetAutoscalerPodPriorityThreshold(); value != 0 {
		podPriority = &value
	}
	componentRoutes := profile.Profile().GetComponentRoutes()
	var componentRoutesValue *string
	if componentRoutes != "" {
		componentRoutesValue = &componentRoutes
	}
	// Capabilities are JSON-encoded lists (matching the cluster manifest
	// variables); an empty profile list leaves the attribute unset.
	jsonList := func(values []string) *string {
		if len(values) == 0 {
			return nil
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			// json.Marshal on a []string cannot practically fail; treat it as a
			// test-harness bug rather than silently creating a cluster without
			// the requested capabilities.
			panic(fmt.Sprintf("jsonList: %v", err))
		}
		encodedStr := string(encoded)
		return &encodedStr
	}
	optionalString := func(value string) *string {
		if value == "" {
			return nil
		}
		return &value
	}
	optionalInt64 := func(value int64) *int64 {
		if value == 0 {
			return nil
		}
		return &value
	}
	// Machine network: the profile value wins, otherwise default to the VPC
	// CIDR (rosa CLI parity — nodes live in that VPC).
	machineCIDR := optionalString(profile.Profile().GetMachineCIDR())
	if machineCIDR == nil {
		machineCIDR = &vpcCIDR
	}
	// API server allow list: default fully open (rosa CLI parity).
	apiAllowedCIDRBlocks := `["0.0.0.0/0"]`
	clusterArgs := &exec.SeaplaneClusterArgs{
		SeaplaneURL:                    &seaplaneURL,
		ReleaseImage:                   &releaseImage,
		AWSRegion:                      &region,
		ClusterName:                    &clusterName,
		OperatorRolesPrefix:            &operatorRolesPrefix,
		SubnetID:                       &subnetID,
		VPCID:                          &vpcID,
		AvailabilityZone:               &availabilityZone,
		OIDCConfigID:                   &oidcOutput.OidcConfigID,
		SchedulerProfile:               schedulerProfileValue,
		AutoscalerMaxNodeProvisionTime: maxProvision,
		AutoscalerMaxNodesTotal:        maxNodes,
		AutoscalerMaxPodGracePeriod:    maxGrace,
		AutoscalerPodPriorityThreshold: podPriority,
		ComponentRoutes:                componentRoutesValue,
		EnabledCapabilities:            jsonList(profile.Profile().GetEnableClusterCapabilities()),
		DisabledCapabilities:           jsonList(profile.Profile().GetDisableClusterCapabilities()),
		// The rosa CLI bridge always defaults the machine network to the VPC
		// CIDR and opens the API server allow list; mirror both so TF-created
		// clusters match CLI-created ones (nodes must reach the API server
		// through the load balancer to register).
		MachineCIDR:          machineCIDR,
		APIAllowedCIDRBlocks: &apiAllowedCIDRBlocks,
		PodCIDR:              optionalString(profile.Profile().GetPodCIDR()),
		ServiceCIDR:          optionalString(profile.Profile().GetServiceCIDR()),
		HostPrefix:           optionalInt64(int64(profile.Profile().GetHostPrefix())),
	}
	if options.DNSReservationID != "" {
		clusterArgs.DNSReservationID = &options.DNSReservationID
	}
	clusterArgs.BaseDomainPrefix = options.BaseDomainPrefix
	// Set the profile's initial customer AWS tags. They can be updated in place
	// after the cluster is created.
	if profile.Profile().IsTagging() {
		tags := profilehandler.Tags
		clusterArgs.Tags = &tags
	}
	if proxyOutput != nil {
		clusterArgs.HTTPProxy = &proxyOutput.HttpProxy
		clusterArgs.HTTPSProxy = &proxyOutput.HttpsProxy
		clusterArgs.NoProxy = &proxyOutput.NoProxy
	}
	backend := &seaplaneBackend{
		service:  service,
		proxy:    proxyService,
		client:   client,
		oidcName: oidcOutput.Name,
		args:     clusterArgs,
	}
	return &Backend{Lifecycle: backend}, nil
}

func (b *seaplaneBackend) FetchClusterID(ctx context.Context) (string, error) {
	clusterID := os.Getenv(config.EnvClusterID)
	if clusterID == "" {
		return "", nil
	}
	clusterName := ""
	if b.args.ClusterName != nil {
		clusterName = *b.args.ClusterName
	}
	if clusterName == "" {
		return "", fmt.Errorf("seaplane cluster name is required to validate cluster ID %q", clusterID)
	}
	cluster, err := b.client.SeaplaneV1alpha1().Clusters().Get(ctx, clusterName, spplatform.GetOptions{})
	if err != nil {
		return "", err
	}
	if string(cluster.UID) != clusterID {
		return "", fmt.Errorf(
			"configured Seaplane cluster ID %q does not match cluster %q UID %q",
			clusterID,
			clusterName,
			cluster.UID,
		)
	}
	return clusterID, nil
}

func (b *seaplaneBackend) Create() (string, error) {
	if _, err := b.service.Apply(b.args); err != nil {
		if output, outputErr := b.service.Output(); outputErr != nil || output == nil || output.ClusterID == "" {
			if b.proxy != nil {
				if _, varsErr := b.proxy.ReadTFVars(); varsErr == nil {
					if _, cleanupErr := b.proxy.Destroy(); cleanupErr != nil {
						return "", fmt.Errorf("%w (also failed to destroy Seaplane proxy endpoint: %v)", err, cleanupErr)
					}
				}
			}
		}
		return "", err
	}
	output, err := b.service.Output()
	if err != nil {
		return "", err
	}
	return output.ClusterID, nil
}

func cleanupSeaplaneProxyOnError(proxy exec.ProxyService, cause error) error {
	if _, varsErr := proxy.ReadTFVars(); varsErr != nil {
		return cause
	}
	if _, cleanupErr := proxy.Destroy(); cleanupErr != nil {
		return fmt.Errorf("%w (also failed to destroy Seaplane proxy endpoint: %v)", cause, cleanupErr)
	}
	return cause
}

// Update applies backend-neutral cluster settings to the Terraform-managed
// Seaplane cluster.
func (b *seaplaneBackend) Update(spec exec.ClusterSpec) error {
	args := &exec.SeaplaneClusterArgs{}
	if spec.Autoscaler != nil {
		args.AutoscalerMaxNodeProvisionTime = spec.Autoscaler.MaxNodeProvisionTime
		args.AutoscalerMaxNodesTotal = spec.Autoscaler.MaxNodesTotal
		args.AutoscalerMaxPodGracePeriod = spec.Autoscaler.MaxPodGracePeriod
		args.AutoscalerPodPriorityThreshold = spec.Autoscaler.PodPriorityThreshold
	}
	args.ComponentRoutes = spec.ComponentRoutes
	args.SchedulerProfile = spec.SchedulerProfile
	return b.update(args)
}

func (b *seaplaneBackend) update(args *exec.SeaplaneClusterArgs) error {
	if args.SeaplaneURL != nil {
		b.args.SeaplaneURL = args.SeaplaneURL
	}
	if args.AWSRegion != nil {
		b.args.AWSRegion = args.AWSRegion
	}
	if args.ClusterName != nil {
		b.args.ClusterName = args.ClusterName
	}
	if args.OperatorRolesPrefix != nil {
		b.args.OperatorRolesPrefix = args.OperatorRolesPrefix
	}
	if args.SubnetID != nil {
		b.args.SubnetID = args.SubnetID
	}
	if args.VPCID != nil {
		b.args.VPCID = args.VPCID
	}
	if args.AvailabilityZone != nil {
		b.args.AvailabilityZone = args.AvailabilityZone
	}
	if args.ExpirationTimestamp != nil {
		b.args.ExpirationTimestamp = args.ExpirationTimestamp
	}
	if args.OIDCConfigID != nil {
		b.args.OIDCConfigID = args.OIDCConfigID
	}
	if args.SchedulerProfile != nil {
		b.args.SchedulerProfile = args.SchedulerProfile
	}
	if args.AutoscalerMaxNodeProvisionTime != nil {
		b.args.AutoscalerMaxNodeProvisionTime = args.AutoscalerMaxNodeProvisionTime
	}
	if args.AutoscalerMaxNodesTotal != nil {
		b.args.AutoscalerMaxNodesTotal = args.AutoscalerMaxNodesTotal
	}
	if args.AutoscalerMaxPodGracePeriod != nil {
		b.args.AutoscalerMaxPodGracePeriod = args.AutoscalerMaxPodGracePeriod
	}
	if args.AutoscalerPodPriorityThreshold != nil {
		b.args.AutoscalerPodPriorityThreshold = args.AutoscalerPodPriorityThreshold
	}
	if args.ComponentRoutes != nil {
		b.args.ComponentRoutes = args.ComponentRoutes
	}
	if args.Tags != nil {
		b.args.Tags = args.Tags
	}
	_, err := b.service.Apply(b.args)
	return err
}

func (b *seaplaneBackend) Read(_ context.Context) (exec.ClusterState, error) {
	output, err := b.service.Output()
	if err != nil {
		return exec.ClusterState{}, err
	}
	return exec.ClusterState{
		ClusterID:        output.ClusterID,
		ClusterName:      output.ClusterName,
		Phase:            output.Phase,
		SchedulerProfile: output.SchedulerProfile,
		ComponentRoutes:  output.ComponentRoutes,
		Autoscaler: &exec.ClusterAutoscalerState{
			MaxNodeProvisionTime: output.AutoscalerMaxNodeProvisionTime,
			MaxNodesTotal:        output.AutoscalerMaxNodesTotal,
			MaxPodGracePeriod:    output.AutoscalerMaxPodGracePeriod,
			PodPriorityThreshold: output.AutoscalerPodPriorityThreshold,
		},
	}, nil
}

func (b *seaplaneBackend) WaitReady(ctx context.Context, _ string) error {
	if err := b.client.SeaplaneV1alpha1().OidcConfigs().WaitUntil(
		ctx,
		b.oidcName,
		func(config *v1alpha1.OidcConfig) bool {
			if config == nil {
				log.Logger.Infof("[seaplane] OIDC config %s not found while waiting for Ready", b.oidcName)
				return false
			}
			log.Logger.Infof("[seaplane] OIDC config %s phase: %s", b.oidcName, config.Status.Phase)
			return config.Status.Phase == v1alpha1.OidcConfigPhaseReady
		},
		30*time.Second,
		15*time.Minute,
	); err != nil {
		return err
	}
	clusterName := ""
	if b.args.ClusterName != nil {
		clusterName = *b.args.ClusterName
	}
	if clusterName == "" {
		return fmt.Errorf("seaplane cluster name is required while waiting for Ready")
	}
	return b.client.SeaplaneV1alpha1().Clusters().WaitUntil(
		ctx,
		clusterName,
		func(cluster *v1alpha1.Cluster) bool {
			if cluster == nil {
				log.Logger.Infof("[seaplane] cluster %s not found while waiting for Ready", clusterName)
				return false
			}
			log.Logger.Infof("[seaplane] cluster %s phase: %s", clusterName, cluster.Status.Phase)
			return cluster.Status.Phase == v1alpha1.ClusterPhaseReady
		},
		30*time.Second,
		90*time.Minute,
	)
}

func waitForSeaplaneClusterDeletion(ctx context.Context, client *seaplane.Clientset, clusterName string) error {
	return client.SeaplaneV1alpha1().Clusters().WaitUntil(
		ctx,
		clusterName,
		func(cluster *v1alpha1.Cluster) bool {
			if cluster == nil {
				log.Logger.Infof("[seaplane] cluster %s deleted", clusterName)
				return true
			}
			log.Logger.Infof("[seaplane] cluster %s phase while deleting: %s", clusterName, cluster.Status.Phase)
			return false
		},
		30*time.Second,
		60*time.Minute,
	)
}

func vpcIDFromSubnet(subnetID, region string) (string, error) {
	ctx := context.Background()
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return "", err
	}
	output, err := ec2.NewFromConfig(awsConfig).
		DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{SubnetIds: []string{subnetID}})
	if err != nil {
		return "", err
	}
	if len(output.Subnets) == 0 || output.Subnets[0].VpcId == nil || *output.Subnets[0].VpcId == "" {
		return "", fmt.Errorf("subnet %q did not return a VPC ID", subnetID)
	}
	return *output.Subnets[0].VpcId, nil
}

func publicSubnetIDFromVPC(vpcID, region string) (string, error) {
	ctx := context.Background()
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("loading AWS config to find a public subnet in VPC %s: %w", vpcID, err)
	}
	ec2Client := ec2.NewFromConfig(awsConfig)
	subnets, err := ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{{Name: aws.String("vpc-id"), Values: []string{vpcID}}},
	})
	if err != nil {
		return "", fmt.Errorf("listing subnets in VPC %s: %w", vpcID, err)
	}
	routeTables, err := ec2Client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []ec2types.Filter{{Name: aws.String("vpc-id"), Values: []string{vpcID}}},
	})
	if err != nil {
		return "", fmt.Errorf("listing route tables in VPC %s: %w", vpcID, err)
	}

	routeTableBySubnet := make(map[string]*ec2types.RouteTable)
	var mainRouteTable *ec2types.RouteTable
	for i := range routeTables.RouteTables {
		routeTable := &routeTables.RouteTables[i]
		for _, association := range routeTable.Associations {
			if association.Main != nil && *association.Main {
				mainRouteTable = routeTable
			}
			if association.SubnetId != nil {
				routeTableBySubnet[*association.SubnetId] = routeTable
			}
		}
	}
	for _, subnet := range subnets.Subnets {
		if subnet.SubnetId == nil {
			continue
		}
		routeTable := routeTableBySubnet[*subnet.SubnetId]
		if routeTable == nil {
			routeTable = mainRouteTable
		}
		if routeTable == nil {
			continue
		}
		for _, route := range routeTable.Routes {
			if route.GatewayId != nil && strings.HasPrefix(*route.GatewayId, "igw-") {
				return *subnet.SubnetId, nil
			}
		}
	}
	return "", fmt.Errorf("VPC %s has no subnet with a route to an internet gateway", vpcID)
}

func buildSeaplaneClient(seaplaneURL, region string) (*seaplane.Clientset, error) {
	ctx := context.Background()
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	identity, err := sts.NewFromConfig(awsConfig).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return seaplane.NewForConfig(&sprest.Config{
		Host:      seaplaneURL,
		Region:    region,
		AccountID: *identity.Account,
		CallerARN: *identity.Arn,
		AWSConfig: awsConfig,
	})
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}
	return value, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

var _ exec.ClusterLifecycle = (*seaplaneBackend)(nil)
