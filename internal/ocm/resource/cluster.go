package resource

***REMOVED***
	"errors"
***REMOVED***
	"regexp"

	"github.com/openshift-online/ocm-common/pkg/cluster/validations"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

var kmsArnRE = regexp.MustCompile(
	`^arn:aws[\w-]*:kms:[\w-]+:\d{12}:key\/mrk-[0-9a-f]{32}$|[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
***REMOVED***
var privateHostedZoneRoleArnRE = regexp.MustCompile(
	`^arn:aws:iam::\d{12}:role\/[A-Za-z0-9]+(?:-[A-Za-z0-9]+***REMOVED***+$`,
***REMOVED***

type Cluster struct {
	clusterBuilder *cmv1.ClusterBuilder
}

func NewCluster(***REMOVED*** *Cluster {
	return &Cluster{
		clusterBuilder: cmv1.NewCluster(***REMOVED***,
	}
}

func (c *Cluster***REMOVED*** GetClusterBuilder(***REMOVED*** *cmv1.ClusterBuilder {
	return c.clusterBuilder
}

func (c *Cluster***REMOVED*** Build(***REMOVED*** (object *cmv1.Cluster, err error***REMOVED*** {
	return c.clusterBuilder.Build(***REMOVED***
}

func (c *Cluster***REMOVED*** CreateNodes(autoScalingEnabled bool, replicas *int64, minReplicas *int64,
	maxReplicas *int64, computeMachineType *string, labels map[string]string,
	availabilityZones []string, multiAZ bool***REMOVED*** error {
	nodes := cmv1.NewClusterNodes(***REMOVED***
	if computeMachineType != nil {
		nodes.ComputeMachineType(
			cmv1.NewMachineType(***REMOVED***.ID(*computeMachineType***REMOVED***,
		***REMOVED***
	}

	if labels != nil {
		nodes.ComputeLabels(labels***REMOVED***
	}

	if availabilityZones != nil {
		if err := validations.ValidateAvailabilityZonesCount(multiAZ, len(availabilityZones***REMOVED******REMOVED***; err != nil {
			return err
***REMOVED***
		nodes.AvailabilityZones(availabilityZones...***REMOVED***
	}

	if autoScalingEnabled {
		if replicas != nil {
			return errors.New("When autoscaling is enabled, replicas should not be configured"***REMOVED***
***REMOVED***

		autoscaling := cmv1.NewMachinePoolAutoscaling(***REMOVED***
		minReplicasVal := 2
		if minReplicas != nil {
			minReplicasVal = int(*minReplicas***REMOVED***
***REMOVED***
		if err := validations.MinReplicasValidator(minReplicasVal, multiAZ, false, 0***REMOVED***; err != nil {
			return err
***REMOVED***
		autoscaling.MinReplicas(minReplicasVal***REMOVED***
		maxReplicasVal := 2
		if maxReplicas != nil {
			maxReplicasVal = int(*maxReplicas***REMOVED***
***REMOVED***
		if err := validations.MaxReplicasValidator(minReplicasVal, maxReplicasVal, multiAZ, false, 0***REMOVED***; err != nil {
			return err
***REMOVED***
		autoscaling.MaxReplicas(maxReplicasVal***REMOVED***
		if !autoscaling.Empty(***REMOVED*** {
			nodes.AutoscaleCompute(autoscaling***REMOVED***
***REMOVED***
	} else {
		if minReplicas != nil || maxReplicas != nil {
			return errors.New("Autoscaling must be enabled in order to set min and max replicas"***REMOVED***
***REMOVED***

		replicasVal := 2
		if replicas != nil {
			replicasVal = int(*replicas***REMOVED***
***REMOVED***
		if err := validations.MinReplicasValidator(replicasVal, multiAZ, false, 0***REMOVED***; err != nil {
			return err
***REMOVED***
		nodes.Compute(replicasVal***REMOVED***
	}

	if !nodes.Empty(***REMOVED*** {
		c.clusterBuilder.Nodes(nodes***REMOVED***
	}

	return nil
}

func (c *Cluster***REMOVED*** CreateAWSBuilder(awsTags map[string]string, ec2MetadataHttpTokens *string, kmsKeyARN *string,
	isPrivateLink bool, awsAccountID *string, stsBuilder *cmv1.STSBuilder, awsSubnetIDs []string,
	privateHostedZoneID *string, privateHostedZoneRoleARN *string***REMOVED*** error {

	if isPrivateLink && awsSubnetIDs == nil {
		return errors.New("Clusters with PrivateLink must have a pre-configured VPC. Make sure to specify the subnet ids."***REMOVED***
	}

	awsBuilder := cmv1.NewAWS(***REMOVED***

	if awsTags != nil {
		awsBuilder.Tags(awsTags***REMOVED***
	}

	ec2MetadataHttpTokensVal := cmv1.Ec2MetadataHttpTokensOptional
	if ec2MetadataHttpTokens != nil {
		ec2MetadataHttpTokensVal = cmv1.Ec2MetadataHttpTokens(*ec2MetadataHttpTokens***REMOVED***
	}
	awsBuilder.Ec2MetadataHttpTokens(ec2MetadataHttpTokensVal***REMOVED***

	if kmsKeyARN != nil {
		if !kmsArnRE.MatchString(*kmsKeyARN***REMOVED*** {
			return errors.New(fmt.Sprintf("Expected a valid value for kms-key-arn matching %s", kmsArnRE***REMOVED******REMOVED***
***REMOVED***
		awsBuilder.KMSKeyArn(*kmsKeyARN***REMOVED***
	}

	if awsAccountID != nil {
		awsBuilder.AccountID(*awsAccountID***REMOVED***
	}

	awsBuilder.PrivateLink(isPrivateLink***REMOVED***

	if awsSubnetIDs != nil {
		awsBuilder.SubnetIDs(awsSubnetIDs...***REMOVED***
	}

	if stsBuilder != nil {
		awsBuilder.STS(stsBuilder***REMOVED***
	}

	if privateHostedZoneID != nil && privateHostedZoneRoleARN != nil {
		if !privateHostedZoneRoleArnRE.MatchString(*privateHostedZoneRoleARN***REMOVED*** {
			return errors.New(fmt.Sprintf("Expected a valid value for PrivateHostedZoneRoleARN matching %s. Got %s", privateHostedZoneRoleArnRE, *privateHostedZoneRoleARN***REMOVED******REMOVED***
***REMOVED***
		if awsSubnetIDs == nil || stsBuilder == nil {
			return errors.New("PrivateHostedZone parameters require STS and SubnetIDs configurations."***REMOVED***
***REMOVED***
		awsBuilder.PrivateHostedZoneID(*privateHostedZoneID***REMOVED***
		awsBuilder.PrivateHostedZoneRoleARN(*privateHostedZoneRoleARN***REMOVED***
	}

	c.clusterBuilder.AWS(awsBuilder***REMOVED***

	return nil
}

func (c *Cluster***REMOVED*** SetAPIPrivacy(isPrivate bool, isPrivateLink bool, isSTS bool***REMOVED*** error {
	if isSTS && !isPrivate && isPrivateLink {
		return errors.New("PrivateLink is only supported on private clusters"***REMOVED***
	}
	api := cmv1.NewClusterAPI(***REMOVED***
	if isPrivate {
		api.Listening(cmv1.ListeningMethodInternal***REMOVED***
	} else {
		api.Listening(cmv1.ListeningMethodExternal***REMOVED***
	}
	c.clusterBuilder.API(api***REMOVED***
	return nil
}

func CreateSTS(installerRoleARN, supportRoleARN, masterRoleARN, workerRoleARN,
	operatorRolePrefix string, oidcConfigID *string***REMOVED*** *cmv1.STSBuilder {
	sts := cmv1.NewSTS(***REMOVED***
	sts.RoleARN(installerRoleARN***REMOVED***
	sts.SupportRoleARN(supportRoleARN***REMOVED***
	instanceIamRoles := cmv1.NewInstanceIAMRoles(***REMOVED***
	instanceIamRoles.MasterRoleARN(masterRoleARN***REMOVED***
	instanceIamRoles.WorkerRoleARN(workerRoleARN***REMOVED***
	sts.InstanceIAMRoles(instanceIamRoles***REMOVED***

	// set OIDC config ID
	if oidcConfigID != nil {
		sts.OidcConfig(cmv1.NewOidcConfig(***REMOVED***.ID(*oidcConfigID***REMOVED******REMOVED***
	}

	sts.OperatorRolePrefix(operatorRolePrefix***REMOVED***
	return sts
}
