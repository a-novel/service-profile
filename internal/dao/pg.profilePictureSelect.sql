SELECT
  picture,
  picture_type,
  picture_updated_at
FROM
  profiles
WHERE
  id = ?0
  AND picture IS NOT NULL;
