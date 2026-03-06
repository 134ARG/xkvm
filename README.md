# XKVM

[English](README.en.md) | [简体中文](README.md)

XKVM 是一个高性能、开源、100% 本地的 KVM-over-IP 解决方案，运行在完整的 Linux 系统上。它提供远程键鼠输入、1080p 60fps的硬件编码视频流，用于远程管理机器。

这玩意目前只适配我自己搓的Radxa Zero 3 (rk3566) + tc358743 HDMI-CSI采集卡的KVM平台。采集卡硬件的立创开源地址[在此](https://oshwhub.com/gth38/hdmi-csi)。可以比较简单地移植到其他rk平台。

## 概述

XKVM基于JetKVM，删掉依大托我用不到的功能，比如云端连接，网络配置，设备管理等等（毕竟跑在一个完整的linux而非buildroot上）。添加了2M - 20M的码率调整以及H.265编码。同时还基于Tauri增加了Linux，Windows和MacOS的前端本地应用，用来解决使用tailscale等VPN进行跨LAN连接时恼人的浏览器WebRTC安全限制。

项目主要开发目的为个人自用，随时弃坑。其他细节见英文版readme。

## 贡献

欢迎fork。

## 许可证

详情请参阅 [LICENSE](LICENSE)。
