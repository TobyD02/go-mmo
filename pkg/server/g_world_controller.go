package server

import (
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/tobyd02/go-mmo/pkg/game"
)

type GWorldController struct {
	GameWorld        *game.GameWorld
	getServerTick    func() int
	getMessageRouter func() *GMessageRouter

	changedPlayers       map[string]struct{}
	changedNpcs          map[string]struct{}
	changedInteractables map[string]struct{}
	changedTiles         map[game.Vec2]struct{}
}

func NewGWorldController(worldWidth, worldHeight int, getTicker func() int, getMessageRouter func() *GMessageRouter) *GWorldController {
	return &GWorldController{
		GameWorld:        game.NewGameWorld(worldWidth, worldHeight),
		getServerTick:    getTicker,
		getMessageRouter: getMessageRouter,

		changedPlayers: make(map[string]struct{}),
		changedNpcs:    make(map[string]struct{}),

		changedInteractables: make(map[string]struct{}),
		changedTiles:         make(map[game.Vec2]struct{}),
	}
}

func (wc *GWorldController) SetupWorld(
	edgeWalls bool,
) {

	npcRegistry, err := game.GetNpcRegistry()
	if err != nil {
		log.Fatalf("failed to get npc registry")
	}

	interactableRegistry, err := game.GetInteractableRegistry()
	if err != nil {
		log.Fatalf("failed to get npc registry")
	}

	for y, row := range wc.GameWorld.Map {
		for x := range row {
			if (x == 0 || x == wc.GameWorld.Width-1) || (y == 0 || y == wc.GameWorld.Height-1) {
				if edgeWalls {
					wc.GameWorld.Map[y][x] = game.TileWall
				}
			} else {
				// Not on a wall, so spawn an interactable (sometimes)

				if x != wc.GameWorld.SpawnPoint.X && y != wc.GameWorld.SpawnPoint.Y {
					randInt := rand.IntN(100)
					if randInt == 1 {
						interactableID, err := game.GetRandomIDFromRegistry(interactableRegistry)
						if err != nil {
							continue
						}
						wc.AddInteractable(game.NewGInteractableInstance(interactableID, x, y))
					} else if randInt == 2 {
						npcID, err := game.GetRandomIDFromRegistry(npcRegistry)
						if err != nil {
							continue
						}
						wc.AddNpc(game.NewGNpcInstance(npcID, x, y))
					}
				}
			}
		}
	}

}

func (wc *GWorldController) BuildWorldDiff() game.GameWorldDiff {
	diff := game.GameWorldDiff{
		PlayersDiff:       make(map[string]*game.GPlayer),
		NpcsDiff:          make(map[string]*game.GNpcInstance),
		InteractablesDiff: make(map[string]*game.GInteractableInstance),
	}

	for playerID := range wc.changedPlayers {
		player, exists := wc.GameWorld.Players[playerID]

		if exists {
			diff.PlayersDiff[playerID] = player
		} else {
			diff.PlayersDiff[playerID] = nil
		}
	}

	for npcID := range wc.changedNpcs {
		npc, exists := wc.GameWorld.Npcs[npcID]

		if exists {
			diff.NpcsDiff[npcID] = npc
		} else {
			diff.NpcsDiff[npcID] = nil
		}
	}

	for id := range wc.changedInteractables {
		interactable, exists := wc.GameWorld.Interactables[id]

		if exists {
			diff.InteractablesDiff[id] = interactable
		} else {
			diff.InteractablesDiff[id] = nil
		}
	}

	for pos := range wc.changedTiles {
		diff.MapDiff = append(
			diff.MapDiff,
			game.GameWorldMapDiff{
				Pos:  pos,
				Tile: wc.GameWorld.Map[pos.Y][pos.X],
			},
		)
	}

	clear(wc.changedPlayers)
	clear(wc.changedInteractables)
	clear(wc.changedTiles)

	return diff
}

func (wc *GWorldController) AddInteractable(interactable *game.GInteractableInstance) {
	if wc.GameWorld.QueryInteractableInstanceAtPosition(interactable.Pos.X, interactable.Pos.Y) != nil {
		return // Only a single interactable in a tile
	}
	if wc.GameWorld.QueryMap(interactable.Pos.X, interactable.Pos.Y) != game.TileWalkable {
		return // Only on walkable tiles, not inside a wall
	}

	wc.GameWorld.Interactables[interactable.ID] = interactable
	wc.changedInteractables[interactable.ID] = struct{}{}
}

