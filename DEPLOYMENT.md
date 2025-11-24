# Deployment Guide - Fly.io

This Sushi Go! game is deployed as a single application on Fly.io, which serves both the frontend and backend with WebSocket support.

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

4. **Open your app**
   ```bash
   fly open
   ```

Your app will be available at `https://your-app-name.fly.dev`

## Review Apps

Review apps are automatically deployed for Copilot pull requests when they are marked as ready for review. These are temporary preview environments for testing changes before merging.

### Automatic Deployment

- Review apps are created automatically via GitHub Actions workflow
- Each PR gets a unique app name: `go-sushi-pr-{number}`
- The PR comment will include the review app URL
- Apps are deployed to Fly.io with the same configuration as production

### Automatic Cleanup

Review apps are automatically scaled to 0 machines after 1 week of inactivity:
- A daily cleanup job runs at 2 AM UTC
- Only review apps (matching `go-sushi-pr-*` pattern) are affected
- Production app (`go-sushi`) is never scaled down by this job
- Apps are scaled to 0 (not deleted) to preserve data while reducing costs

### Manual Management

To manually manage review apps:

```bash
# List all review apps
fly apps list | grep "go-sushi-pr-"

# Check status of a review app
fly status --app go-sushi-pr-123

# Scale a review app back up
fly scale count 1 --app go-sushi-pr-123

# Delete a review app (optional)
fly apps destroy go-sushi-pr-123
```

## Configuration

The app is configured in `fly.toml`:
- Auto-scaling: Minimum 1 machine running
- Health checks on `/health` endpoint
- WebSocket support enabled
- HTTPS enforced

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

1. **Start backend:**
   ```bash
   cd backend
   go run main.go
   ```

2. **Open frontend:**
   - Open `test-frontend/index.html` in your browser
   - It will auto-connect to `ws://localhost:8080/ws`

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
