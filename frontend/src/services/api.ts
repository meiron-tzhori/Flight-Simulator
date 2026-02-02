// API service for Flight Simulator backend communication

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface AircraftState {
  position: {
    latitude: number;
    longitude: number;
    altitude: number;
  };
  velocity: {
    north: number;
    east: number;
    down: number;
  };
  heading: number;
  timestamp: string;
  currentCommand?: string;
  fuel?: number;
}

export interface GoToCommand {
  latitude: number;
  longitude: number;
  altitude: number;
}

export interface TrajectoryPoint {
  latitude: number;
  longitude: number;
  altitude: number;
  timestamp?: string;
}

export interface ApiResponse<T = void> {
  success: boolean;
  data?: T;
  error?: string;
}

class FlightSimulatorAPI {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  /**
   * Fetch current aircraft state
   */
  async getState(): Promise<AircraftState> {
    const response = await fetch(`${this.baseUrl}/state`);
    if (!response.ok) {
      throw new Error(`Failed to fetch state: ${response.statusText}`);
    }
    return response.json();
  }

  /**
   * Send GoTo command
   */
  async sendGoToCommand(command: GoToCommand): Promise<ApiResponse> {
    const response = await fetch(`${this.baseUrl}/command/goto`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        lat: command.latitude,
        lon: command.longitude,
        alt: command.altitude,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`GoTo command failed: ${error}`);
    }

    return {
      success: true,
      data: await response.json().catch(() => undefined),
    };
  }

  /**
   * Send Trajectory command
   */
  async sendTrajectoryCommand(points: TrajectoryPoint[]): Promise<ApiResponse> {
    const response = await fetch(`${this.baseUrl}/command/trajectory`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        waypoints: points.map(p => ({
          lat: p.latitude,
          lon: p.longitude,
          alt: p.altitude,
        })),
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Trajectory command failed: ${error}`);
    }

    return {
      success: true,
      data: await response.json().catch(() => undefined),
    };
  }

  /**
   * Send Stop command
   */
  async sendStopCommand(): Promise<ApiResponse> {
    const response = await fetch(`${this.baseUrl}/command/stop`, {
      method: 'POST',
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Stop command failed: ${error}`);
    }

    return {
      success: true,
      data: await response.json().catch(() => undefined),
    };
  }

  /**
   * Send Hold command
   */
  async sendHoldCommand(): Promise<ApiResponse> {
    const response = await fetch(`${this.baseUrl}/command/hold`, {
      method: 'POST',
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Hold command failed: ${error}`);
    }

    return {
      success: true,
      data: await response.json().catch(() => undefined),
    };
  }

  /**
   * Get SSE stream URL
   */
  getStreamUrl(): string {
    return `${this.baseUrl}/stream`;
  }
}

// Export singleton instance
export const api = new FlightSimulatorAPI();