func (wc *GWorldController) AddNpc(npc *game.GNpcInstance) {
	if wc.GameWorld.QueryMap(npc.Pos.X, npc.Pos.Y) != game.TileWalkable {
		return // Only on walkable tiles, not inside a wall
	}

	wc.GameWorld.Npcs[npc.ID] = npc
	wc.changedNpcs[npc.ID] = struct{}{}
}

func (wc *GWorldController) DeleteNpc(npcID string) {
	if _, exists := wc.GameWorld.Npcs[npcID]; !exists {
		return
	}

	delete(wc.GameWorld.Npcs, npcID)

	wc.changedNpcs[npcID] = struct{}{}
}

func (wc *GWorldController) SpawnNewPlayer(playerID string) error {
	return wc.AddPlayer(playerID, wc.GameWorld.SpawnPoint.X, wc.GameWorld.SpawnPoint.Y)
}

func (wc *GWorldController) AddPlayer(playerID string, x, y int) error {
	if wc.GameWorld.QueryMap(x, y) != game.TileWalkable {
		return fmt.Errorf("Cannot add player")
	}

	wc.addPlayer(game.NewGPlayer(playerID, x, y))

	wc.changedPlayers[playerID] = struct{}{}

	return nil
}

func (wc *GWorldController) DeletePlayer(playerID string) {
	if _, exists := wc.GameWorld.Players[playerID]; !exists {
		return
	}

	delete(wc.GameWorld.Players, playerID)

	wc.changedPlayers[playerID] = struct{}{}
}

func (wc *GWorldController) addPlayer(player *game.GPlayer) {
	wc.GameWorld.Players[player.ID] = player
}

func (wc *GWorldController) MovePlayer(client *GServerClient, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	player, exists := wc.GameWorld.Players[client.ID]
	if !exists { // Safe guard
		return
	}

	newX := player.Pos.X + dx
	newY := player.Pos.Y + dy

	if (player.Pos.Y+dy < 0 || player.Pos.Y+dy >= len(wc.GameWorld.Map)) ||
		(player.Pos.X+dx < 0 || player.Pos.X+dx >= len(wc.GameWorld.Map[player.Pos.Y+dy])) {
		return
	}

	if wc.GameWorld.QueryMap(newX, newY) != game.TileWalkable {
		return // Cannot move to unwalkable tile
	}

	if wc.GameWorld.QueryInteractableInstanceAtPosition(newX, newY) != nil {
		return // Cannot move over interactables?
	}

	player.Pos.X += dx
	player.Pos.Y += dy

	wc.changedPlayers[client.ID] = struct{}{}

	// wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("moved from %d to %d", player.Pos.X, player.Pos.Y))
	// log.Printf("WORLD | %s moved to x: %v y: %v", player.ID, player.Pos.X, player.Pos.Y)
}

func (wc *GWorldController) MoveNpc(npcInstanceID string, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	npc, exists := wc.GameWorld.Npcs[npcInstanceID]
	if !exists { // Safe guard
		return
	}

	newX := npc.Pos.X + dx
	newY := npc.Pos.Y + dy

	if (npc.Pos.Y+dy < 0 || npc.Pos.Y+dy >= len(wc.GameWorld.Map)) ||
		(npc.Pos.X+dx < 0 || npc.Pos.X+dx >= len(wc.GameWorld.Map[npc.Pos.Y+dy])) {
		return
	}

	if wc.GameWorld.QueryMap(newX, newY) != game.TileWalkable {
		return // Cannot move to unwalkable tile
	}

	if wc.GameWorld.QueryInteractableInstanceAtPosition(newX, newY) != nil {
		return // Cannot move over interactables?
	}

	npc.Pos.X += dx
	npc.Pos.Y += dy

	wc.changedNpcs[npcInstanceID] = struct{}{}

	// wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("moved from %d to %d", player.Pos.X, player.Pos.Y))
	// log.Printf("WORLD | %s moved to x: %v y: %v", player.ID, player.Pos.X, player.Pos.Y)
}

