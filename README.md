# Fengchao

FengChao Golang SDK

## Chat Prompt

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

### Prompt

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
