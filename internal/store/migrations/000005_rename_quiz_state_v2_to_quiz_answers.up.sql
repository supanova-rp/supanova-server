ALTER TABLE user_quiz_state RENAME COLUMN quiz_state_v2 TO quiz_answers;
ALTER TABLE user_quiz_state ALTER COLUMN quiz_answers SET DEFAULT '[]'::jsonb;
