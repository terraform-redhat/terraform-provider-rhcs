package machinepool

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func TaintsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"key": {
			Description: "Taint key",
			Type:        schema.TypeString,
			Required:    true,
		},
		"value": {
			Description: "Taint value",
			Type:        schema.TypeString,
			Required:    true,
		},
		"schedule_type": {
			Description: "Taint schedule type",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

type Taint struct {
	Key          string `tfsdk:"key"`
	Value        string `tfsdk:"value"`
	ScheduleType string `tfsdk:"schedule_type"`
}

func ExpandTaintsFromResourceData(resourceData *schema.ResourceData) []Taint {
	list, ok := resourceData.GetOk("taints")
	if !ok {
		return nil
	}

	return ExpandTaintsFromInterface(list)
}

func ExpandTaintsFromInterface(i interface{}) []Taint {
	l := i.([]interface{})
	if len(l) == 0 {
		return nil
	}
	taints := []Taint{}
	for _, taint := range l {
		taintMap := taint.(map[string]interface{})
		taints = append(taints, Taint{
			Key:          taintMap["key"].(string),
			Value:        taintMap["value"].(string),
			ScheduleType: taintMap["schedule_type"].(string),
		})
	}
	return taints
}
func FlatTaints(object *cmv1.MachinePool) []interface{} {
	taints, ok := object.GetTaints()
	if !ok || len(taints) < 0 {
		return nil
	}
	listOfTaints := []interface{}{}

	for _, taint := range taints {
		if taint == nil {
			continue
		}
		taintMap := make(map[string]string)
		taintMap["key"] = taint.Key()
		taintMap["value"] = taint.Value()
		taintMap["schedule_type"] = taint.Effect()
		listOfTaints = append(listOfTaints, taintMap)
	}
	return listOfTaints
}
