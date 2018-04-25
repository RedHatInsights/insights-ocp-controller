%global project       redhatinsights
%global repo          insights-ocp-controller
%global commit        v1.0.0

Name:           insights-ocp-controller
Version:        0.1
Release:        0%{?dist}
Summary:        Tool for extracting and serving content of container images
License:        ASL 2.0
URL:            https://github.com/redhatinsights/insights-ocp-controller
Source0:        ./insights-ocp-controller-0.1.tar.gz
BuildRequires:  golang >= 1.7

%description
Insights scan controller for Openshift Container Platform.

%prep
%setup -qn %{name}

%build
mkdir -p ./_build/src/github.com/RedHatInsights
ln -s $(pwd) ./_build/src/github.com/RedHatInsights/insights-ocp-controller
export GOPATH=$(pwd)/_build:$(pwd)/Godeps/_workspace:%{gopath}
go build -o insights-ocp-controller $(pwd)/_build/src/github.com/RedHatInsights/insights-ocp-controller/cmd.go

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 ./insights-ocp-controller %{buildroot}%{_bindir}/insights-ocp-controller

%files
#%doc LICENSE README.md
%{_bindir}/insights-ocp-controller

%changelog
* Tue Apr 24 2018 Jeremy Crafts <jcrafts@redhat.com> - 1.0.0-0
- Initial Build
