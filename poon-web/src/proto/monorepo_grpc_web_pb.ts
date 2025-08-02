// gRPC-Web client for MonorepoService
import {
  MergePatchRequest,
  MergePatchResponse,
  ReadDirectoryRequest,
  ReadDirectoryResponse,
  ReadFileRequest,
  ReadFileResponse,
  FileHistoryRequest,
  FileHistoryResponse,
  BranchesRequest,
  BranchesResponse,
  CreateBranchRequest,
  CreateBranchResponse,
  DirectoryItem,
} from './monorepo_pb';

// Client interface
export interface MonorepoServiceClient {
  mergePatch(request: MergePatchRequest): Promise<MergePatchResponse>;
  readDirectory(request: ReadDirectoryRequest): Promise<ReadDirectoryResponse>;
  readFile(request: ReadFileRequest): Promise<ReadFileResponse>;
  getFileHistory(request: FileHistoryRequest): Promise<FileHistoryResponse>;
  getBranches(request: BranchesRequest): Promise<BranchesResponse>;
  createBranch(request: CreateBranchRequest): Promise<CreateBranchResponse>;
}

// gRPC-Web client implementation
export class MonorepoServiceClientImpl implements MonorepoServiceClient {
  private hostname: string;

  constructor(hostname: string = 'http://localhost:8080') {
    this.hostname = hostname;
  }

