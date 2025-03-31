-- initialize the database with version tracking table --
CREATE TABLE IF NOT EXISTS version_tracking (
    version INTEGER PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- insert the initial version if it doesn't exist --
INSERT INTO version_tracking (version)
SELECT 0
WHERE NOT EXISTS (
    SELECT 1 FROM version_tracking WHERE version = 0
);
