from flask import Flask, render_template, request
import db
from pprint import pprint
import configuration as c
from flow2pwn import flow2pwn
import time, datetime
import binascii

app = Flask(__name__)

def get_services():
    return [c.services[p] for p in c.services]

def get_service_port(service):
    for p in c.services:
        if c.services[p] == service:
            return p

@app.route("/")
def hello_world():
    filters = request.args
    limit = int(filters.get('nflows', '20'))    
    service_port = get_service_port(filters.get('service'))
    starred = db.get_favorite_connections(db.collConnections)
    f = {}
    if service_port:
        f = {'endpoints':{'$elemMatch':{'$elemMatch':{'$in': [service_port]}}}}
    pprint(f)
    conns   = db.get_connections(db.collConnections, f, limit)
    services = get_services()
    return render_template('index.html', starred=starred, connections=conns, services=services, services_map=c.services)

@app.route("/star/<flow_id>/<sel>", methods=['POST'])
def star(flow_id, sel):
    db.star_one_connection(db.collConnections, flow_id, True if sel == 'true' else False)
    return "ok"

@app.route("/starred", methods=['POST'])
def starred():
    starred = db.get_favorite_connections(db.collConnections)
    pprint(c.services)
    return render_template('starred.html', starred=starred, services_map=c.services)

def flow_to_hex(cur):
    res = []
    for i in range(cur.count()):
        res.append(binascii.hexlify(cur[i]["hex"]))
    return res

@app.route("/flow/<flow_id>", methods=['GET'])
def get_flow(flow_id):
    flow, _ = db.get_nodes_of_a_conn(db.collConnections, db.collNodes, flow_id)
    hexdata = flow_to_hex(flow)
    pprint(flow)
    return render_template('flow.html', flow=zip(flow, hexdata), server=c.vm_ip, flow_id=flow_id)

@app.route("/pwn/<flow_id>", methods=['GET'])
def get_flow2pwn(flow_id):
    c, _ = db.get_flows_of_a_conn(db.collConnections, db.collNodes, flow_id)
    return flow2pwn(c)

@app.template_filter('int_to_time')
def convert_int_to_time(t):
    return datetime.datetime.fromtimestamp(t // 1000000000)

app.run(host='127.0.0.1', port=5001)
