package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Enabled   bool      `json:"enabled"`
	Admin     bool      `json:"admin"`
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

func (user *User) SetPassword(password string) error {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec(context.Background(), "UPDATE users SET password = $1 where id = $2", hashedPassword, user.ID)
	if err != nil {
		return err
	}

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
		return "", err
	}

	return string(token), nil
}

// this works but this would also mean that on restart the secret key changes
// which would "invalidate" all the previous session prior to the restart.
// this could also make it a fresh start :D
var sk = paseto.NewV4AsymmetricSecretKey()
var pk = sk.Public()

func CreateSessionToken(user User, expiry time.Time) string {
	token := paseto.NewToken()

	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(expiry)

	token.Set("user_id", fmt.Sprint(user.ID))

	signed := token.V4Sign(sk, nil)

	return signed
}

const DASH_COOKIE_NAME = "meteorite-cookie"

func CreateDashSessionCookie(user User) http.Cookie {
	// 1 day login validty
	expiry := time.Now().Add(1 * 24 * time.Hour)
	token := CreateSessionToken(user, expiry)

	return http.Cookie{
		Name:    DASH_COOKIE_NAME,
		Value:   token,
		Expires: expiry,
		Path:    "/api",
	}
}

// this is for cookie invalidation on the browser.
func CreateInvalidDashSessionCookie() http.Cookie {
	return http.Cookie{
		Name:    DASH_COOKIE_NAME,
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/api",
	}
}

func VerifySessionToken(signed string) (string, error) {
	parser := paseto.NewParser()
	token, err := parser.ParseV4Public(pk, signed, nil)
	if err != nil {
		log.Println("could not parse token", err)
		return "", err
	}

	userID, err := token.GetString("user_id")
	if err != nil {
		log.Println("could not set user id", err)
		return "", err
	}

	return userID, nil
}

func CreateUser(name string, password string) (User, error) {
	ctx := context.Background()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	row := DB.QueryRow(ctx, "INSERT INTO users (name, password) VALUES ($1, $2) RETURNING id, name, created_at, enabled, admin", name, hashedPassword)

	user, err := scanUserRow(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return User{}, ErrUsernameAlreadyExists
		}

		return User{}, err
	}

	// default token
	// token, _ := CreateToken(user.ID)
	// DB.Exec(ctx, "INSERT INTO tokens VALUES ($1, $2)", user.ID, []byte(token))

	return user, nil
}

func LoginUser(name string, password string) (User, error) {
	var id int
	var pass []byte

	err := DB.QueryRow(context.Background(), "SELECT id, password FROM users WHERE name = $1 AND enabled = true", name).Scan(&id, &pass)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword(pass, []byte(password))
	if err != nil {
		return User{}, err
	}

	return GetUserByID(id)
}
