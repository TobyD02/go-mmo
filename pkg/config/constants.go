// config - manages constants
package config

import "time"

// World / simulation dimensions, measured in tiles.
const ClientSimulationRangeX = 64
const ClientSimulationRangeY = 28

// Client viewport, measured in world tiles.
const BBTClientViewportTilesX = 64
const BBTClientViewportTilesY = 28

// Each world tile is rendered as 2 terminal columns.
const ClientTileWidth = 2

// Client update rates.
const ClientTickSpeed time.Duration = time.Millisecond * 50
const ServerTickSpeed time.Duration = time.Millisecond * 200

// Terminal layout.
const BBTClientRows = 40
const BBTClientColumns = BBTClientViewportTilesX + 2

// Spawn Protection
const SpawnProtectionRadius = 10
const SpawnProtectionRadiusSquared = SpawnProtectionRadius * SpawnProtectionRadius

// Chunking
const ChunkSize = 32

// Game World File
const GameWorldFilePath = "./data/world_new.json"
