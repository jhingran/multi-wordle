# Wordle Web Game

A web-based implementation of the popular Wordle game using Go.

## Prerequisites

- Go 1.21 or later
- Git

## Installation

1. Install Go from [golang.org/dl](https://golang.org/dl)

2. Clone this repository:
```bash
git clone <repository-url>
cd wordle
```

3. Install dependencies:
```bash
go mod download
```

## Running the Game

1. Start the server:
```bash
go run main.go
```

2. Open your web browser and visit:
```
http://localhost:8080
```

## Features

- 6 attempts to guess a 5-letter word
- Color-coded feedback:
  - Green: Correct letter in correct position
  - Yellow: Correct letter in wrong position
  - Gray: Letter not in the word
- Modern web interface
- RESTful API endpoints
- CORS support for cross-origin requests

## API Endpoints

- `GET /` - Web interface
- `POST /new-game` - Create a new game
- `POST /guess` - Make a guess in the current game

## Dependencies

- github.com/gorilla/mux - HTTP router
- github.com/rs/cors - CORS middleware 