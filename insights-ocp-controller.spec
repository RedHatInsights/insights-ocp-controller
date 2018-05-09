%global project       RedHatInsights
%global repo          insights-ocp-controller
%global commit        v0.0.1
%global shortcommit   %(c=%{commit}; echo ${c:0:7})

Name:           insights-ocp-controller
Version:        0.0.1
Release:        2%{?dist}
Summary:        Tool for extracting and serving content of container images
License:        ASL 2.0
URL:            https://github.com/redhatinsights/insights-ocp-controller
Source0:        https://github.com/%{project}/%{repo}/archive/%{commit}/%{repo}-%{version}.tar.gz
Source1:        client-egg.tar.gz
BuildRequires:  golang >= 1.7
Requires:       insights-client >= 3.0.3

%description
Insights scan controller for Openshift Container Platform.

%prep
%setup -qn %{name}-%{version}
%setup -T -D -a 1

%build
mkdir -p ./_build/src/github.com/RedHatInsights
ln -s $(pwd) ./_build/src/github.com/RedHatInsights/insights-ocp-controller
export GOPATH=$(pwd)/_build:$(pwd)/Godeps/_workspace:%{gopath}
go build -o insights-ocp-controller $(pwd)/_build/src/github.com/RedHatInsights/insights-ocp-controller/cmd.go

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 ./insights-ocp-controller %{buildroot}%{_bindir}/insights-ocp-controller
install -m644 ./client-egg/rpm.egg  %{buildroot}/etc/insights-ocp-controller
install -m644 ./client-egg/rpm.egg.asc  %{buildroot}/etc/insights-ocp-controller

%files
#%doc LICENSE README.md
%{_bindir}/insights-ocp-controller
/etc/insights-ocp-controller/rpm.egg
/etc/insights-ocp-controller/rpm.egg.asc

%changelog
* Tue May 08 2018 Lindani Phiri <lphiri@redhat.com> - 0.0.1-2
- Address RPM diff issues 
- Scan in controller

* Wed May 02 2018 Lindani Phiri <lphiri@redhat.com> - 0.0.1-1
- Initial Release

* Wed Apr 25 2018 Lindani Phiri <lphiri@redhat.com> - 0.0.1-0.alpha1
- Initial Build (Alpha)
