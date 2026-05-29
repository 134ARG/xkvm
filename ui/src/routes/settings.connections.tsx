import { useState, useEffect } from "react";
import { useNativeConfig } from "@/stores/nativeConfigStore";
import { Button } from "@/components/Button";
import { InputFieldWithLabel } from "@/components/InputField";
import Card from "@/components/Card";
import { m } from "@localizations/messages.js";

export default function ConnectionSettings() {
  const { config, addConnection, removeConnection, setDefaultConnection, loadConfig } =
    useNativeConfig();
  const [isAdding, setIsAdding] = useState(false);
  const [newConnection, setNewConnection] = useState({ name: "", url: "" });
  const [errors, setErrors] = useState<{ name?: string; url?: string }>({});

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const validateUrl = (url: string): boolean => {
    try {
      const parsed = new URL(url);
      return parsed.protocol === "https:" || parsed.protocol === "http:";
    } catch {
      return false;
    }
  };

  const handleAdd = async () => {
    const newErrors: { name?: string; url?: string } = {};

    if (!newConnection.name.trim()) {
      newErrors.name = m.connection_name_required();
    }

    if (!newConnection.url.trim()) {
      newErrors.url = m.connection_url_required();
    } else if (!validateUrl(newConnection.url)) {
      newErrors.url = m.connection_url_invalid();
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    await addConnection(newConnection.name, newConnection.url);
    setIsAdding(false);
    setNewConnection({ name: "", url: "" });
    setErrors({});
  };

  const handleRemove = async (id: string) => {
    if (confirm(m.connection_confirm_remove())) {
      await removeConnection(id);
    }
  };

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return m.date_never();
    try {
      return new Date(dateStr).toLocaleString();
    } catch {
      return m.date_never();
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">{m.connection_title()}</h2>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {m.connection_manage_description()}
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
                      {m.connection_default()}
                    </span>
                  )}
                </div>
                <p className="mt-1 text-sm text-gray-600 dark:text-gray-300">{conn.url}</p>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {m.connection_last_connected()} {formatDate(conn.last_connected)}
                </p>
              </div>
              <div className="flex gap-2">
                {!conn.is_default && (
                  <Button
                    text={m.connection_set_as_default()}
                    theme="light"
                    size="SM"
                    onClick={() => setDefaultConnection(conn.id)}
                  />
                )}
                <Button
                  text={m.connection_remove()}
                  theme="danger"
                  size="SM"
                  onClick={() => handleRemove(conn.id)}
                  disabled={config.connections.length === 1}
                />
              </div>
            </div>
          </Card>
        ))}

        {config?.connections.length === 0 && !isAdding && (
          <Card>
            <div className="p-8 text-center">
              <p className="text-gray-500 dark:text-gray-400">{m.connection_no_connections()}</p>
            </div>
          </Card>
        )}
      </div>

      {isAdding ? (
        <Card>
          <div className="space-y-4 p-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              {m.connection_add_new()}
            </h3>
            <InputFieldWithLabel
              label={m.connection_name_label()}
              placeholder={m.connection_name_placeholder()}
              value={newConnection.name}
              onChange={e => setNewConnection({ ...newConnection, name: e.target.value })}
              error={errors.name}
            />
            <InputFieldWithLabel
              label={m.connection_backend_url_label()}
              placeholder="https://xkvm.example.com"
              value={newConnection.url}
              onChange={e => setNewConnection({ ...newConnection, url: e.target.value })}
              error={errors.url}
            />
            <div className="flex gap-2">
              <Button text={m.connection_add()} theme="primary" size="SM" onClick={handleAdd} />
              <Button
                text={m.cancel()}
                theme="light"
                size="SM"
                onClick={() => {
                  setIsAdding(false);
                  setNewConnection({ name: "", url: "" });
                  setErrors({});
                }}
              />
            </div>
          </div>
        </Card>
      ) : (
        <Button
          text={m.connection_add()}
          theme="primary"
          size="SM"
          onClick={() => setIsAdding(true)}
        />
      )}
    </div>
  );
}
