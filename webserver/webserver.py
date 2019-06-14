from flask import Flask, render_template, request
import db
from pprint import pprint
import configuration as c
from flow2pwn import flow2pwn

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
    flows   = db.find_what_u_want(db.collConnections, f, limit)
    services = get_services()
    return render_template('index.html', starred=starred, flows=flows, services=services, services_map=c.services)

@app.route("/star/<flow_id>/<sel>", methods=['POST'])
def star(flow_id, sel):
    db.star_one_connection(db.collConnections, flow_id, True if sel == 'true' else False)
    return "ok"

@app.route("/starred", methods=['POST'])
def starred():
    starred = db.get_favorite_connections(db.collConnections)
    pprint(c.services)
    return render_template('starred.html', starred=starred, services_map=c.services)

@app.route("/flow/<flow_id>", methods=['GET'])
def get_flow(flow_id):
    h = True if request.args['hex'] == 'true' else False
    flow, _ = db.get_flows_of_a_conn(db.collConnections, db.collFlows, flow_id)
    pprint(flow)
    return render_template('flow.html', flow=flow, client=c.vm_ip, hex=h, flow_id=flow_id)

@app.route("/pwn/<flow_id>", methods=['GET'])
def get_flow2pwn(flow_id):
    c, _ = db.get_flows_of_a_conn(db.collConnections, db.collFlows, flow_id)
    return flow2pwn(c)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80)