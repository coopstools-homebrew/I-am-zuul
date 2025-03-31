-- after initialization, migrations should not be idempotent --

-- add users table using the github id as the primary key --
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    org_id VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- add table to grant users permissions to orgs --
CREATE TABLE org_permissions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    org_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
