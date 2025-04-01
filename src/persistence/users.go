package persistence

import (
	"database/sql"

	"github.com/coopstools-homebrew/I-am-zuul/src/persistence/queries"
)

type UserInfo struct {
	ID        int32  `json:"id"`
	LoginName string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

type UserTable struct {
	db *sql.DB
}

func NewUserTable(db *sql.DB) *UserTable {
	return &UserTable{db: db}
}

func (ut *UserTable) UpdateUser(user *UserInfo) error {
	stmt, err := ut.db.Prepare(queries.ADD_OR_UPDATE_USER)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.ID, user.LoginName, user.AvatarURL, user.Email)
	return err
}

func (ut *UserTable) GetUserByID(id int32) (*UserInfo, error) {
	row := ut.db.QueryRow(queries.GET_USER_BY_ID, id)
	var user UserInfo
	err := row.Scan(&user.ID, &user.LoginName, &user.AvatarURL, &user.Email)
	return &user, err
}

func (ut *UserTable) GetAllUsers() ([]*UserInfo, error) {
	rows, err := ut.db.Query(queries.GET_ALL_USERS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*UserInfo{}
	for rows.Next() {
		var user UserInfo
		err := rows.Scan(&user.ID, &user.LoginName, &user.AvatarURL, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
