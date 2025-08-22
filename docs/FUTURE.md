# Potential Future Features
Awesome—here’s a concrete, repo-aware implementation plan to add the things we discussed to **Scout-MCP**, wired to your layout (`mcptools/`, approval system, session tokens, etc.). I skimmed your README and repo tree (top-level dirs, tool list, approval model) and will align to that. I couldn’t open the individual Go files under `mcptools/` due to GitHub’s page rendering limits in this environment, but the README and structure are enough to target the right places. ([GitHub][1])

---

## Milestone 1 — Core plumbing (IR, atomic writes, globs)

### 1.1 Create a tiny **edit IR** package (pure Go, no MCP)

Add `internal/edits/` with:

* `edit_ir.go` — the minimal IR structs (`EditBlock`, `SingleAnchor`, `PairAnchors`, `EditOp`).
* `validate.go` — rejects invalid shapes (e.g., `replace_range` requires lines, can’t also have anchors).
* `plan.go` — pure function that takes `[]EditBlock` + file content map and returns planned splices + a preview unified diff (per file).
* `apply.go` — pure function that applies the plan to text (no disk I/O).
* `eol.go` — detect/normalize `\n` vs `\r\n` and re-emit preserved EOL.

> Why: you test this *once* and reuse from every tool (including your Go-specific tools that already edit functions/types).

**Imports to add (go.mod):**

* **atomic writes:** `github.com/google/renameio/v2` (safe replace, fsync, cross-platform). ([GitHub][2], [Go Packages][3])
* **human diffs (preview):** `github.com/pmezard/go-difflib/difflib`. ([GitHub][4], [Go Packages][5])
* **optional machine-patches:** `github.com/sergi/go-diff/diffmatchpatch` (only if you also want DMP patch IO). ([GitHub][6], [Go Packages][7])
* **parse unified diffs (optional input):** `github.com/bluekeyes/go-gitdiff/gitdiff`. ([GitHub][8], [Go Packages][9])
* **globbing:** `github.com/bmatcuk/doublestar/v4` (for `**` patterns). ([GitHub][10], [Go Packages][11])

### 1.2 Atomic writer

Add `fileutil/atomic.go`:

* `WriteFileAtomic(path string, data []byte, mode os.FileMode) error`
  Use `renameio.WriteFile` internally for correct temp→fsync→rename semantics; preserve/restore mode if needed. ([GitHub][2])

* `WriteFileAtomicWithBackup(path, backupExt string, …)` (optional): rename old to `*.bak` before replace.

> We’ll keep your existing `update_file` tool but route it through this writer behind a feature flag, then introduce a safer “atomic” variant (see Milestone 3).

