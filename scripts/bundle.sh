#!/bin/sh
# Assembles bin/SATurn.app from bin/SATurn-gui and signs it.
# SATURN_SIGN_IDENTITY selects the signing identity (default: ad-hoc "-").
set -eu
VERSION=${VERSION:-0.0.0}
APP=bin/SATurn.app
IDENT=${SATURN_SIGN_IDENTITY:--}
[ -x bin/SATurn-gui ] || { echo "missing bin/SATurn-gui; run make app" >&2; exit 1; }
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp bin/SATurn-gui "$APP/Contents/MacOS/SATurn"
cp build/icons.icns "$APP/Contents/Resources/icons.icns"
sed "s/__VERSION__/$VERSION/g" build/Info.plist > "$APP/Contents/Info.plist"
xattr -cr "$APP"
codesign --force --identifier io.github.publicthrone547.saturn --sign "$IDENT" "$APP"
codesign --verify --strict "$APP"
if [ "$IDENT" = "-" ]; then echo "built $APP (ad-hoc signed)"; else echo "built $APP (signed with $IDENT)"; fi
du -sh "$APP"
