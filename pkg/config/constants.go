// config - manages constants
package config

import "time"

// World / simulation dimensions, measured in tiles.
const ClientSimulationRangeX = 32
const ClientSimulationRangeY = 14

// Client viewport, measured in world tiles.
const ClientViewportTilesX = 64
const ClientViewportTilesY = 28

// Each world tile is rendered as 2 terminal columns.
const ClientTileWidth = 2

// Client update rates.
const ClientTickSpeed time.Duration = time.Millisecond * 50
const ServerTickSpeed time.Duration = time.Millisecond * 200

// Terminal layout.
const ClientRows = 40
const ClientColumns = ClientViewportTilesX + 2

// Spawn Protection
const SpawnProtectionRadius = 10
const SpawnProtectionRadiusSquared = SpawnProtectionRadius * SpawnProtectionRadius

// Chunking
const ChunkSize = 32

// Game World File
const GameWorldFilePath = "./data/world_new.json"
