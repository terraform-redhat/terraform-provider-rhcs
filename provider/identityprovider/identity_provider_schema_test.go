/*
Copyright (c) 2026 Red Hat

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package identityprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getSchema() schema.Schema {
	r := New()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)
	return resp.Schema
}

var _ = Describe("Identity provider schema RequiresReplace", func() {

	var s schema.Schema

	BeforeEach(func() {
		s = getSchema()
	})

	Context("String attributes with RequiresReplace", func() {
		DescribeTable("should have RequiresReplace plan modifier",
			func(attrName string) {
				attr, ok := s.Attributes[attrName]
				Expect(ok).To(BeTrue(), "attribute %q must exist", attrName)

				strAttr, ok := attr.(schema.StringAttribute)
				Expect(ok).To(BeTrue(), "attribute %q must be a StringAttribute", attrName)
				Expect(strAttr.PlanModifiers).ToNot(BeEmpty(),
					"attribute %q must have plan modifiers", attrName)

				hasRequiresReplace := false
				for _, pm := range strAttr.PlanModifiers {
					if pm.Description(context.Background()) == "If the value of this attribute changes, Terraform will destroy and recreate the resource." {
						hasRequiresReplace = true
						break
					}
				}
				Expect(hasRequiresReplace).To(BeTrue(),
					"attribute %q must have RequiresReplace plan modifier", attrName)
			},
			Entry("cluster", "cluster"),
			Entry("name", "name"),
			Entry("mapping_method", "mapping_method"),
		)
	})

	Context("Object attributes with RequiresReplace", func() {
		DescribeTable("should have RequiresReplace plan modifier",
			func(attrName string) {
				attr, ok := s.Attributes[attrName]
				Expect(ok).To(BeTrue(), "attribute %q must exist", attrName)

				objAttr, ok := attr.(schema.SingleNestedAttribute)
				Expect(ok).To(BeTrue(), "attribute %q must be a SingleNestedAttribute", attrName)
				Expect(objAttr.PlanModifiers).ToNot(BeEmpty(),
					"attribute %q must have plan modifiers", attrName)

				hasRequiresReplace := false
				for _, pm := range objAttr.PlanModifiers {
					if pm.Description(context.Background()) == "If the value of this attribute changes, Terraform will destroy and recreate the resource." {
						hasRequiresReplace = true
						break
					}
				}
				Expect(hasRequiresReplace).To(BeTrue(),
					"attribute %q must have RequiresReplace plan modifier", attrName)
			},
			Entry("gitlab", "gitlab"),
			Entry("github", "github"),
			Entry("google", "google"),
			Entry("ldap", "ldap"),
			Entry("openid", "openid"),
		)
	})

	Context("HTPasswd attribute", func() {
		It("should NOT have RequiresReplace plan modifier", func() {
			attr, ok := s.Attributes["htpasswd"]
			Expect(ok).To(BeTrue(), "htpasswd attribute must exist")

			objAttr, ok := attr.(schema.SingleNestedAttribute)
			Expect(ok).To(BeTrue(), "htpasswd must be a SingleNestedAttribute")

			for _, pm := range objAttr.PlanModifiers {
				desc := pm.Description(context.Background())
				Expect(desc).ToNot(Equal(
					"If the value of this attribute changes, Terraform will destroy and recreate the resource."),
					"htpasswd must NOT have RequiresReplace plan modifier",
				)
			}
		})
	})
})
