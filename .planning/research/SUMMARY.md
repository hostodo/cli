# Project Research Summary

**Project:** Hostodo CLI Security & Deployment Enhancements
**Domain:** Cloud VPS CLI tool with OAuth security and interactive deployment
**Researched:** 2026-02-15
**Confidence:** HIGH

## Executive Summary

The Hostodo CLI is a cloud VPS management tool that needs security hardening and UX improvements to achieve production readiness. Research shows the current OAuth device flow implementation uses insufficient entropy (8-digit device codes vs industry standard 12+ characters), lacks session audit trails (hard-delete vs soft-delete), and misses standard CLI features users expect from competitors like DigitalOcean and Vultr.

The recommended approach prioritizes security fixes first (device code entropy, session tracking, soft-delete), then adds hostname-based commands to match competitor UX, followed by an interactive deployment wizard that sets the product apart. The Go/Cobra stack is ideal for this domain with Bubbletea providing TUI capabilities that competitors lack. Critical dependencies include zalando/go-keyring for secure token storage and crypto/rand.Text() for RFC 8628-compliant device code generation.

Key risks center on maintaining backward compatibility while restructuring commands, handling hostname ambiguity without database schema changes, and ensuring interactive prompts don't break automation. All risks are mitigated through careful phasing: security first (no breaking changes), then additive features (aliases + hostname support), then new capabilities (deployment wizard with dual-mode support for interactive + automation).

## Key Findings

### Recommended Stack

Go 1.24+ with Cobra framework provides the industry-standard foundation used by Kubernetes, Docker, and GitHub CLI. The latest Go release adds crypto/rand.Text() for secure token generation (128+ bits entropy guaranteed), 15-25% GC pause improvements, and Swiss Tables map implementation for better performance. Bubbletea offers production-ready TUI capabilities based on The Elm Architecture, enabling interactive workflows competitors lack.

