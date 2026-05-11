# Roadmap

A rough, non-binding view of where this provider is going. Items are grouped by the **target release** they are most likely to ship in, but specific scope can shift based on feedback and AceCloud backend availability.

If something here matters to you, please open a GitHub issue and link it to the relevant section — that's the most reliable way to influence sequencing.

> **Status legend:** ✅ shipped · 🚧 in progress · 📋 planned · 💡 considering

---

## Versioning posture

This provider follows [Semantic Versioning](https://semver.org/). While we're on **v0.x.y**:

- Minor bumps (0.1 → 0.2) may contain breaking changes; check `CHANGELOG.md` and the upcoming `UPGRADING.md` before bumping
- Patch bumps (0.1.0 → 0.1.1) are strictly backwards-compatible bug fixes
- Resources and attributes can be deprecated and removed in a single minor cycle

Once **v1.0.0** ships, we adopt the standard SemVer contract: breaking changes only at major bumps, with a documented deprecation cycle of at least one minor before removal.

---

## v0.1.x — patches (rolling)

✅ **v0.1.0** — Initial preview release (2026-04-30). Documented in `CHANGELOG.md`.

📋 **v0.1.1** — bug-fix-only patch covering work since v0.1.0:

- Error envelope handling for the NestJS default error shape (`{error: "Bad Request"}` as string)
- Error-summary phrasing standardised to `"Failed to <verb> <resource>"` everywhere
- Internal-tech leakage scrub in user-facing docs

This will ship as 0.1.1 from `main` once the changes above are tagged.

---

## v0.2.0 — Lifecycle parity + LB unblock

📋 **Target theme:** make every IaaS resource behave the way Terraform users expect — in-place updates where the API supports them, clean lifecycle on destroy, no spurious replacement plans.

Already landed on `main`, will ship as part of v0.2.0:

- ✅ `acecloud_volume_attachment` (new resource)
- ✅ `acecloud_instance.power_state` and `acecloud_instance.locked` attributes
- ✅ `acecloud_instance.flavor_id` in-place resize
- ✅ `acecloud_instance.security_group_ids` in-place update
- ✅ `acecloud_auto_scaling_deployment` in-place update for 10 mutable fields
- ✅ `acecloud_volume` resize fix (use `POST /extend`)
- ✅ Keypair import field-name fix and Read drift fix
- ✅ Instance Delete: unlock + power-on before delete

Still needed before v0.2.0 ships:

- 🚧 Live E2E verification of `acecloud_auto_scaling_deployment` Update (gated by `ACECLOUD_RUN_AS_DEPLOYMENT_ACC=1`, ~30 min per cycle)
- 🚧 Live verification of `acecloud_instance.flavor_id` resize on a multi-host AceCloud cluster (preprod is single-host so the backend silently no-ops the resize)
- 🚧 Backend fix for the LB pool create worker on preprod (escalated to platform team — see `docs/USAGE.md` Troubleshooting). The provider itself is correct; pool create succeeds on dev4 with the same payload.
- 📋 `UPGRADING.md` covering migration from 0.1.x to 0.2.x

Possible additional scope if backend support lands in time:

- 💡 `acecloud_volume.bootable` toggle (currently inferred from `boot` flag in `volumes` block)
- 💡 `acecloud_instance` import support
- 💡 More data sources: snapshots, volumes, load balancers, listeners, pools

---

## v0.3.0 — CaaS + Container Registry + Kubernetes (beta)

📋 **Target theme:** restore the deferred container-platform resources from the `feature/beta-resources` branch. Mark them clearly as Beta in docs until they've had a customer-facing soak period.

Resources to restore (originally built in earlier sessions, removed from the current release branch on 2026-04-23 pending backend stabilisation):

- `acecloud_caas_secret`
- `acecloud_caas_deployment`
- `acecloud_registry_project`
- `acecloud_registry_replication_rule`
- `acecloud_k8s_cluster`
- `acecloud_k8s_node_group`

Prerequisites (mostly backend work, tracked separately):

- 🚧 CaaS deployment endpoint stabilised (was returning `OutOfSync` on dev4)
- 🚧 Container registry endpoint functional on customer-facing environment
- 🚧 K8s cluster long-provisioning behavior (25-35 min creates) handled gracefully — already in place via `PollForResource` with 40-min timeout

Each beta resource will ship with:

- A Beta banner in its registry-facing documentation page
- An explicit note in `CHANGELOG.md` that beta resources may change in non-major bumps
- A separate `tests/live/` acceptance test gated behind an opt-in env var

---

## v0.4.0 — Drift detection + import support

📋 **Target theme:** tighten the loop between Terraform state and live backend state.

- 📋 `terraform import` support for every resource that has a stable identity
- 📋 Improved drift detection for `acecloud_security_group.rules` (currently the provider preserves user-configured rules without reconciling against backend transformations like `tcp/22 → SSH`)
- 📋 Read-only data sources for: snapshots, volumes, load balancers, listeners, pools, health monitors, auto-scaling templates, auto-scaling deployments, keypairs, floating IPs
- 💡 Optional verbose-diff mode for `terraform plan` that shows backend-computed defaults

---

## v0.5.0 — Namespace migration to official AceCloud org

📋 **Target theme:** publish under `acecloud/acecloud` on the Terraform Registry instead of the personal namespace.

- 📋 Claim the `acecloud` publisher account on registry.terraform.io
- 📋 Set up the official-org GitHub repo with appropriate access controls
- 📋 Publish identical `v0.5.0` artifacts under both namespaces for one minor cycle
- 📋 Update README, docs, and examples to reference the new source
- 📋 Add a clear migration note to `UPGRADING.md` — users will need to change their `required_providers` `source` line

Old namespace will continue to receive patches for one minor cycle (`v0.5.x`) after the move, then be archived with a pointer to the new location.

---

## v0.6.x — v0.9.x — Stabilisation

📋 **Target theme:** stop adding net-new resources. Polish what's there. Run a community feedback window for any remaining breaking changes that should land before v1.0.

Expected work:

- Remove anything that was deprecated during v0.2-v0.4
- Tighten validation rules to catch invalid configs at plan-time where possible
- Improve error messages with actionable next steps
- Add `validate` rules that catch cross-resource consistency issues at plan time
- RFC window: file 1-2 GitHub Discussions for any remaining breaking changes; close the window 30 days before the v1.0 cut

---

## v1.0.0 — GA, stability contract

📋 **Target theme:** first release with a public SemVer stability guarantee.

Requirements before tagging v1.0.0:

- ✅ Two consecutive minor releases with no breaking changes to existing resources
- 📋 100% of resources have at least one live acceptance test that passes on the production backend
- 📋 Every resource has an `examples/` folder with a working configuration
- 📋 `UPGRADING.md` documents the full path from each prior 0.x minor to 1.0
- 📋 Security policy updated to cover the previous major in parallel
- 📋 Customer announcement and recommended-upgrade window

After v1.0:

- Breaking changes only at major bumps
- Deprecation requires at least one full minor cycle with a plan-time warning before removal
- Each deprecation documented in `UPGRADING.md` with a recommended migration

---

## Not currently planned

To set expectations, here's what we're **not** planning to ship in the visible roadmap:

- A Pulumi or CDKTF distribution. The HashiCorp tooling generates these automatically from the provider; we don't maintain separate codebases.
- A "raw API" passthrough resource. If something is missing, file an issue — we'd rather build a typed resource than an escape hatch.
- A separate `acecloud_spot_instance` resource. Spot pricing is exposed by the AceCloud platform's instance launch flow; we'd need a clear product-side definition of where spot intersects with `acecloud_instance` before adding it.
- Provider-side cost estimation. Terraform Cloud and Spacelift both surface cost via their own integrations against the registry; we'll lean on those rather than building our own.

If any of these would unblock you, please open an issue — roadmaps are revisable.

---

## How to influence the roadmap

- **Open a GitHub issue** describing the use case and link to any backend documentation
- **React with 👍** to existing issues to signal interest — we use reaction counts for sequencing
- **Submit a PR** for small additions (a new data source, an attribute on an existing resource); larger changes should start with an issue for design alignment
- For AceCloud customers: reach out via your usual support channel and reference the roadmap item

---

*Last updated: 2026-05-11.*
