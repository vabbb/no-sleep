from flask import Flask, render_template, request
import db, math, binascii, time
from pprint import pprint
import configuration as c
from flow2pwn import flow2pwn
from datetime import datetime
from collections import OrderedDict
from xxd import xxd

application = Flask(__name__)

#"round_time": ["trafficSize", "time_of_latest_flow_analyzed"]
cached_rounds = OrderedDict({})

def get_services():
    return [c.services[p] for p in c.services]

def get_service_port(service):
    for p in c.services:
        if c.services[p] == service:
            return p

def flow_time_to_round(flow_time):
    return datetime.fromtimestamp(flow_time // 1_000_000_000 - flow_time // 1_000_000_000 % 300)

def update_cached_rounds(flows):
    global cached_rounds
    for flow in flows:
        curr_round = flow['time'] // 1_000_000_000 - flow['time'] // 1_000_000_000 % 300
        if curr_round not in cached_rounds.keys():
            cached_rounds[curr_round] = [0, flow['time']]
        cached_rounds[curr_round][0] += flow['trafficSize']
        cached_rounds[curr_round][1] = max(cached_rounds[curr_round][1], flow['time'])

def get_rounds():
    global cached_rounds

    if cached_rounds == {}:
        pprint("CACHE MISS!")
        flows = db.get_unsorted_flows({})

        update_cached_rounds(flows)
    else:
        pprint("CACHE HIT!")
        # pprint(cached_rounds)

        #check only flows that are after the last one of the last round
        check_after_this = cached_rounds[next(iter(cached_rounds))][1]

        #we check before the start of the first round
        #it is assumed that time doesnt go backwards
        check_before_this = next(reversed(cached_rounds))*1000000000

        pprint("only checking flows before: " + str(check_before_this))
        pprint("only checking flows after:  " + str(check_after_this))

        f = { 'time': { '$gt': int(check_after_this) }}
        after_flows = db.get_unsorted_flows(f)
        update_cached_rounds(after_flows)

        g = { 'time': { '$lt': int(check_before_this) }}
        before_flows = db.get_unsorted_flows(g)
        update_cached_rounds(before_flows)

    cached_rounds = OrderedDict(sorted(cached_rounds.items(), reverse=True))
    return cached_rounds

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
    for node in nodes:
        node['blob'] = xxd("".join(map(chr, node['blob'])), n=32)
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

@application.route("/round/<rt>", methods=['GET', 'POST'])
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

@application.route("/pwn/<flow_id>", methods=['GET'])
def get_flow2pwn(flow_id):
    flow = db.get_flow(flow_id)
    return flow2pwn(flow)

@application.template_filter('int_to_round_time')
def int_to_round_time(t):
    return datetime.fromtimestamp(t).strftime("%H:%M")

@application.template_filter('unix_to_human_time')
def unix_to_human_time(t):
    unix = t // 1_000_000_000
    nano = t % 1_000_000_000
    ms = nano // 1_000_000
    return datetime.fromtimestamp(unix).strftime("%H:%M:%S")+".<small>"+str(ms)+"</small>"

@application.template_filter('thousand_comma')
def thousand_comma(s):
    return "{:,}".format(s)

@application.template_filter('format_bytes')
def format_bytes(num):
    suffix='B'
    for unit in ['','K','M','G','T','P','E','Z']:
        if abs(num) < 1000.0:
            return "%3.1f%s%s" % (num, unit, suffix)
        num /= 1000.0
    return "%.1f%s%s" % (num, 'Y', suffix)

if __name__ == "__main__":
    application.run(host='0.0.0.0', port=5001)
