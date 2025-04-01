package queries

const (
	GET_VERSION = `
		SELECT version FROM version_tracking 
		ORDER BY version DESC LIMIT 1
	`

	GET_VERSIONS = `
		SELECT version, created_at FROM version_tracking 
		ORDER BY version DESC
	`

	INSERT_VERSION = `
		INSERT INTO version_tracking (version) 
		VALUES ($1)
	`

	ADD_OR_UPDATE_USER = `
		INSERT INTO users (id, login_name, avatar_url, email) 
		VALUES ($1, $2, $3, $4) 
		ON CONFLICT (id) 
		DO UPDATE SET login_name = $2, avatar_url = $3, email = $4
	`

	GET_USER_BY_ID = `
		SELECT id, login_name, avatar_url, email FROM users WHERE id = $1
	`

	GET_ALL_USERS = `
		SELECT id, login_name, avatar_url, email FROM users
	`

	// Given a list of user ids, return a list of permissions for each user with their login_name
	GET_ALL_USER_PERMISSIONS = `
		SELECT user_id, org_id, permission, login_name FROM org_permissions 
		INNER JOIN users ON org_permissions.user_id = users.id 
		WHERE user_id = ANY($1)
	`

	ADD_OR_UPDATE_ORG_PERMISSION = `
		INSERT INTO org_permissions (user_id, org_id, permission) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id, org_id) 
		DO UPDATE SET permission = $3
	 `
)
