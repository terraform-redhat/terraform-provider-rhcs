package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

func ValidatePatchProperties(ctx context.Context, state, plan types.Map) (map[string]string, error) {
	propertiesElements, err := common.OptionalMap(ctx, plan)
	if err != nil {
		return nil, err
	}
	if creatorArnValue, ok := propertiesElements[PropertyRosaCreatorArn]; ok {
		ogProperties, err := common.OptionalMap(ctx, state)
		if err != nil {
			return propertiesElements, err
		}
		if ogCreatorArn, ogOk := ogProperties[PropertyRosaCreatorArn]; ogOk {
			if creatorArnValue != ogCreatorArn {
				return propertiesElements, fmt.Errorf("Shouldn't patch property '%s'", PropertyRosaCreatorArn)
			}
		}
	}
	return propertiesElements, nil
}
