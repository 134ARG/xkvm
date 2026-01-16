import { useState, useEffect } from "react";
import { useNativeConfig, type Connection } from "@/stores/nativeConfigStore";
import { Button } from "@/components/Button";
import { InputFieldWithLabel } from "@/components/InputField";
import Card from "@/components/Card";
import { m } from "@localizations/messages.js";

export default function ConnectionSettings() {
  const { config, addConnection, removeConnection, setDefaultConnection, loadConfig } = useNativeConfig();
  const [isAdding, setIsAdding] = useState(false);
  const [newConnection, setNewConnection] = useState({ name: '', url: '' });
  const [errors, setErrors] = useState<{ name?: string; url?: string }>({});
  
  useEffect(() => {
    loadConfig();
  }, [loadConfig]);
  
  const validateUrl = (url: string): boolean => {
    try {
      const parsed = new URL(url);
      return parsed.protocol === 'https:' || parsed.protocol === 'http:';
    } catch {
      return false;
    }
  };
  
  const handleAdd = async () => {
    const newErrors: { name?: string; url?: string } = {};
    
    if (!newConnection.name.trim()) {
      newErrors.name = 'Name is required';
    }
    
    if (!newConnection.url.trim()) {
      newErrors.url = 'URL is required';
    } else if (!validateUrl(newConnection.url)) {
      newErrors.url = 'Invalid URL (must start with http:// or https://)';
    }
    
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }
    
    await addConnection(newConnection.name, newConnection.url);
    setIsAdding(false);
    setNewConnection({ name: '', url: '' });
    setErrors({});
  };
  
  const handleRemove = async (id: string) => {
    if (confirm('Are you sure you want to remove this connection?')) {
      await removeConnection(id);
    }
  };
  
  const formatDate = (dateStr?: string) => {
    if (!dateStr) return 'Never';
    try {
      return new Date(dateStr).toLocaleString();
    } catch {
      return 'Never';
    }
  };
  
  return (
    <div className="space-y-6 p-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          Backend Connections
        </h2>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Manage your XKVM backend server connections
        </p>
      </div>
      
      <div className="space-y-4">
        {config?.connections.map(conn => (
          <Card key={conn.id}>
            <div className="flex items-center justify-between p-4">
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                    {conn.name}
                  </h3>
                  {conn.is_default && (
                    <span className="rounded-full bg-blue-100 px-2 py-1 text-xs font-medium text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                      Default
                    </span>
                  )}
                </div>
                <p className="mt-1 text-sm text-gray-600 dark:text-gray-300">
                  {conn.url}
                </p>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Last connected: {formatDate(conn.last_connected)}
                </p>
              </div>
              <div className="flex gap-2">
                {!conn.is_default && (
                  <Button
                    variant="secondary"
                    onClick={() => setDefaultConnection(conn.id)}
                  >
                    Set as Default
                  </Button>
                )}
                <Button
                  variant="danger"
                  onClick={() => handleRemove(conn.id)}
                  disabled={config.connections.length === 1}
                >
                  Remove
                </Button>
              </div>
            </div>
          </Card>
        ))}
        
        {config?.connections.length === 0 && !isAdding && (
          <Card>
            <div className="p-8 text-center">
              <p className="text-gray-500 dark:text-gray-400">
                No connections configured. Add your first backend connection to get started.
              </p>
            </div>
          </Card>
        )}
      </div>
      
      {isAdding ? (
        <Card>
          <div className="space-y-4 p-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Add New Connection
            </h3>
            <InputFieldWithLabel
              label="Connection Name"
              placeholder="My XKVM Device"
              value={newConnection.name}
              onChange={(e) => setNewConnection({ ...newConnection, name: e.target.value })}
              error={errors.name}
            />
            <InputFieldWithLabel
              label="Backend URL"
              placeholder="https://xkvm.example.com"
              value={newConnection.url}
              onChange={(e) => setNewConnection({ ...newConnection, url: e.target.value })}
              error={errors.url}
            />
            <div className="flex gap-2">
              <Button onClick={handleAdd}>Add Connection</Button>
              <Button 
                variant="secondary" 
                onClick={() => {
                  setIsAdding(false);
                  setNewConnection({ name: '', url: '' });
                  setErrors({});
                }}
              >
                Cancel
              </Button>
            </div>
          </div>
        </Card>
      ) : (
        <Button onClick={() => setIsAdding(true)}>
          Add Connection
        </Button>
      )}
    </div>
  );
}
