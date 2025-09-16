package imagemirror_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/imagemirror"
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
