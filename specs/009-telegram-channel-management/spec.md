# Feature Specification: Telegram Channel Management Improvements

**Feature Branch**: `[009-telegram-channel-management]`

**Created**: 2026-07-10

**Status**: Ready for planning

**Input**: User description: "Improve Telegram channel deletion, Bot testing, repeated Chat ID notification templates, node scope selection, role clarity, and load-average descriptions."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reliably Manage Bot And Channels (Priority: P1)

As an administrator, I can verify the Telegram Bot, understand which verification stage failed, and visibly delete an obsolete notification channel without guessing what an icon does.

**Why this priority**: A misleading Bot failure and an undiscoverable deletion action prevent basic administration and make a working integration appear broken.

**Independent Test**: Configure a valid Bot, run verification, delete one destination with confirmation, and verify that the Bot remains configured while the deleted destination no longer receives messages.

**Acceptance Scenarios**:

1. **Given** a valid saved Bot credential, **When** the administrator runs Bot verification, **Then** credential connectivity and command-menu initialization are checked and reported separately.
2. **Given** Telegram accepts command-menu initialization with a boolean success result, **When** verification completes, **Then** the interface reports success instead of a response-format error.
3. **Given** a saved notification channel, **When** the administrator selects its clearly labelled delete action, **Then** a confirmation identifies the channel before deletion.
4. **Given** a deletion request fails, **When** the operation returns, **Then** the channel remains visible and the administrator receives an actionable error.
5. **Given** a channel is deleted successfully, **When** the destination list refreshes, **Then** that channel is absent while the shared Bot integration remains unchanged.

---

### User Story 2 - Configure Multiple Delivery Policies For One Chat (Priority: P1)

As an administrator, I can use the same Telegram conversation for multiple notification policies without creating ambiguous command permissions or receiving duplicate copies of the same alert.

**Why this priority**: A single Chat ID often represents one operations group or channel, while different nodes and alert categories require separately maintainable routing policies.

**Independent Test**: Configure two non-overlapping policies for the same Chat ID, trigger matching alerts for each policy, and verify correct delivery, editing, muting, and command authorization behavior.

**Acceptance Scenarios**:

1. **Given** an existing Chat ID, **When** the administrator adds another policy for different nodes or alert categories, **Then** the system accepts the policy according to the confirmed channel-policy model.
2. **Given** multiple policies for one Chat ID, **When** one alert matches more than one policy, **Then** the chat receives no more than one copy of that alert.
3. **Given** multiple policies for one Chat ID, **When** Telegram sends a menu command, **Then** exactly one unambiguous channel-level permission decision is applied.
4. **Given** one policy is deleted, **When** other policies remain for the Chat ID, **Then** the remaining policies and channel authorization continue to work.

---

### User Story 3 - Understand And Select Scope (Priority: P2)

As an administrator managing many nodes, I can choose all nodes or a selected subset efficiently and understand whether future nodes are included.

**Why this priority**: The current empty selection silently means all nodes, while the interface presents only individual checkboxes and does not explain that behavior.

**Independent Test**: Switch between all-node and selected-node modes, use select-all and clear controls, save, reload, and add a new node to verify the documented behavior.

**Acceptance Scenarios**:

1. **Given** the all-node mode is selected, **When** a new node is added later, **Then** the notification policy automatically includes it.
2. **Given** selected-node mode, **When** the administrator chooses select all, **Then** all currently listed nodes become selected and the selected count is visible.
3. **Given** selected-node mode, **When** the administrator clears the selection, **Then** the interface blocks an ambiguous save or explicitly explains the resulting behavior.
4. **Given** an existing destination with no stored node IDs, **When** it is opened after upgrade, **Then** it is shown as all-node mode without changing production behavior.
5. **Given** an existing destination with stored node IDs, **When** it is opened after upgrade, **Then** it is shown as selected-node mode with the same nodes selected.

---

### User Story 4 - Understand Roles And Alert Categories (Priority: P2)

As an administrator, I can understand what each role can see and do, and I can understand what the 1-, 5-, and 15-minute load categories measure before saving a policy.

