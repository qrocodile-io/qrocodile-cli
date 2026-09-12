package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	qrocodile "github.com/qrocodile-io/qrocodile-api-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/qrocodile-io/qrocodile-cli/internal/apiclient"
	"github.com/qrocodile-io/qrocodile-cli/internal/output"
)

var (
	genFormat           string
	genPreset           string
	genModuleStyle      string
	genFinderStyle      string
	genLogo             string
	genLogoColor        string
	genDesign           string
	genSize             int
	genMargin           int
	genModuleColor      string
	genBackground       string
	genErrorCorrection  string
	genFinderFrameColor string
	genFinderEyeColor   string
	genFixContrast      bool
	genOutput           string
)

var generateCmd = &cobra.Command{
	Use:   "generate <content>",
	Short: "Render a QR code",
	Long: `Render a QR code from content.

  qrocodile generate "https://example.com"
  qrocodile generate "https://example.com" --preset ocean -o code.png --format png
  qrocodile generate "https://example.com" --module-style woodwork --finder-style diamond --logo website`,
	// A custom message rather than cobra.ExactArgs(1)'s generic "accepts 1 arg(s), received 0":
	// SilenceUsage is on globally (see root.go) so a runtime error — not logged in, an API
	// failure — doesn't dump the full flag reference underneath it. That means an argument
	// error needs to be self-explanatory on its own, since no usage block follows it either.
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf(`requires exactly one argument: the content to encode, e.g.:

  qrocodile generate "https://example.com"

see 'qrocodile generate --help' for the full flag reference`)
		}
		return nil
	},
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&genFormat, "format", "svg", "output format: svg or png")
	generateCmd.Flags().StringVar(&genPreset, "preset", "", "a built-in design preset")
	generateCmd.Flags().StringVar(&genModuleStyle, "module-style", "", "shape the modules (the pattern) are drawn with")
	generateCmd.Flags().StringVar(&genFinderStyle, "finder-style", "", "shape the three corner finders are drawn with")
	generateCmd.Flags().StringVar(&genLogo, "logo", "", "a built-in logo icon to place on the code — requires --logo-color, the API has no \"original colors\" option")
	generateCmd.Flags().StringVar(&genLogoColor, "logo-color", "", "color to recolor --logo to, e.g. #000000")
	generateCmd.Flags().StringVar(&genDesign, "design", "", "a full design: a JSON file path, - for stdin, or the JSON object itself, e.g. from the QR Designer's \"Copy JSON\"")
	registerValueCompletion(generateCmd, "format", func() []string { return []string{"svg", "png"} })
	registerValueCompletion(generateCmd, "preset", func() []string { return idStrings(qrocodile.PresetIDs()) })
	registerValueCompletion(generateCmd, "module-style", func() []string { return idStrings(qrocodile.ModuleStyleIDs()) })
	registerValueCompletion(generateCmd, "finder-style", func() []string { return idStrings(qrocodile.FinderStyleIDs()) })
	registerValueCompletion(generateCmd, "logo", func() []string { return idStrings(qrocodile.LogoIDs()) })
	generateCmd.Flags().IntVar(&genSize, "size", 0, "image width/height in pixels")
	generateCmd.Flags().IntVar(&genMargin, "margin", -1, "quiet zone in modules")
	generateCmd.Flags().StringVar(&genModuleColor, "module-color", "", "module (foreground) color, e.g. #000000")
	generateCmd.Flags().StringVar(&genBackground, "background", "", "background color, e.g. #ffffff, or \"transparent\"")
	generateCmd.Flags().StringVar(&genErrorCorrection, "error-correction", "", "error correction level: L, M, Q, or H — or auto (the default) to resolve it from the logo/style, e.g. to override a level set by --design")
	registerValueCompletion(generateCmd, "error-correction", func() []string { return []string{"auto", "L", "M", "Q", "H"} })
	generateCmd.Flags().StringVar(&genFinderFrameColor, "finder-frame-color", "", "color of the finder corners' outer frame, e.g. #000000")
	generateCmd.Flags().StringVar(&genFinderEyeColor, "finder-eye-color", "", "color of the finder corners' inner eye, e.g. #000000")
	generateCmd.Flags().BoolVar(&genFixContrast, "fix-contrast", true, "nudge low-contrast colors apart so the code stays scannable")
	generateCmd.Flags().StringVarP(&genOutput, "output", "o", "", "output path, or - for stdout (default: stdout)")
}

