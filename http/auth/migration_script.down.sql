DROP TABLE IF EXISTS auth_role_permissions;
DROP TABLE IF EXISTS auth_user_roles;
DROP TABLE IF EXISTS auth_users;
DROP TABLE IF EXISTS auth_roles;
DROP INDEX IF EXISTS idx_auth_roles_name;
DROP INDEX IF EXISTS idx_auth_users_email;
DROP INDEX IF EXISTS idx_auth_user_roles_role_id_user_id;
DROP INDEX IF EXISTS idx_auth_role_permissions_role_id_resource_method;
