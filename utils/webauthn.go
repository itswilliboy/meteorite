package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/go-webauthn/webauthn/webauthn"
)

var WebAuthnAPI *webauthn.WebAuthn

func InitWebAuthn() error {
	rpID := os.Getenv("WEBAUTHN_RP_ID")
	rpOrigin := os.Getenv("WEBAUTHN_RP_ORIGIN")

	if rpID == "" || rpOrigin == "" {
		log.Println("WEBAUTHN_RP_ID/WEBAUTHN_RP_ORIGIN not set, passkey login is disabled")
		return nil
	}

	config := &webauthn.Config{
		RPDisplayName: "Meteorite",
		RPID:          rpID,
		RPOrigins:     strings.Split(rpOrigin, ","),
	}

	w, err := webauthn.New(config)
	if err != nil {
		return err
	}

	WebAuthnAPI = w
	return nil
}

type webauthnUser struct {
	user        User
	credentials []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte                         { return []byte(strconv.Itoa(u.user.ID)) }
func (u *webauthnUser) WebAuthnName() string                       { return u.user.Name }
func (u *webauthnUser) WebAuthnDisplayName() string                { return u.user.Name }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func LoadWebAuthnUser(userID int) (*webauthnUser, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := DB.Query(context.Background(), "SELECT data FROM webauthn_credentials WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	credentials := []webauthn.Credential{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}

		var cred webauthn.Credential
		if err := json.Unmarshal(raw, &cred); err != nil {
			return nil, err
		}

		credentials = append(credentials, cred)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &webauthnUser{user: user, credentials: credentials}, nil
}

func LoadWebAuthnUserByHandle(rawID, userHandle []byte) (webauthn.User, error) {
	userID, err := strconv.Atoi(string(userHandle))
	if err != nil {
		return nil, err
	}

	return LoadWebAuthnUser(userID)
}

func ParseWebAuthnUserID(user webauthn.User) (int, error) {
	return strconv.Atoi(string(user.WebAuthnID()))
}

func SaveCredential(userID int, name string, cred *webauthn.Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	_, err = DB.Exec(
		context.Background(),
		`
			INSERT INTO webauthn_credentials (id, user_id, name, data)
				VALUES ($1, $2, $3, $4)
			ON CONFLICT (id)
				DO UPDATE SET data = EXCLUDED.data
		`,
		cred.ID, userID, name, data,
	)
	return err
}

func UpdateCredentialSignCount(cred *webauthn.Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	_, err = DB.Exec(context.Background(), "UPDATE webauthn_credentials SET data = $1 WHERE id = $2", data, cred.ID)
	return err
}

type CredentialInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func ListCredentials(userID int) ([]CredentialInfo, error) {
	rows, err := DB.Query(
		context.Background(),
		"SELECT id, name, created_at FROM webauthn_credentials WHERE user_id = $1 ORDER BY created_at",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []CredentialInfo{}
	for rows.Next() {
		var id []byte
		var info CredentialInfo
		if err := rows.Scan(&id, &info.Name, &info.CreatedAt); err != nil {
			return nil, err
		}
		info.ID = base64.RawURLEncoding.EncodeToString(id)
		list = append(list, info)
	}

	return list, rows.Err()
}

func DeleteCredential(userID int, credentialID []byte) error {
	_, err := DB.Exec(context.Background(), "DELETE FROM webauthn_credentials WHERE id = $1 AND user_id = $2", credentialID, userID)
	return err
}

const webauthnChallengeCookieName = "webauthn-challenge"

func CreateChallengeCookie(session *webauthn.SessionData) (http.Cookie, error) {
	data, err := json.Marshal(session)
	if err != nil {
		return http.Cookie{}, err
	}

	expiry := time.Now().Add(5 * time.Minute)

	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(expiry)
	token.Set("challenge", string(data))

	return http.Cookie{
		Name:     webauthnChallengeCookieName,
		Value:    token.V4Encrypt(sessionKey, nil),
		Expires:  expiry,
		Path:     "/api/webauthn",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}, nil
}

func ReadChallengeCookie(r *http.Request) (*webauthn.SessionData, error) {
	cookie, err := r.Cookie(webauthnChallengeCookieName)
	if err != nil {
		return nil, err
	}

	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(sessionKey, cookie.Value, nil)
	if err != nil {
		return nil, err
	}

	raw, err := token.GetString("challenge")
	if err != nil {
		return nil, err
	}

	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func ClearChallengeCookie() http.Cookie {
	return http.Cookie{
		Name:     webauthnChallengeCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/api/webauthn",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