```go
package safeio

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

func AtomicWriteFile(path string, data []byte, perm os.FileMode, makeBackup bool, expectedOldSHA256 string) (newSHA256 string, err error) {
	var f *os.File
	var dir *os.File
	tmp := ""
	backup := path + ".bak"

	// Preserve existing mode if dest exists, unless perm is non-zero.
	fi, statErr := os.Stat(path)
	if statErr == nil {
		if perm == 0 {
			perm = fi.Mode()
		}
	}

	// Optional: verify expected old checksum if provided.
	if expectedOldSHA256 != "" && statErr == nil {
		var old []byte
		old, err = os.ReadFile(path)
		if err != nil {
			goto end
		}
		sum := sha256.Sum256(old)
		if hex.EncodeToString(sum[:]) != expectedOldSHA256 {
			err = fmt.Errorf("precondition failed: expected old sha256 %s", expectedOldSHA256)
			goto end
		}
	}

	// Create temp in same dir
	dirPath := filepath.Dir(path)
	tmpf, cerr := os.CreateTemp(dirPath, "."+filepath.Base(path)+".tmp-*")
	if cerr != nil {
		err = cerr
		goto end
	}
	f = tmpf
	tmp = f.Name()

	// Set perms on temp so the final file has desired mode.
	if perm != 0 {
		cherr := f.Chmod(perm)
		if cherr != nil {
			err = cherr
			goto end
		}
	}

	// Write contents
	_, err = f.Write(data)
	if err != nil {
		goto end
	}

	// Flush file to disk
	err = f.Sync()
	if err != nil {
		goto end
	}

	// Close the file before rename
	err = f.Close()
	if err != nil {
		goto end
	}
	f = nil

	// Optionally rotate a backup of the current file (best-effort)
	if makeBackup {
		if _, e := os.Stat(path); e == nil {
			_ = os.Remove(backup)         // ignore errors
			_ = os.Rename(path, backup)   // not atomic with final rename, but preserves old content
		}
	}

	// On Windows, os.Rename won't replace. Do a best-effort remove first.
	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}

	// Atomic replace
	err = os.Rename(tmp, path)
	if err != nil {
		goto end
	}

	// fsync the directory to persist rename
	dir, err = os.Open(dirPath)
	if err != nil {
		goto end
	}
	_ = dir.Sync() // best-effort on some filesystems
	_ = dir.Close()
	dir = nil

	// Compute checksum of new data to return
	sum := sha256.Sum256(data)
	newSHA256 = hex.EncodeToString(sum[:])

end:
	if f != nil {
		_ = f.Close()
	}
	if tmp != "" {
		_ = os.Remove(tmp) // remove leftover temp on error
	}
	if dir != nil {
		_ = dir.Close()
	}
	return
}
```

### 1.3 Session-scoped ignores

Extend your session struct (where you store the **session token** you already require for all tools except `start_session`) to carry:

```go
type SessionState struct {
    AllowedRoots []string
    Ignores      []string // doublestar patterns, e.g. "**/.git/**", "**/*.min.js"
    // …
}
```

---

## Milestone 2 — New general-purpose tools in `mcptools/`

Your tools are defined under `mcptools/`, so we’ll add one file per tool, following your existing pattern (registration + handler). (Because I couldn’t fetch those files here, I’ll refer to them as `registerX()`/`handleX()`—mirror the names you already use.)

### 2.1 `apply_edit_blocks` (the “one general tool”)

**Path:** `mcptools/apply_edit_blocks.go`
**Risk:** *high* (requires approval token; use your existing `request_approval`/`generate_approval_token` flow). Your README shows that write ops already require confirmation; keep that. ([GitHub][1])

**Input (JSON)** — minimal, stable:

```json
{
  "edits": [ /* []EditBlock (IR) */ ],
  "expectedSha256": { "path/to/file.go": "abcd…" },
  "backup": true,
  "previewOnly": false,
  "stopOnFailure": false
}
```

**Output:**

* `results[]`: `{ path, status: "applied|noop|skipped|failed", message, unifiedDiff? }`
* `previewUnifiedDiff` (if `previewOnly:true`)

**Handler flow:**

1. Validate session token (you already do this across tools). ([GitHub][1])
2. `edits.Validate()` → reject impossible combos (fast fail).
3. Read current file contents (bounded by allowed roots + ignores).
4. `plan := edits.Plan()` → returns in-memory result + unified diff via `difflib`. ([GitHub][4])
5. If `previewOnly` → return preview + per-block plan statuses.
6. Otherwise require approval token (your existing mechanism).
7. For each changed file: `fileutil.WriteFileAtomic(...)`. ([GitHub][2])

> Keep **Go-specific tools** (edit function/type) but have them **lower to the same `[]EditBlock`** and then call the same engine, so all edits share tests.

### 2.2 `parse_edit_blocks` (optional, ergonomic)

**Path:** `mcptools/parse_edit_blocks.go`
Accepts the RFC-822-ish “Keywords + Body” text and returns structured `[]EditBlock` + diagnostics (unknown header, malformed regex, conflicting locators). This is a *helper* for chat authoring; the **canonical** format stays JSON.

**Input:** `{ "rfc822": "…multi chunks separated by ---…" }`
**Output:** `{ "ir": { "edits": [...] }, "diagnostics": [...] }`

### 2.3 `search_content`

