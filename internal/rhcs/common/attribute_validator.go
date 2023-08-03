package common

import (
	"fmt"
)

type SchemaValidateDiagFunc func(interface{}) error

// https://github.com/hashicorp/terraform-plugin-sdk/issues/780.
func ValidAllDiag(validators ...SchemaValidateDiagFunc) SchemaValidateDiagFunc {
	return func(i any) error {
		for _, validator := range validators {
			err := validator(i)
			if err != nil {
				return err
			}

		}
		return nil
	}
}

func ListOfMapValidator(i interface{}) error {
	l, ok := i.([]interface{})
	if !ok {
		return fmt.Errorf("expected type to be list of map")
	}
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	if _, ok = l[0].(map[string]interface{}); !ok {
		return fmt.Errorf("expected type to be list of map")
	}

	return nil
}
