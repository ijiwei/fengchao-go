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

### 🏗
