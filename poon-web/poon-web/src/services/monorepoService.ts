import { 
  DirectoryItem, 
  ReadDirectoryRequest, 
  ReadDirectoryResponse, 
  ReadFileRequest, 
  ReadFileResponse,
  MonorepoServiceClient 
} from '../types/monorepo';

class MonorepoService implements MonorepoServiceClient {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
  }

  async readDirectory(request: ReadDirectoryRequest): Promise<ReadDirectoryResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/api/directory?path=${encodeURIComponent(request.path)}`);
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      
      // Transform the response to match our interface
      // Since we're not using the Go backend anymore, we'll simulate the response structure
      return {
        items: data.items?.map((item: any) => ({
          name: item.name,
          isDir: item.is_dir || item.isDir,
          size: item.size || 0,
          modTime: item.mod_time || item.modTime || Date.now() / 1000,
          hash: item.hash || ''
        })) || []
      };
    } catch (error) {
      console.error('Error reading directory:', error);
      throw error;
    }
  }

  async readFile(request: ReadFileRequest): Promise<ReadFileResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/api/file?path=${encodeURIComponent(request.path)}`);
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const arrayBuffer = await response.arrayBuffer();
      const content = new Uint8Array(arrayBuffer);

      return {
        content,
        hash: response.headers.get('x-file-hash') || '',
        size: content.length
      };
    } catch (error) {
      console.error('Error reading file:', error);
      throw error;
    }
  }

  async mergePatch(): Promise<any> {
    throw new Error('MergePatch not implemented in web client yet');
  }
}

// Mock service for development when backend is not available
class MockMonorepoService implements MonorepoServiceClient {
  private mockData: { [path: string]: DirectoryItem[] } = {
    '/': [
      { name: 'src', isDir: true, size: 0, modTime: Date.now() / 1000, hash: '' },
      { name: 'docs', isDir: true, size: 0, modTime: Date.now() / 1000, hash: '' },
      { name: 'config', isDir: true, size: 0, modTime: Date.now() / 1000, hash: '' },
      { name: 'README.md', isDir: false, size: 1024, modTime: Date.now() / 1000, hash: 'abc123' }
    ],
    '/src': [
      { name: 'frontend', isDir: true, size: 0, modTime: Date.now() / 1000, hash: '' },
      { name: 'backend', isDir: true, size: 0, modTime: Date.now() / 1000, hash: '' },
      { name: 'main.go', isDir: false, size: 2048, modTime: Date.now() / 1000, hash: 'def456' }
    ],
    '/src/frontend': [
      { name: 'app.js', isDir: false, size: 512, modTime: Date.now() / 1000, hash: 'ghi789' },
      { name: 'package.json', isDir: false, size: 256, modTime: Date.now() / 1000, hash: 'jkl012' }
    ],
    '/src/backend': [
      { name: 'server.go', isDir: false, size: 1536, modTime: Date.now() / 1000, hash: 'mno345' },
      { name: 'main.go', isDir: false, size: 768, modTime: Date.now() / 1000, hash: 'pqr678' }
    ],
    '/docs': [
      { name: 'README.md', isDir: false, size: 4096, modTime: Date.now() / 1000, hash: 'stu901' },
      { name: 'api.md', isDir: false, size: 2048, modTime: Date.now() / 1000, hash: 'vwx234' }
    ],
    '/config': [
      { name: 'app.yaml', isDir: false, size: 512, modTime: Date.now() / 1000, hash: 'yza567' }
    ]
  };

  private mockFiles: { [path: string]: string } = {
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
`,
    '/src/main.go': `package main

import (
    "fmt"
    "log"
)

func main() {
    fmt.Println("Hello from Poon monorepo!")
    log.Println("Server starting...")
}
`,
    '/src/frontend/app.js': `// Frontend application entry point
console.log("Poon frontend starting...");

function initApp() {
    document.body.innerHTML = '<h1>Poon Frontend App</h1>';
}

window.onload = initApp;
`,
    '/src/frontend/package.json': `{
  "name": "poon-frontend",
  "version": "1.0.0",
  "main": "app.js",
  "dependencies": {
    "react": "^18.0.0"
  }
}
`,
    '/src/backend/server.go': `package main

import (
    "net/http"
    "log"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Poon backend server"))
    })
    
    log.Println("Backend server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
`,
    '/src/backend/main.go': `package main

import "fmt"

func init() {
    fmt.Println("Backend module initialized")
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

This web interface is built with React and TypeScript.
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
`
  };

  async readDirectory(request: ReadDirectoryRequest): Promise<ReadDirectoryResponse> {
    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, 100 + Math.random() * 200));
    
    const items = this.mockData[request.path] || [];
    return { items: [...items] }; // Return a copy
  }

  async readFile(request: ReadFileRequest): Promise<ReadFileResponse> {
    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, 50 + Math.random() * 150));
    
    const content = this.mockFiles[request.path] || '';
    const contentBytes = new TextEncoder().encode(content);
    
    return {
      content: contentBytes,
      hash: 'mock-hash-' + Math.random().toString(36).substr(2, 9),
      size: contentBytes.length
    };
  }

  async mergePatch(): Promise<any> {
    throw new Error('MergePatch not implemented in mock service');
  }
}

// Export service instances
export const monorepoService = new MonorepoService();
export const mockMonorepoService = new MockMonorepoService();

// Use mock service by default for development
export const defaultService = mockMonorepoService;