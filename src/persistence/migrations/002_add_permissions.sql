-- add table to grant users permissions to orgs --
CREATE TABLE org_permissions (
    user_id INT NOT NULL REFERENCES users(id),
    org_id VARCHAR(255) NOT NULL,
    permission VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- add primary key to org_permissions --
ALTER TABLE org_permissions ADD PRIMARY KEY (user_id, org_id);
