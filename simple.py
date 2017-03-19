from flask import Flask
from flask import request
from flask_cors import CORS, cross_origin
from rpc_client import UserRpcClient

app = Flask(__name__)
CORS(app)

new_user = True

us_rpc = UserRpcClient('localhost','rpc_queue')


@app.route('/login', methods=['POST'])
def hello():
    global new_user
    global us_rpc

    print(request.data)
    print(request.get_json(silent=True)['auth_code'])

    login_request = request.get_json()
    auth_code = login_request['auth_code']
    
    result = us_rpc.login_rpc(auth_code)
    print(result)
    
    if new_user:
        new_user = False
        return "1\n"
    else:
        new_user = True
        return "0\n"

if __name__ == "__main__":
    new_user = True
    app.run()
