var WebSocketClient = require('websocket').client;

var client = new WebSocketClient();

client.on('connectFailed', function(error) {
    console.log('Connect Error: ' + error.toString());
});

client.on('connect', function(connection) {
    console.log('Client Connected');
    connection.on('error', function(error) {
        console.log("Connection Error: " + error.toString());
    });
    connection.on('close', function() {
        console.log('Client Connection Closed');
    });
    connection.on('message', function(message) {
        console.log("Received: '" + JSON.stringify(message) + "'");
    });
});

client.connect('ws://localhost:8080', 'ws');