# Refund untuk Jamaah Gagal Berangkat Implementation Plan

> **For agentic workers:** Steps use checkbox (`- [ ]`) syntax for tracking. Work task-by-task, in order — Task 3 (cascade) assumes Task 1's `payment_method` column exists if you also touch refund creation paths, and Task 4 (docs) should be written last since it documents the final route/behavior surface.

**Goal:** (1) Fix refund journal entries always posting to Bank even for cash payments. (2) Add a working "Ajukan Refund" button to the web app (currently only exists on mobile). (3) Auto-cascade cancel-invoice + initiate-refund when a jamaah is marked "Batal berangkat" in the CRM, without moving any money automatically. (4) Document the `/refunds/*` API.

**Architecture:** See `docs/superpowers/specs/2026-07-01-refund-gagal-berangkat-design.md` for full rationale. Summary: refund's `payment_method` is inherited from the invoice's most recent payment at initiation time and threaded through to the accounting outbox payload; the CRM's existing lost-reason picker gains a stable `lost_reason_code` sent alongside the display label, and `JamaahService.UpdatePipelineStatus` uses it to best-effort call invoice-service's existing cancel/refund endpoints.

**Tech Stack:** Go (Fiber, pgx), PostgreSQL, NATS JetStream (accounting outbox — untouched by this plan), Svelte 5 (runes).

## Global Constraints

- Spec: `docs/superpowers/specs/2026-07-01-refund-gagal-berangkat-design.md`.
- Cascade (Task 3) is **best-effort and never blocks the pipeline-status update**: log failures, report them back via a `cascade` result object, never return an error to the CRM just because invoice-service was unreachable or the caller lacked the finance role.
- Cascade **never completes a refund** — it only creates a `pending` one. Money only moves when finance clicks "Tandai Selesai" through the existing approve→process→complete flow. Do not add any code that auto-approves/processes/completes.
- Commit messages: **no AI co-author line** (repo convention).
- Go build/test from repo root: `go -C "/media/muklis/Data/Codding/suluk" build ./...` / `go -C "/media/muklis/Data/Codding/suluk" test ./...`. Frontend: `cd "/media/muklis/Data/Codding/suluk/frontend-svelte" && npm run build`.

---

### Task 1: Fix refund payment_method bug (Kas vs Bank misclassification)

**Files:**
- Create: `migrations/invoice/008_refund_payment_method.up.sql`, `.down.sql`
- Modify: `internal/invoice/model/model.go`, `internal/invoice/repository/refund.go`, `internal/invoice/service/refund.go`
- Test: `internal/accounting/service/posting_test.go`

**Interfaces:**
- Produces: `model.Refund.PaymentMethod string`. Consumed by Task 3 not at all (cascade doesn't touch this field directly, it flows automatically through `InitiateRefund`).

- [ ] **Step 1: Migration**

`migrations/invoice/008_refund_payment_method.up.sql`:
```sql
-- Refunds need to know which cash/bank account to credit back at CompleteRefund
-- time; previously this was always missing from the accounting payload, so
-- every refund posted as a Bank outflow even when the original payment was
-- cash (tunai).
ALTER TABLE refunds ADD COLUMN payment_method VARCHAR(30) NOT NULL DEFAULT 'transfer_bank';
```

`migrations/invoice/008_refund_payment_method.down.sql`:
```sql
ALTER TABLE refunds DROP COLUMN payment_method;
```

- [ ] **Step 2: Add `PaymentMethod` to the `Refund` model**

In `internal/invoice/model/model.go`, in the `Refund` struct (currently starts `type Refund struct {`), add a field after `RefundPct`:
```go
	PaymentMethod string     `json:"payment_method" db:"payment_method"`
```

- [ ] **Step 3: Persist and read back `payment_method` in the repository**

In `internal/invoice/repository/refund.go`, change `CreateRefund`:
```go
func (r *InvoiceRepo) CreateRefund(ctx context.Context, ref *model.Refund) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO refunds (org_id, invoice_id, amount, refund_pct, reason, notes, payment_method, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')
		RETURNING id, created_at, updated_at
	`, ref.OrgID, ref.InvoiceID, ref.Amount, ref.RefundPct, ref.Reason, ref.Notes, ref.PaymentMethod).Scan(&ref.ID, &ref.CreatedAt, &ref.UpdatedAt)
}
```

Change `CompleteRefund`'s initial SELECT and payload marshal (keep everything else in the function identical):
```go
	// Confirm the refund is processed and capture its amount + invoice.
	var invoiceID uuid.UUID
	var amount int64
	var paymentMethod string
	if err := tx.QueryRow(ctx, `SELECT invoice_id, amount, payment_method FROM refunds WHERE id=$1 AND org_id=$2 AND status='processed' FOR UPDATE`,
		id, orgID).Scan(&invoiceID, &amount, &paymentMethod); err != nil {
		return fmt.Errorf("refund not in processed status")
	}
