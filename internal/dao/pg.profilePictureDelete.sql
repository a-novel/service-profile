UPDATE profiles
SET
  picture = NULL,
  picture_type = NULL,
  picture_updated_at = NULL,
  updated_at = ?1
WHERE
  user_id = ?0
RETURNING
  id,
  username,
  picture_updated_at;
