package attrvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ validator.List = ConflictsWithNotEmptyValidator{}
)

// ConflictsWithNotEmpty ensures that a list parameter is not null and not empty if incompatible with the other paths
// specified in input
// It differs from standard ConflictsWith which only checks if the parameter is not null
func ConflictsWithNotEmpty(expressions ...path.Expression) validator.List {
	return ConflictsWithNotEmptyValidator{
		PathExpressions: expressions,
	}
}

// ConflictsWithNotEmptyValidator is the underlying struct implementing ConflictsWithNotEmpty.
type ConflictsWithNotEmptyValidator struct {
	PathExpressions path.Expressions
}

type ConflictsWithNotEmptyValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type ConflictsWithNotEmptyValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (v ConflictsWithNotEmptyValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ConflictsWithNotEmptyValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that if an attribute is set, these are not set or if set empty: %q",
		v.PathExpressions)
}

func (v ConflictsWithNotEmptyValidator) ValidateList(ctx context.Context, req validator.ListRequest,
	res *validator.ListResponse) {
	// If attribute configuration is null, it cannot conflict with others
	if req.ConfigValue.IsNull() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(v.PathExpressions...)

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		res.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			res.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if !mpVal.IsNull() && len(req.ConfigValue.Elements()) > 0 {
				res.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("Attribute %q cannot be specified and be not empty when %q is specified", mp,
						req.Path),
				))
			}
		}
	}
}
