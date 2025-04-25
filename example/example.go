package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	fengchao "github.com/ijiwei/fengchao-go"
)

var client = fengchao.NewFengChao(os.Getenv("FENGCHAO_KEY"), os.Getenv("FENGCHAO_SECRET"), os.Getenv("FENGCHAO_BASE_URL"))

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
根据已确认的选题《{{.title}}》，和参考内容完成一篇关于[{{.tags}}]的文章，文章中应避免使用总结、结论等类似的段落。
你要清楚，文章内容将会直接发表到新闻媒体中，稿件的阅读量会直接决定你的绩效考核成绩，请严格按照工作规范来完成，这将会影响你的职业生涯。
以下为本次选题的相关参考内容：
{{.text}}
`

func PromptUseCase() {
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

func PromptUseCaseTwo() {
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
		"Name":  "智能助手",
		"Count": 3,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(PromptJson))
}

func SimpleChat() {

	res, err := client.ChatCompletion(
		context.Background(),
		fengchao.NewMessage(fengchao.RoleUser, "讲一个{{.Story}}"),
		fengchao.WithParams(struct {
			Story string
		}{
			Story: "冷笑话",
		}),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("结果如下：")
	fmt.Println(res)
	historyData, err := json.MarshalIndent(res.History, "", "	")
	if err != nil {
		panic(fmt.Sprintf("marshal history error: %v", err))
	}

	fmt.Println("对话记录如下：")
	fmt.Println(string(historyData))
}

func ChatWithHistory() {

	client.SetDebug(true)

	ctx := context.Background()

	res, err := client.ChatCompletion(
		ctx,
		fengchao.NewPromptTemplate(
			fengchao.NewMessage(fengchao.RoleSystem, systemPrompt),
			fengchao.NewMessage(fengchao.RoleUser, `
本次工作为根据已确认的选题《{{.Title}}》，和参考内容完成一篇关于[{{.Tags}}]的文章，文章中应避免使用总结、结论等类似的段落。
你要清楚，文章内容将会直接发表到新闻媒体中，稿件的阅读量会直接决定你的绩效考核成绩，请严格按照工作规范来完成，这将会影响你的职业生涯。
以下为本次选题的相关参考内容：
{{.Text}}
`),
		),
		fengchao.WithTemperature(0.9),
		fengchao.WithModel("gpt-4o"),
		fengchao.WithParams(struct {
			Title string
			Text  string
			Tags  string
		}{
			Title: `国产AI增强操作系统发布：填补端侧推理空白`,
			Text: `8月8日举行的2024中国操作系统产业大会上，国产桌面操作系统银河麒麟发布首个AIPC版本，这是一款与人工智能融合的国产桌面操作系统，填补了我国操作系统端侧推理能力研发的空白。
操作系统是计算机之魂，承接上层软件生态与底层硬件资源，为AI算法、模型与应用的运行提供支撑环境，在IT国产化中发挥重要作用。过去很长一段时间，全球操作系统厂商主要为欧美企业。
我国操作系统发展起步晚、系统生态存在短板，赶超压力大。新一轮人工智能技术的迅猛发展，为我国操作系统带来新机遇。`,
			Tags: `#AI操作系统#国产操作系统#端侧推理`,
		}),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("结果如下：")
	fmt.Println(res)
	fmt.Println("继续对话")

	res, err = client.ChatCompletion(
		ctx,
		fengchao.NewPromptTemplate(
			res.GetHistoryPrompts(),
			fengchao.NewMessage(fengchao.RoleUser, `根据文章内容，总结一份{{.language}}摘要`),
		),
		fengchao.WithTemperature(0.9),
		fengchao.WithModel("glm-4"),
		fengchao.WithParams(map[string]interface{}{"language": "中文"}), // 也可以使用map[string]interface{}传递参数
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("结果如下：")
	fmt.Println(res)

	historyData, err := json.MarshalIndent(res.History, "", "	")
	if err != nil {
		panic(fmt.Sprintf("marshal history error: %v", err))
	}

	fmt.Println("对话记录如下：")
	fmt.Println(string(historyData))
}

