Product Requirements Document (PRD)
Product: mcp-go

Tagline: The Gin framework for Model Context Protocol servers in Go

1. Product Vision
   Vision Statement

Enable Go developers to expose MCP-compliant tools, resources, and prompts with the same simplicity, safety, and ergonomics that Gin brought to HTTP APIs.

If MCP is the control plane for AI tools, mcp-go is the standard Go runtime for it.

2. Problem Statement
   The Problem

Model Context Protocol (MCP) adoption is accelerating across:

AI agents

IDE integrations

Internal automation platforms

AIOps and DevOps tooling

However, Go developers are blocked by:

Low-level MCP bindings

Missing abstractions

No typed handlers

No middleware

No schema ergonomics

No production defaults

Today, building an MCP server in Go feels like writing net/http by hand — not acceptable for modern infra teams.

3. Target Users & Personas
   Primary Persona: Platform / Infra Engineer

Uses Go for backend, CLIs, agents, internal platforms

Wants:

strong typing

correctness by default

predictable behavior

easy observability hooks

Examples:

AI platform teams

DevOps / SRE

Internal tooling teams

Secondary Persona: OSS / Agent Framework Author

Needs a reusable MCP server foundation

Wants:

spec compliance

extensibility

transport flexibility

clean public API

4. Jobs To Be Done (JTBD)

“When I want to expose tools or resources to AI agents using MCP, I want to do it quickly, safely, and idiomatically in Go, without worrying about protocol details.”

Core Jobs

Define MCP tools with typed inputs

Validate and document schemas automatically

Handle MCP transport(s) correctly

Apply auth, logging, limits, and observability consistently

Run in production with confidence

5. Goals & Success Metrics
   Product Goals

Developer Experience First

Hello World MCP server in < 10 minutes

Spec Compliance

Pass MCP compliance tests

Production Readiness

Safe defaults, graceful shutdown, concurrency control

Ecosystem Adoption

Become the default Go MCP recommendation

Success Metrics

⭐ GitHub stars (signal, not vanity)

📦 Downstream usage in other OSS projects

🧩 Adoption in agent frameworks / internal tools

📚 Docs & examples referenced externally

6. Non-Goals (Important)

❌ Building an agent framework

❌ Implementing LLM logic

❌ Opinionating on AI providers

❌ UI or dashboards

❌ Replacing MCP clients

This is infrastructure, not AI logic.

7. Core Value Proposition
   Without mcp-go With mcp-go
   Manual JSON parsing Typed handlers
   Hand-written schemas Auto schema generation
   Ad-hoc error handling MCP-native errors
   No middleware Gin-style middleware
   One-off transports Pluggable transports
   Fragile servers Production-safe defaults
8. Product Principles

Typed > Dynamic

Explicit > Magic

Safe Defaults > Flexibility

Opinionated Core, Extensible Edges

Boring is Good (for infra)

9. Functional Requirements
   9.1 MCP Server Core

Create an MCP server with metadata:

name

version

capabilities

Register:

tools

resources

prompts

Expose introspection / manifest

9.2 Tools

Define tools via builder API

Support:

description

tags

input schema

Typed handler signature:

func(ctx context.Context, input T) (any, error)

Automatic:

decoding

validation

error mapping

9.3 Resources

URI template matching

Path parameter extraction

Typed or param-based handlers

MCP-compliant resource responses

9.4 Prompts (Parity)

Prompt registration

Structured prompt outputs

Optional parameters

9.5 Middleware

Gin-style middleware chain:

Recovery

Request ID

Logging

Timeout

Concurrency limiting

Auth / Principal injection

Middleware must be:

composable

order-dependent

context-aware

9.6 Transports

Pluggable transport layer:

stdio (required for MCP)

HTTP + SSE (for services)

WebSocket (post-MVP)

Single server → multiple transports.

9.7 Schema & Validation

JSON Schema generation from Go structs

Tag-based constraints

Strict decoding by default

Ability to override schema manually

9.8 Errors

MCP-native error envelope

Error codes:

invalid_params

not_found

unauthorized

internal

Custom error mapping hook

9.9 Observability Hooks

Logging hooks

Metrics interface (Prometheus-compatible)

Tracing hooks (OpenTelemetry-friendly)

10. Non-Functional Requirements
    Performance

Minimal allocations

Concurrent-safe

No reflection at runtime per request (pre-compute schemas)

Reliability

Panic recovery

Graceful shutdown

Context propagation everywhere

Security

Strict JSON parsing

No implicit trust of clients

Explicit auth middleware

11. API Design (Public Surface)
    Example
    srv := mcp.NewServer(mcp.ServerInfo{
    Name: "obvia",
    Version: "0.1.0",
    })

srv.Use(
mcp.Recover(),
mcp.RequestID(),
mcp.Timeout(5\*time.Second),
)

srv.Tool("search").
Description("Search incidents").
Input(SearchInput{}).
Handler(searchHandler)

mcp.ListenAndServeHTTP(":8080", srv)

DX parity target: Gin / Cobra / sqlc-level quality

12. MVP Scope (v0.1)
    Included

Server core

Tools

Resources

Typed handlers

JSON Schema generation

stdio transport

HTTP+SSE transport

Middleware: recover, timeout, logging

Examples + docs

Excluded

WebSocket transport

Streaming tool responses

Client SDK

Code generation

13. Roadmap
    v0.1 – Foundation (MVP)

Core server + tools

Spec compliance

stdio + HTTP

Docs + examples

v0.2 – Production Hardening

Auth middleware patterns

Metrics hooks

Rate limiting

Better schema support

v0.3 – Ecosystem

Client SDK

Streaming support

Compliance test suite

Integration examples (Claude, IDEs)

14. Risks & Mitigations
    Risk Mitigation
    MCP spec changes Isolate protocol layer
    Over-engineering Strict MVP scope
    Low adoption Best-in-class DX + docs
    Competing libs emerge Be first + opinionated
15. Open Source Strategy

Apache 2.0 or MIT

Public roadmap

Strong examples

“Official-looking” docs

Encourage downstream frameworks

16. Strategic Fit (Why this matters for you)

This library:

Complements Obvia, Relicta, Ops tooling

Establishes thought leadership in MCP

Becomes infrastructure others build on

Creates long-term leverage (ecosystem gravity)

17. Decision

GO ✔️
This is:

a real gap

the right timing

aligned with your Go + infra + agent focus

feasible to ship incrementally
