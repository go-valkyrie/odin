# Installation

There are several ways to install Odin depending on your platform and preferences.

## Container Image

The easiest way to use Odin is via the official container image:

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/go-valkyrie/odin:latest template /workspace
```

Or with podman:

```bash
podman run --rm -v $(pwd):/workspace:z ghcr.io/go-valkyrie/odin:latest template /workspace
```

The image is built for both `amd64` and `arm64` architectures.

## Binary Releases

Pre-built binaries are available for multiple platforms from the [GitHub Releases](https://github.com/go-valkyrie/odin/releases) page.

### macOS

```bash
# Download the latest release for your architecture
# For Apple Silicon (M1/M2/M3)
curl -LO https://github.com/go-valkyrie/odin/releases/latest/download/odin-darwin-arm64

# For Intel Macs
curl -LO https://github.com/go-valkyrie/odin/releases/latest/download/odin-darwin-amd64

# Make it executable and move to your PATH
chmod +x odin-darwin-*
sudo mv odin-darwin-* /usr/local/bin/odin
```

### Linux

```bash
# For x86_64
curl -LO https://github.com/go-valkyrie/odin/releases/latest/download/odin-linux-amd64

# For ARM64
curl -LO https://github.com/go-valkyrie/odin/releases/latest/download/odin-linux-arm64

# Make it executable and move to your PATH
chmod +x odin-linux-*
sudo mv odin-linux-* /usr/local/bin/odin
```

### Windows

Download the appropriate binary from the [releases page](https://github.com/go-valkyrie/odin/releases) and add it to your PATH.

## Build from Source

If you have Go 1.24.1 or later installed, you can build Odin from source:

```bash
# Clone the repository
git clone https://github.com/go-valkyrie/odin.git
cd odin

# Build the binary
go build ./cmd/odin

# Optionally, install it
go install ./cmd/odin
```

Or install directly:

```bash
go install go-valkyrie.com/odin/cmd/odin@latest
```

## Verify Installation

Once installed, verify Odin is working:

```bash
odin --help
```

You should see the main help output showing available commands.

## Next Steps

Now that you have Odin installed, let's [create your first bundle](./quick-start.md)!
