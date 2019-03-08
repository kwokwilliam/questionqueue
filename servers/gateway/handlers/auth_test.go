package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"questionqueue/servers/gateway/models/users"
	"questionqueue/servers/gateway/sessions"
	"reflect"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// j is a shortcut used for tests to create a new byte buffer
// out of a json string
func j(jsonString string) *bytes.Buffer {
	jsonBuffer := []byte(jsonString)
	return bytes.NewBuffer(jsonBuffer)
}

// createMemStoreRef creates a redis store reference
// for testing
func createMemStoreRef() sessions.Store {
	store := sessions.NewMemStore(time.Second*5, time.Minute*5)
	return store
}

// createMySQLStoreRefForResting creates a mysql store reference
// for testing.
func createMySQLStoreRefForResting(t *testing.T) users.Store {
	mysqlstore := users.NewFakeStore()
	return mysqlstore
}

// createNewHandler creates a new handler context and returns it for the test
// it also catches any errors in handler context creation
func createNewHandler(SigningKey string, t *testing.T) *HandlerContext {
	SessionStore := createMemStoreRef()
	userStore := createMySQLStoreRefForResting(t)
	ctx, err := NewHandlerContext(SigningKey, SessionStore, userStore, "", "")
	if err != nil {
		t.Errorf("Failed to make handler context [%v]", err)
	}
	return ctx
}

// validateStatusCode checks if the response's status code is as expected
func validateStatusCode(rr *httptest.ResponseRecorder, expectedStatus int, t *testing.T) {
	if status := rr.Code; status != expectedStatus {
		t.Errorf("Unexpected status code. Expected [%v] but got [%v]", expectedStatus, status)
	}
}

// validateResponseBodyUser checks if the response's body is as expected for user output
func validateResponseBodyUser(rr *httptest.ResponseRecorder, expectedUser *users.User, t *testing.T) {
	responseUser := &users.User{}
	if err := json.NewDecoder(rr.Body).Decode(responseUser); err != nil {
		t.Errorf("Unexpected error when decoding body [%v]", err)
	}
	expectedUser.Email = ""
	expectedUser.PassHash = []byte(nil)
	if !reflect.DeepEqual(responseUser, expectedUser) {
		t.Errorf("Unexpected user returned to body. Expected [%v] but got [%v]", expectedUser, responseUser)
	}
}

// createNewRequestAndHandleErrors creates
func createNewRequestAndHandleErrors(method string, path string, jsonString string, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, path, j(jsonString))
	if err != nil {
		t.Errorf("Unexpected error occured [%v]", err)
	}
	return req
}

// validateContentType checks if the response's content type is as expected.
func validateContentType(rr *httptest.ResponseRecorder, expectedContentType string, t *testing.T) {
	if ctype := rr.Header().Get("Content-Type"); ctype != expectedContentType {
		t.Errorf("Unexpected content type header. Expected [%v] but got [%v]", expectedContentType, ctype)
	}
}

