# Working with the Things Cloud SDK

This server depends on
[arthursoares/things-cloud-sdk](https://github.com/arthursoares/things-cloud-sdk)
and tracks it. SDK bugs are fixed in the SDK and sent upstream. A release
of this server never waits for an upstream merge.

## Where a fix belongs

- **SDK:** wire format, item kinds, history fetch and commit, HTTP client
  behaviour (errors, retries), identifier encoding.
- **Server (this repo):** MCP tools, filters and views, the per-user sync
  loop in `main.go`, OAuth, anything about how results are presented.

If a server fix needs something the SDK does not expose (a typed error,
for example), add that to the SDK first and use it from here.

## Layout

- Our SDK fork has one long-lived branch, `integration`: upstream `main`
  plus the patches upstream has not merged yet.
- `go.mod` here pins that branch by commit:

  ```
  replace github.com/arthursoares/things-cloud-sdk => github.com/<owner>/things-cloud-sdk <commit>
  ```

- What we carry is always visible with
  `git log upstream/main..integration` in the SDK fork.

## Shipping an SDK fix

1. In the SDK fork, branch off upstream `main`. Make the fix with a test.
2. Open a pull request upstream from that branch.
3. Merge the same branch into `integration`.
4. Bump the pinned commit in this repo's `go.mod`, run `go test ./...`,
   and release.

## When upstream responds

- **Merged:** rebase `integration` onto upstream `main`. The patch drops
  out. Bump the pin here.
- **Changes requested:** update the pull request branch, merge it into
  `integration` again, bump the pin.
- **Declined or unanswered:** the patch stays on `integration`. Record the
  reason in its commit message.
- **`integration` matches upstream:** remove the `replace` line and
  require a tagged upstream release.

## Transition

Until the vendored `things-cloud-sdk/` directory has been converged onto
upstream, `go.mod` still points at that directory and SDK changes are
made there. Convergence means starting `integration` from the latest
upstream release and re-applying each change that exists only in the
vendored copy as its own commit and upstream pull request. After that,
the vendored directory is deleted and the layout above applies.
