ALTER TABLE user_quiz_state ALTER COLUMN quiz_answers SET DEFAULT '{}'::jsonb;
ALTER TABLE user_quiz_state RENAME COLUMN quiz_answers TO quiz_state_v2;
