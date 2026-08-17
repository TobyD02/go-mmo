# Concept

## MVP

- Basic go server (Completely server authoratative)
  - Establishes websocket connections
  - Receives position delta's from clients
  - Each tick, updates world state, and then distributes state to clients

- Basic web clients
  - Establishes websocket connection to server
  - Sends position delta based on input (or random)
  - Receives world state on ticker and updates what its drawing.
