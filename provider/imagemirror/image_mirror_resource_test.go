package imagemirror_test

import (
	"testing"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/imagemirror"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestImageMirror(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ImageMirror Suite")
}

var _ = Describe("ImageMirrorResource", func() {
	It("implements ResourceWithConfigure", func() {
		var _ resource.ResourceWithConfigure = &imagemirror.ImageMirrorResource{}
	})

	It("implements ResourceWithImportState", func() {
		var _ resource.ResourceWithImportState = &imagemirror.ImageMirrorResource{}
	})

	It("creates a new resource instance", func() {
		resource := imagemirror.New()
		Expect(resource).ToNot(BeNil())
	})
})

var _ = Describe("DNS 1123 Subdomain Validation", func() {
	Context("Valid DNS names", func() {
		It("accepts lowercase alphanumeric strings", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("abc123")).To(BeTrue())
		})

		It("accepts strings with hyphens in the middle", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("abc-123")).To(BeTrue())
		})

		It("accepts single character names", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("a")).To(BeTrue())
		})

		It("accepts numbers at start and end", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("1abc2")).To(BeTrue())
		})
	})

	Context("Invalid DNS names", func() {
		It("rejects empty strings", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("")).To(BeFalse())
		})

		It("rejects strings starting with hyphen", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("-abc")).To(BeFalse())
		})

		It("rejects strings ending with hyphen", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("abc-")).To(BeFalse())
		})

		It("rejects uppercase characters", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("ABC")).To(BeFalse())
		})

		It("rejects special characters", func() {
			Expect(imagemirror.IsValidDNS1123Subdomain("abc.123")).To(BeFalse())
			Expect(imagemirror.IsValidDNS1123Subdomain("abc_123")).To(BeFalse())
			Expect(imagemirror.IsValidDNS1123Subdomain("abc@123")).To(BeFalse())
		})

		It("rejects strings longer than 253 characters", func() {
			longString := ""
			for i := 0; i < 254; i++ {
				longString += "a"
			}
			Expect(imagemirror.IsValidDNS1123Subdomain(longString)).To(BeFalse())
		})
	})
})

var _ = Describe("Registry Format Validation", func() {
	Context("Valid registry formats", func() {
		It("accepts simple hostnames", func() {
			Expect(imagemirror.IsValidRegistryFormat("registry")).To(BeTrue())
		})

		It("accepts hostnames with domains", func() {
			Expect(imagemirror.IsValidRegistryFormat("registry.example.com")).To(BeTrue())
		})

		It("accepts registries with paths", func() {
			Expect(imagemirror.IsValidRegistryFormat("registry.example.com/path")).To(BeTrue())
		})

		It("accepts registries with multiple path segments", func() {
			Expect(imagemirror.IsValidRegistryFormat("registry.example.com/org/repo")).To(BeTrue())
		})
	})

	Context("Invalid registry formats", func() {
		It("rejects empty strings", func() {
			Expect(imagemirror.IsValidRegistryFormat("")).To(BeFalse())
		})

		It("rejects registries without hostname", func() {
			Expect(imagemirror.IsValidRegistryFormat("/path/only")).To(BeFalse())
		})
	})
})

var _ = Describe("Zero-Egress Protected Registry Validation", func() {
	Context("Protected registries", func() {
		It("identifies quay.io OpenShift release dev registries", func() {
			Expect(imagemirror.IsZeroEgressProtectedRegistry("quay.io/openshift-release-dev/ocp-v4.0-art-dev")).To(BeTrue())
			Expect(imagemirror.IsZeroEgressProtectedRegistry("quay.io/openshift-release-dev/ocp-release")).To(BeTrue())
		})

		It("identifies quay.io app-sre registry", func() {
			Expect(imagemirror.IsZeroEgressProtectedRegistry("quay.io/app-sre")).To(BeTrue())
		})

		It("identifies Red Hat registries", func() {
			Expect(imagemirror.IsZeroEgressProtectedRegistry("registry.redhat.io")).To(BeTrue())
			Expect(imagemirror.IsZeroEgressProtectedRegistry("registry.access.redhat.com")).To(BeTrue())
		})

		It("identifies registries with subpaths", func() {
			Expect(imagemirror.IsZeroEgressProtectedRegistry("quay.io/openshift-release-dev/ocp-v4.0-art-dev/some/path")).To(BeTrue())
			Expect(imagemirror.IsZeroEgressProtectedRegistry("registry.redhat.io/openshift4/ose-node")).To(BeTrue())
		})
	})

	Context("Non-protected registries", func() {
		It("allows other quay.io registries", func() {
			Expect(imagemirror.IsZeroEgressProtectedRegistry("quay.io/myorg/myrepo")).To(BeFalse())
		})

		It("allows other registries", func() {
			Expect(imagemirror.IsZeroEgressProtectedRegistry("docker.io/library/nginx")).To(BeFalse())
			Expect(imagemirror.IsZeroEgressProtectedRegistry("gcr.io/my-project/image")).To(BeFalse())
			Expect(imagemirror.IsZeroEgressProtectedRegistry("registry.example.com/team/app")).To(BeFalse())
		})

		It("allows empty strings", func() {
			Expect(imagemirror.IsZeroEgressProtectedRegistry("")).To(BeFalse())
		})
	})
})