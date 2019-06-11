from flaw_examples import *

def get_flow_list():
    return flows

def get_flow(id):
    for flow in flows:
        if flow['id'] == id:
            return flow
    return False

def get_flow_data(id):
    f = get_flow(id)
    if (f):
        return f['dataFlow']
    return False

def get_starred():
    res = []
    for flow in flows:
        if flow['favourite']:
            res.append(flow)
    return res

def star_flow(id, sel):
    f = get_flow(id)
    if (f):
        f['favourite'] = True if sel == "true" else False
