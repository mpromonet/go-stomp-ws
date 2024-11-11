go-stomp-ws
---

Tries to use [go-stomp](https://github.com/go-stomp/stomp) over websocket

Start server
---
```
./go-stomp-ws server
```

Run client
---
```
# subscribe to a topic
./go-stomp-ws client

# publish a message to a topic
./go-stomp-ws client -m hello

# in javascript
npm start

# in python
./wspython.py
```