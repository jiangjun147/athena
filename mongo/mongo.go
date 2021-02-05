package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rickone/athena/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
	once   = sync.Once{}
)

func DB(name string) *mongo.Database {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		addr := config.GetString("mongodb", "address") // TODO: auth

		var err error
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", addr)))
		if err != nil {
			panic(err)
		}
	})

	return client.Database(name)
}
