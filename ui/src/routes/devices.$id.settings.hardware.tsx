import { SettingsPageHeader } from "@components/SettingsPageheader";
import { UsbDeviceSetting } from "@components/UsbDeviceSetting";
import { UsbInfoSetting } from "@components/UsbInfoSetting";
import { GPIOSetting } from "@components/GPIOSetting";
import { m } from "@localizations/messages.js";

export default function SettingsHardwareRoute() {
  return (
    <div className="space-y-4">
      <SettingsPageHeader title={m.hardware_title()} description={m.hardware_page_description()} />

      <UsbDeviceSetting />

      <UsbInfoSetting />

      <GPIOSetting />
    </div>
  );
}
