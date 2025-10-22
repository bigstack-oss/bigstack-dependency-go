package mongo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	CreateRecordIfNotExist = options.Update().SetUpsert(true)

	helper *Helper
	once   sync.Once
)

type Client interface {
	Database(string, ...*options.DatabaseOptions) *mongo.Database
	StartSession(...*options.SessionOptions) (mongo.Session, error)
	Disconnect(context.Context) error
}

type DBClient interface {
	Collection(string, ...*options.CollectionOptions) *mongo.Collection
	ListCollectionNames(context.Context, any, ...*options.ListCollectionsOptions) ([]string, error)
}

type TxnClient interface {
	WithTransaction(context.Context, func(mongo.SessionContext) (any, error), ...*options.TransactionOptions) (any, error)
	EndSession(context.Context)
}

type CollClient interface {
	Aggregate(context.Context, any, ...*options.AggregateOptions) (*mongo.Cursor, error)
	BulkWrite(context.Context, []mongo.WriteModel, ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error)
	Find(context.Context, any, ...*options.FindOptions) (*mongo.Cursor, error)
	FindOne(context.Context, any, ...*options.FindOneOptions) *mongo.SingleResult
	FindOneAndUpdate(context.Context, any, any, ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	FindOneAndDelete(context.Context, any, ...*options.FindOneAndDeleteOptions) *mongo.SingleResult
	InsertOne(context.Context, any, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	InsertMany(context.Context, []any, ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)
	DeleteOne(context.Context, any, ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteMany(context.Context, any, ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	UpdateOne(context.Context, any, any, ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(context.Context, any, any, ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	CountDocuments(context.Context, any, ...*options.CountOptions) (int64, error)
	Indexes() mongo.IndexView
}

type GridFSClient interface {
	Find(filter any, opts ...*options.GridFSFindOptions) (*mongo.Cursor, error)
	Delete(fileID any) error
	OpenUploadStream(filename string, opts ...*options.UploadOptions) (*gridfs.UploadStream, error)
	DownloadToStream(fileID any, stream io.Writer) (int64, error)
}

type CursorClient interface {
	All(context.Context, any) error
	Next(context.Context) bool
}

type Helper struct {
	Client
	Options
}

func initOptions(opts []Option) *Options {
	options := &Options{Auth: Auth{Enable: true}}
	for _, o := range opts {
		o(options)
	}

	return options
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	h := &Helper{Options: *initedOpts}

	err := h.SetMongoClient()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func NewGlobalHelper(opts ...Option) error {
	var err error
	once.Do(func() {
		helper, err = NewHelper(opts...)
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) SetMongoClient() error {
	opt := options.Client()
	opt.ApplyURI(h.Uri)

	if h.Auth.Enable {
		opt.Auth = &options.Credential{
			AuthSource: h.Auth.Source,
			Username:   h.Auth.Username,
			Password:   h.Auth.Password,
		}
	}

	if h.ReplicaSet != "" {
		opt.ReplicaSet = &h.ReplicaSet
	}

	mongoCli, err := mongo.Connect(context.Background(), opt)
	if err != nil {
		log.Errorf("err of connect mongo: %s", err.Error())
		return err
	}

	h.Client = mongoCli
	return nil
}

func GetGlobalHelper() *Helper {
	return helper
}

func (h *Helper) NewDBCli(db string) (DBClient, error) {
	if db == "" {
		return nil, fmt.Errorf(
			"db is nil. value: db(%s)",
			db,
		)
	}

	return h.Client.Database(db), nil
}

func (h *Helper) NewCollCli(db, coll string) (CollClient, error) {
	if db == "" || coll == "" {
		return nil, fmt.Errorf(
			"db or coll is nil or both are nil. values: db(%s); coll(%s)",
			db,
			coll,
		)
	}

	dbCli := h.Client.Database(db)
	return dbCli.Collection(coll), nil
}

func (h *Helper) NewTxnCli() (TxnClient, error) {
	s, err := h.Client.StartSession()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (h *Helper) NewGridFSCli(db string) (GridFSClient, error) {
	b, err := gridfs.NewBucket(h.Client.Database(db))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (h *Helper) GetQueryCursor(db, coll string, query bson.M, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return nil, err
	}

	cursor, err := c.Find(context.Background(), query, opts...)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

func (h *Helper) Get(db, coll string, filter bson.M) (*mongo.SingleResult, error) {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	result := c.FindOne(ctx, filter)
	return result, nil
}

func (h *Helper) GetCount(db, coll string, filter bson.M) (int64, error) {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	count, err := c.CountDocuments(ctx, filter)
	defer cancel()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (h *Helper) Insert(db, coll string, data any) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	_, err = c.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) InsertMany(db, coll string, data []any) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	_, err = c.InsertMany(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) UpdateOne(db, coll string, filter any, data any, opts ...*options.UpdateOptions) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	_, err = c.UpdateOne(ctx, filter, data, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) UpdateMany(db, coll string, filter any, data any) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	_, err = c.UpdateMany(ctx, filter, data)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) DeleteOne(db, coll string, filter any) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	_, err = c.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) DeleteAll(db, coll string, filter any) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	_, err = c.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) Aggregate(db, coll string, data any, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	return c.Aggregate(ctx, data, opts...)
}

func (h *Helper) BulkWrite(db, coll string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	_, err = c.BulkWrite(ctx, models, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) UploadFile(db, filename string, file multipart.File, opts ...*options.UploadOptions) error {
	c, err := h.NewGridFSCli(db)
	if err != nil {
		return nil
	}

	uploadStream, err := c.OpenUploadStream(filename, opts...)
	if err != nil {
		return nil
	}
	defer uploadStream.Close()

	_, err = io.Copy(uploadStream, file)
	if err != nil {
		return nil
	}

	return nil
}

func (h *Helper) FindFile(db string, filter any, opts ...*options.GridFSFindOptions) ([]gridfs.File, error) {
	c, err := h.NewGridFSCli(db)
	if err != nil {
		return nil, err
	}

	cursor, err := c.Find(filter, opts...)
	if err != nil {
		return nil, err
	}

	var files []gridfs.File
	if err = cursor.All(context.Background(), &files); err != nil {
		return nil, err
	}

	return files, nil
}

func (h *Helper) DownloadFile(db string, fileID any) (bytes.Buffer, error) {
	c, err := h.NewGridFSCli(db)
	if err != nil {
		return bytes.Buffer{}, err
	}

	var buf bytes.Buffer
	_, err = c.DownloadToStream(fileID, &buf)
	if err != nil {
		return bytes.Buffer{}, err
	}

	return buf, nil
}

func (h *Helper) DeleteFile(db string, fileID any) error {
	c, err := h.NewGridFSCli(db)
	if err != nil {
		return err
	}

	return c.Delete(fileID)
}

func (h *Helper) GetAllCollections(db string) ([]string, error) {
	dbCli, err := h.NewDBCli(db)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	collections, err := dbCli.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	return collections, nil
}

func (h *Helper) CreateExpirationIndex(db, coll string, keys bson.D, seconds int32) error {
	c, err := h.NewCollCli(db, coll)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	indexModel := mongo.IndexModel{Keys: keys, Options: options.Index().SetExpireAfterSeconds(seconds)}
	_, err = c.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) Close() {
	if h.Client == nil {
		return
	}

	err := h.Client.Disconnect(context.Background())
	if err != nil {
		log.Errorf("failed to close mongo connection: %s", err.Error())
	}
}
