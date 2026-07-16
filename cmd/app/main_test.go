package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupMux(t *testing.T) {
	mux := setupMux()
	if mux == nil {
		t.Fatal("setupMux returned nil")
	}

	// Test if expected routes are registered by making dummy requests
	testCases := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/optimize"},
		{"GET", "/api/v1/admin/prices"},
		{"POST", "/api/v1/admin/prices"},
		{"GET", "/swagger/index.html"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			// ServeHTTP will hit the 404 handler if the route doesn't exist
			mux.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found (got 404)", tc.method, tc.path)
			}
		})
	}
}

func TestMain_Success(t *testing.T) {
	// Mock startServer to simulate successful server start
	origStartServer := startServer
	defer func() { startServer = origStartServer }()

	startServerCalled := false
	startServer = func(addr string, handler http.Handler) error {
		startServerCalled = true
		return nil
	}

	// We shouldn't hit osExit, but let's mock it just in case
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	osExitCalled := false
	osExit = func(code int) {
		osExitCalled = true
	}

	// Run main
	main()

	if !startServerCalled {
		t.Error("Expected startServer to be called")
	}
	if osExitCalled {
		t.Error("Did not expect osExit to be called")
	}
}

func TestMain_ListenError(t *testing.T) {
	// Mock startServer to return an error
	origStartServer := startServer
	defer func() { startServer = origStartServer }()

	startServerCalled := false
	startServer = func(addr string, handler http.Handler) error {
		startServerCalled = true
		return errors.New("mock listen error")
	}

	// Mock osExit to prevent actual process exit
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	osExitCode := -1
	osExit = func(code int) {
		osExitCode = code
	}

	// Run main
	main()

	if !startServerCalled {
		t.Error("Expected startServer to be called")
	}
	if osExitCode != 1 {
		t.Errorf("Expected osExit(1) to be called, got %d", osExitCode)
	}
}
