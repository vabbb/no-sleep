package main

import (
	"context"
	"strconv"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// change them if u want
	url         = "mongodb://localhost:27017"
	dbName      = "my_db"
	connections = "connections"
	flows       = "flows"
)

// var for mongoDB
var (
	client          *mongo.Client
	collConnections *mongo.Collection
	collFlows       *mongo.Collection
)

// type connectionDocT struct {
// 	_id              string
// 	srcIP            string
//     dstIP            string
// 	srcPort          uint16
// 	dstPort          uint16
//     lastSeen         int64 // updated with the latest connt.end uploaded
//     favorite         bool  // defaults to false, can only be changed from the front-end
// 	flows 			 []string
// }

// type mongoconntype struct {
// 	_id            string
//     connID         string
//     src            string // "IP:port"
//     dst            string // "IP:port"
//     time           int64 // this is the flow's start time
//     favorite       bool
//     hasSYN, hasFIN bool
//     size           int64
//     data           string // printable representation of the data
//     hex            []byte // hex representation of the data
// }

func connectDB(url string) {
	resClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(url))

	if err != nil {
		log.Fatal(err)
	} else {
		log.Infoln("Connected to mongodb server!")
	}

	client = resClient
}

func getCollectionsFromDB(client *mongo.Client, dbName string, collName string) {
	resCollection := client.Database(dbName).Collection(collName)
	if collName == connections {
		collConnections = resCollection

	} else {
		collFlows = resCollection
	}
}

func insertconntDoc(connt *connt) {

	addNewConnection(connt)
	var err error
	filter := bson.M{"_id": connt.connID}
	// connDoc := collConnections.FindOne(context.TODO(), filter)

	// var connDocDecoded bson.M
	// err = connDoc.Decode(&connDocDecoded)

	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("Found", connDocDecoded["_id"])
	// }

	insertResult, err := collFlows.InsertOne(context.TODO(), bson.M{
		"_id":  connt.connID,
		"src":  connt.srcIP + ":" + strconv.Itoa(int(connt.srcPort)),
		"dst":  connt.dstIP + ":" + strconv.Itoa(int(connt.dstPort)),
		"time": connt.srcPort,
		// "data":   connt.data,
		// "hex":    connt.hex,
	})

	if err != nil {
		log.Infoln("Error in insert connt", connt.connID)
		return
	}

	log.Debugln("Inserting connt doc", insertResult.InsertedID, "...")

	update := bson.M{"$push": bson.M{"flows": connt.connID}, "$set": bson.M{"lastSeen": connt.end}}
	_, err = collConnections.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		log.Infoln(err)
	}

	log.Infoln("Inserted connt", connt.connID, "with lastSeen", connt.end, "at flows array in connection doc", connt.connID)
}

func addNewConnection(connt *connt) {

	insertResult, err := collConnections.InsertOne(context.TODO(), bson.M{
		"_id": connt.connID,
		"endpoints": bson.A{
			bson.A{connt.srcIP, connt.srcPort},
			bson.A{connt.dstIP, connt.dstPort},
		},
		"lastSeen": connt.end,
		"favorite": false,
		"flows":    bson.A{}})

	if err != nil {
		log.Debugln("Object ", connt.connID, " already exists")
		return
	}

	log.Infoln("Inserted connection doc", insertResult.InsertedID)

}
