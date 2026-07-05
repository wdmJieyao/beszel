# Feature Specification: Public Probe Visibility and Refresh Commands

**Feature Branch**: `[007-public-probe-visibility]`

**Created**: 2026-07-04

**Status**: Draft

**Input**: User description: "1、将线路检测里面的“公开展示”选项，迁移到公共看板对应的VPS实例上面去，可以在公共看板设置页面设置，需要考虑迁移对现有数据结构是否有冲突，因为我已经在生产环境部署了。2、产生docker run 命令的地方 可以实现先清理旧镜像，然后重新拉新的镜像吗？"

## Clarifications

### Session 2026-07-04

- Q: How should existing probe-level public visibility migrate to per-VPS visibility? → A: Preserve only the visibility that existed before upgrade for each public VPS; do not expand public probes to additional VPS instances.
- Q: Which generated Docker run commands should refresh images before starting? → A: All generated Docker run commands, including panel, agent, and agent variants.
- Q: Should generated Docker run commands also handle existing containers? → A: Remove existing containers and old images before pulling the new image and starting the new container.
- Q: What should be the default probe visibility for a newly public VPS after migration? → A: Default to no public probe lines selected, with a one-click select-all action available in settings.

### Session 2026-07-05

- Q: Where should administrators manage public probe visibility for a VPS? → A: Only in the public dashboard settings page; do not duplicate the control in the normal VPS edit page.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Public Probe Visibility Per VPS (Priority: P1)

As an administrator, I want to decide which latency probe lines are visible for each VPS on the public dashboard from the public dashboard settings page, so that public exposure is controlled at the same place where each public VPS is managed.

**Why this priority**: This is the primary behavior correction. The current "公开展示" switch on the probe itself is too broad because a line may be safe to show for one public VPS but not for another.

**Independent Test**: Can be tested by enabling a VPS for public display, selecting visible latency probe lines for that VPS, opening the anonymous public dashboard, and confirming only the selected lines appear for that VPS.

**Acceptance Scenarios**:

1. **Given** an administrator is editing a VPS in public dashboard settings, **When** they select one or more latency probe lines for that VPS, **Then** the public dashboard shows those selected lines only for that VPS.
2. **Given** two public VPS instances have different probe visibility selections, **When** an anonymous visitor views the public dashboard, **Then** each VPS shows only its own selected latency probe lines.
3. **Given** a probe line exists but is not selected for a public VPS, **When** the public dashboard renders that VPS, **Then** the line name, target, data, and chart series for that probe are not exposed for that VPS.
4. **Given** a VPS is newly enabled for public display after migration, **When** the administrator opens its public probe visibility settings, **Then** no probe lines are selected by default and the administrator can choose to select all in one action.
5. **Given** an administrator edits a VPS in its normal non-public edit page, **When** they look for public probe visibility controls, **Then** those controls are not duplicated there and remain managed from the public dashboard settings page only.

---

### User Story 2 - Preserve Existing Production Visibility During Upgrade (Priority: P1)

As an administrator with an existing production deployment, I want the upgrade to preserve my current public dashboard behavior, so that already visible monitoring data does not disappear or unexpectedly become more public after the new per-VPS controls are introduced.

**Why this priority**: The user has already deployed to production. A visibility migration is sensitive because it changes public exposure rules and could otherwise cause data loss, hidden charts, or unintended disclosure.

**Independent Test**: Can be tested by preparing existing public VPS and probe visibility settings, upgrading, and confirming the resulting per-VPS visibility selections match the previously visible public dashboard behavior.

**Acceptance Scenarios**:

1. **Given** an existing probe line is currently marked for public display and a VPS is currently public, **When** the system is upgraded, **Then** that probe line remains visible for that public VPS through the new per-VPS setting.
2. **Given** an existing probe line is not currently marked for public display, **When** the system is upgraded, **Then** that probe line is not automatically exposed on public VPS instances.
3. **Given** a VPS is not enabled for the public dashboard, **When** the system is upgraded, **Then** no probe visibility setting causes that VPS to appear publicly.
4. **Given** an upgrade has already run once, **When** the application starts again or the migration is rechecked, **Then** visibility data is not duplicated or reset.
5. **Given** a probe line was visible for some public VPS instances but not others before upgrade, **When** the system is upgraded, **Then** the new per-VPS selections preserve that exact visible set and do not expand the probe to additional public VPS instances.

