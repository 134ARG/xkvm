import { useState } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/solid";
import { motion, AnimatePresence } from "framer-motion";

import { Button } from "@components/Button";
import { GridCard } from "@components/Card";
import { GitHubIcon } from "@components/Icons";
import { JsonRpcResponse, useJsonRpc } from "@hooks/useJsonRpc";
import { useDeviceUiNavigation } from "@hooks/useAppNavigation";
import { useVersion } from "@hooks/useVersion";
import { useDeviceStore } from "@hooks/stores";
import notifications from "@/notifications";
import { m } from "@localizations/messages.js";

interface FailSafeModeOverlayProps {
  reason: string;
}

interface OverlayContentProps {
  readonly children: React.ReactNode;
}

function OverlayContent({ children }: OverlayContentProps) {
  return (
    <GridCard cardClassName="h-full pointer-events-auto outline-hidden!">
      <div className="flex h-full w-full flex-col items-center justify-center rounded-md border border-slate-800/30 dark:border-slate-300/20">
        {children}
      </div>
    </GridCard>
  );
}

export function FailSafeModeOverlay({ reason }: FailSafeModeOverlayProps) {
  const { send } = useJsonRpc();
  useDeviceUiNavigation();
  useVersion();
  useDeviceStore();
  const [isDownloadingLogs, setIsDownloadingLogs] = useState(false);

  const getReasonCopy = () => {
    switch (reason) {
      case "video":
        return {
          message: m.fail_safe_video_message(),
        };
      default:
        return {
          message: m.fail_safe_default_message(),
        };
    }
  };

  const { message } = getReasonCopy();

  const handleReportAndDownloadLogs = () => {
    setIsDownloadingLogs(true);

    send("getDiagnostics", {}, async (resp: JsonRpcResponse) => {
      setIsDownloadingLogs(false);

      if ("error" in resp) {
        notifications.error(m.fail_safe_diagnostics_get_error({ error: resp.error.message }));
        return;
      }

      // Download logs
      const logContent = resp.result as string;
      const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
      const filename = `xkvm-recovery-${reason}-${timestamp}.txt`;

      const blob = new Blob([logContent], { type: "text/plain" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      await new Promise(resolve => setTimeout(resolve, 1000));
      a.click();
      await new Promise(resolve => setTimeout(resolve, 1000));
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      notifications.success(m.fail_safe_logs_download_success());

      // Open GitHub issue

      const issueUrl = "#";

      window.open(issueUrl, "_blank");
    });
  };

  // OTA downgrade functionality disabled
  /*
  const handleDowngrade = () => {
    navigateTo(`/settings/general/update?custom_app_version=${DOWNGRADE_VERSION}`);
  };
  */

  return (
    <AnimatePresence>
      <motion.div
        className="isolate aspect-video h-full w-full"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0, transition: { duration: 0 } }}
        transition={{
          duration: 0.4,
          ease: "easeInOut",
        }}
      >
        <OverlayContent>
          <div className="flex max-w-lg flex-col items-start gap-y-1">
            <ExclamationTriangleIcon className="h-12 w-12 text-yellow-500" />
            <div className="text-left text-sm text-slate-700 dark:text-slate-300">
              <div className="space-y-4">
                <div className="space-y-2 text-black dark:text-white">
                  <h2 className="text-xl font-bold">{m.fail_safe_activated_title()}</h2>
                  <p className="text-sm">{message}</p>
                </div>
                <div className="space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <Button
                      onClick={handleReportAndDownloadLogs}
                      theme="primary"
                      size="SM"
                      disabled={isDownloadingLogs}
                      LeadingIcon={GitHubIcon}
                      loading={isDownloadingLogs}
                      text={
                        isDownloadingLogs
                          ? m.fail_safe_downloading_logs()
                          : m.fail_safe_download_logs_report_issue()
                      }
                    />

                    {/* OTA downgrade button disabled
                    <Button
                      size="SM"
                      onClick={handleDowngrade}
                      theme="light"
                      text={`Downgrade to v${DOWNGRADE_VERSION}`}
                    />
                    */}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </OverlayContent>
      </motion.div>
    </AnimatePresence>
  );
}
