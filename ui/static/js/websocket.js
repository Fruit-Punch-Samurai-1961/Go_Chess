let webSocket;
const msg = document.getElementById("message");
const log = document.getElementById("textarea");

//append the message to textarea
function appendLog(item) {
    log.value += item.textContent || item.innerText;
    log.value += "\r\n";
}

//clean up the message received
function makeComment(messages) {
    messages = messages.split('\n');
    for (let i = 0; i < messages.length; i++) {
        const item = document.createElement("div");
        item.innerText = messages[i];
        appendLog(item);
    }
}

//get the move made by the second player, check if it's illegal and then update the status and also the board position
function performMove(data) {
    const move = game.move({
        from: data.source,
        to: data.target,
        promotion: data.pawn_promotion
    })
    if (move === null) return 'snapback';

    updateStatus();
    board.position(game.fen())
}

//send the move the user made to the second player
function sendMove(source, target, pawnPromotion, newfen) {
    const info = {
        message_type: "send_move",
        source: source,
        target: target,
        pawn_promotion: pawnPromotion,
        game_fen: newfen
    };
    webSocket.send(JSON.stringify(info))
}

//get the message that was posted by the user and send it to server-side which will then serve it to the rest of the connections
document.getElementById("form").onsubmit = function () {
    if (!msg.value) {
        return false;
    }
    if (!webSocket) {
        return false
    }
    webSocket.send(JSON.stringify({message_type: "comment", comment: msg.value}));
    msg.value = "";
    return false;
}


//checks to see if the browser supports websockets or not and makes a websocket is it is supported
if (window["WebSocket"]) {
    webSocket = new WebSocket("wss://" + document.location.host + "/game/" + gameInfo.key + "/wss");
    webSocket.onopen = function (ev) {
        console.log("Successfully opened a websocket connection!")
    }
    webSocket.onmessage = function (evt) {
        const data = JSON.parse(evt.data)
        if (data.message_type === "comment") {
            makeComment(data.comment)
        } else if (data.message_type === "send_move") {
            performMove(data)
        }
    }
} else {
    const item = document.createElement("div");
    item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
    appendLog(item);
}