CREATE TABLE IF NOT EXISTS roles (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT
);
INSERT INTO
  roles (name, description)
VALUES
  (
    'preview user',
    'A user can create posts and comments'
    
  );
INSERT INTO
  roles (name, description)
VALUES
  (
    'user',
    'A user can create posts and comments'
    
  );

INSERT INTO
  roles (name, description)
VALUES
  (
    'moderator',
    'A moderator can update other users posts'
  );

INSERT INTO
  roles (name, description)
VALUES
  (
    'admin',
    'An admin can update and delete other users posts'
  );

  ALTER TABLE
  IF EXISTS users
ADD
  COLUMN role_id INT REFERENCES roles(id) DEFAULT 2;