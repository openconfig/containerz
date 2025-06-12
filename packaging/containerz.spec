%define runopts %{runargs}

Name:		%{name}
Version:	%{version}
Release:	%{release}
Summary:	Containerz Service

Group:		Application/Misc
License:	Apache-2.0
Source0:	image

Requires:	%{requires}
BuildArch:	noarch

%description
containerz running as a container

%install
mkdir -p %{buildroot}/usr/share/%{name}
install -m 0644 %{SOURCE0} %{buildroot}/usr/share/%{name}/image

%files
%defattr(-,root,root,-)
/usr/share/%{name}

%post
docker load -i /usr/share/%{name}/image
docker tag rpmize:latest %{name}:latest
docker run --restart unless-stopped -d %{runopts} --name %{name} %{name}:latest

%preun
if [ $1 == 0 ]; then
  docker rm --force %{name}
fi
