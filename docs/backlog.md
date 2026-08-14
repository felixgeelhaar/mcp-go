
## 2026 Roadmap Alignment

Foundation work for aligning mcp-go with MCP core 2026 roadmap priorities including horizontal scaling, discovery, and enterprise readiness.

---

## Spec Revisions Alignment (2024-11-05 → 2026-07-28)

Bring mcp-go current across every MCP spec revision. Backbone: protocol version negotiation. Phases 0–4 shipped on **v1** (no `/v2` module path): version negotiation, Streamable HTTP, RFC 9728 advertise-only auth metadata, tasks/icons/sampling-tools, then the 2026-07-28 stateless rewrite (`server/discover`, MRTR, `subscriptions/listen`, routing headers, CacheableResult, extensions). Auth stays out-of-library by design. Preserve differentiators: MCP Apps, WebSocket, gRPC, middleware suite. Detailed plan in docs/revisions-roadmap.md.

---
