# Pitfalls Research

**Domain:** CLI Security Hardening & Deployment Features
**Researched:** 2026-02-15
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Device Code Entropy Insufficient for Production

**What goes wrong:**
8-digit numeric device codes (10^8 combinations) can be brute-forced in hours with modern compute. RFC 8628 explicitly warns that device codes "SHOULD use very high entropy" since they're not displayed to users. Attackers can exhaust the keyspace during the code validity window, especially with lax rate limiting.

**Why it happens:**
Developers prioritize UX (shorter codes = less typing for users) over security, forgetting that device codes aren't user-facing—only user codes are typed. Short codes from initial POC/MVP implementations persist into production.

**How to avoid:**
- Use 12+ character alphanumeric codes (36^12 = ~4.7×10^18 combinations)
- Format as dash-separated groups (A3K9-M7P2-X4Q8) for logging/debugging readability
- Implement aggressive rate limiting: RFC 8628 recommends limiting to ~5 attempts for 2^-32 success probability
- Set short validity windows (5-10 minutes max)

**Warning signs:**
- Device code length < 12 characters
- Numeric-only codes (10^n vs 36^n entropy)
- No rate limiting on `/oauth/token` polling endpoint
- Validity period > 15 minutes

**Phase to address:**
Phase 1 (Security Hardening) - Must fix before announcing public CLI availability

