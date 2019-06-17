from pymongo import MongoClient, collection, cursor
from pprint import pprint

url = "mongodb://localhost:27017"

client = MongoClient(url)
db = client.my_db
collFlows = db.flows

# return all docs flow in a descending order by end time
def get_flows(collection, filter, limit=0):
    cursor = collection.find(filter)
    if limit > 0:
        cursor.limit(limit)
    return cursor.sort('time', -1)#, cursor.count()

# return all docs node in a ascending order by time start
def get_nodes(collNodes):
    cursor = collNodes.find()
    return cursor.sort('time', 1), cursor.count()

def get_single_flow(collFlows, id):
    return collFlows.find_one({"_id": id})

def get_single_node(collNodes, id):
    return collNodes.find_one({"_id": id})

def get_favorite_flows(collFlows):
    cursor = collFlows.find({'favorite': True})
    return cursor.sort('lastSeen', -1)

def get_favorite_nodes(collNodes):
    cursor = collNodes.find({'favorite': True})
    return cursor.sort('time', 1)

def star_one_flow(collFlows, id, val):
    filter = {"_id": id}
    update = {"$set": {"favorite": val}}
    collFlows.update_one(filter, update)

def star_one_node(collNodes, id):
    filter = {"_id": id}
    update = {"$set": {"favorite": True}}
    collNodes.update_one(filter, update)

def get_nodes_of_a_conn(collCollections, collNodes, idConn):
    connDoc = collCollections.find_one({"_id": idConn})
    cursor = collNodes.find({"connID": connDoc['_id']})
    return cursor.sort('time', 1), cursor.count()

def find_what_u_want(collection, filter, limit=0):
    cursor = collection.find(filter)
    if limit > 0:
        cursor.limit(limit)
    return cursor#, cursor.count()
