# Kloter Cancel Cascade Implementation Plan

**Goal:** Wire the existing per-jamaah `cascadeGagalBerangkat` cascade into kloter (group) cancellation, so cancelling a kloter cancels+refunds every member instead of leaving their invoices in silent limbo (finding C5).

**Architecture:** See `docs/superpowers/specs/2026-07-02-kloter-cancel-cascade-design.md`. No new cancel/refund logic — a bounded-concurrency loop over the group's members calling the already-built, already-tested `cascadeGagalBerangkat` once per member.

## Global Constraints

- Reuses `cascadeGagalBerangkat` verbatim — do not duplicate its cancel/refund logic.
- Cascade only runs for `to == DepartureBatal`, only when the group has a package and members.
- Money never moves automatically — same guarantee as the individual cascade (cascade only cancels uncollected balance + creates a `pending` refund; finance still approves/processes/completes manually).
- Commit messages: no AI co-author line.

---

### Task 1: `GroupCascadeSummary` + bounded-concurrency loop

**Files:**
- Modify: `internal/jamaah/service/service.go`

**Interfaces:**
- Produces: `GroupCascadeSummary{MembersProcessed, InvoicesCancelled, RefundsInitiated int}`; `(*JamaahService) cascadeGroupCancelled(ctx, members []model.GroupMember, packageID uuid.UUID, authToken string) GroupCascadeSummary`.
- Consumes: existing `cascadeGagalBerangkat(ctx, jamaahID, packageID uuid.UUID, authToken string) CascadeResult` — unchanged.

- [ ] Add the summary type and the loop, capped at 5 concurrent workers via a buffered-channel semaphore + `sync.WaitGroup` + `sync.Mutex` for the aggregate counts (stdlib only — `golang.org/x/sync` is an indirect dependency today, not worth promoting to direct for this).
- [ ] `go build ./... && go vet ./...`
- [ ] Commit.

### Task 2: Wire into `TransitionDeparture`

**Files:**
- Modify: `internal/jamaah/service/service.go` (`TransitionDeparture`), `internal/jamaah/handler/departure.go`

**Interfaces:**
- `(*JamaahService) TransitionDeparture(ctx, groupID, orgID uuid.UUID, status, authToken string) (*model.Group, GroupCascadeSummary, error)` — signature grows by one param + one return value, same shape as yesterday's `UpdatePipelineStatus` change.
- Handler forwards `c.Get("Authorization")`, response becomes `{"group": g, "cascade": summary}`.

- [ ] Grep repo-wide for other callers of `TransitionDeparture` before changing the signature (expect only the handler, same as yesterday's `UpdatePipelineStatus` — verify, don't assume).
- [ ] After a successful `to == DepartureBatal` transition, if `g.PackageID != nil && g.MemberCount > 0`: `members, err := s.repo.ListGroupMembers(ctx, groupID)` (best-effort — log and skip cascade on error, don't fail the whole transition since the status change already committed), then `summary = s.cascadeGroupCancelled(ctx, members, *g.PackageID, authToken)`.
- [ ] `go build ./... && go vet ./... && go test ./...`
- [ ] Commit.

### Task 3: Frontend — confirm before cancelling, show cascade summary

**Files:**
- Modify: `frontend-svelte/src/lib/pages/GroupsPage.svelte`

- [ ] `transitionDep(status)`: when `status === 'batal'`, `confirm()` first, message mentions member count (e.g. "Batalkan kloter ini beserta {N} jamaah di dalamnya?"). Bail if declined.
- [ ] After a successful `batal` transition, if `result.cascade` has any non-zero counts, show a toast summarizing it (invoices cancelled / refunds initiated); otherwise keep the existing generic status-change toast.
- [ ] `npm run build`
- [ ] Commit.

## Self-Review
- Reuses existing, already-tested cascade — no new cancel/refund logic, no new money-movement risk introduced.
- Signature change on `TransitionDeparture` verified against actual call sites before editing (same discipline as yesterday's `UpdatePipelineStatus` change).
- Bounded concurrency prevents large kloter from making the request hang; stdlib only, no new dependency.
