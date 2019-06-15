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
	url    = "mongodb://localhost:27017"
	dbName = "my_db"
	flows  = "flows"
)

var (
	// var for mongoDB
	client    *mongo.Client
	collFlows *mongo.Collection
)

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
	if collName == flows {
		collFlows = resCollection

	} else {
		collFlows = resCollection
	}
}

func (flowt *flowt) uploadToMongo() {

	addNewFlow(flowt)
	filter := bson.M{"_id": flowt.flowID}
	// connDoc := collConnections.FindOne(context.TODO(), filter)

	// var connDocDecoded bson.M
	// err = connDoc.Decode(&connDocDecoded)

	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("Found", connDocDecoded["_id"])
	// }

	insertResult, err := collFlows.InsertOne(context.TODO(), bson.M{
		"_id":  flowt.flowID,
		"src":  flowt.srcIP + ":" + strconv.Itoa(int(flowt.srcPort)),
		"dst":  flowt.dstIP + ":" + strconv.Itoa(int(flowt.dstPort)),
		"time": flowt.start,
		// "data":   flowt.data,
		// "hex":    flowt.hex,
	})

	if err != nil {
		log.Infoln("Error in insert flowt", flowt.flowID)
		return
	}

	log.Debugln("Inserting flowt doc", insertResult.InsertedID, "...")

	update := bson.M{"$push": bson.M{"flows": flowt.flowID}, "$set": bson.M{"lastSeen": flowt.end}}
	_, err = collFlows.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		log.Infoln(err)
	}

	log.Infoln("Inserted flowt", flowt.flowID, "with lastSeen", flowt.end, "at flows array in connection doc", flowt.flowID)
}

func addNewFlow(flowt *flowt) {

	insertResult, err := collFlows.InsertOne(context.TODO(), bson.M{
		"_id":    flowt.flowID,
		"client": bson.A{flowt.srcIP, flowt.srcPort},
		"server": bson.A{flowt.dstIP, flowt.dstPort},

		"lastSeen": flowt.end,
		"favorite": false,
		"flows":    bson.A{}})

	if err != nil {
		log.Debugln("Object ", flowt.flowID, " already exists")
		return
	}

	log.Infoln("Inserted connection doc", insertResult.InsertedID)

}
