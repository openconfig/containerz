#!/bin/bash
set +e

NAME="containerz"
IMAGE=${1:-"containerz:latest"}
VERSION=${2:-"0.01"}
RELEASE="1"
RUNARGS=${3:-"-p 19999:9999 -v /var/run/docker.sock:/var/run/docker.sock"}
REQUIRES="docker-ce"

WORK_DIR="/tmp/work-$(mktemp -u XXXXXX)"

rpmize () {
  mkdir -p $WORK_DIR/BUILD
  docker tag $IMAGE rpmize:latest
  docker save rpmize:latest > $WORK_DIR/image
  cp containerz.spec $WORK_DIR/containerz.spec
  rpmbuild -bb -D "image $IMAGE" -D "name $NAME" -D "release $RELEASE" -D "version $VERSION" -D "requires $REQUIRES" -D "_topdir $WORK_DIR" -D "runargs $RUNARGS" -D "_sourcedir %_topdir" -D "_rpmdir %_topdir" -D "_target_os linux" $WORK_DIR/containerz.spec
  mv $WORK_DIR/noarch/$NAME*rpm .
	

  echo "Created RPM $(du -h $NAME*rpm)"
}

swix () {
  mkdir -p $WORK_DIR/venv
  python3 -m venv $WORK_DIR/venv
  $WORK_DIR/venv/bin/pip3 install switools
  $WORK_DIR/venv/bin/swix-create $NAME.swix $NAME-$VERSION-$RELEASE.noarch.rpm
}

rpmize
swix

rm -Rf $WORK_DIR