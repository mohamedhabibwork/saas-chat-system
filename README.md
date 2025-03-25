# Chat Application

A real-time chat application built with Go, featuring WebSocket support, multi-tenant architecture, and various chat features including bots, channels, and customizable tabs.

## Features

- Real-time messaging using WebSocket
- Multi-tenant support
- User authentication and authorization
- Group chat functionality
- Topic-based notifications
- Bot integration
- Channel management
- Customizable tabs for users
- RESTful API endpoints
- Swagger API documentation

## Prerequisites

- Go 1.16 or higher
- PostgreSQL 12 or higher
- Make (optional, for using Makefile commands)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/chat-app.git
cd chat-app
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the root directory with the following variables:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=messenger
PORT=8080
```

4. Run the application:
```bash
go run main.go
```

## API Documentation

The API documentation is available at `/swagger/index.html` when the server is running.

### Main Endpoints

#### WebSocket
- `GET /ws` - WebSocket connection endpoint

#### Tenants
- `POST /api/tenants` - Create a new tenant
- `GET /api/tenants` - List all tenants
- `GET /api/tenants/{id}` - Get tenant details
- `DELETE /api/tenants/{id}` - Delete a tenant

#### Users
- `POST /api/users/register` - Register a new user
- `GET /api/users` - List users in a tenant
- `GET /api/users/{id}` - Get user details
- `DELETE /api/users/{id}` - Delete a user

#### Groups
- `POST /api/groups` - Create a new group
- `GET /api/groups` - List groups in a tenant
- `GET /api/groups/{id}` - Get group details
- `DELETE /api/groups/{id}` - Delete a group
- `POST /api/groups/{id}/members` - Add member to group
- `DELETE /api/groups/{id}/members/{userId}` - Remove member from group

#### Topics
- `POST /api/topics` - Create a new topic
- `GET /api/topics` - List topics in a tenant
- `GET /api/topics/{id}` - Get topic details
- `DELETE /api/topics/{id}` - Delete a topic
- `POST /api/topics/{id}/subscribe` - Subscribe to topic
- `DELETE /api/topics/{id}/subscribe` - Unsubscribe from topic

#### Bots
- `POST /api/bots` - Create a new bot
- `GET /api/bots/list` - List bots in a tenant
- `DELETE /api/bots/delete` - Delete a bot

#### Channels
- `POST /api/channels` - Create a new channel
- `GET /api/channels/list` - List channels in a tenant
- `POST /api/channels/join` - Join a channel
- `DELETE /api/channels/leave` - Leave a channel

#### Tabs
- `POST /api/tabs` - Create a new tab
- `GET /api/tabs/list` - List user's tabs
- `PUT /api/tabs/order` - Update tab order
- `DELETE /api/tabs/delete` - Delete a tab

## Testing

Run the test suite:
```bash
go test ./...
```

## Development

### Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── controllers/
│   │   ├── base_controller.go
│   │   ├── bot_controller.go
│   │   ├── channel_controller.go
│   │   └── tab_controller.go
│   ├── models/
│   │   └── models.go
│   ├── database/
│   │   └── database.go
│   └── websocket/
│       └── websocket.go
├── pkg/
│   └── swagger/
│       └── swagger.go
├── tests/
│   └── integration/
├── .env
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 