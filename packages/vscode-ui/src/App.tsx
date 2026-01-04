import { useState, useEffect } from 'react';
import { OutputPanel, type ExecuteResponse } from '@yapi/ui';
import { getVSCodeAPI } from './vscode-api';

const vscode = getVSCodeAPI();

function App() {
  const [result, setResult] = useState<ExecuteResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Listen for messages from extension host
    const handleMessage = (event: MessageEvent) => {
      const message = event.data;

      switch (message.type) {
        case 'setLoading':
          setIsLoading(message.loading);
          break;
        case 'setResult':
          setResult(message.result);
          setIsLoading(false);
          break;
        case 'clearResult':
          setResult(null);
          setIsLoading(false);
          break;
      }
    };

    window.addEventListener('message', handleMessage);

    // Request current state on mount
    vscode.postMessage({ type: 'ready' });

    return () => {
      window.removeEventListener('message', handleMessage);
    };
  }, []);

  return (
    <div className="w-full h-full">
      <OutputPanel result={result} isLoading={isLoading} />
    </div>
  );
}

export default App;
