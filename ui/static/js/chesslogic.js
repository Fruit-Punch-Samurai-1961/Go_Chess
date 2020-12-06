let board = null;
const game = new Chess(movesList[movesList.length - 1]);
const $status = $('#status');
const whiteSquareGrey = '#a9a9a9';
const blackSquareGrey = '#696969';

function removeGreySquares() {
    $('#board1 .square-55d63').css('background', '')
}

function updateMoves(moveslist) {
    for (let i = 1; i < moveslist.length; i++) {
        board.position(moveslist[i])
    }
}

function greySquare(square) {
    const $square = $('#board1 .square-' + square);

    let background = whiteSquareGrey;
    if ($square.hasClass('black-3c85d')) {
        background = blackSquareGrey
    }

    $square.css('background', background)
}

function onDragStart(source, piece) {
    // do not pick up pieces if the game is over
    if (game.game_over()) return false

    //only players 1 and 2 can move the pieces
    if (playerNum > 2) {
        return false;
    }

    // only pick up pieces for the side to move and correct player
    if ((game.turn() === 'w' && (piece.search(/^b/) !== -1 || playerNum === 2)) ||
        (game.turn() === 'b' && (piece.search(/^w/) !== -1 || playerNum === 1))) {
        return false
    }
}

function onDrop(source, target) {
    // see if the move is legal
    const move = game.move({
        from: source,
        to: target,
        promotion: 'q'
    });

    // illegal move
    if (move === null) return 'snapback';


    const newFen = game.fen()
    sendMove(source, target, 'q', newFen);
    updateStatus();
}


function onMouseoverSquare(square, piece) {
    // get list of possible moves for this square
    const moves = game.moves({
        square: square,
        verbose: true
    });

    // exit if there are no moves available for this square
    if (moves.length === 0) return

    // highlight the square they moused over
    greySquare(square)

    // highlight the possible squares for this piece
    for (let i = 0; i < moves.length; i++) {
        greySquare(moves[i].to)
    }
}

function onMouseoutSquare(square, piece) {
    removeGreySquares()
}

// update the board position after the piece snap
// for castling, en passant, pawn promotion
function onSnapEnd() {
    board.position(game.fen())
}

function updateStatus() {
    let status = '';

    let moveColor = 'White';
    if (game.turn() === 'b') {
        moveColor = 'Black'
    }

    // checkmate?
    if (game.in_checkmate()) {
        status = 'Game over, ' + moveColor + ' is in checkmate.'
    }

    // draw?
    else if (game.in_draw()) {
        status = 'Game over, drawn position'
    }

    // game still on
    else {
        status = moveColor + ' to move'

        // check?
        if (game.in_check()) {
            status += ', ' + moveColor + ' is in check'
        }
    }

    $status.html(status)
}


const config = {
    draggable: true,
    position: movesList[0],
    onDragStart: onDragStart,
    onDrop: onDrop,
    onSnapEnd: onSnapEnd,
    onMouseoutSquare: onMouseoutSquare,
    onMouseoverSquare: onMouseoverSquare,
    moveSpeed: 'slow',
    orientation: 'white'
};

if (playerNum === 2) {
    config.orientation =  'black';
}

board = Chessboard('board1', config)

updateStatus()
updateMoves(movesList)