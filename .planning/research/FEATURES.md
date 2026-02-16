# Feature Research

**Domain:** Cloud VPS CLI Tools
**Researched:** 2026-02-15
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete or unprofessional.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **JSON output format** | Required for scripting/automation. Every major cloud CLI supports this (AWS, Azure, GCP, DO, Vultr). | LOW | Already implemented. Machine-parseable output is non-negotiable. |
| **Multiple output formats** | Users expect text/table/JSON/YAML options for different contexts (human vs automation). | LOW | Partially implemented (JSON, table). YAML is nice-to-have. |
| **Credential refresh** | Temporary tokens that auto-refresh. Long-lived credentials are a security anti-pattern. | MEDIUM | OAuth device flow provides this. Token refresh logic needed. |
| **Session listing** | Users expect to see active sessions/devices and revoke access per session (like GitHub, AWS). | LOW | Backend supports this. CLI needs `auth sessions list/revoke` commands. |
| **Proper error messages** | Clear, actionable errors (not stack traces). Show next steps when auth fails or resources don't exist. | LOW | Essential for UX. Include error codes, recovery hints. |
| **Non-interactive mode** | Support `--yes` or `--force` flags for CI/CD. Don't block automated workflows with prompts. | LOW | Critical for CI/CD pipelines. All destructive ops need bypass flag. |
| **Wait/polling for async ops** | Commands like `start/stop/reboot` should wait for completion with timeout. Users expect progress indication. | MEDIUM | Partially implemented. Need consistent polling pattern across all ops. |
| **Resource filtering** | Filter by status, region, tag. Standard pattern: `instances list --status=running --region=us-east`. | MEDIUM | Not implemented. Users with 10+ instances need this. |
| **Pagination support** | Handle large result sets (100+ instances). Limit/offset flags. | LOW | Backend supports this. CLI has basic implementation. |
| **Config file support** | Store API URL, default region, output format preferences. Standard location: `~/.hostodo/config.yml`. | LOW | Implemented at `~/.hostodo/config.json`. Consider YAML migration. |
| **Shell completion** | Tab completion for commands, flags, and resource IDs. Expected by devs used to kubectl/doctl. | MEDIUM | Not implemented. Cobra supports this out-of-box. High value for UX. |
| **Help documentation** | Every command needs `--help` with examples. Poor help = frustrated users. | LOW | Implemented via Cobra. Ensure all examples are current. |
| **Version information** | `--version` shows CLI version, API version compatibility. Critical for support. | LOW | Implemented. Include build date/commit for debugging. |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Hostname-based commands** | `hostodo start my-web-server` instead of `hostodo start abc123def`. Eliminates lookup friction. | MEDIUM | Requires hostname→ID resolution with caching. Check for uniqueness. Huge UX win for manual ops. |
| **Interactive deployment wizard** | Guided instance creation: region→plan→OS→SSH keys with smart defaults. | HIGH | Terraform/Heroku pattern. Reduces cognitive load for new users. Can still support `--json` config for automation. |
| **Smart wait with progress** | Show live progress during provisioning: "Creating instance... Configuring network... Installing OS... Ready (2m 34s)". | MEDIUM | Websocket or polling-based status updates. Much better UX than "please wait...". |
| **Audit trail visibility** | `auth sessions` shows login history: device, IP, location, last active. Transparency builds trust. | LOW | Backend already tracks LoginLog. Just expose via CLI. Security-conscious users love this. |
| **Soft-delete sessions** | Distinguish revoke (hide session, preserve audit) vs delete (remove record). Industry standard pattern. | LOW | Backend change needed. Revoke = mark inactive, delete = hard delete. Supports compliance. |
| **Command aliases** | `hostodo i ls` = `hostodo instances list`. Kubectl-style shortcuts for power users. | LOW | Cobra supports aliases trivially. `i/ins` for instances, `a` for auth, etc. |
| **Resource tagging** | Tag instances with labels: `env=prod`, `team=backend`. Filter/group by tags. | MEDIUM | Backend feature. CLI just exposes it. Valuable for teams managing 20+ instances. |
| **Cost estimation** | Show billing impact before deploying: "This instance costs $5/mo". | LOW | Read from plan data. Simple calculation. Prevents bill shock. |
| **SSH key injection** | Add SSH keys during deployment without manual API calls: `--ssh-key=~/.ssh/id_rsa.pub`. | MEDIUM | Common workflow. DO/Linode support this. Reduces provisioning steps. |
| **Bandwidth monitoring** | `instances bandwidth <id>` shows usage graphs in terminal. Prevent surprise overages. | MEDIUM | Requires bandwidth data from backend. Charming/lipgloss for graphs. |
| **Bulk operations** | `instances stop --tag=env:staging` stops multiple instances. Power user feature. | MEDIUM | Requires tag filtering + confirmation. Dangerous but valuable for larger fleets. |
| **Interactive TUI** | Bubble Tea interface for browsing instances, not just listing. | LOW | Already implemented. Maintain and enhance. Sets apart from bare tables. |
| **Pre-flight checks** | Validate operation before executing: "Warning: Stopping this instance will disconnect 3 active SSH sessions". | HIGH | Requires backend integration for connection tracking. Nice-to-have for safety. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Built-in SSH client** | "I want `hostodo ssh <instance>` to connect." | Duplicates `ssh` which is already installed everywhere. Adds complexity for SSH key management, agents, port forwarding. | Document: `eval $(hostodo instances get <id> --format=ssh)` → `ssh root@1.2.3.4`. Or alias helper. |
| **Embedded update mechanism** | "Auto-update the CLI like browsers." | Package managers (brew/apt/yum) handle this. Competing with system tools causes conflicts. | Document: `brew upgrade hostodo`. Trust established update channels. |
| **Local state caching** | "Cache instance list to speed up commands." | Stale data causes confusion. Users expect real-time state. Syncing cache adds complexity. | Fast API responses solve this. Optionally cache for offline `--help`, not live data. |
| **Configuration profiles** | "Let me switch between dev/staging/prod API endpoints." | Confusing for average users. Environment variables handle this: `HOSTODO_API_URL=...`. | Document envvar approach. Simple and standard. |
| **Inline instance creation** | "Pack everything into one command with 20 flags." | Unreadable: `--ram=4096 --cpu=2 --disk=50 --region=us-east --os=ubuntu-22.04 --ssh-key=...`. Error-prone. | Interactive wizard for humans. JSON config file for automation. |
| **Plugin system** | "Let users extend the CLI with plugins." | Maintenance nightmare. Security risk (untrusted code execution). Scope creep. | Keep core focused. If extensibility needed, document API client usage. |
| **Real-time streaming dashboard** | "Live updating TUI showing all instances." | Websocket complexity. Polling overhead. Niche use case (monitoring tools exist). | Offer `watch -n 5 hostodo instances list`. Don't reinvent monitoring. |
| **Multi-cloud support** | "Manage AWS/DO/Vultr from one CLI." | Each provider has unique APIs, authentication, resource models. Unmaintainable abstraction. | Focus on Hostodo excellence. Let users use multiple CLIs. |
| **Credential encryption** | "Encrypt tokens with a master password." | System keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager) handles this securely. Reinventing breaks OS integration. | Use OS keychain (already implemented). Don't roll crypto. |
| **Verbose logging by default** | "Show all HTTP requests for debugging." | Clutters output. Exposes sensitive data (auth tokens in logs). | Offer `--debug` flag. Default to clean, user-facing messages. |

