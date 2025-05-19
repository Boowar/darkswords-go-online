const socket = new WebSocket('ws://localhost:8080/ws');

// DOM элементы
const loginForm = document.getElementById("loginForm");
const gameContainer = document.getElementById("gameContainer");
const playerDisplayName = document.getElementById("playerDisplayName");
const playerStats = document.getElementById("playerStats");
const gameMapDiv = document.getElementById("gameMap");
const chatLog = document.getElementById("chatLog");
const attackMonsterBtn = document.getElementById("attackMonsterBtn");

let playerName = null;
let gameState = { Players: [], Map: [], Log: [] };

console.log("🎮 Клиент запущен");

// --- События WebSocket ---
socket.onopen = () => {
    console.log("🔗 Соединение установлено");
};

socket.onmessage = function(event) {
    console.log("📨 [Сообщение от сервера]");
    console.log("RAW данные:", event.data);

    let msg;
    try {
        msg = JSON.parse(event.data);
        console.log("Разобранные данные:", msg);
    } catch (e) {
        console.error("❌ Не удалось разобрать JSON:", event.data);
        return;
    }

    if (msg.type === "update") {
        console.log("🔄 Получено обновление состояния игры");
        gameState.Players = msg.data.Players; // Сохраняем состояние
        gameState.Map = msg.data.Map; // Сохраняем состояние
        gameState.Log = msg.data.Log; // Сохраняем состояние
        updateUI(msg.data);
    } else if (msg.type === "chat") {
        console.log("💬 Получено чат-сообщение:", msg.data.text);
        logMessage(`💬 ${msg.data.name}: ${msg.data.text}`);
    } else {
        console.warn("❓ Неизвестный тип сообщения:", msg.type);
    }

    console.log("");
};

function logMessage(text) {
    const log = document.getElementById("chatLog");
    log.innerHTML += `<div>${text}</div>`;
    log.scrollTop = log.scrollHeight;
}

// --- Вход в игру ---
function join() {
    console.group("👤 [JOIN] Присоединение игрока");
    const name = document.getElementById("playerName").value.trim();
    const race = document.getElementById("playerRace").value;
    const religion = document.getElementById("playerReligion").value;

    console.log("Имя:", name);
    console.log("Раса:", race);
    console.log("Религия:", religion);

    if (!name || !race || !religion) {
        alert("Введите имя, расу и религию!");
        console.warn("⚠️ Недостающие данные при входе");
        console.groupEnd();
        return;
    }

    playerName = name;
    playerDisplayName.textContent = name;

    console.log("📤 Отправляем JOIN команду на сервер...");
    socket.send(JSON.stringify({
        type: "join",
        data: { name, race, religion }
    }));

    loginForm.style.display = "none";
    gameContainer.style.display = "block";

    console.log("✅ Игрок вошёл в игру");
    console.groupEnd();
}

// --- Выход из игры ---
function leaveGame() {
    console.group("🚪 [LEAVE] Покидание игры");
    if (!playerName) {
        console.warn("🚫 Игрок не авторизован");
        console.groupEnd();
        return;
    }

    console.log("📤 Отправляем LEAVE команду на сервер...");
    socket.send(JSON.stringify({ 
        type: "leave", 
        data: { name: playerName } 
    }));

    playerName = null;
    gameState = { players: [], map: [], log: [] };

    loginForm.style.display = "block";
    gameContainer.style.display = "none";
    document.getElementById("playerName").value = "";
    document.getElementById("playerStats").innerHTML = "";

    console.log("✅ Игрок вышел из игры");
    console.groupEnd();
}

