# Attribute omission and clear semantics

Router: [resource-overview.md](resource-overview.md). Schema flags, defaults, and modifiers: [resource-schema.md](resource-schema.md). Refresh: [resource-read.md](resource-read.md). API updates: [resource-update.md](resource-update.md). Typed nulls: [resource-model.md](resource-model.md). Contract changes: [../breaking-changes.md](../breaking-changes.md).

WHEN implementing or reviewing an attribute whose configuration can be omitted, removed, or explicitly cleared:
- MUST: Review schema, plan modifiers, `Read()`, and `Update()` together. Schema flags alone do not define an API clear/reset operation.
- MUST: Establish the intended contract: clear the remote value, preserve the current value, let the API choose a value, or stop managing it.
- IF omission means stop managing the field, THEN define how that ownership change fits the schema and refresh contract; skipping a PATCH alone does not establish it. Optional+Computed permits an API-selected value, but does not disable refresh or ignore an explicitly configured value.
- MUST NOT: Confuse removing an attribute from configuration with destroying the resource. Resource destruction belongs in [resource-delete.md](resource-delete.md).

WHEN applying the rules below:
- NOTE: These rules cover ordinary stateful resource attributes. Write-only attributes have separate Framework restrictions and are not persisted in plan or state.

## Governing principle: honor the API contract in Terraform

- MUST: Establish what omission, null, empty, false, and defaults mean for this API field, then represent that contract correctly through the Terraform schema, plan, API requests, and state.
- MUST: Distinguish a real remote change from a representation-only difference. Refresh the remote value, but preserve its established Terraform representation when the API result is semantically equivalent and the attribute contract allows it.
- MUST NOT: Treat every API null/nil/empty value as interchangeable, or suppress refresh to hide a genuine change. Normalization preserves equivalent meaning; it does not skip observing the API.
- IF API null means inherit/default and an empty collection means clear all, THEN these are different operations and MUST remain distinct.

## API representation normalization

WHEN the API/client returns a different representation of the same logical value, such as a null/nil/empty collection meaning no tags:
- MUST: Confirm that equivalence for this attribute. A Go nil slice is an API/client representation, not an instruction to store Terraform null.
- WHEN reviewing `Read()`:
  - IF prior state is a known empty collection and the API reports no entries, and the contract considers these equivalent, THEN preserve the empty collection rather than replacing it with null.
  - IF prior state is null and the API reports no entries, THEN preserve null when it matches the attribute contract.
  - IF no prior representation is available, such as during import, THEN choose the null/empty representation defined by the attribute contract.
  - IF the API reports a genuinely different value, THEN record that change. MUST NOT preserve stale entries merely to obtain an empty plan.
- WHEN the mapping is used by `Create()`/`Update()`, MUST preserve the known planned representation if the API result is equivalent; do not replace a planned empty collection with null just because the API returned null.
- MUST: Use the matching typed null or known empty collection deliberately, following [resource-model.md](resource-model.md) and the semantic-preservation guidance in [resource-read.md](resource-read.md).
- NOTE: Framework custom-type semantic equality hooks only run for non-null, known prior and new values. Do not rely on those hooks alone to handle a raw null-versus-empty mismatch; normalize the API-to-state mapping as needed.
- Test empty Terraform collections against API null/nil/empty results, the null-state case, import without an established representation, and real additions/removals. Equivalent representations must not create a loop; real changes must remain observable.

## Optional-only attributes

WHEN `Optional: true`, `Computed: false`, and `Required: false`:
- Configuration omission produces null; removing a previously configured value plans null.
- WHEN reviewing `Read()`:
  - IF it records a non-null Terraform value while configuration omits the attribute, THEN the next plan can propose null again. First check for a contract-equivalent representation difference using the normalization rules above; otherwise review whether the remote value should actually be cleared.
  - IF `Update()` leaves that remote value unchanged and subsequent refresh restores it, THEN the implementation can produce a recurring clear diff.
  - IF the implementation avoids recording the remote value, THEN that value is not represented in state. MUST NOT use this as a generic way to hide drift; follow the refresh requirements in [resource-read.md](resource-read.md).
- WHEN reviewing `Update()`:
  - IF omission means clear, THEN MUST handle a transition from non-null prior state to planned null and send the API-supported clear operation/payload.
  - IF a null plan causes no API change, THEN verify whether the API preserves the value. If state becomes null while the API retains it, the value is effectively unmanaged through this attribute.
  - IF this conflicts with the intended contract or the next refresh, THEN revise the schema/API/state design rather than suppressing refresh.

## Optional and computed attributes

