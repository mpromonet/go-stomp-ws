import { Client } from '@stomp/stompjs';

import { WebSocket } from 'ws';
Object.assign(global, { WebSocket });

const client = new Client({
  brokerURL: 'ws://localhost:8765/ws',
  onConnect: () => {
    client.subscribe('/topic/test01', message =>
      console.log(`Received: ${message.body}`)
    );
    client.publish({ destination: '/topic/test01', body: 'First Message' });
  },
});

client.activate();