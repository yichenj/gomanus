package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/yichenj/gomanus/internal/pkg/agent"
)

func DoConversations(ctx context.Context, a *agent.Agent) {
	history := make([]*schema.Message, 0)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Manus: Hello, what can I do for you?")
	for {
		fmt.Print("User: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Manus: Bye!")
				break
			}
			fmt.Printf("Unknown error: %s\n", err.Error())
		}
		if strings.ToLower(strings.TrimSpace(input)) == "bye" {
			fmt.Println("Manus: Bye!")
			break
		}

		history = append(history, schema.UserMessage(input))
		answer, err := a.Generate(ctx, history, agent.GetCallbackOpts(func(input string) {
			fmt.Println(input)
		}))
		if err != nil {
			fmt.Printf("Application error: %s\n", err.Error())
			break
		}
		history = append(history, answer)
		fmt.Printf("Manus: %s\n", strings.TrimSpace(answer.Content))
		//streamResult, err := agent.Stream(ctx, history)
		//if err != nil {
		//	fmt.Printf("Application error: %s\n", err.Error())
		//	break
		//}
		//fmt.Print("Manus:")
		//answer := ""
		//for {
		//	chunk, err := streamResult.Recv()
		//	if errors.Is(err, io.EOF) {
		//		break
		//	}
		//	if err != nil {
		//		fmt.Printf("Application error: %s\n", err.Error())
		//		break
		//	}
		//	fmt.Print(chunk.Content)
		//	answer += strings.TrimSpace(chunk.Content)
		//}
		//history = append(history, &schema.Message{
		//	Role:    schema.Assistant,
		//	Content: answer,
		//})
		//streamResult.Close()
		//fmt.Print("\n")
	}
}
