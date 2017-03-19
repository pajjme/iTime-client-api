import pika
import uuid

class UserRpcClient:
    def __init__ (self,rabbit_url,rpc_queue):
        self.rpc_queue = rpc_queue

        self.connection = pika.BlockingConnection(
                pika.ConnectionParameters(host=rabbit_url))
        self.channel = self.connection.channel()

        queue_address = self.channel.queue_declare(exclusive=True)
        self.callback_queue = queue_address.method.queue

        #subscribe to the rpc callback queue
        self.channel.basic_consume(
                self.on_response,
                no_ack = True,
                queue = self.callback_queue)

    def on_response(self,channel,method,properties,body):
        if self.correlation_id == properties.correlation_id:
            self.response = body

    def login_rpc(self,auth_code):
        self.response = None
        self.correlation_id = str(uuid.uuid4())
        
        self.channel.basic_publish(exchange='',
                routing_key = self.rpc_queue,
                properties = pika.BasicProperties(
                    reply_to = self.callback_queue,
                    correlation_id = self.correlation_id,
                    ),
                body=str(auth_code))



        #wait for response from user service
        while self.response is None:
            self.connection.process_data_events()
        return str(self.response)

