package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgeflow/usermanagement/internal/apperror"
	"github.com/forgeflow/usermanagement/internal/model"
	"github.com/forgeflow/usermanagement/internal/testutil"
)

func valid(email string) model.UserInput {
	return model.UserInput{FirstName: "John", LastName: "Doe", Email: email, Phone: "+911234567890",
		Department: "Engineering", Designation: "Engineer"}
}

func setup() (*UserService, *testutil.FakeRepo) {
	repo := testutil.NewFakeRepo()
	return NewUserService(repo), repo
}

func kindOf(t *testing.T, err error) apperror.Kind {
	t.Helper()
	var ae *apperror.Error
	require.True(t, errors.As(err, &ae))
	return ae.Kind
}

func TestCreate_Valid(t *testing.T) {
	svc, repo := setup()
	u, err := svc.Create(context.Background(), valid("  John.Doe@Example.com "))
	require.NoError(t, err)
	assert.Equal(t, "john.doe@example.com", u.Email)
	assert.Equal(t, model.StatusActive, u.Status)
	assert.Equal(t, u.CreatedAt, u.UpdatedAt)
	assert.Len(t, repo.Users, 1)
}

func TestCreate_MissingFields(t *testing.T) {
	svc, repo := setup()
	_, err := svc.Create(context.Background(), model.UserInput{})
	require.Error(t, err)
	assert.Equal(t, apperror.KindValidation, kindOf(t, err))
	ae := err.(*apperror.Error)
	assert.Contains(t, ae.Details, "firstName")
	assert.Contains(t, ae.Details, "lastName")
	assert.Contains(t, ae.Details, "email")
	assert.Empty(t, repo.Users)
}

func TestCreate_InvalidEmail(t *testing.T) {
	svc, repo := setup()
	for _, e := range []string{"nope", "a@b", "John <j@x.com>", "@x.com"} {
		_, err := svc.Create(context.Background(), valid(e))
		require.Error(t, err, e)
		assert.Equal(t, apperror.KindValidation, kindOf(t, err), e)
	}
	assert.Empty(t, repo.Users)
}

func TestCreate_InvalidPhone(t *testing.T) {
	svc, _ := setup()
	in := valid("a@b.com")
	in.Phone = "abc"
	_, err := svc.Create(context.Background(), in)
	assert.Equal(t, apperror.KindValidation, kindOf(t, err))
}

func TestCreate_DuplicateEmail(t *testing.T) {
	svc, _ := setup()
	_, err := svc.Create(context.Background(), valid("a@b.com"))
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), valid("A@B.com"))
	assert.Equal(t, apperror.KindConflict, kindOf(t, err))
}

func TestUpdate_RefreshesUpdatedAt(t *testing.T) {
	svc, _ := setup()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return t0 }
	u, err := svc.Create(context.Background(), valid("a@b.com"))
	require.NoError(t, err)
	svc.now = func() time.Time { return t0.Add(time.Hour) }
	in := valid("a@b.com")
	in.FirstName = "Jane"
	up, err := svc.Update(context.Background(), u.ID.Hex(), in)
	require.NoError(t, err)
	assert.Equal(t, "Jane", up.FirstName)
	assert.Equal(t, t0.Add(time.Hour), up.UpdatedAt)
	assert.Equal(t, t0, up.CreatedAt)
}

func TestUpdate_DuplicateEmail(t *testing.T) {
	svc, _ := setup()
	_, _ = svc.Create(context.Background(), valid("a@b.com"))
	u2, _ := svc.Create(context.Background(), valid("c@d.com"))
	_, err := svc.Update(context.Background(), u2.ID.Hex(), valid("a@b.com"))
	assert.Equal(t, apperror.KindConflict, kindOf(t, err))
}

func TestUpdate_NotFoundAndMalformed(t *testing.T) {
	svc, _ := setup()
	_, err := svc.Update(context.Background(), "5f8d0d55b54764421b7156c3", valid("a@b.com"))
	assert.Equal(t, apperror.KindNotFound, kindOf(t, err))
	_, err = svc.Update(context.Background(), "bad", valid("a@b.com"))
	assert.Equal(t, apperror.KindValidation, kindOf(t, err))
}

func TestUpdateStatus_ActiveInactive(t *testing.T) {
	svc, repo := setup()
	u, _ := svc.Create(context.Background(), valid("a@b.com"))
	got, err := svc.UpdateStatus(context.Background(), u.ID.Hex(), model.StatusInactive)
	require.NoError(t, err)
	assert.Equal(t, model.StatusInactive, repo.Users[u.ID].Status)
	assert.Equal(t, model.StatusInactive, got.Status)
	_, err = svc.UpdateStatus(context.Background(), u.ID.Hex(), model.StatusActive)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, repo.Users[u.ID].Status)
}

