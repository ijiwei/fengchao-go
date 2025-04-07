# Fengchao

FengChao Golang SDK

## Quick Start

这是一个通过Prompt模板,来生成文章的例子，在使用时我们可以通过With方法来传递请求的多种参数，实现模板渲染，超时检测，模型的参数控制，最后可以直接输出生成内容的文本。

```go

func ChatWithHistory() {
    client := fengchao.NewFengChao(ApiKey, ApiSecret, BaseUrl)
    client.SetLogger(logrus.StandardLogger())

    res, err := client.ChatCompletion(
        context.Background(),
        fengchao.NewPromptTemplate(
            fengchao.NewMessage(fengchao.RoleSystem, systemPrompt),
            fengchao.NewMessage(fengchao.RoleUser, `
本次工作为根据已确认的选题《{{.title}}》，和参考内容完成一篇关于[{{.tags}}]的文章，文章中应避免使用总结、结论等类似的段落。
你要清楚，文章内容将会直接发表到新闻媒体中，稿件的阅读量会直接决定你的绩效考核成绩，请严格按照工作规范来完成，这将会影响你的职业生涯。
以下为本次选题的相关参考内容：
{{.text}}`),
        ),
        fengchao.WithParams(struct {
            title string
            text  string
            tags  string
        }{
            title: `国产AI增强操作系统发布：填补端侧推理空白`,
            text: `8月8日举行的2024中国操作系统产业大会上，国产桌面操作系统银河麒麟发布首个AIPC版本，这是一款与人工智能融合的国产桌面操作系统，填补了我国操作系统端侧推理能力研发的空白。
操作系统是计算机之魂，承接上层软件生态与底层硬件资源，为AI算法、模型与应用的运行提供支撑环境，在IT国产化中发挥重要作用。过去很长一段时间，全球操作系统厂商主要为欧美企业。
我国操作系统发展起步晚、系统生态存在短板，赶超压力大。新一轮人工智能技术的迅猛发展，为我国操作系统带来新机遇。`,
            tags: `#AI操作系统#国产操作系统#端侧推理`,
        }),
        fengchao.WithTemperature(0.9),
        fengchao.WithModel("gpt-4o"),
    )

    if err != nil {
    panic(err)
    }

    fmt.Println("结果如下：")
    fmt.Println(res)
}
```

如果需要更复杂的使用，请参考下面模块的文档

## Prompt

`Prompt` 在`LLM`的使用中是一个很重要的概念，为了简化用户手动构建Prompt。本项目提供了快速的创造`Prompt`的工具。

### Message

Message是Prompt的基础，一个`Prompt`往往由多个`MessageTemple`组成。
目前我们提供两种创建`Message`的方式， 并提供Render方法提供携带变量的渲染，并获取格式化的Json数据。

#### 使用用`Message`模版创建消息并渲染

```go

import (
    "fmt"
    fengchao "github.com/ijiwei/fengchao-go"
)

func main() {
    UserMessage := fengchao.NewMessage(fengchao.RoleUser, `讲一个关于{{.name}}的笑话吧`)
    MessageJson, err := UserMessage.Render(map[string]interface{}{"name": "小狗"})
    if err != nil {
        panic(err)
    }
    fmt.Println(string(MessageJson)) 
    // output: {"role":"user","content":"讲一个关于小狗的笑话吧"}
}



```

模板渲染采用的是`text/template`，所以你可以使用任何其允许的编写方式来设计你的`Prompt`，甚至是编写循环

```go
func main() {
    UserMessage := fengchao.NewMessage(fengchao.RoleUser, `分别讲一个关于{{range .Items}}、{{.}}{{end}}的笑话吧`)
    MessageJson, err :=UserMessage.Render(map[string]interface{}{"Items": []string{"小猫", "小狗", "小狐狸"}})
    if err != nil {
        panic(err)
    }
    fmt.Println(string(MessageJson)) 
    // output: {"role":"user","content":"分别讲一个关于、小猫、小狗、小狐狸的笑话吧"}
}
```

#### 手动创建的`Message`

```go

