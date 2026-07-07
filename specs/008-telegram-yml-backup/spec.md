# Feature Specification: Telegram Notifications and YML Configuration Backup

**Feature Branch**: `[008-telegram-yml-backup]`

**Created**: 2026-07-06

**Status**: Draft

**Input**: User description: "1、通知考虑接入TG BOT，设计一下接入模式以及功能菜单。2、YML配置备份看一下如何适配我们现有的所有功能，导出配置的时候最好要支持我们现有的功能。3、上一次你做的变更对beszel-agent没有做什么变更吧，我这次只是更新了主面板而已。"

## Clarifications

### Session 2026-07-06

- Q: Telegram Bot 应该采用哪种授权/绑定模式？ → A: 面板手动配置 Telegram chat ID 白名单，并区分管理员高权限绑定与只读通知渠道。
- Q: YML 备份中的敏感值应该如何处理？ → A: 敏感值必须加密导出，未加密导出不得包含明文敏感值，恢复时必须提供解密凭据。
- Q: YML 导入恢复默认采用哪种应用策略？ → A: 合并恢复：创建缺失项、更新匹配项，默认不删除目标实例额外配置。
- Q: YML 导入恢复应该如何匹配现有配置？ → A: 优先使用备份中的稳定标识匹配，名称仅用于显示和冲突提示。
- Q: 只读 Telegram 通知渠道应该接收哪些内容？ → A: 可配置节点和告警级别范围，但只能接收非敏感告警摘要。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Telegram Bot Notification Channel (Priority: P1)

An administrator wants to add Telegram as a notification channel so important monitoring alerts can be delivered to an authorized private chat, group, or channel without relying only on the web panel.

**Why this priority**: Notification delivery is the core operational value. The integration must work reliably before interactive menus or backup workflows are useful.

**Independent Test**: Can be tested by configuring Telegram delivery, sending a test notification, and triggering a representative alert while confirming only authorized recipients receive messages.

**Acceptance Scenarios**:

1. **Given** an administrator has valid Telegram connection details and an authorized destination, **When** they save the notification channel and send a test message, **Then** the destination receives a clearly labeled test notification.
2. **Given** a monitored node changes into an alerting state, **When** Telegram notifications are enabled for that alert path, **Then** the authorized Telegram destination receives an alert containing the affected node, severity, timestamp, and summary.
3. **Given** Telegram delivery fails because the destination is invalid, the bot is blocked, or permissions are missing, **When** the system attempts delivery, **Then** the administrator can see a clear failure state and troubleshooting message without exposing sensitive credentials.

---

### User Story 2 - Telegram Bot Menu and Authorized Queries (Priority: P2)

An administrator wants a Telegram bot menu that can quickly show monitoring information and notification controls from Telegram while avoiding risky operations that should stay in the web panel.

**Why this priority**: A menu improves day-to-day operations, but it must be permissioned and scoped so Telegram is a convenient status surface rather than an unsafe remote admin console.

**Independent Test**: Can be tested by binding an authorized Telegram chat, opening the bot menu, selecting each menu item, and confirming the response matches the same current state visible in the panel.

**Acceptance Scenarios**:

1. **Given** an authorized Telegram user opens the bot menu, **When** they choose "状态总览", **Then** they receive a compact overview of node counts, alert counts, and recently problematic nodes.
2. **Given** an authorized administrator Telegram binding chooses "节点列表", **When** the bot responds, **Then** the response lists configured nodes with concise health indicators and allows the administrator to request a single node summary.
3. **Given** an authorized Telegram user chooses "通知设置", **When** they view the menu, **Then** they can see notification enablement and choose safe actions such as temporary mute or restore notifications.
4. **Given** an unauthorized Telegram user sends a command or opens the bot, **When** the system receives the request, **Then** no monitoring data is disclosed and the request is rejected or ignored with a safe response.
5. **Given** a Telegram destination is configured as a read-only notification channel, **When** matching monitoring events are delivered, **Then** it receives only non-sensitive alert summaries for the configured node and alert-level scope and cannot use privileged bot menu actions.

---

### User Story 3 - YML Configuration Export and Restore (Priority: P3)

An administrator wants to export the panel configuration as a YML backup that covers existing configurable features, including recently added public dashboard and network probe settings, so a deployment can be audited, migrated, or restored.

**Why this priority**: Configuration backup reduces operational risk and supports production recovery. It should reflect current product capabilities, not only legacy settings.

