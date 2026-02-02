# Flight Simulator Frontend

A modern React-based web dashboard for real-time aircraft monitoring and control.

## Features

- 🛩️ **Real-time Aircraft State Display**: Live telemetry including position, velocity, heading, and more
- 🎮 **Flight Command Interface**: Intuitive controls for GoTo, Trajectory, Stop, and Hold commands
- 📡 **Server-Sent Events (SSE)**: Automatic real-time updates from the backend
- 🔄 **Auto-reconnection**: Automatically reconnects if connection is lost
- 📱 **Responsive Design**: Works on desktop, tablet, and mobile devices
- 🌙 **Dark Theme**: Aviation-inspired dark interface

## Technology Stack

- **React 18** with TypeScript
- **Vite** for fast development and building
- **Server-Sent Events (SSE)** for real-time updates
- **CSS3** with modern layouts (Grid, Flexbox)

## Prerequisites

- Node.js 18+ and npm
- Running Flight Simulator backend (Go server on port 8080 by default)

## Installation

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

## Configuration

The frontend connects to the backend API at `http://localhost:8080` by default.

To customize the API URL, create a `.env` file in the `frontend` directory:

```env
VITE_API_URL=http://localhost:8080
```

## Development

Start the development server with hot module replacement:

```bash
npm run dev
```

The application will be available at `http://localhost:5173`

### Development Features
- ⚡ Hot Module Replacement (HMR)
- 🔍 TypeScript type checking
- 🎨 ESLint for code quality

## Building for Production

Create an optimized production build:

```bash
npm run build
```

The build output will be in the `dist` directory.

### Preview Production Build

Test the production build locally:

```bash
npm run preview
```

## Project Structure

```
frontend/
├── src/
│   ├── components/          # React components
│   │   ├── AircraftStateDisplay.tsx
│   │   ├── AircraftStateDisplay.css
│   │   ├── CommandPanel.tsx
│   │   └── CommandPanel.css
│   ├── hooks/              # Custom React hooks
│   │   └── useSSE.ts       # SSE connection hook
│   ├── services/           # API and services
│   │   └── api.ts          # Backend API client
│   ├── App.tsx             # Main application component
│   ├── App.css             # Application styles
│   ├── main.tsx            # Application entry point
│   └── index.css           # Global styles
├── public/                 # Static assets
├── index.html              # HTML template
├── package.json            # Dependencies and scripts
├── tsconfig.json           # TypeScript configuration
├── vite.config.ts          # Vite configuration
└── README.md              # This file
```

## Usage

### Aircraft State Display

The left panel shows real-time aircraft telemetry:
- **Position**: Latitude, Longitude, Altitude
- **Velocity**: North, East, Down components + Ground Speed
- **Flight Data**: Heading, Fuel (if available), Current Command
- **System**: Last update timestamp
- **Connection Status**: Live/Disconnected indicator

### Command Panel

The right panel provides three command interfaces:

#### 1. GoTo Commands
- Enter target latitude, longitude, and altitude
- Click "Send GoTo Command" to execute
- Example: `32.0853, 34.7818, 1000`

#### 2. Trajectory Commands
Two input methods:

**File Upload:**
- Click "Upload CSV File" and select a `.csv` file
- Format: `latitude,longitude,altitude` (one waypoint per line)

**Manual Input:**
- Enter waypoints directly in the text area
- Format: One waypoint per line as `lat,lon,alt`
- Example:
  ```
  32.0853,34.7818,1000
  32.0900,34.7900,1500
  32.1000,34.8000,2000
  ```

#### 3. Control Commands
- **STOP**: Immediately halt all movement and clear commands
- **HOLD**: Maintain current position and altitude

### Notifications

Toast notifications appear in the top-right corner:
- ✓ **Green**: Successful command execution
- ✗ **Red**: Command errors or failures
- Auto-dismiss after 5 seconds

## API Endpoints Used

The frontend communicates with the following backend endpoints:

- `GET /state` - Fetch current aircraft state
- `GET /stream` - SSE stream for real-time updates
- `POST /command/goto` - Send GoTo command
- `POST /command/trajectory` - Send Trajectory command
- `POST /command/stop` - Send Stop command
- `POST /command/hold` - Send Hold command

## Troubleshooting

### Backend Connection Issues

**Problem**: "Connection lost" or "Disconnected" status

**Solutions**:
1. Ensure the Go backend is running on port 8080
2. Check that the backend is accessible at `http://localhost:8080`
3. Verify CORS is configured correctly in the backend
4. Check browser console for detailed error messages

### CORS Errors

If you see CORS errors in the browser console, ensure your Go backend has CORS middleware configured to allow requests from `http://localhost:5173` (Vite dev server).

### Port Already in Use

If port 5173 is already in use, Vite will automatically try the next available port (5174, 5175, etc.).

To specify a custom port:
```bash
npm run dev -- --port 3000
```

### TypeScript Errors

Run type checking:
```bash
npm run build
```

This will show any TypeScript compilation errors.

## Development Tips

1. **Auto-reconnection**: The SSE connection automatically reconnects every 3 seconds if lost
2. **Form Validation**: All command inputs are validated before submission
3. **Loading States**: Buttons show loading state during API calls
4. **Error Handling**: All API errors are caught and displayed as notifications

## License

Same as the main Flight Simulator project.
