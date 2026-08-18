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

### MVP 2

shared:

- World State Type

flow: 1. Client connects, receives full world state from server 2. Clients send messages to server 3. On tick, server reads messages, updates its world state, and packages a `diff` 4. Clients receive `diff` from the server, and update their own world state.
