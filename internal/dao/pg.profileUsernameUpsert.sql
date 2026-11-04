-- The first write of a user creates their profile.
INSERT INTO
  profiles (id, user_id, username, created_at, updated_at)
VALUES
  (?0, ?1, ?2, ?3, ?3)
ON CONFLICT (user_id) DO UPDATE
SET
  username = EXCLUDED.username,
  updated_at = EXCLUDED.updated_at
RETURNING
  id,
  username,
  picture_updated_at;
