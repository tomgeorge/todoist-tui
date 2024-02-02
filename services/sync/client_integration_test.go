package sync

import (
	"context"
	"fmt"
	"testing"

	"github.com/tomgeorge/todoist-tui/types"
)

func TestSync(t *testing.T) {
	cli := NewClient(nil).WithAuthToken("57c7e276c2251e2661a79f678020fdd202cdc97b")
	task, err := cli.AddTask(context.Background(), AddItemArgs{
		Content:     "call Steve tomorrow",
		Labels:      []string{"Urgent"},
		Description: "I need to call Steve!",
		Due: &types.DueDate{
			String: "Tomorrow at 5",
		},
	})
	fmt.Printf("task is %v err is %v", task, err)
}
