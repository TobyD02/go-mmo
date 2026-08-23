package server

import (
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/util"
)

type GWorldController struct {
	GameWorld        *game.GameWorld
	getServerTick    func() int
	getMessageRouter func() *GMessageRouter

	changedPlayers       map[string]struct{}
	changedNpcs          map[string]struct{}
	changedInteractables map[string]struct{}
	changedTiles         map[util.Vec2]struct{}

	npcSpatialIndex          util.GSpatialIndex
	interactableSpatialIndex util.GSpatialIndex
	playerSpatialIndex       util.GSpatialIndex
	// @todo - at the moment interactables cannot move, however if they ever do then the spatial index will need updating
}

func NewGWorldController(worldWidth, worldHeight int, getTicker func() int, getMessageRouter func() *GMessageRouter) *GWorldController {
	gameWorld, err := game.NewGameWorld("./data/world.txt")
	if err != nil {
		panic(err)
	}

	return &GWorldController{
		GameWorld:        gameWorld,
		getServerTick:    getTicker,
		getMessageRouter: getMessageRouter,

		changedPlayers:       make(map[string]struct{}),
		changedNpcs:          make(map[string]struct{}),
		changedInteractables: make(map[string]struct{}),
		changedTiles:         make(map[util.Vec2]struct{}),

		npcSpatialIndex:          util.NewGSpatialIndex(),
		interactableSpatialIndex: util.NewGSpatialIndex(),
		playerSpatialIndex:       util.NewGSpatialIndex(),
	}
}

