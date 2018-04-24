%global project       redhatinsights
%global repo          insights-ocp-controller
%global commit        0.0.1

Name:           insights-ocp-controller
Version:        0.0.1
Release:        0%{?dist}
Summary:        Tool for extracting and serving content of container images
License:        ASL 2.0
URL:            https://github.com/RedHatInsights/insights-ocp-controller
Source0:        ./insights-ocp-controller-0.0.1.tar.gz
BuildRequires:  golang >= 1.7

%description
Insights scan controller for Openshift Container Platform.

%prep
%setup -qn %{name}-%{version}

%build
mkdir -p ./_build/src/github.com/RedHatInsights
ln -s $(pwd) ./_build/src/github.com/RedHatInsights/insights-ocp-controller

export GOPATH=$(pwd)/_build
go build -o insights-controller

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 ./insights-controller %{buildroot}%{_bindir}/insights-controller

%files
#%doc LICENSE README.md
%{_bindir}/insights-controller

%changelog
* Tue Apr 24 2018 Jeremy Crafts <jcrafts@redhat.com> - 1.0.0-0
- Initial Build
