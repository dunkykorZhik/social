CREATE TABLE IF NOT EXISTS posts (
  id bigserial PRIMARY KEY,
  title text NOT NULL,
  user_id bigint NOT NULL,
  content text NOT NULL,
  tags VARCHAR(100) [],
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);


ALTER TABLE
  posts
ADD
  CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id);


ALTER TABLE
  posts
ADD
  COLUMN version INT DEFAULT 0;