<div align="center">
    <img alt="XKVM logo" src="https://xkvm.com/logo-blue.png" height="28">

### KVM

[Discord](https://xkvm.com/discord) | [Website](https://xkvm.com) | [Issues](https://github.com/xkvm/cloud-api/issues) | [Docs](https://xkvm.com/docs)

[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/xkvm.svg?style=social&label=Follow%20%40XKVM)](https://twitter.com/xkvm)

[![Go Report Card](https://goreportcard.com/badge/github.com/xkvm/kvm)](https://goreportcard.com/report/github.com/xkvm/kvm)

</div>

XKVM is a high-performance, open-source KVM over IP (Keyboard, Video, Mouse) solution designed for efficient remote management of computers, servers, and workstations. Whether you're dealing with boot failures, installing a new operating system, adjusting BIOS settings, or simply taking control of a machine from afar, XKVM provides the tools to get it done effectively.

## Features

- **Ultra-low Latency** - 1080p@60FPS video with 30-60ms latency using H.264 encoding. Smooth mouse and keyboard interaction for responsive remote control.
- **Free & Optional Remote Access** - Remote management via XKVM Cloud using WebRTC.
- **Open-source software** - Written in Golang on Linux. Easily customizable through SSH access to the XKVM device.

## Contributing

We welcome contributions from the community! Whether it's improving the firmware, adding new features, or enhancing documentation, your input is valuable. We also have some rules and taboos here, so please read this page and our [Code of Conduct](/CODE_OF_CONDUCT.md) carefully.

## I need help

The best place to search for answers is our [Documentation](https://xkvm.com/docs). If you can't find the answer there, check our [Discord Server](https://xkvm.com/discord).

## I want to report an issue

If you've found an issue and want to report it, please check our [Issues](https://github.com/xkvm/kvm/issues) page. Make sure the description contains information about the firmware version you're using, your platform, and a clear explanation of the steps to reproduce the issue.

# Development

XKVM is written in Go & TypeScript. with some bits and pieces written in C. An intermediate level of Go & TypeScript knowledge is recommended for comfortable programming.

The project contains two main parts, the backend software that runs on the KVM device and the frontend software that is served by the KVM device, and also the cloud.

For comprehensive development information, including setup, testing, debugging, and contribution guidelines, see **[DEVELOPMENT.md](DEVELOPMENT.md)**.

For quick device development, use the `./dev_deploy.sh` script. It will build the frontend and backend and deploy them to the local KVM device. Run `./dev_deploy.sh --help` for more information.

## Backend

The backend is written in Go and is responsible for the KVM device management, the cloud API and the cloud web.

## Frontend

The frontend is written in React and TypeScript and is served by the KVM device. It has three build targets: `device`, `development` and `production`. Development is used for development of the cloud version on your local machine, device is used for building the frontend for the KVM device and production is used for building the frontend for the cloud.
