package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var ctx context.Context

func TestRHCSProvider(t *testing.T) {
	ctx = context.Background()
	RegisterFailHandler(Fail)
	RunSpecs(t, "RHCS Provider Test")

}
