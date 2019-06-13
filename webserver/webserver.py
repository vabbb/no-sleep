from flask import Flask, render_template, request
import fake_db as db
from pprint import pprint
import configuration as c
from flow2pwn import flow2pwn

app = Flask(__name__)

def get_services():
    return [s['name'] for s in c.services]    

@app.route("/")
def hello_world():
    filters = request.args
    pprint(filters)
    starred = db.get_starred()
    flows   = db.get_flow_list()
    services = get_services()
    return render_template('index.html', starred=starred, flows=flows, services=services)

@app.route("/star/<int:flow_id>/<sel>", methods=['POST'])
def star(flow_id, sel):
    db.star_flow(flow_id, sel)
    pprint(db.get_flow_list())
    return "ok"

@app.route("/starred", methods=['POST'])
def starred():
    starred = db.get_starred()
    return render_template('starred.html', starred=starred)

@app.route("/flow/<int:flow_id>", methods=['GET'])
def get_flow(flow_id):
    h = True if request.args['hex'] == 'true' else False
    flow = db.get_flow_data(flow_id)
    pprint(flow)
    return render_template('flow.html', flow=flow, client=c.vm_ip, hex=h)

@app.route("/pwn/<int:flow_id>", methods=['GET'])
def get_flow2pwn(flow_id):
    c = db.get_connection(flow_id)
    return flow2pwn(c)