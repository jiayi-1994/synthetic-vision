---
date: 2026-06-04
topic: unfinished-features-roadmap
title: Unfinished Features Roadmap Requirements
---

# Unfinished Features Roadmap Requirements

## Summary

Turn the two unfinished product surfaces into a staged roadmap: first make Analytics a real, data-backed dashboard using existing generation and credit activity, then make Marketplace a useful preset/template discovery surface that strengthens the generation workflow without becoming a full creator economy.

---

## Problem Frame

`/analytics` and `/marketplace` are currently visible navigation items but still render on-brand "coming soon" placeholders. That creates a product expectation gap: the core generation, gallery, credits, and admin flows exist, while two promised destination pages do not yet return user value.

The highest-leverage move is to complete these surfaces in an order that compounds the existing SaaS loop. Analytics should land first because the product already records generations, statuses, dimensions, costs, and user credit balances. Marketplace should follow as a lightweight preset/template market rather than a broad model marketplace, because the current product is centered on prompt-to-image generation and credit usage, not third-party creator commerce.

---

## Key Decisions

- **Phase Analytics before Marketplace.** Analytics can reuse existing product activity and immediately makes the user's generation history more understandable. Marketplace needs more product definition, so it should not block the lower-risk dashboard work.
- **Use business data before introducing analytics infrastructure.** The first Analytics version should answer practical questions from existing generation and credit records, not depend on a separate event pipeline.
- **Scope Phase 1 Analytics to personal user analytics.** Admin-wide or global platform analytics are deferred so the first phase can ship quickly and avoid mixing personal and operator views.
- **Start Marketplace as curated presets/templates.** The marketplace should help users start better generations faster through styles, prompt patterns, and reusable creative directions. It should not begin as model trading, creator payout, or upload-review infrastructure.
- **Keep the roadmap staged but connected.** Analytics and Marketplace should share the same product goal: make Synthetic Vision feel like a complete image-generation SaaS rather than a demo with two placeholder routes.

---

## Requirements

**Phase 1 — Analytics**

- R1. Replace the Analytics placeholder with a real dashboard that summarizes generation activity for the signed-in user.
- R2. Show core usage metrics that can be derived from existing product data: total generations, completed generations, failed generations, success rate, and recent activity.
- R3. Show credit-oriented metrics that help users understand spend: current balance, credits spent on generation, refunded credits, and approximate cost by resolution where available.
- R4. Show creative-production breakdowns such as resolution mix, aspect-ratio mix, and recent prompt/output history in a lightweight visual or tabular form.
- R5. Preserve the current glassmorphism visual language and navigation shell so Analytics feels like a completed part of the existing product, not a separate admin report.
- R6. If admin-wide analytics are included later, keep them clearly separate from the user's personal analytics so non-admin users never see global platform data.

**Phase 2 — Marketplace**

- R7. Replace the Marketplace placeholder with a curated browsing surface for generation presets, style packs, or prompt templates.
- R8. Each marketplace item should communicate what it helps create, enough preview/detail for the user to choose it, and a clear action to apply it to the generation workspace.
- R9. Applying a marketplace item should reduce prompt setup effort without hiding what will be sent to the generator; users should be able to edit the resulting prompt/style before generating.
- R10. Marketplace v1 should work with built-in or curated content and should not require user uploads, creator moderation, purchase flows, or settlement logic.
- R11. Marketplace should reinforce the existing credit-based generation model but should not introduce separate marketplace billing in the first version.

**Roadmap / Handoff**

- R12. The implementation plan should split Analytics and Marketplace into independently shippable phases, with Analytics as the first execution phase.
- R13. Any backend data additions should be limited to what is necessary to make the visible Analytics or Marketplace behavior true; avoid speculative tracking or commerce abstractions.
- R14. Existing completed flows — login/register, dashboard generation, gallery, admin credits, and image serving — must remain stable while these pages are completed.

---

## Actors

- A1. **Signed-in image creator** — uses Dashboard to generate images, Gallery to review outputs, Analytics to understand usage, and Marketplace to start from stronger presets.
- A2. **Admin/operator** — manages users and credits today; may later need global analytics, but that is not the default Phase 1 scope.
- A3. **Future planner/implementer** — uses this document to split work without inventing product behavior for the two placeholder pages.

---

## Key Flows

- F1. Analytics review
  - **Trigger:** A signed-in user opens `/analytics` from the sidebar.
  - **Actors:** A1
  - **Steps:** The page loads usage, status, credit, and creative breakdowns from product data; the user scans trends and recent activity.
  - **Outcome:** The user understands how much they have generated, what succeeded or failed, and how credits are being consumed.
  - **Covered by:** R1, R2, R3, R4, R5

