import { MonorepoServiceClientImpl } from '@/proto/monorepo_grpc_web_pb';

// Create gRPC client instance
const grpcClient = new MonorepoServiceClientImpl('http://localhost:8080');

// Export the gRPC client as the service
export const monorepoService = grpcClient;

// Use the gRPC client as default
export const defaultService = grpcClient;