## Feature Dependencies

```
OAuth Device Flow (implemented)
    └──requires──> Session Management (partially implemented)
                       └──requires──> Session Listing (not implemented)
                       └──requires──> Session Revocation (not implemented)
                       └──requires──> Soft-Delete Sessions (not implemented)

Hostname-based Commands (planned)
    └──requires──> Hostname→ID Resolution (not implemented)
    └──requires──> Caching Layer (optional, for performance)

Interactive Deployment Wizard (planned)
    └──requires──> SSH Key Management API (backend exists)
    └──requires──> Plan/Region/Template Listing APIs (backend exists)
    └──enhances──> Cost Estimation (not implemented)
    └──enhances──> Pre-flight Checks (not implemented)

Bulk Operations (future)
    └──requires──> Resource Filtering (not implemented)
    └──requires──> Tag Support (backend + CLI)
    └──conflicts──> Non-interactive Mode (must add --yes flag)

Wait/Polling (partially implemented)
    └──enhances──> Start/Stop/Reboot Commands (implemented)
    └──enhances──> Instance Deployment (planned)
    └──requires──> Smart Progress Display (not implemented)

Command Aliases (planned)
    └──requires──> Cobra Alias Configuration (trivial)
    └──independent──> No dependencies
```

### Dependency Notes

