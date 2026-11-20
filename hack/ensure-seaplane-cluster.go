// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

func main() {
	profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load profile: %v\n", err)
		os.Exit(1)
	}
	profileName := profileHandler.Profile().GetName()
	if profileName == "" {
		fmt.Fprintf(os.Stderr, "profile name is empty\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "checking cluster for profile %s\n", profileName)

	service, err := exec.NewSeaplaneClusterService(profileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create cluster service for %s: %v\n", profileName, err)
		os.Exit(1)
	}
	output, err := service.Output()
	if err == nil && output != nil && output.ClusterID != "" {
		fmt.Fprintf(os.Stderr, "cluster already exists for %s: %s (cluster_id=%s)\n", profileName, output.ClusterName, output.ClusterID)
		os.Exit(0)
	}
	if output != nil {
		fmt.Fprintf(os.Stderr, "no cluster in terraform state for %s (output: %+v), creating...\n", profileName, output)
	} else {
		fmt.Fprintf(os.Stderr, "no terraform output for %s, creating...\n", profileName)
	}

	seaplaneURL := os.Getenv("SEAPLANE_URL")
	releaseImage := config.GetSeaplaneReleaseImage()
	if seaplaneURL == "" {
		fmt.Fprintf(os.Stderr, "SEAPLANE_URL is required to create cluster for %s\n", profileName)
		os.Exit(1)
	}
	if releaseImage == "" {
		fmt.Fprintf(os.Stderr, "SEAPLANE_RELEASE_IMAGE is required to create cluster for %s\n", profileName)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "creating seaplane cluster for profile %s with release %s\n", profileName, releaseImage)
	backend, err := clusterworkflow.New(profileHandler, "", profileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create backend for %s: %v\n", profileName, err)
		os.Exit(1)
	}
	ctx := context.Background()
	clusterID, err := exec.CreateAndWaitReady(ctx, backend.Lifecycle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create cluster for %s: %v\n", profileName, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "cluster created for %s: %s\n", profileName, clusterID)
}
