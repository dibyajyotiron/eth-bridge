# Bridging Events Backend

This project collects and processes bridging events from the SocketGate contract on the Ethereum Mainnet. It uses Go, PostgreSQL, Redis, and Go-Migrate for migrations.

### Notion (requirements doc): https://sockettech.notion.site/Assignment-Build-a-real-time-bridging-event-service-ee09be008661443c96425a5eefa21414

---

## Installation Steps

### Using Docker

#### Prerequisites

Before running the backend using docker, ensure you have the following installed:

1. **Docker** (https://www.docker.com/products/docker-desktop/)

Using docker, it's extremely simple.

#### Running

Just create `.env.docker` with the values ->

```sh
DATABASE_URL=postgres://user:password@postgres:5432/events?sslmode=disable
REDIS_URL=redis:6379
ETHEREUM_RPC_URL=wss://mainnet.infura.io/ws/v3/<project_id>
SOCKETGATE_CONTRACT=0x3a23F943181408EAC424116Af7b7790c94Cb97a5
SOCKET_TOPIC_HEX=0x74594da9e31ee4068e17809037db37db496702bf7d8d63afe6f97949277d1609
SERVER_PORT=:8080
REDIS_STREAM=bridging_events
REDIS_STREAM_DLQ=bridging_events_DLQ
```

Now, run ->
`docker-compose up`

Server should be up and running, migrations will run automatically

---

### Running locally

#### Prerequisites

Before running the backend, ensure you have the following installed:

1. **Go** (https://golang.org/doc/install)
2. **Go-Migrate** (https://github.com/golang-migrate/migrate)
3. **PostgreSQL** (https://www.postgresql.org/download)
4. **Redis** (https://redis.io/download)

#### 1. Clone the Repository

```bash
git clone https://github.com/dibyajyotiron/eth-bridge.git
cd bridging-events-backend
```

#### 2. Create a `.env` File

Create a `.env` file in the project root and set the following environment variables:

```dotenv
DATABASE_URL=postgres://localhost:5432/events?sslmode=disable
ETHEREUM_RPC_URL=wss://mainnet.infura.io/ws/v3/someproject
SOCKETGATE_CONTRACT=0x3a23F943181408EAC424116Af7b7790c94Cb97a5
SOCKET_TOPIC_HEX=0x74594da9e31ee4068e17809037db37db496702bf7d8d63afe6f97949277d1609
SERVER_PORT=:8080
REDIS_URL=localhost:6379
REDIS_STREAM="bridging_events"
REDIS_STREAM_DLQ="bridging_events_DLQ"
```

#### 3. Set Up PostgreSQL

Create a PostgreSQL database:

```sql
CREATE DATABASE bridging_events;
```

Or if you've `createdb` command installed,

```sh
createdb bridging_events;
```

#### 4. Run Database Migrations

Run migrations using `makefile`:

```bash
make migrate-up
```

Create new migrations using `makefile`:

```bash
make new-migration NAME=some_migration_name
```

#### 5. Install Dependencies

```bash
go mod tidy
```

#### 6. Run Redis Server

Start the Redis server:

```bash
redis-server
```

---

## Running the Backend

Start the backend server:

```bash
make run
```

The server should now be running at `http://localhost:8080`
(You can change the port in dotenv using `SERVER_PORT`).

---

## API Endpoints

### 1. Fetch Paginated Events

**GET** `/events`

| Query Parameter | Description                  | Example Value |
| --------------- | ---------------------------- | ------------- |
| `last_id`       | ID of the last fetched event | `10`          |
| `limit`         | Number of events per page    | `10`          |

**Param Details**

- `last_id`: First request is meant to be sent without `last_id`,`with limit`. You get `last_id` in the response. When you pass in, the same `last_id` in the next request, you get all items `after that last_id`. Events are ordered in `DESC` order. Last essentially means last item in `DESC` events list.
- `limit`: Number of items required per page. `defaults` to `10`. Maximum is `100`.

**Example Request**:

```bash
curl --location 'localhost:8080/api/v1/events?limit=50&last_id=3'
```

**Example Response**:

```json
{
  "events": [
    {
      "ID": 2,
      "Token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
      "Amount": "1372483935",
      "FromChain": "0x0041B0239420DebF7885433d09AE4f274d3d8AC3",
      "ToChain": "0x0041B0239420DebF7885433d09AE4f274d3d8AC3",
      "Timestamp": "2024-12-14T14:17:03.048677Z",
      "TransactionHash": "0x995f960af8eefc632cdd9b89b546f4069a4098b2b40fd25840048f59d5ee5106"
    },
    {
      "ID": 1,
      "Token": "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE",
      "Amount": "3482483968499194",
      "FromChain": "0x0e186b704783Ba103eE32723084eef498475d50B",
      "ToChain": "0x0e186b704783Ba103eE32723084eef498475d50B",
      "Timestamp": "2024-12-14T14:16:13.354036Z",
      "TransactionHash": "0x40d8a6b3b673c7d234e3da6d5061e51fb6cf9f3239fb2100ddc98742752266ff"
    }
  ],
  "last_id": 1,
  "next_timestamp": "2024-12-14 14:16:13.354036 +0000 UTC"
}
```

---

## Additional Commands

### Run Tests

```bash
go test ./...
```

---

## Further Improvement

1. We can implement retry queues, or retry failed events before sending them to DLQ (For simplicity and time constraint, decided to go with process and send to dlq)
2. We can have more function level documents and more comments around each function to ensure they're easily manageable.
3. Follow `SOLID` principles, although the code follows SOLID principle, but due to time constraint, there is only so much that can be followed.
4. Better logging.

---

## Troubleshooting

1. **Redis Error**: Ensure the Redis server is running on `localhost:6379`.

2. **Database Connection Issues**: Check your `DATABASE_URL` in the `.env` file and verify PostgreSQL is running.

3. **Ethereum RPC Issues**: Ensure your Ethereum node WebSocket URL is correct and accessible. Also, make sure to use `wss` to connect to ethereum otherwise you'll get `unsupported` method.
