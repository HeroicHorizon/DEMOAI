// Package testutil provides an in-memory UserRepository test double.
package testutil

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/forgeflow/usermanagement/internal/apperror"
	"github.com/forgeflow/usermanagement/internal/model"
)

type FakeRepo struct {
	mu    sync.Mutex
	Users map[bson.ObjectID]model.User
	// Err, when set, is returned by every method (simulates a DB outage).
	Err error
	// LastCtx records the context received by the latest call.
	LastCtx context.Context
}

func NewFakeRepo() *FakeRepo { return &FakeRepo{Users: map[bson.ObjectID]model.User{}} }

func (f *FakeRepo) enter(ctx context.Context) error {
	f.LastCtx = ctx
	if f.Err != nil {
		return f.Err
	}
	return ctx.Err()
}

func (f *FakeRepo) emailTaken(email string, except bson.ObjectID) bool {
	for id, u := range f.Users {
		if u.Email == email && id != except {
			return true
		}
	}
	return false
}

func (f *FakeRepo) Create(ctx context.Context, u *model.User) (*model.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.enter(ctx); err != nil {
		return nil, err
	}
	if f.emailTaken(u.Email, bson.NilObjectID) {
		return nil, apperror.Conflict("A user with this email already exists.")
	}
	u.ID = bson.NewObjectID()
	f.Users[u.ID] = *u
	return u, nil
}

func (f *FakeRepo) GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.enter(ctx); err != nil {
		return nil, err
	}
	u, ok := f.Users[id]
	if !ok {
		return nil, apperror.NotFound("User not found.")
	}
	return &u, nil
}

func (f *FakeRepo) List(ctx context.Context, q model.ListQuery) ([]model.User, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.enter(ctx); err != nil {
		return nil, 0, err
	}
	all := []model.User{}
	s := strings.ToLower(q.Search)
	for _, u := range f.Users {
		if q.Status != "" && u.Status != q.Status {
			continue
		}
		if s != "" && !strings.Contains(strings.ToLower(u.FirstName+"\n"+u.LastName+"\n"+u.Email), s) {
			continue
		}
		all = append(all, u)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Email < all[j].Email })
	total := int64(len(all))
	start := (q.Page - 1) * q.PageSize
	if start > len(all) {
		start = len(all)
	}
	end := start + q.PageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (f *FakeRepo) Update(ctx context.Context, id bson.ObjectID, in model.UserInput, now time.Time) (*model.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.enter(ctx); err != nil {
		return nil, err
	}
	u, ok := f.Users[id]
	if !ok {
		return nil, apperror.NotFound("User not found.")
	}
	if f.emailTaken(in.Email, id) {
		return nil, apperror.Conflict("A user with this email already exists.")
	}
	u.FirstName, u.LastName, u.Email, u.Phone = in.FirstName, in.LastName, in.Email, in.Phone
	u.Department, u.Designation, u.UpdatedAt = in.Department, in.Designation, now
	f.Users[id] = u
	return &u, nil
}

func (f *FakeRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, status string, now time.Time) (*model.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.enter(ctx); err != nil {
		return nil, err
	}
	u, ok := f.Users[id]
	if !ok {
		return nil, apperror.NotFound("User not found.")
	}
	u.Status, u.UpdatedAt = status, now
	f.Users[id] = u
	return &u, nil
}

func (f *FakeRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.enter(ctx); err != nil {
		return err
	}
	if _, ok := f.Users[id]; !ok {
		return apperror.NotFound("User not found.")
	}
	delete(f.Users, id)
	return nil
}

func (f *FakeRepo) Stats(ctx context.Context) (model.Stats, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var s model.Stats
	if err := f.enter(ctx); err != nil {
		return s, err
	}
	for _, u := range f.Users {
		s.Total++
		if u.Status == model.StatusActive {
			s.Active++
		} else {
			s.Inactive++
		}
	}
	return s, nil
}
