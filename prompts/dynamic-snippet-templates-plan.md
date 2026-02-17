# Dynamic Snippet Templates - Implementation Plan (v2)

## Issue Reference
GitHub Issue #20: Variable/Placeholder Templates for Dynamic Snippets

---

## 1. Context

### Original Request
Enhance SnipGo with variable/placeholder template support, inspired by pet CLI and chezmoi, enabling snippets to contain dynamic placeholders that are resolved at execution/copy time.

### Current State Summary
- **Snippet struct** (`/Users/abel/github/snipgo/internal/core/snippet.go`): 9 fields (id, title, description, tags, language, is_favorite, created_at, updated_at, body). Body excluded from YAML frontmatter via `yaml:"-"`.
- **Frontmatter** (`/Users/abel/github/snipgo/internal/core/frontmatter.go`): Simple `---\nYAML\n---\n\nBody` parsing. No template awareness.
- **Manager** (`/Users/abel/github/snipgo/internal/core/manager.go`): In-memory cache (`map[string]*Snippet`), thread-safe via `sync.RWMutex`. Returns deep copies. No template expansion.
- **CLI exec** (`/Users/abel/github/snipgo/cmd/snipgo/exec.go`): Runs raw `sh -c snippet.Body` -- no variable substitution.
- **CLI copy** (`/Users/abel/github/snipgo/cmd/snipgo/copy.go`): Copies raw body to clipboard -- no variable substitution.
- **CLI search** (`/Users/abel/github/snipgo/cmd/snipgo/search.go`): Outputs raw body to stdout after fzf selection.
- **Wails bridge** (`/Users/abel/github/snipgo/app/app.go`): 14 methods exposed. No template-related methods.
- **Frontend** (`/Users/abel/github/snipgo/frontend/src/`): React + CodeMirror editor. `SnippetEditor.tsx` with auto-save, tag management, code editing. No variable UI.
- **Config** (`/Users/abel/github/snipgo/internal/config/config.go`): Simple `Config` struct with `DataDirectory` field only.
- **NO existing template/variable code anywhere in the codebase.**
- **Existing frontmatter tests**: 16 test cases total (11 ParseFrontmatter + 5 SerializeFrontmatter) in `frontmatter_test.go`.

### Research Findings
- **Pet CLI**: Uses `<param=default>` syntax with pipe-delimited choices `<subject=|John||Sam||Jane Doe|>`. Has `--raw` flag to skip expansion.
- **Chezmoi**: Uses Go `text/template` with context variables. Has preview-before-apply pattern.
- **SnipGo Issue #20 Requirements**: `${VARIABLE_NAME}` syntax, GUI modal prompts, CLI `-v KEY=VALUE` flags, variable history, frontmatter metadata.

---

## 2. Work Objectives

### Core Objective
Add a template engine to SnipGo that detects `${VARIABLE_NAME}` placeholders in snippet bodies, prompts users for values (CLI and GUI), and produces expanded output -- while maintaining full backward compatibility with existing snippets.

### Deliverables
1. **Template Engine** -- Go package for parsing, extracting, and expanding variables in snippet bodies
2. **Frontmatter Variable Metadata** -- Optional `variables` field in YAML frontmatter for descriptions, defaults, and choices
3. **CLI Variable Support** -- Interactive prompts + `-v KEY=VALUE` flag + `--raw` flag for all output commands
4. **GUI Variable Modal** -- React modal that appears when copying a templated snippet ("Copy Expanded")
5. **Variable History** -- Persistence of recently used values per variable name
6. **Preview Mode** -- Show expanded snippet before execution (chezmoi-inspired); DEFAULT for exec when variables present

### Definition of Done
- All existing tests pass (zero regressions)
- New template engine has 90%+ test coverage
- Existing snippets without variables work identically to before
- CLI: `exec`, `copy`, `search` commands support variable expansion
- GUI: "Copy Expanded" action supports variable expansion via modal
- Variable history persists across sessions
- Frontmatter with `variables:` field parses and serializes correctly

---

## 3. Design Decisions

