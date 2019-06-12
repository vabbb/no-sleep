package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

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
		fmt.Println("Connected to mongodb server!")
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

func insertFlowtDoc(flowt *flowt) {

	addNewConnection(flowt)
	var err error
	filter := bson.M{"_id" : flowt.connID}
	connDoc := collConnections.FindOne(context.TODO(), filter)

	var connDocDecoded bson.M
	err = connDoc.Decode(&connDocDecoded)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Found", connDocDecoded["_id"])
	}

	insertResult, err := collFlows.InsertOne(context.TODO(), bson.M{
		"_id" : flowt.flowID,
		"connID" : flowt.connID,
		"src" : flowt.srcIP + ":" + strconv.Itoa(int(flowt.srcPort)),
		"dst" : flowt.dstIP + ":" + strconv.Itoa(int(flowt.dstPort)),
		"time" : flowt.srcPort,
		"favorite" : flowt.start,
		"hasSYN" : flowt.hasSYN,
		"hasFIN" : flowt.hasFIN,
		"size" : flowt.size,
		"data" : flowt.data,
		"hex" : flowt.hex})

	if err != nil {
		fmt.Println("Error in insert flowt", flowt.flowID)
		return
	}

	fmt.Println("Inserted flowt doc", insertResult.InsertedID)
	
	update := bson.M{"$push": bson.M{"flows": flowt.flowID}} //, "$set": bson.M{"lastSeen": flowt.end}}
	_, err = collConnections.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		fmt.Println(err)
	} 

	fmt.Println("Added flowt", flowt.flowID, "with lastSeen", flowt.end, "at flows array in connection doc", flowt.connID)
}

func addNewConnection(flowt *flowt) {

	insertResult, err := collConnections.InsertOne(context.TODO(), bson.M{
		"_id" : flowt.connID,
		"srcIP" : flowt.srcIP,
		"dstIP" : flowt.dstIP,
		"srcPort" : flowt.srcPort,
		"dstPort" : flowt.dstPort,
		"lastSeen" : flowt.end,
		"favorite" : false,
		"flows" : bson.A{}})
		
	
	if err != nil {
		fmt.Println("Already exists object", flowt.connID)
		return
	}

	fmt.Println("Inserted connection doc", insertResult.InsertedID)

}
