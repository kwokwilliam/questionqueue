package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"questionqueue/src/db"
	"questionqueue/src/model"
	"questionqueue/src/notifier"
	"questionqueue/src/session"
	"strings"
	"time"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrMethodNotAllowed     = errors.New("method not allowed")
)

func (ctx *Context) OkHandler(w http.ResponseWriter, r *http.Request) {
	httpWriter(http.StatusOK, []byte("connected"), "text/plain", w)
	return
}

// TA/teacher control
func (ctx *Context) TeacherHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// Create new TA/teacher.
	case http.MethodPost:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, ErrUnsupportedMediaType.Error(), http.StatusUnsupportedMediaType)
			return
		}

		// decode, verify, save to mongo, save to redis, write results back to client
		nt, err := decodeNewTeacher(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		if err := nt.VerifyNewTeacher(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := ctx.MongoStore.InsertTeacher(nt)
		if err == db.ErrEmailUsed {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t := model.Teacher{
			ID:        res.InsertedID.(primitive.ObjectID),
			Email:     nt.Email,
			FirstName: nt.FirstName,
			LastName:  nt.LastName,
		}

		newSessionState := session.State{
			SessionStart: time.Now(),
			Interface:    t,
		}

		_, err = session.BeginSession(ctx.Key, ctx.SessionStore, newSessionState, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, _ := json.Marshal(t)
		httpWriter(http.StatusCreated, b, "application/json", w)
		return

	//  Update information for a TA/teacher.
	case http.MethodPatch:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, ErrUnsupportedMediaType.Error(), http.StatusUnsupportedMediaType)
			return
		}

		// get current state and session ID
		currentState := &session.State{}
		_, err := session.GetState(r, ctx.Key, ctx.SessionStore, currentState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		sid, err := session.GetSessionID(r, ctx.Key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// decode, verify, save to mongo, save to redis, write results back to client
		tu, err := decodeTeacherUpdate(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		// verify model
		if err := tu.VerifyTeacherUpdate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// check password
		if _, err = ctx.authenticate(tu.Email, tu.OldPassword); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if _, err := ctx.MongoStore.UpdateTeacher(tu); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// retrieve updated record
		// could use id
		currentTeacher, err := ctx.MongoStore.GetTeacherByEmail(tu.Email)
		if len(currentTeacher) > 1 {
			log.Printf("email %v got more than 1 result", tu.Email)
			http.Error(w, "got more than one profile", http.StatusInternalServerError)
		}

		newState := session.State{
			SessionStart: time.Now(),
			Interface:    currentTeacher[0],
		}

		if err := ctx.SessionStore.Save(sid, newState); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// `res.UpsertedID` should match whatever the original session ID is, *double check* redis
		js, err := json.Marshal(currentTeacher[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		httpWriter(http.StatusCreated, js, "application/json", w)
		return

	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
}

func (ctx *Context) TeacherProfileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	id := mux.Vars(r)["id"]
	if id != "me" {
		http.Error(w, "you can only get your own profile", http.StatusForbidden)
	}

	currentState := &session.State{}
	_, err := session.GetState(r, ctx.Key, ctx.SessionStore, currentState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	b, _ := json.Marshal(currentState)

	httpWriter(http.StatusOK, b, "application/json", w)
}

// TA/teacher session control
func (ctx *Context) TeacherSessionHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// login
	case http.MethodPost:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, ErrUnsupportedMediaType.Error(), http.StatusUnsupportedMediaType)
			return
		}

		tl, err := decodeTeacherLogin(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		t, err := ctx.authenticate(tl.Email, tl.Password)
		if err != nil {
			http.Error(w, ErrInvalidCredentials.Error(), http.StatusForbidden)
			return
		}

		newSessionState := session.State{
			SessionStart: time.Now(),
			Interface:    t,
		}
		_, err = session.BeginSession(ctx.Key, ctx.SessionStore, newSessionState, w)

		js, _ := json.Marshal(t)
		httpWriter(http.StatusOK, js, "application/json", w)

	// delete session
	case http.MethodDelete:

		var err error

		// `State` is discarded
		_, err = session.GetState(r, ctx.Key, ctx.SessionStore, &session.State{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		_, err = session.EndSession(r, ctx.Key, ctx.SessionStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		httpWriter(http.StatusOK, []byte("you have been signed out"), "application/json", w)

	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

}

// PostQuestionHandler posts new question to mongo and enqueues the question to redis in the format of
// {queue : [question1, question2, ..., questionN] }
// TODO: redis operations
func (ctx *Context) PostQuestionHandler(w http.ResponseWriter, r *http.Request) {

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, ErrUnsupportedMediaType.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	nq, err := decodeQuestion(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if _, err := ctx.MongoStore.InsertQuestion(nq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := updateQueue(r, ctx, nq, notifier.QuestionNew); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpWriter(http.StatusCreated, nil, "", w)
	return
}

// HttpWriter takes necessary arguments to write back to client.
func httpWriter(statusCode int, body []byte, contentType string, w http.ResponseWriter) {
	if len(contentType) > 0 {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	w.WriteHeader(statusCode)

	if len(body) > 0 {
		_, _ = w.Write(body)
	}
}

// UpdateQueue commits an update to Redis and MessageQueue
func updateQueue(r *http.Request, ctx *Context, i interface{}, messageType string) error {
	// save to redis
	sid, err := session.GetSessionID(r, ctx.Key)
	if err != nil {
		return err
	}

	if err := ctx.SessionStore.Save(sid, i); err != nil {
		return err
	}

	// create message and push to mq
	ctx.Notifier.PublishMessage(&notifier.Message{
		Type:    messageType,
		Content: i,
		UserID:  sid,
	})

	return nil
}

// Authenticate searches existing teachers in the mongo,
// then authenticated against the provided password,
// finally returns the pointer of the matched user.
func (ctx *Context) authenticate(email, password string) (*model.Teacher, error) {
	teachers, err := ctx.MongoStore.GetTeacherByEmail(email)
	if err != nil {
		return nil, err
	}

	if len(teachers) == 0 {
		return nil, ErrInvalidCredentials
	}

	if len(teachers) > 1 {
		log.Printf("got more than 1 matched users.\nemail: %v", email)
		return nil, ErrInvalidCredentials
	}

	// check index 0 as there should only be one matched result
	if err := teachers[0].Authenticate(password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// hide password_hash
	teachers[0].PasswordHash = ""

	return teachers[0], nil
}