```
```go
	// Emit refund.completed so accounting posts Dr Piutang / Cr Kas|Bank.
	payload, _ := json.Marshal(map[string]any{"amount": amount, "invoice_number": invNumber, "payment_method": paymentMethod})
```

- [ ] **Step 4: Default `payment_method` from the invoice's most recent payment**

In `internal/invoice/service/refund.go`, change `InitiateRefund`:
```go
func (s *RefundService) InitiateRefund(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID, req model.InitiateRefundRequest) (*model.Refund, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Amount > inv.AmountPaid {
		return nil, repository.ErrRefundExceedsPaid
	}
	// GetPayments orders by paid_at DESC, so [0] is the most recent payment —
	// that's the account the refund should come back out of.
	paymentMethod := "transfer_bank"
	if payments, err := s.repo.GetPayments(ctx, invoiceID); err == nil && len(payments) > 0 {
		paymentMethod = payments[0].PaymentMethod
	}
	ref := &model.Refund{
		OrgID:         orgID,
		InvoiceID:     invoiceID,
		Amount:        req.Amount,
		RefundPct:     req.RefundPct,
		Reason:        req.Reason,
		Notes:         req.Notes,
		PaymentMethod: paymentMethod,
		Status:        "pending",
	}
	if err := s.repo.CreateRefund(ctx, ref); err != nil {
		return nil, err
	}
	return ref, nil
}
```

- [ ] **Step 5: Add regression test cases to `posting_test.go`**

In `internal/accounting/service/posting_test.go`, inside `TestBuildPostingBalanced`'s `cases` slice, add two new cases (placed anywhere in the slice, e.g. right after the `"payment tunai -> kas"` case):
```go
			{
				name: "refund tunai -> kas",
				env:  mk(events.EventRefundCompleted, map[string]any{"amount": 200000, "payment_method": "tunai", "invoice_number": "INV-3"}),
				wantAcc: map[string][2]int64{
					AccPiutangJemaah: {200000, 0},
					AccKas:           {0, 200000},
				},
			},
			{
				name: "refund transfer -> bank",
				env:  mk(events.EventRefundCompleted, map[string]any{"amount": 200000, "payment_method": "transfer_bank", "invoice_number": "INV-4"}),
				wantAcc: map[string][2]int64{
					AccPiutangJemaah: {200000, 0},
					AccBank:          {0, 200000},
				},
			},
```

- [ ] **Step 6: Build + test**

Run: `go -C "/media/muklis/Data/Codding/suluk" build ./... && go -C "/media/muklis/Data/Codding/suluk" test ./internal/accounting/... ./internal/invoice/...`
Expected: clean build, all tests pass including the two new cases.

- [ ] **Step 7: Commit**

```bash
git add migrations/invoice/008_refund_payment_method.up.sql migrations/invoice/008_refund_payment_method.down.sql internal/invoice/model/model.go internal/invoice/repository/refund.go internal/invoice/service/refund.go internal/accounting/service/posting_test.go
git commit -m "fix(invoice): refund journals credit Kas for cash payments, not always Bank"
```

---

### Task 2: "Ajukan Refund" button on the web app

**Files:**
- Modify: `frontend-svelte/src/lib/pages/CancellationPage.svelte`

**Interfaces:**
- Consumes: `ApiService.listInvoices` (existing, `invoiceApi.js:11`), `ApiService.initiateRefund` (existing, `refundApi.js`, currently only called from mobile).

- [ ] **Step 1: Load invoices alongside refunds/policies**

In `CancellationPage.svelte`, add a new state var near the other `showNewRefundDrawer`/`refundForm` declarations (line ~34-36):
```js
  let invoices = $state([]);
