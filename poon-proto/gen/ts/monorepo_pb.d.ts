import * as jspb from 'google-protobuf'



export class MergePatchRequest extends jspb.Message {
  getPath(): string;
  setPath(value: string): MergePatchRequest;

  getPatch(): Uint8Array | string;
  getPatch_asU8(): Uint8Array;
  getPatch_asB64(): string;
  setPatch(value: Uint8Array | string): MergePatchRequest;

  getMessage(): string;
  setMessage(value: string): MergePatchRequest;

  getAuthor(): string;
  setAuthor(value: string): MergePatchRequest;

  getBranch(): string;
  setBranch(value: string): MergePatchRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MergePatchRequest.AsObject;
  static toObject(includeInstance: boolean, msg: MergePatchRequest): MergePatchRequest.AsObject;
  static serializeBinaryToWriter(message: MergePatchRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MergePatchRequest;
  static deserializeBinaryFromReader(message: MergePatchRequest, reader: jspb.BinaryReader): MergePatchRequest;
}

export namespace MergePatchRequest {
  export type AsObject = {
    path: string,
    patch: Uint8Array | string,
    message: string,
    author: string,
    branch: string,
  }
}

export class MergePatchResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): MergePatchResponse;

  getMessage(): string;
  setMessage(value: string): MergePatchResponse;

  getCommitHash(): string;
  setCommitHash(value: string): MergePatchResponse;

  getConflictsList(): Array<string>;
  setConflictsList(value: Array<string>): MergePatchResponse;
  clearConflictsList(): MergePatchResponse;
  addConflicts(value: string, index?: number): MergePatchResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MergePatchResponse.AsObject;
  static toObject(includeInstance: boolean, msg: MergePatchResponse): MergePatchResponse.AsObject;
  static serializeBinaryToWriter(message: MergePatchResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MergePatchResponse;
  static deserializeBinaryFromReader(message: MergePatchResponse, reader: jspb.BinaryReader): MergePatchResponse;
}

export namespace MergePatchResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    commitHash: string,
    conflictsList: Array<string>,
  }
}

export class ReadDirectoryRequest extends jspb.Message {
  getPath(): string;
  setPath(value: string): ReadDirectoryRequest;

  getBranch(): string;
  setBranch(value: string): ReadDirectoryRequest;

  getRecursive(): boolean;
  setRecursive(value: boolean): ReadDirectoryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadDirectoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ReadDirectoryRequest): ReadDirectoryRequest.AsObject;
  static serializeBinaryToWriter(message: ReadDirectoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadDirectoryRequest;
  static deserializeBinaryFromReader(message: ReadDirectoryRequest, reader: jspb.BinaryReader): ReadDirectoryRequest;
}

export namespace ReadDirectoryRequest {
  export type AsObject = {
    path: string,
    branch: string,
    recursive: boolean,
  }
}

export class ReadDirectoryResponse extends jspb.Message {
  getItemsList(): Array<DirectoryItem>;
  setItemsList(value: Array<DirectoryItem>): ReadDirectoryResponse;
  clearItemsList(): ReadDirectoryResponse;
  addItems(value?: DirectoryItem, index?: number): DirectoryItem;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadDirectoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ReadDirectoryResponse): ReadDirectoryResponse.AsObject;
  static serializeBinaryToWriter(message: ReadDirectoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadDirectoryResponse;
  static deserializeBinaryFromReader(message: ReadDirectoryResponse, reader: jspb.BinaryReader): ReadDirectoryResponse;
}

export namespace ReadDirectoryResponse {
  export type AsObject = {
    itemsList: Array<DirectoryItem.AsObject>,
  }
}

export class DirectoryItem extends jspb.Message {
  getName(): string;
  setName(value: string): DirectoryItem;

  getIsDir(): boolean;
  setIsDir(value: boolean): DirectoryItem;

  getSize(): number;
  setSize(value: number): DirectoryItem;

  getModTime(): number;
  setModTime(value: number): DirectoryItem;

