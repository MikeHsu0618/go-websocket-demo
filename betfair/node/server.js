const express = require('express')
const SocketServer = require('ws').Server
const PORT = 8080 //指定 port
const BETFAIR_STREAM_HOST = 'stream-api.betfair.com'
const BETFAIR_STREAM_PORT = 443
const BETFAIR_APP_KEY = 'PUwqsW5mxVDdsEoU'

const server = express().listen(PORT, () => {
    console.log(`Listening on ${PORT}`)
})
const wss = new SocketServer({ server })

wss.on('connection', ws => {
    console.log('Client connected')

    let tls = require('tls');
    let options = {
        host: BETFAIR_STREAM_HOST,
        port: BETFAIR_STREAM_PORT
    }
    let betfairClient = tls.connect( options,function () {
        console.log("betfairClient Connected");
    });

    betfairClient.write('{"op": "authentication", "appKey": "PUwqsW5mxVDdsEoU", "session": "h04w1+ocOi6mL/RgZeC1c9Syds7jynqc75tMBC/QFvM="}\r\n');

    /*	Subscribe to order/market stream */
    // betfairClient.write('{"op":"orderSubscription","orderFilter":{"includeOverallPosition":false,"customerStrategyRefs":["betstrategy1"],"partitionMatchedByStrategyRef":true},"segmentationEnabled":true}\r\n');
    betfairClient.write('{"op":"marketSubscription","id":2,"marketFilter":{"marketIds":["1.120684740"],"bspMarket":true,"bettingTypes":["ODDS"],"eventTypeIds":["1"],"eventIds":["27540841"],"turnInPlayEnabled":true,"marketTypes":["MATCH_ODDS"],"countryCodes":["ES"]},"marketDataFilter":{}}\r\n');

    betfairClient.on('data', function (data) {
        console.log("betfairClient received", data.toString())
        ws.send(data.toString())
    });

    betfairClient.on('close', function () {
        console.log('betfairClient closed');
        ws.close()
    });

    betfairClient.on('error', function (err) {
        console.log('betfairClient Error:' + err);
        ws.close()
    });

    ws.on('close', () => {
        betfairClient.destroy()
        console.log('Close client connected')
    })
})



