# Changelog

Notable changes to `qrocodile`. Versions follow [SemVer](https://semver.org/), with the usual pre-1.0 caveat: while the major is `0`, a **minor** bump may carry breaking changes.

## 0.2.0

### Added

- `generate` prints a link to stderr that opens the rendered code in the QR Designer, or says why there is none. `--open` opens it in your browser instead, and `--no-designer-url` leaves it out.
- `--preset` accepts `retroConsole`, `cityLights`, `goldenHour`, `snowfall`, `winterNight`, `gingerbread`, and `christmasTree`; `--module-style` accepts `knit`, `bokeh`, `snowfall`, `gingerbread`, and `christmas-tree`; `--finder-style` accepts `bokeh`.

### Changed

- **Breaking:** the `christmas` preset is now called `christmasSweater`, following the API. Designs and scripts that use `christmas` need to switch.
- `--margin` must be at least 1, since the API no longer accepts 0.
- Colors, in flags and in `--design`, accept only `#rgb` and `#rrggbb`; the API no longer accepts `#rrggbbaa`.

## 0.1.0

### Added

- `qrocodile generate <content>` renders a QR code as SVG or PNG (`--format`) via the QRocodile QR Code API.
  - `--preset`, `--module-style`, `--finder-style`, `--logo` (paired with `--logo-color`, because the API requires both and has no “use the icon’s original colors” option), `--module-color`, `--background`, `--finder-frame-color`, `--finder-eye-color`, `--error-correction` (`L`/`M`/`Q`/`H`, or `auto`, the default, to resolve it from the logo or style, or to clear a level set by `--design`), `--margin`, `--size`, `--fix-contrast` for the common styling knobs; `--design <file|-|json>` for a full design, the object the QR Designer’s “Copy JSON” button produces: a file path, `-` for stdin, or the JSON object itself. Flags set alongside `--design`/`--preset` override the loaded design’s own values for that field. `--preset`, `--module-style`, `--finder-style`, `--logo`, and `--error-correction` all tab-complete their valid values.
  - `-o, --output <path|->` writes exactly the bytes produced, whatever the file extension. With no `-o`: SVG prints to stdout unconditionally; PNG does too when stdout isn’t a terminal, but refuses to print raw bytes to an interactive one (`-o -` forces it), matching `curl`’s own guard against corrupting a terminal with binary output.
- `qrocodile auth login` walks through the email + 6-digit code signup and stores the resulting key locally. `--with-key` reads an existing key from stdin instead, the same pattern `gh auth login --with-token` uses.
- `qrocodile auth status` reports whether a key is stored (masked). It does not check that the key is still valid.
- `qrocodile auth logout` removes the stored key.
- `QROCODILE_API_KEY` always overrides the stored key (CI/scripting, matching `GH_TOKEN`/`AWS_ACCESS_KEY_ID`); `QROCODILE_API_BASE_URL` overrides the API base URL.
- `qrocodile version` (and `qrocodile --version`) and shell completion (`qrocodile completion bash|zsh|fish|powershell`, from Cobra).

### Notes

- The stored key lives in a plain, `0600`-permissioned config file (the OS-appropriate config directory), the same place `gh` and `aws` keep theirs by default. It is not stored in the OS keychain.
