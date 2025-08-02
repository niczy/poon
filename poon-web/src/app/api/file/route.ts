import { NextRequest, NextResponse } from 'next/server';

// Mock file contents for development - replace with actual poon-server gRPC calls
const mockFiles: { [path: string]: string } = {
  '/README.md': `# Poon Monorepo

This is the main repository for the Poon monorepo system.

## Structure

- \`src/\` - Source code
- \`docs/\` - Documentation
- \`config/\` - Configuration files

## Getting Started

1. Install dependencies
2. Run the development server
3. Open your browser to localhost:3000

## Components

- **poon-server** - gRPC backend service
- **poon-git** - Git-compatible server
- **poon-cli** - Command-line interface
- **poon-web** - Web interface (this application)
`,
  '/src/main.go': `package main

import (
    "fmt"
    "log"
    "net"
    "google.golang.org/grpc"
)

func main() {
    fmt.Println("Hello from Poon monorepo!")
    log.Println("Starting gRPC server...")
    
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    s := grpc.NewServer()
    // Register services here
    
    log.Printf("Server listening at %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
`,
  '/src/frontend/app.js': `// Frontend application entry point
import React from 'react';
import ReactDOM from 'react-dom/client';
import { FileBrowser } from './components/FileBrowser';

console.log("Poon frontend starting...");

function App() {
    return (
        <div className="app">
            <h1>Poon Frontend App</h1>
            <FileBrowser initialPath="/" />
        </div>
    );
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<App />);
`,
  '/src/frontend/package.json': `{
  "name": "poon-frontend",
  "version": "1.0.0",
  "main": "app.js",
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0",
    "next": "^15.4.0",
    "tailwindcss": "^4.0.0"
  },
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  }
}
`,
  '/src/backend/server.go': `package main

import (
    "context"
    "net/http"
    "log"
    "encoding/json"
    
    "github.com/nic/poon/poon-proto/gen"
)

type server struct {
    repoRoot string
}

func (s *server) ReadDirectory(ctx context.Context, req *gen.ReadDirectoryRequest) (*gen.ReadDirectoryResponse, error) {
    // Implementation for reading directory contents
    log.Printf("Reading directory: %s", req.Path)
    
    // Return mock response
    return &gen.ReadDirectoryResponse{
        Items: []*gen.DirectoryItem{},
    }, nil
}

func (s *server) ReadFile(ctx context.Context, req *gen.ReadFileRequest) (*gen.ReadFileResponse, error) {
    // Implementation for reading file contents
    log.Printf("Reading file: %s", req.Path)
    
    return &gen.ReadFileResponse{
        Content: []byte("file content"),
        Size: 100,
        Hash: "hash123",
    }, nil
}

func main() {
    log.Println("Poon backend server starting on :8080")
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        response := map[string]string{
            "message": "Poon backend server",
            "version": "1.0.0",
        }
        json.NewEncoder(w).Encode(response)
    })
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
`,
  '/src/backend/main.go': `package main

import (
    "fmt"
    "log"
)

func init() {
    fmt.Println("Backend module initialized")
    log.Println("Setting up backend services...")
}

func main() {
    fmt.Println("Backend main function")
}
`,
  '/docs/README.md': `# Poon Documentation

Welcome to the Poon monorepo documentation.

## Architecture

The Poon system consists of several components:

1. **poon-server** - gRPC server for monorepo operations
2. **poon-git** - Git-compatible server for partial checkout
3. **poon-cli** - Command-line interface
4. **poon-web** - Web interface (this application)

## Usage

### Web Interface

Navigate through the file tree using the web interface.
Click on directories to explore their contents.
Click on files to view their source code.

### API Endpoints

- \`GET /api/directory?path=/some/path\` - List directory contents
- \`GET /api/file?path=/some/file\` - Get file contents

## Development

This web interface is built with Next.js, React, and TypeScript.

### Getting Started

\`\`\`bash
npm install
npm run dev
\`\`\`

### Building for Production

\`\`\`bash
npm run build
npm start
\`\`\`

## Features

- 📁 Interactive file browser
- 📄 File content viewer
- 🍞 Breadcrumb navigation  
- 📱 Responsive design
- ⚡ Server-side rendering with Next.js
- 🎨 Modern styling with Tailwind CSS
`,
  '/docs/api.md': `# Poon API Reference

## gRPC Services

### MonorepoService

#### ReadDirectory
- **Request**: \`ReadDirectoryRequest{path: string}\`
- **Response**: \`ReadDirectoryResponse{items: DirectoryItem[]}\`

#### ReadFile  
- **Request**: \`ReadFileRequest{path: string}\`
- **Response**: \`ReadFileResponse{content: bytes, size: number, hash: string}\`

#### MergePatch
- **Request**: \`MergePatchRequest{path, patch, message, author, branch}\`
- **Response**: \`MergePatchResponse{success: boolean, message: string}\`

## REST API

The web interface communicates via REST endpoints that proxy to the gRPC service.

### Directory Listing

\`\`\`
GET /api/directory?path=/some/path
\`\`\`

Response:
\`\`\`json
{
  "items": [
    {
      "name": "filename.txt",
      "isDir": false,
      "size": 1024,
      "modTime": 1703097600,
      "hash": "abc123"
    }
  ]
}
\`\`\`

### File Content

\`\`\`
GET /api/file?path=/some/file
\`\`\`

Returns the raw file content with appropriate Content-Type headers.

## Authentication

Currently no authentication is required for read operations.
Write operations will require authentication in future versions.

## Rate Limiting

API endpoints are rate limited:
- 100 requests per minute for directory listings
- 50 requests per minute for file content

## Error Handling

All endpoints return standard HTTP status codes:
- 200: Success
- 400: Bad Request (invalid path)
- 404: Not Found (file/directory doesn't exist)
- 500: Internal Server Error
`,
  '/config/app.yaml': `# Poon application configuration
environment: development

server:
  grpc_port: 50051
  http_port: 8080

web:
  port: 3000
  api_url: http://localhost:8080

git:
  default_branch: main
  clone_depth: 1

logging:
  level: info
  format: json

features:
  file_browser: true
  syntax_highlighting: true
  download_files: true
  responsive_design: true

ui:
  theme: modern
  primary_color: "#3b82f6"
  secondary_color: "#8b5cf6"
`
};

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const path = searchParams.get('path');

    if (!path) {
      return NextResponse.json(
        { error: 'Path parameter is required' },
        { status: 400 }
      );
    }

    // Add artificial delay to simulate network request
    await new Promise(resolve => setTimeout(resolve, 50 + Math.random() * 150));

    const content = mockFiles[path];
    if (!content) {
      return NextResponse.json(
        { error: 'File not found' },
        { status: 404 }
      );
    }

    // Return file content as plain text
    return new NextResponse(content, {
      headers: {
        'Content-Type': 'text/plain; charset=utf-8',
        'X-File-Hash': 'mock-hash-' + Math.random().toString(36).substr(2, 9),
        'X-File-Size': content.length.toString(),
      },
    });
  } catch (error) {
    console.error('File API error:', error);
    return NextResponse.json(
      { error: 'Failed to read file' },
      { status: 500 }
    );
  }
}