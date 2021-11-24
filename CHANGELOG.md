## 0.1.4 (November 24, 2021)

FEATURES:

* Add `ocm_machine_pool` resource.

* Add `ocm_groups` data source.

* Add `ocm_group_membership` resource.

* Add `ocm_versions` data source.

ENHANCEMENTS:

* Add `api_url` and `console_url` read-only attributes.

* Add support for specifying the number and type of compute nodes with the
  `compute_nodes` and `compute_machine_type` attributes.

* Add support for selection the _OpenShift_ version with the `version`
  attribute.

* Add support for configuring cluster networks with attributes `machine_cidr`,
  `service_cidr`, `pod_cidr` and `host_prefix`.

* Add support for _CCS_ clusters with the `ccs_enabled` attribute.

BREAKING CHANGES:

* Renamed attribute `cluster_id` to `cluster`.
