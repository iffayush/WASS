# rules/ — WASS secret rules + fixtures

This directory holds the YAML secret-detection rules and their fixtures, built in **M2**.

## What ships here (M2)

- `*.yaml` rule files: one rule per exposed-value shape (AWS, GitHub, Stripe, Google, Firebase,
  Supabase, Slack, Twilio, SendGrid, private keys, JWTs, generic high-entropy assignments).
- `fixtures/`: every rule ships with positive and negative fixtures; a rule is not done until its
  fixtures pass. Fixture tests are a CI merge gate.

## Rule format

```yaml
id: aws-access-key-id            # stable rule id, used in fingerprints
description: AWS access key ID
regex: '(?:^|[^A-Z0-9])(AKIA[0-9A-Z]{16})(?:[^A-Z0-9]|$)'
keywords: ["AKIA"]               # prefilter tokens (Aho-Corasick)
entropy: 0                       # optional Shannon-entropy threshold
provider: aws
default_classification: critical # public_by_design | dangerous_if_misconfigured | critical | unknown
remediation_template: |
  1. Deactivate the key in the AWS console (IAM → Users → Security credentials).
  2. Rotate: create a new key, update the workload, then delete the exposed one.
  3. If the key was ever committed, purge history (git filter-repo / BFG) and audit CloudTrail.
```

## Hard rule

**Fixtures use obviously fake, generated values only. Never commit real credentials** — CI runs
gitleaks on every PR and the pre-commit hook runs the same check. Use reserved documentation
values (`AKIAIOSFODNN7EXAMPLE`) or clearly synthetic strings.
