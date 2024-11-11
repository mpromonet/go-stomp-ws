import websocket
import stomper

ws_uri = "ws://{}:{}/ws"

class StompClient(object):

  #Do note that in this case we use jwt_token for authentication hence we are 
  #passing the same in the headers, else we can pass encoded passwords etc. 
  def __init__(self, server_ip="127.0.0.1", port_number=8765, destinations=[]):
    """
    Initializer for the class.
    """
    self.ws_uri = ws_uri.format(server_ip, port_number)
    self.destinations = destinations

  def on_open(self, ws):
    """
    Handler when a websocket connection is opened.

    Args:
      ws(Object): Websocket Object.

    """
    con = stomper.connect("", "", host="")
    ws.send(con)
    
    # Subscribing to all required desitnations. 
    for destination in self.destinations:
      sub = stomper.subscribe(destination, "clientuniqueId", ack="auto")
      ws.send(sub)

      for i in range(10):
        pub = stomper.send(destination, "hello " + str(i))
        ws.send(pub)      

  def create_connection(self):
    """
    Method which starts of the long term websocket connection.
    """

    ws = websocket.WebSocketApp(self.ws_uri,
                                on_message=self.on_msg,
                                on_error=self.on_error,
                                on_close=self.on_closed)
    ws.on_open = self.on_open
    
    # Run until interruption to client or server terminates connection. 
    ws.run_forever()

  def on_msg(self, ws, msg):
    """
    Handler for receiving a message.

    Args:
      msg(str): Message received from stomp watches.

    """
    frame = stomper.Frame()
    unpacked_msg = stomper.Frame.unpack(frame, msg)
    print("Received the message: " + str(unpacked_msg))

  def on_error(self, ws, err):
    """
    Handler when an error is raised.

    Args:
      err(str): Error received.

    """
    print("The Error is:- " + str(err))

  def on_closed(self, ws, status, message):
    """
    Handler when a websocket is closed, ends the client thread.
    """
    print("The websocket connection is closed." + str(message))

if __name__ == "__main__":
    s = StompClient("127.0.0.1", 8765, ["/topic/notifications"])
    s.create_connection()