**Independent Test**: Can be tested by creating a representative configuration, exporting it, validating the exported file contents, restoring it into a clean or test instance, and confirming the resulting configuration matches the original within documented secret-handling rules.

**Acceptance Scenarios**:

1. **Given** an administrator has configured systems, alerting, public dashboard visibility, network probes, and notification channels, **When** they export a YML backup, **Then** the file contains versioned sections for each supported configurable area.
2. **Given** an administrator imports a valid YML backup, **When** they preview the restore, **Then** they can see what will be created, updated, skipped, or rejected before applying changes, with no deletions selected by default.
3. **Given** the YML backup contains unsupported, unknown, or newer-version sections, **When** the administrator imports it, **Then** the system reports compatibility warnings and avoids silently corrupting existing configuration.
4. **Given** sensitive values are present in configurable features, **When** the administrator exports or imports the backup, **Then** sensitive values are included only through encrypted export and import requires the matching decryption credential.

---

### User Story 4 - Preserve Agent Compatibility (Priority: P4)

An administrator who only updates the panel wants confidence that Telegram notifications and YML backup changes do not require updating existing agent deployments unless a future plan explicitly says so.

**Why this priority**: The user is already running production agents. Avoiding unnecessary agent updates lowers rollout risk.

**Independent Test**: Can be tested by updating only the panel in an environment with existing agents and confirming existing metrics, public status, and network probe behavior continue to work.

**Acceptance Scenarios**:

1. **Given** existing agents are reporting to the panel, **When** the panel is updated with this feature, **Then** agents continue reporting without configuration changes.
2. **Given** the feature needs a behavior that cannot work without agent changes, **When** planning identifies that dependency, **Then** the plan must explicitly call it out before implementation begins.

### Edge Cases

- Telegram credentials are syntactically invalid or point to a bot that cannot send messages.
- Telegram destination is a group or channel where the bot lacks permission to post.
- Telegram message size limits require a long alert or node list to be shortened or paginated.
- A Telegram request comes from an unbound or revoked user.
- A configured Telegram destination is read-only and attempts to use administrator-only menu actions.
- A read-only Telegram destination is configured for a subset of nodes or alert levels and an event occurs outside that scope.
- Notification storm conditions produce too many Telegram messages in a short period.
- YML export is requested while configuration changes are being made.
- YML import contains duplicate node names, duplicate probe names, missing references, or references to deleted resources.
- YML import targets a deployment with existing records that conflict with imported records.
- YML import targets a deployment with records that are not present in the backup.
- YML import contains a stable identifier that matches an existing record but the display name has changed.
- A backup was created by an older or newer panel version.
- Sensitive values are unavailable for export because they are intentionally stored or displayed in a protected form.
- An encrypted backup is imported with a missing or incorrect decryption credential.
- A restore is partially invalid; valid sections must not be applied without administrator confirmation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow administrators to configure Telegram as a notification channel with a test-send action and a visible delivery status.
- **FR-002**: System MUST restrict Telegram notifications and menu responses to administrator-approved Telegram chat ID allowlist entries configured in the panel.
- **FR-003**: System MUST deliver monitoring notifications to Telegram with enough context for the administrator to identify the affected node, alert type, severity, time, and current state.
- **FR-004**: System MUST prevent sensitive Telegram connection details from being exposed in ordinary views, error messages, or logs shown to users.
- **FR-005**: System MUST provide administrator Telegram bindings with a bot menu covering at minimum: status overview, alert summary, node list, node detail summary, notification mute or restore, settings/help, and a way to verify the current binding.
- **FR-006**: Telegram menu actions that expose monitoring data MUST require authorization; unauthorized users MUST NOT receive node, alert, probe, or public dashboard details.
- **FR-007**: Telegram menu actions in the first release MUST be limited to read-only status queries and low-risk notification controls; destructive administrative changes MUST remain outside the bot menu unless separately specified.
- **FR-008**: System MUST support read-only Telegram notification destinations, such as ordinary users, groups, or channels, that can receive allowed non-sensitive alert summaries but cannot access privileged bot menu actions or sensitive administrative data.
- **FR-009**: Read-only Telegram notification destinations MUST support configurable delivery scope by node and alert level.
- **FR-010**: System MUST provide a YML configuration export that includes all supported configurable areas currently managed by the panel, including systems, alert settings, notification channels, public dashboard visibility, network probe definitions and assignments, and public probe visibility selections.
- **FR-011**: YML exports MUST include a backup version, source panel version, creation time, and section-level metadata so future imports can validate compatibility.
- **FR-012**: System MUST provide a YML import or restore flow with validation and preview before applying changes.
- **FR-013**: YML import MUST report create, update, skip, conflict, and error decisions in a way an administrator can understand before final confirmation.
- **FR-014**: System MUST encrypt sensitive configuration values when they are included in a YML export; unencrypted exports MUST NOT contain plaintext sensitive values.
- **FR-015**: System MUST require a valid decryption credential before importing encrypted sensitive values, and MUST reject encrypted sensitive sections when the credential is missing or invalid.
- **FR-016**: YML import MUST use merge restore by default: create missing configuration and update matched configuration while preserving target-only configuration.
- **FR-017**: System MUST NOT delete target configuration during import unless a future explicitly specified mode or administrator action enables deletion.
- **FR-018**: YML import MUST match existing configuration primarily by stable identifiers stored in the backup; display names MUST be used for human-readable preview and conflict reporting, not as the primary identity.
- **FR-019**: System MUST support backward-compatible handling for backups exported before future schema additions; unknown sections MUST be reported and handled safely.
- **FR-020**: This feature MUST NOT require beszel-agent changes for its baseline notification and YML backup behavior. Any later agent dependency MUST be documented before implementation.
- **FR-021**: Changed behavior MUST be covered by focused unit tests.
- **FR-022**: If HTTP APIs are added or changed, they MUST be RESTful or explicitly justify an existing non-REST contract.

