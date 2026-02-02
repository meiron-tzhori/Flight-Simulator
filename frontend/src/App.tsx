import { useState } from 'react';
import { AircraftStateDisplay } from './components/AircraftStateDisplay';
import { CommandPanel } from './components/CommandPanel';
import { useSSE } from './hooks/useSSE';
import { api } from './services/api';
import './App.css';

function App() {
  const { state, isConnected, error: sseError } = useSSE(api.getStreamUrl());
  const [notifications, setNotifications] = useState<Array<{ id: number; message: string; type: 'success' | 'error' }>>([]);
  const [notificationId, setNotificationId] = useState(0);

  const addNotification = (message: string, type: 'success' | 'error') => {
    const id = notificationId;
    setNotificationId(id + 1);
    setNotifications(prev => [...prev, { id, message, type }]);

    // Auto-remove notification after 5 seconds
    setTimeout(() => {
      setNotifications(prev => prev.filter(n => n.id !== id));
    }, 5000);
  };

  const handleCommandSent = (command: string) => {
    addNotification(`✓ ${command}`, 'success');
  };

  const handleCommandError = (error: string) => {
    addNotification(`✗ ${error}`, 'error');
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>✈️ Flight Simulator Dashboard</h1>
        <div className="header-subtitle">
          Real-time aircraft monitoring and control
        </div>
      </header>

      {/* Notifications */}
      {notifications.length > 0 && (
        <div className="notifications">
          {notifications.map(notification => (
            <div
              key={notification.id}
              className={`notification notification-${notification.type}`}
            >
              {notification.message}
            </div>
          ))}
        </div>
      )}

      {/* SSE Connection Error */}
      {sseError && (
        <div className="error-banner">
          <strong>Connection Error:</strong> {sseError}
        </div>
      )}

      <main className="app-main">
        <div className="dashboard-grid">
          {/* Aircraft State Display */}
          <section className="dashboard-section">
            <AircraftStateDisplay state={state} isConnected={isConnected} />
          </section>

          {/* Command Panel */}
          <section className="dashboard-section">
            <CommandPanel
              onCommandSent={handleCommandSent}
              onError={handleCommandError}
            />
          </section>
        </div>
      </main>

      <footer className="app-footer">
        <div>Flight Simulator v1.0</div>
        <div>Backend API: {api.getStreamUrl().replace('/stream', '')}</div>
      </footer>
    </div>
  );
}

export default App;
