# Architecture

- MUST: Implement only capabilities available in **ROSA** (via OCM APIs used by this provider).
- MUST NOT: Add provider support for features that are not available in ROSA.
- MUST: Confirm whether a change applies to **Classic**, **HCP**, or both before implementing.
- MUST: Keep Classic and HCP implementations separated (`provider/.../classic/` vs `provider/.../hcp/` and matching `subsystem/` trees).
- MUST: Call out parity or intentional divergence in the PR when a change touches both Classic and HCP.
- MUST: Prefer existing package layouts and patterns in this repo over new architecture or naming.
DEFAULT: When unsure whether support is HCP, Classic, or both — stop and verify before implementation.

**See:**

- ROSA docs: https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/
- OCM API model: https://github.com/openshift-online/ocm-api-model