// --- Обновление интерфейса ---
function updateUI(data) {
    console.group("📊 [UPDATE] Обновление интерфейса");
    console.log("Полученные данные:", data);

    if (!data) {
        console.error("❌ Данные пустые");
        return;
    }

    const player = data.Players.find(p => p.Name === playerName);
    if (!player) {
        console.warn("🚫 Игрок не найден в состоянии игры");
        return;
    }

    console.log("🎮 Информация о персонаже:", player);

    playerStats.innerHTML = `
        <p><strong>🧬 Раса:</strong> ${player.Race}</p>
        <p><strong>🕯️ Религия:</strong> ${player.Religion}</p>
        <p><strong>🏅 Уровень:</strong> ${player.Level}</p>
        <p><strong>⚡ XP:</strong> ${player.CurrentXP} / ${player.RequiredXP}</p>
        <p><strong>❤️ HP:</strong> ${player.HP}, <strong>🧠 MP:</strong> ${player.MP}</p>
        <p><strong>💥 Урон:</strong> ${player.Damage}</p>
        <p><strong>Тело:</strong> ${player.Body}, <strong>Сила:</strong> ${player.Strength}</p>
        <p><strong>Ловкость:</strong> ${player.Dexterity}, <strong>Разум:</strong> ${player.Intelligence}</p>
    `;

    renderMap(data.Map);
    renderInventory(player.Items);

    // --- Отображение надетых предметов ---
    const equippedDiv = document.getElementById("equipped");
    equippedDiv.innerHTML = "<h3>⚔️ Надетые предметы:</h3>";

    if (player.EquippedItems && typeof player.EquippedItems === 'object') {
        Object.entries(player.EquippedItems).forEach(([slot, item]) => {
            equippedDiv.innerHTML += `<p>🗡 ${item.Name} (${item.Type}): ${formatBonus(item.Bonus)}</p><button onclick="unequip('${slot}')">🧦 Снять</button>`;
        });
    } else {
        equippedDiv.innerHTML += "<p>Пока ничего не надето</p>";
    }

    // --- Фильтрация лога: только бой в текущей комнате ---
    const currentRoom = data.Map[player.Position.Y][player.Position.X];

    // Проверяем, есть ли бой в комнате
    if (currentRoom.Monsters && currentRoom.Monsters.length > 0) {
        document.getElementById("chatLog").innerHTML += `<div class="log-event">⚔️ Бой начался в вашей комнате!</div>`;
        currentRoom.Monsters.forEach(monster => {
            document.getElementById("chatLog").innerHTML += `
                <div class="log-event">
                    👹 Монстр: ${monster.Name} | Уровень: ${monster.Level} | HP: ${monster.HP}
                </div>`;
        });
    }

    // --- Лог боя из gameState.Log ---
    if (data.Log && Array.isArray(data.Log)) {
        data.Log.forEach(logEntry => {
            const lower = logEntry.toLowerCase();
            if (
                lower.includes(player.Name.toLowerCase()) || // если это ты
                lower.includes(`[${player.Position.X},${player.Position.Y}]`) // если это твоя комната
            ) {
                logMessage(logEntry);
            }
        });
    }

    // --- Проверяем, есть ли труп в комнате ---
    const corpseInfoDiv = document.getElementById("corpseInfo");
    const corpseNameSpan = document.getElementById("corpseName");
    const lootBtn = document.getElementById("lootBtn");

    if (currentRoom.CorpsePlayer && new Date(currentRoom.CorpseTime) > new Date()) {
        corpseNameSpan.textContent = currentRoom.CorpsePlayer;
        corpseInfoDiv.style.display = "block";
        lootBtn.style.display = "block";
        lootBtn.onclick = () => {
            socket.send(JSON.stringify({
                type: "loot_corpse",
                data: { name: playerName, x: player.Position.X, y: player.Position.Y }
            }));
        };
    } else {
        corpseInfoDiv.style.display = "none";
        lootBtn.style.display = "none";
    }

    console.groupEnd();
}

