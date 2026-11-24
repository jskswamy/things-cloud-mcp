# Outstanding fixes for the standalone Go server — design

Date: 2026-11-24
Status: approved 2026-11-24

## Goal

Fix the seven known defects that the original project only fixed in its
TypeScript rewrite, so the self-hosted Go binary is safe to keep using.
Two of them can corrupt a Things Cloud account beyond client-side
repair, so they go first.

## Out of scope

- Converging the vendored `things-cloud-sdk/` onto upstream
  (`docs/sdk-workflow.md`, "Transition"). That is its own piece of work.
  Every fix here is made against the vendored copy as it is today.
- OAuth bearer 401 (original PR #28): only the hosted service used OAuth.
- A hard-delete tool (#19) and the Things beta schema report (#21).
- Sort index 0 on area creation and -1237 on tag creation. The reported
  crash is specific to `Task6` items.

## Architecture context

Observed at `b03d51d` (baseline `go test ./...` passes in the server and
in the vendored SDK).

- Single package: `main.go` (tools, state, sync loop), `oauth.go`,
  `landing.go`. The SDK is vendored in `things-cloud-sdk/` and bound by a
  `replace` line in `go.mod`.
- The server calls into the SDK at 117 sites across 29 symbols. Nothing
  in the SDK calls back. The seam is one-way and clean.
- All writes go through one function, `writeAndSync` (`main.go:1297`).
  All reads go through `callSerialized` → `syncAndRebuild`
  (`main.go:1276`). These two functions are where most fixes land.
- The server reimplements two things the SDK owns: identifier generation
  (`generateUUID`, `main.go:176`) and "today" computed in UTC.
- The vendored SDK already has `NewUUID`, `ValidateUUID`, a typed
  `APIError` carrying the `things-response` header, and
  `IsAbusePrevention`.

## Fixes

Numbers refer to issues and pull requests in the original project.

### 1. Identifiers Things cannot decode (#30)

`generateUUID` drops leading zero bytes, so about 1 in 255 new items gets
a short identifier. Things crashes on it on every device and the record
cannot be removed by a client.

- `generateUUID` returns `thingscloud.NewUUID()`.
- `identifiableItems` (called first in `writeAndSync`) rejects a
  *creation* whose identifier fails `thingscloud.ValidateUUID`, so a new
  bad identifier can never reach the server whatever produced it. Edits
  and deletions are not checked: they address identifiers that already
  exist in the account, and old accounts hold non-Base58 ones.
- Test: 100,000 generated identifiers all validate; a creation carrying
  a short identifier is refused before any network call; an edit of an
  item with a legacy identifier still goes through.

### 2. Non-positive sort index on edit (PR #18)

Editing or trashing a task whose stored sort index is zero or negative
re-emits that index and crashes Things on every device.

- Apply PR #18 (`healNonPositiveIndex`, called in `writeAndSync` after
  the pre-write sync). It applies cleanly with its tests.
- PR #18 covers map payloads only. Also clamp `Ix: task.Index` to at
  least 1 in the recurring next-occurrence payload (`main.go:668`).
- Test: PR #18's tests, plus one for the struct payload.

### 3. Write lost on 409 OutdatedAncestor (#31)

If another device commits between the pre-write sync and the write, the
commit is rejected cleanly and the tool call fails.

- In `writeAndSync`, when the write fails with `APIError` status 409 and
  `things-response: OutdatedAncestor`, repeat the whole pre-write
  sequence (sync, then the sort-index guard from fix 2, then send), so
  the guard always sees the state the commit is based on. At most 3
  attempts. Every other failure, including
  `CommitUncertainError`, keeps today's behaviour (no automatic retry).
- The vendored `History.Write` already returns `*APIError` for a non-200
  response (`things-cloud-sdk/histories.go`), so no SDK change is needed.
- Test: fake server returns 409 once then 200; exactly one resend, state
  updated. Fake returns 409 three times; error surfaces.

### 4. "Today" computed in UTC (PR #25)

`isToday`, `isScheduledForTodayOrPast` and the recurrence reference use
UTC, so Today is wrong for anyone east or west of UTC around midnight.

- Apply PR #25 (a `userToday()` helper and one location resolved at
  start). Its `main.go` and test changes apply cleanly; the README hunk
  is redone by hand.