**Why this priority**: Controls that appear to have no effect create unsafe assumptions about sensitive information and unreliable notification routing.

**Independent Test**: Compare role descriptions and available controls, then configure each load category and verify that its label and help text identify the correct system load-average window.

**Acceptance Scenarios**:

1. **Given** the role selector is visible, **When** the administrator compares roles, **Then** the interface explains menu access, message detail, link/address redaction, supported chat types, and scope behavior.
2. **Given** a role makes a control ineffective, **When** that role is selected, **Then** the control is hidden or disabled with an explanation rather than silently ignored.
3. **Given** the load alert categories, **When** the administrator reads their labels, **Then** they are identified as system load averages over 1, 5, and 15 minutes, not percentages or alert durations.
4. **Given** a load threshold is configured, **When** help is displayed, **Then** it explains that the value is an absolute average and that CPU logical-core count is the practical reference point.

### Edge Cases

- A Bot credential is valid, but command-menu registration fails due to permissions or an upstream Telegram error.
- The administrator tests a newly entered unsaved credential while a different credential is already stored.
- A destination is being tested while another administrator deletes or edits it.
- Multiple policies for one Chat ID overlap completely, partially, or not at all.
- The last policy for a Chat ID is deleted while that Chat ID has administrator menu access.
- There are no nodes, hundreds of nodes, or nodes are added after an all-node policy is saved.
- Existing production records have an empty node scope or previously ignored alert scopes.
- Load-average data is unavailable for an offline or newly added node.
- A channel or group is assigned an administrator role even though interactive administrator commands are permitted only in private chats.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The destination list MUST provide a visible, labelled deletion action with an accessible name and explanatory hover text.
- **FR-002**: Deleting a destination MUST require confirmation that includes the destination name and masked Chat ID.
- **FR-003**: Successful destination deletion MUST remove only that destination or policy and MUST NOT remove the shared Bot credential or unrelated destinations.
- **FR-004**: Bot verification MUST distinguish credential connectivity, Bot identity retrieval, and command-menu initialization in its result.
- **FR-005**: Bot verification MUST accept every documented Telegram success-result shape and MUST surface sanitized, stage-specific failure information.
- **FR-006**: When an unsaved Token is present, Bot verification MUST clearly test that entered Token; otherwise it MUST test the saved Token without requiring re-entry.
- **FR-007**: Bot verification MUST NOT silently replace the saved Token. Saving credentials remains a separate explicit action.
- **FR-008**: Each Telegram Chat ID MUST identify one unique channel, and each channel MUST support multiple independently named notification policies.
- **FR-009**: Multiple policies matching the same alert and Chat ID MUST produce at most one delivered message for that alert.
- **FR-010**: Channel-level command authorization, mute state, health state, and delivery-policy state MUST remain unambiguous when one Chat ID has multiple policies.
- **FR-011**: Node scope MUST offer an explicit all-node mode that automatically includes nodes added in the future.
- **FR-012**: Node scope MUST offer a selected-node mode with select-all-current, clear-all, selected-count, and searchable or scrollable node selection suitable for at least 500 nodes.
- **FR-013**: Existing empty node scopes MUST migrate or render as all-node mode, while non-empty scopes MUST preserve their exact selected nodes.
- **FR-014**: The interface MUST explain role permissions next to the role selector and MUST prevent unsupported role/chat-type combinations.
- **FR-015**: Node-scope and alert-category filtering MUST apply consistently to both administrator and read-only notification policies. Role MUST affect only interactive menu authorization and sensitive-content redaction.
- **FR-016**: Administrator-level interactive menu access MUST be limited to an explicitly authorized private chat unless a separate secure group-member authorization mechanism is introduced.
- **FR-017**: Read-only delivery MUST continue to hide panel links, network addresses, and other administrator-only details.
- **FR-018**: The alert category list MUST label load-average categories as 1-minute, 5-minute, and 15-minute system load averages.
- **FR-019**: Load-average help MUST state that the value is an absolute average, not a percentage or notification duration, and MUST explain CPU logical-core count as a comparison reference.
- **FR-020**: Controls with no effect for the selected role MUST be hidden or disabled with a reason.
- **FR-021**: All behavior changes MUST include focused regression tests covering deletion, Bot verification result parsing, repeated Chat ID policy behavior, overlap de-duplication, node-scope modes, role enforcement, and load labels.
- **FR-022**: Existing production Bot settings and destination records MUST remain readable and retain their effective delivery behavior after upgrade.
- **FR-023**: Changed HTTP operations MUST remain resource-oriented, authenticated for administrators, and return predictable success, validation, conflict, not-found, and upstream-failure responses.

