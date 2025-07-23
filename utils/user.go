package utils

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Name      string
	CreatedAt time.Time
	Enabled   bool
	Admin     bool
}

func (user *User) SetEnabled(enabled bool) error {
	_, err := DB.Exec(context.Background(), "UPDATE users SET enabled = $1 WHERE id = $2", enabled, user.ID)
	if err != nil {
		return err
	}
	user.Enabled = enabled
	return nil
}

func (user *User) SetAdmin(enabled bool) error {
	_, err := DB.Exec(context.Background(), "UPDATE users SET admin = $1 WHERE id = $2", enabled, user.ID)
	if err != nil {
		return err
	}

	user.Admin = enabled
	return nil
}

func scanUserRow(row pgx.Row) (User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.CreatedAt, &u.Enabled, &u.Admin); err != nil {
		return User{}, err
	}

	return u, nil
}

func GetUserByID(id int) (User, error) {
	row := DB.QueryRow(context.Background(), "SELECT id, name, created_at, enabled, admin FROM users WHERE id = $1", id)

	user, err := scanUserRow(row)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetUserByToken(token string) (User, error) {
	row := DB.QueryRow(
		context.Background(),
		`SELECT id, name, created_at, enabled, admin FROM users u
		JOIN tokens t ON u.id = t.user_id
		WHERE t.token = $1`,
		token,
	)

	user, err := scanUserRow(row)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetToken(user User) (string, error) {
	var token []byte
	err := DB.QueryRow(
		context.Background(),
		`SELECT token FROM tokens WHERE user_id = $1`, user.ID,
	).Scan(&token)

	if err != nil {
		log.Println("error getting token", err)
		return "", err
	}

	return string(token), nil
}

func CreateUser(name string, password string) (string, error) {
	ctx := context.Background()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("error hashing password")
		return "", err
	}

	row := DB.QueryRow(ctx, "INSERT INTO users (name, password) VALUES ($1, $2) RETURNING id, name, created_at, enabled, admin", name, hashedPassword)

	user, err := scanUserRow(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", ErrUsernameAlreadyExists
		}

		log.Println("error scanning user")
		return "", err
	}

	token, _ := CreateToken(user.ID, user.CreatedAt)
	DB.Exec(ctx, "INSERT INTO tokens VALUES ($1, $2)", user.ID, []byte(token))

	return token, nil
}

func LoginUser(name string, password string) (User, error) {
	var id int
	var pass []byte

	err := DB.QueryRow(context.Background(), "SELECT id, password FROM users WHERE name = $1", name).Scan(&id, &pass)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword(pass, []byte(password))
	if err != nil {
		return User{}, err
	}

	return GetUserByID(id)
}
