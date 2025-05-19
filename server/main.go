package main

import (
	"log"
	"net/http"
	"sync"

	//"sync"

	"dark-swords/game"
	"dark-swords/routes"
	"dark-swords/types"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// game.SetGameState()
	// game.SetMutex(&sync.Mutex{})
	var gameState = types.GameState{
		Players: []types.Player{},
		Map:     game.GenerateFixedMap(),
		Log:     []string{},
	}
	game.SetGameState(gameState)
	game.SetMutex(&sync.Mutex{})
	game.SetClients(make(map[*websocket.Conn]string))

	go game.StartMonsterRespawnLoop()
	go game.CheckCorpseLifespan()

	http.HandleFunc("/ws", routes.HandleWebSocket)
	log.Println("🌍 Сервер запущен на ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
