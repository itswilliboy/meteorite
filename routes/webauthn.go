package routes

import (
	"encoding/base64"
	"img/utils"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func webAuthnUnavailable(w http.ResponseWriter) bool {
	if utils.WebAuthnAPI == nil {
		utils.WriteJSONError(w, http.StatusServiceUnavailable, "Passkeys are not configured on this server.")
		return true
	}
	return false
}

func WebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) error {
	if webAuthnUnavailable(w) {
		return nil
	}

	user, err := utils.LoadWebAuthnUser(utils.GetUserID(r))
	if err != nil {
		return err
	}

	options, session, err := utils.WebAuthnAPI.BeginRegistration(
		user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()),
	)
	if err != nil {
		return err
	}

	cookie, err := utils.CreateChallengeCookie(session)
	if err != nil {
		return err
	}
	http.SetCookie(w, &cookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: options.Response})
	return nil
}

func WebAuthnRegisterFinish(w http.ResponseWriter, r *http.Request) error {
	if webAuthnUnavailable(w) {
		return nil
	}

	userID := utils.GetUserID(r)

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Passkey"
	}

	session, err := utils.ReadChallengeCookie(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Passkey registration expired, please try again.")
		return nil
	}

	user, err := utils.LoadWebAuthnUser(userID)
	if err != nil {
		return err
	}

	credential, err := utils.WebAuthnAPI.FinishRegistration(user, *session, r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Could not verify passkey.")
		return nil
	}

	if err := utils.SaveCredential(userID, name, credential); err != nil {
		return err
	}

	clearCookie := utils.ClearChallengeCookie()
	http.SetCookie(w, &clearCookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
	return nil
}

func WebAuthnLoginBegin(w http.ResponseWriter, r *http.Request) error {
	if webAuthnUnavailable(w) {
		return nil
	}

	options, session, err := utils.WebAuthnAPI.BeginDiscoverableLogin()
	if err != nil {
		return err
	}

	cookie, err := utils.CreateChallengeCookie(session)
	if err != nil {
		return err
	}
	http.SetCookie(w, &cookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: options.Response})
	return nil
}

func WebAuthnLoginFinish(w http.ResponseWriter, r *http.Request) error {
	if webAuthnUnavailable(w) {
		return nil
	}

	session, err := utils.ReadChallengeCookie(r)
	if err != nil {
		utils.WriteCodeError(w, http.StatusUnauthorized)
		return nil
	}

	authenticatedUser, credential, err := utils.WebAuthnAPI.FinishPasskeyLogin(utils.LoadWebAuthnUserByHandle, *session, r)
	if err != nil {
		utils.WriteCodeError(w, http.StatusUnauthorized)
		return nil
	}

	if err := utils.UpdateCredentialSignCount(credential); err != nil {
		return err
	}

	userID, err := utils.ParseWebAuthnUserID(authenticatedUser)
	if err != nil {
		return err
	}

	user, err := utils.GetUserByID(userID)
	if err != nil {
		return err
	}

	clearCookie := utils.ClearChallengeCookie()
	http.SetCookie(w, &clearCookie)

	sessionCookie := utils.CreateDashSessionCookie(user)
	http.SetCookie(w, &sessionCookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: user})
	return nil
}

func WebAuthnListCredentials(w http.ResponseWriter, r *http.Request) error {
	list, err := utils.ListCredentials(utils.GetUserID(r))
	if err != nil {
		return err
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: list})
	return nil
}

func WebAuthnDeleteCredential(w http.ResponseWriter, r *http.Request) error {
	id, err := base64.RawURLEncoding.DecodeString(r.PathValue("id"))
	if err != nil {
		utils.WriteCodeError(w, http.StatusBadRequest)
		return nil
	}

	if err := utils.DeleteCredential(utils.GetUserID(r), id); err != nil {
		return err
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
	return nil
}
