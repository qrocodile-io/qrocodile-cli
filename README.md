# QRocodile CLI

Generate styled QR codes (SVG or PNG) from a terminal, script, or CI pipeline, via the [QRocodile QR Code API](https://qrocodile.io/en/qr-code-api/).

## Install

```
go install github.com/qrocodile-io/qrocodile-cli/cmd/qrocodile@latest
```

## Get an API key

```
qrocodile auth login
```

Walks you through the email + 6-digit code signup — free, no card required — and stores the key locally once confirmed. Already have a key from elsewhere ([the website](https://qrocodile.io), [`@qrocodile/api`](https://www.npmjs.com/package/@qrocodile/api), [`qrocodile-api-go`](https://github.com/qrocodile-io/qrocodile-api-go))? Pipe it in instead:

```
echo "$KEY" | qrocodile auth login --with-key
```

`qrocodile auth status` shows whether a key is stored; `qrocodile auth logout` removes it.

## Generate a QR code

```
qrocodile generate "https://example.com"
```

With no flags, that renders a plain SVG to stdout. A few useful combinations:

```bash
# a built-in preset, saved to a file
qrocodile generate "https://example.com" --preset ocean -o code.svg

# PNG instead of SVG
qrocodile generate "https://example.com" --format png -o code.png

# a full design exported from the QR Designer's "Copy JSON" button — a file, stdin (-), or the JSON itself
qrocodile generate "https://example.com" --design my-design.json -o code.png --format png
qrocodile generate "https://example.com" --design '{"preset":"ocean","margin":4}' -o code.png --format png

# recolor a preset — flags set alongside --preset/--design override their values
qrocodile generate "https://example.com" --preset classic --module-color "#ff0000"

# module shape, finder-corner shape, and a built-in logo (--logo needs --logo-color — the
# API has no "use the icon's original colors" option)
qrocodile generate "https://example.com" --module-style woodwork --finder-style morphing --logo website --logo-color "#000000"

# recolor the finder corners independently of the modules, and force a higher error
# correction level (auto, the default, resolves it from the logo/style instead)
qrocodile generate "https://example.com" --finder-frame-color "#1b4332" --finder-eye-color "#52b788" --error-correction H
```

`--preset`, `--module-style`, `--finder-style`, `--logo`, and `--error-correction` all tab-complete their valid values.

`-o` writes exactly the bytes produced to that path (or `-` for stdout, unconditionally). With no `-o` at all: SVG prints to stdout; PNG does too when stdout isn’t a terminal (so `qrocodile generate ... --format png > code.png` works), but refuses to print raw PNG bytes to an interactive terminal — same guard `curl` uses for binary responses.

Full flag reference: `qrocodile generate --help`.

## Authentication in scripts and CI

`QROCODILE_API_KEY` always overrides the stored key, so a script or CI job never needs `auth login`:

```
QROCODILE_API_KEY=qk_live_… qrocodile generate "https://example.com" -o code.svg
```

`QROCODILE_API_BASE_URL` points at a different API instance (a dev/staging deployment, say) instead of the default `https://api.qrocodile.io`.

## Shell completion

```
qrocodile completion bash|zsh|fish|powershell
```

See `qrocodile completion --help` for how to load it in your shell.

## License

MIT
