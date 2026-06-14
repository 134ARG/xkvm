# xKVM Host Agent

Host-side metric collector for xKVM.

The agent runs on the controlled machine, collects CPU, memory, GPU, temperature,
network, uptime, and failed-systemd-unit metrics, then streams newline-delimited
JSON to the xKVM host metrics listener.

## Run Manually

```bash
python3 agents/host/xkvm-host-agent.py --xkvm-addr <xkvm-ip> --xkvm-port 9101
```

## RPM

```bash
# Normal package build; also builds the agent RPM.
make build_packages

# Optional agent-only build.
make build_host_agent_rpm
```

The RPM installs:

- `/usr/bin/xkvm-host-agent`
- `/etc/xkvm-host-agent/agent.conf`
- `/usr/lib/systemd/system/xkvm-host-agent.service`

Edit `/etc/xkvm-host-agent/agent.conf`, then enable the service:

```bash
sudo systemctl enable --now xkvm-host-agent
```
