package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const (
	wordLength = 5
	maxGuesses = 6
)

type SentenceGame struct {
	Words    []string         // The answer words
	Guesses  [][]string       // Each guess is a slice of N words
	Feedback [][][]string     // Feedback per guess per word per letter
	GameOver bool
	Won      bool
}

type GuessResponse struct {
	Valid     bool           `json:"valid"`
	GameOver  bool           `json:"gameOver"`
	Won       bool           `json:"won"`
	Words     []string       `json:"words"`
	Guesses   [][]string     `json:"guesses"`
	Feedback  [][][]string   `json:"feedback"`
}

var game *SentenceGame

func NewSentenceGame(words []string) *SentenceGame {
	return &SentenceGame{
		Words:    words,
		Guesses:  make([][]string, 0),
		Feedback: make([][][]string, 0),
		GameOver: false,
		Won:      false,
	}
}

func (g *SentenceGame) MakeGuess(guesses []string) GuessResponse {
	response := GuessResponse{
		Valid:    false,
		GameOver: g.GameOver,
		Won:      g.Won,
		Words:    g.Words,
		Guesses:  g.Guesses,
		Feedback: g.Feedback,
	}
	if len(guesses) != len(g.Words) {
		return response
	}
	for i := 0; i < len(g.Words); i++ {
		if len(guesses[i]) != wordLength {
			return response
		}
		guesses[i] = strings.ToUpper(guesses[i])
	}
	g.Guesses = append(g.Guesses, guesses)

	// Feedback for this guess: [numWords][wordLength]
	guessFeedback := make([][]string, len(g.Words))
	allCorrect := true
	for w := 0; w < len(g.Words); w++ {
		feedback := make([]string, wordLength)
		used := make([]bool, wordLength)
		word := g.Words[w]
		guess := guesses[w]
		// First pass: correct positions
		for i := 0; i < wordLength; i++ {
			if guess[i] == word[i] {
				feedback[i] = "correct"
				used[i] = true
			} else {
				allCorrect = false
			}
		}
		// Second pass: present but wrong position
		for i := 0; i < wordLength; i++ {
			if feedback[i] == "correct" {
				continue
			}
			for j := 0; j < wordLength; j++ {
				if !used[j] && guess[i] == word[j] {
					feedback[i] = "present"
					used[j] = true
					break
				}
			}
			if feedback[i] == "" {
				feedback[i] = "absent"
			}
		}
		guessFeedback[w] = feedback
	}
	g.Feedback = append(g.Feedback, guessFeedback)

	response.Valid = true
	response.Guesses = g.Guesses
	response.Feedback = g.Feedback

	if allCorrect {
		g.GameOver = true
		g.Won = true
		response.GameOver = true
		response.Won = true
		return response
	}
	if len(g.Guesses) >= maxGuesses {
		g.GameOver = true
		response.GameOver = true
	}
	return response
}

func handleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	game.GameOver = false
	game.Won = false
	game.Guesses = make([][]string, 0)
	game.Feedback = make([][][]string, 0)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"words": game.Words,
	})
}

func handleGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Guesses []string `json:"guesses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	response := game.MakeGuess(request.Guesses)
	json.NewEncoder(w).Encode(response)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Multi-Wordle</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            text-align: center;
        }
        .grids {
            display: flex;
            justify-content: center;
            gap: 40px;
            margin-bottom: 20px;
        }
        .board {
            display: grid;
            grid-template-rows: repeat(6, 1fr);
            gap: 5px;
        }
        .row {
            display: grid;
            grid-template-columns: repeat(5, 1fr);
            gap: 5px;
        }
        .cell {
            width: 60px;
            height: 60px;
            border: 2px solid #ccc;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 24px;
            font-weight: bold;
            text-transform: uppercase;
        }
        .correct { background-color: #6aaa64; color: white; }
        .present { background-color: #c9b458; color: white; }
        .absent { background-color: #787c7e; color: white; }
        input {
            font-size: 18px;
            padding: 10px;
            width: 120px;
            text-transform: uppercase;
            margin: 0 5px;
        }
        button {
            font-size: 18px;
            padding: 10px 20px;
            margin: 10px;
            cursor: pointer;
        }
        .message {
            margin: 20px 0;
            font-size: 18px;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <h1>Multi-Wordle</h1>
    <div class="grids" id="grids"></div>
    <div id="message" class="message"></div>
    <div id="inputs"></div>
    <button onclick="makeGuess()">Guess</button>
    <button onclick="newGame()">New Game</button>
    <script>
        let numWords = {{len .Words}};
        let boards = [];
        let guessInputs = [];
        const message = document.getElementById('message');
        function createBoards() {
            const gridsDiv = document.getElementById('grids');
            gridsDiv.innerHTML = '';
            boards = [];
            for (let b = 0; b < numWords; b++) {
                const board = document.createElement('div');
                board.className = 'board';
                for (let i = 0; i < 6; i++) {
                    const row = document.createElement('div');
                    row.className = 'row';
                    for (let j = 0; j < 5; j++) {
                        const cell = document.createElement('div');
                        cell.className = 'cell';
                        row.appendChild(cell);
                    }
                    board.appendChild(row);
                }
                gridsDiv.appendChild(board);
                boards.push(board);
            }
        }
        function createInputs() {
            const inputsDiv = document.getElementById('inputs');
            inputsDiv.innerHTML = '';
            guessInputs = [];
            for (let b = 0; b < numWords; b++) {
                const input = document.createElement('input');
                input.type = 'text';
                input.maxLength = 5;
                input.placeholder = 'Word ' + (b + 1);
                input.id = 'guess' + b;
                input.style.margin = '0 5px';
                inputsDiv.appendChild(input);
                guessInputs.push(input);
            }
        }
        function updateBoards(guesses, feedback) {
            for (let b = 0; b < numWords; b++) {
                const rows = boards[b].getElementsByClassName('row');
                for (let i = 0; i < guesses.length; i++) {
                    const cells = rows[i].getElementsByClassName('cell');
                    for (let j = 0; j < 5; j++) {
                        cells[j].textContent = guesses[i][b][j];
                        cells[j].className = 'cell ' + feedback[i][b][j];
                    }
                }
            }
        }
        async function newGame() {
            const response = await fetch('/new-game', { method: 'POST' });
            const data = await response.json();
            numWords = data.words.length;
            createBoards();
            createInputs();
            message.textContent = '';
            guessInputs.forEach(input => input.value = '');
            // Add Enter key event
            guessInputs.forEach(input => {
                input.addEventListener('keypress', function(e) {
                    if (e.key === 'Enter') {
                        makeGuess();
                    }
                });
            });
        }
        async function makeGuess() {
            const guesses = guessInputs.map(input => input.value.toUpperCase());
            if (guesses.some(g => g.length !== 5)) {
                message.textContent = 'Please enter ' + numWords + ' 5-letter words!';
                return;
            }
            const response = await fetch('/guess', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ guesses: guesses })
            });
            const data = await response.json();
            if (!data.valid) {
                message.textContent = 'Invalid guess!';
                return;
            }
            updateBoards(data.guesses, data.feedback);
            guessInputs.forEach(input => input.value = '');
            if (data.gameOver) {
                if (data.won) {
                    message.textContent = 'Congratulations! You guessed the sentence!';
                } else {
                    message.textContent = 'Game Over! The sentence was: ' + data.words.join(' ');
                }
            }
        }
        // Start a new game when the page loads
        newGame();
    </script>
</body>
</html>`
	t, err := template.New("home").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, struct{ Words []string }{Words: game.Words})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go 'WORD1 WORD2 ...' (all 5-letter words)")
		os.Exit(1)
	}
	parts := strings.Fields(os.Args[1])
	if len(parts) < 1 {
		fmt.Println("Please provide at least one 5-letter word.")
		os.Exit(1)
	}
	for _, w := range parts {
		if len(w) != wordLength {
			fmt.Printf("Each word must be 5 letters: '%s' is not\n", w)
			os.Exit(1)
		}
	}
	words := make([]string, len(parts))
	for i := 0; i < len(parts); i++ {
		words[i] = strings.ToUpper(parts[i])
	}
	game = NewSentenceGame(words)

	r := mux.NewRouter()
	r.HandleFunc("/new-game", handleNewGame).Methods("POST")
	r.HandleFunc("/guess", handleGuess).Methods("POST")
	r.HandleFunc("/", handleHome).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Handler:      c.Handler(r),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Starting Multi-Wordle server on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
} 