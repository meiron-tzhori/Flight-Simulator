# Flight Simulator API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Base URL](#base-url)
3. [Authentication](#authentication)
4. [Common Response Codes](#common-response-codes)
5. [Error Response Format](#error-response-format)
6. [Endpoints](#endpoints)
   - [Health Check](#health-check)
   - [Submit Go-To Command](#submit-go-to-command)
   - [Submit Trajectory Command](#submit-trajectory-command)
   - [Submit Stop Command](#submit-stop-command-bonus)
   - [Submit Hold Command](#submit-hold-command-bonus)
   - [Get Aircraft State](#get-aircraft-state)
   - [Stream Aircraft State](#stream-aircraft-state-bonus)
7. [Data Models](#data-models)
8. [Examples](#examples)
9. [Rate Limits](#rate-limits)
10. [Versioning](#versioning)

---

## Overview

The Flight Simulator API provides REST endpoints to control a simulated aircraft and query its state. The API supports:

- **Command submission**: Direct the aircraft to fly to specific coordinates or follow a trajectory
- **State queries**: Get current aircraft position, velocity, and heading
- **State streaming**: Receive real-time state updates via Server-Sent Events (SSE)

All endpoints accept and return JSON payloads (except SSE streaming which uses `text/event-stream`).

---

## Base URL

```
http://localhost:8080
```

For production deployments, replace with your actual host and port.

---

## Authentication

**Current version**: No authentication required (demo/assignment purposes)

**Future**: Bearer token authentication may be added for production use.

---

## Common Response Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid request parameters |
| 404 | Not Found | Endpoint not found |
| 422 | Unprocessable Entity | Valid request but cannot be processed (e.g., terrain conflict) |
| 500 | Internal Server Error | Unexpected server error |
| 503 | Service Unavailable | Service temporarily unavailable (e.g., command queue full) |

---

## Error Response Format

All errors return a consistent JSON structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error description",
    "field": "field_name",
    "value": "invalid_value",
    "details": {}
  }
}
```

**Fields**:
- `code`: Machine-readable error code (e.g., `INVALID_COORDINATES`)
- `message`: Human-readable error message
- `field`: (Optional) The field that caused the error
- `value`: (Optional) The invalid value provided
- `details`: (Optional) Additional context

**Example**:
```json
{
  "error": {
    "code": "INVALID_LATITUDE",
    "message": "Latitude must be between -90 and 90 degrees",
    "field": "lat",
    "value": 120.5
  }
}
```

---

## Endpoints

### Health Check

**Description**: Check if the service is running and healthy.

**Endpoint**: `GET /health`

**Request**: None

**Response** (200 OK):
```json
{
  "status": "healthy",
  "simulation_running": true,
  "tick_rate_hz": 30.0,
  "uptime_seconds": 3600,
  "timestamp": "2026-02-01T19:00:00Z"
}
```

**Response Fields**:
- `status`: `"healthy"` or `"degraded"`
- `simulation_running`: Boolean indicating if simulation loop is active
- `tick_rate_hz`: Current simulation tick rate in Hertz
- `uptime_seconds`: Service uptime in seconds
- `timestamp`: Current server time (ISO 8601)

**Curl Example**:
```bash
curl http://localhost:8080/health
```

---

### Submit Go-To Command

**Description**: Direct the aircraft to fly to a specific geographic point.

**Endpoint**: `POST /command/goto`

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "lat": 32.0853,
  "lon": 34.7818,
  "alt": 1000.0,
  "speed": 100.0
}
```

**Request Fields**:
- `lat` (required): Target latitude in degrees (-90 to 90)
- `lon` (required): Target longitude in degrees (-180 to 180)
- `alt` (required): Target altitude in meters MSL (Mean Sea Level), must be ≥ 0
- `speed` (optional): Desired ground speed in m/s (default: configured default speed)

**Response** (200 OK):
```json
{
  "status": "accepted",
  "command_id": "cmd-a3f8b2c1",
  "message": "Go-to command accepted",
  "target": {
    "latitude": 32.0853,
    "longitude": 34.7818,
    "altitude": 1000.0
  },
  "eta_seconds": 120.5
}
```

**Response Fields**:
- `status`: `"accepted"` (command queued successfully)
- `command_id`: Unique identifier for this command
- `message`: Human-readable confirmation
- `target`: Echo of the target coordinates
- `eta_seconds`: Estimated time to reach target in seconds

**Error Responses**:

**400 Bad Request** - Invalid coordinates:
```json
{
  "error": {
    "code": "INVALID_LATITUDE",
    "message": "Latitude must be between -90 and 90 degrees",
    "field": "lat",
    "value": 120.5
  }
}
```

**503 Service Unavailable** - Command queue full:
```json
{
  "error": {
    "code": "QUEUE_FULL",
    "message": "Command queue is full, please retry"
  }
}
```

**422 Unprocessable Entity** - Terrain conflict (bonus feature):
```json
{
  "error": {
    "code": "TERRAIN_CONFLICT",
    "message": "Target altitude below terrain safety margin",
    "details": {
      "terrain_altitude": 800.0,
      "safety_margin": 100.0,
      "minimum_altitude": 900.0,
      "requested_altitude": 500.0
    }
  }
}
```

**Curl Examples**:

```bash
# Basic go-to command
curl -X POST http://localhost:8080/command/goto \
  -H "Content-Type: application/json" \
  -d '{
    "lat": 32.0853,
    "lon": 34.7818,
    "alt": 1000.0
  }'

# Go-to with custom speed
curl -X POST http://localhost:8080/command/goto \
  -H "Content-Type: application/json" \
  -d '{
    "lat": 32.0853,
    "lon": 34.7818,
    "alt": 1000.0,
    "speed": 150.0
  }'
```

---

### Submit Trajectory Command

**Description**: Direct the aircraft to follow a sequence of waypoints.

**Endpoint**: `POST /command/trajectory`

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "waypoints": [
    {
      "lat": 32.0853,
      "lon": 34.7818,
      "alt": 1000.0,
      "speed": 100.0
    },
    {
      "lat": 32.1000,
      "lon": 34.8000,
      "alt": 1500.0,
      "speed": 120.0
    },
    {
      "lat": 32.1200,
      "lon": 34.8200,
      "alt": 2000.0
    }
  ],
  "loop": false
}
```

**Request Fields**:
- `waypoints` (required): Array of waypoint objects (minimum 1)
  - `lat` (required): Waypoint latitude in degrees (-90 to 90)
  - `lon` (required): Waypoint longitude in degrees (-180 to 180)
  - `alt` (required): Waypoint altitude in meters MSL, must be ≥ 0
  - `speed` (optional): Speed to use when flying to this waypoint (m/s)
- `loop` (optional): If `true`, loop back to first waypoint after completing trajectory (default: `false`)

**Response** (200 OK):
```json
{
  "status": "accepted",
  "command_id": "cmd-d7e2c9f4",
  "message": "Trajectory command accepted",
  "waypoint_count": 3,
  "total_distance_meters": 25432.8,
  "eta_seconds": 254.3
}
```

**Response Fields**:
- `status`: `"accepted"`
- `command_id`: Unique identifier for this command
- `message`: Human-readable confirmation
- `waypoint_count`: Number of waypoints in trajectory
- `total_distance_meters`: Total flight path distance
- `eta_seconds`: Estimated time to complete trajectory

**Error Responses**:

**400 Bad Request** - Empty waypoint list:
```json
{
  "error": {
    "code": "EMPTY_WAYPOINTS",
    "message": "Trajectory must contain at least one waypoint",
    "field": "waypoints"
  }
}
```

**400 Bad Request** - Invalid waypoint:
```json
{
  "error": {
    "code": "INVALID_WAYPOINT",
    "message": "Waypoint 2: altitude cannot be negative",
    "field": "waypoints[1].alt",
    "value": -100.0
  }
}
```

**Curl Examples**:

```bash
# Simple trajectory
curl -X POST http://localhost:8080/command/trajectory \
  -H "Content-Type: application/json" \
  -d '{
    "waypoints": [
      {"lat": 32.0853, "lon": 34.7818, "alt": 1000.0},
      {"lat": 32.1000, "lon": 34.8000, "alt": 1500.0},
      {"lat": 32.1200, "lon": 34.8200, "alt": 2000.0}
    ]
  }'

# Looping trajectory with custom speeds
curl -X POST http://localhost:8080/command/trajectory \
  -H "Content-Type: application/json" \
  -d '{
    "waypoints": [
      {"lat": 32.0853, "lon": 34.7818, "alt": 1000.0, "speed": 80.0},
      {"lat": 32.1000, "lon": 34.8000, "alt": 1500.0, "speed": 120.0},
      {"lat": 32.1200, "lon": 34.8200, "alt": 2000.0, "speed": 100.0}
    ],
    "loop": true
  }'
```

---

### Submit Stop Command (Bonus)

**Description**: Immediately stop the aircraft at its current position.

**Endpoint**: `POST /command/stop`

**Request**: None (empty body)

**Response** (200 OK):
```json
{
  "status": "accepted",
  "command_id": "cmd-f1a9c8d2",
  "message": "Aircraft stopped at current position",
  "position": {
    "latitude": 32.0950,
    "longitude": 34.7900,
    "altitude": 1250.0
  }
}
```

**Curl Example**:
```bash
curl -X POST http://localhost:8080/command/stop
```

---

### Submit Hold Command (Bonus)

**Description**: Command the aircraft to hold its current position (orbit/circle).

**Endpoint**: `POST /command/hold`

**Request**: None (empty body)

**Response** (200 OK):
```json
{
  "status": "accepted",
  "command_id": "cmd-b5d3e7a1",
  "message": "Aircraft holding position",
  "hold_position": {
    "latitude": 32.0950,
    "longitude": 34.7900,
    "altitude": 1250.0
  },
  "orbit_radius_meters": 500.0
}
```

**Curl Example**:
```bash
curl -X POST http://localhost:8080/command/hold
```

---

### Get Aircraft State

**Description**: Query the current state of the aircraft.

**Endpoint**: `GET /state`

**Request**: None

**Response** (200 OK):
```json
{
  "position": {
    "latitude": 32.0853,
    "longitude": 34.7818,
    "altitude": 1000.0
  },
  "velocity": {
    "ground_speed": 100.0,
    "vertical_speed": 5.0
  },
  "heading": 45.0,
  "timestamp": "2026-02-01T19:00:00.123Z",
  "active_command": {
    "type": "goto",
    "target": {
      "latitude": 32.1000,
      "longitude": 34.8000,
      "altitude": 1500.0
    },
    "eta_seconds": 95.2
  },
  "environment": {
    "wind": {
      "direction": 270.0,
      "speed": 10.0
    },
    "humidity": 65.0
  }
}
```

**Response Fields**:
- `position`: Current geographic position
  - `latitude`: Degrees (-90 to 90)
  - `longitude`: Degrees (-180 to 180)
  - `altitude`: Meters MSL
- `velocity`: Current velocity vector
  - `ground_speed`: Speed over ground in m/s
  - `vertical_speed`: Climb/descent rate in m/s (positive = climbing)
- `heading`: Direction of flight in degrees (0-360, 0 = North, 90 = East)
- `timestamp`: State timestamp (ISO 8601 with milliseconds)
- `active_command`: Currently executing command (null if none)
  - `type`: Command type (`"goto"`, `"trajectory"`, `"hold"`, `"stop"`)
  - `target`: Target coordinates (for goto/trajectory)
  - `eta_seconds`: Estimated time to completion
- `environment`: Environmental conditions (bonus, null if disabled)
  - `wind`: Wind vector
    - `direction`: Wind direction in degrees
    - `speed`: Wind speed in m/s
  - `humidity`: Relative humidity percentage (0-100)

**Curl Examples**:

```bash
# Get current state
curl http://localhost:8080/state

# Pretty-print with jq
curl -s http://localhost:8080/state | jq .

# Extract just position
curl -s http://localhost:8080/state | jq '.position'

# Watch state continuously (every 1 second)
watch -n 1 'curl -s http://localhost:8080/state | jq .'
```

---

### Stream Aircraft State (Bonus)

**Description**: Subscribe to real-time aircraft state updates via Server-Sent Events (SSE).

**Endpoint**: `GET /stream`

**Request**: None

**Response Headers**:
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Access-Control-Allow-Origin: *
```

**Event Stream Format**:

Each state update is sent as an SSE event:

```
data: {"position":{"latitude":32.0853,"longitude":34.7818,"altitude":1000.0},"velocity":{"ground_speed":100.0,"vertical_speed":5.0},"heading":45.0,"timestamp":"2026-02-01T19:00:00.123Z"}

data: {"position":{"latitude":32.0854,"longitude":34.7819,"altitude":1001.0},"velocity":{"ground_speed":100.0,"vertical_speed":5.0},"heading":45.0,"timestamp":"2026-02-01T19:00:00.223Z"}

data: {"position":{"latitude":32.0855,"longitude":34.7820,"altitude":1002.0},"velocity":{"ground_speed":100.0,"vertical_speed":5.0},"heading":45.0,"timestamp":"2026-02-01T19:00:00.323Z"}

...
```

**Event Fields**: Same as [Get Aircraft State](#get-aircraft-state) response

**Update Frequency**: Configurable (default: 10 Hz = 10 updates per second)

**Connection Management**:
- Server sends periodic heartbeat comments to keep connection alive
- Client disconnection automatically unsubscribes
- Reconnection supported (client should handle)

**Curl Example**:

```bash
# Stream to console
curl -N http://localhost:8080/stream

# Stream and parse with jq (one JSON per line)
curl -N http://localhost:8080/stream | while read line; do
  if [[ $line == data:* ]]; then
    echo "${line:6}" | jq -c '.position'
  fi
done
```

**JavaScript Example**:

```javascript
const eventSource = new EventSource('http://localhost:8080/stream');

eventSource.onmessage = (event) => {
  const state = JSON.parse(event.data);
  console.log('Position:', state.position);
  console.log('Speed:', state.velocity.ground_speed);
  console.log('Heading:', state.heading);
};

eventSource.onerror = (error) => {
  console.error('SSE Error:', error);
  eventSource.close();
};

// Close connection when done
// eventSource.close();
```

**Python Example**:

```python
import requests
import json

url = 'http://localhost:8080/stream'
response = requests.get(url, stream=True)

for line in response.iter_lines():
    if line:
        decoded_line = line.decode('utf-8')
        if decoded_line.startswith('data:'):
            data = json.loads(decoded_line[5:])
            print(f"Position: {data['position']}")
            print(f"Heading: {data['heading']}°")
```

---

## Data Models

### Position

```json
{
  "latitude": 32.0853,
  "longitude": 34.7818,
  "altitude": 1000.0
}
```

**Fields**:
- `latitude`: Decimal degrees, -90 (South Pole) to 90 (North Pole)
- `longitude`: Decimal degrees, -180 (West) to 180 (East)
- `altitude`: Meters above Mean Sea Level (MSL)

**Validation**:
- Latitude: -90 ≤ lat ≤ 90
- Longitude: -180 ≤ lon ≤ 180
- Altitude: alt ≥ 0

### Velocity

```json
{
  "ground_speed": 100.0,
  "vertical_speed": 5.0
}
```

**Fields**:
- `ground_speed`: Speed over ground in meters per second (m/s)
- `vertical_speed`: Vertical rate of climb (+) or descent (-) in m/s

**Reference**:
- 100 m/s ≈ 360 km/h ≈ 194 knots
- Typical airliner cruise: 250 m/s (900 km/h)
- Typical climb rate: 5-10 m/s

### Waypoint

```json
{
  "lat": 32.0853,
  "lon": 34.7818,
  "alt": 1000.0,
  "speed": 100.0
}
```

**Fields**:
- `lat`: Latitude (required)
- `lon`: Longitude (required)
- `alt`: Altitude (required)
- `speed`: Speed to use approaching this waypoint (optional)

---

## Examples

### Complete Flight Scenario

This example demonstrates a complete flight from startup to landing:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

# 1. Check health
echo "1. Checking simulator health..."
curl -s $BASE_URL/health | jq .

# 2. Get initial state
echo "\n2. Getting initial aircraft state..."
curl -s $BASE_URL/state | jq .

# 3. Command: Take off and climb to 1000m
echo "\n3. Taking off to 1000m altitude..."
curl -s -X POST $BASE_URL/command/goto \
  -H "Content-Type: application/json" \
  -d '{
    "lat": 32.0853,
    "lon": 34.7818,
    "alt": 1000.0,
    "speed": 50.0
  }' | jq .

sleep 5

# 4. Check state after takeoff
echo "\n4. Checking state after 5 seconds..."
curl -s $BASE_URL/state | jq '.position, .velocity'

# 5. Fly a triangle trajectory
echo "\n5. Flying triangular pattern..."
curl -s -X POST $BASE_URL/command/trajectory \
  -H "Content-Type: application/json" \
  -d '{
    "waypoints": [
      {"lat": 32.0853, "lon": 34.7818, "alt": 1500.0, "speed": 100.0},
      {"lat": 32.1053, "lon": 34.7818, "alt": 1500.0, "speed": 100.0},
      {"lat": 32.0953, "lon": 34.8018, "alt": 1500.0, "speed": 100.0},
      {"lat": 32.0853, "lon": 34.7818, "alt": 1500.0, "speed": 100.0}
    ]
  }' | jq .

sleep 30

# 6. Check position mid-trajectory
echo "\n6. Position during trajectory..."
curl -s $BASE_URL/state | jq '.position, .active_command'

# 7. Emergency stop
echo "\n7. Emergency stop!..."
curl -s -X POST $BASE_URL/command/stop | jq .

sleep 2

# 8. Hold position
echo "\n8. Holding position..."
curl -s -X POST $BASE_URL/command/hold | jq .

sleep 5

# 9. Resume to destination and descend
echo "\n9. Descending to landing..."
curl -s -X POST $BASE_URL/command/goto \
  -H "Content-Type: application/json" \
  -d '{
    "lat": 32.0853,
    "lon": 34.7818,
    "alt": 100.0,
    "speed": 40.0
  }' | jq .

# 10. Final state
echo "\n10. Final aircraft state..."
sleep 10
curl -s $BASE_URL/state | jq .
```

### Monitoring Script

Continuous monitoring with state updates:

```bash
#!/bin/bash

# Monitor aircraft position every second
while true; do
  clear
  echo "=== Flight Simulator State Monitor ==="
  echo "Timestamp: $(date)"
  echo ""
  
  STATE=$(curl -s http://localhost:8080/state)
  
  echo "Position:"
  echo "  Lat: $(echo $STATE | jq -r '.position.latitude')°"
  echo "  Lon: $(echo $STATE | jq -r '.position.longitude')°"
  echo "  Alt: $(echo $STATE | jq -r '.position.altitude')m"
  echo ""
  
  echo "Velocity:"
  echo "  Ground Speed: $(echo $STATE | jq -r '.velocity.ground_speed')m/s"
  echo "  Vertical Speed: $(echo $STATE | jq -r '.velocity.vertical_speed')m/s"
  echo ""
  
  echo "Heading: $(echo $STATE | jq -r '.heading')°"
  echo ""
  
  COMMAND=$(echo $STATE | jq -r '.active_command.type // "none"')
  echo "Active Command: $COMMAND"
  
  if [ "$COMMAND" != "none" ]; then
    ETA=$(echo $STATE | jq -r '.active_command.eta_seconds // "N/A"')
    echo "  ETA: ${ETA}s"
  fi
  
  sleep 1
done
```

### Load Testing

Simple load test sending multiple commands:

```bash
#!/bin/bash

# Send 100 random goto commands
for i in {1..100}; do
  # Random coordinates around Tel Aviv
  LAT=$(echo "32.0 + ($RANDOM % 100) / 1000" | bc -l)
  LON=$(echo "34.7 + ($RANDOM % 100) / 1000" | bc -l)
  ALT=$(echo "500 + ($RANDOM % 2000)" | bc -l)
  
  curl -s -X POST http://localhost:8080/command/goto \
    -H "Content-Type: application/json" \
    -d "{
      \"lat\": $LAT,
      \"lon\": $LON,
      \"alt\": $ALT
    }" > /dev/null &
  
  echo "Sent command $i"
done

wait
echo "All commands sent!"
```

---

## Rate Limits

**Current version**: No rate limits (demo/assignment purposes)

**Recommendations for production**:
- Command endpoints: 10 requests per second per client
- State query: 30 requests per second per client
- SSE connections: 10 concurrent streams per client

**Backpressure**: 
- Command queue has limited capacity (default: 100)
- Returns `503 Service Unavailable` when full
- Clients should implement retry with exponential backoff

---

## Versioning

**Current API version**: v1 (implicit)

**Version strategy**: URL-based versioning for future releases

**Future**:
```
GET /v1/state
GET /v2/state
```

**Backward compatibility**: Breaking changes require new version

---

## Additional Notes

### Coordinate System

- **Latitude/Longitude**: WGS84 geodetic coordinates
- **Altitude**: Meters above Mean Sea Level (MSL), not Above Ground Level (AGL)
- **Heading**: True heading (not magnetic), 0° = North, clockwise

### Units Summary

| Measurement | Unit | Symbol |
|-------------|------|--------|
| Distance | Meters | m |
| Speed | Meters per second | m/s |
| Altitude | Meters MSL | m |
| Angle/Heading | Degrees | ° |
| Time | Seconds | s |
| Humidity | Percent | % |

### Conversion Reference

**Speed**:
- 1 m/s = 3.6 km/h
- 1 m/s ≈ 1.94 knots
- 100 m/s = 360 km/h ≈ 194 knots

**Distance**:
- 1 nautical mile ≈ 1852 meters
- 1 degree latitude ≈ 111,000 meters
- 1 degree longitude ≈ 111,000 meters × cos(latitude)

### Error Codes Reference

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_LATITUDE` | 400 | Latitude out of range (-90 to 90) |
| `INVALID_LONGITUDE` | 400 | Longitude out of range (-180 to 180) |
| `INVALID_ALTITUDE` | 400 | Altitude negative |
| `INVALID_SPEED` | 400 | Speed negative or exceeds maximum |
| `EMPTY_WAYPOINTS` | 400 | Trajectory has no waypoints |
| `INVALID_WAYPOINT` | 400 | Waypoint has invalid coordinates |
| `MALFORMED_JSON` | 400 | Request body is not valid JSON |
| `QUEUE_FULL` | 503 | Command queue at capacity |
| `SIMULATOR_NOT_RUNNING` | 503 | Simulation engine not active |
| `TERRAIN_CONFLICT` | 422 | Command conflicts with terrain (bonus) |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Support

For issues or questions:
- GitHub: https://github.com/meiron-tzhori/Flight-Simulator/issues
- Documentation: See `docs/` directory in repository

---

**Last Updated**: February 1, 2026  
**API Version**: 1.0  
**Document Version**: 1.0
