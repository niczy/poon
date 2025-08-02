import { NextRequest, NextResponse } from 'next/server';
import { DirectoryItem } from '@/types/monorepo';

// Mock data for development - replace with actual poon-server gRPC calls
const mockData: { [path: string]: DirectoryItem[] } = {
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

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const path = searchParams.get('path') || '/';

    // Add artificial delay to simulate network request
    await new Promise(resolve => setTimeout(resolve, 100 + Math.random() * 200));

    const items = mockData[path] || [];
    
    return NextResponse.json({
      items: items.map(item => ({
        ...item,
        modTime: Math.floor(Date.now() / 1000) - Math.floor(Math.random() * 86400) // Random time within last day
      }))
    });
  } catch (error) {
    console.error('Directory API error:', error);
    return NextResponse.json(
      { error: 'Failed to read directory' },
      { status: 500 }
    );
  }
}