'use client';

import React, { useState, useEffect } from 'react';
import { DirectoryItem } from '@/proto/monorepo_pb';
import { defaultService } from '@/services/monorepoService';
import { DirectoryView } from './DirectoryView';
import { FileView } from './FileView';
import { Breadcrumbs } from './Breadcrumbs';
import { LoadingSpinner } from './LoadingSpinner';
import { ErrorMessage } from './ErrorMessage';

interface FileBrowserProps {
  initialPath?: string;
}

export const FileBrowser: React.FC<FileBrowserProps> = ({ initialPath = '/' }) => {
  const [currentPath, setCurrentPath] = useState<string>(initialPath);
  const [items, setItems] = useState<DirectoryItem[]>([]);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const loadDirectory = async (path: string) => {
    setLoading(true);
    setError(null);
    setSelectedFile(null);
    
    try {
      const response = await defaultService.readDirectory({ path });
      setItems(response.items);
      setCurrentPath(path);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load directory');
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  const handleItemClick = (item: DirectoryItem) => {
    const itemPath = currentPath === '/' ? `/${item.name}` : `${currentPath}/${item.name}`;
    
    if (item.isDir) {
      loadDirectory(itemPath);
    } else {
      setSelectedFile(itemPath);
    }
  };

  const handlePathClick = (path: string) => {
    if (path !== currentPath) {
      loadDirectory(path);
    }
  };

  const handleBackToDirectory = () => {
    setSelectedFile(null);
  };

  useEffect(() => {
    loadDirectory(initialPath);
  }, [initialPath]);

  if (selectedFile) {
    return (
      <div className="bg-white rounded-xl shadow-lg overflow-hidden">
        <Breadcrumbs 
          currentPath={selectedFile} 
          onPathClick={handlePathClick}
          showFileName={true}
        />
        <FileView 
          filePath={selectedFile} 
          onBack={handleBackToDirectory}
        />
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl shadow-lg overflow-hidden">
      <Breadcrumbs 
        currentPath={currentPath} 
        onPathClick={handlePathClick}
      />
      
      {loading && <LoadingSpinner />}
      {error && <ErrorMessage message={error} onRetry={() => loadDirectory(currentPath)} />}
      
      {!loading && !error && (
        <DirectoryView 
          items={items}
          currentPath={currentPath}
          onItemClick={handleItemClick}
        />
      )}
    </div>
  );
};