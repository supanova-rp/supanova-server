CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  name TEXT,
  email TEXT
);

CREATE TABLE IF NOT EXISTS courses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  title TEXT,
  description TEXT
);

CREATE TABLE IF NOT EXISTS videosections (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  title TEXT,
  video_url TEXT,
  position INT,
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
  completed_course BOOLEAN DEFAULT FALSE,

  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_courses FOREIGN KEY(course_id) REFERENCES courses(id) ON DELETE CASCADE,
  CONSTRAINT userprogress_user_course_unique UNIQUE (user_id, course_id)
);