func (wc *GWorldController) SetupWorld() {
	npcRegistry, err := game.GetNpcRegistry()
	if err != nil {
		log.Fatalf("failed to get npc registry")
	}

	interactableRegistry, err := game.GetInteractableRegistry()
	if err != nil {
		log.Fatalf("failed to get npc registry")
	}

	log.Println("Starting world generation")

	for y, row := range wc.GameWorld.Map {
		for x, tile := range row {

			if tile == game.TileWall {
				continue
			}

			if x != wc.GameWorld.SpawnPoint.X && y != wc.GameWorld.SpawnPoint.Y {
				randInt := rand.IntN(100)
				if randInt == 1 {
					interactableID, err := util.GetRandomIDFromRegistry(interactableRegistry)
					if err != nil {
						continue
					}
					wc.AddInteractableInstance(game.NewGInteractableInstance(interactableID, x, y))
				} else if randInt == 2 {
					npcID, err := util.GetRandomIDFromRegistry(npcRegistry)
					if err != nil {
						continue
					}
					wc.AddNpcInstance(game.NewGNpcInstance(npcID, x, y))
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
	clear(wc.changedNpcs)
	clear(wc.changedInteractables)
	clear(wc.changedTiles)

	return diff
}

func (wc *GWorldController) AddInteractableInstance(interactableInstance *game.GInteractableInstance) {
	if len(wc.interactableSpatialIndex.QueryPos(interactableInstance.Pos.X, interactableInstance.Pos.Y)) != 0 {
		return // Only a single interactableInstance in a tile
	}
	if !game.CanWalk(wc.GameWorld.QueryMap(interactableInstance.Pos.X, interactableInstance.Pos.Y)) {
		return // Only on walkable tiles, not inside a wall
	}

	wc.GameWorld.Interactables[interactableInstance.ID] = interactableInstance
	wc.interactableSpatialIndex.Add(interactableInstance.ID, interactableInstance.Pos)
	wc.changedInteractables[interactableInstance.ID] = struct{}{}
}

func (wc *GWorldController) AddNpcInstance(npcInstance *game.GNpcInstance) {
	if !game.CanWalk(wc.GameWorld.QueryMap(npcInstance.Pos.X, npcInstance.Pos.Y)) {
		return // Only on walkable tiles, not inside a wall
	}

	wc.GameWorld.Npcs[npcInstance.ID] = npcInstance
	wc.npcSpatialIndex.Add(npcInstance.ID, npcInstance.Pos)
	wc.changedNpcs[npcInstance.ID] = struct{}{}
}

func (wc *GWorldController) DeleteNpc(npcInstanceID string) {
	npcInstance, exists := wc.GameWorld.Npcs[npcInstanceID]

	if !exists {
		return
	}

	delete(wc.GameWorld.Npcs, npcInstance.ID)
	wc.npcSpatialIndex.Remove(npcInstance.ID, npcInstance.Pos)

	wc.changedNpcs[npcInstance.ID] = struct{}{}
}

func (wc *GWorldController) SpawnNewPlayer(playerID string) error {
	return wc.AddPlayer(playerID, wc.GameWorld.SpawnPoint.X, wc.GameWorld.SpawnPoint.Y)
}

func (wc *GWorldController) AddPlayer(playerID string, x, y int) error {
	// can only spawn on walkable tiles
	if !game.CanWalk(wc.GameWorld.QueryMap(x, y)) {
		return fmt.Errorf("Cannot add player")
	}

	wc.addPlayer(game.NewGPlayer(playerID, x, y))

	wc.playerSpatialIndex.Add(playerID, util.Vec2{X: x, Y: y})
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

	if !game.CanWalk(wc.GameWorld.QueryMap(newX, newY)) {
		return // Cannot move to unwalkable tile
	}

	if len(wc.interactableSpatialIndex.QueryPos(newX, newY)) > 0 {
		return // Cannot move over interactables?
	}

	playerOldPos := player.Pos

	player.Pos.X += dx
	player.Pos.Y += dy

	wc.changedPlayers[client.ID] = struct{}{}
	wc.playerSpatialIndex.Update(client.ID, playerOldPos, player.Pos)
}

func (wc *GWorldController) SetTile(pos util.Vec2, tile game.GameWorldTile) {
	if len(wc.playerSpatialIndex.QueryPos(pos.X, pos.Y)) > 0 {
		return
	}

	if len(wc.npcSpatialIndex.QueryPos(pos.X, pos.Y)) > 0 {
		return
	}

	// If within spawn radius of spawn point then dont allow
	if wc.GameWorld.SpawnPoint.DistanceSquared(pos) <= config.SpawnProtectionRadiusSquared {
		return
	}

	wc.GameWorld.Map[pos.Y][pos.X] = tile
	wc.changedTiles[pos] = struct{}{}
}

func (wc *GWorldController) MoveNpc(npcInstanceID string, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	npcInstance, exists := wc.GameWorld.Npcs[npcInstanceID]
	if !exists { // Safe guard
		return
	}

	newX := npcInstance.Pos.X + dx
	newY := npcInstance.Pos.Y + dy

	if (npcInstance.Pos.Y+dy < 0 || npcInstance.Pos.Y+dy >= len(wc.GameWorld.Map)) ||
		(npcInstance.Pos.X+dx < 0 || npcInstance.Pos.X+dx >= len(wc.GameWorld.Map[npcInstance.Pos.Y+dy])) {
		return
	}

	if !game.CanWalk(wc.GameWorld.QueryMap(newX, newY)) {
		return // Cannot move to unwalkable tile
	}

	if len(wc.interactableSpatialIndex.QueryPos(newX, newY)) > 0 {
		return // Cannot move over interactables?
	}

	oldPos := npcInstance.Pos

	npcInstance.Pos.X += dx
	npcInstance.Pos.Y += dy

	wc.changedNpcs[npcInstanceID] = struct{}{}
	wc.npcSpatialIndex.Update(npcInstance.ID, oldPos, npcInstance.Pos)

	// wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("moved from %d to %d", player.Pos.X, player.Pos.Y))
	// log.Printf("WORLD | %s moved to x: %v y: %v", player.ID, player.Pos.X, player.Pos.Y)
}

func (wc *GWorldController) QuickNpcInstancesQueryAtPos(x, y int) map[string]*game.GNpcInstance {
	ids := wc.npcSpatialIndex.QueryPos(x, y)
	instances := make(map[string]*game.GNpcInstance)
	for instanceID := range ids {
		instances[instanceID] = wc.GameWorld.Npcs[instanceID]
	}

	return instances
}

func (wc *GWorldController) QuickInteractableInstancesQueryAtPos(x, y int) *game.GInteractableInstance {
	ids := wc.interactableSpatialIndex.QueryPos(x, y)
	for instanceID := range ids {
		return wc.GameWorld.Interactables[instanceID]
	}

	return nil
}

func (wc *GWorldController) QuickPlayersQueryAtPos(x, y int) map[string]*game.GPlayer {
	ids := wc.playerSpatialIndex.QueryPos(x, y)
	instances := make(map[string]*game.GPlayer)
	for instanceID := range ids {
		instances[instanceID] = wc.GameWorld.Players[instanceID]
	}

	return instances
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

func (wc *GWorldController) DoInteractables(interactableInstanceIDs map[string]struct{}) {
	for interactableInstanceID := range interactableInstanceIDs {
		interactable := wc.GameWorld.Interactables[interactableInstanceID]
		if interactable == nil {
			continue
		}

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

func (wc *GWorldController) DoNpcs(npcInstanceIDs map[string]struct{}) {

	for npcInstanceID := range npcInstanceIDs {
		npcInstance := wc.GameWorld.Npcs[npcInstanceID]
		if npcInstance == nil {
			continue
		}

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

		if len(wc.interactableSpatialIndex.QueryPos(newX, newY)) > 0 {
			continue
		}

		if !game.CanWalk(wc.GameWorld.QueryMap(newX, newY)) {
			continue
		}

		wc.MoveNpc(npcInstanceID, delta.X, delta.Y)
		wc.changedNpcs[npcInstanceID] = struct{}{}
	}
}
