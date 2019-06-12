from flask import Flask, render_template, request
import fake_db as db
from pprint import pprint

app = Flask(__name__)

@app.route("/")
def hello_world():
    filters = request.args
    pprint(filters)
    starred = db.get_starred()
    flows   = db.get_flow_list()
    services = ['one', 'two']
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

@app.route("/flow/<int:flow_id>", methods=['POST'])
def get_flow(flow_id):
    flow = db.get_flow(id)
    return render_template('flow.html', flow=flow)