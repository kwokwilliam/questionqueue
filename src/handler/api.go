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

	switch mux.Vars(r)["id"] {
	// get current user profile
	case "me":
		currentState := &session.State{}
		_, err := session.GetState(r, ctx.Key, ctx.SessionStore, currentState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		b, _ := json.Marshal(currentState)

		httpWriter(http.StatusOK, b, "application/json", w)

	// get all teachers, they are authorized to do so
	case "all":
		// current state discarded
		_, err := session.GetState(r, ctx.Key, ctx.SessionStore, &session.State{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		teachers, err := ctx.MongoStore.GetAllTeacher()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, _ := json.Marshal(teachers)
		httpWriter(http.StatusOK, b, "application/json", w)
	}
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
// {queue : [id1, id2, ..., idN] }
func (ctx *Context) PostQuestionHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, ErrUnsupportedMediaType.Error(), http.StatusUnsupportedMediaType)
		return
	}

	nq, err := decodeQuestion(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if err := updateQueue(ctx, nq, notifier.QuestionNew); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := ctx.MongoStore.InsertQuestion(nq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpWriter(http.StatusCreated, nil, "", w)
	return
}

// UpdateQueue commits an update to Redis and MessageQueue
func updateQueue(ctx *Context, nq *model.Question, messageType string) error {

	id := nq.BelongsTo

	// get current queue from redis
	currentState := &session.State{}
	err := ctx.SessionStore.Get("queue", currentState)
	if err != nil {
		// `redis: nil` == empty redis, ignore
		if err.Error() != "redis: nil" {
			return err
		}
	}

	// update current queue
	currentID, err := currentQueueMarshaler(currentState.Interface)
	if err != nil {
		return err
	}

	currentID = append(currentID, id)
	newState := session.State{
		SessionStart: time.Now(),
		Interface:    currentID,
	}

	// update redis
	if err := ctx.SessionStore.SetQueue("queue", newState); err != nil {
		return err
	}

	// create message and push to mq
	ctx.Notifier.PublishMessage(&notifier.Message{
		Type:    messageType,
		Content: nq,
		UserID:  id,
	})

	return nil
}

// CurrentQueueMarshaler takes the content of the current queue and marshals into a string slice for manipulation
// Returns any error found
func currentQueueMarshaler(i interface{}) ([]string, error) {
	marshaledCurrentState, err := json.Marshal(i)
	var currentID []string

	if string(marshaledCurrentState) == "null" || len(marshaledCurrentState) == 0 {
		return []string{}, nil
	}

	if err = json.Unmarshal(marshaledCurrentState, &currentID); err != nil {
		return nil, err
	} else {
		return currentID, nil
	}
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