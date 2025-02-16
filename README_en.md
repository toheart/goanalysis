# Project README

## Project Overview

This project is an analysis tool based on Go and Vue.js, primarily used for tracking and analyzing the performance of function calls. It provides a front-end interface where users can input function names to query related goroutines and view detailed trace information and call graphs.

## Tech Stack

- **Backend**: Go (using Kratos framework)
- **Frontend**: Vue.js
- **Database**: SQLite
- **Styles**: Bootstrap

## Features

### 1. Trace Data Viewer

- **Feature Description**: Users can input function names, and the system will display the goroutines related to that function.
- **Component**: `TraceViewer.vue`
- **Implementation Details**:
  - Dynamically filter function names using input boxes and dropdown lists.
  - Fetch related GIDs through API requests.
  - Implement fuzzy search functionality using `fuse.js`.

### 2. Trace Details

- **Feature Description**: Display detailed trace information for a specific GID.
- **Component**: `TraceDetails.vue`
- **Implementation Details**:
  - Fetch and display trace data using GID.
  - Provide functionality to view parameters, allowing users to click a button to see specific function parameters.

### 3. Mermaid Viewer

- **Feature Description**: Graphically display function call relationships.
- **Component**: `MermaidViewer.vue`
- **Implementation Details**:
  - Render function call graphs using Mermaid.js.
  - Support zooming and dragging functionalities.

### 4. Database Operations

- **Feature Description**: Store and query trace data using SQLite database.
- **Implementation Details**:
  - Encapsulate database operations using the `Data` struct.
  - Provide functionalities to get all GIDs, function names, and query GIDs based on function names.

### 5. CORS Support

- **Feature Description**: Resolve cross-origin request issues.
- **Implementation Details**: Configure CORS in the Go server to allow requests from different origins.

## API Endpoints

- `GET /api/gids`: Get all GIDs.
- `GET /api/functions`: Get all function names.
- `POST /api/gids/function`: Get related GIDs based on function name.
- `GET /api/traces/{gid}`: Get trace data for a specific GID.
- `GET /api/params/{id}`: Get parameter data for a specific ID.
- `GET /api/traces/{gid}/mermaid`: Get Mermaid graph data for a specific GID.

## Project Structure

```
.
├── cmd
│   └── server
│      └── server.go          # Server entry point
│   └── rewrite.go           # Rewrite command
│   └── main.go              # Program entry point
├── internal
│   ├── data
│   │   └── data.go            # Database operations
│   ├── service
│   │   └── analysis.go        # Business logic
│   └── server
│       └── server.go          # Server configuration
├── functrace
│   └── trace.go               # Function tracing implementation
├── static
│   └── analysis
│       ├── src
│       │   ├── components
│       │   │   ├── MermaidViewer.vue
│       │   │   ├── TraceDetails.vue
│       │   │   └── TraceViewer.vue
│       │   ├── App.vue
│       │   └── main.js
│       └── vue.config.js       # Vue configuration
└── api
    └── analysis
        └── v1
            ├── analysis.proto   # gRPC interface definition
            └── analysis_grpc.pb.go # gRPC generated code
```

## Installation and Running

### 1. Backend

- Ensure Go environment is installed.
- Run the following command in the project root directory to start the server:

```bash
go run cmd/server/server.go
```

### 2. Frontend

- Ensure Node.js and npm are installed.
- Run the following commands in the `frontWeb/` directory to install dependencies and start the frontend:

```bash
npm install
npm run serve
```

## Contribution

Contributions of any form are welcome! Please submit issues or pull requests.

## License

This project is licensed under the MIT License. Please see the LICENSE file for more information.
