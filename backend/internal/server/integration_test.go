package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/forgeflow/usermanagement/internal/config"
	"github.com/forgeflow/usermanagement/internal/repository"
	"github.com/forgeflow/usermanagement/internal/server"
	"github.com/forgeflow/usermanagement/internal/service"
)

func TestLifecycleAgainstRealMongo(t *testing.T) {
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		t.Skip("MONGODB_TEST_URI not set; skipping real-MongoDB integration test")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)
	db := client.Database(fmt.Sprintf("um_it_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})
	repo := repository.NewMongoUserRepository(db)
	require.NoError(t, repo.EnsureIndexes(context.Background()))
	h := server.NewRouter(service.NewUserService(repo),
		config.Config{CORSOrigin: "http://localhost:5173", MaxBodySize: 1 << 20},
		slog.New(slog.NewTextHandler(io.Discard, nil)))

	do := func(method, path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		return rec
	}
	stats := func() (int, int, int) {
		var out struct {
			Data struct{ Total, Active, Inactive int }
		}
		require.NoError(t, json.Unmarshal(do("GET", "/api/v1/users/stats", "").Body.Bytes(), &out))
		return out.Data.Total, out.Data.Active, out.Data.Inactive
	}

	rec := do("POST", "/api/v1/users", `{"firstName":"John","lastName":"Doe","email":"john@example.com"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var created struct{ Data struct{ ID string } }
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := created.Data.ID
	tot, act, inact := stats()
	assert.Equal(t, [3]int{1, 1, 0}, [3]int{tot, act, inact})

	assert.Equal(t, 409, do("POST", "/api/v1/users", `{"firstName":"J","lastName":"D","email":"john@example.com"}`).Code)

	assert.Equal(t, 200, do("PUT", "/api/v1/users/"+id, `{"firstName":"Johnny","lastName":"Doe","email":"john@example.com"}`).Code)
	rec = do("GET", "/api/v1/users/"+id, "")
	assert.Equal(t, 200, rec.Code)
	var got struct {
		Success bool
		Data    struct{ ID, Email, FirstName string }
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.True(t, got.Success)
	assert.Equal(t, id, got.Data.ID)
	assert.Equal(t, "john@example.com", got.Data.Email)
	assert.Equal(t, "Johnny", got.Data.FirstName)
	assert.Equal(t, 400, do("GET", "/api/v1/users/not-an-id", "").Code)

	assert.Equal(t, 200, do("PATCH", "/api/v1/users/"+id+"/status", `{"status":"INACTIVE"}`).Code)
	tot, act, inact = stats()
	assert.Equal(t, [3]int{1, 0, 1}, [3]int{tot, act, inact})
	assert.Contains(t, do("GET", "/api/v1/users?status=INACTIVE", "").Body.String(), "Johnny")
	assert.NotContains(t, do("GET", "/api/v1/users?status=ACTIVE", "").Body.String(), "Johnny")

	assert.Equal(t, 200, do("DELETE", "/api/v1/users/"+id, "").Code)
	assert.Equal(t, 404, do("GET", "/api/v1/users/"+id, "").Code)
	tot, _, _ = stats()
	assert.Equal(t, 0, tot)
}

func TestMongoUnavailableReturns503(t *testing.T) {
	// Unroutable host with a tiny timeout: no real MongoDB needed.
	client, err := mongo.Connect(options.Client().
		ApplyURI("mongodb://127.0.0.1:1/").SetServerSelectionTimeout(200 * time.Millisecond))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	h := server.NewRouter(service.NewUserService(repository.NewMongoUserRepository(client.Database("x"))),
		config.Config{CORSOrigin: "http://o", MaxBodySize: 1 << 20},
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/users", nil))
	assert.Equal(t, 503, rec.Code)
	assert.NotContains(t, rec.Body.String(), "127.0.0.1")
	assert.NotContains(t, rec.Body.String(), "mongodb://")

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/users/5f8d0d55b54764421b7156c3", nil))
	assert.Equal(t, 503, rec.Code)
	assert.NotContains(t, rec.Body.String(), "127.0.0.1")
	assert.NotContains(t, rec.Body.String(), "mongodb://")
}
