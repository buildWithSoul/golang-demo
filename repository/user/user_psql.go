package userRepository

import (
	"database/sql"

	"../../models"
)

// UserRepository is a skeleton struct
type UserRepository struct{}

func (u UserRepository) Signup(db *sql.DB, user models.User) (models.User, error) {
	stmt := "insert into users (email, password) values($1, $2) RETURNING id;"
	err := db.QueryRow(stmt, user.Email, user.Password).Scan(&user.ID)
	return user, err
}

func (u UserRepository) Login(db *sql.DB, user models.User) (models.User, error) {
	row := db.QueryRow("select * from users where email=$1", user.Email)
	err := row.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		return user, err
	}

	return user, nil
}
