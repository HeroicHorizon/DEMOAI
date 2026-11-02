package repository

import (
	"context"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/forgeflow/usermanagement/internal/apperror"
	"github.com/forgeflow/usermanagement/internal/model"
)

// UserRepository persists users. Errors are returned as *apperror.Error.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	List(ctx context.Context, q model.ListQuery) ([]model.User, int64, error)
	Update(ctx context.Context, id bson.ObjectID, in model.UserInput, now time.Time) (*model.User, error)
	UpdateStatus(ctx context.Context, id bson.ObjectID, status string, now time.Time) (*model.User, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	Stats(ctx context.Context) (model.Stats, error)
}

type MongoUserRepository struct {
	coll *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{coll: db.Collection("users")}
}

// EnsureIndexes creates the unique email index.
func (r *MongoUserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	})
	return err
}

func (r *MongoUserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	res, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		return nil, apperror.FromMongo(err)
	}
	u.ID = res.InsertedID.(bson.ObjectID)
	return u, nil
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	var u model.User
	if err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&u); err != nil {
		return nil, apperror.FromMongo(err)
	}
	return &u, nil
}

func buildFilter(q model.ListQuery) bson.D {
	filter := bson.D{}
	if q.Status != "" {
		filter = append(filter, bson.E{Key: "status", Value: q.Status})
	}
	if q.Search != "" {
		re := bson.Regex{Pattern: regexp.QuoteMeta(q.Search), Options: "i"}
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "firstName", Value: re}},
			bson.D{{Key: "lastName", Value: re}},
			bson.D{{Key: "email", Value: re}},
		}})
	}
	return filter
}

func (r *MongoUserRepository) List(ctx context.Context, q model.ListQuery) ([]model.User, int64, error) {
	filter := buildFilter(q)
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, apperror.FromMongo(err)
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64((q.Page - 1) * q.PageSize)).
		SetLimit(int64(q.PageSize))
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, apperror.FromMongo(err)
	}
	users := []model.User{}
	if err := cur.All(ctx, &users); err != nil {
		return nil, 0, apperror.FromMongo(err)
	}
	return users, total, nil
}

func (r *MongoUserRepository) updateOne(ctx context.Context, id bson.ObjectID, set bson.D) (*model.User, error) {
	var u model.User
	err := r.coll.FindOneAndUpdate(ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: set}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&u)
	if err != nil {
		return nil, apperror.FromMongo(err)
	}
	return &u, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, id bson.ObjectID, in model.UserInput, now time.Time) (*model.User, error) {
	return r.updateOne(ctx, id, bson.D{
		{Key: "firstName", Value: in.FirstName},
		{Key: "lastName", Value: in.LastName},
		{Key: "email", Value: in.Email},
		{Key: "phone", Value: in.Phone},
		{Key: "department", Value: in.Department},
		{Key: "designation", Value: in.Designation},
		{Key: "updatedAt", Value: now},
	})
}

func (r *MongoUserRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status string, now time.Time) (*model.User, error) {
	return r.updateOne(ctx, id, bson.D{
		{Key: "status", Value: status},
		{Key: "updatedAt", Value: now},
	})
}

func (r *MongoUserRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return apperror.FromMongo(err)
	}
	if res.DeletedCount == 0 {
		return apperror.NotFound("User not found.")
	}
	return nil
}

func (r *MongoUserRepository) Stats(ctx context.Context) (model.Stats, error) {
	var s model.Stats
	var err error
	if s.Total, err = r.coll.CountDocuments(ctx, bson.D{}); err != nil {
		return s, apperror.FromMongo(err)
	}
	if s.Active, err = r.coll.CountDocuments(ctx, bson.D{{Key: "status", Value: model.StatusActive}}); err != nil {
		return s, apperror.FromMongo(err)
	}
	if s.Inactive, err = r.coll.CountDocuments(ctx, bson.D{{Key: "status", Value: model.StatusInactive}}); err != nil {
		return s, apperror.FromMongo(err)
	}
	return s, nil
}
