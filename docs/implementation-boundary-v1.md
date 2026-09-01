# Implementation boundary v1

This implementation is intentionally scoped to semantic proof-aware planning. It does not mutate source repositories, create commits, merge pull requests, create tags, or publish releases at runtime. Those operations are operator-owned workflow steps and are recorded separately from the runtime authority counters in CI evidence.

The only persistent inputs are `.gooo` semantics, the fixed denominator contract, fixture identities, and immutable receipt facts. All semantic IR, graph, plans, generated Go, execution evidence, receipts, and reports are caller-owned outputs.
