package repository

const (
	createUserQuery = `INSERT INTO users (first_name, last_name, email, password, role, avatar) 
		VALUES ($1, $2, $3, $4, $5, COALESCE(NULLIF($6, ''), null)) 
		RETURNING user_id, first_name, last_name, email, password, avatar, created_at, updated_at, role`

	findByEmailQuery = `SELECT user_id, email, first_name, last_name, role, avatar, password, created_at, updated_at FROM users WHERE email = $1`

	findByIDQuery = `SELECT user_id, email, first_name, last_name, role, avatar, created_at, updated_at FROM users WHERE user_id = $1`

	updateByIDQuery = `UPDATE users SET first_name = $2, last_name = $3, email = $4, password = $5, role = $6, avatar = $7) WHERE user_id = $1
		RETURNING user_id, first_name, last_name, email, password, avatar, created_at, updated_at, role`

	deleteByIDQuery = `DELETE FROM users WHERE user_id = $1`
)
