# XKVM

[English](README.en.md) | [简体中文](README.md)

XKVM 是一个高性能、开源、100% 本地的 KVM-over-IP 解决方案，运行在完整的 Linux 系统上。它提供远程键鼠输入、1080p 60fps的硬件编码视频流，用于远程管理机器。

这玩意目前只适配我自己搓的Radxa Zero 3 (rk3566) + tc358743 HDMI-CSI采集卡的KVM平台。采集卡硬件的立创开源地址[在此](https://oshwhub.com/gth38/hdmi-csi)。可以比较简单地移植到其他rk平台。

## 概述

XKVM基于JetKVM，删掉依大托我用不到的功能，比如云端连接，网络配置，设备管理等等（毕竟跑在一个完整的linux而非buildroot上）。添加了2M - 20M的码率调整以及H.265编码，可自定义GPIO的ATX电源控制与监控。同时还基于Tauri增加了Linux，Windows和MacOS的前端本地应用，用来解决使用tailscale等VPN进行跨LAN连接时恼人的浏览器WebRTC安全限制。

项目主要开发目的为个人自用，随时弃坑。其他细节见英文版readme。

## 发布包

XKVM 当前发布包分为三类：

| 包 | 安装位置 | 作用 |
|---|---|---|
| `xkvm_<version>_arm64.deb` | XKVM 设备 | 主 XKVM 服务包。包含 KVM 服务、WebUI、视频采集/编码、USB HID 控制、虚拟媒体、GPIO 电源控制等设备端功能。 |
| `xkvm-vfd-agent-<version>-*.rpm` | 被控主机，仅在使用 VFD 显示功能时需要 | 可选的主机端指标上报工具。它会把 CPU、内存、GPU、温度、网络、运行时间、失败的 systemd 单元状态等信息发送给 XKVM 的 VFD 监听端。 |
| `XKVM-Connector_<version>_aarch64.dmg` | Apple Silicon macOS 客户端 | 可选的 macOS 原生前端。适合想用桌面应用，或需要绕开 VPN/跨 LAN 场景下浏览器 WebRTC 限制时使用。 |

大多数用户只需要在 XKVM 设备上安装 Debian 包。VFD agent 和 macOS connector 都是针对特定场景的可选配套包。

## 配置文件路径

XKVM 遵循标准 Linux FHS 规范，路径可通过环境变量覆盖。

| 路径 | 环境变量 | 用途 |
|---|---|---|
| `/etc/xkvm/` | `XKVM_CONFIG_DIR` | 配置文件 |
| `/var/lib/xkvm/` | `XKVM_DATA_DIR` | 数据文件 |
| `/var/log/xkvm/` | `XKVM_LOG_DIR` | 日志 |

主要文件：

| 文件 | 说明 |
|---|---|
| `/etc/xkvm/kvm_config.json` | 主配置（USB、视频、认证、宏等） |
| `/etc/xkvm/tls/` | TLS 证书 |
| `/etc/xkvm/.native-debug-mode` | 创建此文件启用 native 调试模式 |
| `/var/lib/xkvm/images/` | 虚拟介质镜像（ISO/磁盘） |
| `/var/lib/xkvm/crashdump/` | 崩溃日志 |
| `/var/log/xkvm/last.log` | 应用标准输出/错误日志 |

通过 systemd 运行时，目录由 `ConfigurationDirectory`、`StateDirectory`、`LogsDirectory` 自动创建。

## `/etc/xkvm/kvm_config.json` 选项

缺失字段会使用内置默认值。标记为废弃的字段仅为兼容旧配置保留。大多数选项都可以在 WebUI 的 Settings 页面配置；只有在 WebUI 配置被弄乱时，才需要手动编辑这里。

| 键 | 简介 | 取值 |
|---|---|---|
| `cloud_url` | 已废弃的云/API 基础地址 | 已废弃；保持 `""` |
| `public_ipv4_endpoint`<br>`public_ipv6_endpoint` | 公网 IP 查询端点 | 返回纯文本 IP 的 URL/域名；<br>`""` 表示禁用该地址族 |
| `localAuthMode` | 本地认证模式 | `""`, `"password"`, `"noPassword"` |
| `hashed_password`<br>`local_auth_token` | 本地认证凭据 | 字符串，由本地认证逻辑维护 |
| `local_loopback_only` | 本地访问仅绑定 loopback | 布尔值 |
| `tls_mode` | HTTPS 证书模式 | `""`, `"self-signed"`, `"user-defined"` |
| `default_log_level` | 默认日志级别 | `DISABLE`, `NOLEVEL`, `PANIC`, `FATAL`,<br>`ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE` |
| `auto_update_enabled`<br>`include_pre_release` | 已废弃的 OTA 更新行为 | 已废弃/已禁用；布尔值，仅兼容旧配置 |
| `video_quality_factor` | 视频码率 | kbps；UI 通常使用 `2000`-`20000` |
| `video_codec` | 视频编码器 | `0` = H.264，`1` = H.265 |
| `video_sleep_after_sec` | HDMI 睡眠超时 | `0` = 默认 60 秒，正数为秒数，负数禁用 |
| `native_max_restart_attempts` | native helper 重启上限 | 无符号整数 |
| `keyboard_layout` | 键盘映射 | `cs-CZ`, `da-DK`, `de-CH`, `de-DE`, `en-UK`,<br>`en-US`, `es-ES`, `nl-BE`, `fr-CH`, `fr-FR`,<br>`it-IT`, `ja-JP`, `nb-NO`, `sv-SE` |
| `keyboard_macros` | 保存的键盘宏 | `{id,name,sortOrder,steps}` 数组；<br>最多 25 个宏、每个宏 10 步、每步 10 个按键 |
| `hdmi_edid_string` | 保存的 HDMI EDID | UI 保存的 EDID hex/base64 字符串 |
| `active_extension` | 当前加载的控制扩展 | `""`, `"atx-power"`, `"dc-power"`,<br>`"serial-console"` |
| `jiggler_enabled` | 启用鼠标 jiggler | 布尔值 |
| `jiggler_config` | 鼠标 jiggler 调度/限制 | `{inactivity_limit_seconds,jitter_percentage,`<br>`schedule_cron_tab,timezone}`；<br>时区为 IANA/`UTC` |
| `usb_config` | USB gadget 身份信息 | `{vendor_id,product_id,serial_number,`<br>`manufacturer,product}`；<br>ID 为 `0x1d6b` 这类十六进制字符串 |
| `usb_devices` | USB gadget 功能开关 | `{absolute_mouse,relative_mouse,`<br>`keyboard,mass_storage}` 布尔值 |
| `network_config` | 已废弃的网络设置 | 加载时忽略，保存时不会写回 |
| `gpio_pwr_chip/line/active_high`<br>`gpio_rst_chip/line/active_high`<br>`gpio_pwr_led_chip/line/active_high`<br>`gpio_hdd_led_chip/line/active_high` | GPIO ATX/LED 映射 | `chip` 字符串、`line` 整数（`-1` 禁用）、<br>`active_high` 布尔值 |
| `sensor_env_chip`<br>`sensor_soc_chip` | hwmon 芯片选择 | hwmon 芯片名 |
| `sensor_temp_feature`<br>`sensor_hum_feature`<br>`sensor_soc_feature` | hwmon feature 选择 | hwmon feature 名 |
| `vfd_enabled`<br>`vfd_device_path`<br>`vfd_listen_port` | VFD 显示支持 | 布尔值、设备路径字符串、TCP 端口 |
| `serial_port_path` | 串口控制台设备 | 串口设备路径；`""` 禁用串口控制台 |
| `wake_on_lan_devices` | 已废弃的 Wake-on-LAN 列表 | 已废弃的 `{name,macAddress}` 数组 |

## 演示

以下截图来自 Tauri 原生桌面界面（macOS）:

视频设置：支持 2M–20M 码率调节，H.264/H.265 编码切换（H.265 仅限 macOS Safari）。

<img src="images/video-demo.png" />

硬件设置：USB 设备类型、GPIO ATX 控制、机箱传感器与 VFD 显示配置。

<img src="images/gpio-sensor-vfd-settings.png" />

传感器与主机指标：主界面可查看设备环境传感器、主机负载、温度、网络流量与运行状态。

<img src="images/sensors-demo.png" />

## 贡献

欢迎fork。

## 许可证

详情请参阅 [LICENSE](LICENSE)。