func TestUpdateStatus_Invalid(t *testing.T) {
	svc, _ := setup()
	u, _ := svc.Create(context.Background(), valid("a@b.com"))
	_, err := svc.UpdateStatus(context.Background(), u.ID.Hex(), "BANNED")
	assert.Equal(t, apperror.KindValidation, kindOf(t, err))
}

func TestGet_MalformedID(t *testing.T) {
	svc, _ := setup()
	_, err := svc.Get(context.Background(), "xyz")
	assert.Equal(t, apperror.KindValidation, kindOf(t, err))
}

func TestGet_NotFound(t *testing.T) {
	svc, _ := setup()
	_, err := svc.Get(context.Background(), "5f8d0d55b54764421b7156c3")
	assert.Equal(t, apperror.KindNotFound, kindOf(t, err))
}

func TestGet_Found(t *testing.T) {
	svc, _ := setup()
	u, _ := svc.Create(context.Background(), valid("a@b.com"))
	got, err := svc.Get(context.Background(), u.ID.Hex())
	require.NoError(t, err)
	assert.Equal(t, "a@b.com", got.Email)
}

func TestDelete(t *testing.T) {
	svc, repo := setup()
	u, _ := svc.Create(context.Background(), valid("a@b.com"))
	require.NoError(t, svc.Delete(context.Background(), u.ID.Hex()))
	assert.Empty(t, repo.Users)
	assert.Equal(t, apperror.KindNotFound, kindOf(t, svc.Delete(context.Background(), u.ID.Hex())))
	assert.Equal(t, apperror.KindValidation, kindOf(t, svc.Delete(context.Background(), "bad")))
}

func TestList_PaginationSearchFilter(t *testing.T) {
	svc, _ := setup()
	for i := 0; i < 12; i++ {
		u, _ := svc.Create(context.Background(), valid(fmt.Sprintf("u%02d@x.com", i)))
		if i%2 == 0 {
			_, _ = svc.UpdateStatus(context.Background(), u.ID.Hex(), model.StatusInactive)
		}
	}
	users, meta, err := svc.List(context.Background(), model.ListQuery{Page: 2, PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.EqualValues(t, 12, meta.Total)
	assert.Equal(t, 2, meta.Page)

	users, meta, _ = svc.List(context.Background(), model.ListQuery{Status: model.StatusInactive})
	assert.Len(t, users, 6)
	assert.Equal(t, DefaultPageSize, meta.PageSize)

	users, _, _ = svc.List(context.Background(), model.ListQuery{Search: "u03"})
	assert.Len(t, users, 1)

	users, meta, _ = svc.List(context.Background(), model.ListQuery{Search: "zzz"})
	assert.Empty(t, users)
	assert.EqualValues(t, 0, meta.Total)

	_, _, err = svc.List(context.Background(), model.ListQuery{Status: "X"})
	assert.Equal(t, apperror.KindValidation, kindOf(t, err))

	_, meta, _ = svc.List(context.Background(), model.ListQuery{PageSize: 5000})
	assert.Equal(t, MaxPageSize, meta.PageSize)
}

func TestStats(t *testing.T) {
	svc, _ := setup()
	u, _ := svc.Create(context.Background(), valid("a@b.com"))
	_, _ = svc.Create(context.Background(), valid("c@d.com"))
	_, _ = svc.UpdateStatus(context.Background(), u.ID.Hex(), model.StatusInactive)
	s, err := svc.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, model.Stats{Total: 2, Active: 1, Inactive: 1}, s)
}

func TestDBUnavailable(t *testing.T) {
	svc, repo := setup()
	repo.Err = apperror.Unavailable()
	_, err := svc.Create(context.Background(), valid("a@b.com"))
	assert.Equal(t, apperror.KindUnavailable, kindOf(t, err))
	_, err = svc.Stats(context.Background())
	assert.Equal(t, apperror.KindUnavailable, kindOf(t, err))
}

func TestGet_UnexpectedError(t *testing.T) {
	svc, repo := setup()
	repo.Err = errors.New("boom db failure")
	_, err := svc.Get(context.Background(), "5f8d0d55b54764421b7156c3")
	require.Error(t, err)
	assert.Equal(t, apperror.KindInternal, apperror.As(err).Kind)
}

func TestGet_EmptyID(t *testing.T) {
	svc, _ := setup()
	_, err := svc.Get(context.Background(), "")
	assert.Equal(t, apperror.KindValidation, kindOf(t, err))
}
