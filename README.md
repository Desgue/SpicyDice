# SpicyDice Documentation
## 📖 Overview
SpicyDice is a high-performance betting game server that enables real-time dice gambling through WebSocket connections. Built with Go, it features:
- Real-time Gameplay: Instant bet placement and results via WebSocket
- Session Management: Secure single-session system per player
- Persistent Storage: PostgreSQL-backed transaction history
- Clean Architecture: Maintainable and testable codebase

## Prerequisites
- Docker
- Docker Compose
- Make (optional)

## Quick Start with Make
### Full Application Deployment
```bash
# Build and start entire application
make up
```

### Alternative Deployment Options
```bash
# Start in background (detached mode)
make up-d
# Stop application
make down
# Restart application
make restart
# Restart application cleaning the volume
make restart-v
```

## Manual Docker Compose Deployment
### Standard Deployment
```bash
# Build and start services
docker-compose up --build
# Start in background
docker-compose up --build -d
```

## Service Management Commands
### With Make
```bash
# View running containers
make ps
# View all logs
make logs
# View server logs
make logs-server
# View database logs
make logs-db
```

### With Docker Compose
```bash
# View running containers
docker-compose ps
# View logs
docker-compose logs
# Stop and remove all services
docker-compose down
```

## Access Points
- Web Application: `http://localhost:8080`
- Database: `localhost:5432`
  - Username: postgres
  - Password: p4ssw0rd

## Cleanup
```bash
# Remove containers, networks, and images
make clean
```


**Note**: Credentials are for development. Modify for production use.

## WebSocket API
### Connection
```
ws://localhost:8080/ws
```

### Message Structure
```json
{
  "type": "string",
  "payload": {}
}
```

### Endpoints
#### 1. Get Wallet Balance
```json
{
  "type": "wallet",
  "payload": {
    "client_id": 1
  }
}
```
Response:
```json
{
  "client_id": 1,
  "balance": 100.00
}
```

#### 2. Place Bet
```json
{
  "type": "play",
  "payload": {
    "client_id": 1,
    "bet_amount": 10.00,
    "bet_type": "even"  // "even" or "odd"
  }
}
```
Response:
```json
{
  "dice_result": 4,
  "won": true,
  "balance": 90.00
}
```

#### 3. End Play Session
```json
{
  "type": "endplay",
  "payload": {
    "client_id": 1
  }
}
```
Response:
```json
{
  "client_id": 1,
}
```

## Game Rules
- Bet amounts: Min $1.00, Max $1,000.00
- Win multiplier: 2x
- Single active session per player
- 6-sided dice
- Even/Odd betting only

## Error Handling
- Insufficient funds
- Invalid bet amount
- Active session conflicts
- User not found
- Invalid message types

## Architecture
### Core Components
- Clean Architecture pattern
- Services layer for game logic
- Repository layer for data persistence
- PostgreSQL for data storage

## Testing
The project includes basic test coverage as a proof of concept, demonstrating:
- Unit testing methodologies
- Integration testing patterns
- Golang table driven testing practices

For production deployment, additional testing would be required:
- Increase test coverage
- More in depth security testing
- Load and stress testing

## Security Considerations
### Client Identification
- Current implementation uses integer Client IDs
- Production should implement UUIDs to prevent ID spoofing
- Session management should be enhanced with proper authentication

### Data Management
- Production environment should separate game_session table into:
  - game_play: Individual play records
  - game_session: Session management
- Implement comprehensive transaction logging
- Double-entry bookkeeping for financial transactions
- Audit trail for all gaming activities

## Concurrency and Performance
The current implementation provides:
- Basic WebSocket server with concurrent connections
- Session management for multiple users

Production enhancements should include:
- Enhanced state management in WebSocket messages
- Robust connection handling
- Rate limiting
- Advanced error recovery
- Comprehensive load testing

## CI/CD
Basic CI/CD pipeline is implemented using GitHub Actions to demonstrate:
- Automated testing
- Build processes
- Deployment workflows


