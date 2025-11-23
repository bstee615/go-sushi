# Deployment Guide - Fly.io

This Sushi Go! game is deployed as a single application on Fly.io, which serves both the Svelte frontend and Go backend with WebSocket support.

## Prerequisites

Install Fly.io CLI:
```bash
# macOS/Linux
curl -L https://fly.io/install.sh | sh

# Windows (PowerShell)
iwr https://fly.io/install.ps1 -useb | iex
```

## Deploy to Fly.io

1. **Login to Fly.io**
   ```bash
   fly auth login
   ```

2. **Launch the app** (first time only)
   ```bash
   fly launch
   ```
   - Choose a unique app name (or let Fly generate one)
   - Select a region close to your users
   - Don't deploy yet when prompted

3. **Deploy the application**
   ```bash
   fly deploy
   ```
   
   The Dockerfile will automatically:
   - Build the Svelte + TypeScript frontend with Vite
   - Build the Go backend
   - Combine both into a single deployable image

4. **Open your app**
   ```bash
   fly open
   ```

Your app will be available at `https://your-app-name.fly.dev`

## What Gets Deployed

The deployment includes:
- **Frontend**: Production-built Svelte app with Tailwind CSS and PixiJS
  - Minified and optimized (164KB gzipped)
  - Served as static files from `/frontend/dist`
- **Backend**: Go WebSocket server
  - Handles game logic and multiplayer connections
  - Serves frontend files and WebSocket endpoint

## Configuration

The app is configured in `fly.toml`:
- Auto-scaling: Minimum 0 machines (auto-start on request)
- Internal port: 8080
- HTTPS enforced
- WebSocket support enabled

## Game Configuration

To customize game settings, update the backend startup command in `Dockerfile`:

```dockerfile
CMD ["./main", "-rounds", "5", "-cards", "8"]
```

Available flags:
- `-rounds N` - Number of rounds (default: 3)
- `-cards N` - Cards per hand (default: 10)
- `-port :PORT` - Server port (default: :8080)

## Local Development

### Backend
```bash
cd backend
go run main.go
```

### Frontend (Development Mode)
```bash
cd frontend
npm install
npm run dev
```
The dev server runs on `http://localhost:5173` and connects to backend on `ws://localhost:8080/ws`

### Frontend (Production Build)
```bash
cd frontend
npm run build
```
This creates optimized files in `frontend/dist/` that the backend serves

## Monitoring

- **View logs:** `fly logs`
- **Check status:** `fly status`
- **SSH into machine:** `fly ssh console`
- **View metrics:** `fly dashboard`

## Scaling

- **Scale machines:** `fly scale count 2`
- **Change machine size:** `fly scale vm shared-cpu-2x`

## Troubleshooting

If WebSocket connections fail:
- Check logs: `fly logs`
- Verify health check: `curl https://your-app.fly.dev/health`
- Restart app: `fly apps restart your-app-name`

If frontend doesn't load:
- Verify build completed: Check `fly logs` for build output
- Ensure frontend files exist: `fly ssh console` then `ls -la frontend/`
- Check frontend is being served: `curl https://your-app.fly.dev/`