WHEN `Optional: true`, `Computed: true`, and `Required: false`:
- IF configuration supplies a value, THEN it remains practitioner-controlled; known plan values must be respected.
- IF configuration is null, THEN Terraform Core uses prior state as a fallback. When the overall plan differs from state, the Framework may mark unconfigured computed attributes unknown.
- MUST NOT: Assume removing configuration alone triggers `Update()` or resets an API default.
- WHEN reviewing `Read()`, MUST refresh the actual API value according to [resource-read.md](resource-read.md).
- WHEN reviewing `Update()`:
  - IF the plan value is known, THEN ensure the remote operation and response state honor it.
  - IF the plan value is unknown, THEN do not serialize unknown as an API value. Verify which API operation implements the intended result.
  - IF omitting the field from the API request is intended to select/recompute a default, THEN verify that the API actually defines omission that way; a PATCH omission may simply preserve the existing value.
  - IF configuration removal must force a reset, THEN require an explicit reset contract or deliberate planning design. Do not rely on an unknown plan or an absent modifier alone.

## Computed-only and required-only attributes

WHEN `Computed: true`, `Optional: false`, and `Required: false`:
- Practitioner configuration is rejected.
- MUST: Refresh the output from the API; do not send a read-only output as an editable field in `Update()`.
- WHEN considering a state-preserving modifier, follow the stability checks below.

WHEN `Required: true`, `Optional: false`, and `Computed: false`:
- Missing/null configuration is an error; omission is not a clear mechanism.
- WHEN changing an existing attribute to Required, MUST follow [../breaking-changes.md](../breaking-changes.md).

WHEN validating schema flags:
- MUST NOT: Combine Required with Optional or Computed, or leave all three false.

## State-preserving plan modifiers

WHEN reviewing `UseStateForUnknown()`:
- IF the planned value is unknown, prior resource state exists, and configuration is not unknown, THEN it copies the prior attribute value into the plan, including null.
- IF the plan is already known, including null or an explicit empty value, THEN it does nothing. It also does nothing on creation without prior state.
- MUST: Use it only when carrying that value through the update is correct. If the API may change the value during the update, preserving it can create a known-plan/result mismatch.
- MUST NOT: Treat it as an API clear/reset operation, a handler for known null plans, or a mandatory modifier for every computed output.
- IF the modifier is absent and an unconfigured computed value becomes unknown during an update, THEN `(known after apply)` is not, by itself, evidence of a drift loop.

WHEN a nested Optional+Computed child can have null prior state while the API supplies a value:
- MUST: Follow the `UseNonNullStateForUnknown()` rule and version requirements in [resource-schema.md#plan-modifiers](resource-schema.md#plan-modifiers), rather than blindly copying null.
- MUST: Search existing shared modifiers before adding custom behavior, as directed by that section.

## Explicit empty values, defaults, and collections

WHEN configuration contains `""`, `false`, or an empty collection:
- MUST: Distinguish that known value from null; verify whether the API interprets it as clear, disabled, empty, or invalid.
- WHEN an Optional boolean maps to a feature object omitted by the API, review the false/null and existing `IgnoreFalse()` guidance in [resource-schema.md#required-optional-computed](resource-schema.md#required-optional-computed). MUST NOT assume a modifier can legally rewrite a known configured `false` to null: verify configured-value consistency rules, the attribute contract, and modifier behavior before reusing a normalization pattern. IF the plan is known `false` and the API reports the equivalent disabled state, THEN preserve `false` in the apply result rather than replacing it with null.

WHEN a provider `Default` is present:
- MUST: Enable `Computed` as required by Framework schema validation; an Optional-only attribute cannot have a schema default.
- IF configuration is null, THEN review the defaulted plan, not a presumed null plan. A schema default changes planning; provider code still performs the API operation.
- MUST: Follow [resource-schema.md#defaults](resource-schema.md#defaults) and the contract-change guidance.

WHEN maps/objects shrink or are omitted, follow [resource-update.md#clear-on-shrink](resource-update.md#clear-on-shrink). WHEN the API omits/clears collection values, follow the typed-null/empty-contract rules in [resource-model.md](resource-model.md).

## Review and test the full cycle

- MUST: Remember that resource `Read()` receives prior state, not current configuration. `Update()` receives prior state, configuration, and plan.
- MUST: Preserve known/null planned values exactly in the `Create()` and `Update()` response state; unknown planned values may be resolved. If the API cannot produce the known planned result, revisit the contract/planning design. See [Create](resource-create.md) and [Update](resource-update.md).
- MUST: Verify both the remote result and Terraform state for the intended clear/preserve/default behavior.
- Test omission on create; removal after configuration; explicit null versus empty/false; unrelated updates; nested nulls; and refresh after an external API change.
- After apply and refresh, verify that the next plan has no repeated change when the API already matches the intended configuration. Real external drift must still be visible.

Framework references: [attribute configurability](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string), [plan modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification), [Read](https://developer.hashicorp.com/terraform/plugin/framework/resources/read), [Update](https://developer.hashicorp.com/terraform/plugin/framework/resources/update), [Terraform resource instance change lifecycle](https://github.com/hashicorp/terraform/blob/main/docs/resource-instance-change-lifecycle.md).