### API Contracts *(include if HTTP API changes)*

- **Resources**: Authenticated resources for notification channels, Telegram bindings, Telegram test delivery, configuration backups, backup validation, and backup restore previews.
- **Methods**: Read operations must be side-effect-free; create and update operations must validate administrator intent; restore application must require explicit confirmation after preview.
- **Status Codes**: Successful reads, creates, updates, validation failures, authorization failures, and conflict responses must be distinguishable by clients and administrators.
- **Schemas**: Request and response data must identify notification destination state, Telegram binding state, backup metadata, supported sections, validation warnings, restore actions, and conflict details.
- **Compatibility**: Existing notification, public dashboard, network probe, and system configuration behavior must continue to work for deployments that do not enable Telegram or YML restore.

### Key Entities *(include if feature involves data)*

- **Telegram Notification Channel**: Administrator-managed delivery configuration for Telegram alerts, including display name, enabled state, authorized destination metadata, delivery health, and protected connection data.
- **Telegram Binding**: A panel-configured Telegram chat ID allowlist entry with a role, such as administrator binding for privileged menu access or read-only destination for scoped notification-only delivery.
- **Bot Menu Action**: A permitted interaction initiated from Telegram, such as status overview, node summary, alert summary, mute, restore notifications, or help.
- **Configuration Backup**: A versioned YML document containing export metadata and supported configurable sections, with encrypted sensitive values when a full restorable export is requested.
- **Backup Section**: A named portion of the configuration backup, such as systems, alerting, notifications, public dashboard settings, network probes, probe assignments, or public probe visibility.
- **Restore Preview**: A validation result showing what a merge import would create, update by stable identifier, preserve, skip, reject, or require administrator action to resolve.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An administrator can configure Telegram delivery and receive a test message in under 3 minutes using valid Telegram details.
- **SC-002**: 100% of unauthorized Telegram requests tested during validation disclose no node names, metrics, alert details, probe details, or backup data.
- **SC-003**: At least 95% of normal alert notifications reach the configured Telegram destination within 30 seconds when Telegram is available.
- **SC-004**: A representative configuration containing systems, alerts, public dashboard settings, network probes, and notification settings can be exported and restored into a test instance with no undocumented configuration loss.
- **SC-005**: Import preview identifies all intentional updates, target-only preserved records, conflicts, skipped sections, encrypted sensitive-value actions, and decryption failures before any restore is applied.
- **SC-006**: Existing agents continue reporting after a panel-only update in validation, with no required agent-side configuration changes.

## Assumptions

- Telegram integration is intended for administrators, not public visitors.
- Telegram may be used as both a notification destination and a limited interactive status surface.
- Baseline bot menu actions should avoid destructive or high-risk administration.
- YML backup covers configuration, not historical runtime telemetry, metric samples, probe result history, logs, or generated charts.
- Sensitive values are expected to be restorable through encrypted export/import, never through plaintext YML fields.
- Existing authentication and administrator permissions remain the source of truth for who can configure notifications, export backups, or restore backups.
- The previous public probe visibility change did not require source changes under the `agent/` directory; this feature should preserve that panel-only baseline unless planning proves otherwise.
