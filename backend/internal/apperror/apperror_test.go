package apperror

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestFromMongo(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"no docs", mongo.ErrNoDocuments, http.StatusNotFound},
		{"duplicate", mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000}}}, http.StatusConflict},
		{"deadline", context.DeadlineExceeded, http.StatusServiceUnavailable},
		{"canceled", context.Canceled, http.StatusServiceUnavailable},
		{"server selection", errors.New("server selection error: context deadline exceeded, current topology: mongodb://secret-host:27017"), http.StatusServiceUnavailable},
		{"other", errors.New("boom at /src/file.go:12"), http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ae := As(c.err)
			assert.Equal(t, c.status, ae.HTTPStatus())
			assert.NotContains(t, ae.Message, "secret-host")
			assert.NotContains(t, ae.Message, "file.go")
		})
	}
}

func TestAs_PassesThroughAppError(t *testing.T) {
	assert.Equal(t, http.StatusBadRequest, As(Validation("x", nil)).HTTPStatus())
}
