#!/bin/bash
#
# -*- coding: utf-8 -*-
#
# © Copyright 2023 GSI Helmholtzzentrum für Schwerionenforschung
#
# This software is distributed under
# the terms of the GNU General Public Licence version 3 (GPL Version 3),
# copied verbatim in the file "LICENCE".

set -e

function do_checks {

  info_msg="The build script must be executed from the projects base directory!"

  if [ -z "$VERSION" ]; then
    echo "ERROR: Build failed! VERSION file not found" >&2
    echo "INFO: $info_msg"
    exit 1
  fi

}

export VERSION=$(cat VERSION)
export BUILD_DIR=$HOME/rpmbuild
export PKG_DIR=prometheus-lustre-exporter-$VERSION

make build

sed -i "s/VERSION/$(cat VERSION)/" rpm/prometheus-lustre-exporter.spec
mkdir -p $BUILD_DIR/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
mkdir -p $BUILD_DIR/SOURCES/$PKG_DIR/usr/bin
mkdir -p $BUILD_DIR/SOURCES/$PKG_DIR/usr/lib/systemd/system
mkdir -p $BUILD_DIR/SOURCES/$PKG_DIR/etc/sysconfig
mkdir -p $BUILD_DIR/SOURCES/$PKG_DIR/etc/sudoers.d
mkdir -p $BUILD_DIR/SOURCES/$PKG_DIR/var/log/prometheus-lustre-exporter
cp rpm/prometheus-lustre-exporter.spec $BUILD_DIR/SPECS/
cp systemd/prometheus-lustre-exporter.service $BUILD_DIR/SOURCES/$PKG_DIR/usr/lib/systemd/system/
cp systemd/prometheus-lustre-exporter.options $BUILD_DIR/SOURCES/$PKG_DIR/etc/sysconfig/
cp sudoers/prometheus-lustre-exporter $BUILD_DIR/SOURCES/$PKG_DIR/etc/sudoers.d/
cp lustre_exporter $BUILD_DIR/SOURCES/$PKG_DIR/usr/bin/
cd $BUILD_DIR/SOURCES
tar -czvf $PKG_DIR.tar.gz $PKG_DIR
cd $BUILD_DIR
echo build dir is $BUILD_DIR
ls -la $BUILD_DIR/SOURCES
rpmbuild -ba  $BUILD_DIR/SPECS/prometheus-lustre-exporter.spec

