%define name %{getenv:NAME}
%define version %{getenv:VERSION}
%define release %{getenv:RELEASE}
%define cdate %{getenv:CDATE}
Name: %{name}
Version: %{version}
Release: %{release}
Summary: gowhoson
License: MIT
Group: Networking/Daemons
Source0: %{name}
Source1: %{name}.init
Source2: %{name}.json
URL: https://github.com/tai-ga/gowhoson
BuildRoot: /var/tmp/%{name}-build
BuildRequires: initscripts
Requires: initscripts

%description
gowhoson is a golang implementation of the "Whoson" protocol.

%prep

%build

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/sbin
mkdir -p $RPM_BUILD_ROOT/etc/sysconfig
mkdir -p $RPM_BUILD_ROOT/etc/logrotate.d
mkdir -p $RPM_BUILD_ROOT/etc/rc.d/{init.d,rc0.d,rc1.d,rc2.d,rc3.d,rc4.d,rc5.d,rc6.d}
install -m755 $RPM_SOURCE_DIR/%{name} $RPM_BUILD_ROOT/usr/sbin
install -m755 $RPM_SOURCE_DIR/%{name}.init $RPM_BUILD_ROOT/etc/rc.d/init.d/%{name}
install -m644 $RPM_SOURCE_DIR/%{name}.json $RPM_BUILD_ROOT/etc/%{name}.json
install -m644 $RPM_SOURCE_DIR/%{name}.sysconfig $RPM_BUILD_ROOT/etc/sysconfig/%{name}
install -m644 $RPM_SOURCE_DIR/%{name}.logrotate $RPM_BUILD_ROOT/etc/logrotate.d/%{name}

ln -sf ../init.d/%{name} $RPM_BUILD_ROOT/etc/rc.d/rc0.d/K30%{name}
ln -sf ../init.d/%{name} $RPM_BUILD_ROOT/etc/rc.d/rc1.d/K30%{name}
ln -sf ../init.d/%{name} $RPM_BUILD_ROOT/etc/rc.d/rc2.d/S80%{name}
ln -sf ../init.d/%{name} $RPM_BUILD_ROOT/etc/rc.d/rc3.d/S80%{name}
ln -sf ../init.d/%{name} $RPM_BUILD_ROOT/etc/rc.d/rc4.d/S80%{name}
ln -sf ../init.d/%{name} $RPM_BUILD_ROOT/etc/rc.d/rc5.d/S80%{name}
ln -sf ../init.d/%{name} $RPM_BUILD_ROOT/etc/rc.d/rc6.d/K30%{name}

%pre
if ! grep -q "^gowhoson:" /etc/group; then
    groupadd gowhoson
fi
if ! grep -q "^gowhoson:" /etc/passwd; then
    useradd -g gowhoson gowhoson -d /var/empty
fi

%post

%preun
if [ "$1" = 0 ]; then
  /sbin/chkconfig --del %{name}
fi

%files
%defattr(-,root,root)
%attr(4755,root,root)
/usr/sbin/%{name}

%attr(644,root,root)
%config /etc/%{name}.json
%config /etc/sysconfig/%{name}
%config /etc/logrotate.d/%{name}
%config /etc/rc.d/init.d/%{name}
%config(missingok) /etc/rc.d/rc0.d/K30%{name}
%config(missingok) /etc/rc.d/rc1.d/K30%{name}
%config(missingok) /etc/rc.d/rc2.d/S80%{name}
%config(missingok) /etc/rc.d/rc3.d/S80%{name}
%config(missingok) /etc/rc.d/rc4.d/S80%{name}
%config(missingok) /etc/rc.d/rc5.d/S80%{name}
%config(missingok) /etc/rc.d/rc6.d/K30%{name}

%changelog
* Mon Sep 25 2017 Masahiro Ono <masahiro.o@gmail.com>
- Add command dump mode
- Change grpc loging module to zap
- Rename option GRPCPort to ControlPort
- Add command option savefile for server mode
- Fix permission for log directory

* Mon Sep 14 2017 Masahiro Ono <masahiro.o@gmail.com> gowhoson-v0.1.3-1
- Add /etc/sysconfig/gowhoson, /etc/logrotate.d/gowhoson

* Mon Sep 13 2017 Masahiro Ono <masahiro.o@gmail.com> gowhoson-v0.1.2-1
- Update rpm.spec %pre section

* Mon Sep 11 2017 Masahiro Ono <masahiro.o@gmail.com> gowhoson-v0.1.1-1
- Initial build