  getHash(): string;
  setHash(value: string): DirectoryItem;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DirectoryItem.AsObject;
  static toObject(includeInstance: boolean, msg: DirectoryItem): DirectoryItem.AsObject;
  static serializeBinaryToWriter(message: DirectoryItem, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DirectoryItem;
  static deserializeBinaryFromReader(message: DirectoryItem, reader: jspb.BinaryReader): DirectoryItem;
}

export namespace DirectoryItem {
  export type AsObject = {
    name: string,
    isDir: boolean,
    size: number,
    modTime: number,
    hash: string,
  }
}

export class ReadFileRequest extends jspb.Message {
  getPath(): string;
  setPath(value: string): ReadFileRequest;

  getBranch(): string;
  setBranch(value: string): ReadFileRequest;

  getRevision(): string;
  setRevision(value: string): ReadFileRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadFileRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ReadFileRequest): ReadFileRequest.AsObject;
  static serializeBinaryToWriter(message: ReadFileRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadFileRequest;
  static deserializeBinaryFromReader(message: ReadFileRequest, reader: jspb.BinaryReader): ReadFileRequest;
}

export namespace ReadFileRequest {
  export type AsObject = {
    path: string,
    branch: string,
    revision: string,
  }
}

export class ReadFileResponse extends jspb.Message {
  getContent(): Uint8Array | string;
  getContent_asU8(): Uint8Array;
  getContent_asB64(): string;
  setContent(value: Uint8Array | string): ReadFileResponse;

  getHash(): string;
  setHash(value: string): ReadFileResponse;

  getSize(): number;
  setSize(value: number): ReadFileResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReadFileResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ReadFileResponse): ReadFileResponse.AsObject;
  static serializeBinaryToWriter(message: ReadFileResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReadFileResponse;
  static deserializeBinaryFromReader(message: ReadFileResponse, reader: jspb.BinaryReader): ReadFileResponse;
}

export namespace ReadFileResponse {
  export type AsObject = {
    content: Uint8Array | string,
    hash: string,
    size: number,
  }
}

export class FileHistoryRequest extends jspb.Message {
  getPath(): string;
  setPath(value: string): FileHistoryRequest;

  getBranch(): string;
  setBranch(value: string): FileHistoryRequest;

  getLimit(): number;
  setLimit(value: number): FileHistoryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FileHistoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: FileHistoryRequest): FileHistoryRequest.AsObject;
  static serializeBinaryToWriter(message: FileHistoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FileHistoryRequest;
  static deserializeBinaryFromReader(message: FileHistoryRequest, reader: jspb.BinaryReader): FileHistoryRequest;
}

export namespace FileHistoryRequest {
  export type AsObject = {
    path: string,
    branch: string,
    limit: number,
  }
}

export class FileHistoryResponse extends jspb.Message {
  getCommitsList(): Array<Commit>;
  setCommitsList(value: Array<Commit>): FileHistoryResponse;
  clearCommitsList(): FileHistoryResponse;
  addCommits(value?: Commit, index?: number): Commit;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FileHistoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: FileHistoryResponse): FileHistoryResponse.AsObject;
  static serializeBinaryToWriter(message: FileHistoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FileHistoryResponse;
  static deserializeBinaryFromReader(message: FileHistoryResponse, reader: jspb.BinaryReader): FileHistoryResponse;
}

export namespace FileHistoryResponse {
  export type AsObject = {
    commitsList: Array<Commit.AsObject>,
  }
}

export class Commit extends jspb.Message {
  getHash(): string;
  setHash(value: string): Commit;

  getAuthor(): string;
  setAuthor(value: string): Commit;

  getMessage(): string;
  setMessage(value: string): Commit;

  getTimestamp(): number;
  setTimestamp(value: number): Commit;

