const SPIN_DURATION = 3000;
let selectedBetType = null;
let ws = null;
let clientId = Math.floor(Math.random() * 1000);

function connectWebSocket() {
    if (window["WebSocket"]) {
        ws = new WebSocket(`ws://${document.location.host}/ws/spicy-dice`);
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        ws.onclose = () => {
            console.log('Disconnected from WebSocket');
            setTimeout(connectWebSocket, 3000);
        };
        ws.onopen = () => {
            console.log('Connected to WebSocket');
            // Request initial wallet balance
            ws.send(JSON.stringify({
                type: 'wallet',
                payload: {
                    client_id: clientId
                }
            }));
        };
        ws.onmessage = handleWebSocketMessage

    } else {
        var item = document.getElementsByTagName("body");
        item.innerHTML = "<b class='text-center text-2xl'>Your browser does not support WebSockets.</b>";
    }
}

// Handle incoming WebSocket messages
function handleWebSocketMessage(event) {
    const data = JSON.parse(event.data);
    console.log(data)
    switch (data.type) {
        case "wallet":
            console.log("wallet: ", data.payload.balance)
            updateBalance(data.payload.balance);
            break;
        case "play":
            updateBalance(data.payload.balance)
            setTimeout(() => {
                handleGameResult(data.payload);
                // Send end play request
                ws.send(JSON.stringify({
                    type: 'endplay',
                    payload: {
                        client_id: clientId
                    }
                }));

            }, SPIN_DURATION);
            break;
        case "endplay":
            console.log(data.payload)
            updateBalance(data.payload.balance)
            break;

    }
}

// Update UI balance
function updateBalance(amount) {
    document.getElementById('balance').textContent = `$${amount.toFixed(2)}`;
}

// Handle game result
function handleGameResult(result) {
    const diceContainer = document.getElementById('diceContainer');
    const resultDisplay = document.getElementById('result');
    const betEvenBtn = document.getElementById('betEven');
    const betOddBtn = document.getElementById('betOdd');
    const playButton = document.getElementById('playButton');

    // Stop spinning animation
    diceContainer.classList.remove('dice-spin');

    // Show result
    diceContainer.textContent = `${result.dice_result}`;
    resultDisplay.classList.remove('text-transparent');

    if (result.won) {
        resultDisplay.textContent = 'You Won! 🎉';
        resultDisplay.className = 'text-xl font-bold text-green-400';
    } else {
        resultDisplay.textContent = 'You Lost 😢';
        resultDisplay.className = 'text-xl font-bold text-red-400';
    }

    // Re-enable buttons
    playButton.disabled = false;
    betOddBtn.disabled = false;
    betEvenBtn.disabled = false;
}

// Initialize game controls
function initializeGame() {
    const betEvenBtn = document.getElementById('betEven');
    const betOddBtn = document.getElementById('betOdd');
    const playButton = document.getElementById('playButton');
    const betAmount = document.getElementById('betAmount');


    betEvenBtn.addEventListener('click', () => {
        selectedBetType = 'even';
        betEvenBtn.classList.add('ring-2', 'ring-white');
        betOddBtn.classList.remove('ring-2', 'ring-white');
    });

    betOddBtn.addEventListener('click', () => {
        selectedBetType = 'odd';
        betOddBtn.classList.add('ring-2', 'ring-white');
        betEvenBtn.classList.remove('ring-2', 'ring-white');
    });

    playButton.addEventListener('click', () => {
        if (!selectedBetType || !betAmount.value || betAmount.value <= 0) {
            alert('Please select a bet type and enter a valid bet amount');
            return;
        }

        // Start game
        const diceContainer = document.getElementById('diceContainer');
        const resultDisplay = document.getElementById('result');


        // Reset and start spinning animation
        diceContainer.textContent = '🎲';
        diceContainer.classList.add('dice-spin');
        resultDisplay.textContent = '';
        playButton.disabled = true;
        betEvenBtn.disabled = true
        betOddBtn.disabled = true

        // Send play request
        ws.send(JSON.stringify({
            type: 'play',
            payload: {
                client_id: clientId,
                bet_amount: parseFloat(betAmount.value),
                bet_type: selectedBetType
            }
        }));


    });
}

// Start the game
window.onload = function () {
    connectWebSocket();
    initializeGame();

}