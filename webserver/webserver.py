from flask import Flask, render_template, request
import db, math, binascii
from pprint import pprint
import configuration as c
# from flow2pwn import flow2pwn
import time, datetime
from collections import OrderedDict

application = Flask(__name__)

def get_services():
    return [c.services[p] for p in c.services]

def get_service_port(service):
    for p in c.services:
        if c.services[p] == service:
            return p

def flow_time_to_round(flow_time):
    return datetime.datetime.fromtimestamp(flow_time // 1000000000 - flow_time // 1000000000 % 300)

# def remove_duplicates_and_keep_order(my_list):
#     seen = set()
#     seen_add = seen.add
#     return [x for x in my_list if not (x in seen or seen_add(x))]

def get_rounds(flows=None):
    rounds = {}
    f = {}
    if flows == None:
        flows = db.get_unsorted_flows(f)

    for flow in flows:
        curr_round = flow['time'] // 1000000000 - flow['time'] // 1000000000 % 300
        if curr_round not in rounds.keys():
            rounds[curr_round] = 0
        rounds[curr_round] += flow['trafficSize']

    return OrderedDict(sorted(rounds.items(), reverse=True))

@application.route("/")
def index():
    filters = request.args
    limit = int(filters.get('nflows', '20'))
    f = {}
    flows = db.get_flows(f, limit, -1)
    # pprint(flows)
    return render_template( 'index.html',
                            flows=flows,
                            rounds=get_rounds(),
                            services_map=c.services,
                            services=get_services())

def modify_blobs(flow):
    nodes = flow['nodes']
    for i in range(len(nodes)):
        nodes[i]['blob'] = binascii.hexlify(nodes[i]['blob'])
    return flow

@application.route("/flow/<flow_id>", methods=['GET'])
def slash_flow(flow_id):
    flow = db.get_flow(flow_id)
    pprint(flow)
    flow = modify_blobs(flow)
    return render_template( 'flow.html',
                            flow=flow,
                            server="replace this", 
                            hex=request.args['hex'], 
                            flow_id=flow_id)

@application.route("/rounds", methods=['POST'])
def slash_rounds():
    rounds = get_rounds()
    # pprint (rounds)
    return render_template('rounds.html', rounds=rounds)

@application.route("/round/<rt>", methods=['POST'])
@application.route("/round/<rt>", methods=['GET'])
def slash_round(rt):
    if rt == 'ongoing':
        return render_template('round.html', flows={})
    f = { 'time': { '$gte': int(rt)*1000000000, '$lt': (int(rt)+300)*1000000000},  }
    pprint(f)
    flows = db.get_flows(f)
    return render_template( 'round.html',
                            flows=flows,
                            services_map=c.services,
                            services=get_services())


# @application.route("/star/<flow_id>/<sel>", methods=['POST'])
# def star(flow_id, sel):
#     db.star_one_connection(db.collConnections, flow_id, True if sel == 'true' else False)
#     return "ok"

# @application.route("/starred", methods=['POST'])
# def starred():
#     starred = db.get_favorite_connections(db.collConnections)
#     pprint(c.services)
#     return render_template('starred.html', starred=starred, services_map=c.services)

# @application.route("/pwn/<flow_id>", methods=['GET'])
# def get_flow2pwn(flow_id):
#     c, _ = db.get_flows_of_a_conn(db.collConnections, db.collFlows, flow_id)
#     return flow2pwn(c)

@application.template_filter('int_to_round_time')
def int_to_round_time(t):
    return datetime.datetime.fromtimestamp(t).strftime("%H:%M")

@application.template_filter('unix_to_human_time')
def unix_to_human_time(t):
    unix = t // 1000000000
    nano = t % 1000000000
    ms = nano // 1000000
    return datetime.datetime.fromtimestamp(unix).strftime("%H:%M:%S")+".<small>"+str(ms)+"</small>"

@application.template_filter('thousand_comma')
def thousand_comma(s):
    return "{:,}".format(s)

if __name__ == "__main__":
	application.run(host='0.0.0.0', port=5001)
