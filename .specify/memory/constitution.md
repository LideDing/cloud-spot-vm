<!--
  Sync Impact Report
  ==================
  Version change: N/A (initial) → 1.0.0
  Modified principles: N/A (initial creation)
  Added sections:
    - Core Principles (5 principles)
    - Cost Optimization Constraints
    - Development Workflow
    - Governance
  Removed sections: N/A
  Templates requiring updates:
    - .specify/templates/plan-template.md ✅ no update needed (generic template)
    - .specify/templates/spec-template.md ✅ no update needed (generic template)
    - .specify/templates/tasks-template.md ✅ no update needed (generic template)
    - .specify/templates/checklist-template.md ✅ no update needed (generic template)
  Follow-up TODOs:
    - RATIFICATION_DATE set to today (2026-03-31) as initial adoption
-->

# Cloud Spot VM Constitution

## Core Principles

### I. Cost Minimization First

All architectural and operational decisions MUST prioritize minimizing
cloud infrastructure cost. The project exists to leverage Tencent Cloud
Spot CVM instances for the lowest possible compute expense. Every feature
addition MUST be evaluated against its impact on cost efficiency.

- Spot instances MUST be preferred over on-demand instances in all cases
  where workload tolerance allows.
- Automatic fallback strategies MUST consider cost before availability
  unless explicitly overridden by configuration.
- Resource lifecycle management MUST include timely release of unused
  instances to avoid unnecessary charges.

### II. Resilience & Availability

Spot instances are inherently preemptible. The system MUST handle
interruptions gracefully without data loss or prolonged downtime.

- Instance reclamation events MUST be detected and handled automatically.
- State persistence MUST NOT depend on the longevity of any single
  Spot instance.
- Region/zone failover MUST be supported to maintain service continuity
  when spot capacity is unavailable.
- Health checks and automatic recovery MUST be implemented for all
  managed instances.

### III. Configuration-Driven Design

All operational parameters MUST be externalized via configuration
(environment variables or config files), not hard-coded.

- Tencent Cloud credentials, region preferences, instance types, and
  pricing thresholds MUST be configurable without code changes.
- The `.env` file pattern MUST be used for local development; production
  MUST use secure secret management.
- Sensible defaults MUST be provided for all optional configuration
  values.

### IV. Security & Credential Safety

Cloud credentials and sensitive data MUST never be committed to version
control or exposed in logs.

- The `.env` file MUST remain in `.gitignore` at all times.
- API keys and secret IDs MUST be loaded from environment variables or
  a secure vault.
- All external API communication (Tencent Cloud SDK, webhook callbacks)
  MUST use authenticated and encrypted channels.
- Authentication middleware MUST protect all management API endpoints.

### V. Simplicity & Maintainability

The codebase MUST remain simple, idiomatic Go. Avoid premature
abstraction and over-engineering.

- YAGNI: Do not implement features until they are concretely needed.
- Standard library and well-maintained dependencies MUST be preferred
  over custom implementations.
- Package structure MUST follow Go conventions (`cmd/`, `internal/`,
  clear separation of concerns).
- Code MUST be formatted with `gofmt` and pass `go vet` without
  warnings.

## Cost Optimization Constraints

- Instance type selection MUST support automatic comparison of spot
  prices across available zones and regions.
- The system MUST support configurable maximum bid prices to prevent
  cost overruns.
- Monitoring and alerting MUST be in place to detect abnormal spending
  patterns.
- Batch operations (create/destroy multiple instances) MUST be
  supported to reduce API call overhead and improve efficiency.
- All Tencent Cloud API interactions MUST implement retry logic with
  exponential backoff to handle rate limits gracefully.

## Development Workflow

- All changes MUST be submitted via pull requests with at least one
  review before merging.
- The `main` branch MUST always be in a deployable state.
- Commit messages MUST follow conventional commit format
  (e.g., `feat:`, `fix:`, `docs:`, `refactor:`).
- Integration tests covering Tencent Cloud API interactions SHOULD use
  mocked SDK responses to avoid incurring real cloud costs during CI.
- Shell test scripts (`test_*.sh`) MUST be kept up to date with API
  changes.

## Governance

This constitution is the authoritative source of project principles and
constraints. All design decisions, code reviews, and feature proposals
MUST be evaluated against these principles.

- **Amendment procedure**: Any principle change MUST be documented with
  rationale, reviewed, and approved before merging. The constitution
  version MUST be incremented accordingly.
- **Versioning policy**: Semantic versioning (MAJOR.MINOR.PATCH) applies.
  MAJOR for principle removals/redefinitions, MINOR for new principles
  or material expansions, PATCH for clarifications and wording fixes.
- **Compliance review**: All pull requests MUST include a brief
  constitution compliance note when introducing new features or
  architectural changes.

**Version**: 1.0.0 | **Ratified**: 2026-03-31 | **Last Amended**: 2026-03-31
