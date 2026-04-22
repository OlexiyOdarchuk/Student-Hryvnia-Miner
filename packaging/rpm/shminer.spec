Name:           shminer
Version:        @VERSION@
Release:        1%{?dist}
Summary:        Mining Client for S-UAH
License:       GPL-3.0-only
URL:            @URL@
BuildArch:     x86_64

%description
Mining Client for S-UAH cryptocurrency

%prep
mkdir -p %{name}-%{version}

%install
mkdir -p %{buildroot}/%{_bindir}
mkdir -p %{buildroot}/%{_datadir}/icons/hicolor/256x256/apps
mkdir -p %{buildroot}/%{_datadir}/applications
install -Dm755 %{name}-%{version}/shminer %{buildroot}/%{_bindir}/shminer
install -Dm644 %{name}-%{version}/shminer.png %{buildroot}/%{_datadir}/icons/hicolor/256x256/apps/shminer.png
install -Dm644 %{name}-%{version}/shminer.desktop %{buildroot}/%{_datadir}/applications/shminer.desktop

%files
%{_bindir}/shminer
%{_datadir}/icons/hicolor/256x256/apps/shminer.png
%{_datadir}/applications/shminer.desktop

%license LICENSE

%changelog
* Thu Apr 22 2026 iShawyha <shawyhaf@gmail.com> - @VERSION@-1
- Initial RPM package