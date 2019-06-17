from pymongo import MongoClient, collection, cursor
from bson.objectid import ObjectId
from pprint import pprint

url = "mongodb://localhost:27017"

client = MongoClient(url)
db = client.my_db
collFlows = db.flows

# return all docs flow in a descending order by end time
def get_flows(filter, limit=0, sorteh=-1):
    cursor = collFlows.find(filter)
    if limit > 0:
        cursor.limit(limit)
    return cursor.sort('time', sorteh)#, cursor.count()

# return all docs flow in a descending order by end time
def get_unsorted_flows(filter, limit=0):
    cursor = collFlows.find(filter)
    if limit > 0:
        cursor.limit(limit)
    return cursor#, cursor.count()

# return all docs node in a ascending order by time start
def get_nodes(collNodes):
    cursor = collNodes.find()
    return cursor.sort('time', 1), cursor.count()

def get_flow(idFlow):
    flowDoc = collFlows.find_one({"_id": ObjectId(idFlow)})
    return flowDoc

def get_favorite_flows(collFlows):
    cursor = collFlows.find({'favorite': True})
    return cursor.sort('lastSeen', -1)

def star_one_flow(collFlows, id, val):
    filter = {"_id": id}
    update = {"$set": {"favorite": val}}
    collFlows.update_one(filter, update)

def find_what_u_want(collection, filter, limit=0):
    cursor = collection.find(filter)
    if limit > 0:
        cursor.limit(limit)
    return cursor#, cursor.count()
