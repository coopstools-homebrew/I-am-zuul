package queries

const (
	GET_VERSION = `
		SELECT version FROM version_tracking 
		ORDER BY created_at DESC LIMIT 1
	`
	INSERT_VERSION = `
		INSERT INTO version_tracking (version) 
		VALUES ($1)
	`
)
