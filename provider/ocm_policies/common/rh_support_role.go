package common

import (
	"context"
	"encoding/json"
	"fmt"
)

func ParseRhSupportRole(ctx context.Context, awsPolicyDetails string) (string, error) {
	var trustPolicy map[string]any
	if err := json.Unmarshal([]byte(awsPolicyDetails), &trustPolicy); err != nil {
		return "", fmt.Errorf("failed to parse jit support role: %v", err)
	}
	statementList := trustPolicy["Statement"].([]any)
	if len(statementList) != 1 {
		return "", fmt.Errorf("failed to parse jit support role: expected a single statement")
	}
	principalMap := statementList[0].(map[string]any)["Principal"].(map[string]any)
	jitRoles := principalMap["AWS"].([]any)
	if len(jitRoles) != 1 {
		return "", fmt.Errorf("failed to parse jit support role: expected a single role")
	}
	return jitRoles[0].(string), nil
}
