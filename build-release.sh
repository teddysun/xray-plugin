#!/bin/bash
sum="sha1sum"

if ! hash sha1sum 2>/dev/null; then
    if ! hash shasum 2>/dev/null; then
        echo "I can't see 'sha1sum' or 'shasum'"
        echo "Please install one of them!"
        exit
    fi
    sum="shasum"
fi

[[ -z $upx ]] && upx="echo pending"
if [[ $upx == "echo pending" ]] && hash upx 2>/dev/null; then
    upx="upx -9"
fi

VERSION=$(git describe --tags)
LDFLAGS="-X main.VERSION=${VERSION} -s -w -buildid="

# Parse version for versioninfo (format: x.x.x.x or x.x.x)
# Remove 'v' prefix and extract version components
VERSION_CLEAN=${VERSION#v}
# Extract base version (remove -X-gXXXXXXX suffix if exists), e.g., v5.44.1-1-g00092ea -> 5.44.1
VERSION_BASE=$(echo "${VERSION_CLEAN}" | sed -E 's/-[0-9]+-g[0-9a-f]+$//')
IFS='.-' read -r MAJOR MINOR PATCH BUILD <<< "$VERSION_BASE"
# Set defaults if empty
MAJOR=${MAJOR:-0}
MINOR=${MINOR:-0}
PATCH=${PATCH:-0}
BUILD=${BUILD:-0}
# Ensure all version components are pure numbers (remove any non-digit characters)
MAJOR=$(echo "${MAJOR}" | tr -cd '0-9')
MINOR=$(echo "${MINOR}" | tr -cd '0-9')
PATCH=$(echo "${PATCH}" | tr -cd '0-9')
BUILD=$(echo "${BUILD}" | tr -cd '0-9')
# Re-validate defaults
MAJOR=${MAJOR:-0}
MINOR=${MINOR:-0}
PATCH=${PATCH:-0}
BUILD=${BUILD:-0}

# Generate versioninfo.json with actual version
if hash goversioninfo 2>/dev/null; then
    sed -e "s/%MAJOR%/$MAJOR/g" \
        -e "s/%MINOR%/$MINOR/g" \
        -e "s/%PATCH%/$PATCH/g" \
        -e "s/%BUILD%/$BUILD/g" \
        -e "s/%VERSION%/$VERSION_CLEAN/g" \
        versioninfo.json > versioninfo_generated.json
    # Generate platform-specific .syso files for all Windows architectures
    # This creates resource_windows_386.syso, resource_windows_amd64.syso, resource_windows_arm.syso, resource_windows_arm64.syso
    goversioninfo -platform-specific versioninfo_generated.json
fi

OSES=(linux darwin windows freebsd)
ARCHS=(amd64 386)

mkdir bin

for os in "${OSES[@]}"; do
    for arch in "${ARCHS[@]}"; do
        # Go 1.15 drops support for 32-bit binaries on macOS, iOS, iPadOS, watchOS, and tvOS (the darwin/386 and darwin/arm ports)
        # Reference URL: https://tip.golang.org/doc/go1.15#darwin
        if [ "${os}" == "darwin" ] && [ "${arch}" == "386" ]; then
            continue
        fi
        suffix=""
        if [ "${os}" == "windows" ]; then
            suffix=".exe"
        fi
        env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_${os}_${arch}${suffix}
        $upx xray-plugin_${os}_${arch}${suffix} >/dev/null
        tar -zcf bin/xray-plugin-${os}-${arch}-${VERSION}.tar.gz xray-plugin_${os}_${arch}${suffix}
        $sum bin/xray-plugin-${os}-${arch}-${VERSION}.tar.gz
    done
done

# ARM
ARMS=(5 6 7)
for v in "${ARMS[@]}"; do
    env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=${v} go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_arm${v}
done
$upx xray-plugin_linux_arm* >/dev/null
tar -zcf bin/xray-plugin-linux-arm-${VERSION}.tar.gz xray-plugin_linux_arm*
$sum bin/xray-plugin-linux-arm-${VERSION}.tar.gz

# ARM64 (ARMv8 or aarch64)
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_arm64
$upx xray-plugin_linux_arm64 >/dev/null
tar -zcf bin/xray-plugin-linux-arm64-${VERSION}.tar.gz xray-plugin_linux_arm64
$sum bin/xray-plugin-linux-arm64-${VERSION}.tar.gz

# Darwin ARM64
env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_darwin_arm64
$upx xray-plugin_darwin_arm64 >/dev/null
tar -zcf bin/xray-plugin-darwin-arm64-${VERSION}.tar.gz xray-plugin_darwin_arm64
$sum bin/xray-plugin-darwin-arm64-${VERSION}.tar.gz

# Windows ARM
env CGO_ENABLED=0 GOOS=windows GOARCH=arm go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_windows_arm.exe
$upx xray-plugin_windows_arm.exe >/dev/null
tar -zcf bin/xray-plugin-windows-arm-${VERSION}.tar.gz xray-plugin_windows_arm.exe
$sum bin/xray-plugin-windows-arm-${VERSION}.tar.gz

# Windows ARM64
env CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_windows_arm64.exe
$upx xray-plugin_windows_arm64.exe >/dev/null
tar -zcf bin/xray-plugin-windows-arm64-${VERSION}.tar.gz xray-plugin_windows_arm64.exe
$sum bin/xray-plugin-windows-arm64-${VERSION}.tar.gz

# MIPS
MIPSS=(mips mipsle)
for v in "${MIPSS[@]}"; do
    env CGO_ENABLED=0 GOOS=linux GOARCH=${v} go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_${v}
    env CGO_ENABLED=0 GOOS=linux GOARCH=${v} GOMIPS=softfloat go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_${v}_sf
done
$upx xray-plugin_linux_mips* >/dev/null
tar -zcf bin/xray-plugin-linux-mips-${VERSION}.tar.gz xray-plugin_linux_mips*
$sum bin/xray-plugin-linux-mips-${VERSION}.tar.gz

# MIPS64
MIPS64S=(mips64 mips64le)
for v in "${MIPS64S[@]}"; do
    env CGO_ENABLED=0 GOOS=linux GOARCH=${v} go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_${v}
done
tar -zcf bin/xray-plugin-linux-mips64-${VERSION}.tar.gz xray-plugin_linux_mips64*
$sum bin/xray-plugin-linux-mips64-${VERSION}.tar.gz

# ppc64le
env CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_ppc64le
$upx xray-plugin_linux_ppc64le >/dev/null
tar -zcf bin/xray-plugin-linux-ppc64le-${VERSION}.tar.gz xray-plugin_linux_ppc64le
$sum bin/xray-plugin-linux-ppc64le-${VERSION}.tar.gz

# s390x
env CGO_ENABLED=0 GOOS=linux GOARCH=s390x go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_s390x
$upx xray-plugin_linux_s390x >/dev/null
tar -zcf bin/xray-plugin-linux-s390x-${VERSION}.tar.gz xray-plugin_linux_s390x
$sum bin/xray-plugin-linux-s390x-${VERSION}.tar.gz

# riscv64
env CGO_ENABLED=0 GOOS=linux GOARCH=riscv64 go build -v -trimpath -ldflags "${LDFLAGS}" -o xray-plugin_linux_riscv64
$upx xray-plugin_linux_riscv64 >/dev/null
tar -zcf bin/xray-plugin-linux-riscv64-${VERSION}.tar.gz xray-plugin_linux_riscv64
$sum bin/xray-plugin-linux-riscv64-${VERSION}.tar.gz

# Clean up generated files
if [ -f "versioninfo_generated.json" ]; then
    rm -f versioninfo_generated.json
fi
# Clean up generated .syso files
rm -f resource_*.syso 2>/dev/null
