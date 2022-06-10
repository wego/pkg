CREATE TABLE IF NOT EXISTS auth_roles
(
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);


CREATE TABLE IF NOT EXISTS auth_users
(
    id         BIGSERIAL PRIMARY KEY,
    email      TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth_user_roles
(
    id         BIGSERIAL PRIMARY KEY,
    role_id    BIGINT REFERENCES auth_roles (id) NOT NULL,
    user_id    BIGINT REFERENCES auth_users (id) NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth_role_permissions
(
    id         BIGSERIAL PRIMARY KEY,
    role_id    BIGINT REFERENCES auth_roles (id) NOT NULL,
    resource   TEXT                              NOT NULL,
    method     TEXT                              NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_roles_name ON auth_roles (name) WHERE deleted_at IS NULL;;
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_users_email ON auth_users (email) WHERE deleted_at IS NULL;;
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_user_roles_role_id_user_id ON auth_user_roles (role_id, user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_role_permissions_role_id_resource_method ON auth_role_permissions (role_id, resource, method) WHERE deleted_at IS NULL;
