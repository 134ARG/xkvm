import { useState, useEffect } from "react";
import { LuCircleCheck, LuCircleX } from "react-icons/lu";

import { useNativeConfig, Connection } from "@/stores/nativeConfigStore";
import { Button } from "@/components/Button";
import { InputFieldWithLabel } from "@/components/InputField";
import { SelectMenuBasic } from "@/components/SelectMenuBasic";
import { SettingsItem } from "@/components/SettingsItem";
import Card from "@/components/Card";
import LoadingSpinner from "@/components/LoadingSpinner";
import { m } from "@localizations/messages.js";

type TestState = "idle" | "testing" | "ok" | "failed";

const validateUrl = (url: string): boolean => {
  try {
    const parsed = new URL(url);
    return parsed.protocol === "https:" || parsed.protocol === "http:";
  } catch {
    return false;
  }
};

export default function ConnectionSettings() {
  const {
    config,
    currentConnection,
    isLoading,
    error,
    addConnection,
    updateConnection,
    testConnection,
    switchConnection,
    setDefaultConnection,
    removeConnection,
    updateSettings,
    loadConfig,
  } = useNativeConfig();

  // editingId: null = closed, "" = adding new, otherwise editing that connection id
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({ name: "", url: "" });
  const [errors, setErrors] = useState<{ name?: string; url?: string }>({});
  const [testState, setTestState] = useState<TestState>("idle");
  const [testMessage, setTestMessage] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const openAdd = () => {
    setEditingId("");
    setForm({ name: "", url: "" });
    setErrors({});
    setTestState("idle");
    setTestMessage(null);
  };

  const openEdit = (conn: Connection) => {
    setEditingId(conn.id);
    setForm({ name: conn.name, url: conn.url });
    setErrors({});
    setTestState("idle");
    setTestMessage(null);
  };

  const closeForm = () => {
    setEditingId(null);
    setForm({ name: "", url: "" });
    setErrors({});
    setTestState("idle");
    setTestMessage(null);
  };

  const validateForm = () => {
    const newErrors: { name?: string; url?: string } = {};
    if (!form.name.trim()) newErrors.name = m.connection_name_required();
    if (!form.url.trim()) {
      newErrors.url = m.connection_url_required();
    } else if (!validateUrl(form.url)) {
      newErrors.url = m.connection_url_invalid();
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const runTest = async () => {
    if (!form.url.trim() || !validateUrl(form.url)) {
      setErrors(prev => ({ ...prev, url: m.connection_url_invalid() }));
      return false;
    }
    setTestState("testing");
    setTestMessage(null);
    const result = await testConnection(form.url);
    if (result.ok) {
      setTestState("ok");
      setTestMessage(m.connection_test_success());
    } else {
      setTestState("failed");
      setTestMessage(m.connection_test_failed({ error: result.error ?? "" }));
    }
    return result.ok;
  };

  const handleSave = async () => {
    if (!validateForm()) return;

    setIsSaving(true);
    try {
      const ok = await runTest();
      if (!ok) return;

      if (editingId) {
        await updateConnection(editingId, form.name, form.url);
      } else {
        await addConnection(form.name, form.url);
      }
      closeForm();
    } finally {
      setIsSaving(false);
    }
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

      {error && (
        <div className="rounded-md bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/30 dark:text-red-300">
          {m.connection_load_error({ error })}
        </div>
      )}

      {isLoading && !config && (
        <div className="flex justify-center py-8">
          <LoadingSpinner className="h-6 w-6 animate-spin text-blue-600" />
        </div>
      )}

      <div className="space-y-4">
        {config?.connections.map(conn => {
          const isActive = currentConnection?.id === conn.id;
          const isEditing = editingId === conn.id;
          return (
            <Card key={conn.id}>
              <div className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                        {conn.name}
                      </h3>
                      {isActive && (
                        <span className="rounded-full bg-green-100 px-2 py-1 text-xs font-medium text-green-800 dark:bg-green-900 dark:text-green-200">
                          {m.connection_connected()}
                        </span>
                      )}
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
                    {!isActive && (
                      <Button
                        text={m.connection_connect()}
                        theme="primary"
                        size="SM"
                        onClick={() => switchConnection(conn.id)}
                      />
                    )}
                    {!conn.is_default && (
                      <Button
                        text={m.connection_set_as_default()}
                        theme="light"
                        size="SM"
                        onClick={() => setDefaultConnection(conn.id)}
                      />
                    )}
                    <Button
                      text={m.connection_edit()}
                      theme="light"
                      size="SM"
                      onClick={() => (isEditing ? closeForm() : openEdit(conn))}
                    />
                    <Button
                      text={m.connection_remove()}
                      theme="danger"
                      size="SM"
                      onClick={() => handleRemove(conn.id)}
                      disabled={config.connections.length === 1}
                    />
                  </div>
                </div>

                {isEditing && (
                  <ConnectionForm
                    title={m.connection_edit_title()}
                    form={form}
                    errors={errors}
                    testState={testState}
                    testMessage={testMessage}
                    isSaving={isSaving}
                    onChange={setForm}
                    onTest={runTest}
                    onSave={handleSave}
                    onCancel={closeForm}
                    saveLabel={m.connection_save()}
                  />
                )}
              </div>
            </Card>
          );
        })}

        {config?.connections.length === 0 && editingId === null && (
          <Card>
            <div className="p-8 text-center">
              <p className="text-gray-500 dark:text-gray-400">{m.connection_no_connections()}</p>
            </div>
          </Card>
        )}
      </div>

      {editingId === "" ? (
        <Card>
          <div className="p-4">
            <ConnectionForm
              title={m.connection_add_new()}
              form={form}
              errors={errors}
              testState={testState}
              testMessage={testMessage}
              isSaving={isSaving}
              onChange={setForm}
              onTest={runTest}
              onSave={handleSave}
              onCancel={closeForm}
              saveLabel={m.connection_add()}
            />
          </div>
        </Card>
      ) : (
        editingId === null && (
          <Button text={m.connection_add()} theme="primary" size="SM" onClick={openAdd} />
        )
      )}

      {config && (
        <div className="space-y-4 border-t border-gray-200 pt-6 dark:border-gray-700">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
            {m.connection_settings_title()}
          </h3>
          <SettingsItem
            title={m.connection_startup_title()}
            description={m.connection_startup_description()}
          >
            <SelectMenuBasic
              size="SM"
              label=""
              value={config.settings.remember_last_connection ? "last" : "default"}
              options={[
                { value: "default", label: m.connection_startup_default() },
                { value: "last", label: m.connection_startup_last() },
              ]}
              onChange={e =>
                updateSettings({
                  auto_connect: true,
                  remember_last_connection: e.target.value === "last",
                })
              }
            />
          </SettingsItem>
        </div>
      )}
    </div>
  );
}

interface ConnectionFormProps {
  title: string;
  form: { name: string; url: string };
  errors: { name?: string; url?: string };
  testState: TestState;
  testMessage: string | null;
  isSaving: boolean;
  saveLabel: string;
  onChange: (form: { name: string; url: string }) => void;
  onTest: () => void;
  onSave: () => void;
  onCancel: () => void;
}

function ConnectionForm({
  title,
  form,
  errors,
  testState,
  testMessage,
  isSaving,
  saveLabel,
  onChange,
  onTest,
  onSave,
  onCancel,
}: ConnectionFormProps) {
  return (
    <div className="mt-4 space-y-4 border-t border-gray-200 pt-4 dark:border-gray-700">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h3>
      <InputFieldWithLabel
        label={m.connection_name_label()}
        placeholder={m.connection_name_placeholder()}
        value={form.name}
        onChange={e => onChange({ ...form, name: e.target.value })}
        error={errors.name}
      />
      <InputFieldWithLabel
        label={m.connection_backend_url_label()}
        placeholder="https://xkvm.example.com"
        value={form.url}
        onChange={e => onChange({ ...form, url: e.target.value })}
        error={errors.url}
      />

      {testMessage && (
        <div
          className={
            testState === "ok"
              ? "flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400"
              : "flex items-center gap-1.5 text-sm text-red-600 dark:text-red-400"
          }
        >
          {testState === "ok" ? (
            <LuCircleCheck className="h-4 w-4" />
          ) : (
            <LuCircleX className="h-4 w-4" />
          )}
          <span>{testMessage}</span>
        </div>
      )}

      <div className="flex gap-2">
        <Button
          text={saveLabel}
          theme="primary"
          size="SM"
          onClick={onSave}
          loading={isSaving}
          disabled={isSaving}
        />
        <Button
          text={testState === "testing" ? m.connection_testing() : m.connection_test()}
          theme="light"
          size="SM"
          onClick={onTest}
          loading={testState === "testing"}
          disabled={isSaving || testState === "testing"}
        />
        <Button text={m.cancel()} theme="light" size="SM" onClick={onCancel} disabled={isSaving} />
      </div>
    </div>
  );
}
