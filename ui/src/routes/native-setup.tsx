import { useState } from "react";
import { useNativeConfig } from "@/stores/nativeConfigStore";
import { Button } from "@/components/Button";
import { InputFieldWithLabel } from "@/components/InputField";
import Card from "@/components/Card";
import GridBackground from "@/components/GridBackground";
import LogoBlue from "@/assets/logo-blue.svg";
import LogoWhite from "@/assets/logo-white.svg";
import { m } from "@localizations/messages.js";

export default function NativeSetup() {
  const { addConnection, testConnection, switchConnection } = useNativeConfig();
  const [connection, setConnection] = useState({ name: "", url: "" });
  const [errors, setErrors] = useState<{ name?: string; url?: string }>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validateUrl = (url: string): boolean => {
    try {
      const parsed = new URL(url);
      return parsed.protocol === "https:" || parsed.protocol === "http:";
    } catch {
      return false;
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const newErrors: { name?: string; url?: string } = {};

    if (!connection.name.trim()) {
      newErrors.name = m.connection_name_required();
    }

    if (!connection.url.trim()) {
      newErrors.url = m.connection_url_required();
    } else if (!validateUrl(connection.url)) {
      newErrors.url = m.connection_url_invalid();
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    setIsSubmitting(true);
    try {
      // Verify the backend is reachable before saving.
      const test = await testConnection(connection.url);
      if (!test.ok) {
        setErrors({ url: m.connection_test_failed({ error: test.error ?? "" }) });
        setIsSubmitting(false);
        return;
      }

      const added = await addConnection(connection.name, connection.url);

      // Activate it; switchConnection forces a full reload and config re-check.
      await switchConnection(added.id);
    } catch (error) {
      console.error("Failed to add connection:", error);
      setErrors({ url: m.connection_save_error() });
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
            alt={m.xkvm_logo_alt()}
            className="mx-auto mb-4 h-16 w-16 dark:hidden"
          />
          <img
            src={LogoWhite}
            alt={m.xkvm_logo_alt()}
            className="mx-auto mb-4 hidden h-16 w-16 dark:block"
          />
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            {m.welcome_to_xkvm()}
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">{m.native_setup_description()}</p>
        </div>

        <Card>
          <form onSubmit={handleSubmit} className="space-y-4 p-6">
            <InputFieldWithLabel
              label={m.connection_name_label()}
              placeholder={m.connection_name_placeholder()}
              value={connection.name}
              onChange={e => setConnection({ ...connection, name: e.target.value })}
              error={errors.name}
              autoFocus
            />
            <InputFieldWithLabel
              label={m.connection_backend_url_label()}
              placeholder="https://xkvm.example.com"
              value={connection.url}
              onChange={e => setConnection({ ...connection, url: e.target.value })}
              error={errors.url}
            />
            <div className="pt-2">
              <Button
                type="submit"
                size="LG"
                theme="primary"
                fullWidth
                disabled={isSubmitting}
                text={isSubmitting ? m.connecting() : m.connect()}
              />
            </div>
          </form>
        </Card>

        <p className="mt-4 text-center text-sm text-gray-500 dark:text-gray-400">
          {m.native_setup_footer()}
        </p>
      </div>
    </div>
  );
}
