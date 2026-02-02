import { useEffect, useState } from 'react';
import type { AircraftState } from '../services/api';

interface UseSSEReturn {
  state: AircraftState | null;
  isConnected: boolean;
  error: string | null;
}

export function useSSE(url: string): UseSSEReturn {
  const [state, setState] = useState<AircraftState | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let eventSource: EventSource | null = null;
    let reconnectTimeout: NodeJS.Timeout;

    const connect = () => {
      try {
        eventSource = new EventSource(url);

        eventSource.onopen = () => {
          console.log('SSE connected');
          setIsConnected(true);
          setError(null);
        };

        // Listen for 'state' events specifically (backend sends event: state)
        eventSource.addEventListener('state', (event: MessageEvent) => {
          try {
            const data: AircraftState = JSON.parse(event.data);
            console.log('Received state update:', data);
            setState(data);
          } catch (err) {
            console.error('SSE parse error:', err, 'Data:', event.data);
            setError('Failed to parse state data');
          }
        });

        // Also listen for 'connected' event
        eventSource.addEventListener('connected', (event: MessageEvent) => {
          console.log('Connection confirmed:', event.data);
        });

        eventSource.onerror = () => {
          console.error('SSE connection error, attempting reconnect...');
          setIsConnected(false);
          setError('Connection lost');
          eventSource?.close();
          
          // Attempt reconnection after 3 seconds
          reconnectTimeout = setTimeout(connect, 3000);
        };
      } catch (err) {
        console.error('Failed to create EventSource:', err);
        setError('Failed to connect to server');
        setIsConnected(false);
      }
    };

    connect();

    return () => {
      clearTimeout(reconnectTimeout);
      eventSource?.close();
      setIsConnected(false);
    };
  }, [url]);

  return { state, isConnected, error };
}
