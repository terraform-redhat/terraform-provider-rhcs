package logforwarder

import "github.com/hashicorp/terraform-plugin-framework/types"

type LogForwarder struct {
	Cluster      types.String `tfsdk:"cluster"`
	ID           types.String `tfsdk:"id"`
	S3           types.Object `tfsdk:"s3"`
	CloudWatch   types.Object `tfsdk:"cloudwatch"`
	Applications types.List   `tfsdk:"applications"`
	Groups       types.List   `tfsdk:"groups"`
}

type S3Config struct {
	BucketName   types.String `tfsdk:"bucket_name"`
	BucketPrefix types.String `tfsdk:"bucket_prefix"`
}

type CloudWatchConfig struct {
	LogGroupName           types.String `tfsdk:"log_group_name"`
	LogDistributionRoleArn types.String `tfsdk:"log_distribution_role_arn"`
}

type LogForwarderGroup struct {
	ID      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
}
