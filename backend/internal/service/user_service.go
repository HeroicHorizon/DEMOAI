package service

import (
	"context"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/forgeflow/usermanagement/internal/apperror"
	"github.com/forgeflow/usermanagement/internal/model"
	"github.com/forgeflow/usermanagement/internal/repository"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
	maxNameLen      = 100
	maxEmailLen     = 254
	maxFieldLen     = 100
)

var phoneRe = regexp.MustCompile(`^\+?[0-9][0-9 \-]{5,19}$`)

type UserService struct {
	repo repository.UserRepository
	now  func() time.Time
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

// ParseID validates a user ID before any database access.
func ParseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, apperror.Validation("Invalid user ID.", nil)
	}
	return oid, nil
}

func normalize(in model.UserInput) (model.UserInput, error) {
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Phone = strings.TrimSpace(in.Phone)
	in.Department = strings.TrimSpace(in.Department)
	in.Designation = strings.TrimSpace(in.Designation)

	d := map[string]string{}
	required := func(field, val, label string, max int) {
		if val == "" {
			d[field] = label + " is required."
		} else if utf8.RuneCountInString(val) > max {
			d[field] = label + " is too long."
		}
	}
	required("firstName", in.FirstName, "First name", maxNameLen)
	required("lastName", in.LastName, "Last name", maxNameLen)
	required("email", in.Email, "Email", maxEmailLen)
	if _, bad := d["email"]; !bad {
		addr, err := mail.ParseAddress(in.Email)
		if err != nil || addr.Address != in.Email || !strings.Contains(in.Email[strings.LastIndex(in.Email, "@"):], ".") {
			d["email"] = "Email address is invalid."
		}
	}
	if in.Phone != "" && !phoneRe.MatchString(in.Phone) {
		d["phone"] = "Phone number is invalid."
	}
	if utf8.RuneCountInString(in.Department) > maxFieldLen {
		d["department"] = "Department is too long."
	}
	if utf8.RuneCountInString(in.Designation) > maxFieldLen {
		d["designation"] = "Designation is too long."
	}
	if len(d) > 0 {
		return in, apperror.Validation("Validation failed.", d)
	}
	return in, nil
}

func (s *UserService) Create(ctx context.Context, in model.UserInput) (*model.User, error) {
	in, err := normalize(in)
	if err != nil {
		return nil, err
	}
	now := s.now()
	return s.repo.Create(ctx, &model.User{
		FirstName: in.FirstName, LastName: in.LastName, Email: in.Email, Phone: in.Phone,
		Department: in.Department, Designation: in.Designation,
		Status: model.StatusActive, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *UserService) Get(ctx context.Context, id string) (*model.User, error) {
	oid, err := ParseID(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, oid)
}

func (s *UserService) Update(ctx context.Context, id string, in model.UserInput) (*model.User, error) {
	oid, err := ParseID(id)
	if err != nil {
		return nil, err
	}
	in, err = normalize(in)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, oid, in, s.now())
}

func (s *UserService) UpdateStatus(ctx context.Context, id, status string) (*model.User, error) {
	oid, err := ParseID(id)
	if err != nil {
		return nil, err
	}
	if status != model.StatusActive && status != model.StatusInactive {
		return nil, apperror.Validation("Status must be ACTIVE or INACTIVE.", map[string]string{"status": "Unsupported status."})
	}
	return s.repo.UpdateStatus(ctx, oid, status, s.now())
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	oid, err := ParseID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, oid)
}

// List validates and normalizes query parameters, then queries the repository.
func (s *UserService) List(ctx context.Context, q model.ListQuery) ([]model.User, model.Meta, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}
	q.Search = strings.TrimSpace(q.Search)
	if utf8.RuneCountInString(q.Search) > maxEmailLen {
		return nil, model.Meta{}, apperror.Validation("Search term is too long.", nil)
	}
	if q.Status != "" && q.Status != model.StatusActive && q.Status != model.StatusInactive {
		return nil, model.Meta{}, apperror.Validation("Status must be ACTIVE or INACTIVE.", map[string]string{"status": "Unsupported status."})
	}
	users, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, model.Meta{}, err
	}
	return users, model.Meta{Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *UserService) Stats(ctx context.Context) (model.Stats, error) {
	return s.repo.Stats(ctx)
}
