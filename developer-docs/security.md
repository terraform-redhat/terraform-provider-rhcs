# Security

WHEN editing provider code, tests, docs, examples, or logs:
- MUST NOT: Hard code secrets, API keys, tokens, kubeconfigs, AWS credentials, or customer identifiers.
- MUST NOT: Log or print credentials, tokens, or other secrets.
- MUST: Mark secret schema attributes **Sensitive**.
DEFAULT: Prefer placeholders and variables in examples and test fixtures.

## Trivy (IaC misconfiguration)

Repo config: root **`trivy.yaml`**. CodeRabbit may run Trivy when enabled in **`.coderabbit.yaml`**.

WHEN **`trivy config`** reports a **misconfiguration** (IDs like **`AWS-0104`**, **`DS-0002`** — not CVE rows from **`trivy fs`**):
1. MUST: Prefer fixing the HCL/Dockerfile (least privilege, encryption, IMDSv2, non-root, etc.).
2. MAY: `#trivy:ignore:<id>` on the line immediately above the resource or Dockerfile instruction, with a short `#` justification (narrow scope).
3. MAY: `.trivyignore` only when inline suppression is impossible — one ID per line with a `#` justification above each.
DEFAULT: Do not suppress findings without a documented reason.
