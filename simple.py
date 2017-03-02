from flask import Flask
from flask import request
from flask_cors import CORS, cross_origin

app = Flask(__name__)
CORS(app)

@app.route('/login', methods=['POST'])
def hello():
    print(request.data)
    return "Hi Jonas!\n"

if __name__ == "__main__":
    app.run()
