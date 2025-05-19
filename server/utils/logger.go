package utils

import (
	"encoding/json"
	"log"
	"sync"

	"dark-swords/types"

	"github.com/gorilla/websocket"
)

// --- Глобальные переменные ---
var GameState types.GameState
var Clients map[*websocket.Conn]string
var Mutex *sync.Mutex

// --- Установка данных ---
func SetGameState(gs *types.GameState) {
	log.Printf("💬 [logger.go SetGameState] SetGameState")
	GameState = *gs
}

func SetClients(ClientsMap map[*websocket.Conn]string) {
	log.Printf("💬 [logger.go SetClients] SetClients")
	Clients = ClientsMap
}

func SetMutex(m *sync.Mutex) {
	log.Printf("💬 [logger.go SetMutex] SetMutex")
	Mutex = m
}

// --- Отправка сообщения в чат ---
func SendChatMessage(sender string, text string) {
	log.Printf("💬 [Чат] %s: %s", sender, text)

	msg := struct {
		Type string `json:"type"` // chat
		Data struct {
			Name string `json:"name"`
			Text string `json:"text"`
		} `json:"data"`
	}{
		Type: "chat",
		Data: struct {
			Name string `json:"name"`
			Text string `json:"text"`
		}{Name: sender, Text: text},
	}

	data, _ := json.Marshal(msg)

	Mutex.Lock()
	Mutex.Unlock()

	for conn := range Clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("❌ Ошибка отправки чата клиенту: %v", err)
		}
	}
}

// --- Логирование события в Log ---
func LogEvent(event string) {
	log.Printf("📝 [Лог] %s", event)

	Mutex.Lock()
	Mutex.Unlock()

	GameState.Log = append(GameState.Log, event)
	//game.BroadcastGameState()
}

func SendChat(text string) {
	log.Printf("💬 [Чат] %s", text)

	msg := struct {
		Type string `json:"type"` // chat
		Data struct {
			Name string `json:"name"`
			Text string `json:"text"`
		} `json:"data"`
	}{
		Type: "chat",
		Data: struct {
			Name string `json:"name"`
			Text string `json:"text"`
		}{Name: "Система", Text: text},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Ошибка сериализации чата: %v", err)
		return
	}

	for conn := range Clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("❌ Ошибка отправки сообщения в чат: %v", err)
		}
	}
}
