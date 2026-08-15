# Embedded web UI

- `embed.go` embeds the generated `dist` tree for production delivery; `dist` is build output and must not be edited by hand.
- `memory.go` snapshots every embedded asset into a read-only in-memory filesystem during startup; the runtime cache cannot grow after bootstrap.
- `memory_test.go` verifies startup copies remain stable and late assets are not visible.