// registerValueCompletion wires a flag's shell completion to values(), called lazily at
// completion time rather than at init — the enum-ID flags' values() calls the SDK's generated
// value lists, which only need to exist, not be fetched from the network; still no reason to
// pay for it before a shell actually asks.
func registerValueCompletion(cmd *cobra.Command, flag string, values func() []string) {
	_ = cmd.RegisterFlagCompletionFunc(flag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return values(), cobra.ShellCompDirectiveNoFileComp
	})
}

// idStrings converts a slice of the SDK's string-based ID types (qrocodile.PresetID and
// friends) to plain strings, which is what cobra's completion API takes.
func idStrings[T ~string](ids []T) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	return out
}

func runGenerate(cmd *cobra.Command, args []string) error {
	content := args[0]
	flags := cmd.Flags()

	format := strings.ToLower(genFormat)
	switch format {
	case "svg", "png":
	default:
		return fmt.Errorf("--format must be svg or png, got %q", genFormat)
	}

	// The API requires both fields together for a built-in logo — there's no "use the icon's
	// original colors" option — so enforce the pairing here rather than let it round-trip to a
	// confusing 400 from the API.
	if flags.Changed("logo") != flags.Changed("logo-color") {
		return fmt.Errorf("--logo and --logo-color must be set together")
	}

	design, err := buildDesign(cmd, flags)
	if err != nil {
		return err
	}

	input := qrocodile.RenderInput{Content: content, Design: design}
	if flags.Changed("size") {
		input.Size = &genSize
	}
	if flags.Changed("fix-contrast") {
		input.FixContrast = &genFixContrast
	}

	client, err := apiclient.Resolve()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch format {
	case "svg":
		svg, err := client.RenderSVG(ctx, input)
		if err != nil {
			return err
		}
		return output.Write(genOutput, []byte(svg), false, cmd.OutOrStdout())
	case "png":
		png, err := client.RenderPNG(ctx, input)
		if err != nil {
			return err
		}
		return output.Write(genOutput, png, true, cmd.OutOrStdout())
	}
	panic("unreachable: format already validated")
}

// buildDesign assembles the design JSON sent to the API from --design (a base object, if any)
// overlaid with every other design-related flag that was set — matching the API's own rule
// that fields set alongside a preset override the preset's own values.
func buildDesign(cmd *cobra.Command, flags *pflag.FlagSet) (json.RawMessage, error) {
	design := map[string]any{}

	if flags.Changed("design") {
		data, err := readDesignSource(cmd, genDesign)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &design); err != nil {
			return nil, fmt.Errorf("--design does not contain a JSON object: %w", err)
		}
	}
	if flags.Changed("preset") {
		design["preset"] = genPreset
	}
	if flags.Changed("module-style") {
		design["moduleStyleId"] = genModuleStyle
	}
	if flags.Changed("finder-style") {
		design["finderStyleId"] = genFinderStyle
	}
	if flags.Changed("logo") {
		design["logo"] = map[string]any{"id": genLogo, "color": genLogoColor}
	}
	if flags.Changed("module-color") {
		design["moduleColor"] = genModuleColor
	}
	if flags.Changed("background") {
		design["background"] = genBackground
	}
	if flags.Changed("error-correction") {
		// "auto" isn't a real enum value the API accepts — auto-resolving means omitting the
		// field, not sending the literal string. Treated as an explicit clear, so it can strip
		// a level a base --design already set, rather than being silently dropped as a no-op.
		if strings.EqualFold(genErrorCorrection, "auto") {
			delete(design, "errorCorrection")
		} else {
			design["errorCorrection"] = genErrorCorrection
		}
	}
	if flags.Changed("finder-frame-color") {
		design["finderFrameColor"] = genFinderFrameColor
	}
	if flags.Changed("finder-eye-color") {
		design["finderEyeColor"] = genFinderEyeColor
	}
	if flags.Changed("margin") {
		design["margin"] = genMargin
	}

	if len(design) == 0 {
		return nil, nil
	}
	return json.Marshal(design)
}

// readDesignSource resolves --design's value: "-" for stdin, a leading "{" for the JSON
// object given inline (a real file path essentially never starts with "{", so this is
// unambiguous in practice), otherwise a file path.
func readDesignSource(cmd *cobra.Command, value string) ([]byte, error) {
	if value == "-" {
		return io.ReadAll(cmd.InOrStdin())
	}
	if strings.HasPrefix(strings.TrimSpace(value), "{") {
		return []byte(value), nil
	}
	data, err := os.ReadFile(value)
	if err != nil {
		return nil, fmt.Errorf("reading --design: %w", err)
	}
	return data, nil
}
