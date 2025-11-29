CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  name TEXT,
  email TEXT
);

CREATE TABLE IF NOT EXISTS courses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  title TEXT,
  description TEXT,
  completion_title TEXT,
  completion_message TEXT
);

CREATE TABLE IF NOT EXISTS course_materials (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  course_id UUID NOT NULL,
  storage_key UUID NOT NULL,
  name TEXT NOT NULL,
  position INT,

  CONSTRAINT fk_courses FOREIGN KEY(course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS videosections (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  title TEXT,
  position INT,
  storage_key UUID,
  course_id UUID,

  CONSTRAINT fk_courses FOREIGN KEY(course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS quizsections (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  position INT,
  course_id UUID,

  CONSTRAINT fk_courses FOREIGN KEY(course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS quizquestions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  question TEXT,
  position INT,
  quiz_section_id UUID,
  is_multi_answer BOOLEAN DEFAULT FALSE NOT NULL,

  CONSTRAINT fk_quizsections FOREIGN KEY(quiz_section_id) REFERENCES quizsections(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS quizanswers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  answer TEXT,
  correct_answer BOOLEAN,
  quiz_question_id UUID,
  position INT,

  CONSTRAINT fk_quizquestions FOREIGN KEY(quiz_question_id) REFERENCES quizquestions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS usercourses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  user_id TEXT,
  course_id UUID NOT NULL,

  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_courses FOREIGN KEY(course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS userprogress (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  user_id TEXT NOT NULL,
  course_id UUID NOT NULL,
  completed_section_ids UUID[] NOT NULL,
  completed_intro BOOLEAN DEFAULT FALSE,
  completed_course BOOLEAN DEFAULT FALSE,

  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_courses FOREIGN KEY(course_id) REFERENCES courses(id) ON DELETE CASCADE,
  CONSTRAINT userprogress_user_course_unique UNIQUE (user_id, course_id)
);

CREATE TABLE IF NOT EXISTS user_quiz_state (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  user_id TEXT NOT NULL,
  quiz_id UUID NOT NULL,
  quiz_state JSONB NOT NULL DEFAULT '{}'::jsonb,
  attempts INT NOT NULL DEFAULT 0,

  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_quizsections FOREIGN KEY(quiz_id) REFERENCES quizsections(id) ON DELETE CASCADE,
  CONSTRAINT userquizstate_user_quiz_unique UNIQUE (user_id, quiz_id)
);