```

Change `loadData()` to also fetch invoices:
```js
  async function loadData() {
    loading = true;
    try {
      const [refundData, policyData, invoiceData] = await Promise.all([
        ApiService.listRefunds({ status: statusFilter === 'all' ? '' : statusFilter }),
        ApiService.listPolicies(),
        ApiService.listInvoices({ pageSize: 100 }).catch(() => null),
      ]);
      refunds = refundData.refunds || [];
      total = refundData.total || 0;
      policies = policyData.policies || [];
      const invList = invoiceData?.invoices || invoiceData?.data || (Array.isArray(invoiceData) ? invoiceData : []) || [];
      invoices = invList.filter((i) => i.status !== 'batal' && (i.amount_paid ?? 0) > 0);
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      loading = false;
    }
  }
```

- [ ] **Step 2: Add open/select/save handlers**

Add after `openRefundDetail`/before `refundAction` (or anywhere in the `<script>` block):
```js
  function openNewRefund() {
    refundForm = { invoice_id: '', amount: 0, refund_pct: 100, reason: '', notes: '' };
    showNewRefundDrawer = true;
  }

  function selectInvoiceForRefund(id) {
    refundForm.invoice_id = id;
    const inv = invoices.find((i) => i.id === id);
    refundForm.amount = inv?.amount_paid ?? 0;
  }

  async function saveNewRefund() {
    if (!refundForm.invoice_id || !refundForm.amount) {
      showToast('Pilih invoice dan pastikan ada nominal yang dibayar', 'error');
      return;
    }
    savingRefund = true;
    try {
      await ApiService.initiateRefund(refundForm.invoice_id, {
        amount: Number(refundForm.amount),
        refund_pct: Number(refundForm.refund_pct) || 100,
        reason: refundForm.reason,
        notes: refundForm.notes,
      });
      showToast('Pengajuan refund dibuat');
      showNewRefundDrawer = false;
      await loadData();
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingRefund = false;
    }
  }
```

- [ ] **Step 3: Add the trigger button**

In the template, change the `PageHeader` actions snippet (currently only has the "Kebijakan" button):
```svelte
    {#snippet actions()}
      <Button variant="ghost" icon={Plus} onclick={openNewPolicy}>Kebijakan</Button>
      <Button variant="primary" icon={Plus} onclick={openNewRefund}>Ajukan Refund</Button>
    {/snippet}
```

- [ ] **Step 4: Add the drawer**

Add a new `<SlideDrawer>` block after the closing `</SlideDrawer>` of "Refund Detail Drawer" (end of file):
```svelte
<!-- New Refund Drawer -->
<SlideDrawer open={showNewRefundDrawer} onClose={() => showNewRefundDrawer = false} title="Ajukan Refund Baru" width="480px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1">
      <label for="ref-invoice" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Invoice</label>
      <select id="ref-invoice" value={refundForm.invoice_id} onchange={(e) => selectInvoiceForRefund(e.target.value)} class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)">
        <option value="">Pilih invoice...</option>
        {#each invoices as inv}
          <option value={inv.id}>{(inv.jamaah_name || inv.invoice_number || inv.id)} · {formatIDR(inv.amount_paid)}</option>
        {/each}
      </select>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="ref-amount" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Jumlah Refund (Rp)</label>
        <input id="ref-amount" type="number" bind:value={refundForm.amount} class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)" />
      </div>
      <div class="flex flex-col gap-1">
        <label for="ref-pct" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Persentase (%)</label>
        <input id="ref-pct" type="number" bind:value={refundForm.refund_pct} min="0" max="100" step="0.01" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)" />
      </div>
    </div>
    <div class="flex flex-col gap-1">
      <label for="ref-reason" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Alasan</label>
      <textarea id="ref-reason" bind:value={refundForm.reason} rows="2" placeholder="Contoh: Jamaah gagal berangkat" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)"></textarea>
    </div>
    <div class="flex flex-col gap-1">
      <label for="ref-notes" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Catatan (opsional)</label>
      <textarea id="ref-notes" bind:value={refundForm.notes} rows="2" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)"></textarea>
    </div>
    <div class="flex gap-2 pt-2">
      <Button variant="ghost" full onclick={() => showNewRefundDrawer = false}>Batal</Button>
      <Button variant="primary" full disabled={savingRefund} onclick={saveNewRefund}>{savingRefund ? 'Menyimpan...' : 'Ajukan'}</Button>
    </div>
  </div>
