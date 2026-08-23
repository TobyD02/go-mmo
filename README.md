# Concept

**Known Issues**

- Large number of clients connecting when world size is _**very**_ big can cause issues.
  - Most likely because we are packaging the full world payload.
  - Could only send chunk diffs - would require a lot of rewriting
    - Would also make the whole async world connection kind of redundant
  - probably a smart fix somewhere
  - For now though - stick to 1000x1000 world (or smaller if need be)
    - Plan is to eventually spawn multiple servers anyway - optimizations on server were secondary solution

## Building (web)

1. Runs the POC ebit client in web browser (http://localhost:8081)

```bash
go run github.com/hajimehoshi/wasmserve@latest -http=localhost:8081 ./cmd/ebit_client/
```

2. Alternatively, Running the following command will run a native ebit client

```
go run cmd/ebit_client/
```

## Running client swarm

1. Run server and build image

```bash
docker compose up --build -d

```

2. Run replicas clients

```bash
docker service create \
  --name mmo-clients \
  --replicas 100 \
  --env G_SERVER=ws://host.docker.internal:8080 \
  --entrypoint headless_client \
  mmo:latest
```

3. Commands:

```
# Scale (increase or decrease number of mock clients)
docker service scale mmo-clients:{{ number_of_clients }}

# Delete the clients
docker service rm mmo-clients
```

## TODO

- [ ] Generate items - i.e. some database that is serialised
- [ ] Serialise data. relational database or something?
- [ ] Simple Auth - user's login with an ID, which retrieves their data. Rather than random uuid each time a client is connected.

# @TODO - Replace all that below

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

### MVP 3

- Gameplay interactions.
  - I think proof of concept needs 1 additional tile type that is interactable.
  - Client can send an interaction action to the server
  - Server determines whether the client can interact
  - At some point, the interactable tile's state switches, perhaps some tick cool down to refresh?

- Persistent Storage
  - Store player state?
  - Change login from a random uuid to a typed? then when client loads in it either loads existing data or new data.
