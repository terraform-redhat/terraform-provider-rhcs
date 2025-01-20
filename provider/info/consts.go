/*
Copyright (c) 2021 Red Hat, Inc.

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
package info

type ocmEnv string

const (
	ocmEnvProd         ocmEnv = "production"
	ocmEnvStage        ocmEnv = "stage"
	ocmEnvInt          ocmEnv = "integration"
	ocmEnvFedRAMPProd  ocmEnv = "fedramp-production"
	ocmEnvFedRAMPStage ocmEnv = "fedramp-stage"
	ocmEnvFedRAMPInt   ocmEnv = "fedramp-int"
)

var ocmAWSAccounts = map[ocmEnv]string{
	ocmEnvProd:         "710019948333",
	ocmEnvStage:        "644306948063",
	ocmEnvInt:          "896164604406",
	ocmEnvFedRAMPProd:  "448648337690",
	ocmEnvFedRAMPStage: "448870092490",
	ocmEnvFedRAMPInt:   "449053620653",
}
