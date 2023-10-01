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
|____tf-manifests                                 ---- tf-manifests folder will contain all of the tf configuration files, separated by provider
      |__aws                                      ---- Prepare user data prepared on AWS for cluster creation 
      |     |__vpc                                ---- vpc created on AWS for cluster creation
      |     |__account-roles
      |     |__proxy
      |     |__…
      |__rhcs                                     ---- rhcs provider, separated by resources
      |     |__clusters                           ---- clusters creation separated by resource key
      |     |   	|___rosa-classic
      |     |     |___osd-ccs
      |     |     |___rosa-hcp
      |     |     |___aro
      |     |___idps                              ---- idps creation
      |     |___machine-pools                     ---- machinepools creation
      |__azure                                    ---- azure provider
```

## Contibute to terraform rhcs provider tests
Please read the structure and contribte code to the correct place
### Contribute to day1 case
* Define manifests files under terraform-provider-rhcs/tests/tf-manifests/rhcs/clusters
* Define profile in terraform-provider-rhcs/tests/ci/profiles/tf_cluster_profile.yml
* If any other data can be created by terrafom provder, define steps under terraform-provider-rhcs/tests/tf-manifests/<provider>/<resource name>
* Define prepare steps according to the profile in terraform-provider-rhcs/tests/ci/profile_handler.go
#### Day1 cases only do creation step, we have to define a day1 case to verify the day1 works well
* Create the case in terraform-provider-rhcs/tests/e2e/<feature name>_test.go
* Label the case with ***CI.Day1Post***
* Label the case with importance ***CI.Critical***
* Don't need to run creation step, just in BeforeEach step call function  ***PrepareRHCSClusterByProfileENV()*** it will load the clusterID prepared
* Code for checking steps  only in the case
### Contribute to day2
* Create the case in terraform-provider-rhcs/tests/e2e/<feature name>_test.go
* Label the case with ***CI.Day2***
* Label the case with importance ***CI.Critical*** or ***CI.High***
* Don't need to run creation step, just in BeforeEach step call function  ***PrepareRHCSClusterByProfileENV()*** it will load the clusterID prepared
* Code for day2 actions and check step
* Every case need to recover the cluster after the case run finished unless it's un-recoverable
### Labels
* Label your case with the ***CI.Feature<feature name>*** defined in terraform-provider-rhcs/tests/ci/labels.go
* Label your case with importance defined in terraform-provider-rhcs/tests/ci/labels.go
* Label your case with ***CI.Day1/CI.Day2*** according to the case runtime
* Label your case with ***CI.Exclude*** if it fails CI all  the time and you can't fix it in time

## Running
The cluster created by the automation scripts are shared across all of the test cases. [Why we need do this?](./docs/challenge.md).

### Prerequisite
Please read repo's [README.md](../README.md)
#### Users and Tokens
To execute the test cases, we need to prepare the [offline token](https://console.redhat.com/openshift/token/show). Get the offline token and export as an ENV variable 

**rhcs_TF_TOKEN**.
* export RHCS_TOKEN=<offline token>

#### Global variables
The default rhcs enviroment is `production`. To specify running on staging or local, can export the global vairable 
* export RHCSEnv = "staging"

To declare the profile cluster type to be run, use the below variable::
* export CLUSTER_PROFILE = <rosa-tf-profile>

### To run with ginkgo directly
This only allow running the test cases one by one. option.
* Running all the CMS test cases
  * `ginkgo -v ./tests/e2e`
* Running a single test case with scenario
  * `ginkgo -v -focus <case title> ./tests/e2e`
* Running a set of test case with label filter
  * `ginkgo run --label-filter <filter> ./tests/e2e`

### To run on local simulate CI
This allows run by case filter to simulate CI. Anybody can customize the case label filter to filter out the cases want to run
* Prepare cluster with profile
  * Check file terraform-provider-rhcs/tests/ci/profiles/tf_cluster_profile.yml to find the supported profiles and detailed configuration
    * `export CLUSTER_PROFILE=<profile name>` 
  * Export the token
    * `export RHCS_TOKEN=<rhcs token>`
  * Export the label filter to locate the preparation case
    * `export LabelFilter="day1-prepare"`
  * Run make command
    * `make e2e_test`
* Run day2 day1 post cases with profile
  * Check file terraform-provider-rhcs/tests/ci/profiles/tf_cluster_profile.yml to find the supported profiles and detailed configuration
    * `export CLUSTER_PROFILE=<profile name>` #If it had been run, then you can skip
  * Export the token
    * `export RHCS_TOKEN=<rhcs token>` #If it had been run, then you can skip
  * Export the label filter to locate the preparation case
    * `export LabelFilter="(Critical,High)&&(day1-post,day2)&&!Exclude"` #This filter only choose all critical and high priority cases with day1-post and day2
  * Run make command
    * `make e2e_test`
* Run destroy with profile
  * Check file terraform-provider-rhcs/tests/ci/profiles/tf_cluster_profile.yml to find the supported profiles and detailed configuration
    * `export CLUSTER_PROFILE=<profile name>` #If it had been run, then you can skip
  * Export the token
    * `export RHCS_TOKEN=<rhcs token>` #If it had been run, then you can skip
  * Export the label filter to locate the preparation case
    * `export LabelFilter="destroy"` #This filter only choose all critical and high priority cases with day1-post and day2
  * Run make command
    * `make e2e_test`

### Set log level
* Log level defined in terraform-provider-rhcs/tests/utils/exec/tf-exec.go
```
logger.SetLevel(logging.InfoLevel)
```