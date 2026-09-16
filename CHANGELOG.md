# Changelog

Notable changes to `qrocodile`. Versions follow [SemVer](https://semver.org/), with the usual pre-1.0 caveat: while the major is `0`, a **minor** bump may carry breaking changes.

## 0.1.0

### Added

- `qrocodile generate <content>` — renders a QR code as SVG or PNG (`--format`) via the QRocodile API.
  - `--preset`, `--module-style`, `--finder-style`, `--logo` (paired with `--logo-color` — the API requires both, there’s no “use the icon’s original colors” option), `--module-color`, `--background`, `--finder-frame-color`, `--finder-eye-color`, `--error-correction` (`L`/`M`/`Q`/`H`, or `auto` — the default — to resolve it from the logo/style, or to clear a level set by `--design`), `--margin`, `--size`, `--fix-contrast` for the common styling knobs; `--design <file|-|json>` for a full design — a file path, `-` for stdin, or the JSON object itself — the object the QR Designer’s “Copy JSON” button produces. Flags set alongside `--design`/`--preset` override the loaded design’s own values for that field. `--preset`, `--module-style`, `--finder-style`, `--logo`, and `--error-correction` all tab-complete against the API’s own live enum.
  - `-o, --output <path|->` — writes exactly the bytes produced, no extension-sniffing. With no `-o`: SVG prints to stdout unconditionally; PNG does too when stdout isn’t a terminal, but refuses to print raw bytes to an interactive one (`-o -` forces it), matching `curl`’s own guard against corrupting a terminal with binary output.
- `qrocodile auth login` — the email + 6-digit code signup flow, storing the resulting key locally. `--with-key` reads an existing key from stdin instead, the same pattern `gh auth login --with-token` uses.
- `qrocodile auth status` — reports whether a key is stored (masked), without spending a real API call to validate it.
- `qrocodile auth logout` — removes the stored key.
- `QROCODILE_API_KEY` always overrides the stored key (CI/scripting, matching `GH_TOKEN`/`AWS_ACCESS_KEY_ID`); `QROCODILE_API_BASE_URL` overrides the API base URL.
- `qrocodile version` (and `qrocodile --version`) and shell completion (`qrocodile completion bash|zsh|fish|powershell`, from Cobra).

### Notes

- The stored key lives in a plain, `0600`-permissioned config file (the OS-appropriate config directory), the same tier `gh` and `aws` default to — not the OS keychain. It’s stored as a typed record (`{"type": "api_key", ...}`) so a future credential kind (an OAuth token pair, say) is a second variant of the same shape, not a breaking migration.
