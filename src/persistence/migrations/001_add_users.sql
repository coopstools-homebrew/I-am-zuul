-- after initialization, migrations should not be idempotent --

-- add users table using the github id as the primary key --
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    login_name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- TODO: make login_name the primary key --
