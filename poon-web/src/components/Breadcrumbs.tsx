'use client';

import React from 'react';

interface BreadcrumbsProps {
  currentPath: string;
  onPathClick: (path: string) => void;
  showFileName?: boolean;
}

interface Breadcrumb {
  name: string;
  path: string;
  isFile?: boolean;
}

export const Breadcrumbs: React.FC<BreadcrumbsProps> = ({ 
  currentPath, 
  onPathClick, 
  showFileName = false 
}) => {
  const generateBreadcrumbs = (): Breadcrumb[] => {
    const parts = currentPath.split('/').filter(part => part !== '');
    const breadcrumbs: Breadcrumb[] = [{ name: 'Home', path: '/' }];
    
    let currentBreadcrumbPath = '';
    for (let i = 0; i < parts.length; i++) {
      currentBreadcrumbPath += '/' + parts[i];
      
      // If this is the last part and showFileName is false, skip it (it's a directory)
      // If showFileName is true, the last part is a file and should be shown but not clickable
      const isLast = i === parts.length - 1;
      const isFile = showFileName && isLast;
      
      breadcrumbs.push({
        name: parts[i],
        path: currentBreadcrumbPath,
        isFile
      });
    }
    
    return breadcrumbs;
  };

  const breadcrumbs = generateBreadcrumbs();

  return (
    <nav className="bg-gray-50 border-b border-gray-200 px-6 py-4">
      <div className="flex flex-wrap items-center gap-2 mb-2">
        {breadcrumbs.map((breadcrumb, index) => (
          <React.Fragment key={breadcrumb.path}>
            {index > 0 && <span className="text-gray-400 font-medium">/</span>}
            
            {breadcrumb.isFile ? (
              <span className="flex items-center gap-1 px-2 py-1 bg-green-100 text-green-800 rounded text-sm font-medium">
                📄 {breadcrumb.name}
              </span>
            ) : (
              <button
                className={`flex items-center gap-1 px-2 py-1 rounded text-sm font-medium transition-colors ${
                  breadcrumb.path === currentPath 
                    ? 'bg-blue-500 text-white' 
                    : 'text-gray-600 hover:bg-gray-200 hover:text-gray-900'
                }`}
                onClick={() => onPathClick(breadcrumb.path)}
              >
                {index === 0 ? '🏠' : '📁'} {breadcrumb.name}
              </button>
            )}
          </React.Fragment>
        ))}
      </div>
      
      <div className="text-xs text-gray-500">
        <code className="bg-gray-200 px-2 py-1 rounded">{currentPath}</code>
      </div>
    </nav>
  );
};