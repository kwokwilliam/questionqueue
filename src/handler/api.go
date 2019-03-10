package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"questionqueue/src/db"
	"questionqueue/src/model"
	"questionqueue/src/session"
	"questionqueue/src/websocket"
	"strings"
)

const (
	ErrUnsupportedMediaType = "unsupported media type"
	ErrInvalidCredentials   = "invalid credentials"
	ErrMethodNotAllowed     = "method not allowed"
	ErrUnauthorizedSession  = "unauthorized session"
)

// TA/teacher control
func (ctx *Context) TeacherHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// Create new TA/teacher.
	case http.MethodPost:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, ErrUnsupportedMediaType, http.StatusUnsupportedMediaType)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t := model.Teacher{
			ID:        res.InsertedID,
			Email:     nt.Email,
			FirstName: nt.FirstName,
			LastName:  nt.LastName,
		}

		js, _ := json.Marshal(t)

		if err := ctx.SessionStore.Save(res.InsertedID, t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		httpWriter(http.StatusCreated, js, "application/json", w)
		return

	//  Update information for a TA/teacher.
	case http.MethodPatch:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, ErrUnsupportedMediaType, http.StatusUnsupportedMediaType)
			return
		}

		currentState := &SessionState{}
		_, err := session.GetState(r, ctx.Key, ctx.SessionStore, currentState)
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

		// since `SessionState` can be any interface, force cast into `TeacherUpdate`
		if tu.Email != currentState.Interface.(model.TeacherUpdate).Email {
			http.Error(w, "you can only modify your profile", http.StatusForbidden)
			return
		}

		if err := tu.VerifyTeacherUpdate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := ctx.MongoStore.UpdateTeacher(tu)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		js, _ := json.Marshal(tu)

		// TODO: `res.UpsertedID` should match whatever the original session ID is, *double check* redis
		if err := ctx.SessionStore.Save(res.UpsertedID, tu); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		httpWriter(http.StatusCreated, js, "application/json", w)
		return

	default:
		http.Error(w, ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
}

func (ctx *Context) TeacherProfileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	id := mux.Vars(r)["id"]
	if id != "me" {
		http.Error(w, "you can only get your own profile", http.StatusForbidden)
	}

	currentState := &SessionState{}
	_, err := session.GetState(r, ctx.Key, ctx.SessionStore, currentState)
	if err != nil {
		http.Error(w, ErrUnauthorizedSession, http.StatusUnauthorized)
		return
	}

	httpWriter(http.StatusOK, currentState.Interface, "application/json", w)
}

// TA/teacher session control
func (ctx *Context) TeacherSessionHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// login
	case http.MethodPost:

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, ErrUnsupportedMediaType, http.StatusUnsupportedMediaType)
			return
		}

		tl, err := decodeTeacherLogin(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		teachers, err := ctx.MongoStore.GetTeacherByEmail(tl.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(teachers) == 1 {
			http.Error(w, ErrInvalidCredentials, http.StatusForbidden)
			return
		}

		if len(teachers) > 1 {
			log.Printf("got more than 1 matched users.\nemail: %v", tl.Email)
			http.Error(w, ErrInvalidCredentials, http.StatusForbidden)
		}

		// check index 0 as there should only be one matched result
		if err := teachers[0].Authenticate(tl.Password); err != nil {
			http.Error(w, ErrInvalidCredentials, http.StatusForbidden)
			return
		}

		js, _ := json.Marshal(teachers[0])

		_, err = session.BeginSession(ctx.Key, ctx.SessionStore, js, w)

		httpWriter(http.StatusOK, js, "application/json", w)

	// delete session
	case http.MethodDelete:

		var err error

		// `SessionState` is discarded
		_, err = session.GetState(r, ctx.Key, ctx.SessionStore, &SessionState{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		_, err = session.EndSession(r, ctx.Key, ctx.SessionStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		httpWriter(http.StatusOK, "you have been signed out", "application/json", w)

	default:
		http.Error(w, ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

}

// Post new question and enqueue the user.
func (ctx *Context) PostQuestionHandler(w http.ResponseWriter, r *http.Request) {

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, ErrUnsupportedMediaType, http.StatusUnsupportedMediaType)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, ErrMethodNotAllowed, http.StatusMethodNotAllowed)
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

	if err := updateQueue(r, ctx, nq, websocket.QuestionNew); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpWriter(http.StatusCreated, nil, "", w)
	return
}

// decoders; probably cannot be further refactored
func decodeQuestion(d io.ReadCloser) (*model.Question, error) {
	decoder := json.NewDecoder(d)
	var i *model.Question

	if err := decoder.Decode(i); err != nil {
		return nil, err
	} else {
		return i, nil
	}
}

func decodeNewTeacher(d io.ReadCloser) (*model.NewTeacher, error) {
	decoder := json.NewDecoder(d)
	var i *model.NewTeacher

	if err := decoder.Decode(i); err != nil {
		return nil, err
	} else {
		return i, nil
	}
}

//func decodeTeacher(d io.ReadCloser) (*model.Teacher, error) {
//	decoder := json.NewDecoder(d)
//	var i *model.Teacher
//
//	if err := decoder.Decode(i); err != nil {
//		return nil, err
//	} else {
//		return i, nil
//	}
//}

func decodeTeacherUpdate(d io.ReadCloser) (*model.TeacherUpdate, error) {
	decoder := json.NewDecoder(d)
	var i *model.TeacherUpdate

	if err := decoder.Decode(i); err != nil {
		return nil, err
	} else {
		return i, nil
	}
}

func decodeTeacherLogin(d io.ReadCloser) (*model.TeacherLogin, error) {
	decoder := json.NewDecoder(d)
	var i *model.TeacherLogin

	if err := decoder.Decode(i); err != nil {
		return nil, err
	} else {
		return i, nil
	}
}

// HttpWriter takes necessary arguments to write back to client.
func httpWriter(statusCode int, body interface{}, contentType string, w http.ResponseWriter) {

	// check marshalling doesnt error out
	marshaledBody, err := json.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// write results
	if len(contentType) > 0 {
		w.Header().Set("Content-Type", contentType)
	}

	w.WriteHeader(statusCode)

	if len(marshaledBody) > 0 {
		_, _ = w.Write(marshaledBody)
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
	ctx.Notifier.PublishMessage(&websocket.Message{
		Type:    messageType,
		Content: i,
		UserID:  sid,
	})

	return nil
}
