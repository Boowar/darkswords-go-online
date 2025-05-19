package game

import (
	"dark-swords/types"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// --- Глобальные переменные ---
var GameState types.GameState
var Clients map[*websocket.Conn]string
var Mutex *sync.Mutex

// --- Установка данных ---
func SetGameState(gs types.GameState) {
	GameState = gs
}

func SetMutex(m *sync.Mutex) {
	Mutex = m
}

func SetClients(clientsMap map[*websocket.Conn]string) {
	Clients = clientsMap
}

// --- Рассылка игрового состояния ---
func BroadcastGameState() {
	log.Printf("🔄 Рассылка состояния мира всем клиентам")
	//log.Printf("🎮 [broadcast] Текущее состояние игры: %+v", gameState)
	log.Printf("👥 [broadcast] Подключено клиентов: %d", len(Clients))
	Mutex.Lock()
	log.Printf("🔒 Мьютекс захвачен")
	defer func() {
		Mutex.Unlock()
		log.Printf("🔓 Мьютекс освобождён")
	}()
	GameState.Map = updateMapWithPlayers(GameState.Players, GameState.Map)

	// Создаём копию текущего состояния
	dataToSend := GameState

	// Очищаем лог перед следующим раундом
	GameState.Log = nil

	data, err := json.Marshal(struct {
		Type string          `json:"type"`
		Data types.GameState `json:"data"`
	}{
		Type: "update",
		Data: dataToSend,
	})

	if err != nil {
		log.Printf("❌ Ошибка сериализации игрового состояния: %v", err)
		return
	}

	//log.Printf("📤 Отправляемое сообщение: %s", data)

	for conn := range Clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("❌ Ошибка отправки клиенту: %v", err)
		} else {
			log.Printf("✅ Отправлено сообщение клиенту")
		}
	}
}
