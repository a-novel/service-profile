SELECT
  id,
  username,
  picture_updated_at
FROM
  profiles
WHERE
  user_id = ?0;