  getChangedFilesList(): Array<string>;
  setChangedFilesList(value: Array<string>): Commit;
  clearChangedFilesList(): Commit;
  addChangedFiles(value: string, index?: number): Commit;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Commit.AsObject;
  static toObject(includeInstance: boolean, msg: Commit): Commit.AsObject;
  static serializeBinaryToWriter(message: Commit, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Commit;
  static deserializeBinaryFromReader(message: Commit, reader: jspb.BinaryReader): Commit;
}

export namespace Commit {
  export type AsObject = {
    hash: string,
    author: string,
    message: string,
    timestamp: number,
    changedFilesList: Array<string>,
  }
}

export class BranchesRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BranchesRequest.AsObject;
  static toObject(includeInstance: boolean, msg: BranchesRequest): BranchesRequest.AsObject;
  static serializeBinaryToWriter(message: BranchesRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BranchesRequest;
  static deserializeBinaryFromReader(message: BranchesRequest, reader: jspb.BinaryReader): BranchesRequest;
}

export namespace BranchesRequest {
  export type AsObject = {
  }
}

export class BranchesResponse extends jspb.Message {
  getBranchesList(): Array<string>;
  setBranchesList(value: Array<string>): BranchesResponse;
  clearBranchesList(): BranchesResponse;
  addBranches(value: string, index?: number): BranchesResponse;

  getDefaultBranch(): string;
  setDefaultBranch(value: string): BranchesResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BranchesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: BranchesResponse): BranchesResponse.AsObject;
  static serializeBinaryToWriter(message: BranchesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BranchesResponse;
  static deserializeBinaryFromReader(message: BranchesResponse, reader: jspb.BinaryReader): BranchesResponse;
}

export namespace BranchesResponse {
  export type AsObject = {
    branchesList: Array<string>,
    defaultBranch: string,
  }
}

export class CreateBranchRequest extends jspb.Message {
  getName(): string;
  setName(value: string): CreateBranchRequest;

  getFromBranch(): string;
  setFromBranch(value: string): CreateBranchRequest;

  getFromCommit(): string;
  setFromCommit(value: string): CreateBranchRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateBranchRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateBranchRequest): CreateBranchRequest.AsObject;
  static serializeBinaryToWriter(message: CreateBranchRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateBranchRequest;
  static deserializeBinaryFromReader(message: CreateBranchRequest, reader: jspb.BinaryReader): CreateBranchRequest;
}

export namespace CreateBranchRequest {
  export type AsObject = {
    name: string,
    fromBranch: string,
    fromCommit: string,
  }
}

export class CreateBranchResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): CreateBranchResponse;

  getMessage(): string;
  setMessage(value: string): CreateBranchResponse;

  getBranchName(): string;
  setBranchName(value: string): CreateBranchResponse;

  getCommitHash(): string;
  setCommitHash(value: string): CreateBranchResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateBranchResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateBranchResponse): CreateBranchResponse.AsObject;
  static serializeBinaryToWriter(message: CreateBranchResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateBranchResponse;
  static deserializeBinaryFromReader(message: CreateBranchResponse, reader: jspb.BinaryReader): CreateBranchResponse;
}

export namespace CreateBranchResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    branchName: string,
    commitHash: string,
  }
}

export class CreateWorkspaceRequest extends jspb.Message {
  getName(): string;
  setName(value: string): CreateWorkspaceRequest;

  getTrackedPathsList(): Array<string>;
  setTrackedPathsList(value: Array<string>): CreateWorkspaceRequest;
  clearTrackedPathsList(): CreateWorkspaceRequest;
  addTrackedPaths(value: string, index?: number): CreateWorkspaceRequest;

  getBaseBranch(): string;
  setBaseBranch(value: string): CreateWorkspaceRequest;

  getMetadataMap(): jspb.Map<string, string>;
  clearMetadataMap(): CreateWorkspaceRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateWorkspaceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateWorkspaceRequest): CreateWorkspaceRequest.AsObject;
  static serializeBinaryToWriter(message: CreateWorkspaceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateWorkspaceRequest;
  static deserializeBinaryFromReader(message: CreateWorkspaceRequest, reader: jspb.BinaryReader): CreateWorkspaceRequest;
}

export namespace CreateWorkspaceRequest {
  export type AsObject = {
    name: string,
    trackedPathsList: Array<string>,
    baseBranch: string,
    metadataMap: Array<[string, string]>,
  }
}

