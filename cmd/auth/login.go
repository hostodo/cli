package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
	"github.com/hostodo/hostodo-cli/pkg/auth"
	"github.com/hostodo/hostodo-cli/pkg/config"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Hostodo",
	Long: `Authenticate with your Hostodo account using device flow.

This will:
1. Display a code to enter in your browser
2. Open your browser to the authorization page
3. Copy the code to your clipboard
4. Wait for you to authorize

Example:
  hostodo auth login`,
	Run: runLogin,
}

func init() {
	AuthCmd.AddCommand(loginCmd)
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))

	codeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981"))

	urlStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Underline(true)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true)
)

func runLogin(cmd *cobra.Command, args []string) {
	// Check if already authenticated
	if auth.IsAuthenticated() {
		fmt.Println(warningStyle.Render("You are already logged in."))
		fmt.Println("Run 'hostodo logout' first if you want to re-authenticate.")
		return
	}

	// Load config for API URL
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render("Error: ") + "Failed to load config: " + err.Error())
		os.Exit(1)
	}

	// Create OAuth client
	oauthClient := auth.NewDeviceFlowClient(cfg.APIURL)

	// Get device name
	deviceName := auth.GetDeviceName()

	fmt.Println()
	fmt.Println(titleStyle.Render("Hostodo CLI Authentication"))
	fmt.Println()

	// Initiate device flow
	deviceCode, err := oauthClient.InitiateDeviceFlow(deviceName)
	if err != nil {
		fmt.Println(errorStyle.Render("Error: ") + "Failed to start authentication: " + err.Error())
		os.Exit(1)
	}

	// Display the user code prominently using ASCII art
	displayUserCode(deviceCode.UserCode)

	// Show verification URL
	fmt.Println()
	fmt.Printf("  Visit: %s\n", urlStyle.Render(deviceCode.VerificationURI))
	fmt.Println()

	// Copy code to clipboard
	codeForClipboard := formatCodeWithDash(deviceCode.UserCode)
	if err := clipboard.WriteAll(codeForClipboard); err == nil {
		fmt.Println(successStyle.Render("  ✓") + " Code copied to clipboard")
	} else {
		fmt.Println(warningStyle.Render("  ⚠") + " Could not copy to clipboard")
	}

	// Open browser
	if err := browser.OpenURL(deviceCode.VerificationURI); err != nil {
		fmt.Println(warningStyle.Render("  ⚠") + " Could not open browser automatically")
		fmt.Printf("  Please visit the URL above manually\n")
	} else {
		fmt.Println(successStyle.Render("  ✓") + " Browser opened")
	}

	fmt.Println()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		cancel()
	}()

	// Poll for authorization with spinner
	token, err := pollWithSpinner(ctx, oauthClient, deviceCode)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			fmt.Println()
			fmt.Println(warningStyle.Render("Authentication cancelled"))
			os.Exit(0)
		}
		if errors.Is(err, auth.ErrAccessDenied) {
			fmt.Println()
			fmt.Println(errorStyle.Render("Access denied. ") + "Authorization was rejected.")
			os.Exit(1)
		}
		if errors.Is(err, auth.ErrExpiredToken) {
			fmt.Println()
			fmt.Println(errorStyle.Render("Code expired. ") + "Please try again.")
			os.Exit(1)
		}
		fmt.Println()
		fmt.Println(errorStyle.Render("Error: ") + err.Error())
		os.Exit(1)
	}

	// Save token to keychain
	if err := auth.SaveToken(token.AccessToken); err != nil {
		fmt.Println(errorStyle.Render("Error: ") + "Failed to save credentials: " + err.Error())
		os.Exit(1)
	}

	// Success!
	fmt.Println()
	fmt.Println(successStyle.Render("✓ Successfully authenticated!"))
	fmt.Println()
	fmt.Println("  You can now use the Hostodo CLI.")
	fmt.Println("  Try: hostodo instances list")
	fmt.Println()
}

// displayUserCode renders the user code as large ASCII art
func displayUserCode(userCode string) {
	// Format with dash: ABCD-EFGH
	formatted := formatCodeWithDash(userCode)

	// Create ASCII art
	fig := figure.NewColorFigure(formatted, "standard", "green", true)

	fmt.Println()
	fmt.Println("  Enter this code in your browser:")
	fmt.Println()
	fig.Print()
}

// formatCodeWithDash formats the 8-char code as XXXX-XXXX
func formatCodeWithDash(code string) string {
	code = strings.ToUpper(strings.ReplaceAll(code, "-", ""))
	if len(code) == 8 {
		return code[:4] + "-" + code[4:]
	}
	return code
}

// Spinner model for polling
type spinnerModel struct {
	spinner spinner.Model
	done    bool
	err     error
	token   *auth.TokenResponse
}

type pollingDoneMsg struct {
	token *auth.TokenResponse
	err   error
}

func pollWithSpinner(ctx context.Context, client *auth.DeviceFlowClient, deviceCode *auth.DeviceCodeResponse) (*auth.TokenResponse, error) {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := spinnerModel{
		spinner: s,
	}

	// Create Bubble Tea program
	p := tea.NewProgram(m)

	// Start polling in background
	go func() {
		interval := time.Duration(deviceCode.Interval) * time.Second
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.Send(pollingDoneMsg{err: ctx.Err()})
				return
			case <-ticker.C:
				token, err := client.PollForToken(ctx, deviceCode.DeviceCode, deviceCode.Interval)
				if err == nil {
					p.Send(pollingDoneMsg{token: token})
					return
				}
				if errors.Is(err, auth.ErrAuthorizationPending) {
					// Keep polling
					continue
				}
				if errors.Is(err, auth.ErrSlowDown) {
					// Increase interval
					interval += 5 * time.Second
					ticker.Reset(interval)
					continue
				}
				// Other errors (access_denied, expired_token)
				p.Send(pollingDoneMsg{err: err})
				return
			}
		}
	}()

	// Run spinner
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	result := finalModel.(spinnerModel)
	if result.err != nil {
		return nil, result.err
	}

	return result.token, nil
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pollingDoneMsg:
		m.done = true
		m.token = msg.token
		m.err = msg.err
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.done = true
			m.err = context.Canceled
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("  %s Waiting for authorization...\n", m.spinner.View())
}