import fengchao "github.com/ijiwei/fengchao-go"

func main() {
    message := &fengchao.Message{
        Role: fengchao.RoleUser
        Content: "讲一个笑话吧"
    }
}
```

当然手动创建的`Message`同样也可以进行`Render`:

```go

import fengchao "github.com/ijiwei/fengchao-go"

func main() {
    message := &fengchao.Message{
        Role: fengchao.RoleUser
        Content: "讲一个笑话吧"
    }
    MessageJson, err := message.Render(nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(MessageJson)) 
    // output: {"role":"user","content":"讲一个笑话吧"}
}

```

### Template

熟悉了`Message`创建之后，就可以创造第一个`Prompt`了

```go


import fengchao "github.com/ijiwei/fengchao-go"

func main() {
    prompt := fengchao.NewPromptTemplate(
        fengchao.NewMessage(fengchao.RoleSystem, `你是一个非常厉害的{{.Name}}！`),
        fengchao.NewMessage(fengchao.RoleUser, `分别讲一个关于{{range .Items}}、{{.}}{{end}}的笑话吧`),
    )
    prompt.HumanFriendly = true
    PromptJson, err := prompt.Render(map[string]interface{}{
        "Items": []string{"小猫", "小狗", "小狐狸"},
        "Name": "智能助手",
    })

    if err != nil {
        panic(err)
    }
    fmt.Println(string(PromptJson)) 
}

```

output:

```json
[
  {
    "role": "system",
    "content": "你是一个非常厉害的智能助手！"
  },
  {
    "role": "user",
    "content": "分别讲一个关于、小猫、小狗、小狐狸的笑话吧"
  }
]
```

Prompt也可以嵌套使用：

```go

import fengchao "github.com/ijiwei/fengchao-go"