- **Session Management requires OAuth Device Flow:** Sessions are created via device flow login. Already implemented. Need to expose session list/revoke endpoints.
- **Hostname-based commands require resolution logic:** Must query API to map hostname→instance ID. Add caching to avoid repeated lookups. Check for duplicate hostnames (error if ambiguous).
- **Interactive wizard enhances deployment UX:** Independent feature, but benefits from cost estimation and pre-flight checks if available.
- **Bulk operations conflict with safety:** Dangerous without confirmation. Must implement `--yes` flag AND clear warning messages.
- **Smart progress requires polling:** Long-running operations (deploy, stop, start) need status polling. Unify polling pattern across all commands.
- **Command aliases are independent:** Pure CLI sugar. No backend dependencies. Implement early for quick UX win.

## MVP Definition

### Launch With (v1 - Current State)

Minimum viable product — what's needed to validate the concept.

- [x] OAuth device flow authentication
- [x] Session storage in system keychain
- [x] List instances (interactive TUI + JSON)
- [x] Get instance details
- [x] Start/stop/reboot instances with wait
- [x] Multiple output formats (JSON, table, interactive)
- [x] Basic error handling
- [x] Help documentation

**Status:** ✅ MVP complete. CLI is functional for basic instance management.

### Add After Validation (v1.x - Security & UX Hardening)

Features to add once core is working. Prioritize based on user feedback.

**Security Improvements (Priority 1):**
- [ ] Session listing (`auth sessions list`)
- [ ] Session revocation (`auth sessions revoke <session-id>`)
- [ ] Soft-delete sessions (revoke vs delete)
- [ ] Increase device code entropy (12+ chars)
- [ ] Last-used timestamp tracking for sessions
- [ ] Audit trail visibility in CLI

**UX Improvements (Priority 2):**
- [ ] Command aliases (`i` for instances, `a` for auth)
- [ ] Shell completion (bash/zsh/fish)
- [ ] Hostname-based commands (`hostodo start my-server`)
- [ ] Resource filtering (`--status=running --region=us-east`)
- [ ] Config file for defaults (output format, API URL)
- [ ] Smart progress indicators during wait

**Automation Support (Priority 3):**
- [ ] Non-interactive mode (`--yes` flag)
- [ ] YAML output format
- [ ] Exit codes for scripting
- [ ] Consistent polling pattern
- [ ] Timeout configuration for waits

### Future Consideration (v2+ - Advanced Features)

Features to defer until product-market fit is established.

**Deployment Features:**
- [ ] Interactive deployment wizard
- [ ] SSH key management
- [ ] Cost estimation during creation
- [ ] Template selection UI

**Management Features:**
- [ ] Bandwidth monitoring and graphs
- [ ] Backup management
- [ ] Reverse DNS configuration
- [ ] Billing/invoice viewing

**Power User Features:**
- [ ] Resource tagging
- [ ] Bulk operations (`--tag` filters)
- [ ] Pre-flight checks
- [ ] Custom output templates

