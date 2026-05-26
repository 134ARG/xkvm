# xkVM VFD Agent

Host-side metric collector for the xkVM VFD display.

The agent runs on the controlled machine, collects CPU, memory, GPU, temperature,
network, uptime, and failed-systemd-unit metrics, then streams newline-delimited
JSON to the xkVM VFD listener.

## Run Manually

```bash
python3 agents/vfd/xkvm-vfd-agent.py --host <xkvm-ip> --port 9101
```

## RPM

```bash
# Normal package build; also builds the agent RPM.
make build_packages

# Optional agent-only build.
make build_vfd_agent_rpm
```

The RPM installs:

- `/usr/bin/xkvm-vfd-agent`
- `/etc/xkvm-vfd-agent/agent.conf`
- `/usr/lib/systemd/system/xkvm-vfd-agent.service`

Edit `/etc/xkvm-vfd-agent/agent.conf`, then enable the service:

```bash
sudo systemctl enable --now xkvm-vfd-agent
```
