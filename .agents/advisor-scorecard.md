# Advisor Scorecard (GPT / GLM / DeepSeek / OpenRouter)

Tracks how well each advisor actually performs per task domain, so Claude Code (the executor, see
`.agents/rules/bc-account-core-rules.md` → "Multi-Model Role Division") can route future work to
whoever is actually strong at it instead of a fixed assignment forever. Append-only running log,
same convention as `worklog.md` — read the summary table before picking an advisor for a new kind
of task, append one row every time an advisor is consulted and its answer is checked against real
evidence.

**Scoring rubric (1-10 per consult):**
- Accuracy (0-4): did Claude's own verification against real source/runtime confirm the advice, or
  did it turn out wrong/hallucinated?
- Usefulness (0-4): how much of the answer was actually adoptable/adopted as-is vs discarded?
- Reliability (0-2): did the call succeed cleanly, or did it need retries/time out/error?

10 = fully correct, fully adopted, first try. 0 = wrong and unusable. Show the score to Jead every
time an advisor is consulted, inline in the task report (for example "GLM: 8/10 — ...").

## Summary (running average per advisor x domain, update after each new row below)

| Advisor | Domain | Avg score | # consults | Notes |
|---|---|---|---|---|
| GLM | Frontend/React/Next.js/full-stack | 6.0/10 | 5 | Wide swing: one timeout (2/10), strong runs (8, 5, 9, 6/10) — reliability is the real open risk, not answer quality when it actually responds; strong on scope-boundary/defer-list calls, but assumes codebase shape without seeing it (proposed a form-context layer the real per-tab-component code doesn't need, assumed a tenant business-type flag that doesn't exist on Shop). |
| DeepSeek | Backend/API/MongoDB | 5.6/10 | 5 | Consistently right on high-level direction/consensus, consistently weaker on inventing exact field shapes or over-scoped schema (new collections, arrays, index fields like `isactive`/`searchkeyword` that don't exist on the real Product model) not grounded in real source — verify its concrete schema proposals harder than its high-level ones. |
| GPT | Requirement/UX/Thai copy | 8.0/10 | 5 | Most consistent advisor so far across general UX calls and a real Thai tax-law citation check (9/10) — candidate to expand into more Thai-legal/domain research going forward per the adaptive-routing rule. |
| GLM | (other domains) | — | 0 | Expand here if GLM proves strong outside its default domain. |
| DeepSeek | (other domains) | — | 0 | Expand here if DeepSeek proves strong outside its default domain. |
| GPT | (other domains) | — | 0 | Expand here if GPT proves strong outside its default domain. |
| OpenRouter | Wildcard/model-picker (any domain) | — | 0 | Added 2026-07-05, working via `~/.claude/tools/openrouter-ask.py` (verified live: `--model openai/gpt-5.2` answered correctly, `--list-models --filter deepseek` returned real pricing). Log the underlying model name in the domain/task column below (e.g. "OpenRouter → anthropic/claude-sonnet-5 — ..."), not just "OpenRouter", since the average here spans many different underlying models. |

## Log (append one row per consult, most recent last)

| Date | Advisor | Domain/task | Score | Reason |
|---|---|---|---|---|
| 2026-07-04 | GPT | Requirement/UX — creditor/debtor screen design | 8/10 | SME-friendly naming, 7-section layout, and the Auth-field UX call were sound and substantially adopted; call succeeded cleanly. |
| 2026-07-04 | DeepSeek | Backend — creditor/debtor schema | 7/10 | Correctly flagged the Auth plain-text-password risk and proposed WHT/bank-account/credit-limit fields that were adopted; also proposed PaymentTermType/DiscountDays fields that were judged out of scope and not adopted. |
| 2026-07-04 | GLM | Frontend/fullstack — creditor/debtor screen design | 2/10 | Call timed out after 3 retries (120s/180s/300s); honestly reported as a failure rather than fabricated, but delivered zero usable content this round. |
| 2026-07-05 | GPT | Requirement/UX — cost-center/project/job/channel screens | 7/10 | Reasonable SME UX and scope-trim suggestions, partly folded into the do-not-build list; nothing repo-specific to verify against. |
| 2026-07-05 | DeepSeek | Backend — cost-center/project/job schema | 5/10 | Correctly backed the 3-separate-modules direction, but proposed a `parentcode`+Job-FK shape and a non-repo `Names{Th,En}` struct that Fable rejected as not matching real source (YAGNI/ungrounded specifics). |
| 2026-07-05 | GLM | Full-stack architecture — cost-center/project/job module split | 8/10 | Converged on the same no-parentcode/no-new-renderer/leave-channelprice-alone position Fable ultimately adopted almost verbatim; call succeeded cleanly this time. |
| 2026-07-05 | GPT | Thai tax law — counterparty branch number on tax invoice/WHT cert | 9/10 | Correctly distinguished Section 86/4 (general minimum fields) from Notification No. 199 ข้อ 7-9 (specific branch/HQ requirement, VAT-registered buyer only) exactly matching primary-source text pulled directly from rd.go.th; correctly said WHT cert has no branch-code field; recommended option (b) one-master-with-branch-list, matching the final synthesis. |
| 2026-07-05 | DeepSeek | Thai tax law — counterparty branch number on tax invoice/WHT cert | 6/10 | Right on WHT cert (no branch requirement) and right that option (c)/(b) dominate over (a); overstated invoice rule as unconditional ("Yes, mandatory") and missed the VAT-registered-buyer-only condition confirmed in Notification No. 199 by primary source. |
| 2026-07-05 | GLM | Thai tax law — counterparty branch number on tax invoice/WHT cert | 5/10 | Correctly said WHT cert has no branch requirement and correctly favored (b)/(c) over (a); wrongly asserted counterparty branch is merely "best practice, not legally required" on the tax invoice, contradicted by Notification No. 199 ข้อ 9 which is a real legal mandate (conditional on buyer being VAT-registered) — its own Section 86/4 citation was also imprecise (invoice minimum fields don't literally require branch; that's ฉบับที่ 199's addition). |
| 2026-07-05 | GPT | Requirement/UX — unified business-partner design (creditor/debtor as same real party) | 8/10 | Shared partner identity + role flags + badge/combined-view (no auto-netting) + repeatable bank-account rows + defer list matched Fable's final spec almost exactly. |
| 2026-07-05 | DeepSeek | Backend — unified business-partner schema | 6/10 | Correctly identified `taxid` as the natural shared key (adopted); also proposed a new `creditlimit` collection and a `Branches` array restructure that Fable explicitly rejected as premature machinery/unnecessary churn on live screens. |
| 2026-07-05 | GLM | Frontend architecture — Thai address genericization + unified-partner scope boundaries | 9/10 | Migrate-branch-first, reuse-image-gallery-pattern-for-bank-accounts, separate `addressforactual` field (not reusing shipping array), and the exact phase-1/phase-2 scope split were all adopted near-verbatim by Fable; only its `variant` config option and one branch-code legal claim were rejected. |
| 2026-07-06 | GPT | Requirement/UX — product/barcode/product-set screen redesign (12-tab overload, bundle-vs-BOM) | 8/10 | Daily-use-vs-advanced tab grouping, "Product Set = ขายเป็นชุด / BOM = ผลิต" clarifier, frontend-only restaurant/marketplace hiding, and the defer list were all adopted into the final spec; its Product Set feature wish-list (stock-availability preview, COGS rollup) correctly self-deferred but slightly over-scoped for a first pass. |
| 2026-07-06 | DeepSeek | Backend — Product/ProductBarcode schema strategy + index audit | 4/10 | High-level "strictly additive, never rename" stance is correct but already project law; claimed Product Set has "no structural validation" when backend ValidateProductClassification actually double-guards itemtype/materialtype (verified in UAT); index recommendations cite fields (`isactive`, `searchkeyword`, `nameen`) that don't exist on the real Product model; dual-write RestaurantDetail migration machinery is over-engineered for a pre-launch disposable DB. |
| 2026-07-06 | GLM | Frontend architecture — product-screen decomposition + tenant gating + test plan | 6/10 | Extraction-order plan, hide-don't-disable, and scope-creep defer list adopted in modified form; but its form-context refactor is unnecessary (real code already has per-tab components taking {value,onChange} props — mechanical file moves suffice), and its tenant-gating step assumes an existing business-type flag that verification shows does not exist on the Shop model. |
| 2026-07-07 | DeepSeek | Backend — Product Mongo→Kafka→PG pipeline design | 6/10 | Dedicated `when-product-*` topics, metadata-only writes (never touch stock columns), and the JOIN-ordering race analysis were all correct and adopted; but "extend inventory.go + reuse biapi-inventory-consumer group" contradicts the repo's real per-entity registration pattern in handlers/kafka.go, its plain-INSERT-no-ON-CONFLICT breaks idempotent resync, and it missed the DatabaseRebuildAll-recreates-empty-table trap entirely. |
| 2026-07-07 | GLM | Backend — Product Kafka pipeline file layout + registration | 9/10 | Nearly everything verified true against source and adopted: emit via a small producer-side helper (real pattern = MessageQueueRepository), new sibling consumer file kafka/product.go, new `biapi-product-consumer` group, registration at the exact same site as biapi-inventory-consumer (handlers/kafka.go StartConsumers — verified), upsert-on-itemcode leaving stock columns alone, and flagging rebuild-backfill as mandatory; only deviation is Fable includes the minimal resync endpoint in this pass instead of deferring it. |
| 2026-07-07 | GPT | Requirement — backfill existing Product data or prospective-only | 8/10 | Correct business call that prospective-only looks like "สินค้าหาย" to the owner and that repair must be a reusable tenant-scoped action (not a one-off script) because PG is disposable/rebuildable — adopted as POST /product/resync reusing the new pipeline; honestly reported its sandbox read failure instead of fabricating source claims; per-count reporting suggestion adopted in reduced form. |
