# Architecture Research

**Domain:** CLI tool enhancements (security, hostname support, deployment)
**Researched:** 2026-02-15
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Layer (Cobra)                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │  auth   │  │instances│  │ deploy  │  │  root   │        │
│  │ (cmds)  │  │ (cmds)  │  │ (cmds)  │  │ aliases │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │              │
├───────┴────────────┴────────────┴────────────┴──────────────┤
│                     Service Layer (pkg/)                     │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ api/client   │  │auth/keychain │  │  ui/styles   │      │
│  │ api/sessions │  │auth/oauth    │  │  ui/table    │      │
│  │api/instances │  │auth/resolver │  │ui/formatters │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
├─────────────────────────────────────────────────────────────┤
│                    External Systems                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Backend    │  │   Keychain   │  │  Terminal    │      │
│  │  REST API    │  │ (OS secure)  │  │   (TTY/UI)   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **cmd/*** | CLI command definitions, argument parsing, user interaction | Cobra commands with RunE functions |
| **pkg/api** | HTTP client, REST API communication, response parsing | net/http client with methods per resource |
| **pkg/auth** | OAuth device flow, token storage (keychain), authentication state | Keychain integration + OAuth polling |
| **pkg/config** | Config file management, device ID persistence | JSON serialization to ~/.hostodo/config.json |
| **pkg/ui** | Output formatting, interactive TUI, tables, styles | Bubble Tea + lipgloss for rich terminal UI |
| **Backend** | Django REST API, OAuth endpoints, instance management | Existing odopanel backend (not modified) |

## Recommended Project Structure

Current structure is solid, changes are additive:

```
cmd/
├── auth/              # OAuth commands (login, logout, sessions, whoami)
├── instances/         # Instance operations (existing: list, get, start, stop, reboot)
│   └── deploy.go      # NEW: Interactive deployment command
├── root.go            # Root command + top-level aliases
└── aliases.go         # NEW: Additional command aliases (list, start, stop at root)

pkg/
├── api/               # API client layer
│   ├── client.go      # HTTP client base
│   ├── auth.go        # Authentication endpoints
│   ├── instances.go   # Instance CRUD operations
│   ├── sessions.go    # CLI session management
│   ├── models.go      # Response/request structs
│   └── resolver.go    # NEW: Hostname → Instance ID resolution
├── auth/              # Authentication logic
│   ├── keychain.go    # Secure token storage (OS keychain)
│   ├── oauth.go       # Device flow implementation
│   └── device.go      # NEW: Persistent device ID tracking
├── config/            # Configuration management
│   └── config.go      # Config file load/save, device ID persistence
├── ui/                # User interface
│   ├── formatters.go  # Data formatting (JSON, text)
│   ├── styles.go      # Terminal color/style definitions
│   ├── table.go       # Interactive table (Bubble Tea TUI)
│   └── prompts.go     # NEW: Interactive deployment prompts
└── utils/             # Shared utilities
    └── validation.go  # Input validation helpers
```

### Structure Rationale

- **cmd/ organized by domain**: Commands grouped by functional area (auth, instances), not verb (get, list). This scales better as more features are added.
- **pkg/ organized by technical concern**: API client separate from auth logic separate from UI, enabling clean testing and reuse.
- **Minimal new files**: Most changes enhance existing files (instances.go, models.go) rather than creating new structure.
- **Backend boundary clear**: All backend communication flows through pkg/api, making it easy to mock for testing.

## Architectural Patterns

### Pattern 1: Hostname Resolution with Client-Side Caching

**What:** CLI accepts hostname arguments, resolves to instance_id via API lookup, caches result per session.

**When to use:** Any command accepting instance identifier (get, start, stop, reboot, deploy).

**Trade-offs:**
- **Pro:** UX improvement (users think in hostnames, not IDs)
- **Pro:** No backend changes needed (API already returns hostname field)
- **Con:** Extra API call on first use (mitigated by session caching)
- **Con:** Stale cache if hostname changes (acceptable; rare operation)

**Example:**
```go
// pkg/api/resolver.go
type InstanceResolver struct {
    client *Client
    cache  map[string]string // hostname → instance_id
}

func (r *InstanceResolver) Resolve(identifier string) (string, error) {
    // If already looks like instance_id, return as-is
    if isInstanceID(identifier) {
        return identifier, nil
    }

    // Check cache
    if instanceID, ok := r.cache[identifier]; ok {
        return instanceID, nil
    }

    // Fetch all instances and build cache
    instances, err := r.client.ListInstances(100, 0)
    if err != nil {
        return "", err
    }

    for _, inst := range instances.Results {
        r.cache[inst.Hostname] = inst.InstanceID
    }

    // Lookup hostname
    if instanceID, ok := r.cache[identifier]; ok {
        return instanceID, nil
    }

    return "", fmt.Errorf("instance not found: %s", identifier)
}
```

### Pattern 2: Dual-Mode Commands (Interactive + Flags)

**What:** Commands support both interactive prompts (for humans) and flag-based input (for automation), with TTY detection to choose appropriate mode.

**When to use:** Complex operations with multiple inputs (deploy, configure).

**Trade-offs:**
- **Pro:** Great UX for interactive users (guided prompts)
- **Pro:** Automation-friendly (all inputs via flags, no TTY needed)
- **Pro:** Self-documenting (prompts teach flag names)
- **Con:** More code to maintain (two input paths)
- **Con:** Flag validation must match prompt validation

**Example:**
```go
// cmd/instances/deploy.go
var deployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Deploy a new VPS instance",
    Long: `Deploy a new instance interactively or via flags.

Interactive mode (default):
  hostodo deploy

Flag mode (for automation):
  hostodo deploy --hostname myserver --plan vps-1gb --template ubuntu-22.04 --region us-east`,
    RunE: runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
    // Detect if stdin is a terminal
    isTTY := isatty.IsTerminal(os.Stdin.Fd())

    // If all flags provided, use flag mode
    if allFlagsProvided() {
        return deployFromFlags()
    }

    // If TTY, use interactive mode
    if isTTY {
        return deployInteractive()
    }

    // Non-TTY without all flags = error
    return fmt.Errorf("deployment requires either interactive mode (TTY) or all flags: --hostname, --plan, --template, --region")
}
```

### Pattern 3: Backward-Compatible Command Aliases

**What:** Add root-level aliases (e.g., `hostodo list`) while preserving old paths (`hostodo instances list`).

**When to use:** Simplifying common command paths without breaking existing scripts.

**Trade-offs:**
- **Pro:** Reduces typing for common operations
- **Pro:** Zero breaking changes (old commands still work)
- **Con:** Two ways to do same thing (documentation overhead)
- **Con:** Help text needs to clarify alias relationships

**Example:**
```go
// cmd/root.go
func init() {
    // Original commands under namespaces
    rootCmd.AddCommand(auth.AuthCmd)      // hostodo auth login
    rootCmd.AddCommand(instances.InstancesCmd) // hostodo instances list

    // Top-level aliases for auth (already implemented)
    rootCmd.AddCommand(loginAliasCmd)     // hostodo login
    rootCmd.AddCommand(logoutAliasCmd)    // hostodo logout
    rootCmd.AddCommand(whoamiAliasCmd)    // hostodo whoami

    // NEW: Top-level aliases for instances (default scope)
    rootCmd.AddCommand(createListAlias())  // hostodo list → instances list
    rootCmd.AddCommand(createGetAlias())   // hostodo get <id> → instances get <id>
    rootCmd.AddCommand(createStartAlias()) // hostodo start <id> → instances start <id>
    rootCmd.AddCommand(createStopAlias())  // hostodo stop <id> → instances stop <id>
    rootCmd.AddCommand(createRebootAlias())// hostodo reboot <id> → instances reboot <id>
    rootCmd.AddCommand(createDeployAlias())// hostodo deploy → instances deploy
}

func createListAlias() *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List instances (alias for 'instances list')",
        Run: func(cmd *cobra.Command, args []string) {
            // Delegate to instances.listCmd
            instances.ListCmd.Run(cmd, args)
        },
    }
}
```

## Data Flow

### Request Flow: Hostname-Based Instance Operation

```
User Input: "hostodo start myserver"
    ↓
[CLI Parser] → Parse command + arguments
    ↓
[Resolver] → Check if "myserver" is instance_id or hostname
    ↓ (not instance_id format)
[Resolver] → GET /client/instances/?limit=100 (fetch all instances)
    ↓
[Resolver] → Build hostname→instance_id cache map
    ↓
[Resolver] → Lookup "myserver" in cache → returns "abc123"
    ↓
[Client] → POST /client/instances/abc123/power/ {"action": "start"}
    ↓
[Response] → Success message displayed to user
```

### State Management: OAuth Device Flow with Session Tracking

```
[Login Command]
    ↓
[OAuth Client] → POST /v1/oauth/device/authorize
                 {device_name: "MacBook-Pro", device_id: "550e8400-..."}
    ↓
[Backend] → Generate device_code (48 chars, high entropy)
            Generate user_code (12 chars, A3K9-M7P2-X4Q8 format)
            Create DeviceCode record with device_id
    ↓
[CLI] → Display user_code + verification URL
        Poll POST /v1/oauth/token every 5 seconds
    ↓
[User] → Authorizes in browser
    ↓
[Backend] → Create CLISession record
            Link device_id to user
            Return access_token
    ↓
[CLI] → Store token in OS keychain
        Store device_id in ~/.hostodo/config.json
    ↓
[Subsequent Commands]
    ↓
[API Client] → GET /client/instances/
               Header: Authorization: Bearer <token>
    ↓
[Backend] → Validate token
            Update CLISession.last_used_at (NEW)
    ↓
[Response] → Instance data returned
```

### Key Data Flows

1. **Authentication Flow:** Device ID persists across logins, user_code visible to user, device_code hidden and high entropy, token stored in keychain, session tracks usage.

2. **Hostname Resolution Flow:** First instance command fetches all instances, caches hostname→ID mapping, subsequent commands use cached mapping, cache invalidates on error (instance not found).

3. **Deployment Flow (Interactive):** TTY detected → show wizard prompts → collect inputs (hostname, plan, template, region, SSH key) → POST /client/orders/ → poll order status → display success/failure.

4. **Deployment Flow (Flags):** All flags provided → validate inputs → POST /client/orders/ → poll order status → display success/failure (suitable for CI/CD).

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1-100 instances | Current architecture sufficient; single ListInstances call builds cache |
| 100-1000 instances | Consider pagination in resolver; add `--force-refresh` flag to bypass cache |
| 1000+ instances | Backend should add `/client/instances/by-hostname/<hostname>/` endpoint to avoid client-side filtering |

### Scaling Priorities

1. **First bottleneck:** Hostname resolution with 100+ instances (mitigated by caching, acceptable latency).
2. **Second bottleneck:** Interactive TUI rendering with 500+ instances (mitigated by Bubble Tea's virtual scrolling).
3. **Third bottleneck:** Keychain operations on Windows (platform-specific, not critical).

## Anti-Patterns

### Anti-Pattern 1: Storing Tokens in Config File

**What people do:** Store access_token in ~/.hostodo/config.json alongside device_id.

**Why it's wrong:**
- Config files are often committed to dotfiles repos
- Tokens exposed to any process reading home directory
- Violates OAuth security best practices

**Do this instead:**
- Use OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Only store non-sensitive data in config (API URL, device_id)
- Already implemented via pkg/auth/keychain.go

### Anti-Pattern 2: Backend-Dependent Hostname Resolution

**What people do:** Add `/instances/<hostname>/` endpoint to backend, require backend deployment for CLI feature.

**Why it's wrong:**
- Couples CLI feature to backend deployment schedule
- Adds endpoint used only by CLI (not web UI)
- Increases backend attack surface

**Do this instead:**
- Client-side resolution using existing `/instances/` list endpoint
- Hostname field already in API response (no backend change needed)
- Cache results to minimize API calls

### Anti-Pattern 3: Complex Flag Parsing for Interactive Mode

**What people do:** Parse flags even in interactive mode, merge flag values with prompt values, handle conflicts.

**Why it's wrong:**
- Confusing UX (what takes precedence?)
- Brittle logic (hard to test all combinations)
- Violates principle of least surprise

**Do this instead:**
- Detect TTY at command start
- If TTY + no flags → full interactive mode
- If TTY + all flags → skip prompts, use flags
- If no TTY + missing flags → error with helpful message
- Keep input paths completely separate

### Anti-Pattern 4: Generating New Device ID on Every Login

**What people do:** Generate UUID on each login attempt, don't persist device_id.

**Why it's wrong:**
- Backend sees each login as new device
- Session list cluttered with duplicate entries
- Can't track device-specific audit trail

**Do this instead:**
- Generate device_id once, persist in config.json
- Reuse device_id for all login attempts from same machine
- Already implemented via config.GetOrCreateDeviceID()

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Backend REST API | HTTP client with Bearer token auth | Uses pkg/api/client.go doRequest() |
| OS Keychain | Platform-specific keychain libraries | Uses github.com/zalando/go-keyring |
| Terminal (TTY) | isatty detection + Bubble Tea TUI | Interactive mode only if stdin is terminal |
| Config Storage | JSON file at ~/.hostodo/config.json | Stores device_id, api_url (not tokens) |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| cmd/ ↔ pkg/api | Direct function calls | Commands create Client, call methods |
| cmd/ ↔ pkg/auth | Direct function calls | Commands check IsAuthenticated(), GetToken() |
| cmd/ ↔ pkg/ui | Direct function calls | Commands pass data to formatters |
| pkg/api ↔ Backend | HTTP REST | All API calls centralized in pkg/api/* |
| pkg/auth ↔ Keychain | Library calls | Platform-specific keyring operations |

## Component Boundaries: Backend vs CLI Changes

### Backend Changes (Django/odopanel)

**Required (Security Hardening):**
1. **Device Code Format** (odoauth/services/DeviceFlowService.py):
   - Change from 8-digit numeric to 12-char alphanumeric
   - Format: `secrets.token_urlsafe(9)[:12].upper()` → "A3K9M7P2X4Q8"
   - Insert dashes for display: "A3K9-M7P2-X4Q8"
   - Strip dashes on validation

2. **Last Used Tracking** (odoauth/models.py + views.py):
   - Add `CLISession.last_used_at` datetime field
   - Update on every authenticated API call via middleware or view mixin
   - Migration: Add nullable field, backfill with `created_at`

3. **Soft Delete Sessions** (odoauth/models.py + serializers.py):
   - Add `CLISession.revoked_at` nullable datetime field
   - Filter queries to exclude revoked sessions (WHERE revoked_at IS NULL)
   - Revoke endpoint sets `revoked_at = now()` instead of DELETE

**Not Required (CLI Handles):**
- Hostname lookup endpoint (CLI uses existing list endpoint)
- Deployment workflow (CLI uses existing POST /client/orders/)
- Command aliasing (pure CLI feature)

### CLI Changes (Go/Cobra)

**New Files:**
1. **pkg/api/resolver.go**: Hostname → instance_id resolution with caching
2. **pkg/ui/prompts.go**: Interactive deployment wizard (Bubble Tea)
3. **cmd/instances/deploy.go**: Deployment command (interactive + flags)

**Modified Files:**
1. **cmd/root.go**: Add root-level instance command aliases
2. **cmd/instances/get.go**: Accept hostname via resolver
3. **cmd/instances/start.go**: Accept hostname via resolver
4. **cmd/instances/stop.go**: Accept hostname via resolver
5. **cmd/instances/reboot.go**: Accept hostname via resolver
6. **pkg/api/models.go**: Potentially add deployment request/response structs
7. **pkg/auth/oauth.go**: Handle new device_code format (12 chars with dashes)

**No Changes Needed:**
- pkg/auth/keychain.go (already secure)
- pkg/config/config.go (device_id already persisted)
- pkg/api/client.go (bearer auth already implemented)

## Build Order and Dependencies

### Phase 1: Backend Security Hardening (Independent)
**Dependencies:** None (pure backend work)
**Components:**
1. Device code entropy increase (odoauth)
2. Last used tracking (odoauth)
3. Soft delete sessions (odoauth)

**Why First:** Security fixes should ship ASAP, not blocked on CLI features.

### Phase 2: CLI Hostname Support (Depends on Backend Data)
**Dependencies:** Backend already returns `hostname` field in Instance model
**Components:**
1. Create pkg/api/resolver.go
2. Modify instance commands to use resolver
3. Update help text to mention hostname support
4. Add tests for resolver

**Why Second:** Core UX improvement, unblocks deployment command.

### Phase 3: CLI Command Aliases (Depends on Phase 2)
**Dependencies:** Resolver must exist for `hostodo get <hostname>` to work
**Components:**
1. Add root-level aliases in cmd/root.go
2. Update README examples
3. Update help text to show both paths

**Why Third:** Purely additive, doesn't block other work.

### Phase 4: Deployment Command (Depends on Phase 2)
**Dependencies:** Resolver needed for instance-based operations post-deploy
**Components:**
1. Create pkg/ui/prompts.go (interactive wizard)
2. Create cmd/instances/deploy.go
3. Add deployment models to pkg/api/models.go
4. Add deployment methods to pkg/api/client.go

**Why Fourth:** Most complex feature, benefits from all previous work.

## Migration Strategy

### Device Code Format Migration

**Problem:** Existing users may have 8-digit codes in-flight during deployment.

**Solution:**
1. Backend validates both old (8-digit) and new (12-char) formats for 30 days
2. After 30 days, backend rejects old format
3. CLI doesn't change (just accepts what backend returns)

**Code:**
```python
# odoauth/services/DeviceFlowService.py
def validate_user_code(self, user_code):
    # Strip dashes
    clean_code = user_code.replace('-', '')

    # Try new format first (12 alphanumeric)
    if len(clean_code) == 12 and clean_code.isalnum():
        return self.lookup_by_user_code(clean_code)

    # TEMPORARY: Support old format (8 digits) until 2026-03-15
    if settings.ALLOW_LEGACY_USER_CODES and len(clean_code) == 8 and clean_code.isdigit():
        return self.lookup_by_user_code(clean_code)

    raise InvalidUserCodeError()
```

### Session Tracking Migration

**Problem:** Existing CLISession records don't have `last_used_at`.

**Solution:**
1. Migration adds nullable field, backfills with `created_at`
2. All active sessions get "last used" data immediately
3. New code updates field on every API call
4. No user-facing impact

### Backward Compatibility for Commands

**Problem:** Users have scripts using `hostodo instances list`.

**Solution:**
1. Keep all original command paths working
2. Add aliases at root level
3. Documentation shows both (prefer shorter form)
4. No deprecation warnings (both are equally valid)

**Example:**
```bash
# Both work indefinitely
hostodo instances list
hostodo list

# Both work indefinitely
hostodo instances get myserver
hostodo get myserver
```

## Sources

- [Cobra Framework Documentation](https://cobra.dev/)
- [RFC 8628: OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)
- [Command Line Interface Guidelines](https://clig.dev/)
- [Go Keychain Package](https://github.com/keybase/go-keychain)
- [SCS Session Manager for Go](https://github.com/alexedwards/scs)
- [Building CLI Applications in Go using Cobra](https://oneuptime.com/blog/post/2026-01-07-go-cobra-cli/view)
- [10 Design Principles for Delightful CLIs](https://www.atlassian.com/blog/it-teams/10-design-principles-for-delightful-clis)

---
*Architecture research for: Hostodo CLI v2.0 enhancements*
*Researched: 2026-02-15*
