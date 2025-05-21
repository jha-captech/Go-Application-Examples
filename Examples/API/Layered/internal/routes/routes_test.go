package routes_test

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestRoutes_AddRoute(t *testing.T) {
// 	router := mux.NewRouter()
// 	AddRoute(router, "/test", TestHandler)

// 	req, err := http.NewRequest("GET", "/test", nil)
// 	assert.NoError(t, err)

// 	rr := httptest.NewRecorder()
// 	router.ServeHTTP(rr, req)

// 	assert.Equal(t, http.StatusOK, rr.Code)
// 	assert.Equal(t, "Test successful", rr.Body.String())
// }