</SlideDrawer>
```

- [ ] **Step 5: Build**

Run: `cd "/media/muklis/Data/Codding/suluk/frontend-svelte" && npm run build`
Expected: clean build, no Svelte compiler errors/warnings about unused/undefined state.

- [ ] **Step 6: Manual smoke test**

Run dev server (`npm run dev`), open the "Pembatalan" page, click "Ajukan Refund", pick an invoice with `amount_paid > 0`, submit, confirm it appears in the table with status "Pending" and can be approved/processed/completed through the existing detail drawer.

- [ ] **Step 7: Commit**

```bash
git add frontend-svelte/src/lib/pages/CancellationPage.svelte
git commit -m "feat(web): add missing 'Ajukan Refund' button/drawer to Pembatalan page"
```

---

### Task 3: Cascade — "Batal berangkat" auto-cancels invoice + initiates refund

**Files:**
- Modify: `internal/jamaah/model/model.go`, `internal/jamaah/service/service.go`, `internal/jamaah/handler/handler.go`
- Modify: `frontend-svelte/src/lib/services/apiDomains/jamaahApi.js`, `frontend-svelte/src/lib/pages/CRMPage.svelte`

**Interfaces:**
- Produces: `service.CascadeResult{InvoiceCancelled, RefundInitiated, Attempted bool}`; `(*JamaahService) UpdatePipelineStatus(ctx, orgID, userID, jamaahID, packageID uuid.UUID, status, reason, lostReason, lostReasonCode, authToken string) (*model.JamaahPackageRegistration, CascadeResult, error)`.
- Consumes: Task 1's `payment_method` column indirectly (not directly touched here — cascade just calls the existing `/invoices/:id/refund` endpoint which now defaults `payment_method` correctly per Task 1).

- [ ] **Step 1: Add `lost_reason_code` to the request model**

In `internal/jamaah/model/model.go`, change `UpdatePipelineStatusRequest`:
```go
type UpdatePipelineStatusRequest struct {
	PipelineStatus string `json:"pipeline_status" validate:"required,oneof=prospek survey booking dp cicilan lunas berangkat selesai batal"`
	Reason         string `json:"reason,omitempty" validate:"max=255"`
	LostReason     string `json:"lost_reason,omitempty" validate:"max=40"`
	LostReasonCode string `json:"lost_reason_code,omitempty" validate:"max=20"`
}
```

- [ ] **Step 2: Add `CascadeResult` and the cascade function**

In `internal/jamaah/service/service.go`, add near the top of the file (after imports, before `JamaahService` struct or after `UpdatePipelineStatus` — either is fine, group logically with `UpdatePipelineStatus`):
```go
// CascadeResult reports what the "jamaah gagal berangkat" cascade actually
// did, so the CRM can tell staff whether finance still needs to act manually
// (e.g. the caller lacked the finance role, or invoice-service errored).
type CascadeResult struct {
	InvoiceCancelled bool `json:"invoice_cancelled"`
	RefundInitiated  bool `json:"refund_initiated"`
	Attempted        bool `json:"attempted"`
}

// jamaahCascadeInvoice is the subset of invoice fields the cascade needs from
// invoice-service's GET /api/v1/invoices/jamaah/:id response.
type jamaahCascadeInvoice struct {
	ID              uuid.UUID `json:"id"`
	PackageID       uuid.UUID `json:"package_id"`
	Status          string    `json:"status"`
	AmountPaid      int64     `json:"amount_paid"`
	AmountRemaining int64     `json:"amount_remaining"`
}

