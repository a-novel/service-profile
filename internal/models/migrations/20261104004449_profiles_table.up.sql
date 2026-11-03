CREATE TABLE profiles (
  -- A UUIDv7 sorts by creation time, so the primary key also orders profiles by age.
  id uuid PRIMARY KEY NOT NULL DEFAULT uuidv7(),
  /* The user who owns the profile, from the authentication service. A user holds one profile. */
  user_id uuid NOT NULL CONSTRAINT profiles_user_id_key UNIQUE,
  /* The name the profile shows. Null until the user sets one. */
  username text CONSTRAINT profiles_username_check CHECK (
    username <> ''
    AND octet_length(username) <= 2048
  ),
  /* The picture's encoded bytes, a square WebP or JPEG. */
  picture bytea CONSTRAINT profiles_picture_check CHECK (octet_length(picture) <= 262144),
  /* The media type the picture is encoded in. */
  picture_type text CONSTRAINT profiles_picture_type_check CHECK (picture_type IN ('image/webp', 'image/jpeg')),
  /* When the picture last changed. Picture URLs carry it as their version. */
  picture_updated_at timestamp with time zone,
  -- Full precision lets updated_at distinguish two updates made within the same second.
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- A picture comes and goes with its type and its version.
  CONSTRAINT profiles_picture_set_check CHECK (
    (picture IS NULL) = (picture_type IS NULL)
    AND (picture IS NULL) = (picture_updated_at IS NULL)
  )
);
