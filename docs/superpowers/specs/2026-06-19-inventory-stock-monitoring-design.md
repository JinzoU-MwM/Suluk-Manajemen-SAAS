# Inventory Stock Monitoring + Auto-Deduct on Departure — Design

- **Date:** 2026-06-19
- **Status:** Approved — proceeding to implementation plan (v1)
- **Area:** `inventory-service` (Go) + `frontend-svelte` Inventaris page
- **Related events:** `group.departed` (already emitted by jamaah-service)

## 1. Context & problem

Today the Inventaris module is built entirely on `member_equipment`: a per-group
equipment workflow (forecast koper/ihram/mukena, mark-received per member, QR
handover at equipment/luggage checkpoints). It answers *"who received what."*

It does **not** answer *"how much stock do we have left."* The DB already has a
dormant `inventory_items` table (`name, category, unit, stock, min_stock`) and a
matching `InventoryItem` Go struct, but **no API, no UI, and no movement
history** — the stock side was scaffolded and never built.

We want stock monitoring: staff add umroh stock (koper, seragam, kain ihram,
…), track levels, and have stock **auto-deducted when a group departs**.

## 2. Goals

- Add stock-level monitoring **alongside** the existing distribution/QR workflow
  (nothing existing is removed).
- Let staff add stock items, restock, adjust, and see current levels + history.
- Define a **departure kit per package** (items + quantity per jamaah).
- On group departure, **automatically deduct** `qty_per_jamaah × member_count`
  for each kit item, recorded in an auditable ledger.
- Surface low-stock and negative-stock as alerts.

## 3. Non-goals for v1 — planned for v2

Deliberately deferred to keep v1 focused. These are slated for a **v2** iteration;
v1 should leave room for them but not build them:

- **Auto-reversal** when a departure is undone (`berangkat` → back) — restore the
  deducted stock automatically. v1 corrects via a manual adjustment in the ledger.
- **Purchase orders / supplier management / costing** (stock valuation, reorder POs).
- **Handover-based deduction** — deduct from `received_items` at QR scan, as an
  alternative/complement to departure-driven deduction.
- **Per-warehouse / multi-location stock** — v1 is a single org-wide pool.

The v1 data model (append-only movement ledger, signed deltas, `reason` enum) is
intentionally general enough to extend for all of these without a rewrite — e.g.
a `reversal` reason, a `handover` reason, or a `location_id` column.

## 4. Decisions (locked in brainstorming)

| Topic | Decision |
|-------|----------|
| Existing workflow | **Keep**; add stock monitoring as a second tab. |
| Deduct rule | **Named kit per package**: `qty_per_jamaah × member_count`. |
| History | **Full movement ledger** (`stock_movements`). |
| Shortfall | **Allow negative** stock + loud alert. |
| Trigger | **Event-driven**: subscribe to `group.departed`. |
| Kit config location | Inside the new **Stok** tab ("Kit per Paket"). |

## 5. Data model

`inventory_items` is reused as-is. Two new tables; all rows are org-scoped.

### 5.1 `package_kit_items` (new)
A package's departure kit = its rows.

```
id              UUID PK
org_id          UUID NOT NULL
package_id      UUID NOT NULL
item_id         UUID NOT NULL  -- FK-by-convention -> inventory_items.id
qty_per_jamaah  INTEGER NOT NULL CHECK (qty_per_jamaah > 0)
created_at      TIMESTAMPTZ DEFAULT NOW()
updated_at      TIMESTAMPTZ DEFAULT NOW()
UNIQUE (org_id, package_id, item_id)
INDEX (org_id, package_id)
```

### 5.2 `stock_movements` (new)
Append-only ledger. Every stock change is one row.

```
id            UUID PK
org_id        UUID NOT NULL
item_id       UUID NOT NULL
delta         INTEGER NOT NULL          -- signed: +restock/+initial, -deduct/±adjust
reason        TEXT NOT NULL             -- 'initial' | 'restock' | 'adjustment' | 'departure'
note          TEXT DEFAULT ''
group_id      UUID                      -- set for reason='departure'
package_id    UUID                      -- set for reason='departure'
created_by    UUID                      -- NULL for system (event consumer)
created_at    TIMESTAMPTZ DEFAULT NOW()
INDEX (org_id, item_id, created_at DESC)
-- Idempotency: one departure deduction per (group, item).
UNIQUE INDEX uniq_departure_movement ON stock_movements(group_id, item_id)
    WHERE reason = 'departure'
```

`inventory_items.stock` is the source of truth for the current level; movements
are the history. They are kept consistent by always writing both inside one
transaction (§7).

## 6. Backend API (inventory-service)

New routes, added next to the existing inventory endpoints and proxied through
api-gateway the same way. All require auth; `org_id` comes from JWT claims.

