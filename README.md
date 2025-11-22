# Sushi Go! Game

A multiplayer implementation of the Sushi Go! card game with a Go backend and React frontend.

## Backend Dependencies

- **Go 1.21+**
- **gorilla/websocket** (v1.5.1) - WebSocket implementation
- **gopter** (v0.2.9) - Property-based testing library

## Frontend Dependencies

- **React 18+** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling framework
- **DaisyUI** - Component library
- **fast-check** - Property-based testing library

## Getting Started

### Backend

```bash
cd backend
go mod download
go run main.go
```

The server will start on `http://localhost:8080`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:5173`

## Development

The project follows a client-server architecture where:
- The Go backend manages authoritative game state
- The React frontend provides the user interface
- WebSocket connections enable real-time multiplayer gameplay

## Testing

### Backend Tests
```bash
cd backend
go test ./...
```

### Frontend Tests
```bash
cd frontend
npm test
```

## License

MIT
