package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Episode represents the schema for the "Episodes" collection
type Episode struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Podcast     primitive.ObjectID `bson:"podcast,omitempty"`
	Title       string             `bson:"title,omitempty"`
	Description string             `bson:"description,omitempty"`
	Duration    int                `bson:"duration,omitempty"`
}

// const uri = `mongodb://localhost:27018,localhost:27019,localhost:27020/?replicaSet=myReplicaSet`
const uri = `mongodb://mongo1:27017,mongo2:27017,mongo3:27017/?replicaSet=myReplicaSet`

var Hostname, _ = os.Hostname()

func main() {

	e := echo.New()
	e.Debug = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})
	e.GET("/error", func(c echo.Context) error {
		return errors.New("nothing error")
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	e.GET("/db", func(c echo.Context) error {
		return DB(c)
	})
	e.GET("/db2", func(c echo.Context) error {
		return DB2(c)
	})

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8090"
	}
	

	e.Logger.Fatal(e.Start(":" + httpPort))
}

func DB(c echo.Context) error {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	defer client.Disconnect(context.TODO())

	database := client.Database("quickstart")
	episodesCollection := database.Collection("episodes", &options.CollectionOptions{})
	fmt.Printf("episodesCollection: %v\n", episodesCollection)

	// database.RunCommand(context.TODO(), bson.D{{"create", "episodes"}})

	//////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////

	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	errSession := mongo.WithSession(context.Background(), session,
		func(sessionContext mongo.SessionContext) error {

			if err = session.StartTransaction(txnOpts); err != nil {
				return err
			}
			result, err := episodesCollection.InsertOne(
				sessionContext,
				Episode{
					Title:    "A Transaction Episode for the Ages",
					Duration: 15,
				},
			)
			if err != nil {
				return err
			}
			fmt.Println(result.InsertedID)
			v := rand.Intn(5)
			fmt.Printf("v: %v\n", v)
			result, err = episodesCollection.InsertOne(
				sessionContext,
				Episode{
					Title:    "Transactions for All " + Hostname,
					Duration: v,
				},
			)
			if err != nil {

				return err
			}
			fmt.Println(result.InsertedID)
			if err = session.CommitTransaction(sessionContext); err != nil {
				fmt.Printf("err: %v\n", err)
				return err
			}
			return nil
		})

	if errSession != nil {
		if abortErr := session.AbortTransaction(context.Background()); abortErr != nil {
			fmt.Printf("abortErr: %v\n", abortErr)
			return abortErr
		}
		fmt.Printf("errSession: %v\n", errSession)
		return errSession
	}
	return c.HTML(http.StatusOK, "no error")

}

func DB2(c echo.Context) error {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	defer client.Disconnect(context.TODO())

	database := client.Database("quickstart")
	episodesCollection := database.Collection("episodes", &options.CollectionOptions{})
	fmt.Printf("episodesCollection: %v\n", episodesCollection)

	// database.RunCommand(context.TODO(), bson.D{{"create", "episodes"}})

	//////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////

	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	errSession := mongo.WithSession(context.Background(), session,
		func(sessionContext mongo.SessionContext) error {

			if err = session.StartTransaction(txnOpts); err != nil {
				return err
			}
			result, err := episodesCollection.InsertOne(
				sessionContext,
				Episode{
					Title:    "A Transaction Episode for the Ages",
					Duration: 15,
				},
			) 
			if err != nil {
				return err
			}
			fmt.Println(result.InsertedID)
			if err = session.CommitTransaction(sessionContext); err != nil {
				fmt.Printf("err: %v\n", err)
				return err
			}
			// cannot call abortTransaction after calling commitTransaction
			return errors.New("wanted error for failing transction ")
		})

	if errSession != nil {
		if abortErr := session.AbortTransaction(context.Background()); abortErr != nil {
			fmt.Printf("abortErr: %v\n", abortErr)
			return abortErr
		}
		fmt.Printf("errSession: %v\n", errSession)
		return errSession
	}
	return c.HTML(http.StatusOK, "no error")

}