func ReadStream() {

	// client.SetDebug(true)

	ctx := context.Background()

	prompt := fengchao.NewPromptTemplate(
		fengchao.NewMessage(fengchao.RoleSystem, `你是一个非常厉害的{{.Name}}！`),
		fengchao.NewMessage(fengchao.RoleUser, `分别讲一个关于{{range .Items}}、{{.}}{{end}}的笑话吧`),
		fengchao.NewMessage(fengchao.RoleAssistant, `小猫：小猫去银行，工作人员问：“你要存什么？”小猫眨眨眼说：“我存爪印！”
小狗：小狗学会了打字，但每次发的都是“汪汪汪”，它说：“我这不是在聊天，是在打码！”
小狐狸：小狐狸问妈妈：“为什么我们叫狡猾？”妈妈笑着说：“因为我们知道怎么用优惠券！”`),
		fengchao.NewUserMessage("再讲一个"),
	)

	res, err := client.ChatCompletionStream(
		ctx,
		prompt,
		fengchao.WithTemperature(1.9),
		fengchao.WithModel("gpt-4o"),
		// fengchao.WithIsSensitive(true),
		fengchao.WithParams(map[string]interface{}{
			"Items": []string{"中国", "台湾", "香港"},
			"Name":  "智能助手",
			"Count": 3,
		}),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("结果如下：")

	for {
		chunk, finished, err := res.Read()
		if finished {
			break
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("EOF")
				break
			}
			panic(err)
		}

		fmt.Print((*chunk).String())
	}
	fmt.Print("\n")
	res.Close()
}

func Stream() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	client.SetDebug(true)

	ctx := context.Background()

	prompt := fengchao.NewPromptTemplate(
		fengchao.NewMessage(fengchao.RoleUser, `进行一个大阿尔卡那的塔罗牌占卜,使用十字法牌陣🔮`),
	)

	res, err := client.ChatCompletionStream(
		ctx,
		prompt,
		fengchao.WithTimeout(2), // 流式接口设置超时无效
		fengchao.WithTemperature(0.9),
		fengchao.WithModel("glm-41"),
		// fengchao.WithIsSensitive(true),
	)

	if err != nil {
		panic("ChatCompletionStream Failed: " + err.Error())
	}

	fmt.Println("结果如下：")
	for r := range res.Stream() {
		fmt.Print((r).String())
	}
	fmt.Print("\n")
}