**Path:** `mcptools/search_content.go`
Greps contents with globs + ignores (session-scoped). Returns excerpts + line numbers (read-only).

**Input:**
`{ "query": "regexp or literal", "regex": false, "paths": ["."], "includeGlobs": ["**/*.go"], "excludeGlobs": [], "maxMatchesPerFile": 20 }`

**Output:**
`{ "results": [ { "path": "…", "matches": [ { "line": 42, "excerpt": "…" } ] } ] }`

Use `doublestar` for matching; respect `SessionState.Ignores`. ([GitHub][10])

### 2.4 `get_file_metadata`

**Path:** `mcptools/get_file_metadata.go`
Returns size/mtime/mode/sha256 and `isBinary` (simple heuristic: NUL byte check + MIME sniff).

**Output:**
`{ "size": 1234, "mtime": "RFC3339", "mode": "0644", "sha256": "…", "isBinary": false, "mime": "text/x-go" }`

### 2.5 `list_directory`

**Path:** `mcptools/list_directory.go`
Enumerate a directory with `depth`, `includeGlobs`, `excludeGlobs`, and a `dirsOnly` flag. (Helps models plan edits without reading files.)

### 2.6 **(Optional)** `generate_diff` / `apply_unified_diff`

If you want to accept *unified diffs* from a client, add:

* `generate_diff` — produces unified diffs (uses `difflib`, already in the plan). ([GitHub][4])
* `apply_unified_diff` — parse with `bluekeyes/go-gitdiff` and apply strictly (or add a small “fuzz” window). Note: `go-gitdiff` applies in strict mode by default; drift-tolerant application is an open topic. ([GitHub][8])

---

## Milestone 3 — Integrate approval & de-risk “whole-file write”

You already have `request_approval` and `generate_approval_token`, with a preview-before-apply model and risk levels. Keep that and:

* Mark `apply_edit_blocks` as **high risk** (multiple files/edits).
* Add `update_file_atomic` (new tool) as **medium risk** and **deprecate** `update_file` behind a feature flag so you stop unsafe overwrites by default. (Internally both route through `fileutil.WriteFileAtomic`.) ([GitHub][1])

---

## Milestone 4 — Extend **start\_session** instructions & tool\_help

Update the instructions that `start_session` returns so the model/you follow a safe loop:

1. **Plan**: use `search_content`/`read_files`
2. **Propose**: produce `edits` (JSON) or RFC-822 text → `parse_edit_blocks`
3. **Preview**: call `apply_edit_blocks` with `previewOnly:true`
4. **Approve**: call `request_approval` → `generate_approval_token`
5. **Apply**: call `apply_edit_blocks` with token

(This mirrors your existing session/instruction design. Your README explicitly says the session returns guidelines; add these there.) ([GitHub][1])

---

## Milestone 5 — Tests (fast + exhaustive without permutations hell)

Create `test/edits/` and `testdata/edits/`:

### 5.1 Unit tests (pure)

* `internal/edits/validate_test.go` — bad shapes are rejected.
* `internal/edits/plan_test.go` — table tests per **locator family** × **op**:

    * line range (in/out of bounds; overlap)
    * single anchor (first/last/nth; offset; not found)
    * start/end anchors (both missing / only one / nested)
    * ops: insert/replace/delete/append
* `internal/edits/eol_test.go` — CRLF/LF round-trip.

### 5.2 Golden tests

Each case folder:

```
case-010-insert-after-anchor/
  before.go
  edits.json
  after.go
  preview.diff
```

Runner loads `before`, reads `edits.json` → `Plan()` → compare `preview.diff`, then `Apply()` → compare `after`.

### 5.3 Fuzz/property tests

* Fuzz anchors: insert then delete restores original.
* Idempotency: apply twice with `Idempotent:true` yields same hash.
* Mixed EOL files stay stable.

### 5.4 MCP integration tests

Under `test/`, spin a temp workspace (respecting allowed roots); call your **handlers** (not stdio) for:

