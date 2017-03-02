from flask import Flask
from flask import request
app = Flask(__name__)

@app.route('/login', methods=['POST'])
def hello():
    print(request.data)
    return "Hi Jonas!\n"

if __name__ == "__main__":
    app.run()
