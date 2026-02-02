import { useEffect, useState } from 'react';
import { AircraftState } from '../services/api';

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

        eventSource.onmessage = (event) => {
          try {
            const data: AircraftState = JSON.parse(event.data);
            setState(data);
          } catch (err) {
            console.error('SSE parse error:', err);
            setError('Failed to parse state data');
          }
        };

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
