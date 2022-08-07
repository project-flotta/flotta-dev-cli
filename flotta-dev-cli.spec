Name:       flotta-dev-cli
Version:    0.2.0
Release:    4%{?dist}
Summary:    CLI for flotta development
ExclusiveArch: %{go_arches}
Group:      Flotta
License:    ASL 2.0
Source0:    %{name}-%{version}.tar.gz

BuildRequires:  golang
BuildRequires:  bash

Provides:       %{name} = %{version}-%{release}
Provides:       golang(%{go_import_path}) = %{version}-%{release}

%description
The flotta-dev-cli provides a tool to manage devices and predefined workloads for flotta. It emulates a device by running it as a docker container.
The device will be registered to flotta Edge API service as deployed on a local k8s cluster.

%prep
tar fx %{SOURCE0}

%build
cd flotta-dev-cli-%{VERSION}
export GOFLAGS="-mod=vendor"
go build -o ./bin/flotta ./main.go

%install
cd flotta-dev-cli-%{VERSION}
mkdir -p %{buildroot}%{_bindir}/
install -m 755 ./bin/flotta %{buildroot}%{_bindir}/flotta

%files
%{_bindir}/flotta

%changelog
* Sun Aug 7 2022 Moti Asayag <masayag@redhat.com> 0.2.0-4
- Pin edgedevice image to a stable version

* Wed Jul 27 2022 Moti Asayag <masayag@redhat.com> 0.2.0-3
- Pull edgedevice image if not available

* Wed  Jul 27 2022 Moti Asayag <masayag@redhat.com> 0.2.0-2
- Replaced nameless args with named flags in cli commands.
- Allow adding workloads only to existing devices.

* Thu Jul 21 2022 Moti Asayag <masayag@redhat.com> 0.2.0-1
- Initial release.
- Added support for managing edge devices and their registration to local k8s cluster.
- Added support for creating predefined workloads, assigning them to specific devices and removing them.