### API Contracts *(include if HTTP API changes)*

- **Resources**: Bot integration, Telegram channels, and notification policies are separate administrator-managed resources; one channel may own multiple policies.
- **Methods**: Reading uses safe retrieval; creation creates a new resource; partial editing updates one resource; deletion removes one identified resource; verification is an explicit non-persistent integration check.
- **Status Codes**: Successful reads and edits return success; deletion returns no content; invalid scope or role combinations return validation failure; missing records return not found; uniqueness conflicts return conflict; Telegram upstream failures return a sanitized verification result.
- **Schemas**: Responses MUST distinguish channel identity and permission fields from policy scope fields if they become separate entities.
- **Compatibility**: Existing destination records MUST be migrated without loss, duplicate delivery, or a temporary authorization gap. Existing clients MUST receive a documented compatibility response during any transition period.

### Key Entities *(include if feature involves data)*

- **Bot Integration**: The shared Telegram Bot identity, credential state, polling state, command-menu state, and sanitized health information.
- **Telegram Channel**: One Telegram Chat ID and chat type, with channel-level name, enabled state, role or permissions, mute state, test status, delivery health, and menu authorization.
- **Notification Policy**: A named set of node scope and alert-category scope rules that determines which alerts a Telegram channel receives; each channel may own multiple policies.
- **Node Scope**: Either dynamic all-node coverage or an explicit set of current node identifiers.
- **Alert Category**: A documented alert classification such as status, CPU, memory, disk, temperature, bandwidth, GPU, battery, or a specific load-average window.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An administrator can identify and delete a destination in no more than three interactions, with zero accidental deletion during confirmation testing.
- **SC-002**: A valid Bot verification completes with an accurate stage-by-stage result in at least 99% of attempts when Telegram is reachable.
- **SC-003**: Administrators can assign all current nodes to a selected-node policy with one action and can identify whether future nodes are included without consulting documentation.
- **SC-004**: For 500 listed nodes, scope selection remains usable without text overlap and an administrator can locate and select a named node within 10 seconds.
- **SC-005**: Every alert matching multiple policies for one Chat ID produces exactly one message in that chat.
- **SC-006**: In permission testing, read-only channels expose zero administrator-only links or network addresses and execute zero administrator menu commands.
- **SC-007**: In usability review, administrators correctly explain the difference between roles and the meaning of all three load-average categories after reading only the settings interface.
- **SC-008**: Upgrading a production dataset preserves 100% of existing enabled states, node scopes, alert scopes, mute states, and Chat IDs.

## Assumptions

- The reported Node Scope issue means the current interface lacks an efficient and explicit all-node selection mode.
- Empty node scope continues to mean all nodes, including future nodes, for backward compatibility.
- The requested deletion applies to an individual Telegram destination or notification policy, not deletion of the shared Bot integration.
- The Bot Token remains a single shared integration credential for the panel.
- Each Chat ID represents exactly one Telegram channel, while that channel may own multiple independently managed notification policies.
- Administrator and read-only policies both honor node and alert-category scopes; administrator role permits authorized private-chat menu access and full message details, while read-only role denies menu access and receives redacted content.
- Telegram groups and channels are primarily notification recipients; administrator menu commands remain private-chat-only unless secure sender-level authorization is separately designed.
- Load-average thresholds continue to use the operating system's absolute 1-, 5-, and 15-minute load averages.
- Historical telemetry, alert definitions themselves, and agent behavior are outside this feature's scope.
