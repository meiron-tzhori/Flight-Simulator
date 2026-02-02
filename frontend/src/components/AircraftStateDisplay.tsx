import type { AircraftState } from '../services/api';
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
            {isConnected ? '\u25cf Connected' : '\u25cb Disconnected'}
          </span>
        </div>
        <div className="state-loading">
          {isConnected ? 'Waiting for data...' : 'Disconnected from server'}
        </div>
      </div>
    );
  }

  const formatNumber = (value: number | undefined, decimals: number = 2): string => {
    if (value === undefined || value === null) return 'N/A';
    return value.toFixed(decimals);
  };

  return (
    <div className="aircraft-state">
      <div className="state-header">
        <h2>Aircraft State</h2>
        <span className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
          {isConnected ? '\u25cf Live' : '\u25cb Disconnected'}
        </span>
      </div>

      <div className="state-grid">
        {/* Position Section */}
        <div className="state-section">
          <h3>Position</h3>
          <div className="state-item">
            <span className="label">Latitude:</span>
            <span className="value">{formatNumber(state.position?.latitude, 6)}\u00b0</span>
          </div>
          <div className="state-item">
            <span className="label">Longitude:</span>
            <span className="value">{formatNumber(state.position?.longitude, 6)}\u00b0</span>
          </div>
          <div className="state-item">
            <span className="label">Altitude:</span>
            <span className="value">{formatNumber(state.position?.altitude, 2)} m</span>
          </div>
        </div>

        {/* Velocity Section */}
        <div className="state-section">
          <h3>Velocity</h3>
          <div className="state-item state-item-highlight">
            <span className="label">Ground Speed:</span>
            <span className="value">{formatNumber(state.velocity?.ground_speed, 2)} m/s</span>
          </div>
          <div className="state-item">
            <span className="label">Vertical Speed:</span>
            <span className="value">{formatNumber(state.velocity?.vertical_speed, 2)} m/s</span>
          </div>
        </div>

        {/* Flight Data Section */}
        <div className="state-section">
          <h3>Flight Data</h3>
          <div className="state-item">
            <span className="label">Heading:</span>
            <span className="value">{formatNumber(state.heading, 1)}\u00b0</span>
          </div>
          {state.active_command && (
            <div className="state-item state-item-highlight">
              <span className="label">Active Command:</span>
              <span className="value command-value">{state.active_command.type}</span>
            </div>
          )}
          {state.active_command?.eta_seconds !== undefined && (
            <div className="state-item">
              <span className="label">ETA:</span>
              <span className="value">{Math.round(state.active_command.eta_seconds)}s</span>
            </div>
          )}
        </div>

        {/* Environment Section */}
        {state.environment && (
          <div className="state-section">
            <h3>Environment</h3>
            {state.environment.wind && (
              <>
                <div className="state-item">
                  <span className="label">Wind Direction:</span>
                  <span className="value">{formatNumber(state.environment.wind.direction, 0)}\u00b0</span>
                </div>
                <div className="state-item">
                  <span className="label">Wind Speed:</span>
                  <span className="value">{formatNumber(state.environment.wind.speed, 1)} m/s</span>
                </div>
              </>
            )}
            {state.environment.humidity !== undefined && (
              <div className="state-item">
                <span className="label">Humidity:</span>
                <span className="value">{formatNumber(state.environment.humidity, 0)}%</span>
              </div>
            )}
          </div>
        )}

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
