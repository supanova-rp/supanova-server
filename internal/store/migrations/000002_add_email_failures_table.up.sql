CREATE TABLE email_failures (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  error TEXT NOT NULL,
  template_name TEXT NOT NULL,
  template_params JSONB NOT NULL DEFAULT '{}'::jsonb,
  email_name TEXT NOT NULL,
  retries INT NOT NULL DEFAULT 5 
);