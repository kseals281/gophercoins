package main

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "log"
)

type Coin int32

const (
    HardChore = 50
)

type Account struct {
    ID     primitive.ObjectID `bson:"_id,omitempty"`
    Name   string             `bson:"name,omitempty"`
    Amount float64            `bson:"amount,omitempty"`
}

func createAccount(ctx context.Context, col *mongo.Collection, data interface{}) {
    _, err := col.InsertOne(ctx, data)
    if err != nil {
        log.Fatal(err)
    }
}

func readAccount(ctx context.Context, col *mongo.Collection, filter interface{}) (result bson.M) {
    if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
        log.Printf("unable to find account: %v", err)
    }
    return result
}

func updateAccount(ctx context.Context, col *mongo.Collection, id primitive.ObjectID, update interface{}) {
    _, err := col.UpdateByID(ctx, id, update)
    if err != nil {
        log.Printf("unable to update account: %v", err)
    }
}
