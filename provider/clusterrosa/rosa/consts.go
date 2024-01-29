package rosa

import (
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
)

const (
	poolMessage = "This attribute is specifically applies for the Worker %[1]s Pool and becomes irrelevant once the resource is created. Any modifications to the default Machine Pool should be made through the Terraform imported Machine Pool resource. For more details, refer to [Worker %[1]s Pool in ROSA Cluster](../guides/worker-machine-pool.md)"

	tagsPrefix            = "rosa_"
	tagsOpenShiftVersion  = tagsPrefix + "openshift_version"
	propertyRosaTfVersion = tagsPrefix + "tf_version"
	propertyRosaTfCommit  = tagsPrefix + "tf_commit"
)

type ClusterTopology string

const (
	Classic ClusterTopology = "classic"
	Hcp     ClusterTopology = "hcp"
)

func GeneratePoolMessage(ct ClusterTopology) string {
	switch ct {
	case Hcp:
		return fmt.Sprintf(poolMessage, "Node")
	default:
		return fmt.Sprintf(poolMessage, "Machine")
	}
}

var OCMProperties = map[string]string{
	propertyRosaTfVersion: build.Version,
	propertyRosaTfCommit:  build.Commit,
}