**Why defer:**
- Deployment wizard needs user research (what options matter?)
- Tagging requires backend schema changes
- Bulk ops are dangerous; need careful design
- Bandwidth/billing are nice-to-have; core is instance management

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority | Complexity |
|---------|------------|---------------------|----------|------------|
| **Session listing/revocation** | HIGH | LOW | P1 | LOW |
| **Soft-delete sessions** | MEDIUM | LOW | P1 | LOW |
| **Increase device code entropy** | MEDIUM | LOW | P1 | LOW |
| **Command aliases** | HIGH | LOW | P1 | LOW |
| **Shell completion** | HIGH | MEDIUM | P1 | MEDIUM |
| **Hostname-based commands** | HIGH | MEDIUM | P1 | MEDIUM |
| **Non-interactive mode** | HIGH | LOW | P1 | LOW |
| **Resource filtering** | HIGH | MEDIUM | P2 | MEDIUM |
| **Config file defaults** | MEDIUM | LOW | P2 | LOW |
| **YAML output** | MEDIUM | LOW | P2 | LOW |
| **Smart progress indicators** | MEDIUM | MEDIUM | P2 | MEDIUM |
| **Consistent polling pattern** | MEDIUM | MEDIUM | P2 | MEDIUM |
| **Interactive deployment wizard** | HIGH | HIGH | P2 | HIGH |
| **SSH key injection** | MEDIUM | MEDIUM | P2 | MEDIUM |
| **Cost estimation** | MEDIUM | LOW | P3 | LOW |
| **Bandwidth monitoring** | MEDIUM | MEDIUM | P3 | MEDIUM |
| **Resource tagging** | HIGH | HIGH | P3 | HIGH |
| **Bulk operations** | HIGH | MEDIUM | P3 | MEDIUM |
| **Pre-flight checks** | LOW | HIGH | P3 | HIGH |

**Priority key:**
- **P1:** Must have for production readiness (security + essential UX)
- **P2:** Should have for competitive parity (match DO/Vultr CLIs)
- **P3:** Nice to have for differentiation (above and beyond)

**Implementation cost:**
- **LOW:** < 1 day (simple flag, API call, config change)
- **MEDIUM:** 1-3 days (requires design, testing, edge cases)
- **HIGH:** 5+ days (needs research, backend changes, or complex UX)

## Competitor Feature Analysis

| Feature | AWS CLI | DigitalOcean doctl | Vultr CLI | Hostodo CLI | Our Approach |
|---------|---------|-------------------|-----------|-------------|--------------|
| **JSON output** | ✅ (default) | ✅ | ✅ | ✅ | Same. Table stakes. |
| **YAML output** | ✅ | ✅ | ✅ | ❌ | Add in v1.x. Low effort. |
| **Multiple formats** | ✅ (json/yaml/text/table) | ✅ (json/yaml/text) | ✅ (json/yaml/text) | ✅ (json/table/interactive) | Add YAML for parity. |
| **Shell completion** | ✅ | ✅ | ✅ | ❌ | P1. Cobra built-in. |
| **OAuth device flow** | ✅ (SSO) | ❌ (API tokens only) | ❌ (API tokens only) | ✅ | Differentiator. Better security. |
| **Session management** | ✅ (IAM session tokens) | ❌ | ❌ | Partial | Expose session list/revoke. Security win. |
| **Interactive TUI** | ❌ (text only) | ❌ (text only) | ❌ (text only) | ✅ (Bubble Tea) | Differentiator. Keep and enhance. |
| **Hostname support** | ✅ (Name tags) | ✅ (Droplet names) | ✅ (--host flag) | ❌ | P1. Essential for usability. |
| **Resource filtering** | ✅ (extensive) | ✅ (--tag, --region) | ✅ (filters) | ❌ | P2. Need for 10+ instances. |
| **Wait/polling** | ✅ (--wait flag) | ✅ (automatic) | ✅ (status polling) | ✅ (implemented) | Enhance with better progress. |
| **Bulk operations** | ✅ (via filters) | ✅ (via tags) | ❌ | ❌ | P3. Risky but valuable. |
| **SSH integration** | ❌ | ✅ (doctl compute ssh) | ❌ | ❌ | Anti-feature. Don't build. |
| **Interactive wizard** | ❌ | ❌ | ❌ | ❌ | Differentiator. Build for v2. |
| **Cost estimation** | ✅ (AWS Pricing API) | ✅ (plan details) | ✅ (plan pricing) | ❌ | P3. Read from plan data. |
| **Audit logging** | ✅ (CloudTrail) | ❌ | ❌ | Partial (backend) | Expose via CLI. Security feature. |

