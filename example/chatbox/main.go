package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	fengchaogo "github.com/ijiwei/fengchao-go"
)

var client = fengchaogo.NewFengChao(os.Getenv("FENGCHAO_KEY"), os.Getenv("FENGCHAO_SECRET"), os.Getenv("FENGCHAO_BASE_URL"))

var systemMessage = fengchaogo.NewMessage(fengchaogo.RoleSystem, `你是一名善于理解问题的助手，你要按照以下的规则与用户对话:
1. 采用风趣幽默的回答，适当添加Emoji来让回答更加形象
2. 回答的内容尽可能丰富，如果篇幅过长，你可以先对问题进行总结并生成大纲，并通过多次对话的方式分步进行回答
3. 你的回答要有主观性，不要拿用户的意见和建议作为依据
`)

func ChatBox() {
	fmt.Println("FENGCHAO(https://github.com/ijiwei/fengchao-go)")
	fmt.Println("---------------------")
	fmt.Println("输入:help 获取帮助信息")
	fmt.Print("> ")
	s := bufio.NewScanner(os.Stdin)
	var historyMessage *fengchaogo.PromptTemplate
	for s.Scan() {
		input := s.Text()

		switch input {
		case ":help":
			fmt.Println("clear: 清除历史消息")
			fmt.Println("history: 显示历史消息")
			fmt.Println("exit: 退出")
			fmt.Print("> ")
		case ":clear":
			historyMessage = nil
			fmt.Print("已清除历史消息\n> ")
			continue
		case ":exit":
			return
		case ":history":
			historyDisplay(historyMessage)
			continue
		case "":
			continue
		}

		inputMessage := fengchaogo.NewMessage(fengchaogo.RoleUser, input)
		res, err := client.ChatCompletionStream(
			context.Background(),
			fengchaogo.NewPromptTemplate(
				systemMessage,
				historyMessage,
				inputMessage,
			),
			fengchaogo.WithIsSensitive(true),
			fengchaogo.WithModel("glm-4"),
		)
		if err != nil {
			panic(err)
		}

		answer := ""
		for r := range res.Stream() {
			fmt.Print(r.String())
			answer = answer + r.String()
		}

		historyMessage = fengchaogo.NewPromptTemplate(
			historyMessage,
			inputMessage,
			fengchaogo.NewMessage(fengchaogo.RoleAssistant, answer),
		)
		fmt.Print("\n> ")
	}
}

func historyDisplay(history *fengchaogo.PromptTemplate) {
	if history == nil {
		fmt.Println("没有历史消息")
		fmt.Print("> ")
	}
	messages, _ := history.RenderMessages(nil)
	for _, m := range messages {
		fmt.Printf(">> %s: %s\n", m.Role, m.Content)
	}
	fmt.Print("> ")
}

func main() {
	ChatBox()
}
