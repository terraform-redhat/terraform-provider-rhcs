package ci

import (
	. "github.com/onsi/ginkgo/v2"
)

// Features
var FeatureMachinepool = Label("feature-machinepool")

// day1/day1-post and day2
var Day1 = Label("day1")
var Day1Prepare = Label("day1-prepare")
var Day1Post = Label("day1-post")
var Day2 = Label("day2")

// destroy
var Destroy = Label("destroy")

// importance
var Critical = Label("Critical")
var High = Label("High")
var Medium = Label("Medium")
var Low = Label("Low")

// exclude
var Exclude = Label("Exclude")
