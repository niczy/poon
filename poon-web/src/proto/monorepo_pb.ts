// Generated TypeScript interfaces from monorepo.proto

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
  conflicts: string[];
}

export interface ReadDirectoryRequest {
  path: string;
  branch?: string;
  recursive?: boolean;
}

export interface ReadDirectoryResponse {
  items: DirectoryItem[];
}

export interface DirectoryItem {
  name: string;
  isDir: boolean;
  size: number;
  modTime: number; // Unix timestamp
  hash: string;
}

export interface ReadFileRequest {
  path: string;
  branch?: string;
  revision?: string;
}

export interface ReadFileResponse {
  content: Uint8Array;
  hash: string;
  size: number;
}

export interface FileHistoryRequest {
  path: string;
  branch?: string;
  limit?: number;
}

export interface FileHistoryResponse {
  commits: Commit[];
}

export interface Commit {
  hash: string;
  author: string;
  message: string;
  timestamp: number;
  changedFiles: string[];
}

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface BranchesRequest {}

export interface BranchesResponse {
  branches: string[];
  defaultBranch: string;
}

export interface CreateBranchRequest {
  name: string;
  fromBranch?: string;
  fromCommit?: string;
}

export interface CreateBranchResponse {
  success: boolean;
  message: string;
  branchName: string;
  commitHash: string;
}