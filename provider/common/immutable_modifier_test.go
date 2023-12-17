package common

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
)

var _ = Describe("Immutable function tests- attributeChanged", func() {
	ctx := context.Background()
	It("attribute wasn't changed - should return false", func() {
		attributeConfig := AttributeConfig{
			attributeName:     "test",
			stateRawIsNull:    false,
			planRawIsNull:     false,
			stateValue:        types.StringValue("123"),
			planValue:         types.StringValue("123"),
			isAttrComputed:    false,
			configValueString: "123",
		}

		wasChanged := attributeChanged(ctx, attributeConfig)
		Expect(wasChanged).ToNot(BeTrue())
	})

	It("attribute value was changed - should return true", func() {
		attributeConfig := AttributeConfig{
			attributeName:     "test",
			stateRawIsNull:    false,
			planRawIsNull:     false,
			stateValue:        types.StringValue("123"),
			planValue:         types.StringValue("456"),
			isAttrComputed:    false,
			configValueString: "123",
		}

		wasChanged := attributeChanged(ctx, attributeConfig)
		Expect(wasChanged).To(BeTrue())
	})

	It("attribute was null and has a value not computed - should return true", func() {
		attributeConfig := AttributeConfig{
			attributeName:     "test",
			stateRawIsNull:    false,
			planRawIsNull:     false,
			stateValue:        types.StringNull(),
			planValue:         types.StringValue("456"),
			isAttrComputed:    false,
			configValueString: "123",
		}

		wasChanged := attributeChanged(ctx, attributeConfig)
		Expect(wasChanged).To(BeTrue())
	})

	It("attribute become to null not computed - should return true", func() {
		attributeConfig := AttributeConfig{
			attributeName:     "test",
			stateRawIsNull:    false,
			planRawIsNull:     false,
			stateValue:        types.StringValue("456"),
			planValue:         types.StringNull(),
			isAttrComputed:    false,
			configValueString: "123",
		}

		wasChanged := attributeChanged(ctx, attributeConfig)
		Expect(wasChanged).To(BeTrue())
	})

	It("attribute computed and unknown - should return false", func() {
		attributeConfig := AttributeConfig{
			attributeName:     "test",
			stateRawIsNull:    false,
			planRawIsNull:     false,
			stateValue:        types.StringValue("456"),
			planValue:         types.StringUnknown(),
			isAttrComputed:    true,
			configValueString: "123",
		}

		wasChanged := attributeChanged(ctx, attributeConfig)
		Expect(wasChanged).ToNot(BeTrue())

		attributeConfig = AttributeConfig{
			attributeName:     "test",
			stateRawIsNull:    false,
			planRawIsNull:     false,
			stateValue:        types.StringNull(),
			planValue:         types.StringUnknown(),
			isAttrComputed:    true,
			configValueString: "123",
		}

		wasChanged = attributeChanged(ctx, attributeConfig)
		Expect(wasChanged).ToNot(BeTrue())

	})
})
