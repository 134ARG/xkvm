import { useCallback, useEffect, useState } from "react";

import { useSettingsStore } from "@hooks/stores";
import { Button } from "@components/Button";
import { TextAreaWithLabel } from "@components/TextArea";
import { JsonRpcResponse, useJsonRpc } from "@/hooks/useJsonRpc";
import { SettingsItem } from "@components/SettingsItem";
import { SettingsPageHeader } from "@components/SettingsPageheader";
import { SelectMenuBasic } from "@components/SelectMenuBasic";
import { NestedSettingsGroup } from "@components/NestedSettingsGroup";
import Fieldset from "@components/Fieldset";
import notifications from "@/notifications";
import { supportsH265WebRTC } from "@/utils";
import { m } from "@localizations/messages.js";

const defaultEdid =
  "00ffffffffffff0052620188008888881c150103800000780a0dc9a05747982712484c00000001010101010101010101010101010101023a801871382d40582c4500c48e2100001e011d007251d01e206e285500c48e2100001e000000fc00543734392d6648443732300a20000000fd00147801ff1d000a202020202020017b";
const xkvmUltimateEdid =
  "00ffffffffffff00617601000100000016240104a23c22781eee95a3544c99260f5054210800d1c0b300a9c0950081808140810081c0023a801871382d40582c450058522100001e70c200a0a0a055503020350058522100001a000000fd0018781eff3c010a202020202020000000fc00584b564d0a202020202020202001760203307051901f222120041301020311125d5e5f6061230907078301000067030c001000007866d85dc4017888e2000008e80030f2705a80b0588a0058522100001e1dd800a0a0405f603020360058772100001aef6800a0a0402e603020360058772100001a7a5e00a0a0a02a503020350058522100001a00000000000000b3";
const hdmi2k120Edid =
  "00ffffffffffff0031d8341200000000221a0103806036780fee91a3544c99260f50542fcf00315945598180814090409500a940b3006fc200a0a0a055503020" +
  "3500c01c3200001a000000fd0030781ec83c000a202020202020000000fc0068646d692d326b2d3132300a200000000000000000000000000000000000000176" +
  "020330f148101f041322021101230907078301000067030c001000003c67d85dc401788008e200cae3050000e30601006fc200a0a0a0555030203500c01c3200" +
  "001a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000054";
const edids = [
  {
    value: defaultEdid,
    label: m.video_edid_xkvm_default(),
  },
  {
    value: xkvmUltimateEdid,
    label: m.video_edid_xkvm_ultimate(),
  },
  {
    value: hdmi2k120Edid,
    label: m.video_edid_hdmi_2k_120hz(),
  },
  {
    value:
      "00ffffffffffff0031d8341200000000221a0103806036780fee91a3544c99260f50542fcf00315945598180814090409500a940b30008e80030f2705a80b0588a00c01c3200001e000000fd00185518873c000a202020202020000000fc0068646d692d346b2d3630300a20000000100000000000000000000000000000012c020342f15161605f5e5d101f0413222120051402110123090707830100006d030c001000003c21006001020367d85dc401788008e200cae3050000e3060100e20d614dd000a0f0703e8030203500c01c3200001e1a3680a070381f4030203500c01c3200001a1a1d008051d01c2040803500c01c3200001c00000000000000a7",
    label: m.video_edid_hdmi_4k_600mhz(),
  },
  {
    value:
      "00FFFFFFFFFFFF00047265058A3F6101101E0104A53420783FC125A8554EA0260D5054BFEF80714F8140818081C081008B009500B300283C80A070B023403020360006442100001A000000FD00304C575716010A202020202020000000FC0042323436574C0A202020202020000000FF0054384E4545303033383532320A01F802031CF14F90020304050607011112131415161F2309070783010000011D8018711C1620582C250006442100009E011D007251D01E206E28550006442100001E8C0AD08A20E02D10103E9600064421000018C344806E70B028401720A80406442100001E00000000000000000000000000000000000000000000000000000096",
    label: m.video_edid_acer_b246wl(),
  },
  {
    value:
      "00FFFFFFFFFFFF0006B3872401010101021F010380342078EA6DB5A7564EA0250D5054BF6F00714F8180814081C0A9409500B300D1C0283C80A070B023403020360006442100001A000000FD00314B1E5F19000A202020202020000000FC00504132343851560A2020202020000000FF004D314C4D51533035323135370A014D02032AF14B900504030201111213141F230907078301000065030C001000681A00000101314BE6E2006A023A801871382D40582C450006442100001ECD5F80B072B0374088D0360006442100001C011D007251D01E206E28550006442100001E8C0AD08A20E02D10103E960006442100001800000000000000000000000000DC",
    label: m.video_edid_asus_pa248qv(),
  },
  {
    value:
      "00FFFFFFFFFFFF0010AC132045393639201E0103803C22782ACD25A3574B9F270D5054A54B00714F8180A9C0D1C00101010101010101023A801871382D40582C450056502100001E000000FF00335335475132330A2020202020000000FC0044454C4C204432373231480A20000000FD00384C1E5311000A202020202020018102031AB14F90050403020716010611121513141F65030C001000023A801871382D40582C450056502100001E011D8018711C1620582C250056502100009E011D007251D01E206E28550056502100001E8C0AD08A20E02D10103E960056502100001800000000000000000000000000000000000000000000000000000000004F",
    label: m.video_edid_dell_d2721h(),
  },
  {
    value:
      "00ffffffffffff0010ac0100020000000111010380221bff0a00000000000000000000adce0781800101010101010101010101010101000000ff0030303030303030303030303030000000ff0030303030303030303030303030000000fd00384c1f530b000a000000000000000000fc0044454c4c2049445241430a2020000a",
    label: m.video_edid_dell_idrac(),
  },
];

