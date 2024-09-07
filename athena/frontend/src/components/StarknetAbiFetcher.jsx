import React, { useState } from 'react';

const StarknetAbiFetcher = () => {
  const [classHash, setClassHash] = useState('');
  const [jsonRpcUrl, setJsonRpcUrl] = useState('');
  const [abi, setAbi] = useState(null);
  const [error, setError] = useState(null);
  const handleFetchAbi = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/abi', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          classHash: classHash,
          jsonRpcUrl: jsonRpcUrl,
        }),
      });
  
      if (!response.ok) {
        throw new Error(`HTTP error ${response.status}`);
      }
  
      const abi = await response.json();
      setAbi(abi);
      setError(null);
    } catch (err) {
      setError(err.message);
      setAbi(null);
    }
  };
  return (
    <div className="p-4">
      <h2 className="text-2xl font-bold mb-4">Fetch StarkNet ABI</h2>

      <div className="space-y-4">
        <div>
          <label htmlFor="classHash" className="block font-medium mb-1">
            Contract Class Hash
          </label>
          <input
            type="text"
            id="classHash"
            className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            value={classHash}
            onChange={(e) => setClassHash(e.target.value)}
            placeholder="Enter the contract class hash"
          />
        </div>

        <div>
          <label htmlFor="jsonRpcUrl" className="block font-medium mb-1">
            JSON-RPC URL
          </label>
          <input
            type="text"
            id="jsonRpcUrl"
            className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            value={jsonRpcUrl}
            onChange={(e) => setJsonRpcUrl(e.target.value)}
            placeholder="Enter the JSON-RPC URL"
          />
        </div>

        <button
          className="bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-md"
          onClick={handleFetchAbi}
        >
          Fetch ABI
        </button>
      </div>

      {abi && (
        <div className="mt-4">
          <h3 className="text-xl font-bold mb-2">ABI Data</h3>
          <pre className="bg-gray-100 p-4 rounded-md overflow-auto">
            {JSON.stringify(abi, null, 2)}
          </pre>
        </div>
      )}

      {error && (
        <div className="mt-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative">
          <strong>Error:</strong> {error}
        </div>
      )}
    </div>
  );
};

export default StarknetAbiFetcher;