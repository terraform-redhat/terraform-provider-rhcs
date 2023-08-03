package clusterschema

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"strings"
)

func StsFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"oidc_endpoint_url": {
			Description: "OIDC Endpoint URL",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"oidc_config_id": {
			Description: "OIDC Configuration ID",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"thumbprint": {
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"role_arn": {
			Description: "Installer Role",
			Type:        schema.TypeString,
			Required:    true,
		},
		"support_role_arn": {
			Description: "Support Role",
			Type:        schema.TypeString,
			Required:    true,
		},
		"instance_iam_roles": {
			Description: "Instance IAM Roles",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: instanceRoleFields(),
			},
			Required: true,
		},
		"operator_role_prefix": {
			Description: "Operator IAM Role prefix",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func instanceRoleFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"master_role_arn": {
			Description: "Master/Controller Plane Role ARN",
			Type:        schema.TypeString,
			Required:    true,
		},
		"worker_role_arn": {
			Description: "Worker Node Role ARN",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

type Sts struct {
	// required
	RoleARN            string          `tfsdk:"role_arn"`
	SupportRoleArn     string          `tfsdk:"support_role_arn"`
	InstanceIAMRoles   InstanceIAMRole `tfsdk:"instance_iam_roles"`
	OperatorRolePrefix string          `tfsdk:"operator_role_prefix"`

	// optional
	OIDCEndpointURL *string `tfsdk:"oidc_endpoint_url"`
	OIDCConfigID    *string `tfsdk:"oidc_config_id"`

	// computed
	Thumbprint string `tfsdk:"thumbprint"`
}

type InstanceIAMRole struct {
	MasterRoleARN string `tfsdk:"master_role_arn"`
	WorkerRoleARN string `tfsdk:"worker_role_arn"`
}

func ExpandInstanceIamRolesFromInterface(i interface{}) InstanceIAMRole {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return InstanceIAMRole{}
	}

	instanceRoleMap := l[0].(map[string]interface{})
	return InstanceIAMRole{
		MasterRoleARN: instanceRoleMap["master_role_arn"].(string),
		WorkerRoleARN: instanceRoleMap["worker_role_arn"].(string),
	}
}

func ExpandStsFromResourceData(resourceData *schema.ResourceData) *Sts {
	list, ok := resourceData.GetOk("sts")
	if !ok {
		return nil
	}

	return ExpandStsFromInterface(list)
}

func ExpandStsFromInterface(i interface{}) *Sts {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	stsMap := l[0].(map[string]interface{})

	return &Sts{
		RoleARN:            stsMap["role_arn"].(string),
		SupportRoleArn:     stsMap["support_role_arn"].(string),
		InstanceIAMRoles:   ExpandInstanceIamRolesFromInterface(stsMap["instance_iam_roles"]),
		OperatorRolePrefix: stsMap["operator_role_prefix"].(string),
		OIDCEndpointURL:    common2.GetOptionalStringFromMapString(stsMap, "oidc_endpoint_url"),
		OIDCConfigID:       common2.GetOptionalStringFromMapString(stsMap, "oidc_config_id"),
	}
}

func FlatSts(object *cmv1.Cluster) []interface{} {
	sts, ok := object.AWS().GetSTS()
	if !ok || sts == nil {
		// TODO: does it correct??
		return nil
	}

	result := make(map[string]interface{})

	result["oidc_endpoint_url"] = strings.TrimPrefix(sts.OIDCEndpointURL(), "https://")
	result["role_arn"] = sts.RoleARN()
	result["support_role_arn"] = sts.SupportRoleARN()
	// TODO: check if the bug was fixed in the uhc-cluster-services
	result["operator_role_prefix"] = sts.OperatorRolePrefix()

	thumbprint, err := common2.GetThumbprint(sts.OIDCEndpointURL(), common2.DefaultHttpClient{})
	if err == nil {
		result["operator_role_prefix"] = thumbprint
	}
	oidcConfig, ok := sts.GetOidcConfig()
	if ok && oidcConfig != nil && oidcConfig.ID() != "" {
		result["oidc_config_id"] = oidcConfig.ID()
	}

	instanceIAMRoles := sts.InstanceIAMRoles()
	if instanceIAMRoles != nil {
		instanceRolesMap := make(map[string]interface{})
		instanceRolesMap["master_role_arn"] = instanceIAMRoles.MasterRoleARN()
		instanceRolesMap["worker_role_arn"] = instanceIAMRoles.WorkerRoleARN()

		result["instance_iam_roles"] = []interface{}{instanceRolesMap}
	}

	return []interface{}{result}
}
