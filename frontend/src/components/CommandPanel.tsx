import { useState } from 'react';
import { api } from '../services/api';
import type { GoToCommand, TrajectoryPoint } from '../services/api';
import './CommandPanel.css';

interface CommandPanelProps {
  onCommandSent?: (command: string) => void;
  onError?: (error: string) => void;
}

export function CommandPanel({ onCommandSent, onError }: CommandPanelProps) {
  const [activeTab, setActiveTab] = useState<'goto' | 'trajectory' | 'control'>('goto');
  const [isLoading, setIsLoading] = useState(false);

  // GoTo command state
  const [gotoLat, setGotoLat] = useState('');
  const [gotoLon, setGotoLon] = useState('');
  const [gotoAlt, setGotoAlt] = useState('');

  // Trajectory state
  const [trajectoryInput, setTrajectoryInput] = useState('');
  const [trajectoryFile, setTrajectoryFile] = useState<File | null>(null);

  const handleGoToSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const command: GoToCommand = {
        latitude: parseFloat(gotoLat),
        longitude: parseFloat(gotoLon),
        altitude: parseFloat(gotoAlt),
      };

      await api.sendGoToCommand(command);
      onCommandSent?.(`GoTo: ${gotoLat}, ${gotoLon}, ${gotoAlt}m`);
      
      // Clear form
      setGotoLat('');
      setGotoLon('');
      setGotoAlt('');
    } catch (error) {
      onError?.(error instanceof Error ? error.message : 'Failed to send GoTo command');
    } finally {
      setIsLoading(false);
    }
  };

  const handleTrajectorySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      let points: TrajectoryPoint[] = [];

      if (trajectoryFile) {
        // Parse CSV file
        const text = await trajectoryFile.text();
        points = parseCSV(text);
      } else if (trajectoryInput) {
        // Parse manual input
        points = parseCSV(trajectoryInput);
      } else {
        throw new Error('Please provide trajectory points via CSV or manual input');
      }

      await api.sendTrajectoryCommand(points);
      onCommandSent?.(`Trajectory: ${points.length} waypoints`);
      
      // Clear form
      setTrajectoryInput('');
      setTrajectoryFile(null);
    } catch (error) {
      onError?.(error instanceof Error ? error.message : 'Failed to send Trajectory command');
    } finally {
      setIsLoading(false);
    }
  };

  const parseCSV = (text: string): TrajectoryPoint[] => {
    const lines = text.trim().split('\n');
    const points: TrajectoryPoint[] = [];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line || line.startsWith('#')) continue; // Skip empty lines and comments

      const parts = line.split(',').map(p => p.trim());
      if (parts.length < 3) {
        throw new Error(`Invalid CSV format at line ${i + 1}: expected lat,lon,alt`);
      }

      points.push({
        latitude: parseFloat(parts[0]),
        longitude: parseFloat(parts[1]),
        altitude: parseFloat(parts[2]),
      });
    }

    if (points.length === 0) {
      throw new Error('No valid trajectory points found');
    }

    return points;
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setTrajectoryFile(file);
      setTrajectoryInput(''); // Clear manual input if file is selected
    }
  };

  const handleStopCommand = async () => {
    setIsLoading(true);
    try {
      await api.sendStopCommand();
      onCommandSent?.('STOP command sent');
    } catch (error) {
      onError?.(error instanceof Error ? error.message : 'Failed to send Stop command');
    } finally {
      setIsLoading(false);
    }
  };

  const handleHoldCommand = async () => {
    setIsLoading(true);
    try {
      await api.sendHoldCommand();
      onCommandSent?.('HOLD command sent');
    } catch (error) {
      onError?.(error instanceof Error ? error.message : 'Failed to send Hold command');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="command-panel">
      <h2>Flight Commands</h2>

      <div className="tab-buttons">
        <button
          className={`tab-button ${activeTab === 'goto' ? 'active' : ''}`}
          onClick={() => setActiveTab('goto')}
        >
          GoTo
        </button>
        <button
          className={`tab-button ${activeTab === 'trajectory' ? 'active' : ''}`}
          onClick={() => setActiveTab('trajectory')}
        >
          Trajectory
        </button>
        <button
          className={`tab-button ${activeTab === 'control' ? 'active' : ''}`}
          onClick={() => setActiveTab('control')}
        >
          Control
        </button>
      </div>

      <div className="tab-content">
        {activeTab === 'goto' && (
          <form onSubmit={handleGoToSubmit} className="command-form">
            <div className="form-group">
              <label htmlFor="goto-lat">Latitude (°)</label>
              <input
                id="goto-lat"
                type="number"
                step="any"
                value={gotoLat}
                onChange={(e) => setGotoLat(e.target.value)}
                placeholder="e.g., 32.0853"
                required
                disabled={isLoading}
              />
            </div>

            <div className="form-group">
              <label htmlFor="goto-lon">Longitude (°)</label>
              <input
                id="goto-lon"
                type="number"
                step="any"
                value={gotoLon}
                onChange={(e) => setGotoLon(e.target.value)}
                placeholder="e.g., 34.7818"
                required
                disabled={isLoading}
              />
            </div>

            <div className="form-group">
              <label htmlFor="goto-alt">Altitude (m)</label>
              <input
                id="goto-alt"
                type="number"
                step="any"
                value={gotoAlt}
                onChange={(e) => setGotoAlt(e.target.value)}
                placeholder="e.g., 1000"
                required
                disabled={isLoading}
              />
            </div>

            <button type="submit" className="submit-button" disabled={isLoading}>
              {isLoading ? 'Sending...' : 'Send GoTo Command'}
            </button>
          </form>
        )}

        {activeTab === 'trajectory' && (
          <form onSubmit={handleTrajectorySubmit} className="command-form">
            <div className="form-group">
              <label htmlFor="trajectory-file">Upload CSV File</label>
              <input
                id="trajectory-file"
                type="file"
                accept=".csv"
                onChange={handleFileChange}
                disabled={isLoading}
              />
              {trajectoryFile && (
                <div className="file-info">Selected: {trajectoryFile.name}</div>
              )}
            </div>

            <div className="form-divider">OR</div>

            <div className="form-group">
              <label htmlFor="trajectory-input">Manual Input (CSV format)</label>
              <textarea
                id="trajectory-input"
                value={trajectoryInput}
                onChange={(e) => {
                  setTrajectoryInput(e.target.value);
                  setTrajectoryFile(null); // Clear file if manual input is used
                }}
                placeholder="Enter waypoints (lat,lon,alt per line):\n32.0853,34.7818,1000\n32.0900,34.7900,1500\n32.1000,34.8000,2000"
                rows={6}
                disabled={isLoading}
              />
              <div className="form-hint">Format: latitude,longitude,altitude (one per line)</div>
            </div>

            <button type="submit" className="submit-button" disabled={isLoading}>
              {isLoading ? 'Sending...' : 'Send Trajectory'}
            </button>
          </form>
        )}

        {activeTab === 'control' && (
          <div className="control-buttons">
            <button
              onClick={handleStopCommand}
              className="control-button stop-button"
              disabled={isLoading}
            >
              🛑 STOP
            </button>
            <p className="control-description">
              Immediately stop all movement and clear current command
            </p>

            <button
              onClick={handleHoldCommand}
              className="control-button hold-button"
              disabled={isLoading}
            >
              ⏸️ HOLD
            </button>
            <p className="control-description">
              Hold current position (maintain altitude and coordinates)
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