func TestUsersHandler(t *testing.T) {
	SigningKey := "random string 123"
	cases := []struct {
		name                string
		method              string
		endpoint            string
		jsonString          string
		dontSetJSON         bool
		expectedCode        int
		checkContentType    bool
		expectedContentType string
		checkBody           bool
		expectedUser        *users.User
	}{
		{
			name:             "Gets status method not allowed for GET requests",
			method:           "GET",
			endpoint:         "/v1/users",
			jsonString:       "",
			expectedCode:     http.StatusMethodNotAllowed,
			checkContentType: false,
			checkBody:        false,
		},
		{
			name:         "Test posting unsupported media type",
			method:       "POST",
			endpoint:     "/v1/users",
			dontSetJSON:  true,
			jsonString:   "",
			expectedCode: http.StatusUnsupportedMediaType,
		},
		{
			name:         "Make sure invalid JSON cannot be decoded",
			method:       "POST",
			endpoint:     "/v1/users",
			jsonString:   "123",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "Attempt to create invalid user",
			method:       "POST",
			endpoint:     "/v1/users",
			jsonString:   `{"email": "abc@cdf.com", "password":"password", "passwordConf":"mismatch"}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:     "Check valid input",
			method:   "POST",
			endpoint: "/v1/users",
			jsonString: `{"email": "abc@cdf.com", "password":"password2", "passwordConf":"password2",
					"firstName": "firstname", "lastName": "lastname"}`,
			expectedCode: http.StatusCreated,
			checkBody:    true,
			expectedUser: &users.User{
				ID: int64(1)},
		},
	}

	for _, c := range cases {
		// create the request
		req := createNewRequestAndHandleErrors(c.method, c.endpoint, c.jsonString, t)
		// if dontSetJson is true, do not set the content type json header
		if !c.dontSetJSON {
			req.Header.Set(headerContentType, contentTypeJSON)
		}
		ctx := createNewHandler(SigningKey, t)

		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/v1/users", ctx.UsersHandler)
		router.ServeHTTP(rr, req)

		// Will always check an expected code
		validateStatusCode(rr, c.expectedCode, t)
		// check content type if requested
		if c.checkContentType {
			validateContentType(rr, c.expectedContentType, t)
		}

		// check content type if requested
		if c.checkBody {
			validateResponseBodyUser(rr, c.expectedUser, t)
		}
	}
}

func TestSpecificUsersHandler(t *testing.T) {
	SigningKey := "random string 123"
	cases := []struct {
		name                string
		method              string
		endpoint            string
		jsonString          string
		doNotAuthorize      bool
		dummyUser           *users.User
		dontSetJSON         bool
		expectedCode        int
		checkContentType    bool
		expectedContentType string
		checkBody           bool
		expectedUser        *users.User
		checkExpectedUser   bool
	}{
		{
			name:           "Test unauthorized request",
			method:         "GET",
			endpoint:       "/v1/users/2",
			doNotAuthorize: true,
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:       "",
			expectedCode:     http.StatusUnauthorized,
			checkContentType: false,
			checkBody:        false,
		},
		{
			name:     "Test invalid ID",
			method:   "GET",
			endpoint: "/v1/users/dfsafj",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:       "",
			expectedCode:     http.StatusBadRequest,
			checkContentType: false,
			checkBody:        false,
		},
		{
			name:     "Test GET unfound user",
			method:   "GET",
			endpoint: "/v1/users/4",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:       "",
			expectedCode:     http.StatusNotFound,
			checkContentType: false,
			checkBody:        false,
		},
		{
			name:     "Test GET found user and body match",
			method:   "GET",
			endpoint: "/v1/users/1",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedUser: &users.User{
				ID: 1},
			checkExpectedUser:   true,
			jsonString:          "asdf",
			expectedCode:        http.StatusOK,
			checkContentType:    true,
			expectedContentType: contentTypeJSON,
			checkBody:           true,
		},
		{
			name:     "Test PATCH forbidden",
			method:   "PATCH",
			endpoint: "/v1/users/1",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:       "asdf",
			expectedCode:     http.StatusForbidden,
			checkContentType: false,
			checkBody:        false,
		},
		{
			name:     "Test PATCH Content type JSON fail",
			method:   "PATCH",
			endpoint: "/v1/users/2",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			dontSetJSON:      true,
			jsonString:       "asdf",
			expectedCode:     http.StatusUnsupportedMediaType,
			checkContentType: false,
			checkBody:        false,
		},
		{
			name:     "Test PATCH Successful and writes response",
			method:   "PATCH",
			endpoint: "/v1/users/2",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedUser: &users.User{
				ID: 1},
			checkExpectedUser:   true,
			jsonString:          `{"firstName":"newfirst", "lastName": "newlast"}`,
			expectedCode:        http.StatusCreated,
			checkContentType:    true,
			expectedContentType: contentTypeJSON,
			checkBody:           true,
		},
		{
			name:     "Test POST invalid method",
			method:   "POST",
			endpoint: "/v1/users/2",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, c := range cases {
		// create the request
		req := createNewRequestAndHandleErrors(c.method, c.endpoint, c.jsonString, t)
		// if dontSetJson is true, do not set the content type json header
		if !c.dontSetJSON {
			req.Header.Set(headerContentType, contentTypeJSON)
		}
		ctx := createNewHandler(SigningKey, t)

		rr := httptest.NewRecorder()

		if !c.doNotAuthorize {
			sessionState := NewSessionState(time.Now(), c.dummyUser)
			_, err := sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, rr)
			req.Header.Set("Authorization", rr.Header().Get("Authorization"))
			if err != nil {
				t.Errorf("Unexpected error occurred when creating new session [%v]", err)
			}
		}
		router := mux.NewRouter()
		router.HandleFunc("/v1/users/{uid}", ctx.SpecificUsersHandler)
		router.ServeHTTP(rr, req)

		// Will always check an expected code
		validateStatusCode(rr, c.expectedCode, t)
		// check content type if requested
		if c.checkContentType {
			validateContentType(rr, c.expectedContentType, t)
		}

		// check content type if requested
		if c.checkBody {
			validateResponseBodyUser(rr, c.expectedUser, t)
		}
	}
}

func TestSessionsHandler(t *testing.T) {
	SigningKey := "random string 123"
	cases := []struct {
		name                string
		method              string
		endpoint            string
		jsonString          string
		doNotAuthorize      bool
		dummyUser           *users.User
		dontSetJSON         bool
		expectedCode        int
		checkContentType    bool
		expectedContentType string
		checkBody           bool
		expectedUser        *users.User
		checkExpectedUser   bool
	}{
		{
			name:     "Test bad method",
			method:   "GET",
			endpoint: "/v1/sessions",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:        "Test unauthorized request",
			method:      "POST",
			endpoint:    "/v1/sessions",
			dontSetJSON: true,
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedCode: http.StatusUnsupportedMediaType,
		},
		{
			name:     "Test invalid json",
			method:   "POST",
			endpoint: "/v1/sessions",
			dummyUser: &users.User{
				ID:        2,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:   `asdf`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:     "Test profile not found",
			method:   "POST",
			endpoint: "/v1/sessions",
			dummyUser: &users.User{
				ID:        2,
				Email:     "emailthatwillfail@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:   `{"email":"emailcannotfound@email.com", "password":"password"}`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:     "Test failed credentials",
			method:   "POST",
			endpoint: "/v1/sessions",
			dummyUser: &users.User{
				ID:        2,
				Email:     "emailthatwillfail@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:   `{"email":"emailfailedcreds@email.com", "password":"password"}`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:     "Test working credentials",
			method:   "POST",
			endpoint: "/v1/sessions",
			dummyUser: &users.User{
				ID:        2,
				Email:     "emailthatwillfail@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			jsonString:          `{"email":"emailworkingcreds@email.com", "password":"password"}`,
			expectedCode:        http.StatusCreated,
			checkContentType:    true,
			expectedContentType: contentTypeJSON,
			checkBody:           true,
			expectedUser:        &users.User{ID: 1},
		},
	}

	for _, c := range cases {
		// create the request
		req := createNewRequestAndHandleErrors(c.method, c.endpoint, c.jsonString, t)
		// if dontSetJson is true, do not set the content type json header
		if !c.dontSetJSON {
			req.Header.Set(headerContentType, contentTypeJSON)
		}
		ctx := createNewHandler(SigningKey, t)

		rr := httptest.NewRecorder()

		if !c.doNotAuthorize {
			sessionState := NewSessionState(time.Now(), c.dummyUser)
			_, err := sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, rr)
			req.Header.Set("Authorization", rr.Header().Get("Authorization"))
			if err != nil {
				t.Errorf("Unexpected error occurred when creating new session [%v]", err)
			}
		}
		router := mux.NewRouter()
		router.HandleFunc("/v1/sessions", ctx.SessionsHandler)
		router.ServeHTTP(rr, req)

		// Will always check an expected code
		validateStatusCode(rr, c.expectedCode, t)
		// check content type if requested
		if c.checkContentType {
			validateContentType(rr, c.expectedContentType, t)
		}

		// check content type if requested
		if c.checkBody {
			validateResponseBodyUser(rr, c.expectedUser, t)
		}
	}
}

func TestSpecificSessionsHandler(t *testing.T) {
	SigningKey := "random string 123"
	cases := []struct {
		name                string
		method              string
		endpoint            string
		jsonString          string
		doNotAuthorize      bool
		dummyUser           *users.User
		dontSetJSON         bool
		expectedCode        int
		checkContentType    bool
		expectedContentType string
		checkBody           bool
		expectedBody        string
	}{
		{
			name:     "Test bad method",
			method:   "GET",
			endpoint: "/v1/sessions/mine",
			dummyUser: &users.User{
				ID:        1,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "Test forbidden path",
			method:   "DELETE",
			endpoint: "/v1/sessions/notmine",
			dummyUser: &users.User{
				ID:        1,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name:     "Test session delete success",
			method:   "DELETE",
			endpoint: "/v1/sessions/mine",
			dummyUser: &users.User{
				ID:        1,
				Email:     "email@email.com",
				PassHash:  []byte("passHash"),
				FirstName: "firstname",
				LastName:  "lastname",
			},
			expectedCode:        http.StatusOK,
			checkContentType:    true,
			expectedContentType: "text/plain",
			checkBody:           true,
			expectedBody:        "signed out",
		},
	}

	for _, c := range cases {
		// create the request
		req := createNewRequestAndHandleErrors(c.method, c.endpoint, c.jsonString, t)
		// if dontSetJson is true, do not set the content type json header
		if !c.dontSetJSON {
			req.Header.Set(headerContentType, contentTypeJSON)
		}
		ctx := createNewHandler(SigningKey, t)

		rr := httptest.NewRecorder()

		if !c.doNotAuthorize {
			sessionState := NewSessionState(time.Now(), c.dummyUser)
			_, err := sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, rr)
			req.Header.Set("Authorization", rr.Header().Get("Authorization"))
			if err != nil {
				t.Errorf("Unexpected error occurred when creating new session [%v]", err)
			}
		}
		router := mux.NewRouter()
		router.HandleFunc("/v1/sessions/{uid}", ctx.SpecificSessionHandler)
		router.ServeHTTP(rr, req)

		// Will always check an expected code
		validateStatusCode(rr, c.expectedCode, t)
		// check content type if requested
		if c.checkContentType {
			validateContentType(rr, c.expectedContentType, t)
		}

		// check content type if requested
		if c.checkBody && rr.Body.String() != c.expectedBody {
			t.Errorf("Body not as expected. Expected [%v] got [%v]", c.expectedBody, rr.Body.String())
		}
	}
}