func main() {

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

```

output:

```json
[
  {
    "role": "system",
    "content": "你是一个非常厉害的智能助手！"
  },
  {
    "role": "user",
    "content": "分别讲一个关于、小猫、小狗、小狐狸的笑话吧"
  },
  {
    "role": "assistant",
    "content": "小猫：小猫去银行，工作人员问：“你要存什么？”小猫眨眨眼说：“我存爪印！”\n小狗：小狗学会了打字，但每次发的都是“汪汪汪”，它说：“我这不是在聊天，是在打码！”\n小狐狸：小狐狸问妈妈：“为什么我们叫狡猾？”妈妈笑着说：“因为我们知道怎么用优惠券！”"
  },
  {
    "role": "user",
    "content": "太好笑了😂，再讲3个好不好？"
  }
]
```

## Chat

## Chat Completion

在进行对话前，首先需要获取到一对API KEY和Secret，以及Fengchao服务的Url

### 同步Chat

`ChatCopletion`方法会在`API`完成响应的返回`ChatCopletion`对象, 可以获取对话相关的信息，也可以直接打印（已经实现了`String()`方法）。

```go

func SimpleChat() {
    apiKey := "you api key"
    apiSecret := "you api secret"
    client := fengchao.NewFengChao(apiKey, apiSecret, "http://fengchao.api")
    res, err := client.ChatCompletion(
        context.Background(),
        fengchao.NewMessage(fengchao.RoleUser, "讲一个冷笑话"),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("结果如下：")
    fmt.Println(res)
}

```

也可以使用复杂的`Prompt Template`来进行生成

```go
func SimpleChat() {
    apiKey := "you api key"
    apiSecret := "you api secret"
    client := fengchao.NewFengChao(apiKey, apiSecret, "http://fengchao.api")
    res, err := client.ChatCompletion(
        context.Background(),
        fengchao.NewMessage(fengchao.RoleUser, "讲一个{{.Story}}"),
        fengchao.WithParams(map[string]string{
            "Story": "鬼故事",
        }),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("结果如下：")
    fmt.Println(res)
}
```

除此之外，我们也可通过其`History`属性，获取对话的列表数据

```go
func SimpleChat() {
    apiKey := "you api key"
    apiSecret := "you api secret"
    client := fengchao.NewFengChao(apiKey, apiSecret, "http://fengchao.api")
    res, err := client.ChatCompletion(
        context.Background(),
        fengchao.NewMessage(fengchao.RoleUser, "讲一个{{.Story}}"),
        fengchao.WithParams(map[string]string{
            "Story": "鬼故事",
        }),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("结果如下：")
    fmt.Println(res)
    historyData, err := json.MarshalIndent(res.History, "", "   ")
    if err != nil {
        panic(fmt.Sprintf("marshal history error: %v", err))
    }

    fmt.Println("对话记录如下：")
    fmt.Println(string(historyData))
}
```

当我们想使用对话记录快速构建`Prompt`的时候，对话记录也提供了一个构建`PromptTemplate`的方法

```go
func SimpleChat() {
    apiKey := "you api key"
    apiSecret := "you api secret"
    client := fengchao.NewFengChao(apiKey, apiSecret, "http://fengchao.api")
    res, err := client.ChatCompletion(
        context.Background(),
        fengchao.NewMessage(fengchao.RoleUser, "讲一个{{.Story}}"),
        fengchao.WithParams(map[string]string{
            "Story": "鬼故事",
        }),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("结果如下：")
    fmt.Println(res)

    promptTemplate := fengchao.NewPromptTemplate(
        res.GetHistoryPrompts(),
        fengchao.NewMessage(fengchao.RoleUser, `根据文章内容，总结一份{{.language}}摘要`),
    )
}
```

我们可以使用它来继续构建下一次的对话

```go
func SimpleChat() {
    apiKey := "you api key"
    apiSecret := "you api secret"
    client := fengchao.NewFengChao(apiKey, apiSecret, "http://fengchao.api")
    res, err := client.ChatCompletion(
        context.Background(),
        fengchao.NewMessage(fengchao.RoleUser, "讲一个{{.Story}}"),
        fengchao.WithParams(map[string]string{
            "Story": "鬼故事",
        }),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("结果如下：")
    fmt.Println(res)

    prompt := fengchao.NewPromptTemplate(
        res.GetHistoryPrompts(),
        fengchao.NewMessage(fengchao.RoleUser, `根据文章内容，总结一份{{.language}}摘要`),
    )

    res, err = client.ChatCompletion(
        ctx,
        prompt,
        fengchao.WithTemperature(0.9),
        fengchao.WithModel("glm-4"),
        fengchao.WithParams(map[string]interface{}{"language": "英文"})
    )

    if err != nil {
        panic(err)
    }

    fmt.Println("结果如下：")
    fmt.Println(res)
}
```

### 快速生成

使用预定义的模板`prompt` 进行快速的文本生成

⚠️ 这种方式必须手动指定`Query`和`PredefinedPrompt`

```go

func QuickChatCompletion() {
    client.SetDebug(true)
    res, err := client.QuickCompletion(
        context.Background(),
        fengchao.WithPredefinedPrompts("多译英"),
        fengchao.WithQuery(`命运之轮象征着命运的起伏和变化，它代表着生活中不可预测的转变和机遇。这张牌可能意味着你正处在一个重要的转折点，你将会经历一些意想不到的改变。这些改变可能会带来新的机会和挑战，需要你灵活适应并做好准备。
命运之轮也提醒我们，生活中的好运和不幸都是暂时的，一切都在不断变化中。这张牌鼓励你保持乐观和开放的态度，相信未来会带来更好的机会和成长。同时，也要学会珍惜当下，充分利用现有的资源和机会。`),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("结果如下：")
    fmt.Println(res)
}

```

### 批量生成

`BatchChatCompletionBuilder`提供了一种，批量进行请求的方法，可以实现并发的请求，等到所有请求都结束后同步返回
⚠️ 单次批量的调用，最多添加5个

```go

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

```

### 流式请求

我们可以通过`ChatCompletionStream` 方法，来获取一个`StreamReader`, 然后手动处理数据包

`StreamReader`的`Read`方法，返回三个参数，分别为数据包，是否完成，和错误，可以自行处理其逻辑

流式请求因为要接管`response`, 所以超时时间属性不会起作用，如果需要控制超时时间，需要在外部通过上下文手动控制，内部已经实现上下文的取消策略

```go
func ReadStream() {

    // client.SetDebug(true)

    ctx := context.Background()

    prompt := fengchao.NewPromptTemplate(
        fengchao.NewMessage(fengchao.RoleSystem, `你是一个非常厉害的{{.Name}}！`),
        fengchao.NewMessage(fengchao.RoleUser, `分别讲一个关于{{range .Items}}、{{.}}{{end}}的笑话吧`),
        fengchao.NewMessage(fengchao.RoleAssistant, `小猫：小猫去银行，工作人员问：“你要存什么？”小猫眨眨眼说：“我存爪印！”
小狗：小狗学会了打字，但每次发的都是“汪汪汪”，它说：“我这不是在聊天，是在打码！”
小狐狸：小狐狸问妈妈：“为什么我们叫狡猾？”妈妈笑着说：“因为我们知道怎么用优惠券！”`),
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
            fmt.Println("结束了")
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
        // time.Sleep(time.Millisecond * 100)
    }
    fmt.Print("\n")
    res.Close()
}

```

当然我们也提供了一个更简单的方式来进行处理流式的请求，我们提供一个go1.23.0的新特性的生成器函数

```go
func Stream() {
    defer func() {
        if err := recover(); err != nil {
            fmt.Println("recover error: ", err)
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
        // fengchao.WithIsSensitive(true),
    )

    if err != nil {
        panic("stream error: " + err.Error())
    }

    fmt.Println("结果如下：")
    for r := range res.Stream() {
        if r.Error != nil {
            fmt.Print(r.Error)
        }
        fmt.Print((r).String())
    }
    fmt.Print("\n")
}

```

## 支持历史记录的聊天对话示例

```go

import (
    "bufio"
    "context"
    "fmt"
    "os"

    fengchao "github.com/ijiwei/fengchao-go"
)

var client = fengchao.NewFengChao(os.Getenv("FENGCHAO_KEY"), os.Getenv("FENGCHAO_SECRET"), os.Getenv("FENGCHAO_BASE_URL"))

var systemMessage = fengchao.NewMessage(fengchao.RoleSystem, `你是一名善于理解问题的助手，你要按照以下的规则与用户对话:
1. 采用风趣幽默的回答，适当添加Emoji来让回答更加形象
2. 回答的内容尽可能丰富，如果篇幅过长，你可以先对问题进行总结并生成大纲，并通过多次对话的方式分步进行回答
3. 你的回答要有主观性，不要拿用户的意见和建议作为依据
`)

func ChatBox() {
    fmt.Println("FENGCHAO-CHATBOX")
    fmt.Println("---------------------")
    fmt.Print("> ")
    s := bufio.NewScanner(os.Stdin)
    var historyMessage *fengchao.PromptTemplate
    for s.Scan() {
        input := s.Text()

        switch input {
        case ":clear":
            historyMessage = nil
            fmt.Println("已清除历史消息\n>")
            continue
        case ":exit":
            return
        case ":history":
            historyDisplay(historyMessage)
            continue
        case "":
            continue
        }

        inputMessage := fengchao.NewMessage(fengchao.RoleUser, input)
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

        historyMessage = fengchao.NewPromptTemplate(
            historyMessage,
            inputMessage,
            fengchao.NewMessage(fengchao.RoleAssistant, answer),
        )
        fmt.Print("\n> ")
    }
}

func historyDisplay(history *fengchao.PromptTemplate) {
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
```
