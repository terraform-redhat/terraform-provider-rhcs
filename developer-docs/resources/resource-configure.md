# Resource configure

Router: [`resource-overview.md`](resource-overview.md). Offline validation (no client): [`resource-schema.md`](resource-schema.md). Diagnostics: [`errors.md`](../errors.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN deciding whether Configure is required → [When Configure is needed](#when-configure-is-needed)
- WHEN implementing Configure → [Resource Configure method](#resource-configure-method)
- DEFAULT: Open [When Configure is needed](#when-configure-is-needed) first.

## When Configure is needed

WHEN the resource calls OCM (or other APIs) during plan/apply:
- MUST: Implement `resource.ResourceWithConfigure` and a Configure method.
WHEN the resource needs no provider client:
- MUST NOT: Add empty Configure boilerplate without a reason.
- MUST NOT: Invent a second, incompatible `ResourceData` type for one resource without updating provider Configure and all consumers.

## Resource Configure method

WHEN implementing Configure:
- MUST: Return early if `req.ProviderData == nil` (config validation can run before provider Configure).
- MUST: Type-assert `ProviderData` to the expected type; on mismatch, add a clear diagnostic and return (actionable message for provider developers — see [`errors.md`](../errors.md)).
- MUST: Store only what CRUD needs (e.g. collection/client handles), following analogous packages.
- DEFAULT: For new types, follow HashiCorp Configure (https://developer.hashicorp.com/terraform/plugin/framework/resources/configure) if `ResourceData` stays whatever provider Configure already exports (`*sdk.Connection` today). WHEN editing an existing package, match that package.
- EXAMPLE (Configure shape only — nil check + `*sdk.Connection` assert): `provider/imagemirror`. Prefer a package closest to your feature over copying unrelated CRUD/schema from that package.