// cascadeGagalBerangkat runs when a registration is marked "batal" with lost
// reason code "tidak_jadi" (jamaah gagal berangkat). It looks up the
// registration's invoice and, best-effort, cancels the uncollected remainder
// and files a pending refund request for whatever was already paid, so staff
// no longer has to remember the separate manual cancel + refund steps. Every
// step is best-effort and logged rather than returned as an error: a failure
// here (e.g. the caller isn't owner/admin/finance, so invoice-service 403s)
// must never fail the pipeline-status update that already succeeded.
func (s *JamaahService) cascadeGagalBerangkat(ctx context.Context, jamaahID, packageID uuid.UUID, authToken string) CascadeResult {
	var invoices []jamaahCascadeInvoice
	if err := s.httpc.GetJSON(ctx, s.invoiceAddr, "/api/v1/invoices/jamaah/"+jamaahID.String(), authToken, &invoices); err != nil {
		if s.log != nil {
			s.log.Warnw("cascade gagal berangkat: list invoices", "jamaah_id", jamaahID, "err", err)
		}
		return CascadeResult{}
	}

	var inv *jamaahCascadeInvoice
	for i := range invoices {
		if invoices[i].PackageID == packageID && invoices[i].Status != "batal" {
			inv = &invoices[i]
			break
		}
	}
	if inv == nil {
		return CascadeResult{}
	}

	result := CascadeResult{Attempted: true}
	headers := map[string]string{"Authorization": authToken}

	if inv.AmountRemaining > 0 {
		body := map[string]string{"reason": "Jamaah gagal berangkat"}
		if err := s.httpc.PostJSON(ctx, s.invoiceAddr, "/api/v1/invoices/"+inv.ID.String()+"/cancel", headers, body, nil); err != nil {
			if s.log != nil {
				s.log.Warnw("cascade gagal berangkat: cancel invoice", "invoice_id", inv.ID, "err", err)
			}
		} else {
			result.InvoiceCancelled = true
		}
	}

	if inv.AmountPaid > 0 {
		body := map[string]any{"amount": inv.AmountPaid, "refund_pct": 100, "reason": "Jamaah gagal berangkat"}
		if err := s.httpc.PostJSON(ctx, s.invoiceAddr, "/api/v1/invoices/"+inv.ID.String()+"/refund", headers, body, nil); err != nil {
			if s.log != nil {
				s.log.Warnw("cascade gagal berangkat: initiate refund", "invoice_id", inv.ID, "err", err)
			}
		} else {
			result.RefundInitiated = true
		}
	}

	return result
}
```

- [ ] **Step 3: Wire the cascade into `UpdatePipelineStatus`**

In `internal/jamaah/service/service.go`, change the `UpdatePipelineStatus` signature and its two `return nil, err` early-exits plus the final return:
```go
func (s *JamaahService) UpdatePipelineStatus(ctx context.Context, orgID, userID, jamaahID, packageID uuid.UUID, status, reason, lostReason, lostReasonCode, authToken string) (*model.JamaahPackageRegistration, CascadeResult, error) {
	reg, err := s.repo.GetRegistration(ctx, orgID, jamaahID, packageID)
	if err != nil {
		return nil, CascadeResult{}, err
	}

	// Gate advancing to lunas/berangkat on document completeness (+ mahram).
	if err := s.checkTransitionGate(ctx, orgID, jamaahID, status, reg); err != nil {
		return nil, CascadeResult{}, err
	}

	dpDate, lunasDate, berangkatDate := reg.DPDate, reg.LunasDate, reg.BerangkatDate
	now := time.Now()
	switch status {
	case string(model.StatusProspek):
		dpDate = nil
		lunasDate = nil
		berangkatDate = nil
	case string(model.StatusDP):
		dpDate = &now
		lunasDate = nil
		berangkatDate = nil
	case string(model.StatusLunas):
		dpDate = reg.DPDate
		lunasDate = &now
		berangkatDate = nil
	case string(model.StatusBerangkat):
		dpDate = reg.DPDate
		lunasDate = reg.LunasDate
		berangkatDate = &now
	}

	// lost_reason is only meaningful when the lead is lost.
	if status != string(model.StatusBatal) {
		lostReason = ""
		lostReasonCode = ""
	}
	if err := s.repo.UpdatePipelineStatus(ctx, orgID, jamaahID, packageID, repository.PipelineUpdate{
		Status:        status,
		FromStatus:    reg.PipelineStatus,
		DPDate:        dpDate,
		LunasDate:     lunasDate,
		BerangkatDate: berangkatDate,
		LostReason:    lostReason,
		Reason:        reason,
		ChangedBy:     userID,
	}); err != nil {
		return nil, CascadeResult{}, err
	}
	s.recompute(ctx, orgID, jamaahID, packageID) // stage change moves the score

	var cascade CascadeResult
	if status == string(model.StatusBatal) && lostReasonCode == "tidak_jadi" {
		cascade = s.cascadeGagalBerangkat(ctx, jamaahID, packageID, authToken)
	}

	updated, err := s.repo.GetRegistration(ctx, orgID, jamaahID, packageID)
	return updated, cascade, err
}
```

(Only the signature, the `lostReasonCode = ""` reset line, the three `CascadeResult{}` additions to early returns, and the new cascade block + final return are new — the DP/lunas/berangkat date logic in the middle is unchanged, reproduced above only so the diff context is unambiguous.)

- [ ] **Step 4: Update the handler**

In `internal/jamaah/handler/handler.go`, change `UpdatePipelineStatus`:
```go
	reg, cascade, err := h.svc.UpdatePipelineStatus(c.Context(), claims.OrgID, claims.UserID, jamaahID, packageID, req.PipelineStatus, req.Reason, req.LostReason, req.LostReasonCode, c.Get("Authorization"))
	if err != nil {
		if errors.Is(err, service.ErrGate) {
			return response.BadRequest(c, err.Error())
		}
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"registration": reg, "cascade": cascade})
```

- [ ] **Step 5: Build**

Run: `go -C "/media/muklis/Data/Codding/suluk" build ./...`
Expected: clean build (only call site of the changed signature is the handler, already updated).

- [ ] **Step 6: Send `lost_reason_code` from the frontend**

In `frontend-svelte/src/lib/services/apiDomains/jamaahApi.js`, change `updatePipelineStatus`:
```js
    async updatePipelineStatus(jamaahId, packageId, { pipeline_status, reason = '', lost_reason = '', lost_reason_code = '' }) {
      const response = await apiFetch(`${API_URL}/jamaah/${jamaahId}/registrations/${packageId}/status`, {
        method: 'PATCH',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify({ pipeline_status, reason, lost_reason, lost_reason_code }),
      });
      if (!response.ok) throw new Error(await parseError(response));
      return unwrapData(await response.json());
    },
