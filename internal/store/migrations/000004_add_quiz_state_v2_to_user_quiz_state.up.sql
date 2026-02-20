ALTER TABLE user_quiz_state ADD COLUMN quiz_state_v2 JSONB NOT NULL DEFAULT '{}'::jsonb;