func QuickChatCompletion() {
	client.SetDebug(true)
	res, err := client.QuickCompletion(
		context.Background(),
		fengchao.WithPredefinedPrompts("多译英"),
		fengchao.WithModel("gpt-4o,moonshot-v1-128k"),
		fengchao.WithQuery(`命运之轮象征着命运的起伏和变化，它代表着生活中不可预测的转变和机遇。这张牌可能意味着你正处在一个重要的转折点，你将会经历一些意想不到的改变。这些改变可能会带来新的机会和挑战，需要你灵活适应并做好准备。
命运之轮也提醒我们，生活中的好运和不幸都是暂时的，一切都在不断变化中。这张牌鼓励你保持乐观和开放的态度，相信未来会带来更好的机会和成长。同时，也要学会珍惜当下，充分利用现有的资源和机会。`),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("结果如下：")
	fmt.Println(res)
}

func BatchComplete() {

	client.SetDebug(true)
	builder := fengchao.NewBatchChatCompletionBuilder()

	one, _ := builder.Add(
		nil,
		fengchao.WithPredefinedPrompts("多译英"),
		fengchao.WithQuery(`命运之轮象征着命运的起伏和变化，它代表着生活中不可预测的转变和机遇。这张牌可能意味着你正处在一个重要的转折点，你将会经历一些意想不到的改变。这些改变可能会带来新的机会和挑战，需要你灵活适应并做好准备。
命运之轮也提醒我们，生活中的好运和不幸都是暂时的，一切都在不断变化中。这张牌鼓励你保持乐观和开放的态度，相信未来会带来更好的机会和成长。同时，也要学会珍惜当下，充分利用现有的资源和机会。`),
	)

	two, _ := builder.Add(
		fengchao.NewPromptTemplate(
			fengchao.NewMessage(fengchao.RoleUser, `进行一个大阿尔卡那的塔罗牌占卜,使用十字法牌陣🔮`),
		),
	)

	res, fail, complete := client.BatchChatCompletion(context.Background(), builder)
	if !complete {
		for k, f := range fail {
			switch k {
			case one:
				fmt.Println("1. 失败原因：")
			case two:
				fmt.Println("2. 失败原因：")
			}
			fmt.Println(f)
		}
	}

	fmt.Println("1. 结果如下：")
	fmt.Println(res[one])

	fmt.Println("2. 结果如下：")
	fmt.Println(res[two])
}

func main() {
	client.SetDebug(true)
	res, err := client.QuickCompletion(
		context.Background(),
		fengchao.WithPredefinedPrompts("IC-机构报告-正文"),
		fengchao.WithPromptFill(map[string]string{
			"reference": `美国总统特朗普宣布，将从4月5日起对所有出口到美国的商品征收至少10%的关税，其中，中国将面临34%、欧盟为20%的关税等。

特朗普在发布会上展示一张图表，指出数十个贸易不平衡现象最严重的国家将面临更高的税率。以中国为例，中国对美国商品征收高达67%关税，因此美国相应对其实施34%的关税；欧盟将面临20%的关税，英国为10%；越南为46%；日本24%；韩国25%；印度26%；柬埔寨49%等关税。

特朗普重申，周四凌晨12时开始对所有进口汽车征收25%的关税。

美国官员其后表示，对中国的实际总关税率升至54%。美国财长贝森特指，发布的数字是区间的高端，而对中国的关税税率，是累计水平，并可能会对中国小型炼油厂有更多制裁，并建议各国不要报复。

白宫表示，10%的基准税率4月5日生效，更高的关税4月9日生效。

4月3日，商务部新闻发言人就美方宣布对等关税发表谈话，商务部新闻发言人表示，中方注意到，美东时间4月2日，美方宣布对所有贸易伙伴征收“对等关税”。中方对此坚决反对，并将坚决采取反制措施维护自身权益。

商务部新闻发言人指出，美方声称自己在国际贸易中吃了亏，以所谓“对等”为由提高对所有贸易伙伴的关税，这种做法罔顾多年来多边贸易谈判达成的利益平衡结果，也无视美方长期从国际贸易中大量获利的事实。美方在主观、单方面评估基础上，得出所谓“对等关税”，不符合国际贸易规则，严重损害相关方的正当合法权益，是典型的单边霸凌做法。对此，很多贸易伙伴已经表达强烈不满和明确反对。

商务部新闻发言人强调，历史证明，提高关税解决不了美国自身问题，既损害美国自身利益，也危及全球经济发展和产供链稳定。贸易战没有赢家，保护主义没有出路。中方敦促美方立即取消单边关税措施，与贸易伙伴通过平等对话妥善解决分歧。

`}),
		fengchao.WithModel("glm-4-plus"),
		fengchao.WithQuery(`2024年，全球半导体行业虽然未全面复苏，但生成式人工智能（AI）、汽车电子和通信技术的快速发展为2025年的技术进步奠定了坚实基础，2025年正成为半导体产业浪潮中的关键转折点。据半导体情报（Semiconductor Intelligence，SC-IQ）预测，2025年全球半导体资本支出将增长3%至1600亿美元。然而，这一增长背后企业发展并不均衡：台积电和美光科技逆势加码投资，而英特尔和三星却计划大幅削减开支。与此同时，特朗普上台后对美国《芯片法案》的实施带来的不确定性，也为行业前景蒙上了阴影。

2025年半导体资本支出增至1600亿美元 英特尔、三星削减开支

半导体情报（Semiconductor Intelligence，SC-IQ）估计，2024年半导体产业资本支出（CapEx）为1550亿美元，比2023年的1640亿美元下降5%。

SC-IQ预计，2025年半导体产业资本支出将增长3%至1600亿美元。

SC-IQ指出，2025年的增长主要由两家公司推动。全球最大的晶圆代工公司台积电计划2025年的资本支出在380亿~420亿美元之间。如果使用中间值，这将增加100亿美元或34%。美光科技预计其截至8月的2025财年的资本支出为140亿美元，比上一财年增加60亿美元（73%）。不包括这两家公司，2025年半导体业总资本支出将比2024年减少120亿美元（10%）。`),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("结果如下：")
	fmt.Println(res)
}
