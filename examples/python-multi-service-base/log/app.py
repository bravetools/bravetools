import datetime

from flask import Flask


app = Flask(__name__)

@app.route("/", methods=['GET'])
def log():
    return f"request logged at: {datetime.datetime.now()}"
