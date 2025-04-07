package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	fengchaogo "github.com/ijiwei/fengchao-go"
)

const supervisorSystem = `
你将充当一个监管者（Supervisor），负责控制和调度工作代理（Work Agent）的行为。在执行过程中，你首先需要进行推理，然后根据结果控制其他代理的行为，提供其他代理执行结果的反馈，评估是否继续执行任务，或根据需要调整执行路径。你不会直接回答问题，而是提供反馈意见，并评估代理的状态。

主要职责：
1. 分析并评估目标要求，确定代理的行动路径和执行步骤。
2. 实时监督其他代理的任务执行，确保其按预期执行。
3. 提供反馈意见，建议代理调整策略或改变执行路径。
4. 控制任务是否继续执行，或暂停并重新评估。

输出：
反馈应以JSON格式呈现，不使用任何 markdown 代码块或标记，包含以下字段：
1. feedback: 反馈意见或指令（字符串）
2.decision: 是否继续执行（布尔值）
3. reasoning: 监管者推理过程，说明为何做出此决定（字符串）
4. status: 当前代理状态（枚举："running"、"paused"、"completed"、 "failed"）

注意！
1. 确保所有信息都按照格式返回，避免遗漏或混淆。
2. 关注代理状态的变化，确保任务可以顺利完成。

示例：
输入
收集当天关于主机游戏的最新消息，并将我感兴趣的内容整合到一起，另存为PDF发送给我
输出:
{
	"feedback": "通过互联网搜索以下关键词《主机游戏》《steam》《Platstation》",
	"decision": true,
	"reasoning": "用户想要收集当天关于主机游戏的最新消息，并将感兴趣的内容整合起来，首先要联网获取当前最新的主机游戏消息, 然后将这些消息整合到一起，最后另存为PDF发送给用户。",
	"status": "running"
}
Work Agent:
关于主机游戏的最新信息：
1. 万众期待的中世纪开放世界角色扮演游戏《天国：拯救2》现已正式推出，相信有不少人已经准备踏上这段中世纪的史诗级旅程了。在这段刺激、惊奇的冒险当中，沿途或许也有不可错过的美景相伴。
2. 2月4日，任天堂公布了2025财年第3季度的财报。数据显示，2024年4~12月，任天堂销售额为9562亿日元，同比减少31.4%；营业利润为2475亿日元，同比减少46.7%。
3. 《无人深空》作为一款让玩家沉浸在浩瀚宇宙中的冒险游戏，再次带来了让人期待已久的更新——“世界 第二部分”。这一次，开发者不单单满足于扩大游戏的星系范围，更是带来了玩家一直渴望的内容：气态行星和深邃海洋的神秘世界。对于所有热衷探索和发现的玩家来说，这无疑是一次全新的、震撼的冒险体验！
输出:
{
	"feedback": "总结本周关于游戏的新闻, 注意体现关键词：主机游戏，steam，Platstation",
	"decision": true,
	"reasoning": "用户想要收集当天关于主机游戏的最新消息，并将感兴趣的内容整合起来，首先要联网获取当前最新的主机游戏消息, 然后将这些消息整合到一起，最后另存为PDF发送给用户。",
	"status": "running"
}
Work Agent: 
本周游戏界有几则重要新闻：
《天国：拯救2》正式发布：备受期待的中世纪开放世界角色扮演游戏《天国：拯救2》现已正式推出。玩家将踏上一段充满刺激与惊奇的中世纪史诗级旅程，沿途还能欣赏到不可错过的美景。
任天堂财报公布：2月4日，任天堂公布了2025财年第3季度的财报。数据显示，2024年4月至12月，任天堂的销售额为9562亿日元，同比减少31.4%；营业利润为2475亿日元，同比减少46.7%。
《无人深空》更新“世界 第二部分”：《无人深空》带来了令人期待已久的更新——“世界 第二部分”。此次更新不仅扩大了游戏的星系范围，还新增了气态行星和深邃海洋等玩家一直渴望的内容，为热衷探索和发现的玩家提供了全新的冒险体验。
输出:
{
	"feedback": "保存到《本周主机游戏要闻.pdf》中",
	"decision": true,
	"reasoning": "用户想要收集当天关于主机游戏的最新消息，并将感兴趣的内容整合起来，首先要联网获取当前最新的主机游戏消息, 然后将这些消息整合到一起，最后另存为PDF发送给用户。",
	"status": "running"
}
Work Agent:
已保存到 本周主机游戏要闻.pdf中，请检查，如果有任何疑问，欢迎随时与我们联系
输出:
{
	"feedback": "已完成",
	"decision": false,
	"reasoning": "用户想要收集当天关于主机游戏的最新消息，并将感兴趣的内容整合起来，首先要联网获取当前最新的主机游戏消息, 然后将这些消息整合到一起，最后另存为PDF发送给用户。",
	"status": "completed"
}
`

var ErrorCompleted = fmt.Errorf("supervisor already completed")

var client = fengchaogo.NewFengChao(os.Getenv("FENGCHAO_KEY"), os.Getenv("FENGCHAO_SECRET"), os.Getenv("FENGCHAO_BASE_URL"))

