package main

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
    "os"
    "reflect"
    "testing"
    "time"
)

type testDatabase struct {
    client *mongo.Client
    col    *mongo.Collection
    ctx    context.Context
}

func (t *testDatabase) cleanDatabase() error {
    if _, err := t.col.DeleteMany(t.ctx, bson.D{}); err != nil {
        return err
    }
    return nil
}

func (t *testDatabase) contextClientCollection() {
    var err error
    t.client, err = mongo.NewClient(options.Client().ApplyURI(
        os.Getenv("GOPHERCOINS_DB_URI")))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = t.client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    t.col = t.client.Database("coin_bank").Collection("test")
}

func Test_createAccount(t *testing.T) {
    tests := []struct {
        name string
        data bson.D
        want int
    }{
        {
            "single insert",
            bson.D{
                {"name", "foo"},
                {"amount", 0},
            },
            1,
        },
    }
    
    testDB := new(testDatabase)
    testDB.contextClientCollection()
    defer testDB.client.Disconnect(testDB.ctx)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := testDB.cleanDatabase()
            if err != nil {
                t.Errorf("unable to sanitize test db: %s", err)
            }
            createAccount(testDB.ctx, testDB.col, tt.data)
            if cur, _ := testDB.col.Find(testDB.ctx, bson.D{}); cur.RemainingBatchLength() != tt.want {
                t.Logf("inequal number of documents inserted got %d. want %d", cur.RemainingBatchLength(), tt.want)
                t.Fail()
            }
    
            _ = testDB.cleanDatabase()
        })
    }
}

func Test_readAccount(t *testing.T) {
    tests := []struct {
        name             string
        filter           bson.D
        existingAccounts bson.D
    }{
        {
            "found foo account",
            bson.D{{"name", "foo"}},
            bson.D{
                {"name", "foo"},
                {"amount", 0},
            },
        }, {
            "found bar account",
            bson.D{{}},
            bson.D{
                {"name", "bar"},
                {"amount", 5},
            },
        },
    }
    
    testDB := new(testDatabase)
    testDB.contextClientCollection()
    defer testDB.client.Disconnect(testDB.ctx)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if err := testDB.cleanDatabase(); err != nil {
                t.Errorf("unable to sanitize database")
            }
            
            result, err := testDB.col.InsertOne(testDB.ctx, tt.existingAccounts) // Insert what we're trying to read
            if err != nil {
                t.Fatalf("unable to insert document for reading: %v", err)
            }
            
            gotResult := readAccount(testDB.ctx, testDB.col, tt.filter)
            if !reflect.DeepEqual(gotResult["_id"], result.InsertedID) {
                t.Errorf("readAccount() = %v, want %v", gotResult, result.InsertedID)
            }
            
            _ = testDB.cleanDatabase()
        })
    }
}

func Test_updateAccount(t *testing.T) {
    tests := []struct {
        name             string
        existingAccounts bson.D
        id               primitive.ObjectID
        update           bson.D
        wantAmount       int32
    }{
        {
            "increase",
            bson.D{
                {"name", "foo"},
                {"amount", 0},
            },
            primitive.ObjectID{},
            bson.D{
                {"$inc", bson.D{
                    {"amount", 1},
                }},
            },
            1,
        },
    }
    testDB := new(testDatabase)
    testDB.contextClientCollection()
    defer testDB.client.Disconnect(testDB.ctx)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var account bson.M
            
            // Setup account to update
            if err := testDB.cleanDatabase(); err != nil {
                t.Errorf("unable to sanitize database")
            }
            result, err := testDB.col.InsertOne(testDB.ctx, tt.existingAccounts) // Insert what we're trying to read
            if err != nil {
                t.Fatalf("unable to insert document for reading: %v", err)
            }
            id := result.InsertedID.(primitive.ObjectID)
            
            // Update
            updateAccount(testDB.ctx, testDB.col, id, tt.update)
            
            // Check account updated
            _ = testDB.col.FindOne(testDB.ctx, bson.D{{"_id", result.InsertedID}}).Decode(&account)
            if !reflect.DeepEqual(account["amount"], tt.wantAmount) {
                t.Errorf("account did not update amount.\tgot %v, want %v", account["amount"], tt.wantAmount)
            }
    
            _ = testDB.cleanDatabase()
        })
    }
}
