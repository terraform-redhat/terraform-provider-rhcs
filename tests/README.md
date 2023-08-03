# RHCS Provider Function Verification Testing
This package is the automation package for Function Verification Testing on the terraform rhcs provider. 

## Structure of tests
```sh
tests
|____e2e                                          ---- e2e folder contains all of the e2e cases
|    |____init_test.go
|    |____machine_pool_test.go
|    |____...
|    |____cluster_creation_validation_test.go
|    |____cluster_creation.go                     ---- cluster_creation.go will handle cluster creation by profile
|____utils
|    |____exec                                    ---- exec will package the tf commands like init state apply destroy
|    |    |____tf.go
|    |    |____cluster.go
|    |    |____...
|    |    |____machine_pool.go
|    |    |____idp.go
|    |____helper                                  ---- helper will provide support functions during testing
|    |    |_____helper.go
|    |____openshift                               ---- openshift will support e2e check in cluster
|    |____constants                               ---- all of the constants will be defined in constants
|         |____constants.go
|       
|____prow_ci.sh
|
|____tf-manifests                                 ---- tf-manifests folder will contain all of the tf configuration files, separated by clusterservice
      |__aws                                      ---- Prepare user data prepared on AWS for cluster creation 
      |     |__vpc                                ---- vpc created on AWS for cluster creation
      |     |__account-roles
      |     |__proxy
      |     |__…
      |__rhcs                                     ---- rhcs clusterservice, separated by resources
      |     |__clusters                           ---- clusters creation separated by resource key
      |     |   	|___rosa-classic
      |     |     |___osd-ccs
      |     |     |___rosa-hcp
      |     |     |___aro
      |     |___idps                              ---- idps creation
      |     |___machine-pools                     ---- machinepools creation
      |__azure                                    ---- azure clusterservice
```

## Contibute to terraform rhcs provider tests
Please read the structure and contribte code to the correct place

## Running
The cluster created by the automation scripts are shared across all of the test cases. [Why we need do this?](./docs/challenge.md).

### Prerequisite
Please read repo's [README.md](../README.md)
#### Users and Tokens
To execute the test cases, we need to prepare the [offline token](https://console.redhat.com/openshift/token/show). Get the offline token and export as an ENV variable **rhcs_TF_TOKEN**.
* export rhcs_TF_TOKEN=<offline token>


#### Global variables
The default rhcs enviroment is `production`. To specify running on staging or local, can export the global vairable 
* export GATE_WAY="https://api.stage.openshift.com"


### To run with ginkgo directly
This only allow running the test cases one by one. option.
* Running all the CMS test cases
  * `ginkgo -v ./tests/e2e`
* Running a single test case with scenario
  * `ginkgo -v -focus <case title> ./tests/e2e`
* Running a set of test case with label filter
  * `ginkgo run --label-filter <filter> ./tests/e2e`
