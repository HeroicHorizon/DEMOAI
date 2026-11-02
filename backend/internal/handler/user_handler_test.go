package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgeflow/usermanagement/internal/config"
	"github.com/forgeflow/usermanagement/internal/server"
	"github.com/forgeflow/usermanagement/internal/service"
	"github.com/forgeflow/usermanagement/internal/testutil"
)

const origin = "http://localhost:5173"

func newServer(repo *testutil.FakeRepo, maxBody int64) http.Handler {
	cfg := config.Config{CORSOrigin: origin, MaxBodySize: maxBody}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return server.NewRouter(service.NewUserService(repo), cfg, log)
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

const goodUser = `{"firstName":"John","lastName":"Doe","email":"john@example.com","phone":"+911234567890"}`

func createUser(t *testing.T, h http.Handler, body string) string {
	rec := do(h, "POST", "/api/v1/users", body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var out struct{ Data struct{ ID string } }
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	return out.Data.ID
}

func TestCreateAndGet(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	id := createUser(t, h, goodUser)
	rec := do(h, "GET", "/api/v1/users/"+id, "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"success":true`)
	assert.Contains(t, rec.Body.String(), "john@example.com")
}

func TestCreate_Validation400(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	assert.Equal(t, 400, do(h, "POST", "/api/v1/users", `{}`).Code)
	assert.Equal(t, 400, do(h, "POST", "/api/v1/users", `{"firstName":"a","lastName":"b","email":"bad"}`).Code)
	assert.Equal(t, 400, do(h, "POST", "/api/v1/users", `not json`).Code)
	assert.Equal(t, 400, do(h, "POST", "/api/v1/users", `{"unknown":1}`).Code)
}

func TestCreate_Duplicate409(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	createUser(t, h, goodUser)
	rec := do(h, "POST", "/api/v1/users", goodUser)
	assert.Equal(t, 409, rec.Code)
	assert.Contains(t, rec.Body.String(), "A user with this email already exists.")
}

func TestGet_400And404(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	assert.Equal(t, 400, do(h, "GET", "/api/v1/users/not-an-id", "").Code)
	assert.Equal(t, 404, do(h, "GET", "/api/v1/users/5f8d0d55b54764421b7156c3", "").Code)
}

func TestUpdate(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	id := createUser(t, h, goodUser)
	other := createUser(t, h, `{"firstName":"A","lastName":"B","email":"other@example.com"}`)
	rec := do(h, "PUT", "/api/v1/users/"+id, `{"firstName":"Jane","lastName":"Doe","email":"john@example.com"}`)
	assert.Equal(t, 200, rec.Code)
	assert.Contains(t, rec.Body.String(), "Jane")
	rec = do(h, "PUT", "/api/v1/users/"+other, `{"firstName":"A","lastName":"B","email":"john@example.com"}`)
	assert.Equal(t, 409, rec.Code)
	assert.Equal(t, 404, do(h, "PUT", "/api/v1/users/5f8d0d55b54764421b7156c3", goodUser).Code)
}

func TestStatus(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	id := createUser(t, h, goodUser)
	rec := do(h, "PATCH", "/api/v1/users/"+id+"/status", `{"status":"INACTIVE"}`)
	assert.Equal(t, 200, rec.Code)
	assert.Contains(t, rec.Body.String(), "INACTIVE")
	assert.Equal(t, 400, do(h, "PATCH", "/api/v1/users/"+id+"/status", `{"status":"WAT"}`).Code)
	assert.Equal(t, 400, do(h, "PATCH", "/api/v1/users/bad/status", `{"status":"ACTIVE"}`).Code)
}

func TestDelete(t *testing.T) {
	repo := testutil.NewFakeRepo()
	h := newServer(repo, 1<<20)
	id := createUser(t, h, goodUser)
	assert.Equal(t, 200, do(h, "DELETE", "/api/v1/users/"+id, "").Code)
	assert.Empty(t, repo.Users)
	assert.Equal(t, 404, do(h, "DELETE", "/api/v1/users/"+id, "").Code)
}

func TestListAndStats(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	id := createUser(t, h, goodUser)
	createUser(t, h, `{"firstName":"A","lastName":"B","email":"other@example.com"}`)
	do(h, "PATCH", "/api/v1/users/"+id+"/status", `{"status":"INACTIVE"}`)

	rec := do(h, "GET", "/api/v1/users?status=ACTIVE&pageSize=1", "")
	assert.Equal(t, 200, rec.Code)
	var out struct {
		Data []map[string]any
		Meta map[string]float64
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Len(t, out.Data, 1)
	assert.EqualValues(t, 1, out.Meta["total"])
	assert.Equal(t, 400, do(h, "GET", "/api/v1/users?status=BAD", "").Code)

	rec = do(h, "GET", "/api/v1/users/stats", "")
	assert.Equal(t, 200, rec.Code)
	assert.Contains(t, rec.Body.String(), `"total":2`)
	assert.Contains(t, rec.Body.String(), `"active":1`)
	assert.Contains(t, rec.Body.String(), `"inactive":1`)
}

func TestDBUnavailable_503NoLeak(t *testing.T) {
	repo := testutil.NewFakeRepo()
	repo.Err = errors.New("server selection error: mongodb://admin:pw@secret-host:27017 goroutine 1 [running]")
	h := newServer(repo, 1<<20)
	for _, req := range [][3]string{
		{"GET", "/api/v1/users", ""}, {"GET", "/api/v1/users/stats", ""}, {"POST", "/api/v1/users", goodUser},
	} {
		rec := do(h, req[0], req[1], req[2])
		assert.Equal(t, 503, rec.Code, req[1])
		body := rec.Body.String()
		for _, leak := range []string{"mongodb://", "secret-host", "admin", "goroutine", "27017"} {
			assert.NotContains(t, body, leak)
		}
	}
}

func TestOversizedBodyRejected(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 64)
	rec := do(h, "POST", "/api/v1/users", `{"firstName":"`+strings.Repeat("a", 500)+`"}`)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestCORS(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	req := httptest.NewRequest("OPTIONS", "/api/v1/users", nil)
	req.Header.Set("Origin", origin)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, 204, rec.Code)
	assert.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"))

	req = httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestContextCancellationPropagated(t *testing.T) {
	repo := testutil.NewFakeRepo()
	h := newServer(repo, 1<<20)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest("GET", "/api/v1/users", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.NotNil(t, repo.LastCtx)
	assert.ErrorIs(t, repo.LastCtx.Err(), context.Canceled)
	assert.Equal(t, 503, rec.Code)
}

func TestHealthz(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	assert.Equal(t, 200, do(h, "GET", "/healthz", "").Code)
}

type envelope struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func parseEnvelope(t *testing.T, rec *httptest.ResponseRecorder) envelope {
	t.Helper()
	var env envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env), rec.Body.String())
	return env
}

func TestGet_200EnvelopeShape(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	id := createUser(t, h, goodUser)
	rec := do(h, "GET", "/api/v1/users/"+id, "")
	require.Equal(t, http.StatusOK, rec.Code)
	env := parseEnvelope(t, rec)
	assert.True(t, env.Success)
	assert.Equal(t, id, env.Data["id"])
	assert.Equal(t, "john@example.com", env.Data["email"])
	assert.Equal(t, "John", env.Data["firstName"])
	assert.Nil(t, env.Error)
}

func TestGet_InvalidIDVariants(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	for _, id := range []string{"not-an-id", "%20", "5f8d0d55b54764421b7156c"} {
		rec := do(h, "GET", "/api/v1/users/"+id, "")
		assert.Equal(t, 400, rec.Code, id)
		env := parseEnvelope(t, rec)
		assert.False(t, env.Success)
		require.NotNil(t, env.Error, id)
		assert.Equal(t, "VALIDATION_ERROR", env.Error.Code)
		assert.Equal(t, "Invalid user ID.", env.Error.Message)
	}
	rec := do(h, "GET", "/api/v1/users/5f8d0d55b54764421b7156c3", "")
	assert.Equal(t, 404, rec.Code)
	env := parseEnvelope(t, rec)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "NOT_FOUND", env.Error.Code)
	assert.Equal(t, "User not found.", env.Error.Message)
}

func TestGet_Unexpected500NoLeak(t *testing.T) {
	repo := testutil.NewFakeRepo()
	repo.Err = errors.New("boom db failure")
	h := newServer(repo, 1<<20)
	rec := do(h, "GET", "/api/v1/users/5f8d0d55b54764421b7156c3", "")
	assert.Equal(t, 500, rec.Code)
	env := parseEnvelope(t, rec)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
	assert.NotContains(t, rec.Body.String(), "boom")
}

func TestGet_MissingID(t *testing.T) {
	h := newServer(testutil.NewFakeRepo(), 1<<20)
	rec := do(h, "GET", "/api/v1/users/", "")
	// No route matches an empty id segment, so the router answers 404 (not the JSON envelope).
	assert.Equal(t, 404, rec.Code)
}
