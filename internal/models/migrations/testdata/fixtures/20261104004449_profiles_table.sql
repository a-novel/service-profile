-- A fixture runs after its matching migration and seeds data for the next migration.
INSERT INTO
  profiles (id, user_id, username)
VALUES
  (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    'fixture profile'
  );
