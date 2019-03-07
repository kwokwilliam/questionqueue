package handlers

import (
	"assignments-kwokwilliam/servers/gateway/models/users"
	"assignments-kwokwilliam/servers/gateway/sessions"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

const headerAccessControlAllowOrigin = "Access-Control-Allow-Origin"
const allOrigins = "*"

const headerContentType = "Content-Type"
const contentTypeJSON = "application/json"

// writeUserToResponse will write the given user and status code to
// the http responsewriter as JSON
func writeUserToResponse(w http.ResponseWriter, statusCode int, user *users.User) error {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	err := enc.Encode(user)
	return err
}

// UsersHandler handles requests for the "users" resource.
func (ctx *HandlerContext) UsersHandler(w http.ResponseWriter, r *http.Request) {
	// If the method is POST
	if r.Method == "POST" {
		// Make sure content type is JSON
		reqContentTypeHeader := r.Header.Get("Content-Type")
		if !strings.HasPrefix(reqContentTypeHeader, contentTypeJSON) {
			http.Error(w, "Unsupported Media Type - Request body must be in JSON", http.StatusUnsupportedMediaType)
			return
		}

		// decode json in body
		newUser := &users.NewUser{}
		jsonDecoder := json.NewDecoder(r.Body)
		if err := jsonDecoder.Decode(newUser); err != nil {
			http.Error(w, "Internal server error - Unable to decode JSON", http.StatusBadRequest)
			return
		}

		// Create new user (validates data as well)
		newUserAsUser, err := newUser.ToUser()
		if err != nil || newUserAsUser == nil {
			http.Error(w, "Internal server error - Unable to create new user", http.StatusBadRequest)
			return
		}

		// Create new user account in database
		newUserAsUser, err = ctx.userStore.Insert(newUserAsUser)
		if err != nil || newUserAsUser == nil {
			print(err)
			fmt.Print(newUserAsUser)
			http.Error(w, "User already exists", http.StatusBadRequest)
			return
		}

		// Add user to trie
		ctx.trie.AddUserToTrie(newUserAsUser.UserName, newUserAsUser.FirstName, newUserAsUser.LastName, newUserAsUser.ID)

		// Create a new session for the user
		sessionState := NewSessionState(time.Now(), newUserAsUser)
		sessionID, err := sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, w)
		ctx.SetSessionID(sessionID)
		if err != nil {
			http.Error(w, "Internal server error - Unable to begin new session", http.StatusInternalServerError)
			return
		}

		err = writeUserToResponse(w, http.StatusCreated, newUserAsUser)
		if err != nil {
			http.Error(w, "Internal server error - Unable to encode", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == "GET" {
		sessionState := &SessionState{}

		// Getting state should validate session id
		_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
		if err != nil {
			http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
			return
		}

		prefix := r.URL.Query().Get("q")
		if prefix == "" {
			http.Error(w, "q parameter cannot be left out", http.StatusBadRequest)
			return
		}

		uids := ctx.trie.Find(strings.ToLower(prefix), 20)
		usersArr, err := ctx.userStore.GetMultipleUsersByID(uids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set(headerContentType, contentTypeJSON)
		enc := json.NewEncoder(w)
		if err = enc.Encode(usersArr); err != nil {
			http.Error(w, "Unable to parse users", http.StatusInternalServerError)
			return
		}
		return
	}
	http.Error(w, "Method not allowed - must be a POST request", http.StatusMethodNotAllowed)
	return
}

// SpecificUsersHandler handles requests for a specific user.
// Because a lot of these are error handled functions, the best way to
// go about it would be to not split it apart. This lends itself to be
// more readable as well.
func (ctx *HandlerContext) SpecificUsersHandler(w http.ResponseWriter, r *http.Request) {
	sessionState := &SessionState{}

	// Getting state should validate session id
	sid, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
	if err != nil {
		http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
		return
	}

	// Check if a valid path has been given
	reqIDStr := mux.Vars(r)["uid"]
	var reqIDNum int64
	if reqIDStr == "me" {
		reqIDNum = sessionState.User.ID
	} else {
		reqIDNum, err = strconv.ParseInt(reqIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID provided", http.StatusBadRequest)
			return
		}
	}

	// If the method is GET, get user profile associated with requested ID
	// If it cannot be found, return an error.
	if r.Method == "GET" {
		user, err := ctx.userStore.GetByID(reqIDNum)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		err = writeUserToResponse(w, http.StatusOK, user)
		// print(w.Header().Get(headerContentType))
		if err != nil {
			http.Error(w, "Internal server error - Unable to encode", http.StatusInternalServerError)
			return
		}
		return
	}

	// If the method is patch, update the user with given data
	if r.Method == "PATCH" {
		if reqIDNum != sessionState.User.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Make sure content type is JSON
		reqContentTypeHeader := r.Header.Get("Content-Type")
		if !strings.HasPrefix(reqContentTypeHeader, contentTypeJSON) {
			http.Error(w, "Unsupported Media Type - Request body must be in JSON", http.StatusUnsupportedMediaType)
			return
		}

		// decode json in body
		userUpdates := &users.Updates{}
		jsonDecoder := json.NewDecoder(r.Body)
		if err := jsonDecoder.Decode(userUpdates); err != nil {
			http.Error(w, "Internal server error - Unable to decode JSON", http.StatusInternalServerError)
			return
		}

		// Keep track of old firstname and lastname
		oldFirstName := sessionState.User.FirstName
		oldLastName := sessionState.User.LastName

		// update user profile
		updatedUser, err := ctx.userStore.Update(sessionState.User.ID, userUpdates)
		if err != nil {
			http.Error(w, "Internal server error - Unable to update user", http.StatusInternalServerError)
			return
		}

		// update user session
		sessionState.User = *updatedUser
		if err = ctx.SessionStore.Save(sid, sessionState); err != nil {
			http.Error(w, "Error saving new session", http.StatusInternalServerError)
			return
		}

		// update user in trie
		ctx.trie.RemoveNamesInTrie(oldFirstName, oldLastName, sessionState.User.ID)
		ctx.trie.AddUserToTrie(updatedUser.UserName, updatedUser.FirstName, updatedUser.LastName, sessionState.User.ID)

		// write user to response body
		err = writeUserToResponse(w, http.StatusCreated, updatedUser)
		if err != nil {
			http.Error(w, "Internal server error - Unable to encode", http.StatusInternalServerError)
			return
		}
		return
	}

	// If not GET or PATCH, return an error
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}

// SessionsHandler handles the request for the sessions resource and allows clients to
// begin a new session using an existing user's credentials
func (ctx *HandlerContext) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed - must be a POST request", http.StatusMethodNotAllowed)
		return
	}

	// Make sure content type is JSON
	reqContentTypeHeader := r.Header.Get("Content-Type")
	if !strings.HasPrefix(reqContentTypeHeader, contentTypeJSON) {
		http.Error(w, "Unsupported Media Type - Request body must be in JSON", http.StatusUnsupportedMediaType)
		return
	}

	// decode json in body
	creds := &users.Credentials{}
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(creds); err != nil {
		http.Error(w, "Internal server error - Unable to decode JSON", http.StatusInternalServerError)
		return
	}

	// try find user profile
	user, err := ctx.userStore.GetByEmail(creds.Email)
	if err != nil {
		bcrypt.CompareHashAndPassword([]byte("dummy"), []byte("dummy"))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set authentication tries
	numTries, err := ctx.SessionStore.SetTries(creds.Email)
	if err != nil || numTries >= 6 {
		http.Error(w, "Blocked repeated failed sign ins", http.StatusUnauthorized)
		return
	}

	// try authenticate user
	if err = user.Authenticate(creds.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// if auth success, begin new session
	sessionState := NewSessionState(time.Now(), user)
	sessionID, err := sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, w)
	ctx.SetSessionID(sessionID)
	if err != nil {
		http.Error(w, "Internal server error - Unable to begin new session", http.StatusInternalServerError)
		return
	}

	// Track sign in
	ip := r.RemoteAddr
	if r.Header.Get("X-Forwarded-For") != "" {
		ips := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		ip = ips[0]
	}
	if err = ctx.userStore.InsertSignIn(user.ID, ip); err != nil {
		fmt.Printf("FAILED TO TRACK USER [%v]", user.ID)
	}

	// respond to client with status code of http.StatusCreated
	err = writeUserToResponse(w, http.StatusCreated, user)
	if err != nil {
		http.Error(w, "Internal server error - Unable to encode", http.StatusInternalServerError)
		return
	}
}

// SpecificSessionHandler handles request related to a specific authenticated session
// Currently supported operations:
// 		DELETE - Ends the user's session
func (ctx *HandlerContext) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check if valid path given (must be "mine")
	reqString := mux.Vars(r)["uid"]
	if reqString != "mine" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// End the current session otherwise
	if err := ctx.DeleteCurrentSession(); err != nil {
		http.Error(w, "Internal server error - Session delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("signed out"))
	return
}

// ProfilePhotoHandler handles the requests related to
func (ctx *HandlerContext) ProfilePhotoHandler(w http.ResponseWriter, r *http.Request) {
	sessionState := &SessionState{}

	// Getting state should validate session id
	_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
	if err != nil {
		http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
		return
	}
	var reqIDNum int64
	reqIDStr := mux.Vars(r)["uid"]
	if reqIDStr == "me" {
		reqIDNum = sessionState.User.ID
		reqIDStr = strconv.FormatInt(sessionState.User.ID, 10)
	} else {
		// Throw error if invalid ID provided
		reqIDNum, err = strconv.ParseInt(reqIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID provided", http.StatusBadRequest)
			return
		}
	}
	fileName := "./avatars/" + reqIDStr + ".png"
	if r.Method == "GET" {
		if _, err := os.Stat(fileName); err == nil {
			http.ServeFile(w, r, fileName)
			return
		}
		http.Error(w, "Cannot find avatar", http.StatusNotFound)
		return
	}

	if r.Method == "PUT" {
		// Check that provided id is user
		if reqIDNum != sessionState.User.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Get file
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			http.Error(w, "Unable to parse file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Create avatar folder if it doesn't exist
		avatarFolder := filepath.Join(".", "avatars")
		if err = os.MkdirAll(avatarFolder, os.ModePerm); err != nil {
			http.Error(w, "Unable to save file internally", http.StatusInternalServerError)
			return
		}

		f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		// Set user's new Photo URL
		updates := &users.Updates{PhotoURL: "https://api.uwinfotutor.me/v1/users/" + reqIDStr + "/avatar"}
		if err = sessionState.User.ApplyUpdates(updates); err != nil {
			http.Error(w, "Unable to apply user updates", http.StatusInternalServerError)
			return
		}
		if _, err := ctx.userStore.UpdateImage(reqIDNum, updates); err != nil {
			fmt.Print(err)
			http.Error(w, "Failure to save new avatar", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Avatar updated"))
		return
	}

	// If not GET or PUT, return an error
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}

// ResetCodesHandler is the handler for resetting codes
func (ctx *HandlerContext) ResetCodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// decode json in body
	creds := &users.Credentials{}
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(creds); err != nil {
		http.Error(w, "Internal server error - Unable to decode JSON", http.StatusInternalServerError)
		return
	}

	// set reset code
	resetCode, err := ctx.userStore.SetResetCode(creds.Email)
	if err != nil {
		http.Error(w, "Unable to reset code. Email may not exist.", http.StatusNotFound)
		return
	}

	// Send email, currently uses my email until I figure out how to send via the server.
	auth := smtp.PlainAuth("", ctx.accessKey, ctx.secretKey, "email-smtp.us-west-2.amazonaws.com")
	to := []string{creds.Email}
	msg := []byte("To: " + creds.Email + "\r\n" +
		"Subject: Password Reset Confirmation Code\r\n" +
		"\r\n" +
		"This is your reset code: " + resetCode + "\r\n")
	err = smtp.SendMail("email-smtp.us-west-2.amazonaws.com:25", auth, "mockingod@gmail.com", to, msg)
	if err != nil {
		http.Error(w, "Unable to generate reset code", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Confirmation code sent to email"))
	return
}

// ResetPassword handles the password reset
func (ctx *HandlerContext) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Grab email and check database for it.
	email := mux.Vars(r)["email"]
	resetCode, resetTime, err := ctx.userStore.GetResetCodeByEmail(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// decode json in body
	resetData := &users.UserResetPass{}
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(resetData); err != nil {
		http.Error(w, "Internal server error - Unable to decode JSON", http.StatusInternalServerError)
		return
	}

	// check matching reset codes, time is valid, and passwords match
	if resetCode != resetData.ResetCode {
		http.Error(w, "Invalid reset code", http.StatusBadRequest)
		return
	}
	if time.Since(resetTime).Minutes() > 5 {
		http.Error(w, "Reset code expired", http.StatusBadRequest)
		return
	}
	if resetData.Password != resetData.PasswordConf {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Create new password hash (a lot easier to do this than call the object itself)
	passHash, err := bcrypt.GenerateFromPassword([]byte(resetData.Password), bcrypt.MinCost)
	if err != nil {
		http.Error(w, "Password could not be saved", http.StatusInternalServerError)
		return
	}

	if err = ctx.userStore.UpdatePassword(email, passHash); err != nil {
		http.Error(w, "Error saving password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Password updated"))
	return
}
