package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"dark-swords/game"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Не удалось обновить соединение до WebSocket: %v", err)
		return
	}
	log.Printf("🔌 Новое соединение установлено")

	defer func() {
		game.RemoveClient(conn)
	}()

	for {
		msg := struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}{}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("⚠️ Ошибка чтения сообщения: %v", err)
			break
		}

		log.Printf("📨 Получено сообщение: type=%s", msg.Type)

		switch msg.Type {
		case "join":
			game.HandleJoin(conn, msg.Data)
		case "move":
			game.HandleMove(conn, msg.Data)
		case "attack_monster":
			game.HandleAttackMonster(conn, msg.Data)
		case "loot_corpse":
			game.HandleLootCorpse(conn, msg.Data)
		case "equip":
			game.HandleEquip(conn, msg.Data)
		case "unequip":
			game.HandleUnequip(conn, msg.Data)
		case "use_spell":
			game.HandleUseSpell(conn, msg.Data)
		default:
			log.Printf("❓ Неизвестный тип сообщения: %s", msg.Type)
		}
	}
}
