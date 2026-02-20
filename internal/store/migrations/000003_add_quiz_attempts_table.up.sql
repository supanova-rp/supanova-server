CREATE TABLE quiz_attempts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  user_id TEXT NOT NULL,
  quiz_id UUID NOT NULL,
  attempt_data JSONB NOT NULL DEFAULT '{}'::jsonb,
  attempt_number INT NOT NULL,

  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_quizsections FOREIGN KEY(quiz_id) REFERENCES quizsections(id) ON DELETE CASCADE,
  CONSTRAINT quiz_attempt_history_user_quiz_attempt_unique UNIQUE (user_id, quiz_id, attempt_number)
);
