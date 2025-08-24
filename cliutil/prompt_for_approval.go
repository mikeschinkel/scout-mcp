package cliutil

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/term"
)

func PromptForApproval(ctx context.Context, prompt string) (ar ApprovalResponse, err error) {
	var char byte
	var buffer [1]byte
	var stdinFd int
	var oldState *term.State

	ensureLogger()

	fmt.Print(prompt + "\n[ Y(es) / N(o) / A(ll) / add (D)elay / C(ancel) ]: ")

	// Try raw mode for single character input
	stdinFd = int(os.Stdin.Fd())
	oldState, err = term.MakeRaw(stdinFd)
	if err != nil {
		// Fallback for non-TTY environments (like GoLand console)
		if IsTerminalError(err) {
			fmt.Print("\n(Running in non-TTY environment; type choice and press Enter): ")
			ar, err = fallbackLineInput(ctx)
			goto end
		}
		goto end
	}

	// Read single character
	_, err = os.Stdin.Read(buffer[:])
	if err != nil {
		goto end
	}
	// Restore terminal immediately after reading
	defer must(term.Restore(stdinFd, oldState))
	_ = term.Restore(stdinFd, oldState)

	char = buffer[0]

	// Echo the character and add newline (now in normal mode)
	if char != 3 { // Don't echo Ctrl-C
		fmt.Printf("%c\n", char)
	} else {
		fmt.Println("^C")
	}
	ar, err = parseResponse(buffer[:])
end:
	return ar, err
}

func fallbackLineInput(ctx context.Context) (ar ApprovalResponse, err error) {
	var input string
	var inputChan chan string
	var errChan chan error

	inputChan = make(chan string, 1)
	errChan = make(chan error, 1)

	// Read input in a goroutine
	go func() {
		var scanInput string
		_, scanErr := fmt.Scanln(&scanInput)
		if scanErr != nil {
			errChan <- scanErr
			return
		}
		inputChan <- scanInput
	}()

	// Wait for either input or cancellation
	select {
	case <-ctx.Done():
		fmt.Println() // Add newline for clean output
		err = ctx.Err()
		goto end
	case err = <-errChan:
		goto end
	case input = <-inputChan:
		// Continue with processing
	}

	// Process the input (accept first character or full words)
	ar, err = parseResponse([]byte(input))

end:
	return ar, err
}

type ApprovalResponse = byte

const (
	NoResponse     = 'n'
	YesResponse    = 'y'
	AllResponse    = 'a'
	DelayResponse  = 'd'
	CancelResponse = 'c'
)

func parseResponse(resp []byte) (ar byte, err error) {
	// Process the response (
	if len(resp) > 0 {
		switch resp[0] {
		case 'y', 'Y':
			ar = YesResponse
		case 'a', 'A':
			ar = AllResponse
		case 'n', 'N':
			ar = NoResponse
		case 'd', 'D':
			ar = DelayResponse
		case 'c', 'C', 3: // Ctrl-C:
			ar = CancelResponse
			err = context.Canceled
		default:
			err = fmt.Errorf("invalid input: %s (expected Y/N/A/D/C)", string(resp))
		}
	}
	return ar, err
}

// ApprovalFunc is called before moving each message
// Returns: approved (move this message), approveAll (auto-approve remaining), error
// Context allows cancellation via Ctrl-C
type ApprovalFunc = func(context.Context, string) (ar ApprovalResponse, err error)
