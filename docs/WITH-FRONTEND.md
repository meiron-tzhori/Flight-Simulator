# Frontend Dashboard Design (with-frontend)

## 1. Goals

The purpose of the `with-frontend` branch is to add a simple but solid web UI on top of the existing Flight Simulator backend, without changing the core concurrency and simulation architecture.

The frontend should:

- Visualize the **current aircraft state** in real time.
- Provide an **intuitive command interface** for:
  - GoTo commands.
  - Trajectory commands.
- Consume the existing HTTP API:
  - `GET /state`
  - `GET /stream` (SSE)
  - `POST /command/goto`
  - `POST /command/trajectory`
  - `POST /command/stop`
  - `POST /command/hold`
- Be easy to develop, build, and serve alongside the Go backend.

## 2. High-Level Architecture

```text
+--------------------------+        +---------------------------+
|        React UI          |  HTTP  |      Go Backend API       |
|  (Vite dev / static)     +<------>+  /state, /stream, /command|
+-------------+------------+        +---------------------------+
              |
              | SSE (EventSource)
              v
        Live Aircraft State
