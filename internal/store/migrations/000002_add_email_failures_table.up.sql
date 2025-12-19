CREATE TABLE email_failures (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  error TEXT NOT NULL,
  params JSONB NOT NULL DEFAULT '{}'::jsonb,
  retries INT NOT NULL DEFAULT 5 
);