# Member connections

- `types.go` defines the transient connection projection and capability claims; raw provider IP results are never stored.
- `handles.go` signs owner, scan, node, IP, and expiry into a 15-minute HMAC capability accepted by drop operations.
- `service.go` starts and polls metadata-only scans, then attaches ephemeral handles.
- `worker.go` starts provider scans and moves interrupted or ambiguous starts to durable `pending_review` without retrying the provider call.
- `drop_service.go` stores only a target hash and seals the capability into the outbox payload.
- `drop_worker.go` drops once, then reconciles uncertain outcomes through a fresh scan.
- `drop_reconcile.go` bounds fresh-scan polling and checks the selected node/IP.
- `handles_test.go` covers owner binding, tamper rejection, and strict expiry without making provider calls.
- `worker_test.go` covers ambiguous start responses, interrupted starts, terminal polling projection, and replay suppression.

Provider scans and drops are invoked by durable operation workers through the shared Remnawave queue. Browser payloads never choose a node or IP directly.