export class CreateWorkspaceResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): CreateWorkspaceResponse;

  getMessage(): string;
  setMessage(value: string): CreateWorkspaceResponse;

  getWorkspaceId(): string;
  setWorkspaceId(value: string): CreateWorkspaceResponse;

  getRemoteUrl(): string;
  setRemoteUrl(value: string): CreateWorkspaceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateWorkspaceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateWorkspaceResponse): CreateWorkspaceResponse.AsObject;
  static serializeBinaryToWriter(message: CreateWorkspaceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateWorkspaceResponse;
  static deserializeBinaryFromReader(message: CreateWorkspaceResponse, reader: jspb.BinaryReader): CreateWorkspaceResponse;
}

export namespace CreateWorkspaceResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    workspaceId: string,
    remoteUrl: string,
  }
}

export class GetWorkspaceRequest extends jspb.Message {
  getWorkspaceId(): string;
  setWorkspaceId(value: string): GetWorkspaceRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetWorkspaceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetWorkspaceRequest): GetWorkspaceRequest.AsObject;
  static serializeBinaryToWriter(message: GetWorkspaceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetWorkspaceRequest;
  static deserializeBinaryFromReader(message: GetWorkspaceRequest, reader: jspb.BinaryReader): GetWorkspaceRequest;
}

export namespace GetWorkspaceRequest {
  export type AsObject = {
    workspaceId: string,
  }
}

export class GetWorkspaceResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): GetWorkspaceResponse;

  getMessage(): string;
  setMessage(value: string): GetWorkspaceResponse;

  getWorkspace(): WorkspaceInfo | undefined;
  setWorkspace(value?: WorkspaceInfo): GetWorkspaceResponse;
  hasWorkspace(): boolean;
  clearWorkspace(): GetWorkspaceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetWorkspaceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetWorkspaceResponse): GetWorkspaceResponse.AsObject;
  static serializeBinaryToWriter(message: GetWorkspaceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetWorkspaceResponse;
  static deserializeBinaryFromReader(message: GetWorkspaceResponse, reader: jspb.BinaryReader): GetWorkspaceResponse;
}

export namespace GetWorkspaceResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    workspace?: WorkspaceInfo.AsObject,
  }
}

export class UpdateWorkspaceRequest extends jspb.Message {
  getWorkspaceId(): string;
  setWorkspaceId(value: string): UpdateWorkspaceRequest;

  getTrackedPathsList(): Array<string>;
  setTrackedPathsList(value: Array<string>): UpdateWorkspaceRequest;
  clearTrackedPathsList(): UpdateWorkspaceRequest;
  addTrackedPaths(value: string, index?: number): UpdateWorkspaceRequest;

  getMetadataMap(): jspb.Map<string, string>;
  clearMetadataMap(): UpdateWorkspaceRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateWorkspaceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateWorkspaceRequest): UpdateWorkspaceRequest.AsObject;
  static serializeBinaryToWriter(message: UpdateWorkspaceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateWorkspaceRequest;
  static deserializeBinaryFromReader(message: UpdateWorkspaceRequest, reader: jspb.BinaryReader): UpdateWorkspaceRequest;
}

export namespace UpdateWorkspaceRequest {
  export type AsObject = {
    workspaceId: string,
    trackedPathsList: Array<string>,
    metadataMap: Array<[string, string]>,
  }
}

export class UpdateWorkspaceResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): UpdateWorkspaceResponse;

  getMessage(): string;
  setMessage(value: string): UpdateWorkspaceResponse;

  getWorkspace(): WorkspaceInfo | undefined;
  setWorkspace(value?: WorkspaceInfo): UpdateWorkspaceResponse;
  hasWorkspace(): boolean;
  clearWorkspace(): UpdateWorkspaceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateWorkspaceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateWorkspaceResponse): UpdateWorkspaceResponse.AsObject;
  static serializeBinaryToWriter(message: UpdateWorkspaceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateWorkspaceResponse;
  static deserializeBinaryFromReader(message: UpdateWorkspaceResponse, reader: jspb.BinaryReader): UpdateWorkspaceResponse;
}

