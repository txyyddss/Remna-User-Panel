# Frontend maintenance scripts

- `compact-contract.mjs` keeps the generated OpenAPI type artifact below the
  repository line ceiling without changing its TypeScript declarations.
- `check-structure.mjs` audits source line limits, folder documentation,
  locale parity, script/template copy, and local icon-asset policy in CI.
