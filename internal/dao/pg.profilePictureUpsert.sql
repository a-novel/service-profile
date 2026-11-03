-- The first write of a user creates their profile.
INSERT INTO
  profiles (
    id,
    user_id,
    picture,
    picture_type,
    picture_updated_at,
    created_at,
    updated_at
  )
VALUES
  (?0, ?1, ?2, ?3, ?4, ?4, ?4)
ON CONFLICT (user_id) DO UPDATE
SET
  picture = EXCLUDED.picture,
  picture_type = EXCLUDED.picture_type,
  picture_updated_at = EXCLUDED.picture_updated_at,
  updated_at = EXCLUDED.updated_at
RETURNING
  id,
  username,
  picture_updated_at;
