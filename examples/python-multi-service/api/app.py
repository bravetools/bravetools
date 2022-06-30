import json
import os
import urllib.request

from flask import Flask


AUTH_ADDR = "http://" + os.environ.get('AUTH_ADDR', "")
LOG_ADDR = "http://" + os.environ.get('LOG_ADDR', "")

app = Flask(__name__)


def check_auth():
    try:
        data = json.dumps({'user': 'test_user'})

        headers = {'Content-Type': 'application/json'}
        req = urllib.request.Request(AUTH_ADDR, data=data.encode(), headers=headers)
        resp = urllib.request.urlopen(req)
        r = resp.read().decode('utf-8')
        return r

    except urllib.error.URLError as e:
        raise ConnectionError(f"Could not connect to auth service at '{AUTH_ADDR}'.", e.reason)

def log_request():
    try:
        resp = urllib.request.urlopen(LOG_ADDR)
        r = resp.read().decode('utf-8')
        return r

    except urllib.error.URLError as e:
        raise ConnectionError(f"Could not connect to logging service at '{LOG_ADDR}'.", e.reason)


@app.route("/")
def serve():
    try:
        auth_status = check_auth()
    except Exception as e:
        auth_status = f"failed to authenticate: {e}"
    
    try:
        log_status = log_request()
    except Exception as e:
        log_status = f"failed to log request: {e}"

    return f"<p>{auth_status}</p><p>{log_status}</p>"

