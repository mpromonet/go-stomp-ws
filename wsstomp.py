#!/usr/bin/env python3
import websocket
import stomper

ws_uri = "ws://{}:{}/ws"

class StompClient(object):

  def __init__(self, server_ip="127.0.0.1", port_number=8765, destinations=[]):
    self.ws_uri = ws_uri.format(server_ip, port_number)
    self.destinations = destinations

  def on_open(self, ws):

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
    ws = websocket.WebSocketApp(self.ws_uri,
                                on_message=self.on_msg,
                                on_error=self.on_error,
                                on_close=self.on_closed)
    ws.on_open = self.on_open
    
    # Run until interruption to client or server terminates connection. 
    ws.run_forever()

  def on_msg(self, ws, msg):
    frame = stomper.Frame()
    unpacked_msg = stomper.Frame.unpack(frame, msg)
    print("Received the message: " + str(unpacked_msg))

  def on_error(self, ws, err):
    print("The Error is:- " + str(err))

  def on_closed(self, ws, status, message):
    print("The websocket connection is closed." + str(message))

if __name__ == "__main__":
    s = StompClient("127.0.0.1", 8765, ["/topic/notifications"])
    s.create_connection()
