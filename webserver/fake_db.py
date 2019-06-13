from flaw_examples import *
from pprint import pprint

def get_flow_list():
    return connections


def get_connection(id):
    for c in connections:
        if c['connID'] == str(id):
            return c
    return False


def get_info_connection(id):
    c = get_connection(id)
    if c:
        res = {}
        res['connID'] = c['connID']
        res['srcIP'] = c['srcIP']
        res['dstIP'] = c['dstIP']
        res['srcPort'] = c['srcPort']
        res['dstPort'] = c['dstPort']
        res['lastSeen'] = c['lastSeen']
        res['favorite'] = c['favorite']

        return res

    return False


def get_flow_data(id):
    f = get_connection(id)
    if f:
        return f['flows']
    return False


def get_starred():
    res = []
    for c in connections:
        if c['favorite']:
            res.append(c)
    return res


def star_flow(id, sel):
    c = get_connection(id)
    if c:
        c['favorite'] = True if sel == "true" else False
