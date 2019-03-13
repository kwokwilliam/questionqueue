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

const (
	MimeJson  = "application/json"
	MimePlain = "text/plain"
)

func (ctx *Context) OkHandler(w http.ResponseWriter, r *http.Request) {
	httpWriter(http.StatusOK, []byte("connected"), MimePlain, w)
	return
}

// TA/teacher control
func (ctx *Context) TeacherHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// Create new TA/teacher.
	case http.MethodPost:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), MimeJson) {
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
		httpWriter(http.StatusCreated, b, MimeJson, w)
		return

	//  Update information for a TA/teacher.
	case http.MethodPatch:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), MimeJson) {
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

		httpWriter(http.StatusCreated, js, MimeJson, w)
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

		httpWriter(http.StatusOK, b, MimeJson, w)

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
		httpWriter(http.StatusOK, b, MimeJson, w)
	}
}

// TA/teacher session control
func (ctx *Context) TeacherSessionHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// login
	case http.MethodPost:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), MimeJson) {
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
		httpWriter(http.StatusOK, js, MimeJson, w)

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

		httpWriter(http.StatusOK, []byte("you have been signed out"), MimeJson, w)

	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

}

// PostQuestionHandler posts new question to mongo and enqueues the question to redis in the format of
// {queue : [
// 		{model.Question1}, {model.Question2}, ..., {model.QuestionN}
// ]}
func (ctx *Context) PostQuestionHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// POST new question and enqueue
	case http.MethodPost:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), MimeJson) {
			http.Error(w, ErrUnsupportedMediaType.Error(), http.StatusUnsupportedMediaType)
			return
		}

		nq, err := decodeQuestion(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		if err := enqueueQuestion(ctx, nq); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := ctx.MongoStore.InsertQuestion(nq); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, _ := json.Marshal(nq)

		httpWriter(http.StatusCreated, b, MimeJson, w)
		return

	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
}

// DeleteQuestionHandler removes a question from the redis question queue
func (ctx *Context) DeleteQuestionHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodDelete:

		id := mux.Vars(r)["id"]
		if len(id) == 0 {
			http.Error(w, "you have to provide an question ID to dequeue", http.StatusBadRequest)
			return
		}

		err := dequeueQuestion(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		httpWriter(http.StatusOK, []byte(id), MimePlain, w)

	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
}

// UpdateQueue commits an update to Redis and MessageQueue
func enqueueQuestion(ctx *Context, nq *model.Question) error {

	// get current queue from redis
	currentQueue := model.QuestionQueue{}
	err := ctx.SessionStore.GetQueue(&currentQueue)
	if err != nil {
		// `redis: nil` == empty redis, ignore
		if err.Error() != "redis: nil" {
			return err
		}
	}

	currentQueue.Queue = append(currentQueue.Queue, nq)

	// update redis
	if err := ctx.SessionStore.SetQueue(currentQueue); err != nil {
		return err
	}

	// create message and push to mq
	ctx.Notifier.PublishMessage(&notifier.Message{
		Type:    notifier.QuestionNew,
		Content: nq,
		UserID:  nq.ID,
	})

	return nil
}

func dequeueQuestion(ctx *Context, id string) error {
	// get current queue from redis
	currentQueue := model.QuestionQueue{}
	err := ctx.SessionStore.GetQueue(&currentQueue)
	if err != nil {
		// `redis: nil` == empty redis, ignore
		if err.Error() != "redis: nil" {
			return err
		}
	}

	// TODO: remove question from currentQueue

	// update redis
	if err := ctx.SessionStore.SetQueue(currentQueue); err != nil {
		return err
	}

	// create message and push to mq
	ctx.Notifier.PublishMessage(&notifier.Message{
		Type:    notifier.QuestionDelete,
		Content: id,
		UserID:  id,
	})

	return nil
}

// HttpWriter takes necessary arguments to write back to client.
func httpWriter(statusCode int, body []byte, contentType string, w http.ResponseWriter) {
	if len(contentType) > 0 {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", MimePlain)
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