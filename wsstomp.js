import { Client } from '@stomp/stompjs';

import { WebSocket } from 'ws';
Object.assign(global, { WebSocket });

const client = new Client({
  webSocketFactory : function () {
    return new WebSocket("ws://localhost:8765/ws");
  },
  onConnect: () => {
    client.subscribe('/topic/notifications', message =>
      console.log(`Received: ${message.body}`)
    );
    client.publish({ destination: '/topic/notifications', body: 'First Message' });
  },
});

client.activate();