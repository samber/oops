package oopsrecoverygin_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/samber/oops/recovery/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGinOopsRecovery_NoPanic(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(GinOopsRecovery())
	router.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
}

func TestGinOopsRecovery_PanicWithString(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(GinOopsRecovery())
	router.GET("/panic-string", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-string", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGinOopsRecovery_PanicWithError(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(GinOopsRecovery())
	router.GET("/panic-error", func(c *gin.Context) {
		panic(errors.New("db failure"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-error", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGinOopsRecovery_PanicWithInt(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(GinOopsRecovery())
	router.GET("/panic-int", func(c *gin.Context) {
		panic(42)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-int", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGinOopsRecovery_ErrorAddedToContext(t *testing.T) {
	t.Parallel()

	// The error-capturing middleware must be registered BEFORE GinOopsRecovery()
	// so that when it resumes after c.Next() returns, GinOopsRecovery has already
	// added the error to c.Errors.
	router := gin.New()

	capturedErrorCount := 0
	router.Use(func(c *gin.Context) {
		c.Next()
		capturedErrorCount = len(c.Errors)
	})

	router.Use(GinOopsRecovery())

	router.GET("/panic-ctx", func(c *gin.Context) {
		panic("capture me")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-ctx", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	if capturedErrorCount == 0 {
		t.Error("expected error to be added to gin context, got none")
	}
}

func TestGinOopsRecovery_NoPanicDoesNotAbort(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(GinOopsRecovery())

	handlerCalled := false
	router.GET("/chain", func(c *gin.Context) {
		handlerCalled = true
		c.String(http.StatusOK, "chained")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/chain", nil)
	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called when no panic occurs")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
