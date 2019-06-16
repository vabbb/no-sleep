package main

import (
	"context"

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
	collFlows = resCollection
}

func addNewFlowToDB(flowt *flowt)  {

	var nodes []bson.M
	for i:=0; i<len(flowt.nodes); i++{
		toAppend := bson.M{
			"fromSrc": flowt.nodes[i].isSrc,
			"time": flowt.nodes[i].time,
			"printableData": flowt.nodes[i].printableData,
			"blob": flowt.nodes[i].blob,
			"hasFlag": flowt.nodes[i].hasFlag}

		nodes = append(nodes, toAppend)
	} 

	insertResult, err := collFlows.InsertOne(context.TODO(), bson.M{
		"time": flowt.start,
		"duration": flowt.end - flowt.start,// flowt.end - flowt.start
        "srcIP": flowt.srcIP,
        "srcPort": flowt.srcPort ,
        "dstIP": flowt.dstIP,
        "dstPort": flowt.dstPort,
        "hasFlag": flowt.hasFlag,
        "trafficSize": flowt.trafficSize, //measured in bytes
        "favorite": false,
        "seenSYN": flowt.seenSYN,
        "seenFIN": flowt.seenFIN,
        "nodes": nodes})

	if err != nil {
		log.Debugln("Error in insert flowt", flowt.flowID)
		return
	}

	log.Infoln("Inserting flowt", flowt.flowID, "with ObjectID", insertResult.InsertedID, "!")
}
