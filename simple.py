from flask import Flask
from flask import request
from flask_cors import CORS, cross_origin

app = Flask(__name__)
CORS(app)

new_user = True

@app.route('/login', methods=['POST'])
def hello():
    global new_user
    print(request.data)
    if new_user:
        new_user = False
        return "New User\n"
    else:
        new_user = True
        return "Existing User\n"

if __name__ == "__main__":
    new_user = True
    app.run()
