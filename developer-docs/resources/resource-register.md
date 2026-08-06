# Resource register

Router: [`resource-overview.md`](resource-overview.md). Type names: [`naming.md`](../naming.md). Layout: [`package-layout.md`](../packages/package-layout.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN defining the resource type (factory / Metadata) → [Define the resource type](#define-the-resource-type)
- WHEN wiring the type into the provider → [Register on the provider](#register-on-the-provider)
- DEFAULT: Open [Define the resource type](#define-the-resource-type) first, then [Register on the provider](#register-on-the-provider).

## Define the resource type

WHEN adding a new resource:
- MUST: Implement `resource.Resource` in the package (factory `New()` returning `resource.Resource`, matching analogous packages).
- MUST: Set `Metadata` `TypeName` to `req.ProviderTypeName + "_…"` consistent with [`naming.md`](../naming.md) (e.g. `rhcs_dns_domain`).
- MUST: Follow the nearest analogous package for constructor and interface assertions (`ResourceWithConfigure`, import, etc. only when that type needs them).
- DEFAULT: Prefer Framework patterns from the HashiCorp **terraform-provider-development** skill for new types; WHEN editing an existing package, match that package.

## Register on the provider

WHEN the resource should be usable in Terraform:
- MUST: Add the factory to `Provider.Resources()` in [`provider/provider.go`](../../provider/provider.go) (same pattern as neighboring entries).
- MUST NOT: Leave a resource implemented but unregistered.
- DEFAULT: Keep list style consistent with existing entries in that function.

NOTE: Registration alone is not enough — continue the overview step sequence (configure → schema → …). Subsystem coverage: [`testing.md`](../testing.md).