* `parse_edit_blocks` → `apply_edit_blocks` (preview) → approval → apply.
* `search_content` with globs/ignores.

---

## Milestone 6 — Performance & limits

* **Caps:** `maxFilesPerCall`, `maxEditsPerFile`, `maxBytesPerFile` (reject if exceeded with actionable error).
* **Binary guard:** refuse content edits on binary files unless `force:true`.
* **Concurrency:** lock per path (advisory) to avoid simultaneous writers; ok to start with a simple `sync.Map[path]*mutex`.
* **Logging:** keep your current access/approval logs; add a short per-file “applied/skip/fail + sha256(before→after)” line (no content).

---

## File-level changes to expect

```
/internal/edits/
  edit_ir.go
  validate.go
  plan.go
  apply.go
  eol.go

/fileutil/
  atomic.go                   // renameio-based writer

/mcptools/
  apply_edit_blocks.go        // new
  parse_edit_blocks.go        // new (optional, ergonomic)
  search_content.go           // new
  get_file_metadata.go        // new
  list_directory.go           // new
  update_file_atomic.go       // new (wrap existing update_file)

/docs/
  tools-apply-edit-blocks.md  // short reference for schemas and examples
```

And a small tweak wherever you **register** tools today (likely in `mcp.go` or a central registry) to add the five new tools.

---

## Schemas you can paste into your tool registry

### `apply_edit_blocks` (input)

```json
{
  "type": "object",
  "properties": {
    "edits": {
      "type": "array",
      "items": { "$ref": "#/$defs/EditBlock" },
      "minItems": 1
    },
    "expectedSha256": {
      "type": "object",
      "additionalProperties": { "type": "string" }
    },
    "backup": { "type": "boolean", "default": true },
    "previewOnly": { "type": "boolean", "default": false },
    "stopOnFailure": { "type": "boolean", "default": false }
  },
  "required": ["edits"],
  "$defs": {
    "EditBlock": {
      "type": "object",
      "properties": {
        "path": { "type": "string" },
        "op": { "enum": ["insert_after","insert_before","replace_range","replace_anchor","delete_range","delete_anchor","append"] },
        "startLine": { "type": "integer", "minimum": 1 },
        "endLine": { "type": "integer", "minimum": 1 },
        "anchor": {
          "type": "object",
          "properties": {
            "regex": { "type": "string" },
            "where": { "enum": ["first","last","nth"] },
            "n": { "type": "integer", "minimum": 1 },
            "offsetLines": { "type": "integer", "default": 0 }
          },
          "required": ["regex"]
        },
        "anchors": {
          "type": "object",
          "properties": {
            "start": { "type": "string" },
            "end": { "type": "string" }
          },
          "required": ["start","end"]
        },
        "body": { "type": "string" },
        "ifNotFound": { "enum": ["error","skip","create"], "default": "error" },
        "ifMultiple": { "enum": ["error","first","last","nth","all"], "default": "error" },
        "nth": { "type": "integer", "minimum": 1 },
        "idempotent": { "type": "boolean", "default": false },
        "contextBefore": { "type": "integer", "minimum": 0, "default": 0 },
        "contextAfter": { "type": "integer", "minimum": 0, "default": 0 }
      },
      "required": ["path","op"]
    }
  }
}
```

### `apply_edit_blocks` (output)

```json
{
  "type": "object",
  "properties": {
    "results": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "path": { "type": "string" },
          "status": { "enum": ["applied","noop","skipped","failed"] },
          "message": { "type": "string" },
          "unifiedDiff": { "type": "string" }
        },
        "required": ["path","status"]
      }
    },
    "previewUnifiedDiff": { "type": "string" }
  },
  "required": ["results"]
}
```

Similar small schemas for the other tools (`search_content`, `get_file_metadata`, `list_directory`, `parse_edit_blocks`).

---

## Glue code patterns (aligned to your style)

You prefer explicit two-line short-var then `if`. I’ll mirror that here.

**Atomic write wrapper:**

