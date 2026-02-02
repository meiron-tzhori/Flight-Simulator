import { useEffect, useState } from 'react'

interface AircraftState {
  position: {
    latitude: number
    longitude: number
    altitude: number
  }
  velocity: {
    groundSpeed: number
    verticalSpeed: number
  }
  heading: number
  timestamp: string
}

export function useSSE(): AircraftState | null {
  const [state, setState] = useState<AircraftState | null>(null)

  useEffect(() => {
    const eventSource = new EventSource('/stream')

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        setState(data)
      } catch (error) {
        console.error('SSE parse error:', error)
      }
    }

    eventSource.onerror = () => {
      console.error('SSE connection lost, reconnecting...')
    }

    return () => {
      eventSource.close()
    }
  }, [])

  return state
}
