<<<<<<<< HEAD:migrations/versioned/000110_knowledge_base_desensitization_config.up.sql
-- Migration 000110: opt-in per-KB document desensitization configuration.
========
-- Migration 000108: opt-in per-KB document desensitization configuration.
>>>>>>>> 234ac35 (fix(kb): bump desensitization migrations past message_artifacts_deleted_at):migrations/versioned/000108_knowledge_base_desensitization_config.up.sql
ALTER TABLE knowledge_bases ADD COLUMN IF NOT EXISTS desensitization_config JSONB;
