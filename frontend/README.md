# Sushi Go! Frontend

A modern frontend for the Sushi Go! game built with Svelte, Tailwind CSS, and PixiJS.

## Technologies

- **Svelte** - Reactive UI framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Utility-first CSS framework for UI components
- **PixiJS** - 2D WebGL renderer for the game canvas
- **Vite** - Fast build tool and dev server

## Features

- Clean, modern UI with Tailwind CSS
- Real-time multiplayer via WebSocket
- Interactive game canvas using PixiJS
- Card visualization with placeholder graphics
- Smooth animations and transitions
- Responsive design

## Development

### Prerequisites

- Node.js 18+ 
- Backend server running on `localhost:8080`

### Setup

```bash
cd frontend
npm install
```

### Run Development Server

```bash
npm run dev
```

The frontend will be available at `http://localhost:5173`

### Build for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

### Preview Production Build

```bash
npm run preview
```

## Project Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── websocket.ts          # WebSocket connection management
│   │   ├── LoginScreen.svelte     # Login/lobby screen
│   │   ├── PlayingScreen.svelte   # Main game screen
│   │   └── GameCanvas.svelte      # PixiJS game canvas
│   ├── App.svelte                 # Root component
│   ├── app.css                    # Global styles with Tailwind
│   └── main.ts                    # Application entry point
├── tailwind.config.js             # Tailwind configuration
├── vite.config.ts                 # Vite configuration
└── package.json
```

## Architecture

### WebSocket Communication

The `websocket.ts` store manages all WebSocket communication with the backend:
- Auto-connects on startup
- Auto-reconnects on disconnection
- Handles all message types (game_state, round_end, game_end, etc.)
- Provides reactive stores for connection status and game state

### UI Components

1. **LoginScreen** - Handles player login, game creation/joining, and displays active games list
2. **PlayingScreen** - Main game interface with:
   - Game canvas for card rendering
   - Player list and scores
   - Collection display
   - Game controls
3. **GameCanvas** - PixiJS-powered canvas that renders cards with:
   - Card placeholders with colors and emojis
   - Interactive card selection
   - Hover effects
   - Selection highlighting

### Styling

The app uses Tailwind CSS for all UI elements, providing:
- Responsive layouts
- Consistent color scheme matching the test-frontend
- Smooth transitions and animations
- Clean, modern design

### Game Canvas (PixiJS)

The game canvas uses PixiJS for high-performance 2D rendering:
- Cards are rendered as Graphics objects with rounded rectangles
- Each card type has a unique color
- Cards display emoji icons for visual identification
- Interactive hover and selection states
- Responsive layout that adapts to window size

## Differences from test-frontend

While maintaining the same functionality and layout as test-frontend, this new frontend offers:

1. **Modern Framework** - Built with Svelte for better performance and developer experience
2. **Component Architecture** - Modular, reusable components
3. **Type Safety** - Full TypeScript support
4. **Better Styling** - Tailwind CSS instead of vanilla CSS
5. **Canvas Rendering** - PixiJS for hardware-accelerated card rendering
6. **Better State Management** - Svelte stores for reactive state

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)

WebGL support is required for PixiJS rendering.

## Recommended IDE Setup

[VS Code](https://code.visualstudio.com/) + [Svelte](https://marketplace.visualstudio.com/items?itemName=svelte.svelte-vscode).
