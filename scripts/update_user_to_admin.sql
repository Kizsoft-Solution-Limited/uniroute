-- Script to update a user to admin role
-- Usage: Update the email address below and run this script

-- STEP 1: Check current state
SELECT 
  id, 
  email, 
  name,
  roles,
  pg_typeof(roles) as roles_type,
  array_length(roles, 1) as roles_count,
  'user' = ANY(roles) as has_user_role,
  'admin' = ANY(roles) as has_admin_role
FROM users 
WHERE email = 'adikekizinho@gmail.com';

-- STEP 2: Update the user to have both 'user' and 'admin' roles
-- IMPORTANT: Use ARRAY['user', 'admin']::TEXT[] syntax
UPDATE users 
SET roles = ARRAY['user', 'admin']::TEXT[],
    updated_at = NOW()
WHERE email = 'adikekizinho@gmail.com';

-- STEP 3: Verify the update worked
SELECT 
  id, 
  email, 
  name,
  roles,
  array_length(roles, 1) as roles_count,
  'user' = ANY(roles) as has_user_role,
  'admin' = ANY(roles) as has_admin_role
FROM users 
WHERE email = 'adikekizinho@gmail.com';

-- Expected result: has_admin_role should be TRUE
-- Expected roles: {"user","admin"} or ['user', 'admin']

