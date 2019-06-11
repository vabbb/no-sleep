from flask import Flask, render_template
import fake_db as db

app = Flask(__name__)

@app.route("/")
def hello_world():
    starred = db.get_starred()
    flows   = db.get_flow_list()
    return render_template('index.html', starred=starred, flows=flows)

@app.route("/star/<int:flow_id>/<sel>", methods=['POST'])
def star(flow_id, sel):
    db.star_flow(flow_id, sel)
    return "ok"

