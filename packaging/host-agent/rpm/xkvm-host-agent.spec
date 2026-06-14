%{!?xkvm_version:%{error:xkvm_version must be defined by the package build script}}

Name:           xkvm-host-agent
Version:        %{xkvm_version}
Release:        1%{?dist}
Summary:        Host metric sender for xKVM

License:        GPL-3.0-only
URL:            https://github.com/134ARG/xkvm
Source0:        %{name}-%{version}.tar.gz

BuildArch:      noarch
Requires:       python3
Requires:       python3-psutil
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd
BuildRequires:  systemd-rpm-macros

%description
xkvm-host-agent collects host metrics and streams newline-delimited JSON to
the xKVM host metrics listener.

%prep
%autosetup

%build
# Python script; nothing to build.

%install
install -D -m 0755 agents/host/xkvm-host-agent.py %{buildroot}%{_bindir}/xkvm-host-agent
install -D -m 0644 packaging/host-agent/systemd/xkvm-host-agent.service %{buildroot}%{_unitdir}/xkvm-host-agent.service
install -D -m 0644 packaging/host-agent/xkvm-host-agent.conf %{buildroot}%{_sysconfdir}/xkvm-host-agent/agent.conf

%post
%systemd_post xkvm-host-agent.service

%preun
%systemd_preun xkvm-host-agent.service

%postun
%systemd_postun_with_restart xkvm-host-agent.service

%files
%doc agents/host/README.md
%{_bindir}/xkvm-host-agent
%{_unitdir}/xkvm-host-agent.service
%dir %{_sysconfdir}/xkvm-host-agent
%config(noreplace) %{_sysconfdir}/xkvm-host-agent/agent.conf
