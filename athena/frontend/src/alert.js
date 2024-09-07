// src/components/ui/alert.js
import React from 'react';

export const Alert = ({ children, className, variant = 'default' }) => {
  // Define different styles for each alert type
  const variantStyles = {
    default: 'bg-blue-100 border-blue-500 text-blue-700',
    success: 'bg-green-100 border-green-500 text-green-700',
    destructive: 'bg-red-100 border-red-500 text-red-700',
  };

  return (
    <div className={`p-4 border-l-4 rounded-md ${variantStyles[variant]} ${className}`}>
      {children}
    </div>
  );
};

export const AlertTitle = ({ children }) => (
  <h2 className="font-bold text-lg mb-2">{children}</h2>
);

export const AlertDescription = ({ children }) => (
  <p>{children}</p>
);
