from pymongo import MongoClient
from pymongo import collection
from pymongo import cursor
from pprint import pprint

url = "mongodb://localhost:27017"

client = MongoClient(url)
db = client.my_db
collConnections = db.connections
collFlows = db.flows

# return all docs connection in a descending order by end time
def get_connections(collConnections):
    cursor = collConnections.find()
    return cursor.sort('lastSeen', -1)

# return all docs flow in a ascending order by time start
def get_flows(collFlows):
    cursor = collFlows.find()
    return cursor.sort('time', 1)

def get_single_connection(collConnections, id):
    return collConnections.find_one({"_id": id})

def get_single_flow(collFlows, id):
    return collFlows.find_one({"_id": id})

def get_favorite_connections(collConnections):
    cursor = collConnections.find({'favorite': True})
    return cursor.sort('lastSeen', -1)

def get_favorite_flows(collFlows):
    cursor = collFlows.find({'favorite': True})
    return cursor.sort('time', 1)

def star_one_connection(collConnections, id):
    filter = {"_id": id}
    update = {"$set": {"favorite": True}}
    collConnections.update_one(filter, update)

def star_one_flow(collFlows, id):
    filter = {"_id": id}
    update = {"$set": {"favorite": True}}
    collFlows.update_one(filter, update)

def get_flows_of_a_conn(collCollections, collFlows, idConn):
    connDoc = collCollections.find_one({"_id": idConn})
    cursor = collFlows.find({"connID": connDoc['_id']})
    return cursor.sort('time', 1)

def find_what_u_want(collection, filter):
    return collection.find(filter)

