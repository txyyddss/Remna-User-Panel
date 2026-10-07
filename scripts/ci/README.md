# Hosted Go race checks

`race_shards.py` is called only by GitHub Actions. Four independent runners split
the database package's top-level tests, executable examples and fuzz seed tests,
and partition every other Go package. New packages and checks are discovered
from Go itself; no handwritten allowlist can silently omit them.

`race_inventory.py` owns discovery and exact-once partition validation. Each
runner uploads its complete inventory, assigned checks, completion marker and
JSON diagnostics. The existing Go race and coverage gate validates all four
manifests before running the unchanged domain coverage packages and 70% gate.

Per-runner Go parallelism follows available CPUs. Database tests retain their
existing two-open migration semaphore and isolated temporary databases. Shards
have independent processes and CPU allocations, so migration concurrency does
not multiply on one machine. Do not invoke these scripts locally: repository
instructions permit automated suites only in hosted CI.

Main pushes always run the backend gates. A frontend-only repair must not cancel
an unfinished backend run and publish without validating the upgraded Go tree.
Pull requests retain changed-area filtering.
The aggregate gate follows the actual shard result, including failures, rather
than reevaluating changed-area output after the matrix has completed.
