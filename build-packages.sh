#!/bin/bash

# Build script for hani packages
set -e

VERSION="1.2.1"
ARCH="amd64"

echo "🏗️  Building hani binary..."
mise exec -- make build

echo "📦 Creating package directories..."
mkdir -p hani_deb_package/DEBIAN hani_deb_package/usr/bin
mkdir -p hani_rpm_package/SPECS hani_rpm_package/SOURCES hani_rpm_package/BUILD hani_rpm_package/RPMS hani_rpm_package/SRPMS

echo "📋 Creating .deb package..."
cp hani hani_deb_package/usr/bin/
chmod +x hani_deb_package/usr/bin/hani

# Create control file for .deb
cat > hani_deb_package/DEBIAN/control << EOF
Package: hani
Version: $VERSION
Section: editors
Priority: optional
Architecture: $ARCH
Maintainer: Tim Apple <timappledotcom@users.noreply.github.com>
Description: A TUI Markdown Editor
 Hani is a terminal-based markdown editor with vim-like keybindings
 and live preview functionality. Built with Go and Bubbletea.
 .
 Features:
  - Vim-like modal editing (normal/insert modes)
  - Live markdown preview
  - Syntax highlighting
  - Tab-based interface
  - File creation and editing
  - Catppuccin color scheme
EOF

# Build .deb package
dpkg-deb --build hani_deb_package hani_${VERSION}_${ARCH}.deb

echo "📋 Creating .rpm package..."
# Create spec file for .rpm
cat > hani_rpm_package/SPECS/hani.spec << EOF
Name: hani
Version: $VERSION
Release: 1
Summary: A TUI Markdown Editor
License: MIT
Group: Applications/Editors
Source0: hani
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
BuildArch: x86_64

%description
Hani is a terminal-based markdown editor with vim-like keybindings
and live preview functionality. Built with Go and Bubbletea.

Features:
- Vim-like modal editing (normal/insert modes)
- Live markdown preview
- Syntax highlighting
- Tab-based interface
- File creation and editing
- Catppuccin color scheme

%prep

%build

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}/usr/bin
cp %{_sourcedir}/hani %{buildroot}/usr/bin/
chmod +x %{buildroot}/usr/bin/hani

%files
%defattr(-,root,root,-)
/usr/bin/hani

%changelog
* $(date "+%a %b %d %Y") Tim Apple <timappledotcom@users.noreply.github.com> - $VERSION-1
- Major stability and feature release
- Added comprehensive test suite
- Enhanced syntax highlighting
- Improved error handling and performance
- Production-ready with robust file management
EOF

# Copy binary to SOURCES
cp hani hani_rpm_package/SOURCES/

# Build .rpm package (if rpmbuild is available)
if command -v rpmbuild &> /dev/null; then
    rpmbuild --define "_topdir $(pwd)/hani_rpm_package" -bb hani_rpm_package/SPECS/hani.spec
    cp hani_rpm_package/RPMS/x86_64/hani-${VERSION}-1.x86_64.rpm ./hani_${VERSION}_x86_64.rpm
    echo "✅ RPM package created: hani_${VERSION}_x86_64.rpm"
else
    echo "⚠️  rpmbuild not available, skipping RPM package creation"
    echo "   Install rpm-build package to create RPM packages"
fi

echo "✅ DEB package created: hani_${VERSION}_${ARCH}.deb"

echo "🧹 Cleaning up temporary directories..."
rm -rf hani_deb_package hani_rpm_package

echo "🎉 Package build complete!"
echo "📦 Created packages:"
ls -la *.deb *.rpm 2>/dev/null || echo "   - hani_${VERSION}_${ARCH}.deb"
