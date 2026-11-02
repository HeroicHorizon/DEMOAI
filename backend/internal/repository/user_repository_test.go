package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/forgeflow/usermanagement/internal/apperror"
	"github.com/forgeflow/usermanagement/internal/model"
)

func newRepo(t *testing.T) *MongoUserRepository {
	t.Helper()
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		t.Skip("MONGODB_TEST_URI not set; skipping real-MongoDB repository test")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)
	db := client.Database(fmt.Sprintf("um_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})
	repo := NewMongoUserRepository(db)
	require.NoError(t, repo.EnsureIndexes(context.Background()))
	return repo
}

func mk(email, first string, status string, at time.Time) *model.User {
	return &model.User{FirstName: first, LastName: "Doe", Email: email, Status: status, CreatedAt: at, UpdatedAt: at}
}

func TestBuildFilter_EscapesRegex(t *testing.T) {
	f := buildFilter(model.ListQuery{Search: "a.b*(c", Status: "ACTIVE"})
	assert.Len(t, f, 2)
	or := f[1].Value.(bson.A)
	re := or[0].(bson.D)[0].Value.(bson.Regex)
	assert.Equal(t, `a\.b\*\(c`, re.Pattern)
}

func TestCRUD(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	u, err := repo.Create(ctx, mk("a@x.com", "A", "ACTIVE", now))
	require.NoError(t, err)
	got, err := repo.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "a@x.com", got.Email)

	later := now.Add(time.Hour)
	up, err := repo.Update(ctx, u.ID, model.UserInput{FirstName: "B", LastName: "Z", Email: "b@x.com"}, later)
	require.NoError(t, err)
	assert.Equal(t, "B", up.FirstName)
	assert.True(t, up.UpdatedAt.Equal(later))
	assert.True(t, up.CreatedAt.Equal(now))

	st, err := repo.UpdateStatus(ctx, u.ID, "INACTIVE", later)
	require.NoError(t, err)
	assert.Equal(t, "INACTIVE", st.Status)

	require.NoError(t, repo.Delete(ctx, u.ID))
	_, err = repo.GetByID(ctx, u.ID)
	assert.Equal(t, apperror.KindNotFound, apperror.As(err).Kind)
	assert.Equal(t, apperror.KindNotFound, apperror.As(repo.Delete(ctx, u.ID)).Kind)
}

func TestUniqueEmail(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()
	_, err := repo.Create(ctx, mk("a@x.com", "A", "ACTIVE", now))
	require.NoError(t, err)
	_, err = repo.Create(ctx, mk("a@x.com", "B", "ACTIVE", now))
	assert.Equal(t, apperror.KindConflict, apperror.As(err).Kind)

	u2, err := repo.Create(ctx, mk("c@x.com", "C", "ACTIVE", now))
	require.NoError(t, err)
	_, err = repo.Update(ctx, u2.ID, model.UserInput{FirstName: "C", LastName: "D", Email: "a@x.com"}, now)
	assert.Equal(t, apperror.KindConflict, apperror.As(err).Kind)
}

func TestListSearchFilterPaginationStats(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	base := time.Now().UTC().Truncate(time.Millisecond)
	for i := 0; i < 5; i++ {
		status := "ACTIVE"
		if i%2 == 1 {
			status = "INACTIVE"
		}
		_, err := repo.Create(ctx, mk(fmt.Sprintf("user%d@x.com", i), fmt.Sprintf("First%d", i), status, base.Add(time.Duration(i)*time.Minute)))
		require.NoError(t, err)
	}
	_, err := repo.Create(ctx, mk("dot.user@x.com", "D.t", "ACTIVE", base.Add(time.Hour)))
	require.NoError(t, err)

	users, total, err := repo.List(ctx, model.ListQuery{Page: 1, PageSize: 4})
	require.NoError(t, err)
	assert.EqualValues(t, 6, total)
	assert.Len(t, users, 4)
	assert.Equal(t, "dot.user@x.com", users[0].Email) // newest first
	users, _, _ = repo.List(ctx, model.ListQuery{Page: 2, PageSize: 4})
	assert.Len(t, users, 2)

	users, total, _ = repo.List(ctx, model.ListQuery{Page: 1, PageSize: 10, Status: "INACTIVE"})
	assert.EqualValues(t, 2, total)
	assert.Len(t, users, 2)

	users, _, _ = repo.List(ctx, model.ListQuery{Page: 1, PageSize: 10, Search: "FIRST3"})
	assert.Len(t, users, 1)
	users, _, _ = repo.List(ctx, model.ListQuery{Page: 1, PageSize: 10, Search: "user1@x"})
	assert.Len(t, users, 1)

	// "." must be literal: only the "D.t" user matches, not every user.
	users, _, _ = repo.List(ctx, model.ListQuery{Page: 1, PageSize: 10, Search: "D.t"})
	assert.Len(t, users, 1)
	users, _, _ = repo.List(ctx, model.ListQuery{Page: 1, PageSize: 10, Search: ".*"})
	assert.Empty(t, users)

	s, err := repo.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, model.Stats{Total: 6, Active: 4, Inactive: 2}, s)
}
