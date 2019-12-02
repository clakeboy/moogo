package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

//dsn mongodb://root:WiaQ82n7B3L5Cz*2#10m@172.18.76.150:27017?authSource=admin
type Config struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Auth     string `json:"auth" yaml:"auth"`
	DBName   string `json:"db_name" yaml:"db_name"`
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
}

//mongodb orm use official driver
type Database struct {
	client        *mongo.Client
	currentDBName string
}

func NewDatabase(conf *Config) (*Database, error) {
	opts := options.Client()
	opts.SetHosts([]string{fmt.Sprintf("%s:%s", conf.Host, conf.Port)})
	opts.SetMaxPoolSize(uint64(conf.PoolSize))
	opts.SetConnectTimeout(20 * time.Second)
	opts.SetCompressors([]string{"snappy", "zlib"})
	opts.SetZlibLevel(7)
	if conf.Auth != "" {
		opts.SetAuth(options.Credential{
			AuthMechanism: "SCRAM-SHA-1",
			AuthSource:    conf.Auth,
			Username:      conf.User,
			Password:      conf.Password,
			PasswordSet:   true,
		})
	}

	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}

	return &Database{
		client:        client,
		currentDBName: conf.DBName,
	}, nil
}

//connect to mongodb
func (d *Database) Open() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err := d.client.Connect(ctx)
	if err != nil {
		return err
	}

	return d.Ping()
}

//disconnect mongodb
func (d *Database) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	return d.client.Disconnect(ctx)
}

//select database
func (d *Database) SelectDatabase(dbName string) *Database {
	d.currentDBName = dbName
	return d
}

func (d *Database) ListDatabase() (mongo.ListDatabasesResult, error) {
	return d.client.ListDatabases(nil, bson.M{})
}

func (d *Database) ListDatabaseNames() ([]string, error) {
	return d.client.ListDatabaseNames(nil, bson.M{})
}

//get default database
func (d *Database) Database(dbName ...string) *mongo.Database {
	if len(dbName) > 0 {
		return d.client.Database(dbName[0])
	}

	return d.client.Database(d.currentDBName)
}

//return high level collection
func (d *Database) Collection(collName string, dbName ...string) *Collection {
	coll := d.Database(dbName...).Collection(collName)
	return NewCollection(collName, coll)
}

//ping mongodb server
func (d *Database) Ping() error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return d.client.Ping(ctx, readpref.Primary())
}