  async mergePatch(_request: MergePatchRequest): Promise<MergePatchResponse> {
    // For now, return a mock response since we need a gRPC-Web proxy
    // In production, this would make an actual gRPC-Web call
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          success: false,
          message: 'gRPC-Web proxy not configured - using mock data',
          commitHash: '',
          conflicts: []
        });
      }, 100);
    });
  }

  async readDirectory(request: ReadDirectoryRequest): Promise<ReadDirectoryResponse> {
    // Simulate gRPC call - in production this would call the actual gRPC-Web endpoint
    const mockResponse: ReadDirectoryResponse = {
      items: this.getMockDirectoryItems(request.path)
    };
    
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(mockResponse);
      }, 100 + Math.random() * 200);
    });
  }

  async readFile(request: ReadFileRequest): Promise<ReadFileResponse> {
    const content = this.getMockFileContent(request.path);
    const contentBytes = new TextEncoder().encode(content);
    
    const mockResponse: ReadFileResponse = {
      content: contentBytes,
      hash: 'mock-hash-' + Math.random().toString(36).substr(2, 9),
      size: contentBytes.length
    };
    
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(mockResponse);
      }, 50 + Math.random() * 150);
    });
  }

  async getFileHistory(request: FileHistoryRequest): Promise<FileHistoryResponse> {
    const mockResponse: FileHistoryResponse = {
      commits: [
        {
          hash: 'abc123def456',
          author: 'developer@example.com',
          message: 'Update ' + request.path,
          timestamp: Math.floor(Date.now() / 1000) - 3600,
          changedFiles: [request.path]
        },
        {
          hash: 'def456ghi789',
          author: 'maintainer@example.com',
          message: 'Initial commit for ' + request.path,
          timestamp: Math.floor(Date.now() / 1000) - 86400,
          changedFiles: [request.path]
        }
      ]
    };
    
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(mockResponse);
      }, 200);
    });
  }

  async getBranches(_request: BranchesRequest): Promise<BranchesResponse> {
    const mockResponse: BranchesResponse = {
      branches: ['main', 'develop', 'feature/new-feature', 'hotfix/critical-fix'],
      defaultBranch: 'main'
    };
    
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(mockResponse);
      }, 100);
    });
  }

  async createBranch(request: CreateBranchRequest): Promise<CreateBranchResponse> {
    const mockResponse: CreateBranchResponse = {
      success: true,
      message: 'Branch created successfully',
      branchName: request.name,
      commitHash: 'def456ghi789'
    };
    
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(mockResponse);
      }, 300);
    });
  }

  // Mock data helpers (same as API routes but using gRPC interface)
  private getMockDirectoryItems(path: string) {
    const mockData: { [path: string]: DirectoryItem[] } = {
      '/': [
        { name: 'src', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 3600, hash: '' },
        { name: 'docs', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 7200, hash: '' },
        { name: 'config', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 1800, hash: '' },
        { name: 'README.md', isDir: false, size: 1024, modTime: Math.floor(Date.now() / 1000) - 900, hash: 'abc123' },
        { name: 'go.mod', isDir: false, size: 256, modTime: Math.floor(Date.now() / 1000) - 1200, hash: 'def456' }
      ],
      '/src': [
        { name: 'frontend', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 2400, hash: '' },
        { name: 'backend', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 3000, hash: '' },
        { name: 'main.go', isDir: false, size: 2048, modTime: Math.floor(Date.now() / 1000) - 600, hash: 'ghi789' },
        { name: 'utils.go', isDir: false, size: 512, modTime: Math.floor(Date.now() / 1000) - 1500, hash: 'jkl012' }
      ],
      '/src/frontend': [
        { name: 'components', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 1800, hash: '' },
        { name: 'pages', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 2100, hash: '' },
        { name: 'app.tsx', isDir: false, size: 1536, modTime: Math.floor(Date.now() / 1000) - 300, hash: 'mno345' },
        { name: 'package.json', isDir: false, size: 768, modTime: Math.floor(Date.now() / 1000) - 4800, hash: 'pqr678' }
      ],
      '/src/backend': [
        { name: 'handlers', isDir: true, size: 0, modTime: Math.floor(Date.now() / 1000) - 2700, hash: '' },
        { name: 'server.go', isDir: false, size: 3072, modTime: Math.floor(Date.now() / 1000) - 450, hash: 'stu901' },
        { name: 'main.go', isDir: false, size: 1024, modTime: Math.floor(Date.now() / 1000) - 1800, hash: 'vwx234' },
        { name: 'config.go', isDir: false, size: 512, modTime: Math.floor(Date.now() / 1000) - 3600, hash: 'yza567' }
      ],
      '/docs': [
        { name: 'architecture.md', isDir: false, size: 2048, modTime: Math.floor(Date.now() / 1000) - 7200, hash: 'bcd890' },
        { name: 'api.md', isDir: false, size: 4096, modTime: Math.floor(Date.now() / 1000) - 1800, hash: 'efg123' },
        { name: 'README.md', isDir: false, size: 1536, modTime: Math.floor(Date.now() / 1000) - 3600, hash: 'hij456' }
      ],
      '/config': [
        { name: 'app.yaml', isDir: false, size: 1024, modTime: Math.floor(Date.now() / 1000) - 1200, hash: 'klm789' },
        { name: 'database.yaml', isDir: false, size: 512, modTime: Math.floor(Date.now() / 1000) - 2400, hash: 'nop012' }
      ]
    };

    return mockData[path] || [];
  }

  private getMockFileContent(path: string): string {
    const mockFiles: { [path: string]: string } = {
      '/README.md': `# Poon Monorepo

This is the main repository for the Poon monorepo system, now with gRPC support!

## Architecture

The system uses gRPC for communication between components:

- **poon-server** - gRPC backend service (port 50051)
- **poon-git** - Git-compatible server  
- **poon-cli** - Command-line interface with gRPC client
- **poon-web** - Web interface with gRPC-Web client

## gRPC Services

The MonorepoService provides these operations:

### File Operations
- \`ReadDirectory(path)\` - List directory contents
- \`ReadFile(path)\` - Get file contents
- \`GetFileHistory(path)\` - View commit history for a file

### Branch Operations  
- \`GetBranches()\` - List available branches
- \`CreateBranch(name, from)\` - Create new branches

### Patch Operations
- \`MergePatch(path, patch, message, author)\` - Apply patches to files

## Getting Started

1. Start the gRPC server:
   \`\`\`bash
   cd poon-server && ./poon-server
   \`\`\`

2. Start the web interface:
   \`\`\`bash
   cd poon-web && npm run dev
   \`\`\`

3. Browse to http://localhost:3000

## Development

The web client uses gRPC-Web to communicate with the backend service through a proxy.
In development mode, mock data is provided when the gRPC server is not available.
`,
      '/go.mod': `module github.com/nic/poon

go 1.21

require (
    google.golang.org/grpc v1.58.0
    google.golang.org/protobuf v1.31.0
    github.com/spf13/cobra v1.7.0
)

require (
    github.com/golang/protobuf v1.5.3 // indirect
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    golang.org/x/net v0.15.0 // indirect
    golang.org/x/sys v0.12.0 // indirect
    golang.org/x/text v0.13.0 // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20230920204549-e6e6cdab5c13 // indirect
)
`,
      '/src/main.go': `package main

import (
    "context"
    "log"
    "net"
    "os"
    "path/filepath"
    
    "google.golang.org/grpc"
    pb "github.com/nic/poon/poon-proto/gen/go"
)

type server struct {
    pb.UnimplementedMonorepoServiceServer
    repoRoot string
}

func (s *server) ReadDirectory(ctx context.Context, req *pb.ReadDirectoryRequest) (*pb.ReadDirectoryResponse, error) {
    log.Printf("gRPC ReadDirectory: %s", req.Path)
    
    fullPath := filepath.Join(s.repoRoot, req.Path)
    entries, err := os.ReadDir(fullPath)
    if err != nil {
        return nil, err
    }
    
    var items []*pb.DirectoryItem
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue
        }
        
        items = append(items, &pb.DirectoryItem{
            Name:    entry.Name(),
            IsDir:   entry.IsDir(),
            Size:    info.Size(),
            ModTime: info.ModTime().Unix(),
            Hash:    "", // TODO: Calculate git hash
        })
    }
    
    return &pb.ReadDirectoryResponse{
        Items: items,
    }, nil
}

func (s *server) ReadFile(ctx context.Context, req *pb.ReadFileRequest) (*pb.ReadFileResponse, error) {
    log.Printf("gRPC ReadFile: %s", req.Path)
    
    fullPath := filepath.Join(s.repoRoot, req.Path)
    content, err := os.ReadFile(fullPath)
    if err != nil {
        return nil, err
    }
    
    return &pb.ReadFileResponse{
        Content: content,
        Size:    int64(len(content)),
        Hash:    "", // TODO: Calculate git hash
    }, nil
}

func (s *server) MergePatch(ctx context.Context, req *pb.MergePatchRequest) (*pb.MergePatchResponse, error) {
    log.Printf("gRPC MergePatch: %s", req.Path)
    
    // TODO: Implement patch merging
    return &pb.MergePatchResponse{
        Success: false,
        Message: "Patch merging not yet implemented",
    }, nil
}

func main() {
    repoRoot := os.Getenv("REPO_ROOT")
    if repoRoot == "" {
        repoRoot = "."
    }
    
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterMonorepoServiceServer(s, &server{
        repoRoot: repoRoot,
    })

    log.Printf("gRPC server listening at %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
`,
      '/src/frontend/app.tsx': `import React, { useState, useEffect } from 'react';
import { FileBrowser } from '@/components/FileBrowser';
import { MonorepoServiceClientImpl } from '@/proto/monorepo_grpc_web_pb';

const grpcClient = new MonorepoServiceClientImpl('http://localhost:8080');

export default function App() {
  const [branches, setBranches] = useState<string[]>([]);
  const [currentBranch, setCurrentBranch] = useState<string>('main');

  useEffect(() => {
    // Load available branches on app start
    grpcClient.getBranches({}).then(response => {
      setBranches(response.branches);
      setCurrentBranch(response.defaultBranch);
    }).catch(err => {
      console.log('Using mock data - gRPC server not available');
    });
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-gradient-to-r from-blue-600 to-purple-700 text-white">
        <div className="max-w-6xl mx-auto px-4 py-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-bold">Poon Monorepo Browser</h1>
              <p className="text-xl opacity-90">gRPC-powered file browsing</p>
            </div>
            
            <div className="text-right">
              <div className="text-sm opacity-75">Current Branch</div>
              <select 
                className="bg-white/20 text-white border border-white/30 rounded px-3 py-1"
                value={currentBranch}
                onChange={(e) => setCurrentBranch(e.target.value)}
              >
                {branches.map(branch => (
                  <option key={branch} value={branch} className="text-gray-900">
                    {branch}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </header>
      
      <main className="max-w-6xl mx-auto p-6">
        <FileBrowser 
          initialPath="/" 
          branch={currentBranch}
          grpcClient={grpcClient}
        />
      </main>
    </div>
  );
}
`,
      '/config/app.yaml': `# Poon system configuration
grpc:
  server_host: localhost
  server_port: 50051
  web_proxy_port: 8080

web:
  port: 3000
  grpc_web_endpoint: http://localhost:8080

repository:
  root_path: /path/to/monorepo
  default_branch: main
  
features:
  file_browser: true
  branch_switching: true
  file_history: true
  patch_merging: true
  
logging:
  level: info
  format: json

# gRPC-Web proxy configuration (for production)
proxy:
  enabled: true
  cors_enabled: true
  allowed_origins:
    - http://localhost:3000
    - https://poon-web.example.com
`
    };

    return mockFiles[path] || `// Generated file content for ${path}
    
This file is part of the Poon monorepo system.
It demonstrates gRPC communication between components.

Content generated via gRPC ReadFile operation.
Timestamp: ${new Date().toISOString()}
Path: ${path}
`;
  }
}