func (wc *GWorldController) InteractWith(client *GServerClient, interactableInstanceID string) {
	player, exists := wc.GameWorld.Players[client.ID]
	if !exists { // Safe guard
		return
	}

	interactableInstance, exists := wc.GameWorld.Interactables[interactableInstanceID]
	if !exists { // Safe guard
		return
	}

	if !interactableInstance.PlayerCanOccupyOrWork(player) {
		return
	}

	interactableInstance.DoWork(wc.getServerTick())

	interactableName := game.GetInteractableNameFromRegistry(interactableInstance.InteractableID)
	wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("Worked %s", interactableName))

	if interactableInstance.WorkIsDone() {
		yields := interactableInstance.GetYieldAndTriggerCooldown()
		player.AddToInventory(yields)

		// Added to player inventory, so player has changed
		wc.changedPlayers[player.ID] = struct{}{}

		for itemID, yield := range yields {
			itemName := game.GetItemNameFromRegistry(itemID)
			wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("Received %dx %s", yield, itemName))
		}

		if len(yields) == 0 {
			wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", "Received nothing")
		}
	}

	wc.changedInteractables[interactableInstanceID] = struct{}{}
}

func (wc *GWorldController) DoTickers() {
	// @todo - need to tick occupancy as well.
	// - if occupant hasn't worked this tick, it should be cleared

	for _, interactable := range wc.GameWorld.Interactables {

		// If its occupied and hasn't been worked this tick
		if interactable.IsOccupied() && !interactable.DidWorkThisTick(wc.getServerTick()) {
			if interactable.OccupantCooldown <= 0 {
				interactable.ClearOccupant()
			} else {
				interactable.OccupantCooldown--
			}

			wc.changedInteractables[interactable.ID] = struct{}{}
		}

		if interactable.CurrentTickCooldown <= 0 {
			continue
		}

		interactable.CurrentTickCooldown--

		if interactable.CurrentTickCooldown <= 0 {
			interactable.OccupiedBy = ""
			interactable.CurrentTicksWorked = 0
			log.Printf("WORLD | %s interactable was reset", interactable.ID)
		}

		wc.changedInteractables[interactable.ID] = struct{}{}
	}
}

func (wc *GWorldController) AttackNpc(client *GServerClient, npcInstanceID string) {
	player, exists := wc.GameWorld.Players[client.ID]
	if !exists { // Safe guard
		return
	}

	npcInstance, exists := wc.GameWorld.Npcs[npcInstanceID]
	if !exists { // Safe guard
		return
	}

	if !npcInstance.PlayerCanAttack(player) {
		return
	}

	dmg := npcInstance.TakePlayerDamage(player)
	npc := game.GetNpcFromRegistry(npcInstance.NpcID)

	npcName := ""
	npcMaxHealth := 0
	if npc != nil {
		npcName = npc.Name
		npcMaxHealth = npc.MaxHealth
	}

	wc.getMessageRouter().PushClientLogMessage(
		client.ID,
		"CLIENT",
		fmt.Sprintf("Dealt %d dmg to %s | %s health: %d/%d", dmg, npcName, npcName, npcInstance.Health, npcMaxHealth),
	)

	wc.changedNpcs[npcInstance.ID] = struct{}{}

	if npcInstance.Health <= 0 {
		npcLoot := npcInstance.GetLoot()
		player.AddToInventory(npcLoot)

		for itemID, amount := range npcLoot {
			itemName := game.GetItemNameFromRegistry(itemID)
			wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("Received %dx %s", amount, itemName))
		}

		wc.DeleteNpc(npcInstance.ID)
	}

}

func (wc *GWorldController) DoNpcs() {

	for npcInstanceID, npcInstance := range wc.GameWorld.Npcs {
		npc := game.GetNpcFromRegistry(npcInstance.NpcID)
		if npc == nil {
			continue
		}

		serverTick := wc.getServerTick()
		if !npcInstance.CanDoCombat(serverTick) && !npcInstance.CanDoPatrol(serverTick) {
			continue
		}

		playerTarget := wc.GameWorld.Players[npcInstance.PlayerTargetID]
		npcTarget := wc.GameWorld.Npcs[npcInstance.NpcTargetID]

		delta := npcInstance.Think(wc.getServerTick(), playerTarget, npcTarget)

		newX := npcInstance.Pos.X + delta.X
		newY := npcInstance.Pos.Y + delta.Y

		if free := wc.GameWorld.QueryInteractableInstanceAtPosition(newX, newY); free != nil {
			continue
		}

		if free := wc.GameWorld.QueryMap(newX, newY); free != game.TileWalkable {
			continue
		}

		if delta.X != 0 || delta.Y != 0 {
			wc.MoveNpc(npcInstanceID, delta.X, delta.Y)
			wc.changedNpcs[npcInstanceID] = struct{}{}
		}
	}
}