- Change how the location is resolved: `MCP_TIMEZONE` if set, otherwise
  Go's `time.Local`. `time.Local` already honours `TZ` and the system
  zone, so PR #25's separate `TZ` lookup and its UTC fallback are
  removed. This is a self-hosted binary, usually on the user's own
  machine, so the system zone is the right default with no configuration.
- Test: PR #25's tests, adjusted for the new fallback.

### 5. Stuck after a Things Cloud account reset (#27, #35)

After a reset the stored cursor points past the new history, and every
incremental fetch returns `404 AccountIssue`.

- In `syncAndRebuild`, when `incrementalSync` fails with `APIError` 404
  and `things-response: AccountIssue`, re-resolve the account's history
  with `client.OwnHistory()`. Only if its ID differs from the one in use
  (the signature of a reset) switch to it and run `fullRebuild`.
  Otherwise, or if the rebuild fails, return the original error. This
  keeps a persistent 404 from refetching the whole history on every call.
- This deliberately narrows the existing rule that a full rebuild is
  "never an automatic error fallback": it applies to this one error
  only, where the old cursor is known to be meaningless.
- Test: fake server returns 404 AccountIssue for the old history and a
  different, valid own history; state is rebuilt. Same history ID, or any
  other error, surfaces without a rebuild.

### 6. things_diagnose returns null (PR #26)

Nil slices marshal to `null`, which fails the tool's own output schema
on a healthy account.

- Apply PR #26. It applies cleanly with its tests.

### 7. No backoff on 429 AbusePrevention (#32)

- In the vendored SDK client, retry a GET up to 3 times with 1s, 2s, 4s
  waits when `IsAbusePrevention` is true. Writes are never retried here.
- SDK-side change: upstream candidate under `docs/sdk-workflow.md`.
- Test: fake server returns 429 AbusePrevention twice then 200; a POST
  that gets 429 is not retried. The wait is injectable so the test does
  not sleep.

### 8. Release

- Bump the version constant, merge to `main`.
- Outside this repository: update `rev`, `hash` and `vendorHash` in the
  Nix package that builds this server, replace its "temporary fork pin"
  comment, and verify the system build.

## Order and dependencies

1, 2 and 6 are independent and go first (1 and 2 are the dangerous
ones). 3 and 5 both edit the sync and write path and go one after the
other. 4 and 7 are independent. 8 depends on all of them.

## Testing

Each fix adds tests to the existing test files and is verified with
`go test ./...` in the server, and in `things-cloud-sdk/` for fix 7.
Nothing here is tested against a live Things account; writes to a real
account stay a manual check with a disposable account.

## Evidence

- Code revision: `b03d51d` (`main`), plus the documentation commit
  `1b3a5f6` on `standalone-docs`. No other uncommitted changes.
- Analysed: `main.go`, `go.mod`, `things-cloud-sdk/{base58,client,histories,types}.go`,
  `things-cloud-sdk/sync/sync.go`; upstream SDK at v0.6.0 (`6919edc`).
- Seam analysis: graph index of this repository and of the upstream SDK,
  2026-11-24.
- Patch checks: PRs #18 and #26 apply cleanly; PR #25 applies except its
  README hunk.
- Gaps: whether the existing test helpers can fake a 409 and a 404 from
  Things Cloud, or fixes 3 and 5 need a small fake server added.

## Review record

Review mode: Lightweight — user selected Ponytail only (2026-11-24) |
Scope: fixes inside existing functions of one package and one SDK client
function, no tool contract changes | Review status: Ponytail design
review completed 2026-11-24 against the code at `b03d51d` (no
implementation tests exist yet) | Final approval: given by the user 2026-11-24

Findings and dispositions:

1. Major — validating every identifier would block edits of items with
   legacy identifiers. Fixed: only creations are validated (fix 1).
2. Major — retrying after 409 without re-running the sort-index guard
   could send a commit based on stale state. Fixed: the whole pre-write
   sequence repeats (fix 3).
3. Major — rebuilding on every 404 AccountIssue would refetch the full
   history on each call if the error persists. Fixed: rebuild only when
   the own-history ID has changed (fix 5).
4. Nit — PR #25's `TZ` lookup duplicates what `time.Local` does. Fixed:
   removed (fix 4).
5. Nit — fix 7 changes the vendored SDK and will be redone on the SDK
   integration branch later. Accepted: it is small and becomes the first
   upstream pull request.