---

### User Story 3 - Generate Refresh-Safe Docker Run Commands (Priority: P2)

As an administrator installing or updating agents from generated commands, I want the generated Docker run command to remove the old local image and pull the latest image before starting the container, so that copied commands reliably deploy the current published image instead of reusing a stale cached image.

**Why this priority**: This improves deployment reliability after GHCR image updates. It is less sensitive than public visibility but directly affects whether users actually run the latest custom images.

**Independent Test**: Can be tested by copying a generated Docker command into a host with an older local image and confirming the command refreshes the image before starting the container.

**Acceptance Scenarios**:

1. **Given** a generated Docker command is displayed, **When** an administrator reads it, **Then** the command includes steps to remove the old local image for the relevant image name before pulling the latest image.
2. **Given** no previous local image exists, **When** the generated command runs, **Then** the cleanup step does not stop the rest of the command from pulling and starting the container.
3. **Given** an administrator reruns the generated command after a new image is published, **When** the command completes successfully, **Then** the started container uses the newly pulled image.
4. **Given** generated Docker run commands are shown for panel, agent, or agent variants, **When** an administrator copies any of those commands, **Then** the same image cleanup and pull behavior is included.
5. **Given** a previous container with the generated command's container name already exists, **When** the generated command runs, **Then** it removes the old container before creating the replacement container.

### Edge Cases

- A VPS is public, but no latency probe lines are selected for it: the public dashboard should show the VPS without latency probe chart data and without leaking unselected probe names or targets.
- A latency probe is disabled or deleted after being selected for a VPS: the public dashboard should ignore unavailable selections and the settings page should make stale selections recoverable.
- A newly added VPS is made public after the migration: it should not inherit public probe exposure accidentally, should start with no selected probe lines, and should support an explicit select-all action.
- Existing production data contains public probes but no public VPS instances: the migration should preserve probe records without creating public VPS exposure.
- Existing production data contains public VPS instances but no public probes: the migration should not create visible probe lines.
- Generated Docker commands should tolerate cleanup failures caused by a missing local image or missing old container while still failing clearly on real pull or container start errors.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST remove the public visibility control from the latency probe configuration experience as the primary way to decide anonymous public dashboard exposure.
- **FR-002**: The system MUST provide public latency probe visibility controls on each VPS entry in the public dashboard settings experience.
- **FR-003**: Administrators MUST be able to choose zero, one, or multiple latency probe lines to display for each public VPS.
- **FR-004**: The public dashboard settings page MUST be the only user-facing place to manage per-VPS public probe visibility, and the normal VPS edit page MUST NOT duplicate that control.
- **FR-005**: Newly public VPS instances created or enabled after migration MUST default to no public probe lines selected.
- **FR-006**: Anonymous public dashboard viewers MUST only see latency probe chart data for probe lines selected on the corresponding public VPS.
- **FR-007**: Anonymous public dashboard viewers MUST NOT receive names, targets, status, results, or metadata for unselected latency probe lines on a VPS.
- **FR-008**: The public dashboard settings experience MUST provide a one-click action to select all available probe lines for a VPS.
- **FR-009**: Existing production deployments MUST migrate current public probe visibility into the new per-VPS visibility model without requiring manual reconfiguration for previously visible dashboard data.
- **FR-010**: Migration MUST NOT make a non-public VPS public.
- **FR-011**: Migration MUST NOT make a previously non-public latency probe visible on any VPS.
- **FR-012**: Migration MUST be repeatable without duplicating, resetting, or widening per-VPS probe visibility selections.
- **FR-013**: Migration MUST preserve only the anonymous public dashboard visibility that existed before upgrade for each public VPS and MUST NOT expand a public probe line to VPS instances where it was not previously visible.
- **FR-014**: If legacy public probe fields remain for compatibility, the user-facing settings experience MUST make the per-VPS public dashboard setting the source of truth for public dashboard exposure.
- **FR-015**: Generated Docker run commands MUST include a best-effort cleanup step for the old local image associated with the command.
- **FR-016**: Generated Docker run commands MUST pull the configured image before creating or starting the container.
- **FR-017**: Generated Docker run commands MUST continue when the cleanup step has nothing to remove, but MUST fail clearly if pulling the new image or starting the container fails.
- **FR-018**: The refresh behavior MUST apply to all generated Docker run commands shown by the product, including panel, agent, and agent variant commands.
- **FR-019**: Generated Docker run commands MUST remove an existing container with the generated command's container name before creating the replacement container.
- **FR-020**: Changed behavior MUST be covered by focused unit tests.
- **FR-021**: If HTTP APIs are added or changed, they MUST be RESTful or explicitly justify an existing non-REST contract.

