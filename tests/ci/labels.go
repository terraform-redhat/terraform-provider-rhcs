package ci

import (
	. "github.com/onsi/ginkgo/v2"
)

// Features
var FeatureClusterautoscaler = Label("feature-clusterautoscaler")
var FeatureMachinepool = Label("feature-machinepool")
var FeatureIDP = Label("feature-idp")
var FeatureImport = Label("feature-import")

// day1/day1-post and day2
var Day1 = Label("day1")
var Day1Prepare = Label("day1-prepare")
var Day1Negative = Label("day1-negative")
var Day1Post = Label("day1-post")
var Day2 = Label("day2")
var Upgrade = Label("upgrade")

// day3 : the test cases will destroy default resource
var Day3 = Label("day3")

// destroy
var Destroy = Label("destroy")

// importance
var Critical = Label("Critical")
var High = Label("High")
var Medium = Label("Medium")
var Low = Label("Low")

// exclude
var Exclude = Label("Exclude")

// Cluster Type
var NonClassicCluster = Label("NonClassicCluster")
var NonHCPCluster = Label("NonHCPCluster")
