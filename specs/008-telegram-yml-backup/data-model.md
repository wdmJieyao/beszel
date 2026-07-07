# Data Model: Telegram Notifications and YML Configuration Backup

## Telegram Bot Settings

Singleton panel-level configuration controlling whether Telegram integration is
available.

**Fields**

- `id`: stable record identifier.
- `enabled`: whether the hub should send Telegram messages and process bot
  menu updates.
- `botTokenEncrypted`: encrypted Telegram bot token.
- `botUsername`: optional display value resolved from Telegram for operator
  confirmation.
- `pollingEnabled`: whether hub long polling is enabled.
- `lastPollOffset`: Telegram update offset used to avoid reprocessing updates.
- `lastError`: latest non-sensitive delivery or polling error.
- `created`, `updated`: audit timestamps.

**Validation**

- Only administrators can create or update settings.
- Token is never returned in plaintext by ordinary read APIs.
- Disabling settings stops menu polling and delivery without deleting
  destinations.

## Telegram Destination

Panel-managed chat ID allowlist entry.

**Fields**

- `id`: stable record identifier exported in backups.
- `user`: optional owner/admin user relation for menu personalization and
  alert ownership.
- `name`: operator-visible destination name.
- `chatId`: Telegram chat ID. May represent private chat, group, supergroup, or
  channel.
- `chatType`: `private`, `group`, `supergroup`, `channel`, or `unknown`.
- `role`: `admin` or `read_only`.
- `enabled`: delivery enabled flag.
- `nodeScope`: list of system IDs allowed for read-only delivery; empty means
  all systems visible to the owner/admin scope.
- `alertLevelScope`: alert levels or categories allowed for read-only delivery.
- `muteUntil`: optional timestamp for temporary mute.
- `lastTestAt`: timestamp for last successful test.
- `lastDeliveryAt`: timestamp for last successful alert delivery.
- `lastError`: latest non-sensitive delivery error.
- `created`, `updated`: audit timestamps.

**Validation**

- `chatId` must be unique per destination.
- `role=admin` can use menu actions; `role=read_only` cannot.
- Read-only destinations must receive sanitized alert summaries only.
- Node and alert-level scopes must reference existing systems/known alert
  classes at validation time.

**State transitions**

- `disabled` -> `enabled`: destination becomes eligible for delivery.
- `enabled` -> `muted`: temporary mute suppresses delivery until `muteUntil`.
- `muted` -> `enabled`: delivery resumes when mute expires or is cleared.
- Any state -> `error`: last failure recorded without exposing token or secret.

## Bot Menu Action

An inbound Telegram command or callback processed by the hub.

**Attributes**

- `chatId`: source Telegram chat.
- `action`: `status_overview`, `alert_summary`, `node_list`, `node_detail`,
  `mute_notifications`, `restore_notifications`, `settings_help`,
  `verify_binding`, or `help`.
- `roleRequired`: `admin` unless explicitly safe for read-only.
- `responseSensitivity`: `non_sensitive_summary` or `admin_detail`.

**Rules**

- Unknown or unauthorized chats receive no monitoring data.
- Read-only chats cannot invoke admin menu actions.
- Long responses are paginated or summarized to fit Telegram limits.

## Configuration Backup Document

Versioned YML document for portable panel configuration.

**Top-level sections**

- `meta`: backup version, source panel version, created timestamp, instance
  label, and export options.
- `encryption`: algorithm, key derivation metadata, salt, nonce, and encrypted
  section references when secrets are included.
- `users`: non-auth user references needed for ownership mapping, such as email
  and stable ID. Passwords, sessions, MFA, and OAuth data are excluded.
- `systems`: system records and related fingerprint/token configuration.
- `alerts`: alert definitions and quiet-hour windows.
- `notifications`: existing email/webhook settings plus Telegram settings and
  destinations.
- `publicStatus`: public dashboard visibility, metric flags, display names, and
  public probe selections.
- `networkProbes`: probe definitions and probe-system assignments.

**Out of scope**

- Runtime metric history, alert history, network probe result history, logs,
  auth sessions, OAuth secrets, password hashes, generated charts, and agent
  runtime cache.

## Backup Section

Named, versioned unit inside the backup document.

**Fields**

- `name`: section identifier.
- `version`: section schema version.
- `items`: ordered configuration items.
- `warnings`: optional export warnings, such as skipped unavailable secrets.

**Validation**

- Unknown sections are reported during preview and skipped unless explicitly
  supported later.
- Newer major section versions are rejected or skipped with a warning.

## Backup Secret Envelope

Encrypted representation of sensitive backup values.

**Fields**

- `ciphertext`: encrypted bytes encoded for YML.
- `nonce`: nonce for authenticated encryption.
- `salt`: key derivation salt.
- `kdf`: key derivation metadata.
- `algorithm`: authenticated encryption algorithm identifier.
- `contentType`: identifies the protected value or section.

**Sensitive values**

- Telegram bot tokens.
- Webhook/Shoutrrr URLs containing credentials.
- Agent or fingerprint tokens.
- Any future notification credentials.

**Validation**

- Plaintext export must omit or redact sensitive values.
- Encrypted import requires the correct decryption credential.
- Decryption failure blocks applying encrypted sensitive sections.

## Restore Preview

Dry-run result for an import.

**Fields**

- `backupMeta`: parsed backup metadata.
- `mode`: `merge`.
- `items`: list of per-record decisions.
- `summary`: counts for create, update, preserve, skip, conflict, error.
- `requiresCredential`: whether encrypted values need a credential.
- `warnings`: compatibility and scope warnings.

**Decision types**

- `create`: record does not exist and can be created.
- `update`: stable identifier matches an existing record and can be updated.
- `preserve`: target-only record is not in the backup and will remain.
- `skip`: item is unsupported or intentionally omitted.
- `conflict`: item needs administrator action before apply.
- `error`: item cannot be applied.

## Restore Apply Request

Administrator-confirmed application of a previously previewed backup.

**Fields**

- `backup`: YML document content or upload reference.
- `mode`: must be `merge` for this feature.
- `decryptionCredential`: required only when encrypted sensitive values are
  present.
- `confirmPreviewId`: preview identifier or checksum to prevent applying a
  different file than the reviewed preview.

**Rules**

- Apply must not delete target-only configuration.
- Apply must use stable identifiers before display names.
- Apply should be transactional per section where practical; failures must be
  reported with enough detail to understand partial progress.
