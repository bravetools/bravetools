from flask import Flask, request


app = Flask(__name__)

@app.route("/", methods=['POST'])
def authenticate():
    data = request.json
    return f"authenticated user: {data}"
