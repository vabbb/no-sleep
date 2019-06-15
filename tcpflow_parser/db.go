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
	nodes       = "nodes"
)

// var for mongoDB
var (
	client          *mongo.Client
	collConnections *mongo.Collection
	collNodes       *mongo.Collection
)

// type connectionDocT struct {
// 	_id              string
// 	srcIP            string
//     dstIP            string
// 	srcPort          uint16
// 	dstPort          uint16
//     lastSeen         int64 // updated with the latest flowt.end uploaded
//     favorite         bool  // defaults to false, can only be changed from the front-end
// 	flows 			 []string
// }

// type mongoFlowType struct {
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
		collNodes = resCollection
	}
}

func insertNodetDoc(nodet *nodet) {

	addNewConnection(nodet)
	var err error
	filter := bson.M{"_id": nodet.connID}
	// connDoc := collConnections.FindOne(context.TODO(), filter)

	// var connDocDecoded bson.M
	// err = connDoc.Decode(&connDocDecoded)

	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("Found", connDocDecoded["_id"])
	// }

	insertResult, err := collNodes.InsertOne(context.TODO(), bson.M{
		"_id":    nodet.nodeID,
		"connID": nodet.connID,
		"src":    nodet.srcIP + ":" + strconv.Itoa(int(nodet.srcPort)),
		"dst":    nodet.dstIP + ":" + strconv.Itoa(int(nodet.dstPort)),
		"time":   nodet.time,
		"size":   len(nodet.data),
		"data":   nodet.data,
		"hex":    nodet.hex})

	if err != nil {
		log.Infoln("Error in insert flowt", nodet.nodeID)
		return
	}

	log.Debugln("Inserting flowt doc", insertResult.InsertedID, "...")

	update := bson.M{"$push": bson.M{"nodes": nodet.nodeID}, "$set": bson.M{"lastSeen": nodet.time}}
	_, err = collConnections.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		log.Infoln(err)
	}

	log.Infoln("Inserted nodet", nodet.nodeID, "with lastSeen", nodet.time, "at nodes array in connection doc", nodet.connID)
}

func addNewConnection(nodet *nodet) {

	insertResult, err := collConnections.InsertOne(context.TODO(), bson.M{
		"_id": nodet.connID,
		"endpoints": bson.A{
			bson.A{nodet.srcIP, nodet.srcPort},
			bson.A{nodet.dstIP, nodet.dstPort},
		},
		"lastSeen": nodet.time,
		"favorite": false,
		"nodes":    bson.A{}})

	if err != nil {
		log.Debugln("Object ", nodet.connID, " already exists")
		return
	}

	log.Infoln("Inserted connection doc", insertResult.InsertedID)

}
