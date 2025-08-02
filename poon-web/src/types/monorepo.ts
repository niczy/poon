// TypeScript interfaces for Poon monorepo gRPC service

export interface DirectoryItem {
  name: string;
  isDir: boolean;
  size: number;
  modTime: number; // Unix timestamp
  hash: string;
}

export interface ReadDirectoryRequest {
  path: string;
}

export interface ReadDirectoryResponse {
  items: DirectoryItem[];
}

export interface ReadFileRequest {
  path: string;
}

export interface ReadFileResponse {
  content: Uint8Array;
  hash: string;
  size: number;
}

export interface MergePatchRequest {
  path: string;
  patch: Uint8Array;
  message: string;
  author: string;
  branch: string;
}

export interface MergePatchResponse {
  success: boolean;
  message: string;
  commitHash: string;
}

// Client interface for the MonorepoService
export interface MonorepoServiceClient {
  readDirectory(request: ReadDirectoryRequest): Promise<ReadDirectoryResponse>;
  readFile(request: ReadFileRequest): Promise<ReadFileResponse>;
  mergePatch(request: MergePatchRequest): Promise<MergePatchResponse>;
}