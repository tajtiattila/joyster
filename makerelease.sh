#!/bin/sh

if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
	echo "WARNING: Changed files present"
fi

if [ -n "$(git ls-files --others --exclude-standard)" ]; then
	echo "WARNING: Untracked files present"
fi

git status --short

if [ -z "$(git describe --exact-match HEAD 2>/dev/null)" ]; then
	echo "WARNING: No annotated tag for current version"
fi

PRG=joyster
DIR=release/$PRG

# HEAD must have an annotated tag
VER=$(git describe HEAD)

rm -rf release
mkdir -p release release/joyster/32bit $DIR/64bit

cp README.md $DIR
cp -a misc $DIR/misc

VJOYSDK=D:/dev/vJoy204SDK-080714
cp $VJOYSDK/lib/vJoyInterface.dll $DIR/32bit/
cp $VJOYSDK/lib/amd64/vJoyInterface.dll $DIR/64bit/

build () {
	go build -ldflags "-X main.Version $VER"
}

GOARCH=386 build && cp $PRG.exe $DIR/32bit/
GOARCH=amd64 build && cp $PRG.exe $DIR/64bit/

cd release
7z a joyster-$VER.zip joyster