```

- [ ] **Step 7: Track and send the reason code from the CRM lost-reason picker**

In `frontend-svelte/src/lib/pages/CRMPage.svelte`, add a state var next to `let lostReason = $state('');` (line ~51):
```js
  let lostReasonCode = $state('');
```

Change the lost-reason button click handler (line ~892, currently `onclick={() => (lostReason = r.label)}`):
```svelte
            onclick={() => { lostReason = r.label; lostReasonCode = r.id; }}
```

Change `confirmBatal()` (line ~337-343) to pass the code through, and reset it alongside `lostReason = ''` (line ~320):
```js
  function confirmBatal() {
    if (!pendingBatal) return;
    const { idx, packageId } = pendingBatal;
    showLostModal = false;
    persistStage(idx, packageId, 'batal', lostReason || '', lostReasonCode || '');
    pendingBatal = null;
  }
```

Change `persistStage` to accept and forward the code, and to surface the cascade outcome:
```js
  async function persistStage(idx, packageId, toCol, lost_reason, lost_reason_code = '') {
    const prev = jamaah[idx];
    jamaah[idx] = { ...prev, pipeline_status: toCol }; // optimistic
    try {
      const result = await ApiService.updatePipelineStatus(prev.id, packageId, { pipeline_status: toCol, lost_reason, lost_reason_code });
      showToast(`Lead dipindahkan ke ${stageLabel(toCol)}`, 'success');
      if (result?.cascade?.invoice_cancelled || result?.cascade?.refund_initiated) {
        showToast('Invoice dibatalkan & refund otomatis diajukan — cek menu Pembatalan', 'success');
      } else if (result?.cascade?.attempted) {
        showToast('Jamaah gagal berangkat: proses cancel & refund manual di menu Pembatalan', 'warning');
      }
    } catch (e) {
      jamaah[idx] = prev; // rollback
      showToast(mapError(e.message), 'error');
    }
  }
