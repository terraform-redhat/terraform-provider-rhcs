package ci

import (
	. "github.com/onsi/ginkgo/v2"
)

// Features
var FeatureClusterAutoscaler = Label("feature-cluster-autoscaler")
var FeatureClusterCompute = Label("feature-cluster-compute")
var FeatureClusterDefault = Label("feature-cluster-default")
var FeatureClusterEncryption = Label("feature-cluster-encryption")
var FeatureClusterIMDSv2 = Label("feature-cluster-imdsv2")
var FeatureClusterMisc = Label("feature-cluster-misc")
var FeatureClusterNetwork = Label("feature-cluster-network")
var FeatureClusterPrivate = Label("feature-cluster-private")
var FeatureClusterProxy = Label("feature-cluster-proxy")
var FeatureClusterRegistryConfig = Label("feature-cluster-registry-config")
var FeatureIngress = Label("feature-ingress")
var FeatureMachinepool = Label("feature-machinepool")
var FeatureIDP = Label("feature-idp")
var FeatureImport = Label("feature-import")
var FeatureTuningConfig = Label("feature-tuning-config")
var FeatureExternalAuth = Label("feature-external-auth")
var FeatureImageMirror = Label("feature-image-mirror")

// day1/day1-post and day2
var Day1 = Label("day1")
var Day1Prepare = Label("day1-prepare")
var Day1Negative = Label("day1-negative")
var Day1Supplemental = Label("day1-supplemental")
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
