# misePTR

**misePTR** is a lightweight Kubernetes controller that watches new nodes and automatically updates their **DNS records**.

[![Test miseptr](https://github.com/vulnebify/miseptr/actions/workflows/test.yaml/badge.svg)](https://github.com/vulnebify/miseptr/actions/workflows/test.yaml)

---

## Installation

### Build locally

```bash
make build
```

---

## Usage

```bash
./bin/miseptr watch --provider vultr --suffix example.com
```

✅ Connects automatically to in-cluster Kubernetes or local `~/.kube/config`.

---

## Commands

| Command                | Description                                |
|-------------------------|--------------------------------------------|
| `miseptr watch`    | Watch nodes and update PTR records         |


### Flags

| Flag         | Description                           | Default    |
|--------------|---------------------------------------|------------|
| `--provider` | Hosting provider (e.g. `vultr`)       | `vultr`    |
| `--dns`      | DNS provider (e.g. `cloudflare`)      | *optional* |
| `--suffix`   | Suffix for PTR template (`%s.suffix`) | *required* |

---

## Testing

Install setup-envtest and run tests:

```bash
make setup-envtest
sudo make fetch-envtest-binaries
make test
```

✅ This sets up a local Kubernetes control plane for integration testing.

---

## GitHub Release

To create a versioned release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The binary will appear under [Releases](../../releases).

---

## License

This project is licensed under the [MIT License](./LICENSE).
