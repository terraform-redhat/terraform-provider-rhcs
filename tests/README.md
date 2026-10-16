# RHCS Provider Function Verification Testing

This package is the automation package for Function Verification Testing on the terraform rhcs provider.

BEFORE CREATING COMMITS FROM THIS CLONE, YOU MUST RUN `make install-hooks` ONCE. YOU MUST LET LOCAL HOOKS RUN ON EVERY COMMIT AND PUSH.

## Structure of tests

```bash
tests
|____e2e  
    |____cluster_creation_test.go                 ---- handles cluster creation by profile
    |____cluster_destroy_test.go
    |____e2e_suite_test.go                        ---- test suite for all the e2e tests
    |____idps_test.go
    |____...
    |____machine_pool_test.go
    |____verfication_post_day1_test.go            ---- post day1 verification tests
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
* If any other data can be created by terrafom provder, define steps under terraform-provider-rhcs/tests/tf-manifests/\<provider\>/\<resource name\>
* Define prepare steps according to the profile in terraform-provider-rhcs/tests/ci/profile_handler.go

#### Day1 cases only do creation step, we have to define a day1 case to verify the day1 works well

* Create the case in terraform-provider-rhcs/tests/e2e/\<feature name\>_test.go
* Label the case with `CI.Day1Post`
* Label the case with importance `CI.Critical`
* Don't need to run creation step, just in BeforeEach step call function  ***PrepareRHCSClusterByProfileENV()*** it will load the clusterID prepared
* Code for checking steps  only in the case

### Contribute to day2

* Create the case in terraform-provider-rhcs/tests/e2e/\<feature name\>_test.go
* Label the case with `CI.Day2`
* Label the case with importance `CI.Critical` or `CI.High`
* Don't need to run creation step, just in BeforeEach step call function  ***PrepareRHCSClusterByProfileENV()*** it will load the clusterID prepared
* Code for day2 actions and check step
* Every case need to recover the cluster after the case run finished unless it's un-recoverable

### Contribute to day3

* day3 cases means the cases is destructive and cannot recover anymore. The actions will affect other cases like default worker pool deletion, it will affect many day1 post verification
* Create the case in terraform-provider-rhcs/tests/e2e/\<feature name\>_test.go
* Label the case with `CI.Day3`
* Label the case with importance `CI.Critical` or `CI.High`
* Don't need to run creation step, just in BeforeEach step call function  ***PrepareRHCSClusterByProfileENV()*** it will load the clusterID prepared
* Code for day3 actions and check step
* Every case need to recover the cluster after the case run finished unless it's un-recoverable

### Labels

* Label your case with the `CI.Feature<feature name>` defined in terraform-provider-rhcs/tests/ci/labels.go
* Label your case with importance defined in terraform-provider-rhcs/tests/ci/labels.go
* Label your case with `CI.Day1Post/CI.Day2/CI.Day3` according to the case runtime
* Label your case with `CI.Exclude` if it fails CI all  the time and you can't fix it in time

## Running

The cluster created by the automation scripts are shared across all of the test cases. [Why we need do this?](./docs/challenge.md).

### Prerequisite

Please read repo's [README.md](../README.md)
For the test cases, we need `make install` to make the terraform provider installed to local

#### Global variables

##### RHCS_TOKEN

To execute the test cases, we need to prepare the [offline token](https://console.redhat.com/openshift/token/show). Get the offline token and export as an ENV variable:

* export RHCS_TOKEN=\<offline token\>

##### RHCS_URL

The default rhcs enviroment is `staging`. To specify running on production, integration or local, you can export the global variable:

* export RHCS_URL = "https://api.openshift.com"

##### CLUSTER_PROFILE

To declare the profile cluster type to be run, use the below variable::

* export CLUSTER_PROFILE = \<rosa-tf-profile\>

### Running a local CI simulation

This feature allows for running tests through a case filter to simulate CI. Anyone can customize the case label filter to select the specific cases that would be run.

* Prepare cluster with profile
  * Check file terraform-provider-rhcs/tests/ci/profiles/tf_cluster_profile.yml to find the supported profiles and detailed configuration
    * `export CLUSTER_PROFILE=<profile name>`
  * Export the token
    * `export RHCS_TOKEN=<rhcs token>`
  * Set up the OCM URL (if different from staging)
    * `export RHCS_URL="https://api.openshift.com"`
  * Run ginkgo run command
    * `ginkgo run --label-filter day1-prepare --timeout 2h tests/e2e`
* Run day2 day1 post cases with profile
  * Check file terraform-provider-rhcs/tests/ci/profiles/tf_cluster_profile.yml to find the supported profiles and detailed configuration
    * `export CLUSTER_PROFILE=<profile name>` #If it had been run, then you can skip
  * Export the token
    * `export RHCS_TOKEN=<rhcs token>` #If it had been run, then you can skip
  * Set up the OCM URL (if different from staging)
    * `export RHCS_URL="https://api.openshift.com"`
  * Run ginkgo run command
    * `ginkgo run --label-filter '(Critical,High)&&(day1-post,day2)&&!Exclude' tests/e2e`
* Run a specified case to debug
  * `ginkgo -focus <case id> tests/e2e`
* Run destroy with profile
  * Check file terraform-provider-rhcs/tests/ci/profiles/tf_cluster_profile.yml to find the supported profiles and detailed configuration
    * `export CLUSTER_PROFILE=<profile name>` #If it had been run, then you can skip
  * Export the token
    * `export RHCS_TOKEN=<rhcs token>` #If it had been run, then you can skip
  * Set up the OCM URL (if different from staging)
    * `export RHCS_URL="https://api.openshift.com"`
   * Run ginkgo run command
     * `ginkgo run --label-filter destroy tests/e2e`

### Shared HCP/HyperFleet resource E2E workflow

Shared resource E2E tests use `CLUSTER_PROFILE` to select the HCP or HyperFleet
backend. The shared NodePool IMDSv2 test is one example: it configures the
metadata-token requirement on each NodePool and verifies both `optional` and
`required` values.

For HyperFleet, use the same API URL for cluster setup, the test, and teardown.
Set the release image for cluster creation, and make AWS credentials available
to Terraform. `RHCS_TOKEN` is only needed for HCP/OCM workflows. The
`hyperfleet-basic` profile is defined in
`tests/ci/profiles/tf_hyperfleet_cluster_profiles.yml`.
The AWS region can come from the profile or be inferred from the region in the
execute-api URL; set `REGION` explicitly only if neither supplies it.
`hyperfleet-basic` is no-proxy; use `hyperfleet-explicit-proxy` for the explicit
HTTP(S) proxy profile. `hyperfleet-hcp-ad` remains proxy-enabled. Destroy a
cluster created with an older proxy-enabled `hyperfleet-basic` definition before
using the updated no-proxy profile, so its proxy Terraform state is cleaned up.
`hyperfleet-basic` is no-proxy; use `hyperfleet-explicit-proxy` to exercise the
explicit HTTP(S) proxy configuration (not transparent proxying). The
`hyperfleet-hcp-ad` profile remains proxy-enabled. Tear down a cluster created
with a proxy-enabled profile before changing that profile's proxy setting, so
the matching proxy Terraform state is cleaned up.

Create the cluster from the selected profile:

```bash
TF_EXEC_STREAM_OUTPUT=true CLUSTER_PROFILE=hyperfleet-basic HYPERFLEET_URL=https://<api-id>.execute-api.<api-region>.amazonaws.com/prod HYPERFLEET_RELEASE_IMAGE=<release-image> make e2e_test LabelFilter=day1-prepare
```

Run only the shared IMDSv2 test:

```bash
TF_EXEC_STREAM_OUTPUT=true CLUSTER_PROFILE=hyperfleet-basic HYPERFLEET_URL=https://<api-id>.execute-api.<api-region>.amazonaws.com/prod go run github.com/onsi/ginkgo/v2/ginkgo run --focus='sets IMDSv2 per nodepool through the selected backend' --focus-file='tests/e2e/hcp_machine_pool_test.go' --label-filter='day2 && feature-machinepool && hyperfleet-validated' --timeout=5h tests/e2e
```

The test deletes its NodePools after verification. Tear down the profile-managed
cluster resources afterward. Set `WAIT_FOR_NODEPOOL_DELETION=true` in the
environment if you want the test to wait for each asynchronous NodePool deletion;
it defaults to false.

```bash
TF_EXEC_STREAM_OUTPUT=true CLUSTER_PROFILE=hyperfleet-basic HYPERFLEET_URL=https://<api-id>.execute-api.<api-region>.amazonaws.com/prod make e2e_test LabelFilter=destroy
```

For an already provisioned cluster, skip the setup command. The suite reads the
cluster ID from profile Terraform state; provide `CLUSTER_ID=<cluster-id>` on
the test and teardown commands if the cluster is not recorded in that workspace.

### Set log level

* Log level defined in terraform-provider-rhcs/tests/utils/utils/log/logger.go

```golang
logger.SetLevel(logging.InfoLevel)
```
