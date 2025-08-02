'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { defaultService } from '@/services/monorepoService';
import { LoadingSpinner } from './LoadingSpinner';
import { ErrorMessage } from './ErrorMessage';

interface FileViewProps {
  filePath: string;
  onBack: () => void;
}

export const FileView: React.FC<FileViewProps> = ({ filePath, onBack }) => {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [fileSize, setFileSize] = useState<number>(0);
  const [isText, setIsText] = useState<boolean>(true);

  const loadFile = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await defaultService.readFile({ path: filePath });
      const decoder = new TextDecoder('utf-8');
      const textContent = decoder.decode(response.content);
      
      // Simple check for binary content (presence of null bytes)
      const isBinary = textContent.includes('\0');
      
      setContent(textContent);
      setFileSize(response.size);
      setIsText(!isBinary);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load file');
    } finally {
      setLoading(false);
    }
  }, [filePath]);

  const getLanguageFromExtension = (filename: string): string => {
    const ext = filename.split('.').pop()?.toLowerCase();
    const langMap: { [key: string]: string } = {
      'js': 'javascript',
      'jsx': 'javascript',
      'ts': 'typescript',
      'tsx': 'typescript',
      'go': 'go',
      'py': 'python',
      'java': 'java',
      'cpp': 'cpp',
      'c': 'c',
      'rs': 'rust',
      'md': 'markdown',
      'html': 'html',
      'css': 'css',
      'scss': 'scss',
      'json': 'json',
      'yaml': 'yaml',
      'yml': 'yaml',
      'xml': 'xml',
      'sh': 'bash',
      'bash': 'bash'
    };
    return langMap[ext || ''] || 'text';
  };

  const formatFileSize = (bytes: number): string => {
    const sizes = ['B', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const downloadFile = () => {
    const blob = new Blob([content], { type: 'application/octet-stream' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filePath.split('/').pop() || 'file';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  useEffect(() => {
    loadFile();
  }, [filePath, loadFile]);

  const fileName = filePath.split('/').pop() || '';
  const language = getLanguageFromExtension(fileName);

  return (
    <div className="p-6">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 pb-4 border-b border-gray-200">
        <div className="flex items-center gap-3 mb-4 md:mb-0">
          <span className="text-2xl">📄</span>
          <span className="text-xl font-semibold text-gray-900">{fileName}</span>
        </div>
        <div className="flex gap-3 w-full md:w-auto">
          <button 
            className="flex-1 md:flex-initial px-4 py-2 bg-gray-500 text-white rounded-lg hover:bg-gray-600 transition-colors flex items-center justify-center gap-2"
            onClick={onBack}
          >
            ← Back
          </button>
          <button 
            className="flex-1 md:flex-initial px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors flex items-center justify-center gap-2"
            onClick={downloadFile}
          >
            📥 Download
          </button>
        </div>
      </div>

      <div className="flex gap-6 mb-6 text-sm text-gray-600">
        <span>Size: {formatFileSize(fileSize)}</span>
        <span>Type: {isText ? `${language} (text)` : 'Binary file'}</span>
      </div>

      {loading && <LoadingSpinner />}
      {error && <ErrorMessage message={error} onRetry={loadFile} />}

      {!loading && !error && (
        <>
          {isText ? (
            <div className="bg-gray-50 border border-gray-200 rounded-lg overflow-hidden">
              <pre className="p-6 overflow-x-auto text-sm font-mono leading-relaxed bg-white text-gray-900 whitespace-pre-wrap break-words">
                <code>{content}</code>
              </pre>
            </div>
          ) : (
            <div className="text-center py-16 bg-gray-50 border border-gray-200 rounded-lg">
              <div className="text-6xl mb-4">🔒</div>
              <div className="text-lg text-gray-600 mb-6">
                This is a binary file and cannot be displayed in the browser.
              </div>
              <button 
                className="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors flex items-center justify-center gap-2 mx-auto"
                onClick={downloadFile}
              >
                📥 Download File
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};