### API Contracts *(include if HTTP API changes)*

- **Resources**: Public dashboard VPS settings and latency probe visibility selections are the affected resources.
- **Methods**: Reading settings must return each public VPS with its selected public probe lines. Updating settings must support replacing the selected probe line set for a VPS without altering unrelated VPS settings.
- **Status Codes**: Successful reads and updates must use normal success responses. Invalid VPS IDs, invalid probe IDs, unauthorized access, and stale deleted selections must return predictable error responses or safe ignored states as appropriate.
- **Schemas**: Public VPS settings must include a collection of selected latency probe identifiers. Anonymous public dashboard responses must include only selected public probe series for each VPS.
- **Compatibility**: Existing deployments using probe-level public visibility must be migrated so currently visible public dashboard probe data remains visible for the same eligible VPS instances after upgrade.

### Key Entities *(include if feature involves data)*

- **Public VPS Visibility Setting**: Represents whether a VPS appears on the anonymous public dashboard and which latency probe lines are allowed to appear for that VPS.
- **Latency Probe Line**: Represents a configured network probe line that can collect latency results and may be selected for display on one or more public VPS instances.
- **Legacy Probe Public Visibility**: Represents the existing probe-level visibility flag that must be interpreted during migration but should no longer be the primary public exposure control.
- **Generated Docker Command**: Represents a copyable deployment command shown to administrators for starting Beszel components or agents with the configured image source.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After upgrade, 100% of previously visible public dashboard latency probe charts remain visible for the same eligible public VPS instances.
- **SC-002**: After upgrade, 0 non-public VPS instances become public due to the visibility migration.
- **SC-003**: After upgrade, 0 previously non-public latency probe lines are exposed on the anonymous public dashboard.
- **SC-004**: Administrators can configure public probe visibility for a VPS from the public dashboard settings page in under 2 minutes.
- **SC-005**: Administrators can change a newly public VPS from no visible probe lines to all visible probe lines in one settings action.
- **SC-006**: Anonymous public dashboard responses expose no unselected probe names, targets, or result series for a VPS in all tested visibility combinations.
- **SC-007**: Generated Docker commands refresh the relevant image before container start in 100% of command examples shown by the product.
- **SC-008**: Re-running the migration or restarting after migration preserves the same per-VPS probe visibility selections without duplicates.
- **SC-009**: 100% of generated Docker run command variants shown for panel and agent deployment include consistent image refresh behavior.
- **SC-010**: Re-running any generated Docker run command succeeds on hosts that already have the previous generated container name and image, assuming the new image can be pulled and required ports are available.

## Assumptions

- Existing probe-level "公开展示" values are treated as the migration source for deciding which probe lines should initially be visible on the same public VPS instances where they were visible before upgrade.
- Existing public VPS visibility remains the gate for whether a VPS appears on the public dashboard at all.
- The migration should favor preserving existing public dashboard output, not creating stricter defaults that hide charts immediately after upgrade.
- Newly configured public VPS instances after migration start with no selected public probe lines and may be expanded quickly with a deliberate select-all control in the settings experience.
- Docker command cleanup is best effort for missing local images or missing old containers but should not mask failures that prevent pulling or starting the requested image.