type Agent struct {
	Context *Context
	client  *fengchaogo.FengChao
	Stream  bool
}

func (a *Agent) AppendMessage(message fengchaogo.Prompt) {
	a.Context.History = fengchaogo.NewPromptTemplate(
		a.Context.History,
		message,
	)
}

func (a *Agent) Execute(ctx context.Context) (*fengchaogo.ChatCompletionResult, error) {
	res, err := a.client.ChatCompletion(
		ctx,
		a.Context.History,
		fengchaogo.WithModel("ERNIE-Bot-4"),
	)
	if err != nil {
		return nil, fmt.Errorf("agent chat failed, %s", err)
	}

	fmt.Println("Worker:")
	fmt.Println(res)

	a.Context.History = res.GetHistoryPrompts()
	return res, nil
}

type Supervisor struct {
	Context *Context
	client  *fengchaogo.FengChao
	Worker  *Agent
	Target  string
	Rounds  []Round
	Status  Status
}

// 初始化Supervisor, 将目标设置为Target
func NewSupervisor(Target string) *Supervisor {
	supervisor := &Supervisor{
		Context: &Context{
			History: fengchaogo.NewPromptTemplate(
				fengchaogo.NewMessage(
					fengchaogo.RoleSystem,
					supervisorSystem,
				),
				fengchaogo.NewMessage(
					fengchaogo.RoleUser,
					Target,
				),
			),
		},
		client: client,
		Target: Target,
		Worker: CreateWorker(),
	}
	return supervisor
}

// 对话
func (s *Supervisor) Execute(ctx context.Context) error {

	if s.Status == Completed {
		return ErrorCompleted
	}

	fmt.Println("Supervisor: 正在运行中...")
	// 对话将会添加一个轮次
	res, err := s.client.ChatCompletion(ctx, s.Context.History, fengchaogo.WithModel("gpt-4o"))
	if err != nil {
		return fmt.Errorf("supervisor chat failed, %s", err)
	}

	var result StepResult
	if err := json.Unmarshal([]byte(res.Choices[0].Message.Content), &result); err != nil {
		return fmt.Errorf("unmarshal llm ouput step result failed, %s", err)
	}

	fmt.Println("Supervisor Result:")
	fmt.Println("Status:", result.Status)
	fmt.Println("Feedback:", result.Feedback)
	fmt.Println("Reasoning:", result.Reasoning)

	s.Context.History = fengchaogo.NewPromptTemplate(
		s.Context.History,
		fengchaogo.NewMessage(
			fengchaogo.RoleAssistant,
			res.Choices[0].Message.Content,
		),
	)

	s.Worker.AppendMessage(fengchaogo.NewMessage(
		fengchaogo.RoleUser,
		result.Feedback,
	))

	currentRound := Round{
		Manager:  s,
		Worker:   s.Worker,
		Instruct: result,
	}

	s.Status = result.Status
	if result.Status != Completed {
		err := currentRound.Execute(ctx)
		if err != nil {
			return fmt.Errorf("execute worker failed, %s", err)
		}
	}

	s.Context.History = fengchaogo.NewPromptTemplate(
		s.Context.History,
		fengchaogo.NewMessage(
			fengchaogo.RoleUser,
			fmt.Sprintf("Worker Agent: \n%s", currentRound.Result.Choices[0].Message.Content),
		),
	)

	s.Rounds = append(s.Rounds, currentRound)
	return nil
}

type Context struct {
	History *fengchaogo.PromptTemplate
}

type Status string

const (
	Completed Status = "completed"
	Running   Status = "running"
	Failed    Status = "failed"
)

type StepResult struct {
	Feedback  string `json:"feedback"`
	Decision  bool   `json:"decision"`
	Reasoning string `json:"reasoning"`
	Status    Status `json:"status"`
}

type Round struct {
	Manager  *Supervisor
	Worker   *Agent
	Instruct StepResult
	Result   fengchaogo.ChatCompletionResult
}

func (r *Round) Execute(ctx context.Context) error {
	result, err := r.Worker.Execute(ctx)
	if err != nil {
		return fmt.Errorf("execute worker failed, %s", err)
	}
	r.Result = *result
	return nil
}

func CreateWorker() *Agent {
	worker := &Agent{
		Context: &Context{
			History: fengchaogo.NewPromptTemplate(
				fengchaogo.NewMessage(
					fengchaogo.RoleSystem,
					"you are a helpful assistant",
				),
			),
		},
		client: client,
	}
	return worker
}

func Start() {
	client.SetDebug(true)
	s := NewSupervisor("针对特朗普征收关税一事的最新情况写一篇公众号的文章")
	stdin := bufio.NewScanner(os.Stdin)
	fmt.Println("输入:exit 退出")
	for stdin.Scan() {
		input := stdin.Text()

		switch input {
		case ":exit":
			return
		}
		err := s.Execute(context.Background())
		if err != nil {
			if errors.Is(err, ErrorCompleted) {
				fmt.Println("任务已完成，退出")
				return
			}
			panic(err)
		}

	}

}

func main() {
	Start()
}
