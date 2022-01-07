package db

import (
	"bytes"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"searchproxy/fram/config"
	"searchproxy/fram/utils"
	"sync"
)

type MongoConfig struct {
	Uri    string `json:"uri"`
	Dbname string `json:"dbname"`
}

type Models struct {
	database *mongo.Database
	config   *MongoConfig
}

var db Models
var dance sync.Once

func MongoInstance() *Models {
	dance.Do(func() {
		db = Models{}
		var mg MongoConfig
		config.Install().Get("mongo",&mg)
		db.config = &mg
		clientOptions := options.Client().ApplyURI(db.config.Uri)
		client, err := mongo.Connect(context.Background(), clientOptions)
		utils.FatalAssert(err)
		err = client.Ping(context.Background(), nil)
		utils.FatalAssert(err)
		db.database = client.Database(db.config.Dbname)
	})
	return &db
}

func NewMongo(cfg *MongoConfig) *Models  {
	db = Models{}
	db.config = cfg
	clientOptions := options.Client().ApplyURI(db.config.Uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	utils.FatalAssert(err)
	err = client.Ping(context.Background(), nil)
	utils.FatalAssert(err)
	db.database = client.Database(db.config.Dbname)
	return &db
}

func (m Models) FindOne(table string, filter interface{}, result interface{} ) error {
	//var result models.User
	err := m.database.Collection(table).FindOne(context.Background(), filter).Decode(result)
	if err != nil {
		return err
	}
	return nil
}

func (m Models) GetDBNames() []string {
	listc, _ := m.database.ListCollectionNames(context.Background(), bson.M{})
	return listc
}

func (m Models) Aggregate(table string, pipeline interface{}, result *[]interface{}) error {
	cur, err := m.database.Collection(table).Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	for cur.Next(context.Background()) {
		elem := make(map[string]interface{})
		err := cur.Decode(elem)
		if err != nil {
			return err
		}
		*result = append(*result, elem)
	}
	return nil
}

func (m Models) FindSort(table string, filter interface{}, result *[]interface{}, sort interface{} ) error {
	findOptions := options.Find()
	findOptions.SetSort(sort)
	//findOptions.SetLimit(limit)
	//findOptions.SetSkip(skip)
	cur, err := m.database.Collection(table).Find(context.Background(), filter, findOptions)
	//cur, err := m.database.Collection(table).Find(context.Background(), filter)
	if err != nil {
		return err
	}
	for cur.Next(context.Background()) {
		elem := make(map[string]interface{})
		err := cur.Decode(elem)
		if err != nil {
			return err
		}
		*result = append(*result, elem)
	}
	return nil
}

func (m Models) FindMany(table string, filter interface{}, result *[]interface{} ) error {
	cur, err := m.database.Collection(table).Find(context.Background(), filter)
	if err != nil {
		return err
	}
	for cur.Next(context.Background()) {
		elem := make(map[string]interface{})
		err := cur.Decode(elem)
		if err != nil {
			return err
		}
		*result = append(*result, elem)
	}
	return nil
}

func (m Models) FindManyLimit(table string, filter interface{}, result *[]interface{} , limit, skip int64) error {
	findOptions := options.Find()
	findOptions.SetLimit(limit)
	findOptions.SetSkip(skip)
	cur, err := m.database.Collection(table).Find(context.Background(), filter, findOptions)
	if err != nil {
		return err
	}
	for cur.Next(context.Background()) {
		elem := make(map[string]interface{})
		err := cur.Decode(elem)
		if err != nil {
			return err
		}
		*result = append(*result, elem)
	}
	return nil
}

func (m Models) FindManySortLimit(table string, filter interface{}, sort interface{}, result *[]interface{}, limit, skip int64) error {
	findOptions := options.Find()
	findOptions.SetSort(sort)
	findOptions.SetLimit(limit)
	findOptions.SetSkip(skip)
	cur, err := m.database.Collection(table).Find(context.Background(), filter, findOptions)
	if err != nil {
		return err
	}
	for cur.Next(context.Background()) {
		elem := make(map[string]interface{})
		err := cur.Decode(elem)
		if err != nil {
			return err
		}
		*result = append(*result, elem)
	}
	return nil
}

func (m Models) InsertOne(table string, data interface{}) (*mongo.InsertOneResult, error) {
	id, err := m.database.Collection(table).InsertOne(context.Background(), data)
	//id.InsertedID
	if err != nil {
		return nil, err
	}
	return id, nil
}

func (m Models) InsertMany(table string, data *[]interface{}) error {
	_, err := m.database.Collection(table).InsertMany(context.Background(), *data)
	if err != nil {
		return err
	}
	return nil
}

func (m Models) UpdateOne(table string, filter interface{}, data interface{}) (*mongo.UpdateResult, error) {

	_id, err := m.database.Collection(table).UpdateOne(context.Background(), filter, data)
	if err != nil {
		return _id, err
	}
	return _id, nil
}

func (m Models) SaveOne(table string, filter interface{}, data interface{}) {
	_, _ = m.database.Collection(table).ReplaceOne(context.Background(), filter, data)
}

func (m Models) UpdateMany(table string, filter interface{}, data interface{}) error {
	_, err := m.database.Collection(table).UpdateMany(context.Background(), filter, data)
	if err != nil {
		return err
	}
	return nil
}

func (m Models) DeleteMany(table string, filter interface{}) error {
	_, err := m.database.Collection(table).DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}


func (m Models) getGridfsBucket(collName string) *gridfs.Bucket {
	var bucket *gridfs.Bucket
	// 使用默认文件集合名称
	if collName == "" || collName == options.DefaultName {
		bucket, _ = gridfs.NewBucket(m.database)
	} else {
		// 使用传入的文件集合名称
		bucketOptions := options.GridFSBucket().SetName(collName)
		bucket, _ = gridfs.NewBucket(m.database, bucketOptions)
	}
	return bucket
}

// 上传文件
// collName:文件集合名称 fileID:文件ID，必须唯一，否则会覆盖
// fileName:文件名称 fileContent:文件内容
func (m Models) GridfsUploadWithID(collName, fileID, fileName string, fileContent []byte) error {
	bucket := m.getGridfsBucket(collName)
	err := bucket.UploadFromStreamWithID(fileID, fileName, bytes.NewBuffer(fileContent))
	if err != nil {
		return err
	}
	return nil
}

// 下载文件
// 返回文件内容
func (m Models) GridfsDownload(collName, fileID string) (fileContent []byte, err error) {
	bucket := m.getGridfsBucket(collName)
	fileBuffer := bytes.NewBuffer(nil)
	if _, err = bucket.DownloadToStream(fileID, fileBuffer); err != nil {
		panic(err)
		return nil, err
	}
	return fileBuffer.Bytes(), nil
}

// 删除文件
func (m Models) GridfsDelete(collName, fileID string) error {
	bucket := m.getGridfsBucket(collName)
	if err := bucket.Delete(fileID); err != nil && err != gridfs.ErrFileNotFound {
		panic(err)
		return err
	}
	return nil
}

func (m Models) CountDocuments(table string, filter interface{},) (int64,error)  {
	return m.database.Collection(table).CountDocuments(context.TODO(), filter)
}