**Items**
- `GET    /items` — list with current stock, min, unit, category, low/negative flags.
- `POST   /items` — create `{name, category, unit, min_stock, initial_stock}`; writes an `initial` movement when `initial_stock > 0`.
- `PATCH  /items/:id` — edit `{name, category, unit, min_stock}` (not stock directly).
- `POST   /items/:id/restock` — `{qty>0, note}` → `+qty`, movement `restock`.
- `POST   /items/:id/adjust` — `{delta, note}` (signed) → movement `adjustment`.
- `GET    /items/:id/movements` — ledger for the item (paged).
- `DELETE /items/:id` — blocked (409) if referenced by any `package_kit_items`.

**Kits**
- `GET /kits/:packageId` — kit items for a package.
- `PUT /kits/:packageId` — replace the package's kit with `[{item_id, qty_per_jamaah}]`.

**Event consumer (internal, not an endpoint)**
- inventory-service connects to NATS and durably subscribes to `group.departed`.

Reads/writes are scoped by `org_id`; cross-org access returns 404.

## 7. Auto-deduct flow

On `group.departed` (payload `{group_id, package_id, member_count}`):

1. If `package_id` empty or `member_count < 1` → ack, no-op.
2. Load `package_kit_items` by `package_id` (a UUID, globally unique to one org);
   the rows carry `org_id`. Empty → ack, log, no-op (no kit configured).
3. **One transaction:**
   - For each kit item:
     - `INSERT INTO stock_movements (… reason='departure', group_id, package_id, delta = -(qty_per_jamaah × member_count)) ON CONFLICT (uniq_departure_movement) DO NOTHING`.
     - **Only if the insert created a row** (RETURNING / rows-affected), apply `UPDATE inventory_items SET stock = stock - (qty×N)`. (Guarantees idempotency: a re-delivered event inserts nothing and updates nothing.)
   - Commit.
4. Ack the message.

- **Negative allowed:** the `UPDATE` is unconditional; stock may go below zero.
- **Idempotency:** the partial unique index makes re-delivery a no-op per item;
  the conditional update means the level can't drift on retries.
- Alerts (low/negative) are **computed** from `stock` vs `min_stock`, not stored.

The event payload carries **no `org_id`**; the consumer derives it from the kit
rows (each `package_id` belongs to exactly one org) and scopes every deduction
query to it. A package with no kit rows is simply skipped.

## 8. Frontend (Inventaris page)

The Inventaris page gains a tab switcher (existing app idiom, e.g. `FilterTabs`):

- **Tab "Distribusi"** — the current `member_equipment` workflow (forecast,
  mark-received, QR handover). **Unchanged.**
- **Tab "Stok"** (new):
  - **Summary**: total items, low-stock count, negative-stock count.
  - **Alert banner** when any item is low/negative.
  - **Item table**: name, category, unit, **current stock** (red when ≤0 or
    `< min_stock`), min stock, a "kit" marker if used in any kit. Row actions:
    **Restock**, **Adjust**, **History**.
  - **Add item** form: name, category, unit, initial stock, min stock.
  - **History** drawer per item: the movement ledger (date, reason, Δ, note,
    linked group for departures).
  - **"Kit per Paket"** section: pick a package → add items + qty per jamaah →
    save (`PUT /kits/:packageId`). Needs the package list (from package-service,
    already available to the frontend).

All new UI uses the app's Tailwind idiom (rounded-xl cards, `ring-slate-200/60`,
`focus:ring-primary-100`, serif slate headings, soft callouts) established in the
Pusat Bantuan polish.

## 9. Edge cases

- **Shortfall** → negative stock + red flag on the item + alert banner.
- **Re-delivered event** → no-op (idempotent, §7).
- **No kit / zero members** → no-op.
- **Item deleted while in a kit** → blocked (409); remove from kits first.
- **Departure undone** → no auto-reversal (v1); correct via manual adjustment.
- **Org isolation** → every query scoped by `org_id`.

## 10. Testing

- **Go unit/service tests:** deduct math (`qty×N`); idempotency (same
  `group.departed` twice → one deduction); negative/shortfall; no-kit and
  zero-member no-ops; delete-blocked-when-in-kit.
- **Event consumer test:** simulate a `group.departed` envelope end-to-end.
- **Frontend:** logic-level where testable (the repo's vitest is node-env);
  manual verification of the Stok tab, kit config, and alert states.
- Gates: `go test ./...`, `go vet ./...`, `npm run check`, `npm test`.

## 11. Rollout

- Single migration `migrations/inventory/003_stock_monitoring.{up,down}.sql`
  (new tables + indexes).
- inventory-service gains a NATS connection + durable consumer (new dependency
  for that service; bus already runs in compose).
- Frontend Inventaris page + new Stok components.
- Deploy: rebuild `inventory-service` + `frontend` (surgical, as before).
