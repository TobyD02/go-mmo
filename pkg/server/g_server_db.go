package server

import (
	"github.com/tobyd02/go-mmo/pkg/database"
	"github.com/tobyd02/go-mmo/pkg/game"
)

type GServerDB struct {
	db database.DB
}

func NewGServerDB(db database.DB) *GServerDB {
	return &GServerDB{db}
}

func (d *GServerDB) PlayerExists(playerID string) (bool, error) {
	err := d.db.Connect()
	if err != nil {
		return false, err
	}

	defer d.db.Close()

	playerRow := d.db.QueryRow("SELECT player_id, pos_x, pos_y FROM player WHERE player_id=?", playerID)
	var player game.GPlayer
	err = playerRow.Scan(&player.ID, &player.Pos.X, &player.Pos.Y)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (d *GServerDB) LoadPlayer(playerID string) (*game.GPlayer, error) {
	err := d.db.Connect()
	if err != nil {
		return nil, err
	}

	defer d.db.Close()

	playerRow := d.db.QueryRow("SELECT player_id, pos_x, pos_y FROM player WHERE player_id=?", playerID)
	inventoryRows, err := d.db.Query("SELECT item_name, amount FROM player_inventory WHERE player_id=?", playerID)

	if err != nil {
		return nil, err
	}

	var player game.GPlayer
	err = playerRow.Scan(&player.ID, &player.Pos.X, &player.Pos.Y)
	if err != nil {
		return nil, err
	}

	player.Inventory = make(map[string]int, 0)

	for inventoryRows.Next() {
		var name string
		var amount int

		if err := inventoryRows.Scan(&name, &amount); err != nil {
			return nil, err
		}

		player.AddToInventory(map[string]int{name: amount})
	}

	return &player, nil
}

func (d *GServerDB) SavePlayer(player *game.GPlayer) error {
	err := d.db.Connect()
	if err != nil {
		return err
	}

	defer d.db.Close()

	err = d.db.BeginTransaction()

	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			d.db.RollbackTransaction()
		}
	}()

	_, err = d.db.Exec(`
	INSERT INTO player(player_id, pos_x, pos_y)
	VALUES (?, ?, ?)
	ON CONFLICT(player_id)
	DO UPDATE SET 
		pos_x = excluded.pos_x, 
		pos_y = excluded.pos_y
	`, player.ID, player.Pos.X, player.Pos.Y)

	if err != nil {
		return err
	}

	// delete player inventory in db first
	_, err = d.db.Exec(
		"DELETE FROM player_inventory WHERE player_id = ?",
		player.ID,
	)
	if err != nil {
		return err
	}

	for itemName, amount := range player.Inventory {
		if amount <= 0 {
			continue
		}

		_, err = d.db.Exec(`
		INSERT INTO player_inventory (
			player_id, item_name, amount
		)
		VALUES (?, ?, ?)
		`, player.ID, itemName, amount)

		if err != nil {
			return err
		}
	}

	err = d.db.CommitTransaction()
	if err != nil {
		return err
	}

	committed = true
	return nil
}

func (d *GServerDB) Init() error {
	err := d.db.Connect()
	if err != nil {
		return err
	}

	defer d.db.Close()

	_, err = d.db.Exec(`
	CREATE TABLE IF NOT EXISTS player(
		player_id TEXT PRIMARY KEY NOT NULL,
		pos_x INTEGER NOT NULL,
		pos_y INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS player_inventory(
		player_id TEXT NOT NULL,
		item_name TEXT NOT NULL,
		amount INTEGER NOT NULL,

		PRIMARY KEY (player_id, item_name),

		FOREIGN KEY(player_id)
			REFERENCES player(player_id)
			ON DELETE CASCADE
	);
	`)

	return err
}