```go
data, ok := newContentBytes
if !ok {
    // construct or fetch your bytes; placeholder for your flow
}

err := renameio.WriteFile(path, data, 0o644)
if err != nil {
    return fmt.Errorf("atomic write failed for %s: %w", path, err)
}
```

*(In real code, read current mode; if zero, preserve.)* ([GitHub][2])

**Plan→Preview in handler:**

```go
plan, ok := edits.Plan(ctx, req.Edits, readFilesFn)
if !ok {
    return errorPlanning
}
diff, ok := edits.PreviewUnified(plan)
if !ok {
    // handle
}
```

**Regex anchors (multiline mode):**

```go
re, err := regexp.Compile("(?m)" + anchor.Regex)
if err != nil {
    return fmt.Errorf("bad regex: %w", err)
}
```

---

## Docs you should update

* **README → API Tools:** add the 4–5 new tools with short examples (your README already lists 20 tools with concise blurbs—mirror that format). ([GitHub][1])
* **CLAUDE.md / start\_session instructions:** add the plan→preview→approve→apply loop.
* **Security section:** note size/binary caps; note that all writes are now atomic by default. ([GitHub][1])

---

## Rollout strategy (minimal disruption)

1. Land **IR + atomic writer** (no new tools yet).
2. Ship `search_content` + `get_file_metadata` (read-only).
3. Add `apply_edit_blocks` **behind a feature flag** in `start_session` instructions (“use this path when available”).
4. Route existing Go-specific tools to compile → IR → same engine.
5. Add `update_file_atomic` and mark legacy `update_file` as **dangerous / deprecated** in `tool_help`. ([GitHub][1])

---

## Notes on dependencies/choices (why these)

* `renameio` is the least-surprising, cross-platform safe file replace (temp → fsync → rename; also fsyncs dir). ([GitHub][2])
* `difflib` renders human-readable unified diffs easily (for previews and approval screens). ([GitHub][4])
* `diffmatchpatch` is handy if you ever want drift-tolerant machine patches; otherwise your **anchor/range edits** plus strict unified diffs are enough. ([GitHub][6])
* `go-gitdiff` only if you *accept* unified diffs as input (it parses & can apply in strict mode). ([GitHub][8])
* `doublestar` gives you `**` in includes/excludes and aligns with how devs expect globs to work. ([GitHub][10])

---


[1]: https://github.com/mikeschinkel/scout-mcp "GitHub - mikeschinkel/scout-mcp: An MCP Server for enabling Claude UI to access local files using a Cloudflare Tunnel."
[2]: https://github.com/google/renameio?utm_source=chatgpt.com "Package renameio provides a way to atomically create or ..."
[3]: https://pkg.go.dev/github.com/google/renameio/v2?utm_source=chatgpt.com "renameio package - github.com/google ..."
[4]: https://github.com/pmezard/go-difflib?utm_source=chatgpt.com "pmezard/go-difflib: Partial port of Python difflib package to Go"
[5]: https://pkg.go.dev/github.com/pmezard/go-difflib/difflib?utm_source=chatgpt.com "difflib"
[6]: https://github.com/sergi/go-diff?utm_source=chatgpt.com "sergi/go-diff: Diff, match and patch text in Go"
[7]: https://pkg.go.dev/github.com/sergi/go-diff/diffmatchpatch?utm_source=chatgpt.com "diffmatchpatch package - github.com/sergi/go-diff ..."
[8]: https://github.com/bluekeyes/go-gitdiff?utm_source=chatgpt.com "bluekeyes/go-gitdiff: Go library for parsing and applying ..."
[9]: https://pkg.go.dev/github.com/bluekeyes/go-gitdiff/gitdiff?utm_source=chatgpt.com "gitdiff package - github.com/bluekeyes ..."
[10]: https://github.com/bmatcuk/doublestar?utm_source=chatgpt.com "bmatcuk/doublestar: Implements support for double star ..."
[11]: https://pkg.go.dev/github.com/bmatcuk/doublestar?utm_source=chatgpt.com "doublestar package - github.com/bmatcuk ..."