### D1: Variable Syntax -- `${VARIABLE_NAME}`
**Choice:** `${VARIABLE_NAME}` (as specified in issue #20)
**Rationale:**
- Familiar to shell users and developers
- Distinct from Go template `{{.Var}}` syntax (avoids confusion with chezmoi)
- Distinct from pet's `<param>` syntax (avoids conflict with shell redirects in snippet bodies)
- Regex: `\$\{([A-Za-z_][A-Za-z0-9_]*)\}`

**Escape mechanism:** `$${VAR}` produces literal `${VAR}` (double-dollar escapes). This matches shell conventions.

### D2: Frontmatter Variable Metadata -- Optional `variables` Map
**Choice:** Add optional `variables` map to YAML frontmatter.
```yaml
---
id: abc123
title: Deploy Script
variables:
  SERVER:
    description: "Target server hostname"
    default: "prod-01.example.com"
  ENV:
    description: "Deployment environment"
    default: "staging"
    choices:
      - staging
      - production
      - development
---

ssh ${SERVER} "deploy --env ${ENV}"
```

**Rationale:**
- Backward compatible -- existing snippets have no `variables` field, YAML parser ignores unknown fields
- Aligns with pet's concept of defaults and choices
- Keeps metadata close to usage (in the same file)
- `choices` field inspired by pet's multiple-choice syntax but using cleaner YAML lists

### D3: Storage for Variable History -- JSON File
**Choice:** Store variable history in `~/.config/snipgo/var_history.json`
```json
{
  "SERVER": ["prod-01.example.com", "staging-01.example.com"],
  "ENV": ["staging", "production"],
  "_meta": { "max_entries_per_var": 10 }
}
```
**Rationale:**
- Simple, human-readable
- Separate from snippet files (variable values may be sensitive)
- Same config directory as existing config.yaml
- Capped at 10 entries per variable to prevent unbounded growth

### D4: Template Engine Location -- New `internal/tmpl/` Package
**Choice:** Create `internal/tmpl/` package (NOT `template` -- avoids shadowing Go stdlib `text/template`)
**Rationale:**
- Single responsibility -- template logic is distinct from snippet management
- Testable in isolation without storage/manager dependencies
- Can be imported by both CLI and Wails bridge independently
- Name `tmpl` avoids Go stdlib shadow warning

### D5: CLI Flag Design
**Choice:** Add flags to `exec`, `copy`, and `search` commands:
- `-v KEY=VALUE` (repeatable) -- provide variable values inline
- `--raw` -- output unexpanded snippet body (pet-inspired)
- `--preview` -- show expanded snippet before executing (exec only). **DEFAULT when variables are present** for safety (see D8)

### D6: GUI Interaction Model
**Choice:** Modal dialog triggered when user clicks "Copy Expanded" on a snippet containing variables.
- Modal shows one input per detected variable
- Pre-fills with: (1) frontmatter default, (2) last-used value from history, (3) empty
- Dropdown for variables with `choices` defined
- "Cancel" returns to editor; "Apply" expands and copies to clipboard
- **Phase 2 scope:** "Copy Expanded" button only. GUI "Execute" button deferred to future work.

### D7: Snippet Struct Extension
**Choice:** Add `Variables` field to `Snippet` struct with `yaml:"variables,omitempty"`.
**Rationale:** `omitempty` ensures existing snippets serialize identically. The field is only present when variables are explicitly defined.

### D8: Shell Injection Safety for `exec` with Variables (CRITICAL)
**Choice:** When `exec` runs a snippet that contains variables, `--preview` is ON by default.
**Rationale:**
- `exec.go` runs `sh -c expanded_body`. If a user provides `; rm -rf /` as a variable value, it executes.
- Forcing preview-before-execute when variables are present gives the user a chance to inspect the expanded command.
- User can bypass with `--no-preview` if they trust the values (e.g., from `-v` flags in a script).
- A warning is printed to stderr: `WARNING: Variable values are interpolated directly into the shell command. Review the expanded command carefully.`

### D9: Variable Name Hydration on YAML Unmarshal
**Choice:** The `Variable` struct has `Name` field tagged `yaml:"-" json:"name"`.
**Rationale:**
- YAML frontmatter uses map keys as variable names (`variables:\n  SERVER:\n    description: ...`)
- After `yaml.Unmarshal`, `map[string]*Variable` has the name as map key but `Variable.Name` is empty
- Post-unmarshal hydration in `ParseFrontmatter` sets `v.Name = key` for each entry
- This is 3 lines of code and keeps the struct self-describing for JSON serialization to frontend

### D10: ExtractVariables Merge Behavior (Body is Source of Truth)
**Choice:** When merging body-detected variables with frontmatter metadata:
1. **Variables in body but NOT in frontmatter** -- returned with empty metadata (name only, no description/default/choices)
2. **Variables in frontmatter but NOT in body** -- EXCLUDED (body is source of truth for what variables are needed)
3. **Variables in both** -- frontmatter enriches the body-detected variable with description, default, choices
4. **Ordering** -- body occurrence order (first appearance), matching `ExtractVariables` output

**Rationale:** The body is what gets expanded, so only variables actually referenced in the body matter. Frontmatter metadata is supplementary. This prevents stale frontmatter entries from causing phantom prompts.

### D11: Choice Prompt UX in CLI
**Choice:** Simple numbered list with `fmt.Scanln` for choice-based variables:
```
SERVER (Target server) [prod-01]:
  1) prod-01 (default)
  2) staging-01
  3) dev-01
Enter choice [1] or custom value:
```
**Rationale:**
- No external dependency needed (readline is already available but not required for this)
- User can type a number to select, or type a custom value
- Default is pre-selected (just press Enter)
- Clear, familiar UX pattern

---

## 4. Must Have / Must NOT Have

### Must Have (MVP - Phase 1)
- [ ] `${VAR}` detection and extraction from snippet body
- [ ] Variable expansion with provided values
- [ ] Escape mechanism (`$${}`)
- [ ] CLI interactive prompts for missing variables (with choice display per D11)
- [ ] CLI `-v KEY=VALUE` flag
- [ ] CLI `--raw` flag
- [ ] CLI `--preview` default-on for exec when variables present (D8)
- [ ] Shell injection warning on stderr for exec with variables (D8)
- [ ] Frontmatter `variables` field with `description`, `default`, `choices`
- [ ] Variable.Name hydration after YAML unmarshal (D9)
- [ ] ExtractVariables merge behavior per D10
- [ ] Backward compatibility with all existing snippets
- [ ] Template engine unit tests (90%+ coverage)

### Must Have (Phase 2)
- [ ] GUI variable prompt modal ("Copy Expanded" action)
- [ ] Variable history (read/write)
- [ ] History-based auto-suggestions in GUI modal
- [ ] Wails bridge methods for template operations

### Must NOT Have (Out of Scope)
- Go `text/template` syntax (too complex, different paradigm)
- Shell parameter expansion (`${VAR:-default}`, `${VAR:+alt}`) -- keep syntax simple
- Nested variable references (`${${PREFIX}_SERVER}`)
- Remote sync of variable history
- Variable validation/type constraints beyond string
- Encrypted variable storage
- Environment variable auto-injection (security concern)
- Conditional logic in templates
- GUI "Execute" button (deferred to future -- "Copy Expanded" only in Phase 2)

---

## 5. Task Flow and Dependencies

```
Phase 1: Core Engine (no external dependencies)
  T1: Template types and interfaces               [no deps]
  T2: Variable parser (extract ${VAR} from body)  [depends: T1]
  T3: Variable expander (substitute values)        [depends: T1]
  T4: Escape handling ($${VAR} -> literal)          [depends: T2, T3]
  T5: Snippet struct extension (Variables field)    [depends: T1]
  T6: Frontmatter parse/serialize for variables     [depends: T5]
       + Variable.Name hydration (D9)
       + choices support in parse/serialize
  T7: CLI -v flag, --raw flag, --preview flag       [depends: T2, T3]
       + Shell injection warning (D8)
       + --preview default-on for exec with vars
       + Choice prompt display (D11)
  T8: CLI interactive prompts for variables         [depends: T2, T7]
       + ExtractVariables merge logic (D10)

Phase 2: Persistence and GUI
  T9:  Variable history store                       [no deps]
  T10: Wire history into CLI prompts                [depends: T8, T9]
  T11: Wails bridge methods for templates           [depends: T3, T5, T9]
       + ExtractVariables with merge behavior (D10)
  T12: GUI VariablePromptModal component            [depends: T11]
       + Dropdown for choices
  T13: Wire modal into SnippetEditor ("Copy Expanded") [depends: T12]
  T14: CLI Integration tests (end-to-end)           [depends: T8, T10]
  T15: Frontend tests for modal                     [depends: T12]
```

---

## 6. Detailed TODOs

### Phase 1: Core Template Engine and CLI Integration

#### T1: Define Template Types and Interfaces
**Files to create:**
- `/Users/abel/github/snipgo/internal/tmpl/types.go`

**What:**
```go
package tmpl

// Variable represents a template variable definition.
// Name is populated via post-unmarshal hydration (see D9), not from YAML.
type Variable struct {
    Name        string   `yaml:"-" json:"name"`
    Description string   `yaml:"description,omitempty" json:"description,omitempty"`
    Default     string   `yaml:"default,omitempty" json:"default,omitempty"`
    Choices     []string `yaml:"choices,omitempty" json:"choices,omitempty"`
}

// TemplateResult holds the expansion output.
type TemplateResult struct {
    Expanded  string            // The expanded body text
    Variables map[string]string // Variable name -> resolved value
}
```

**IMPORTANT:** No `VariableValue` struct -- it is not needed. Values are passed as `map[string]string`.

**Acceptance Criteria:**
- Types compile without errors
- `Variable.Name` has `yaml:"-"` tag (not populated from YAML, hydrated post-unmarshal)
- `Variable.Name` has `json:"name"` tag (serialized to frontend)
- JSON and YAML tags present for all other fields
- Zero external dependencies in this file
- No `VariableValue` type exists

---

#### T2: Variable Parser -- Extract Variables from Body
**Files to create:**
- `/Users/abel/github/snipgo/internal/tmpl/parser.go`
- `/Users/abel/github/snipgo/internal/tmpl/parser_test.go` (WRITE FIRST - TDD)

**What:**
- `ExtractVariables(body string) []string` -- returns unique variable names in order of first appearance
- Regex: `\$\{([A-Za-z_][A-Za-z0-9_]*)\}`
- Must skip escaped `$${VAR}` patterns
- Must handle duplicate variable names (return unique list)
- Must preserve order of first appearance

**Test Cases (write first):**
| Input | Expected Output |
|-------|----------------|
| `"hello ${NAME}"` | `["NAME"]` |
| `"${A} and ${B} and ${A}"` | `["A", "B"]` |
| `"no variables here"` | `[]` |
| `""` | `[]` |
| `"$${ESCAPED}"` | `[]` |
| `"${valid_name_123}"` | `["valid_name_123"]` |
| `"${123invalid}"` | `[]` |
| `"${A}\n${B}\n${C}"` | `["A", "B", "C"]` |
| `"mix ${REAL} and $${ESCAPED} and ${ALSO_REAL}"` | `["REAL", "ALSO_REAL"]` |

**Acceptance Criteria:**
- All test cases pass
- Returns empty slice (not nil) for no variables
- Order preserved, duplicates removed

---

#### T3: Variable Expander -- Substitute Values into Body
**Files to create:**
- `/Users/abel/github/snipgo/internal/tmpl/expander.go`
- `/Users/abel/github/snipgo/internal/tmpl/expander_test.go` (WRITE FIRST - TDD)

**What:**
- `Expand(body string, values map[string]string) (*TemplateResult, error)`
- Replaces each `${VAR}` with corresponding value from map
- Returns error if a variable has no value and no default
- Converts `$${VAR}` escape sequences to literal `${VAR}` in output

**Test Cases (write first):**
| Input Body | Values | Expected Output | Error? |
|-----------|--------|-----------------|--------|
| `"hi ${NAME}"` | `{"NAME": "World"}` | `"hi World"` | No |
| `"${A} ${B}"` | `{"A": "1", "B": "2"}` | `"1 2"` | No |
| `"${MISSING}"` | `{}` | - | Yes: "missing value for variable: MISSING" |
| `"no vars"` | `{}` | `"no vars"` | No |
| `"$${ESCAPED}"` | `{}` | `"${ESCAPED}"` | No |
| `"${A} $${B} ${C}"` | `{"A":"x","C":"z"}` | `"x ${B} z"` | No |

**Acceptance Criteria:**
- All test cases pass
- Error message includes the missing variable name
- Escaped sequences handled after variable expansion

---

#### T4: Escape Handling Integration
**Covered by T2 and T3 tests.** No separate task file needed -- ensure escape tests exist in both parser and expander test files.

---

#### T5: Extend Snippet Struct with Variables Field
**Files to modify:**
- `/Users/abel/github/snipgo/internal/core/snippet.go`

**What:**
Add to `Snippet` struct:
```go
Variables map[string]*tmpl.Variable `yaml:"variables,omitempty" json:"variables,omitempty"`
```

Import `snipgo/internal/tmpl` package (NOT `template`).

**Also update:**
- `copySnippet()` in `/Users/abel/github/snipgo/internal/core/manager.go` -- deep copy the Variables map (iterate map, copy each Variable struct)
- `NewSnippet()` -- initialize Variables as nil (omitempty handles serialization)

**Acceptance Criteria:**
- Existing snippet files without `variables:` parse identically (nil Variables field)
- Serialized output of snippets without variables has NO `variables:` line
- `copySnippet` properly deep-copies the Variables map (modifying copy does not affect original)
- All existing tests in `snippet_test.go` still pass

---

#### T6: Frontmatter Parse/Serialize with Variables (includes choices and Name hydration)
**Files to modify:**
- `/Users/abel/github/snipgo/internal/core/frontmatter.go` -- ADD post-unmarshal hydration for Variable.Name (D9)
- `/Users/abel/github/snipgo/internal/core/frontmatter_test.go` -- ADD new test cases

**Post-unmarshal hydration in ParseFrontmatter (D9):**
After `yaml.Unmarshal(yamlBytes, snippet)`, add:
```go
for name, v := range snippet.Variables {
    v.Name = name
}
```
This ensures each `Variable.Name` is set from the map key.

**New test cases to add (extending existing 16 tests):**
1. Parse frontmatter WITH `variables:` section (including choices) -- verify Variable.Name is hydrated
2. Parse frontmatter WITHOUT `variables:` section (backward compat -- Variables is nil)
3. Serialize snippet WITH variables and verify round-trip (parse -> serialize -> parse)
4. Serialize snippet WITHOUT variables and verify NO `variables:` in output
5. Parse frontmatter with variables that have choices list
6. Round-trip preserves choices ordering

**Test input:**
```yaml
---
id: test-id
title: Deploy Script
variables:
  SERVER:
    description: Target server
    default: prod-01
  ENV:
    description: Environment
    choices:
      - staging
      - production
---
ssh ${SERVER} "deploy --env ${ENV}"
```

**Verify after parse:**
- `snippet.Variables["SERVER"].Name == "SERVER"`
- `snippet.Variables["SERVER"].Description == "Target server"`
- `snippet.Variables["SERVER"].Default == "prod-01"`
- `snippet.Variables["ENV"].Choices == ["staging", "production"]`

**Acceptance Criteria:**
- Round-trip (parse -> serialize -> parse) preserves variables including choices
- All existing 16 test cases in frontmatter_test.go still pass unchanged
- New test cases cover: variables present, variables absent, choices present, Name hydration
- Variable.Name is correctly populated from map key after unmarshal

---

#### T7: CLI Flags (--var, --raw, --preview) and Shell Safety
**Files to modify:**
- `/Users/abel/github/snipgo/cmd/snipgo/exec.go`
- `/Users/abel/github/snipgo/cmd/snipgo/copy.go`
- `/Users/abel/github/snipgo/cmd/snipgo/search.go`

**File to create:** `/Users/abel/github/snipgo/cmd/snipgo/template_helpers.go`

**What:**
Add flags to each command:
```go
cmd.Flags().StringArrayP("var", "v", []string{}, "Set variable value (KEY=VALUE, repeatable)")
cmd.Flags().Bool("raw", false, "Output snippet body without variable expansion")
```

Add to `exec` only:
```go
cmd.Flags().Bool("preview", false, "Preview expanded command before executing (default when variables present)")
cmd.Flags().Bool("no-preview", false, "Skip preview even when variables are present")
```

**Shell injection safety (D8) -- exec.go only:**
When exec expands variables:
1. Print to stderr: `WARNING: Variable values are interpolated directly into the shell command. Review the expanded command carefully.`
2. If `--no-preview` is NOT set AND snippet has variables: force preview mode
3. Preview shows expanded command and asks `Execute? [Y/n]:`
4. User must confirm before execution

**Choice display in prompts (D11):**
For variables with `choices` defined, display:
```
SERVER (Target server) [prod-01]:
  1) prod-01 (default)
  2) staging-01
  3) dev-01
Enter choice [1] or custom value:
```
Use `fmt.Scanln` for input (no external dependency needed).

**Shared helper functions in template_helpers.go:**
```go
// parseVarFlags parses -v KEY=VALUE flags into a map
func parseVarFlags(flags []string) (map[string]string, error)

// mergeVariables merges body-detected variables with frontmatter metadata (D10).
// Returns variables in body occurrence order. Frontmatter-only variables are excluded.
func mergeVariables(bodyVarNames []string, frontmatterVars map[string]*tmpl.Variable) []*tmpl.Variable

// expandSnippetBody extracts variables, merges with provided values and defaults,
// and returns the expanded body. If interactive and missing values, prompts user.
func expandSnippetBody(snippet *core.Snippet, providedVars map[string]string, raw bool) (string, error)
```

**Acceptance Criteria:**
- `snipgo exec --raw` outputs unexpanded body
- `snipgo exec -v SERVER=prod-01 -v ENV=staging` expands without prompting
- `snipgo exec` with variables shows preview by default and warns on stderr
- `snipgo exec --no-preview -v SERVER=x` skips preview
- `snipgo copy --raw "query"` copies unexpanded body
- Flags parse correctly (KEY=VALUE splitting)
- Invalid flag format (no `=`) returns clear error
- Choice variables show numbered list

---

#### T8: CLI Interactive Prompts for Variables
**Files to modify:**
- `/Users/abel/github/snipgo/cmd/snipgo/template_helpers.go` (created in T7)

**What:**
When variables are detected and not all provided via `-v` flags:
1. Call `mergeVariables()` to get ordered list with metadata (D10)
2. For each missing variable, prompt using the format from D11
3. For variables with `choices`: show numbered list, accept number or custom value
4. For variables without `choices`: simple `VARIABLE_NAME (description) [default]:` prompt
5. Empty input uses default value
6. No default and empty input = error

**ExtractVariables merge behavior (D10):**
- Body is source of truth: only variables found in body are prompted
- Frontmatter enriches body-detected variables with description/default/choices
- Frontmatter-only variables (not in body) are silently ignored
- Order follows body occurrence (first appearance)

**Test file to create:** `/Users/abel/github/snipgo/cmd/snipgo/template_helpers_test.go`

**Test Cases:**
- `parseVarFlags(["A=1", "B=2"])` -> `{"A":"1", "B":"2"}`
- `parseVarFlags(["INVALID"])` -> error
- `parseVarFlags(["A=1=2"])` -> `{"A":"1=2"}` (split on first `=` only)
- `parseVarFlags([])` -> `{}`
- `mergeVariables(["A","B"], {"A": varWithDesc})` -> `[varA_with_desc, varB_name_only]`
- `mergeVariables(["A"], {"A": varA, "STALE": varStale})` -> `[varA]` (STALE excluded)
- `mergeVariables([], {"A": varA})` -> `[]` (no body vars = nothing)

**Acceptance Criteria:**
- Interactive prompt shows variable description if available
- Default value shown in brackets
- Choices shown as numbered list per D11
- Frontmatter-only variables are NOT prompted (D10)
- All test cases pass

---

### Phase 2: Persistence, GUI, and Polish

#### T9: Variable History Store
**Files to create:**
- `/Users/abel/github/snipgo/internal/history/var_history.go`
- `/Users/abel/github/snipgo/internal/history/var_history_test.go` (WRITE FIRST - TDD)

**What:**
```go
type VarHistory struct {
    mu        sync.RWMutex
    entries   map[string][]string // variable name -> recent values (newest first)
    path      string              // file path for persistence
    maxPerVar int
}

// NewVarHistory creates a history store at the given path. Loads existing data if present.
// Main code passes config-derived path; tests pass temp file path.
func NewVarHistory(path string) (*VarHistory, error)

func (h *VarHistory) Get(varName string) []string
func (h *VarHistory) Add(varName, value string) error  // Persists to disk
func (h *VarHistory) GetAll() map[string][]string
```

**CRITICAL (D4 fix from critic):** Constructor takes `path string` parameter. This makes the type testable -- tests pass a temp file path, production code passes `~/.config/snipgo/var_history.json`. No hardcoded paths in the constructor.

**Design details:**
- Max 10 entries per variable (configurable via struct field)
- Most recently used first
- Duplicate values moved to front (not duplicated)
- Thread-safe
- Auto-creates file on first write
- Gracefully handles missing/corrupt file (returns empty history)

**Test Cases:**
- Add single value, retrieve it
- Add duplicate value, verify it moves to front
- Add 11 values, verify oldest dropped
- Get non-existent variable returns empty slice
- Round-trip: write to temp file, read back, verify
- Missing file: returns empty history, no error
- Corrupt file: returns empty history, no error (graceful degradation)

**Acceptance Criteria:**
- Constructor takes `path string` parameter (testable)
- All test cases pass using temp files
- File I/O errors don't crash (graceful degradation)
- Thread-safe reads and writes

---

#### T10: Wire History into CLI Prompts
**Files to modify:**
- `/Users/abel/github/snipgo/cmd/snipgo/template_helpers.go`
- `/Users/abel/github/snipgo/cmd/snipgo/main.go` (initialize history store)

**What:**
- Initialize `VarHistory` in `PersistentPreRun` alongside manager, passing `~/.config/snipgo/var_history.json` as path
- When prompting for a variable, show last-used value as default (if no frontmatter default)
- After successful expansion, save all used values to history

**Priority for default values:**
1. CLI `-v` flag value (highest)
2. Frontmatter `default` value
3. Last-used value from history
4. Empty (user must provide)

**Acceptance Criteria:**
- History persists across CLI invocations
- History values appear as defaults in prompts
- Frontmatter defaults take precedence over history

---

#### T11: Wails Bridge Methods for Templates
**Files to modify:**
- `/Users/abel/github/snipgo/app/app.go`

**What:**
Add methods:
```go
// ExtractVariables returns variables found in a snippet's body, enriched with
// frontmatter metadata. Follows D10 merge rules: body is source of truth.
// Variables in frontmatter but not in body are excluded.
// Returns variables in body occurrence order.
func (a *App) ExtractVariables(snippetID string) ([]*tmpl.Variable, error)

// ExpandSnippet expands a snippet's body with the given variable values.
func (a *App) ExpandSnippet(snippetID string, values map[string]string) (string, error)

// GetVariableHistory returns recent values for a variable name.
func (a *App) GetVariableHistory(varName string) ([]string, error)

// SaveVariableHistory saves variable values to history.
func (a *App) SaveVariableHistory(values map[string]string) error
```

**Also:**
- Initialize `VarHistory` in `NewApp()` with config-derived path
- Store as field on `App` struct

**ExtractVariables implementation (D10):**
1. Get snippet by ID
2. Call `tmpl.ExtractVariables(snippet.Body)` to get body var names
3. Call `mergeVariables(bodyVarNames, snippet.Variables)` to merge with frontmatter
4. Return enriched list in body order

**Acceptance Criteria:**
- All 4 methods callable from frontend via Wails IPC
- ExtractVariables follows D10 merge rules (body is source of truth)
- ExpandSnippet returns error for missing values
- TypeScript bindings regenerated (`wails generate module`)

---

#### T12: GUI VariablePromptModal Component (includes choices dropdown)
**Files to create:**
- `/Users/abel/github/snipgo/frontend/src/components/VariablePromptModal.tsx`
- `/Users/abel/github/snipgo/frontend/src/components/VariablePromptModal.test.tsx` (WRITE FIRST - TDD)

**What:**
React modal component:
```typescript
interface VariablePromptModalProps {
  variables: Variable[];          // From ExtractVariables
  onSubmit: (values: Record<string, string>) => void;
  onCancel: () => void;
}
```

**UI Design:**
- Overlay modal with semi-transparent backdrop
- Title: "Fill in Variables"
- For each variable:
  - Label: variable name
  - Helper text: description (if available)
  - Input: text field (pre-filled with default or history value)
  - Or: `<select>` dropdown (if choices available)
- Footer: "Cancel" and "Copy Expanded" buttons
- Keyboard: Enter submits, Escape cancels

**Component size:** Should be under 150 lines. Extract `VariableInput` sub-component if needed.

**Test Cases:**
- Renders all variable inputs
- Pre-fills default values
- Shows dropdown for variables with choices
- Submit returns all values
- Cancel calls onCancel
- Validates all fields have values (no empty submissions for required vars)

**Acceptance Criteria:**
- Under 200 lines
- All test cases pass
- Accessible (labels linked to inputs)
- Consistent with existing Tailwind CSS styling
- Dropdown rendered for choices, text input for non-choices

---

#### T13: Wire Modal into SnippetEditor ("Copy Expanded" only)
**Files to modify:**
- `/Users/abel/github/snipgo/frontend/src/components/ActionButtons.tsx`
- `/Users/abel/github/snipgo/frontend/src/components/SnippetEditor.tsx`
- `/Users/abel/github/snipgo/frontend/src/bridge.ts`
- `/Users/abel/github/snipgo/frontend/src/types.ts`

**What:**
1. Add `Variable` type to `types.ts`:
```typescript
export interface Variable {
  name: string;
  description?: string;
  default?: string;
  choices?: string[];
}
```

2. Add bridge methods to `bridge.ts`:
```typescript
ExtractVariables(snippetID: string): Promise<Variable[]>;
ExpandSnippet(snippetID: string, values: Record<string, string>): Promise<string>;
GetVariableHistory(varName: string): Promise<string[]>;
SaveVariableHistory(values: Record<string, string>): Promise<void>;
```

3. Modify `ActionButtons.tsx`:
- "Copy" button: if variables detected, show modal first, then copy expanded body
- **NO "Execute" button** -- deferred to future work. Phase 2 ships "Copy Expanded" only.

4. Modify `SnippetEditor.tsx`:
- Add state for modal visibility: `showVariableModal`
- Add state for pending action: `pendingAction: 'copy' | null`
- Wire modal submit to perform the pending action with expanded body

**Flow:**
```
User clicks Copy -> Check for variables ->
  No variables: copy raw body (existing behavior)
  Has variables: show modal -> user fills values -> expand -> copy expanded body -> save to history
```

**Acceptance Criteria:**
- Snippets without variables: zero behavior change
- Snippets with variables: modal appears before copy
- Modal pre-fills from history via bridge
- After expansion, values saved to history
- TypeScript types match Go types
- NO "Execute" button in GUI (deferred)

---

#### T14: CLI Integration Tests (end-to-end)
**Files to create:**
- `/Users/abel/github/snipgo/cmd/snipgo/template_integration_test.go`

**What:**
End-to-end tests that:
1. Create a snippet with variables via the manager
2. Verify `parseVarFlags` correctly
3. Verify `expandSnippetBody` with various combinations
4. Test the full flow with mocked stdin for interactive prompts
5. Verify shell injection warning is printed to stderr for exec
6. Verify --preview is default-on for exec with variables

**Acceptance Criteria:**
- Tests run without fzf dependency (mock selection)
- Tests cover: all flags provided, partial flags (prompt needed), --raw, --preview, --no-preview
- Tests verify history is updated after expansion
- Tests verify merge behavior (D10): frontmatter-only vars not prompted
- Tests verify shell injection warning on stderr

---

#### T15: Frontend Tests for Variable Modal
**Covered by T12 test file.** Additional integration-level tests:

**Files to create/modify:**
- Add test cases to `/Users/abel/github/snipgo/frontend/src/components/SnippetEditor.test.tsx`

**What:**
- Verify Copy button triggers modal when snippet has variables
- Verify Copy button does NOT trigger modal when snippet has no variables
- Verify expanded text is copied after modal submission

**Acceptance Criteria:**
- All new tests pass
- Existing SnippetEditor tests still pass
- No snapshot regressions

---

## 7. Commit Strategy

| Commit | Tasks | Message |
|--------|-------|---------|
| 1 | T1, T2, T3 | `feat(core): add template engine with variable parsing and expansion` |
| 2 | T5, T6 | `feat(core): extend snippet struct with variables metadata in frontmatter` |
| 3 | T7, T8 | `feat(cli): add variable expansion support to exec, copy, and search commands` |
| 4 | T9, T10 | `feat(core): add variable history persistence and CLI integration` |
| 5 | T11 | `feat(gui): add Wails bridge methods for template expansion` |
| 6 | T12, T13 | `feat(gui): add variable prompt modal for Copy Expanded action` |
| 7 | T14, T15 | `test: add integration tests for template system (CLI and GUI)` |

Branch name: `feature/dynamic-snippet-templates`

---

## 8. Risk Identification and Mitigations

### R1: Backward Compatibility Breakage
**Risk:** Adding `Variables` field to Snippet struct could break existing YAML parsing.
**Mitigation:** Use `omitempty` tag. YAML parser ignores unknown fields by default. Add explicit backward-compat test case (existing snippet file parsed with new code).
**Severity:** HIGH | **Likelihood:** LOW

### R2: Circular Import (core <-> tmpl)
**Risk:** `core` package imports `tmpl` for Variable type; `tmpl` might need `core.Snippet`.
**Mitigation:** Keep `tmpl` package independent. It operates on raw strings, not Snippet objects. The `Variable` type lives in `tmpl` package; `core.Snippet` references it via import. No reverse dependency.
**Severity:** MEDIUM | **Likelihood:** MEDIUM

### R3: Regex Edge Cases
**Risk:** Variable regex might match inside code blocks, strings, or comments where expansion is unwanted.
**Mitigation:** For MVP, expand all occurrences. Document that users should use `$${VAR}` to escape. Future enhancement: skip expansion inside backtick code blocks.
**Severity:** LOW | **Likelihood:** MEDIUM

### R4: History File Corruption
**Risk:** Concurrent CLI invocations could corrupt var_history.json.
**Mitigation:** Use file locking (or atomic write via temp file + rename). Read-before-write pattern. Graceful degradation if file is corrupt (start fresh).
**Severity:** LOW | **Likelihood:** LOW

### R5: Large Variable Values
**Risk:** User could paste large text as variable value, causing performance issues.
**Mitigation:** Cap variable values at 10KB. Show warning for values over 1KB.
**Severity:** LOW | **Likelihood:** LOW

### R6: Frontend Bundle Size
**Risk:** New modal component could increase bundle size.
**Mitigation:** Component is small (< 200 lines). No new npm dependencies needed. Modal uses existing Tailwind utilities.
**Severity:** LOW | **Likelihood:** LOW

### R7: Shell Injection via Variable Values (CRITICAL)
**Risk:** `exec` runs `sh -c expanded_body`. Malicious or accidental variable values (e.g., `; rm -rf /`) execute as shell commands.
**Mitigation:**
1. `--preview` is default-on when variables are present in exec (D8)
2. Warning printed to stderr before execution
3. User must explicitly confirm after seeing expanded command
4. `--no-preview` available for scripted/trusted use cases
5. Documented in `snipgo exec --help`
**Severity:** HIGH | **Likelihood:** MEDIUM

---

## 9. Verification Steps

### Unit Test Verification
```bash
# After each phase, run:
go test ./internal/tmpl/... -v -cover     # Template engine (target: 90%+)
go test ./internal/core/... -v -cover     # Core (verify no regression)
go test ./internal/history/... -v -cover  # History store
go test ./cmd/snipgo/... -v               # CLI tests
cd frontend && pnpm test                  # Frontend tests
```

### Manual Verification Checklist

**Phase 1 (CLI):**
1. Create snippet with `${SERVER}` in body via `snipgo new`
2. `snipgo exec -v SERVER=myhost` -- verifies flag-based expansion (shows preview by default)
3. `snipgo exec` (same snippet) -- verifies interactive prompt appears, then preview
4. `snipgo exec --raw` -- verifies raw output (no expansion)
5. `snipgo exec --no-preview -v SERVER=myhost` -- verifies preview skipped
6. `snipgo copy --raw "snippet"` -- verifies raw copy
7. Create snippet WITHOUT variables -- verify zero behavior change
8. Create snippet with `choices` in frontmatter -- verify numbered list in CLI prompt

**Phase 2 (GUI):**
1. Open GUI, select snippet with variables
2. Click Copy -- verify modal appears
3. Fill values, click "Copy Expanded" -- verify expanded text in clipboard
4. Re-open modal -- verify last-used values pre-filled
5. Select snippet WITHOUT variables, click Copy -- verify no modal (direct copy)
6. Verify no "Execute" button in GUI (deferred)

### Lint Verification
```bash
make lint          # Both Go and TypeScript
make type-check    # TypeScript type checking
```

---

## 10. File Change Summary

### New Files (10)
| File | Task | Description |
|------|------|-------------|
| `internal/tmpl/types.go` | T1 | Variable and result types |
| `internal/tmpl/parser.go` | T2 | Variable extraction from body |
| `internal/tmpl/parser_test.go` | T2 | Parser tests (TDD) |
| `internal/tmpl/expander.go` | T3 | Variable substitution engine |
| `internal/tmpl/expander_test.go` | T3 | Expander tests (TDD) |
| `internal/history/var_history.go` | T9 | Variable history persistence |
| `internal/history/var_history_test.go` | T9 | History tests (TDD) |
| `cmd/snipgo/template_helpers.go` | T7 | Shared CLI template utilities (includes mergeVariables) |
| `cmd/snipgo/template_helpers_test.go` | T8 | Template helper tests (includes merge behavior tests) |
| `cmd/snipgo/template_integration_test.go` | T14 | End-to-end CLI tests |

### New Frontend Files (2)
| File | Task | Description |
|------|------|-------------|
| `frontend/src/components/VariablePromptModal.tsx` | T12 | GUI variable modal |
| `frontend/src/components/VariablePromptModal.test.tsx` | T12 | Modal unit tests (TDD) |

### Modified Files (11)
| File | Task | Change |
|------|------|--------|
| `internal/core/snippet.go` | T5 | Add Variables field to Snippet struct |
| `internal/core/manager.go` | T5 | Deep copy Variables in copySnippet() |
| `internal/core/frontmatter.go` | T6 | Add 3-line Variable.Name hydration after unmarshal |
| `internal/core/frontmatter_test.go` | T6 | Add 6 new test cases (variables, choices, Name hydration) |
| `cmd/snipgo/exec.go` | T7 | Add -v, --raw, --preview, --no-preview flags + expansion + shell warning |
| `cmd/snipgo/copy.go` | T7 | Add -v, --raw flags + expansion |
| `cmd/snipgo/search.go` | T7 | Add -v, --raw flags + expansion |
| `cmd/snipgo/main.go` | T10 | Initialize VarHistory with config path |
| `app/app.go` | T11 | Add 4 template bridge methods + VarHistory field |
| `frontend/src/types.ts` | T13 | Add Variable interface |
| `frontend/src/bridge.ts` | T13 | Add 4 bridge method wrappers |

### Modified Frontend Files (2)
| File | Task | Change |
|------|------|--------|
| `frontend/src/components/ActionButtons.tsx` | T13 | Variable-aware copy flow (no Execute button) |
| `frontend/src/components/SnippetEditor.tsx` | T13 | Modal state + wiring |

### Unchanged Files (critical for backward compat)
- `internal/storage/filesystem.go` -- No changes needed
- `internal/config/config.go` -- No changes needed
- All existing test files -- Must continue passing

---

## 11. Success Criteria

1. **Zero regressions**: `go test ./...` passes, `pnpm test` passes, `make lint` clean
2. **Template engine coverage**: `go test ./internal/tmpl/... -cover` shows 90%+
3. **Backward compatibility**: Existing snippet `.md` files parse and serialize identically
4. **CLI functional**: All three output commands (`exec`, `copy`, `search`) support `-v`, `--raw`
5. **CLI safe**: `exec` with variables shows preview by default, warns about shell injection
6. **GUI functional**: Variable prompt modal appears for templated snippets with "Copy Expanded", pre-fills from history
7. **GUI scoped**: No "Execute" button in GUI (deferred to future)
8. **History persists**: Close and reopen app, previous variable values appear as defaults
9. **History testable**: VarHistory constructor accepts path parameter, tests use temp files
10. **Documentation**: `snipgo exec --help` shows new flags with examples and shell safety warning

---

## 12. Critic Issue Resolution Tracker

| # | Issue | Resolution | Where |
|---|-------|------------|-------|
| C1 | Variable.Name empty after YAML unmarshal | `yaml:"-" json:"name"` tag + post-unmarshal hydration | D9, T1, T6 |
| C2 | ExtractVariables merge behavior unspecified | Body is source of truth; explicit 4-rule merge spec | D10, T7, T8, T11 |
| C3 | Shell injection risk in exec | --preview default-on + stderr warning + --no-preview opt-out | D8, T7, R7 |
| C4 | VarHistory constructor untestable | `NewVarHistory(path string)` parameter | D3 note, T9 |
| C5 | Choice prompt UX unspecified | Numbered list with fmt.Scanln, concrete format | D11, T7, T8 |
| M1 | Dead VariableValue struct | Removed from T1 types | T1 |
| M2 | Package shadows stdlib template | Renamed to `internal/tmpl/` | D4, all refs |
| M3 | GUI Execute button premature | Deferred; Phase 2 = "Copy Expanded" only | D6, T13 |
| M4 | T14 ambiguous | Folded into T6 (choices parse) and T12 (choices dropdown) | T6, T12 |
| M5 | Frontmatter test count wrong | Corrected to 16 (was 10) | Section 1, T6 |
