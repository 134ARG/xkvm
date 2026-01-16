import { useState } from "react";
import { useNavigate } from "react-router";
import { useNativeConfig } from "@/stores/nativeConfigStore";
import { Button } from "@/components/Button";
import { InputFieldWithLabel } from "@/components/InputField";
import Card from "@/components/Card";
import GridBackground from "@/components/GridBackground";
import LogoBlue from "@/assets/logo-blue.svg";
import LogoWhite from "@/assets/logo-white.svg";

export default function NativeSetup() {
  const navigate = useNavigate();
  const { addConnection } = useNativeConfig();
  const [connection, setConnection] = useState({ name: '', url: '' });
  const [errors, setErrors] = useState<{ name?: string; url?: string }>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const validateUrl = (url: string): boolean => {
    try {
      const parsed = new URL(url);
      return parsed.protocol === 'https:' || parsed.protocol === 'http:';
    } catch {
      return false;
    }
  };
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const newErrors: { name?: string; url?: string } = {};
    
    if (!connection.name.trim()) {
      newErrors.name = 'Name is required';
    }
    
    if (!connection.url.trim()) {
      newErrors.url = 'URL is required';
    } else if (!validateUrl(connection.url)) {
      newErrors.url = 'Invalid URL (must start with http:// or https://)';
    }
    
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }
    
    setIsSubmitting(true);
    try {
      // Add connection and wait for it to be set as current
      await addConnection(connection.name, connection.url);
      
      // Navigate to home page (forces full reload and config re-check)
      window.location.href = '/';
    } catch (error) {
      console.error('Failed to add connection:', error);
      setErrors({ url: 'Failed to save connection. Please try again.' });
      setIsSubmitting(false);
    }
  };
  
  return (
    <div className="relative flex h-full w-full items-center justify-center">
      <GridBackground />
      <div className="z-10 w-full max-w-md px-4">
        <div className="mb-8 text-center">
          <img
            src={LogoBlue}
            alt="XKVM Logo"
            className="mx-auto mb-4 h-16 w-16 dark:hidden"
          />
          <img
            src={LogoWhite}
            alt="XKVM Logo"
            className="mx-auto mb-4 hidden h-16 w-16 dark:block"
          />
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Welcome to XKVM
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Connect to your XKVM backend to get started
          </p>
        </div>
        
        <Card>
          <form onSubmit={handleSubmit} className="space-y-4 p-6">
            <InputFieldWithLabel
              label="Connection Name"
              placeholder="My XKVM Device"
              value={connection.name}
              onChange={(e) => setConnection({ ...connection, name: e.target.value })}
              error={errors.name}
              autoFocus
            />
            <InputFieldWithLabel
              label="Backend URL"
              placeholder="https://xkvm.example.com"
              value={connection.url}
              onChange={(e) => setConnection({ ...connection, url: e.target.value })}
              error={errors.url}
            />
            <div className="pt-2">
              <Button
                type="submit"
                size="LG"
                theme="primary"
                fullWidth
                disabled={isSubmitting}
                text={isSubmitting ? 'Connecting...' : 'Connect'}
              />
            </div>
          </form>
        </Card>
        
        <p className="mt-4 text-center text-sm text-gray-500 dark:text-gray-400">
          You can manage multiple connections later in Settings
        </p>
      </div>
    </div>
  );
}