```

- [ ] **Step 8: Build**

Run: `cd "/media/muklis/Data/Codding/suluk/frontend-svelte" && npm run build`
Expected: clean build.

- [ ] **Step 9: Manual smoke test**

Using an owner/admin/finance-role login: create/find a jamaah with a package registration that has an invoice with `amount_paid > 0`. In CRM, drag it to "Batal", pick "Batal berangkat", confirm. Verify: toast says invoice cancelled + refund initiated; open menu Pembatalan and confirm a new `pending` refund appears for that invoice with the correct amount. Then log in as a non-finance role and repeat — verify pipeline status still updates and the toast says to process manually, with no new refund created.

- [ ] **Step 10: Commit**

```bash
git add internal/jamaah/model/model.go internal/jamaah/service/service.go internal/jamaah/handler/handler.go frontend-svelte/src/lib/services/apiDomains/jamaahApi.js frontend-svelte/src/lib/pages/CRMPage.svelte
git commit -m "feat(jamaah): cascade cancel invoice + initiate refund when jamaah gagal berangkat"
```

---

### Task 4: Document `/refunds/*` in API_REFERENCE.md

**Files:**
- Modify: `API_REFERENCE.md`

**Interfaces:** none (docs only).

- [ ] **Step 1: Add the Refunds section**

In `API_REFERENCE.md`, insert a new `## Refunds` section immediately after the `## Invoices` section (which currently ends with the `/invoices/:id/payments` row, right before `## Finance / Expenses`):
```markdown
## Refunds

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/invoices/:id/refund` | Finance | Initiate a refund for an invoice |
| GET | `/refunds/` | Yes | List refunds (paginated, filterable by status) |
| GET | `/refunds/:id` | Yes | Get refund detail |
| GET | `/refunds/by-invoice/:id` | Yes | List refunds for an invoice |
| PUT | `/refunds/:id/approve` | Finance | Approve a pending refund |
| PUT | `/refunds/:id/process` | Finance | Mark an approved refund as processed |
| PUT | `/refunds/:id/complete` | Finance | Complete a refund (posts the accounting journal) |
| PUT | `/refunds/:id/reject` | Finance | Reject a pending refund |
| GET | `/refunds/policies` | Yes | List refund policies |
| POST | `/refunds/policies` | Finance | Create a refund policy |
| PUT | `/refunds/policies/:id` | Finance | Update a refund policy |
| DELETE | `/refunds/policies/:id` | Finance | Delete a refund policy |
```
("Finance" auth = `RequireRole(owner, admin, finance)`, matching the style already used elsewhere in this doc if such a distinction exists — otherwise fall back to "Yes" for all rows and add one line above the table noting which routes require the finance role, whichever matches the existing convention in the surrounding sections once you check them.)

- [ ] **Step 2: Commit**

```bash
git add API_REFERENCE.md
git commit -m "docs: document /refunds API endpoints"
```

---

## Self-Review

- **Spec coverage:** payment_method bug fixed at the source (InitiateRefund default + CompleteRefund payload) with regression tests (Task 1) ✓; web "Ajukan Refund" button/drawer wired to existing dead state + existing API client (Task 2) ✓; cascade triggers only on `lost_reason_code=="tidak_jadi"`, best-effort, never auto-completes money movement, reports result back to CRM (Task 3) ✓; docs (Task 4) ✓. All of spec §3-§7 mapped to a task.
- **Placeholder scan:** every step shows real Go/SQL/Svelte matching the exact current file contents read from the repo (function bodies, line context, existing field names) — no invented APIs. Migration numbered `003` (next after `002_refund`).
- **Type consistency:** `UpdatePipelineStatus` signature change (Task 3, Step 3) has exactly one call site (the handler, Step 4) — verified via repo-wide grep before writing this plan. `CascadeResult` produced in Step 2, returned in Step 3, consumed by handler in Step 4, consumed by frontend in Step 7. `payment_method` produced in Task 1 Step 4, consumed in Step 3 (CompleteRefund read-back) — consistent.
- **DB-test honesty:** repo/SQL layer (Task 1 migration + queries, Task 3 cascade HTTP calls) is build-verified + integration only, per repo convention (no test DB available). The one piece of pure, deterministic logic at risk of silent regression — the Kas-vs-Bank account branch in `posting.go` — already had zero test coverage for the refund event specifically; Task 1 Step 5 closes that gap.
- **Backward compatibility:** `payment_method` column defaults `'transfer_bank'` (matches `payments.payment_method` table default) so existing rows need no backfill; `UpdatePipelineStatusRequest.LostReasonCode` and the handler's `cascade` response field are both additive (omitempty / new field), so any client not yet updated keeps working unchanged.
