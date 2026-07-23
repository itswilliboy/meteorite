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

func WebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
	if webAuthnUnavailable(w) {
		return
	}

	user, err := utils.LoadWebAuthnUser(utils.GetUserID(r))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	options, session, err := utils.WebAuthnAPI.BeginRegistration(
		user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()),
	)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	cookie, err := utils.CreateChallengeCookie(session)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	http.SetCookie(w, &cookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: options.Response})
}

func WebAuthnRegisterFinish(w http.ResponseWriter, r *http.Request) {
	if webAuthnUnavailable(w) {
		return
	}

	userID := utils.GetUserID(r)

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Passkey"
	}

	session, err := utils.ReadChallengeCookie(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Passkey registration expired, please try again.")
		return
	}

	user, err := utils.LoadWebAuthnUser(userID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	credential, err := utils.WebAuthnAPI.FinishRegistration(user, *session, r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Could not verify passkey.")
		return
	}

	if err := utils.SaveCredential(userID, name, credential); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	clearCookie := utils.ClearChallengeCookie()
	http.SetCookie(w, &clearCookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
}

func WebAuthnLoginBegin(w http.ResponseWriter, r *http.Request) {
	if webAuthnUnavailable(w) {
		return
	}

	options, session, err := utils.WebAuthnAPI.BeginDiscoverableLogin()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	cookie, err := utils.CreateChallengeCookie(session)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	http.SetCookie(w, &cookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: options.Response})
}

func WebAuthnLoginFinish(w http.ResponseWriter, r *http.Request) {
	if webAuthnUnavailable(w) {
		return
	}

	session, err := utils.ReadChallengeCookie(r)
	if err != nil {
		utils.WriteCodeError(w, http.StatusUnauthorized)
		return
	}

	authenticatedUser, credential, err := utils.WebAuthnAPI.FinishPasskeyLogin(utils.LoadWebAuthnUserByHandle, *session, r)
	if err != nil {
		utils.WriteCodeError(w, http.StatusUnauthorized)
		return
	}

	if err := utils.UpdateCredentialSignCount(credential); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	userID, err := utils.ParseWebAuthnUserID(authenticatedUser)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	user, err := utils.GetUserByID(userID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	clearCookie := utils.ClearChallengeCookie()
	http.SetCookie(w, &clearCookie)

	sessionCookie := utils.CreateDashSessionCookie(user)
	http.SetCookie(w, &sessionCookie)

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: user})
}

func WebAuthnListCredentials(w http.ResponseWriter, r *http.Request) {
	list, err := utils.ListCredentials(utils.GetUserID(r))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: list})
}

func WebAuthnDeleteCredential(w http.ResponseWriter, r *http.Request) {
	id, err := base64.RawURLEncoding.DecodeString(r.PathValue("id"))
	if err != nil {
		utils.WriteCodeError(w, http.StatusBadRequest)
		return
	}

	if err := utils.DeleteCredential(utils.GetUserID(r), id); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
}
