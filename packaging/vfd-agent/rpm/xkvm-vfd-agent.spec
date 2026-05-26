%{!?xkvm_version:%global xkvm_version 0.1.5}

Name:           xkvm-vfd-agent
Version:        %{xkvm_version}
Release:        1%{?dist}
Summary:        Host metric sender for the xkVM VFD display

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
xkvm-vfd-agent collects host metrics and streams newline-delimited JSON to
the xkVM VFD listener.

%prep
%autosetup

%build
# Python script; nothing to build.

%install
install -D -m 0755 agents/vfd/xkvm-vfd-agent.py %{buildroot}%{_bindir}/xkvm-vfd-agent
install -D -m 0644 packaging/vfd-agent/systemd/xkvm-vfd-agent.service %{buildroot}%{_unitdir}/xkvm-vfd-agent.service
install -D -m 0644 packaging/vfd-agent/xkvm-vfd-agent.conf %{buildroot}%{_sysconfdir}/xkvm-vfd-agent/agent.conf

%post
%systemd_post xkvm-vfd-agent.service

%preun
%systemd_preun xkvm-vfd-agent.service

%postun
%systemd_postun_with_restart xkvm-vfd-agent.service

%files
%doc agents/vfd/README.md
%{_bindir}/xkvm-vfd-agent
%{_unitdir}/xkvm-vfd-agent.service
%dir %{_sysconfdir}/xkvm-vfd-agent
%config(noreplace) %{_sysconfdir}/xkvm-vfd-agent/agent.conf

%changelog
* Tue May 26 2026 134ARG <xen134@outlook.com> - 0.1.5-1
- Add xkVM VFD metric agent package.