export namespace UpdateWorkspaceResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    workspace?: WorkspaceInfo.AsObject,
  }
}

export class DeleteWorkspaceRequest extends jspb.Message {
  getWorkspaceId(): string;
  setWorkspaceId(value: string): DeleteWorkspaceRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteWorkspaceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteWorkspaceRequest): DeleteWorkspaceRequest.AsObject;
  static serializeBinaryToWriter(message: DeleteWorkspaceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteWorkspaceRequest;
  static deserializeBinaryFromReader(message: DeleteWorkspaceRequest, reader: jspb.BinaryReader): DeleteWorkspaceRequest;
}

export namespace DeleteWorkspaceRequest {
  export type AsObject = {
    workspaceId: string,
  }
}

export class DeleteWorkspaceResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): DeleteWorkspaceResponse;

  getMessage(): string;
  setMessage(value: string): DeleteWorkspaceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteWorkspaceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteWorkspaceResponse): DeleteWorkspaceResponse.AsObject;
  static serializeBinaryToWriter(message: DeleteWorkspaceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteWorkspaceResponse;
  static deserializeBinaryFromReader(message: DeleteWorkspaceResponse, reader: jspb.BinaryReader): DeleteWorkspaceResponse;
}

export namespace DeleteWorkspaceResponse {
  export type AsObject = {
    success: boolean,
    message: string,
  }
}

export class WorkspaceInfo extends jspb.Message {
  getId(): string;
  setId(value: string): WorkspaceInfo;

  getName(): string;
  setName(value: string): WorkspaceInfo;

  getTrackedPathsList(): Array<string>;
  setTrackedPathsList(value: Array<string>): WorkspaceInfo;
  clearTrackedPathsList(): WorkspaceInfo;
  addTrackedPaths(value: string, index?: number): WorkspaceInfo;

  getCreatedAt(): string;
  setCreatedAt(value: string): WorkspaceInfo;

  getLastSync(): string;
  setLastSync(value: string): WorkspaceInfo;

  getStatus(): WorkspaceStatus;
  setStatus(value: WorkspaceStatus): WorkspaceInfo;

  getMetadataMap(): jspb.Map<string, string>;
  clearMetadataMap(): WorkspaceInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WorkspaceInfo.AsObject;
  static toObject(includeInstance: boolean, msg: WorkspaceInfo): WorkspaceInfo.AsObject;
  static serializeBinaryToWriter(message: WorkspaceInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WorkspaceInfo;
  static deserializeBinaryFromReader(message: WorkspaceInfo, reader: jspb.BinaryReader): WorkspaceInfo;
}

export namespace WorkspaceInfo {
  export type AsObject = {
    id: string,
    name: string,
    trackedPathsList: Array<string>,
    createdAt: string,
    lastSync: string,
    status: WorkspaceStatus,
    metadataMap: Array<[string, string]>,
  }
}

export class SparseCheckoutRequest extends jspb.Message {
  getPathsList(): Array<string>;
  setPathsList(value: Array<string>): SparseCheckoutRequest;
  clearPathsList(): SparseCheckoutRequest;
  addPaths(value: string, index?: number): SparseCheckoutRequest;

  getTargetDir(): string;
  setTargetDir(value: string): SparseCheckoutRequest;

  getWorkspaceId(): string;
  setWorkspaceId(value: string): SparseCheckoutRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SparseCheckoutRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SparseCheckoutRequest): SparseCheckoutRequest.AsObject;
  static serializeBinaryToWriter(message: SparseCheckoutRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SparseCheckoutRequest;
  static deserializeBinaryFromReader(message: SparseCheckoutRequest, reader: jspb.BinaryReader): SparseCheckoutRequest;
}

export namespace SparseCheckoutRequest {
  export type AsObject = {
    pathsList: Array<string>,
    targetDir: string,
    workspaceId: string,
  }
}

export class SparseCheckoutResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): SparseCheckoutResponse;

  getMessage(): string;
  setMessage(value: string): SparseCheckoutResponse;

  getConfiguredPathsList(): Array<string>;
  setConfiguredPathsList(value: Array<string>): SparseCheckoutResponse;
  clearConfiguredPathsList(): SparseCheckoutResponse;
  addConfiguredPaths(value: string, index?: number): SparseCheckoutResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SparseCheckoutResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SparseCheckoutResponse): SparseCheckoutResponse.AsObject;
  static serializeBinaryToWriter(message: SparseCheckoutResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SparseCheckoutResponse;
  static deserializeBinaryFromReader(message: SparseCheckoutResponse, reader: jspb.BinaryReader): SparseCheckoutResponse;
}