**Sources:**
- [RFC 8628 - OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)
- [Security Considerations - OAuth 2.0 Simplified](https://www.oauth.com/oauth2-servers/device-flow/security-considerations/)

---

### Pitfall 2: Hard-Deleted Sessions Create Compliance Gaps

**What goes wrong:**
Hard-deleting OAuth sessions on revocation destroys audit trails. Compliance teams cannot answer "When was this session revoked?", "Who revoked it?", or "What sessions were active during the security incident window?" GDPR/SOC2 audits fail when you can't prove deletion occurred or when.

**Why it happens:**
Developers conflate "user can't use it anymore" with "delete the record." Simple `DELETE` statements are easier than soft-delete infrastructure. Privacy regulations misunderstood as requiring immediate data destruction (they typically require de-identification or anonymization, not deletion).

**How to avoid:**
- Add `revoked_at` timestamp field (nullable)
- Query filters: `WHERE revoked_at IS NULL` for active sessions
- Separate archival process for compliance retention (e.g., move to cold storage after 90 days)
- Document retention policy: "Revoked sessions retained for 90 days for audit, then archived"
- Use event-sourcing pattern: emit `SessionRevoked` event for immutable audit log

**Warning signs:**
- `DELETE FROM cli_sessions WHERE...` in revocation logic
- No `revoked_at`, `deleted_at`, or similar timestamp fields
- Compliance questions about "when was session X revoked" cannot be answered
- Ghost data in reports (forgot WHERE deleted_at IS NULL filter)

**Phase to address:**
Phase 1 (Security Hardening) - Implement before GA; migrating later breaks production data

**Sources:**
- [Soft delete vs hard delete: choose the right data lifecycle | AppMaster](https://appmaster.io/blog/soft-delete-vs-hard-delete)
- [10 Essential Audit Trail Best Practices for 2026 – OpsHub Signal](https://signal.opshub.me/audit-trail-best-practices/)
- [Deleting data: soft, hard or audit? | Marty Friedel](https://www.martyfriedel.com/blog/deleting-data-soft-hard-or-audit)

---

### Pitfall 3: Token Rotation Race Conditions Break Active Sessions

**What goes wrong:**
Implementing refresh token rotation carelessly creates race conditions: client requests refresh while old token is being invalidated, causing "random logout" bugs. Users see 401 errors mid-session when access token expires and refresh fails. Multi-device users experience frequent auth failures when devices race to rotate the same token.

**Why it happens:**
Authorization servers invalidate old tokens before issuing new ones (non-atomic). Clock skew between client and server causes premature token expiration. Developers test single-device flows and miss race conditions that manifest in production multi-device scenarios.

**How to avoid:**
- Atomic rotation: issue new token BEFORE invalidating old one (grace period)
- Implement token families: track lineage, detect reuse across family
- Reuse detection: if stolen token used after rotation, revoke entire family
- Clock skew tolerance: accept tokens ±5 minutes from `exp` claim
- Idempotent refresh: same refresh token can be used twice within 5-second window

**Warning signs:**
- "Random logout" bug reports
- 401 errors clustering around token expiration times
- Refresh endpoint has no reuse detection
- Multiple devices per user showing higher auth failure rates
- Refresh token rotation without token family tracking

**Phase to address:**
Phase 1 (Security Hardening) - Critical for production multi-device support

**Sources:**
- [Hardening OAuth Tokens in API Security | Clutch Events](https://www.clutchevents.co/resources/hardening-oauth-tokens-in-api-security-token-expiry-rotation-and-revocation-best-practices)
- [Refresh Token Rotation - Auth0 Docs](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation)
- [Access Token vs Refresh Token: OAuth 2026 | TheLinuxCode](https://thelinuxcode.com/access-token-vs-refresh-token-a-practical-breakdown-for-modern-oauth-2026/)

---

### Pitfall 4: Hostname Ambiguity Without Resolution Strategy

**What goes wrong:**
Users create multiple instances with same hostname (dev/staging/prod all named "webserver"). CLI command `hostodo get webserver` fails with "multiple instances match" error or worse—silently operates on wrong instance. User deletes production thinking it's staging. SSH hostname conflicts prevent connection.

**Why it happens:**
CLI accepts hostname as argument without uniqueness validation. Database allows duplicate hostnames (no UNIQUE constraint). Developers assume hostnames are unique like IDs. Hostname resolution logic uses `WHERE hostname = ?` instead of handling duplicates.

**How to avoid:**
- Detection: `SELECT COUNT(*) WHERE hostname = ?`; if > 1, enter disambiguation flow
- Interactive mode: show list of matches with distinguishing info (region, IP, status)
  ```
  Multiple instances found named "webserver":
  [1] webserver (192.168.1.10, US-East, running)
  [2] webserver (10.0.0.5, EU-West, stopped)
  Select instance [1-2]:
  ```
- Non-interactive mode: require `--region` or `--id` flag when ambiguous
- Fuzzy matching with confirmation: "Did you mean webserver-prod?"
- Add `display_name` field separate from `hostname` for user-friendly labels
- Validate uniqueness at creation: warn "hostname already exists, append -2?"

**Warning signs:**
- Database schema allows `hostname` duplicates
- Get/start/stop commands use `WHERE hostname = ?` without LIMIT check
- Error handling missing for "multiple results" case
- No disambiguation UI/UX designed
- Test data has unique hostnames (doesn't catch production behavior)

**Phase to address:**
Phase 2 (Hostname Support) - Must solve before shipping hostname features

**Sources:**
- [AmbiSQL: Interactive Ambiguity Detection and Resolution for Text-to-SQL](https://arxiv.org/html/2508.15276)
- [SOLVED - avahi Host name conflict, retrying with hostname-2 | Arch Linux Forums](https://bbs.archlinux.org/viewtopic.php?id=284081)

---

### Pitfall 5: Interactive Prompts Break Automation Scripts

**What goes wrong:**
CLI adds interactive deployment wizard with prompts for region/plan/template. Existing automation scripts (CI/CD, Terraform, cron jobs) hang indefinitely waiting for stdin input that never comes. Scripts that worked pre-update now fail silently or timeout.

**Why it happens:**
Developers test CLI manually in terminal (TTY detected, prompts work). Forget that scripts run without TTY (stdin is pipe/redirect). No `--non-interactive` or `--yes` flag for automation. Prompts don't check `isatty(stdin)` before triggering.

**How to avoid:**
- Detect TTY: `if stdin.IsTerminal()` before showing prompts
- Non-TTY behavior: require flags or error with usage message
  ```
  Error: --region required when running non-interactively
  Use --region, --plan, --template flags or run in interactive terminal
  ```
- Explicit `--non-interactive` flag: skip all prompts, use defaults/required flags
- `--yes` / `-y` flag: auto-confirm dangerous operations (delete, force reboot)
- Environment variable fallback: `HOSTODO_REGION`, `HOSTODO_PLAN` for Docker/CI
- Document automation patterns in README with example scripts

**Warning signs:**
- Prompts without TTY detection
- No `--non-interactive` flag exists
- Deployment command requires interaction (no flag-based mode)
- CI/CD examples missing from documentation
- Test suite doesn't cover non-TTY scenarios

**Phase to address:**
Phase 2 (Deployment Features) - Design flag-based mode alongside interactive wizard

**Sources:**
- [Command Line Interface Guidelines](https://clig.dev/)
- [Interactive CLI Automation with Python | The Green Report](https://www.thegreenreport.blog/articles/interactive-cli-automation-with-python/interactive-cli-automation-with-python.html)

---

### Pitfall 6: Backward Compatibility Breaks Without Deprecation Path

**What goes wrong:**
CLI v2.0 changes command structure (`hostodo instances list` → `hostodo list`). Existing scripts, CI/CD pipelines, and user muscle memory break immediately. Users downgrade to v1.x or abandon CLI entirely. GitHub issues flood with "breaking changes" complaints.

**Why it happens:**
Developers see UX improvements as "obvious wins" and ship without migration plan. SemVer major version (v2.0) used as license to break everything. No cost/benefit analysis of breaking 90% of existing usage. Frustration with legacy command structure drives aggressive refactoring.

**How to avoid:**
- Support BOTH old and new commands for 2+ major versions
- Deprecation warnings: print to stderr when old command used
  ```
  Warning: 'hostodo instances list' is deprecated, use 'hostodo list'
  This alias will be removed in v3.0 (June 2027)
  ```
- Changelog prominently documents breaking changes and migration guide
- `hostodo migrate` command: analyzes scripts, suggests replacements
- Beta flag for new behavior: `--use-new-structure` before becoming default
- Gradual rollout: aliases first, then swap default, then remove (expand/migrate/contract)

**Warning signs:**
- Major version bump (v1.x → v2.0) with command structure changes
- No deprecation warnings in v1.x preparing users
- Documentation only shows new commands (old commands undocumented)
- Migration guide missing or buried
- Breaking changes shipped in minor/patch versions

**Phase to address:**
Phase 2 (CLI Restructuring) - Plan deprecation before shipping new structure

**Sources:**
- [Backward Compatibility: Versioning, Migrations, and Testing | Medium](https://medium.com/@QuarkAndCode/backward-compatibility-versioning-migrations-and-testing-b69637ca5e3d)
- [AWS CLI version 2 Migration Guide](https://docs.aws.amazon.com/cli/latest/userguide/cliv2-migration.html)

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Numeric device codes | Faster initial POC, simpler code generation | Brute-force vulnerability, security incident | **Never** - RFC 8628 explicitly requires high entropy |
| Hard-delete sessions | Simpler DELETE queries, less storage | Compliance failures, no audit trail, regulatory fines | Never in production - only in local dev/testing |
| Single hostname field | Matches VPS hostname exactly | Ambiguity when duplicates exist, poor UX | If enforcing UNIQUE constraint + validation at creation |
| No TTY detection in prompts | Works in manual testing | Breaks all automation, CI/CD failures | Only in pure interactive-only tools (not server CLIs) |
| Breaking changes in v2.0 without aliases | Clean codebase, better structure | User churn, support burden, ecosystem fragmentation | Only if user base < 100 and pre-GA |
| Storing tokens in JSON config | Simple file I/O, portable | Token theft if file permissions wrong | Acceptable with strict `chmod 0600` enforcement |
| Config duplication across files | Localized logic, easy to write | Maintenance nightmare, drift, bugs | Only during rapid prototyping - refactor before v1.0 |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| OAuth backend | Trusting 8-digit codes from POC | Validate entropy: device_code >= 12 chars, user_code follows RFC 8628 |
| Device flow polling | Hardcoded 5-second interval | Respect `interval` from backend, implement exponential backoff for `slow_down` |
| Session API | Fetching all sessions on login | Paginate: fetch only active sessions, lazy-load revoked for audit views |
| Instance API by hostname | `GET /instances?hostname=X` without duplicate check | Check count first, disambiguate if > 1, cache hostname→ID mapping |
| Keychain storage (macOS/Linux) | Assuming keychain available | Fallback to encrypted file if keychain missing, document dependencies |
| SSH key management | Storing private keys in API | Only store public keys server-side, private keys stay client-local |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Fetching all instances on every list | Fast with 5 instances | Paginate with `--limit`/`--offset`, cache locally, lazy-load details | 50+ instances (2+ second load times) |
| Polling for status without backoff | Responsive during development | Exponential backoff: 1s, 2s, 4s, 8s max, respect `Retry-After` header | API rate limits kick in (429 errors) |
| N+1 queries for instance details | Unnoticeable with single instance | Batch API: `GET /instances/batch?ids=1,2,3`, client-side join | List view with 20+ instances |
| Linear search through hostnames | Instant with 10 instances | Build index: map[hostname][]instanceID on first list, O(1) lookup | 100+ instances (noticeable lag) |
| Synchronous power operations | Good UX for single instance | Async with progress bar, parallel operations for bulk commands | Multi-instance operations |
| Full table scan on revoked sessions | Fast with 100 sessions | Index on `revoked_at`, partition by date, archive after 90 days | 10k+ sessions in database |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Device code brute-force window | Attacker obtains auth before user notices | 12+ char codes, 5-minute expiry, rate limit to 5 attempts |
| Token in command-line args | Visible in `ps aux`, shell history | Use stdin, env vars, or config file - never CLI args |
| Config file world-readable | Token theft via local access | Enforce `0600` permissions on write, error if world-readable on read |
| No last_used tracking | Can't detect zombie sessions or token theft | Log `last_used_at` on every API call, alert on geolocation changes |
| Revoked tokens still work | Sessions persist after revocation | Backend validates token NOT in revoked set, client clears on 401 |
| Missing token family tracking | Can't detect token replay attacks | Track refresh token lineage, revoke family on reuse detection |
| Plain-text tokens in logs | Tokens leaked in crash reports, debug logs | Redact tokens: log first 8 chars only (e.g., `abc123...`) |
| No scope limitation | CLI has god-mode access to account | Implement least-privilege scopes: `read:instances`, `write:instances` |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Requiring instance IDs everywhere | Users memorize IDs or copy-paste constantly | Accept hostnames, show hostname in all output, fuzzy search |
| No feedback during long operations | "Is it frozen?" - users Ctrl+C and retry | Progress indicators: "Starting instance (2/30s)", spinner, percentage |
| Error messages without solutions | "Error: 404" - user stuck | Actionable errors: "Instance not found. List instances: hostodo list" |
| Verbose command structure | `hostodo instances list` for 90% of usage | Default to instances: `hostodo list` (instances implied) |
| No confirmation on destructive ops | Accidentally delete production | Require `--confirm` or `--yes` for delete/terminate, show impact |
| Generic "authentication failed" | Can't distinguish wrong password from expired session | Specific: "Session expired, please re-authenticate: hostodo login" |
| Table output truncated in narrow terminals | Data hidden, users can't see important fields | Responsive layouts, horizontal scroll hint, `--simple` for CI |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **OAuth Device Flow:** Coded 8-digit codes — verify RFC 8628 compliance (12+ chars, high entropy)
- [ ] **Session Management:** Working revoke endpoint — verify audit trail (soft-delete, not hard-delete)
- [ ] **Token Rotation:** Refresh tokens rotate — verify race condition handling (atomic swap, token families)
- [ ] **Hostname Support:** Accepts hostname args — verify ambiguity resolution (duplicate detection, disambiguation UI)
- [ ] **Interactive Prompts:** Nice wizard for deployment — verify non-interactive mode (TTY detection, `--non-interactive` flag)
- [ ] **Command Restructuring:** Cleaner `hostodo list` command — verify backward compat (aliases, deprecation warnings)
- [ ] **Config Management:** Loading from file works — verify security (0600 permissions, no world-readable)
- [ ] **Error Handling:** API errors caught — verify actionable messages (next steps, not generic "failed")
- [ ] **Power Operations:** Start/stop/reboot work — verify async handling (progress bars, timeout handling)
- [ ] **Last Used Tracking:** `last_used_at` field exists — verify updated on every auth'd request, not just login

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| 8-digit codes in production | MEDIUM | Migrate: support both 8 and 12-digit, expire old codes after 7 days, force re-auth |
| Hard-deleted sessions | HIGH | Add `revoked_at` field, cannot restore lost data, document gap in audit logs |
| Token rotation race conditions | LOW | Add grace period (5s), implement token families, notify users of forced re-auth |
| Hostname ambiguity without UI | MEDIUM | Add disambiguation, update docs, notify users via changelog + in-app banner |
| Broken automation from prompts | LOW | Ship hotfix with `--non-interactive`, document in release notes, apologize |
| Breaking changes without migration | HIGH | Emergency patch with aliases, extend deprecation timeline, publish migration guide |
| World-readable config file | LOW | Force `chmod 0600` on next write, warn users to check existing permissions |
| No last_used tracking | MEDIUM | Add field, backfill with `created_at` or NULL, implement going forward |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Device code entropy (8-digit) | Phase 1: Security Hardening | Generate 1000 codes, verify all >= 12 chars, alphanumeric, formatted |
| Hard-deleted sessions | Phase 1: Security Hardening | Revoke session, verify `revoked_at` set, record still queryable |
| Token rotation race conditions | Phase 1: Security Hardening | Multi-device stress test, verify no 401s during concurrent refresh |
| Hostname ambiguity | Phase 2: Hostname Support | Create 2 instances with same hostname, verify disambiguation prompt |
| Interactive prompt automation breaks | Phase 2: Deployment Features | Run `hostodo deploy` with stdin=closed, verify error (not hang) |
| Breaking changes without migration | Phase 2: CLI Restructuring | Run deprecated command, verify warning + still works |
| No last_used tracking | Phase 1: Security Hardening | Auth, make API calls, verify `last_used_at` updates in DB |
| Config duplication | Phase 2: CLI Restructuring | Refactor config loading, verify DRY principle, single source of truth |
| Missing TTY detection | Phase 2: Deployment Features | Test in Docker (no TTY), verify falls back to flag mode |
| No token scope limitation | Future (OAuth Scopes) | Request read-only token, verify write operations rejected |

## Sources

### OAuth & Security
- [RFC 8628 - OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)
- [Security Considerations - OAuth 2.0 Simplified](https://www.oauth.com/oauth2-servers/device-flow/security-considerations/)
- [Hardening OAuth Tokens in API Security](https://www.clutchevents.co/resources/hardening-oauth-tokens-in-api-security-token-expiry-rotation-and-revocation-best-practices)
- [Refresh Token Rotation - Auth0 Docs](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation)
- [OAuth 2.1 vs 2.0: What developers need to know](https://stytch.com/blog/oauth-2-1-vs-2-0/)

### Audit & Compliance
- [Soft delete vs hard delete | AppMaster](https://appmaster.io/blog/soft-delete-vs-hard-delete)
- [Deleting data: soft, hard or audit?](https://www.martyfriedel.com/blog/deleting-data-soft-hard-or-audit)
- [10 Essential Audit Trail Best Practices for 2026](https://signal.opshub.me/audit-trail-best-practices/)

### CLI Best Practices
- [Command Line Interface Guidelines](https://clig.dev/)
- [Interactive CLI Automation with Python](https://www.thegreenreport.blog/articles/interactive-cli-automation-with-python/interactive-cli-automation-with-python.html)

### Backward Compatibility
- [Backward Compatibility: Versioning, Migrations, and Testing](https://medium.com/@QuarkAndCode/backward-compatibility-versioning-migrations-and-testing-b69637ca5e3d)
- [AWS CLI version 2 Migration Guide](https://docs.aws.amazon.com/cli/latest/userguide/cliv2-migration.html)

### Hostname Resolution
- [AmbiSQL: Interactive Ambiguity Detection and Resolution](https://arxiv.org/html/2508.15276)
- [SOLVED - avahi Host name conflict](https://bbs.archlinux.org/viewtopic.php?id=284081)

---
*Pitfalls research for: Hostodo CLI v2.0 Security & Deployment*
*Researched: 2026-02-15*