// --- Отрисовка карты ---
function renderMap(map) {
    gameMapDiv.innerHTML = "";

    for (let y = 0; y < map.length; y++) {
        for (let x = 0; x < map[y].length; x++) {
            const room = map[y][x];
            const roomDiv = document.createElement("div");
            roomDiv.className = "room empty-room";

            if (room.BiomeType === "forest") {
                roomDiv.classList.replace("empty-room", "biome-forest");
                roomDiv.innerHTML += `<div class="icon">🌲</div>`;
            } else if (room.BiomeType === "desert") {
                roomDiv.classList.replace("empty-room", "biome-desert");
                roomDiv.innerHTML += `<div class="icon">☀️</div>`;
            } else if (room.BiomeType === "mountains") {
                roomDiv.classList.replace("empty-room", "biome-mountains");
                roomDiv.innerHTML += `<div class="icon">⛰</div>`;
            } else if (room.BiomeType === "swamp") {
                roomDiv.classList.replace("empty-room", "biome-swamp");
                roomDiv.innerHTML += `<div class="icon">🌿</div>`;
            } else if (room.BiomeType === "road") {
                roomDiv.classList.replace("empty-room", "road");
                roomDiv.innerHTML += `<div class="icon">🛣</div>`;
            }
            // --- Установка класса по типу ---
            if (room.LocationType === "shop") {
                //roomDiv.classList.replace("empty-room", "safe-room");
                roomDiv.innerHTML += `<div class="icon">🛒</div>`;
            } else if (room.LocationType === "guild") {
                //roomDiv.classList.replace("empty-room", "safe-room");
                roomDiv.innerHTML += `<div class="icon">🏛</div>`;
            } else if (room.LocationType === "stair") {
                //roomDiv.classList.replace("empty-room", "safe-room");
                roomDiv.innerHTML += `<div class="icon">🪜</div>`;
            }  

            if (room.CorpsePlayer && new Date(room.CorpseTime) > new Date()) {
                //roomDiv.classList.replace("empty-room", "corpse-room");
                roomDiv.innerHTML += `<div class="icon">☠️</div>`;
            } else if (room.LocationType === "safe") {
                //roomDiv.classList.replace("empty-room", "safe-room");
                roomDiv.innerHTML += `<div class="icon">🏠</div>`;
            } else if (room.Monsters && room.Monsters.length > 0) {
                //roomDiv.classList.replace("empty-room", "monster-room");
                roomDiv.innerHTML += `<div class="icon">👹</div><div>${room.Monsters.length}</div>`;
            }

            // Игроки в комнате
            const playersHere = gameState.Players.filter(p => p.Position.X === x && p.Position.Y === y).map(p => p.Name);
            if (playersHere.length > 0) {
                const names = document.createElement("div");
                names.style.fontSize = "0.6em";
                names.textContent = playersHere.join(", ");
                roomDiv.appendChild(names);
            }

            roomDiv.onclick = () => {
                if (playerName) {
                    movePlayer(x, y);
                }
            };

            gameMapDiv.appendChild(roomDiv);
        }
    }

    // --- Отрисовка информации о монстре в комнате игрока ---
    const player = gameState.Players.find(p => p.Name === playerName);
    if (!player) return;

    const currentRoom = map[player.Position.Y][player.Position.X];

    const monsterInfoDiv = document.getElementById("monsterInfo");
    const monsterNameSpan = document.getElementById("monsterName");
    const monsterLevelSpan = document.getElementById("monsterLevel");
    const monsterHPSpan = document.getElementById("monsterHP");
    const monsterMaxHPSpan = document.getElementById("monsterMaxHP");
    const monsterDamageSpan = document.getElementById("monsterDamage");
    const monsterSchoolSpan = document.getElementById("monsterSchool");
    const monsterResistSpan = document.getElementById("monsterResist");

    if (currentRoom.Monsters && currentRoom.Monsters.length > 0) {
        const monster = currentRoom.Monsters[0];
        console.log("👹 Монстр в комнате:", monster);

        monsterNameSpan.textContent = monster.Name || "Неизвестно";
        monsterLevelSpan.textContent = monster.Level || 0;
        monsterHPSpan.textContent = monster.HP || 0;
        monsterMaxHPSpan.textContent = monster.MaxHP || 0;
        monsterDamageSpan.textContent = monster.Damage || 0;
        monsterSchoolSpan.textContent = monster.School || "нет школы";
        monsterResistSpan.textContent = Object.entries(monster.Resist || {}).map(([k, v]) => `${k}:${v}`).join(" ") || "нет сопротивления";

        monsterInfoDiv.style.display = "block";

        console.log("⚔️ Монстр в комнате найден:", currentRoom.Monsters[0].Name);
        attackMonsterBtn.style.display = "block";
        attackMonsterBtn.onclick = () => {
            console.log("🗡️ Игрок начал атаковать монстра в комнате");
            attackMonster(player.Position.X, player.Position.Y);
        };
    } else {
        monsterInfoDiv.style.display = "none";
        console.log("🛡️ В комнате нет монстров");
        attackMonsterBtn.style.display = "none";
    }
}