- F2. Marketplace preset application
  - **Trigger:** A signed-in user opens `/marketplace` and chooses a preset/template.
  - **Actors:** A1
  - **Steps:** The user browses curated items, opens or inspects one, applies it to the generation workspace, then edits prompt/style before generating.
  - **Outcome:** The marketplace item accelerates creation while leaving the final generation parameters under user control.
  - **Covered by:** R7, R8, R9, R10, R11

- F3. Roadmap execution
  - **Trigger:** Planning begins from this requirements document.
  - **Actors:** A3
  - **Steps:** The planner creates separate implementation phases, verifies Phase 1 can ship independently, then scopes Phase 2 without adding unsupported commerce features.
  - **Outcome:** The unfinished pages are completed in a low-risk order with explicit non-goals.
  - **Covered by:** R12, R13, R14

---

## Acceptance Examples

- AE1. Analytics no longer looks unfinished
  - **Given:** A signed-in user has existing generation records.
  - **When:** They open `/analytics`.
  - **Then:** They see real usage and credit information rather than a "coming soon" placeholder.
  - **Covers:** R1, R2, R3

- AE2. Analytics handles empty history
  - **Given:** A newly registered user has no generations yet.
  - **When:** They open `/analytics`.
  - **Then:** They see a polished empty state that explains what will appear after generating, not broken charts or zero-only noise.
  - **Covers:** R1, R5

- AE3. Marketplace applies without locking the user in
  - **Given:** A user selects a curated marketplace preset.
  - **When:** They apply it.
  - **Then:** The generation workspace is prefilled with useful prompt/style direction and the user can edit before spending credits.
  - **Covers:** R8, R9, R11

- AE4. Marketplace avoids premature commerce
  - **Given:** Marketplace v1 is being implemented.
  - **When:** scope questions arise around creator uploads, paid packs, moderation, or payouts.
  - **Then:** those items are deferred rather than folded into the first implementation phase.
  - **Covers:** R10, R11, R12

---

## Success Criteria

- Analytics and Marketplace both stop presenting as unfinished placeholders.
- Analytics can be planned and shipped before Marketplace without waiting on a new event pipeline.
- Marketplace v1 makes generation easier through presets/templates while keeping generation review and credit spend in the existing Dashboard flow.
- A downstream `ce-plan` can split the work into clear Phase 1 and Phase 2 units without deciding product scope from scratch.

---

## Scope Boundaries

**Deferred for later**

- Admin-wide or platform-wide analytics beyond the minimum needed for the first dashboard.
- Advanced BI, custom date ranges, export/reporting, cohort analysis, or a separate telemetry/event system.
- User-submitted marketplace content, creator profiles, review queues, paid packs, payouts, or revenue sharing.
- Separate marketplace billing that is distinct from generation credit spend.

**Outside the first-pass product identity**

- A full third-party model store.
- A creator-economy marketplace with settlement, moderation, and ranking systems.
- A broad rewrite of generation, credits, gallery, or admin flows solely to support these pages.

---

## Dependencies / Assumptions

- Existing generation records are sufficient for the first Analytics version: status, cost, resolution, aspect ratio, timestamps, and image/prompt metadata.
- Existing credit transaction records are sufficient for basic spend/refund views, though planning should verify whether an authenticated user-facing credit ledger endpoint is needed.
- Marketplace v1 can start with curated/static content if no admin-managed preset model exists yet.
- The current README/SPEC intentionally list Marketplace and Analytics as placeholder routes, so completing them is aligned with the documented product shape.

---

## Outstanding Questions

**Resolved**

- Phase 1 Analytics is strictly personal-user analytics. Admin/global analytics are deferred for a later phase.

**Deferred to Planning**

- Which exact data contract should feed Analytics: extend existing stats endpoints, add analytics-specific endpoints, or derive some data client-side from generation lists?
- Should Marketplace presets be static bundled content for v1, database-backed curated content, or admin-editable content?
- What is the minimal navigation handoff from Marketplace to Dashboard: route query params, shared store state, or another product pattern?

---

## Sources / Research

- `README.md` documents `/marketplace` and `/analytics` as on-brand "Coming soon" placeholders.
- `SPEC.md` defines the existing product contract and acceptance checklist for generation, gallery, admin, credits, and the two placeholder routes.
- `frontend/src/views/Marketplace.vue` and `frontend/src/views/Analytics.vue` currently render placeholder pages.
- `backend/internal/models/models.go`, `backend/internal/handlers/profile.go`, and `backend/internal/handlers/admin.go` show existing generation, credit, user stats, and admin data surfaces that planning can verify for reuse.



