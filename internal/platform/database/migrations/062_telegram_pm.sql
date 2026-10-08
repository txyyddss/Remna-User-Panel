ALTER TABLE users ADD COLUMN pm_blocked INTEGER NOT NULL DEFAULT 0 CHECK (pm_blocked IN (0,1));
ALTER TABLE users ADD COLUMN pm_muted INTEGER NOT NULL DEFAULT 0 CHECK (pm_muted IN (0,1));

CREATE TABLE telegram_pm_conversations (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chat_id INTEGER NOT NULL CHECK (chat_id<0),
    topic_id INTEGER CHECK (topic_id>1),
    profile_message_id INTEGER CHECK (profile_message_id>0),
    profile_state TEXT NOT NULL DEFAULT 'missing' CHECK (profile_state IN ('missing','ready','pending_review')),
    topic_state TEXT NOT NULL DEFAULT 'new' CHECK (topic_state IN ('new','creating','ready','pending_review')),
    topic_operation_id TEXT REFERENCES provider_operations(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (user_id,chat_id)
);
CREATE UNIQUE INDEX telegram_pm_topic_idx ON telegram_pm_conversations(chat_id,topic_id) WHERE topic_id IS NOT NULL;
CREATE INDEX telegram_pm_conversation_page_idx ON telegram_pm_conversations(created_at DESC,id DESC);

-- Serialize append-only profile creation and explicit repair for a conversation.
CREATE UNIQUE INDEX telegram_pm_profile_attempt_idx ON provider_operation_items(
    CASE WHEN target_type='pm_topic_recovery' THEN substr(target_id,1,instr(target_id,':')-1) ELSE target_id END
) WHERE status='processing' AND ((target_type='pm_conversation' AND item_key='profile') OR target_type='pm_topic_recovery');

-- Bounded webhook tombstones contain no Telegram message bodies.
CREATE TABLE telegram_pm_updates (
    update_id INTEGER PRIMARY KEY,
    created_at TEXT NOT NULL
);
CREATE INDEX telegram_pm_update_retention_idx ON telegram_pm_updates(created_at);
