package main

import (
	"fmt"

	fengchao "github.com/ijiwei/fengchao-go"
)

const systemPrompt = `
你是一名有多年经验的文字内容创作者，你的工作内容包含：
1. 针对参考内容进行分析，并确定文章和稿件的选题。
2. 根据已确定的选题和内容撰写文章和稿件的正文。
你要严格准守以下的工作规范和要求：
1. 信息必须经过充分的事实核查，确保内容的真实性和准确性，杜绝虚假、夸大的内容。
2. 内容需通过反抄袭检测工具，重复率不得超过10%。
3. 段落长度适中，使用小标题、列表等形式优化内容结构，增强可读性。
4. 语言表达需简洁流畅，避免使用晦涩难懂的术语和复杂的句式。风格应符合目标读者群体的阅读习惯。
5. 遵守内容生产的行业道德规范，避免涉及敏感话题、歧视性言论、暴力和色情内容。
`

const ContentGeneratorPrompt = `
本次工作为根据已确认的选题《{{.title}}》，和参考内容完成一篇关于[{{.tags}}]的文章，文章中应避免使用总结、结论等类似的段落。
你要清楚，文章内容将会直接发表到新闻媒体中，稿件的阅读量会直接决定你的绩效考核成绩，请严格按照工作规范来完成，这将会影响你的职业生涯。
以下为本次选题的相关参考内容：
{{.text}}
`

func example() {
	// 渲染系统消息
	systemMessage := fengchao.NewMessage(fengchao.RoleSystem, systemPrompt)
	sm, err := systemMessage.Render(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(sm))
	sm, err = systemMessage.Render(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(sm))

	// 渲染 Prompt
	promptOne := fengchao.NewPromptTemplate(
		systemMessage,
		// 如果直接使用Message, 无法渲染，所以如果有变量的话，更推荐直接使用NewMessage
		// 当然你不想用提供的模板变量，也可以自己生成message的content
		&fengchao.Message{
			Role: fengchao.RoleUser,
			Content: `1+1=2
			对吗？`,
		},
		fengchao.NewMessage(fengchao.RoleAssistant, "对的"),
		fengchao.NewMessage(fengchao.RoleUser, "你的名字是{{.name}}吗?"),
	)
	promptOne.HumanFriendly = true
	m, err := promptOne.Render(map[string]interface{}{"name": "fengchao"})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(m))
	m, err = promptOne.Render(map[string]interface{}{"name": "fengchao"})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(m))

	// 使用NewMessage可以使用 template 来渲染
	prompt := fengchao.NewPromptTemplate(
		promptOne,
		fengchao.NewMessage(fengchao.RoleAssistant, "是的"),
		fengchao.NewMessage(fengchao.RoleUser, ContentGeneratorPrompt),
	)
	prompt.HumanFriendly = true
	m, err = prompt.Render(map[string]interface{}{"name": "fengchao", "title": "文章标题", "tags": "文章标签", "text": "文章内容"})

	if err != nil {
		panic(err)
	}
	fmt.Println(string(m))
}

func main() {
	// example()
    prompt := fengchao.NewPromptTemplate(
        fengchao.NewMessage(fengchao.RoleSystem, `你是一个非常厉害的{{.Name}}！`),
        fengchao.NewMessage(fengchao.RoleUser, `分别讲一个关于{{range .Items}}、{{.}}{{end}}的笑话吧`),
		fengchao.NewMessage(fengchao.RoleAssistant, `小猫：小猫去银行，工作人员问：“你要存什么？”小猫眨眨眼说：“我存爪印！”
小狗：小狗学会了打字，但每次发的都是“汪汪汪”，它说：“我这不是在聊天，是在打码！”
小狐狸：小狐狸问妈妈：“为什么我们叫狡猾？”妈妈笑着说：“因为我们知道怎么用优惠券！”`),
    )
	prompt = fengchao.NewPromptTemplate(
		prompt,
        fengchao.NewMessage(fengchao.RoleUser, `再讲{{.Count}}个好不好？`),
	)
	prompt.HumanFriendly = true
    PromptJson, err := prompt.Render(map[string]interface{}{
        "Items": []string{"小猫", "小狗", "小狐狸"},
        "Name": "智能助手",
		"Count": 3,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(string(PromptJson)) 

}