export namespace SparseCheckoutResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    configuredPathsList: Array<string>,
  }
}

export class DownloadPathRequest extends jspb.Message {
  getPath(): string;
  setPath(value: string): DownloadPathRequest;

  getBranch(): string;
  setBranch(value: string): DownloadPathRequest;

  getFormat(): string;
  setFormat(value: string): DownloadPathRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DownloadPathRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DownloadPathRequest): DownloadPathRequest.AsObject;
  static serializeBinaryToWriter(message: DownloadPathRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DownloadPathRequest;
  static deserializeBinaryFromReader(message: DownloadPathRequest, reader: jspb.BinaryReader): DownloadPathRequest;
}

export namespace DownloadPathRequest {
  export type AsObject = {
    path: string,
    branch: string,
    format: string,
  }
}

export class DownloadPathResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): DownloadPathResponse;

  getMessage(): string;
  setMessage(value: string): DownloadPathResponse;

  getContent(): Uint8Array | string;
  getContent_asU8(): Uint8Array;
  getContent_asB64(): string;
  setContent(value: Uint8Array | string): DownloadPathResponse;

  getFilename(): string;
  setFilename(value: string): DownloadPathResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DownloadPathResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DownloadPathResponse): DownloadPathResponse.AsObject;
  static serializeBinaryToWriter(message: DownloadPathResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DownloadPathResponse;
  static deserializeBinaryFromReader(message: DownloadPathResponse, reader: jspb.BinaryReader): DownloadPathResponse;
}

export namespace DownloadPathResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    content: Uint8Array | string,
    filename: string,
  }
}

export class AddTrackedPathRequest extends jspb.Message {
  getWorkspaceId(): string;
  setWorkspaceId(value: string): AddTrackedPathRequest;

  getPath(): string;
  setPath(value: string): AddTrackedPathRequest;

  getBranch(): string;
  setBranch(value: string): AddTrackedPathRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddTrackedPathRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AddTrackedPathRequest): AddTrackedPathRequest.AsObject;
  static serializeBinaryToWriter(message: AddTrackedPathRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddTrackedPathRequest;
  static deserializeBinaryFromReader(message: AddTrackedPathRequest, reader: jspb.BinaryReader): AddTrackedPathRequest;
}

export namespace AddTrackedPathRequest {
  export type AsObject = {
    workspaceId: string,
    path: string,
    branch: string,
  }
}

export class AddTrackedPathResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): AddTrackedPathResponse;

  getMessage(): string;
  setMessage(value: string): AddTrackedPathResponse;

  getCommitHash(): string;
  setCommitHash(value: string): AddTrackedPathResponse;

  getNewVersion(): number;
  setNewVersion(value: number): AddTrackedPathResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddTrackedPathResponse.AsObject;
  static toObject(includeInstance: boolean, msg: AddTrackedPathResponse): AddTrackedPathResponse.AsObject;
  static serializeBinaryToWriter(message: AddTrackedPathResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddTrackedPathResponse;
  static deserializeBinaryFromReader(message: AddTrackedPathResponse, reader: jspb.BinaryReader): AddTrackedPathResponse;
}

export namespace AddTrackedPathResponse {
  export type AsObject = {
    success: boolean,
    message: string,
    commitHash: string,
    newVersion: number,
  }
}

export enum WorkspaceStatus { 
  ACTIVE = 0,
  SYNCING = 1,
  ERROR = 2,
  SUSPENDED = 3,
}