export default function SettingsVideoRoute() {
  const { send } = useJsonRpc();
  const h265Supported = supportsH265WebRTC();
  const [streamQuality, setStreamQuality] = useState(5000);
  const [videoCodec, setVideoCodec] = useState<number>(0);
  const [customEdidValue, setCustomEdidValue] = useState<string | null>(null);
  const [edid, setEdid] = useState<string | null>(null);
  const [edidLoading, setEdidLoading] = useState(true);
  const { debugMode } = useSettingsStore();
  // Video enhancement settings from store
  const {
    videoSaturation,
    setVideoSaturation,
    videoBrightness,
    setVideoBrightness,
    videoContrast,
    setVideoContrast,
    videoSharpness,
    setVideoSharpness,
  } = useSettingsStore();

  useEffect(() => {
    send("getStreamQualityFactor", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setStreamQuality(Number(resp.result));
    });

    send("getVideoCodec", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) return;
      setVideoCodec(Number(resp.result));
    });

    send("getEDID", {}, (resp: JsonRpcResponse) => {
      setEdidLoading(false);
      if ("error" in resp) {
        notifications.error(
          m.video_failed_get_edid({ error: resp.error.data || m.unknown_error() }),
        );
        return;
      }

      const receivedEdid = resp.result as string;

      const matchingEdid = edids.find(x => x.value.toLowerCase() === receivedEdid.toLowerCase());

      if (matchingEdid) {
        setEdid(matchingEdid.value);
        // Reset custom EDID value
        setCustomEdidValue(null);
      } else {
        setEdid("custom");
        setCustomEdidValue(receivedEdid);
      }
    });
  }, [send]);

  const handleStreamQualityChange = (bitrate: number) => {
    send("setStreamQualityFactor", { factor: bitrate }, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        notifications.error(
          m.video_failed_set_stream_quality({ error: resp.error.data || m.unknown_error() }),
        );
        return;
      }

      notifications.success(
        m.video_stream_quality_set({
          quality: `${bitrate} kbps`,
        }),
      );
      setStreamQuality(bitrate);
    });
  };

  const handleVideoCodecChange = (codec: number) => {
    if (codec === 1 && !h265Supported) {
      notifications.error(m.video_failed_set_codec({ error: m.video_h265_safari_only_error() }));
      setVideoCodec(0);
      return;
    }

    send("setVideoCodec", { codec }, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        notifications.error(
          m.video_failed_set_codec({ error: resp.error.data || m.unknown_error() }),
        );
        return;
      }

      notifications.success(m.video_codec_set_success({ codec: codec === 1 ? "H.265" : "H.264" }));
      setVideoCodec(codec);

      setTimeout(() => {
        window.location.reload();
      }, 1000); // 1s delay
    });
  };

  const handleEDIDChange = (newEdid: string) => {
    setEdidLoading(true);
    send("setEDID", { edid: newEdid }, (resp: JsonRpcResponse) => {
      setEdidLoading(false);
      if ("error" in resp) {
        notifications.error(
          m.video_failed_set_edid({ error: resp.error.data || m.unknown_error() }),
        );
        return;
      }

      notifications.success(
        m.video_edid_set_success({
          edid: edids.find(x => x.value === newEdid)?.label ?? m.video_custom_edid_fallback(),
        }),
      );
      // Update the EDID value in the UI
      setEdid(newEdid);
    });
  };

  const [debugInfo, setDebugInfo] = useState<string | null>(null);
  const [debugInfoLoading, setDebugInfoLoading] = useState(false);
  const getDebugInfo = useCallback(() => {
    setDebugInfoLoading(true);
    send("getVideoLogStatus", {}, (resp: JsonRpcResponse) => {
      if ("error" in resp) {
        notifications.error(
          m.video_failed_get_debug_info({ error: resp.error.data || m.unknown_error() }),
        );
        setDebugInfoLoading(false);
        return;
      }
      const data = resp.result as string;
      setDebugInfo(
        data
          .split("\n")
          .map(line => line.trim().replace(/^\[\s*\d+\.\d+\]\s*/, ""))
          .join("\n"),
      );
      setDebugInfoLoading(false);
    });
  }, [send]);

  return (
    <div className="space-y-3">
      <div className="space-y-4">
        <SettingsPageHeader title={m.video_title()} description={m.video_description()} />

        <div className="space-y-4">
          <div className="space-y-4">
            <SettingsItem
              title={m.video_stream_quality_title()}
              description={m.video_stream_quality_description()}
            >
              <div className="flex items-center gap-3">
                <input
                  type="range"
                  min="1000"
                  max="20000"
                  step="100"
                  value={streamQuality}
                  onChange={e => setStreamQuality(Number(e.target.value))}
                  onMouseUp={e =>
                    handleStreamQualityChange(Number((e.target as HTMLInputElement).value))
                  }
                  onTouchEnd={e =>
                    handleStreamQualityChange(Number((e.target as HTMLInputElement).value))
                  }
                  className="h-2 w-48 cursor-pointer appearance-none rounded-lg bg-gray-200 dark:bg-gray-700"
                />
                <span className="min-w-[90px] text-sm font-medium">{streamQuality} kbps</span>
              </div>
            </SettingsItem>

            <SettingsItem title={m.video_codec_title()} description={m.video_codec_description()}>
              <SelectMenuBasic
                size="SM"
                label=""
                value={videoCodec.toString()}
                onChange={e => handleVideoCodecChange(Number(e.target.value))}
                options={[
                  { value: "0", label: m.video_codec_h264_avc() },
                  {
                    value: "1",
                    label: h265Supported
                      ? m.video_codec_h265_hevc()
                      : m.video_codec_h265_hevc_safari_only(),
                    disabled: !h265Supported,
                  },
                ]}
              />
            </SettingsItem>

            {/* Video Enhancement Settings */}
            <SettingsItem
              title={m.video_enhancement_title()}
              description={m.video_enhancement_description()}
            />

            <NestedSettingsGroup>
              <SettingsItem
                title={m.video_saturation_title()}
                description={m.video_saturation_description({ value: videoSaturation.toFixed(1) })}
              >
                <input
                  type="range"
                  min="0.5"
                  max="2.0"
                  step="0.1"
                  value={videoSaturation}
                  onChange={e => setVideoSaturation(Number.parseFloat(e.target.value))}
                  className="h-2 w-32 cursor-pointer appearance-none rounded-lg bg-gray-200 dark:bg-gray-700"
                />
              </SettingsItem>

              <SettingsItem
                title={m.video_brightness_title()}
                description={m.video_brightness_description({ value: videoBrightness.toFixed(1) })}
              >
                <input
                  type="range"
                  min="0.5"
                  max="1.5"
                  step="0.1"
                  value={videoBrightness}
                  onChange={e => setVideoBrightness(Number.parseFloat(e.target.value))}
                  className="h-2 w-32 cursor-pointer appearance-none rounded-lg bg-gray-200 dark:bg-gray-700"
                />
              </SettingsItem>

              <SettingsItem
                title={m.video_contrast_title()}
                description={m.video_contrast_description({ value: videoContrast.toFixed(1) })}
              >
                <input
                  type="range"
                  min="0.5"
                  max="2.0"
                  step="0.1"
                  value={videoContrast}
                  onChange={e => setVideoContrast(Number.parseFloat(e.target.value))}
                  className="h-2 w-32 cursor-pointer appearance-none rounded-lg bg-gray-200 dark:bg-gray-700"
                />
              </SettingsItem>

              <SettingsItem
                title={m.video_sharpness_title()}
                description={m.video_sharpness_description({ value: videoSharpness.toFixed(2) })}
              >
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.05"
                  value={videoSharpness}
                  onChange={e => setVideoSharpness(Number.parseFloat(e.target.value))}
                  className="h-2 w-32 cursor-pointer appearance-none rounded-lg bg-gray-200 dark:bg-gray-700"
                />
              </SettingsItem>

              <div className="flex gap-2">
                <Button
                  size="SM"
                  theme="light"
                  text={m.video_reset_to_default()}
                  onClick={() => {
                    setVideoSaturation(1);
                    setVideoBrightness(1);
                    setVideoContrast(1);
                    setVideoSharpness(0);
                  }}
                />
              </div>
            </NestedSettingsGroup>
            <Fieldset disabled={edidLoading} className="space-y-2">
              <SettingsItem
                title={m.video_edid_title()}
                description={m.video_edid_description()}
                loading={edidLoading}
              >
                <SelectMenuBasic
                  size="SM"
                  label=""
                  fullWidth
                  value={customEdidValue ? "custom" : edid || ""}
                  onChange={e => {
                    if (e.target.value === "custom") {
                      setEdid("custom");
                      setCustomEdidValue("");
                    } else {
                      setCustomEdidValue(null);
                      handleEDIDChange(e.target.value);
                    }
                  }}
                  options={[...edids, { value: "custom", label: m.video_edid_custom() }]}
                />
              </SettingsItem>
              {customEdidValue !== null && (
                <>
                  <SettingsItem
                    title={m.video_custom_edid_title()}
                    description={m.video_custom_edid_description()}
                  />
                  <TextAreaWithLabel
                    label={m.video_edid_file_label()}
                    placeholder="00F..."
                    rows={3}
                    value={customEdidValue}
                    onChange={e => setCustomEdidValue(e.target.value)}
                  />
                  <div className="flex justify-start gap-x-2">
                    <Button
                      size="SM"
                      theme="primary"
                      text={m.video_set_custom_edid()}
                      loading={edidLoading}
                      onClick={() => handleEDIDChange(customEdidValue)}
                    />
                    <Button
                      size="SM"
                      theme="light"
                      text={m.video_restore_to_default()}
                      loading={edidLoading}
                      onClick={() => {
                        setCustomEdidValue(null);
                        handleEDIDChange(defaultEdid);
                      }}
                    />
                  </div>
                </>
              )}
            </Fieldset>
          </div>

          {debugMode && (
            <div className="space-y-4">
              <SettingsItem
                title={m.video_debugging_info_title()}
                description={m.video_debugging_info_description()}
              >
                <Button
                  size="SM"
                  theme="primary"
                  text={m.video_get_debugging_info()}
                  loading={debugInfoLoading}
                  disabled={debugInfoLoading}
                  onClick={() => {
                    getDebugInfo();
                  }}
                />
              </SettingsItem>
              {debugInfo && (
                <div className="max-h-64 overflow-y-auto rounded-md bg-gray-100 p-2 font-mono text-xs dark:bg-gray-800">
                  <pre className="whitespace-pre-wrap">{debugInfo}</pre>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
