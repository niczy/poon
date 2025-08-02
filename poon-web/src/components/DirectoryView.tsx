'use client';

import React from 'react';
import { DirectoryItem } from '@/proto/monorepo_pb';

interface DirectoryViewProps {
  items: DirectoryItem[];
  currentPath: string;
  onItemClick: (item: DirectoryItem) => void;
}

export const DirectoryView: React.FC<DirectoryViewProps> = ({ 
  items, 
  onItemClick 
}) => {
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '-';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const formatDate = (timestamp: number): string => {
    return new Date(timestamp * 1000).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getFileIcon = (item: DirectoryItem): string => {
    if (item.isDir) return '📁';
    
    const ext = item.name.split('.').pop()?.toLowerCase();
    const iconMap: { [key: string]: string } = {
      'js': '📄',
      'ts': '📘',
      'tsx': '📘',
      'jsx': '📄',
      'go': '🐹',
      'py': '🐍',
      'java': '☕',
      'cpp': '⚙️',
      'c': '⚙️',
      'rs': '🦀',
      'md': '📝',
      'txt': '📄',
      'json': '📋',
      'yaml': '📋',
      'yml': '📋',
      'xml': '📋',
      'html': '🌐',
      'css': '🎨',
      'scss': '🎨',
      'png': '🖼️',
      'jpg': '🖼️',
      'jpeg': '🖼️',
      'gif': '🖼️',
      'svg': '🖼️',
      'pdf': '📕',
      'zip': '📦',
      'tar': '📦',
      'gz': '📦'
    };
    
    return iconMap[ext || ''] || '📄';
  };

  if (items.length === 0) {
    return (
      <div className="p-6">
        <div className="text-center py-16 text-gray-500">
          <div className="text-6xl mb-4">📭</div>
          <div className="text-lg">This directory is empty</div>
        </div>
      </div>
    );
  }

  // Sort items: directories first, then files, both alphabetically
  const sortedItems = [...items].sort((a, b) => {
    if (a.isDir !== b.isDir) {
      return a.isDir ? -1 : 1;
    }
    return a.name.localeCompare(b.name);
  });

  return (
    <div className="p-6">
      <div className="border border-gray-200 rounded-lg overflow-hidden">
        <div className="hidden md:grid md:grid-cols-3 gap-4 px-4 py-3 bg-gray-50 font-semibold text-sm text-gray-700">
          <span>Name</span>
          <span className="text-right">Size</span>
          <span className="text-right">Modified</span>
        </div>
        
        {sortedItems.map((item, index) => (
          <div 
            key={`${item.name}-${index}`}
            className={`grid grid-cols-1 md:grid-cols-3 gap-4 px-4 py-3 cursor-pointer transition-colors border-b border-gray-100 last:border-b-0 hover:bg-gray-50 ${
              item.isDir ? 'hover:bg-blue-50' : ''
            }`}
            onClick={() => onItemClick(item)}
          >
            <div className="flex items-center gap-3 min-w-0">
              <span className="text-xl flex-shrink-0">{getFileIcon(item)}</span>
              <span className={`font-medium truncate ${
                item.isDir ? 'text-blue-600' : 'text-gray-900'
              }`}>
                {item.name}
                {item.isDir && '/'}
              </span>
            </div>
            <span className="text-sm text-gray-500 md:text-right">
              {formatFileSize(item.size)}
            </span>
            <span className="text-sm text-gray-500 md:text-right">
              {formatDate(item.modTime)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};