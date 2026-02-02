import { AircraftState } from '../services/api';
import './AircraftStateDisplay.css';

interface AircraftStateDisplayProps {
  state: AircraftState | null;
  isConnected: boolean;
}

export function AircraftStateDisplay({ state, isConnected }: AircraftStateDisplayProps) {
  if (!state) {
    return (
      <div className="aircraft-state">
        <div className="state-header">
          <h2>Aircraft State</h2>
          <span className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '● Connected' : '○ Disconnected'}
          </span>
        </div>
        <div className="state-loading">
          {isConnected ? 'Waiting for data...' : 'Disconnected from server'}
        </div>
      </div>
    );
  }

  const formatCoordinate = (value: number, decimals: number = 6): string => {
    return value.toFixed(decimals);
  };

  const formatVelocity = (vel: { north: number; east: number; down: number }): number => {
    return Math.sqrt(vel.north ** 2 + vel.east ** 2 + vel.down ** 2);
  };

  return (
    <div className="aircraft-state">
      <div className="state-header">
        <h2>Aircraft State</h2>
        <span className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
          {isConnected ? '● Live' : '○ Disconnected'}
        </span>
      </div>

      <div className="state-grid">
        {/* Position Section */}
        <div className="state-section">
          <h3>Position</h3>
          <div className="state-item">
            <span className="label">Latitude:</span>
            <span className="value">{formatCoordinate(state.position.latitude)}°</span>
          </div>
          <div className="state-item">
            <span className="label">Longitude:</span>
            <span className="value">{formatCoordinate(state.position.longitude)}°</span>
          </div>
          <div className="state-item">
            <span className="label">Altitude:</span>
            <span className="value">{formatCoordinate(state.position.altitude, 2)} m</span>
          </div>
        </div>

        {/* Velocity Section */}
        <div className="state-section">
          <h3>Velocity</h3>
          <div className="state-item">
            <span className="label">North:</span>
            <span className="value">{formatCoordinate(state.velocity.north, 2)} m/s</span>
          </div>
          <div className="state-item">
            <span className="label">East:</span>
            <span className="value">{formatCoordinate(state.velocity.east, 2)} m/s</span>
          </div>
          <div className="state-item">
            <span className="label">Down:</span>
            <span className="value">{formatCoordinate(state.velocity.down, 2)} m/s</span>
          </div>
          <div className="state-item state-item-highlight">
            <span className="label">Ground Speed:</span>
            <span className="value">{formatCoordinate(formatVelocity(state.velocity), 2)} m/s</span>
          </div>
        </div>

        {/* Flight Data Section */}
        <div className="state-section">
          <h3>Flight Data</h3>
          <div className="state-item">
            <span className="label">Heading:</span>
            <span className="value">{formatCoordinate(state.heading, 1)}°</span>
          </div>
          {state.fuel !== undefined && (
            <div className="state-item">
              <span className="label">Fuel:</span>
              <span className="value">{formatCoordinate(state.fuel, 1)}%</span>
            </div>
          )}
          {state.currentCommand && (
            <div className="state-item state-item-highlight">
              <span className="label">Current Command:</span>
              <span className="value command-value">{state.currentCommand}</span>
            </div>
          )}
        </div>

        {/* Timestamp Section */}
        <div className="state-section">
          <h3>System</h3>
          <div className="state-item">
            <span className="label">Last Update:</span>
            <span className="value timestamp-value">
              {new Date(state.timestamp).toLocaleTimeString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