**Key takeaways:**
1. **Parity gaps:** Shell completion, hostname support, YAML output, resource filtering.
2. **Differentiators:** OAuth device flow (vs API tokens), Interactive TUI, session management, planned wizard.
3. **Anti-features:** Don't copy doctl's SSH integration. Document the simple approach instead.
4. **Security advantage:** Hostodo has better auth (OAuth device flow + session tracking) than competitors (static API tokens).

## Sources

### Official Documentation
- [AWS CLI Output Formats](https://docs.aws.amazon.com/cli/v1/userguide/cli-usage-output-format.html) - MEDIUM confidence
- [DigitalOcean doctl CLI Reference](https://docs.digitalocean.com/reference/doctl/) - MEDIUM confidence
- [Vultr CLI Documentation](https://docs.vultr.com/reference/vultr-cli) - MEDIUM confidence
- [Google Cloud gcloud CLI Scripting](https://docs.cloud.google.com/sdk/docs/scripting-gcloud) - HIGH confidence
- [Azure CLI Output Formats](https://learn.microsoft.com/en-us/cli/azure/format-output-azure-cli) - HIGH confidence

### Security & Authentication
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics) - HIGH confidence
- [WorkOS CLI Authentication Best Practices](https://workos.com/blog/best-practices-for-cli-authentication-a-technical-guide) - HIGH confidence
- [OAuth 2.1 Features You Can't Ignore in 2026](https://rgutierrez2004.medium.com/oauth-2-1-features-you-cant-ignore-in-2026-a15f852cb723) - MEDIUM confidence

### UX & Design Patterns
- [CLI Best Practices](https://hackmd.io/@arturtamborski/cli-best-practices) - MEDIUM confidence
- [The Poetics of CLI Command Names](https://smallstep.com/blog/the-poetics-of-cli-command-names/) - MEDIUM confidence
- [Terraform Apply Command Guide](https://www.env0.com/blog/terraform-apply-guide-command-options-and-examples) - MEDIUM confidence

### Long-Running Operations
- [Google Cloud Polling Long Running Operations](https://docs.cloud.google.com/service-infrastructure/docs/polling-operations) - HIGH confidence
- [Long Running Tasks in MCP: The Call-Now, Fetch-Later Pattern](https://agnost.ai/blog/long-running-tasks-mcp/) - MEDIUM confidence
- [AWS Session Manager Logging](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-auditing.html) - HIGH confidence

### Anti-Patterns
- [Feature Creep: Causes and How to Avoid It](https://www.june.so/blog/feature-creep-causes-consequences-and-how-to-avoid-it) - MEDIUM confidence
- [The Feature Creep Anti-Pattern](https://develpreneur.com/the-feature-creep-anti-pattern/) - MEDIUM confidence

### Competitor Analysis
- [GitHub vultr-cli](https://github.com/vultr/vultr-cli) - HIGH confidence (official source)
- [Vultr CLI Management Guide](https://blogs.vultr.com/How-to-Easily-Manage-Instances-with-Vultr-CLI) - MEDIUM confidence
- [DigitalOcean vs Linode Comparison](https://blog.back4app.com/digitalocean-vs-linode-vs-heroku/) - LOW confidence

---
*Feature research for: Hostodo CLI Security & Deployment Improvements*
*Researched: 2026-02-15*
*Confidence: HIGH (verified with official docs + industry standards)*