// --- Команды ---
function movePlayer(x, y) {
    console.log(`👣 Игрок отправляет команду MOVE на [${x},${y}]`);
    socket.send(JSON.stringify({
        type: "move",
        data: {
            name: playerName,
            to: { x, y }
        }
    }));
}

function attackMonster(x, y) {
    console.log(`⚔️ Игрок отправляет команду ATTACK_MONSTER на [${x},${y}]`);
    socket.send(JSON.stringify({
        type: "attack_monster",
        data: {
            name: playerName,
            x, y
        }
    }));
}

function lootCorpse() {
    const player = gameState.Players.find(p => p.Name === playerName);
    if (!player) return;

    const x = player.Position.X;
    const y = player.Position.Y;

    socket.send(JSON.stringify({
        type: "loot_corpse",
        data: { name: playerName, x, y }
    }));
}

function renderInventory(items) {
    const inventoryList = document.getElementById("inventoryList");
    if (!items || items.length === 0) {
        inventoryList.innerHTML = `<p>🎒 Ваш инвентарь пуст</p>`;
        return;
    }

    inventoryList.innerHTML = "";
    items.forEach((item, index) => {
        const itemDiv = document.createElement("div");
        itemDiv.className = "item";
        itemDiv.style.margin = "5px 0";

        itemDiv.innerHTML = `
            <strong>${item.Name}</strong> (${item.Type})<br>
            <small>Мин. уровень: ${item.MinLevel} | Бонус: ${formatBonus(item.Bonus)}</small><br>
        `;

        // --- Кнопки ---
        let button = null;

        if (item.Type === "weapon" || item.Type === "ring") {
            button = document.createElement("button");
            button.textContent = "🧦 Надеть";
            button.onclick = () => equipItem(item.Name);
        } else if (item.Type === "potion" || item.Type === "scroll") {
            button = document.createElement("button");
            button.textContent = "🧪 Использовать";
            button.onclick = () => useItem(item.Name);
        }

        if (button) {
            itemDiv.appendChild(button);
        }

        inventoryList.appendChild(itemDiv);
    });
}

function formatBonus(bonus) {
    if (typeof bonus === "number") {
        return `+${bonus}`;
    } else if (typeof bonus === "object" && bonus !== null) {
        let result = "";
        for (const key in bonus) {
            result += `${key}: +${bonus[key]}, `;
        }
        return result.slice(0, -2); // убираем последнюю запятую
    }
    return "нет бонусов";
}

function equipItem(itemName) {
    socket.send(JSON.stringify({
        type: "equip",
        data: { name: playerName, item_name: itemName }
    }));
}

function useItem(itemName) {
    equipItem(itemName); // те же данные, но сервер проверяет тип
}

function unequip(itemType) {
    socket.send(JSON.stringify({
        type: "unequip",
        data: { name: playerName, item_type: itemType }
    }));
}

function castSpell(school) {
    const player = gameState.Players.find(p => p.Name === playerName);
    if (!player) return;

    socket.send(JSON.stringify({
        type: "use_spell",
        data: {
            user: playerName,
            target: "Гоблин", // можно выбрать цель или использовать текущего монстра
            school: school
        }
    }));
}