#!/usr/bin/env sh
# ds-rescue installer — curl this from README quickstart
set -e

REPO="zred0627/ds-rescue-cc"
PREFIX="${PREFIX:-/usr/local/bin}"

uname_s="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$uname_s" in
  linux)  os="linux" ;;
  darwin) os="darwin" ;;
  *) echo "Unsupported OS: $uname_s. Try Windows scripts/install.ps1"; exit 1 ;;
esac

uname_m="$(uname -m)"
case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported arch: $uname_m"; exit 1 ;;
esac

# Resolve latest tag via GitHub API
tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep -E '"tag_name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')"
[ -n "$tag" ] || { echo "Failed to fetch latest tag"; exit 1; }

archive="ds-rescue_${tag#v}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${tag}/${archive}"

tmpdir="$(mktemp -d)"
echo "Downloading $url ..."
curl -fsSL "$url" | tar -xz -C "$tmpdir"

install -m 0755 "$tmpdir/ds-rescue" "$PREFIX/ds-rescue"
rm -rf "$tmpdir"

echo "Installed ds-rescue to $PREFIX/ds-rescue"
echo "Verify: ds-rescue --version"
echo ""
echo "Next steps:"
echo "  1. Set API key: export DEEPSEEK_API_KEY=sk-... (or write to ~/.ds-rescue/key)"
echo "  2. Run self-check: ds-rescue --check"
