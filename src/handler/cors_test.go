package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

const expOrigin = "*"
const expAllowedMethods = "GET, PUT, POST, PATCH, DELETE"
const expAllowedHeaders = "Content-Type, Authorization"
const expExposedHeaders = "Authorization"
const expMaxAge = "600"

func TestCORSHandler(t *testing.T) {
	jsonBuffer := []byte("")
	req, err := http.NewRequest("GET", "/test", bytes.NewBuffer(jsonBuffer))
	if err != nil {
		t.Errorf("Unexpected error occurred [%v]", err)
	}

	sessionStore := createMemStoreRef()
	userStore := createMySQLStoreRefForResting(t)
	ctx, err := NewHandlerContext("random key", sessionStore, userStore, "", "")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(ctx.UsersHandler)
	newCorsHandler := NewCORS(handler)

	newCorsHandler.ServeHTTP(rr, req)

	accessControlAllowOriginGet := rr.Header().Get("Access-Control-Allow-Origin")
	accessControlAllowMethodsGet := rr.Header().Get("Access-Control-Allow-Methods")
	accessControlAllowHeadersGet := rr.Header().Get("Access-Control-Allow-Headers")
	accessControlExposedHeadersGet := rr.Header().Get("Access-Control-Expose-Headers")
	accessControlMaxAgeGet := rr.Header().Get("Access-Control-Max-Age")

	if accessControlAllowOriginGet != expOrigin {
		t.Errorf("Access control allow origin not correct. Got [%v] expected [%v]", accessControlAllowOriginGet, expOrigin)
	}

	if accessControlAllowMethodsGet != expAllowedMethods {
		t.Errorf("Access control allow method not correct. Got [%v] expected [%v]", accessControlAllowMethodsGet, expAllowedMethods)
	}

	if accessControlAllowHeadersGet != expAllowedHeaders {
		t.Errorf("Access control allow headers not correct. Got [%v] expected [%v]", accessControlAllowHeadersGet, expAllowedHeaders)
	}

	if accessControlExposedHeadersGet != expExposedHeaders {
		t.Errorf("Access control exposed not correct. Got [%v] expected [%v]", accessControlExposedHeadersGet, expExposedHeaders)
	}

	if accessControlMaxAgeGet != expMaxAge {
		t.Errorf("Access control age not correct. Got [%v] expected [%v]", accessControlMaxAgeGet, expMaxAge)
	}
}
