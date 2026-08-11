# Frontend maintenance scripts

- `compact-contract.mjs` keeps the generated OpenAPI type artifact below the
  repository line ceiling without changing its TypeScript declarations.
- `check-structure.mjs` audits source line limits, folder documentation,
  locale parity, script/template copy, local icon-asset policy, and coverage
  of production-source icons by the Vite client bundle scan or explicit list.