**Core technologies:**
- **Go 1.24+**: Latest stable with crypto/rand.Text() for secure device codes, performance improvements, WASI support
- **Cobra v1.10+**: Industry standard CLI framework (173k+ projects), command structure, shell completion, help generation
- **Bubbletea v1.3+**: Interactive TUI framework for complex workflows, state management, differentiator from competitors
- **go-keyring v0.2.6**: Cross-platform secure credential storage (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **Huh v2+**: Interactive forms/prompts with accessibility, theme support, 5 field types for deployment wizard

**Critical security requirements:**
- Use crypto/rand.Text() for device codes (12+ chars, 128-bit entropy minimum)
- Store tokens in OS keychain, never in config files (MANDATORY)
- SHA-256 hash tokens before database storage
- Implement soft-delete for session audit trails

### Expected Features

Research across AWS CLI, DigitalOcean doctl, and Vultr CLI reveals clear table stakes vs differentiators.

**Must have (table stakes):**
- JSON/YAML/table output formats (all competitors support this)
- Shell completion (bash/zsh/fish) - missing, competitors all have it
- Hostname-based commands (DigitalOcean, Vultr support this) - missing, critical UX gap
- Session listing and revocation (AWS has this, others don't) - partial implementation
- Non-interactive mode with --yes flag for CI/CD
- Proper error messages with actionable next steps
- Resource filtering (--status, --region) for users with 10+ instances

**Should have (competitive differentiation):**
- OAuth device flow (better than competitors' static API tokens) - implemented, needs hardening
- Interactive TUI (Bubble Tea) - implemented, maintain and enhance as differentiator
- Interactive deployment wizard (no competitor has this) - planned, major UX win
- Soft-delete sessions with audit trail (AWS has this, others don't)
- Command aliases (i for instances, a for auth) - not implemented, quick UX win
- Smart progress indicators during long operations
- Cost estimation during instance creation

**Defer (v2+):**
- Bandwidth monitoring with terminal graphs
- Resource tagging and bulk operations
- Pre-flight checks for operations
- Built-in SSH client (anti-feature - document the simple approach instead)
- Plugin system (maintenance nightmare, scope creep)

### Architecture Approach

Standard three-layer architecture with clean separation of concerns: CLI layer (Cobra commands) delegates to service layer (pkg/api, pkg/auth, pkg/ui) which communicates with external systems (backend REST API, OS keychain, terminal). This structure scales well as new features are added and enables clean testing with mocked backends.

**Major components:**
1. **cmd/auth, cmd/instances** - Command definitions, argument parsing, user interaction via Cobra
2. **pkg/api** - HTTP client, hostname resolver with caching, REST API communication
3. **pkg/auth** - OAuth device flow, keychain token storage, device ID persistence
4. **pkg/ui** - Formatters (JSON/table), Bubbletea TUI, interactive prompts (Huh forms)
5. **pkg/config** - Config file management (~/.hostodo/config.json), device ID tracking

**Key patterns:**
- Hostname resolution with client-side caching (no backend changes needed)
- Dual-mode commands: interactive prompts for humans, flags for automation
- Backward-compatible aliases (hostodo list = hostodo instances list)
- Soft-delete sessions with revoked_at timestamp (not hard DELETE)

### Critical Pitfalls

1. **Device Code Entropy Insufficient** - 8-digit numeric codes (10^8 combinations) are brute-forceable in hours. RFC 8628 explicitly requires "very high entropy" since device codes aren't user-facing. Fix: 12+ alphanumeric characters (36^12 = 4.7x10^18 combinations), format as A3K9-M7P2-X4Q8 with dashes, use crypto/rand.Text(). Address in Phase 1.

2. **Hard-Deleted Sessions Break Audit Compliance** - DELETE FROM sessions destroys audit trails needed for GDPR/SOC2. Can't answer "when was session revoked" or "what sessions were active during incident." Fix: Add revoked_at timestamp, query WHERE revoked_at IS NULL for active sessions, retain revoked sessions for 90 days. Address in Phase 1.

3. **Hostname Ambiguity Without Resolution Strategy** - Users create multiple instances named "webserver" (dev/staging/prod), CLI command operates on wrong instance or fails. Fix: Detect duplicates (SELECT COUNT), show disambiguation UI in interactive mode, require --id flag in non-interactive mode. Address in Phase 2.

4. **Interactive Prompts Break Automation** - Deployment wizard adds prompts that hang CI/CD scripts indefinitely. Fix: Detect TTY with isatty(stdin), require flags if non-TTY, provide --non-interactive mode, document automation patterns. Address in Phase 2.

5. **Breaking Changes Without Deprecation Path** - Changing command structure (hostodo instances list → hostodo list) breaks existing scripts. Fix: Support both old and new commands for 2+ versions, print deprecation warnings, document migration guide. Address in Phase 2.

## Implications for Roadmap

Based on research, suggested 3-phase structure prioritizing security, then UX parity with competitors, then differentiation:

### Phase 1: Security Hardening & Session Management
**Rationale:** Security vulnerabilities must be fixed before announcing public CLI availability. All changes are backend-focused with minimal CLI impact, avoiding breaking changes while establishing foundation for session management features.

**Delivers:**
- RFC 8628-compliant device codes (12+ chars, high entropy)
- Session audit trail (soft-delete with revoked_at timestamps)
- Last-used tracking for session monitoring
- Token rotation hardening (atomic swap, token families)
- CLI session listing and revocation commands

**Addresses features:**
- Session listing/revocation (table stakes from FEATURES.md)
- Soft-delete sessions (security improvement)
- Audit trail visibility (differentiator)
- Device code security (critical pitfall #1)

**Avoids pitfalls:**
- Pitfall #1: Device code entropy (fixes brute-force vulnerability)
- Pitfall #2: Hard-deleted sessions (enables compliance)
- Pitfall #3: Token rotation race conditions (prevents random logouts)

**Backend changes required:**
- odoauth/services/DeviceFlowService.py (device code generation)
- odoauth/models.py (add revoked_at, last_used_at fields)
- odoauth/views.py (update last_used_at on auth'd requests)

**CLI changes required:**
- cmd/auth/sessions.go (new: list/revoke commands)
- pkg/api/sessions.go (new: session API client)
- pkg/auth/oauth.go (handle 12-char device codes with dashes)

**Research flag:** SKIP - OAuth patterns well-documented, RFC 8628 explicit, established best practices

### Phase 2: Hostname Support & Command Usability
**Rationale:** After security is solid, add UX features to match competitor capabilities. Hostname-based commands are the most frequently requested feature and enable more intuitive instance management. Command aliases reduce typing friction for power users.

**Delivers:**
- Hostname-based instance operations (hostodo start myserver)
- Root-level command aliases (hostodo list = hostodo instances list)
- Hostname ambiguity detection and disambiguation UI
- Shell completion (bash/zsh/fish)
- Resource filtering (--status, --region)

**Addresses features:**
- Hostname support (critical gap vs competitors)
- Command aliases (UX improvement)
- Shell completion (table stakes)
- Resource filtering (needed for 10+ instances)

**Avoids pitfalls:**
- Pitfall #4: Hostname ambiguity (detects duplicates, shows disambiguation)
- Pitfall #6: Breaking changes (aliases preserve old commands)

**Uses stack elements:**
- Client-side hostname resolver (no backend changes)
- Cobra alias configuration (trivial to implement)
- Cobra shell completion (built-in feature)

**Implements architecture:**
- pkg/api/resolver.go (hostname→instance_id with caching)
- cmd/root.go (root-level aliases)
- Modification of cmd/instances/* (use resolver for all commands)

**Research flag:** SKIP - Standard CLI patterns, well-documented in Cobra, no novel integration

### Phase 3: Interactive Deployment & Advanced Features
**Rationale:** Once security and UX parity are achieved, add features that differentiate from competitors. Interactive deployment wizard provides guided experience competitors lack, while maintaining automation support through flag-based mode.

**Delivers:**
- Interactive deployment wizard (region→plan→OS→SSH keys)
- Dual-mode support (interactive prompts + automation flags)
- Smart progress indicators during provisioning
- Cost estimation before instance creation
- SSH key injection during deployment
- Non-interactive mode for CI/CD

**Addresses features:**
- Interactive deployment wizard (major differentiator)
- SSH key management (competitive feature)
- Cost estimation (user-requested)
- Non-interactive mode (table stakes for automation)
- Smart progress indicators (UX improvement)

**Avoids pitfalls:**
- Pitfall #5: Interactive prompts breaking automation (TTY detection, flag mode)

**Uses stack elements:**
- Huh v2+ for interactive forms
- Bubbles spinner/progress for status updates
- TTY detection (isatty) for mode selection

**Implements architecture:**
- pkg/ui/prompts.go (interactive wizard)
- cmd/instances/deploy.go (new command with dual-mode)
- pkg/api/client.go (deployment API methods)

**Research flag:** LIKELY NEEDS RESEARCH - Deployment API contracts, order provisioning flow, status polling patterns need backend investigation during phase planning.

### Phase Ordering Rationale

- **Security first:** Vulnerabilities must be fixed before public announcement; delay increases risk window. Backend-only changes minimize coordination overhead.
- **UX parity second:** Hostname support is prerequisite for deployment wizard (users expect to reference newly created instances by name). Aliases provide quick wins while deployment wizard is being built.
- **Differentiation third:** Deployment wizard is complex and benefits from hostname resolver and command structure being stable. Phasing allows user feedback on core features to inform wizard design.
- **Dependency management:** Phase 2 depends on Phase 1 session infrastructure for auth; Phase 3 depends on Phase 2 hostname resolver for post-deployment operations.
- **Risk mitigation:** Each phase is independently shippable, allowing early feedback and course correction.

### Research Flags

**Phases likely needing deeper research during planning:**
- **Phase 3 (Deployment):** Backend order API contracts, provisioning status polling, error handling for failed deployments, template/plan/region listing endpoints need investigation via /gsd:research-phase.

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Security):** RFC 8628 explicit, OAuth best practices well-documented, GORM soft-delete is standard Django pattern.
- **Phase 2 (Hostname):** Client-side resolution pattern established, Cobra aliases trivial, shell completion built-in to framework.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Go 1.24 release notes verified, Cobra/Bubbletea production usage confirmed, RFC 8628 explicit on requirements |
| Features | HIGH | Competitor analysis (AWS/DO/Vultr CLIs) comprehensive, table stakes vs differentiators clear from user patterns |
| Architecture | HIGH | Standard three-layer CLI architecture, patterns verified in kubectl/doctl/gh implementations |
| Pitfalls | HIGH | Security issues (device code entropy, session audit) backed by RFC standards and compliance requirements |

**Overall confidence:** HIGH

### Gaps to Address

Research was comprehensive with official sources for all critical areas. Minor gaps to resolve during implementation:

- **Deployment API contracts:** Backend order provisioning endpoint structure needs investigation (suggest /gsd:research-phase during Phase 3 planning). FEATURES.md references existing POST /client/orders/ but polling/status patterns need verification.

- **Hostname uniqueness validation:** Need to confirm whether backend enforces UNIQUE constraint on hostname field or allows duplicates. If duplicates allowed, Phase 2 must implement disambiguation UI. If enforced, simpler error handling suffices.

- **Token refresh interval:** Backend refresh token expiry not documented in research. Need to verify whether access tokens auto-refresh or require explicit refresh flow. Impacts session management UX in Phase 1.

- **Template/plan listing pagination:** FEATURES.md mentions pagination support but limit on template/plan lists unknown. May impact deployment wizard if 100+ templates exist. Address during Phase 3 planning.

All gaps are implementable details, not architectural unknowns. Proceed to requirements with current research.

## Sources

### Primary (HIGH confidence)
- [RFC 8628 - OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628) - Device code security requirements
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24) - crypto/rand.Text() details, performance improvements
- [Cobra Documentation](https://cobra.dev/) - CLI framework best practices
- [GORM Delete Documentation](https://gorm.io/docs/delete.html) - Soft delete patterns
- [AWS CLI Output Formats](https://docs.aws.amazon.com/cli/v1/userguide/cli-usage-output-format.html) - Competitor feature analysis
- [DigitalOcean doctl CLI Reference](https://docs.digitalocean.com/reference/doctl/) - Feature comparison
- [Command Line Interface Guidelines](https://clig.dev/) - Industry best practices

### Secondary (MEDIUM confidence)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics) - Token rotation patterns
- [Charmbracelet Huh](https://github.com/charmbracelet/huh) - Interactive forms library
- [Bubbles Components](https://pkg.go.dev/github.com/charmbracelet/bubbles) - TUI components
- [WorkOS CLI Authentication Best Practices](https://workos.com/blog/best-practices-for-cli-authentication-a-technical-guide) - Auth patterns
- [10 Essential Audit Trail Best Practices for 2026](https://signal.opshub.me/audit-trail-best-practices/) - Compliance requirements
- [Soft delete vs hard delete | AppMaster](https://appmaster.io/blog/soft-delete-vs-hard-delete) - Database patterns
- [AWS CLI version 2 Migration Guide](https://docs.aws.amazon.com/cli/latest/userguide/cliv2-migration.html) - Backward compatibility patterns

### Tertiary (LOW confidence)
- [Vultr CLI Management Guide](https://blogs.vultr.com/How-to-Easily-Manage-Instances-with-Vultr-CLI) - Feature comparison (blog post)
- [OAuth 2.1 Features You Can't Ignore in 2026](https://rgutierrez2004.medium.com/oauth-2-1-features-you-cant-ignore-in-2026-a15f852cb723) - Future OAuth trends

---
*Research completed: 2026-02-15*
*Ready for roadmap: yes*
