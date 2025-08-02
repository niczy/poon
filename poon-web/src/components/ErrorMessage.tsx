'use client';

import React from 'react';

interface ErrorMessageProps {
  message: string;
  onRetry?: () => void;
}

export const ErrorMessage: React.FC<ErrorMessageProps> = ({ message, onRetry }) => {
  return (
    <div className="flex items-start gap-4 p-6 m-6 bg-red-50 border border-red-200 rounded-lg text-red-800">
      <div className="text-2xl flex-shrink-0">❌</div>
      <div className="flex-1">
        <div className="text-lg font-semibold mb-2">Error</div>
        <div className="mb-4 leading-relaxed">{message}</div>
        {onRetry && (
          <button 
            className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors flex items-center gap-2"
            onClick={onRetry}
          >
            🔄 Retry
          </button>
        )}
      </div>
    </div>
  );
};