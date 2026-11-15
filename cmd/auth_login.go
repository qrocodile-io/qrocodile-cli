package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	qrocodile "github.com/qrocodile-io/qrocodile-api-go"
	"github.com/spf13/cobra"

	"github.com/qrocodile-io/qrocodile-cli/internal/apiclient"
	"github.com/qrocodile-io/qrocodile-cli/internal/config"
)

var authLoginWithKey bool

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in, storing an API key locally",
	Long: `Sign in, storing an API key locally.

With no flags, walks the email + 6-digit code signup flow and stores the key it receives.

--with-key reads an existing key from stdin instead, for a key you got elsewhere (the
website, @qrocodile/api) or for scripting:

  echo "$KEY" | qrocodile auth login --with-key`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if authLoginWithKey {
			return loginWithKey(cmd)
		}
		return loginInteractive(cmd)
	},
}

func init() {
	authLoginCmd.Flags().BoolVar(&authLoginWithKey, "with-key", false, "read an existing API key from stdin instead of signing up")
}

func loginWithKey(cmd *cobra.Command) error {
	// The prompt goes to stderr, like loginInteractive's prompts, so it stays out of anything
	// reading stdout.
	fmt.Fprint(cmd.ErrOrStderr(), "Paste your API key, then press Enter: ")
	// A single line, not io.ReadAll: reading to EOF would block forever on an interactive
	// terminal (EOF there is Ctrl+D, not Enter) even though piped input works fine either way.
	key, err := readLine(bufio.NewReader(cmd.InOrStdin()))
	if err != nil {
		return fmt.Errorf("reading key from stdin: %w", err)
	}
	if key == "" {
		return fmt.Errorf("no key read from stdin")
	}
	if err := storeKey(key); err != nil {
		return err
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "API key saved.")
	return nil
}

func loginInteractive(cmd *cobra.Command) error {
	errOut := cmd.ErrOrStderr()
	in := bufio.NewReader(cmd.InOrStdin())

	fmt.Fprint(errOut, "Email: ")
	email, err := readLine(in)
	if err != nil {
		return err
	}
	if email == "" {
		return fmt.Errorf("no email entered")
	}

	client := apiclient.Unauthenticated()

	// Each API call gets its own timeout, created right before it. One shared timeout would
	// also count the time spent reading the email and typing the code, which can easily exceed
	// any reasonable request timeout.
	registerCtx, cancelRegister := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelRegister()
	if _, err := client.RegisterKey(registerCtx, email, ""); err != nil {
		return fmt.Errorf("requesting a key: %w", err)
	}
	fmt.Fprintln(errOut, "Check your email for the verification code.")

	fmt.Fprint(errOut, "Code: ")
	code, err := readLine(in)
	if err != nil {
		return err
	}
	if code == "" {
		return fmt.Errorf("no code entered")
	}

	confirmCtx, cancelConfirm := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelConfirm()
	result, err := client.ConfirmKey(confirmCtx, code, qrocodile.ConfirmKeyIdentifier{Email: email})
	if err != nil {
		return fmt.Errorf("confirming the code: %w", err)
	}
	if err := storeKey(result.ApiKey); err != nil {
		return err
	}

	if result.Replaced {
		fmt.Fprintln(errOut, "API key saved (replaced the previous key for this address).")
	} else {
		fmt.Fprintln(errOut, "API key saved.")
	}
	return nil
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func storeKey(key string) error {
	return config.Save(&config.Config{
		Credential: &config.Credential{Type: config.CredentialTypeAPIKey, Value: key},
	})
}
