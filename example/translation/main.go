package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	fengchao "github.com/ijiwei/fengchao-go"
)

const ChunkSize = 4000

const prompt = `
you are a highly skilled translator tasked with translating various types of content from other languages into Chinese. Follow these instructions carefully to complete the translation task:

Input
Depending on the type of input, follow these specific instructions:

If the input is a URL or a request to translate a URL:
First, request the built-in Action to retrieve the URL content. Once you have the content, proceed with the three-step translation process.

If the input is an image or PDF:
Get the content from image (by OCR) or PDF, and proceed with the three-step translation process.

Otherwise, proceed directly to the three-step translation process.
Strategy
You will follow a three-step translation process:

Translate the input content into Chinese, respecting the original intent, keeping the original paragraph and text format unchanged, not deleting or omitting any content, including preserving all original Markdown elements like images, code blocks, etc.

Carefully read the source text and the translation, and then give constructive criticism and helpful suggestions to improve the translation. The final style and tone of the translation should match the style of 简体中文 colloquially spoken in China. When writing suggestions, pay attention to whether there are ways to improve the translation's

(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),

(ii) fluency (by applying Chinese grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),

(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),

(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms Chinese).

Based on the results of steps 1 and 2, refine and polish the translation
Glossary
Here is a glossary of technical terms to use consistently in your translations:

AGI -> 通用人工智能

LLM/Large Language Model -> 大语言模型

Transformer -> Transformer

Token -> Token

Generative AI -> 生成式 AI

AI Agent -> AI 智能体

prompt -> 提示词

zero-shot -> 零样本学习

few-shot -> 少样本学习

multi-modal -> 多模态

fine-tuning -> 微调

Output
For each step of the translation process, output your results within the appropriate XML tags:

<step1_initial_translation>

[Insert your initial translation here]

</step1_initial_translation>

<step2_reflection>

[Insert your reflection on the translation, write a list of specific, helpful and constructive suggestions for improving the translation. Each suggestion should address one specific part of the translation.]

</step2_reflection>

<step3_refined_translation>

[Insert your refined and polished translation here]

</step3_refined_translation>

Remember to consistently use the provided glossary for technical terms throughout your translation. Ensure that your final translation in step 3 accurately reflects the original meaning while sounding natural in Chinese.
`

var client = fengchao.NewFengChao(os.Getenv("FENGCHAO_KEY"), os.Getenv("FENGCHAO_SECRET"), os.Getenv("FENGCHAO_BASE_URL"))

func main() {
	// 定义命令行参数
	fileFlag := flag.String("file", "", "要读取的文件路径")
	textFlag := flag.String("text", "", "要处理的文本内容")

	flag.Parse()

	// 如果提供了文件路径参数
	if *fileFlag != "" {
		err := translateFile(*fileFlag)
		if err != nil {
			panic(err)
		}
	} else if *textFlag != "" {
		transalteText(*textFlag)
	} else {
		fmt.Println("请提供 -file 或 -text 参数")
	}
}

func transalteText(text string) {
	fmt.Println(text)
	fmt.Println(strings.Repeat("-", 20))
	fmt.Println(translation(text))
}

func translateFile(filename string) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0)

	if err != nil {
		return fmt.Errorf("open file %s error: %s", filename, err)
	}
	srcFileStat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file %s error: %s", filename, err)
	}
	defer f.Close()
	// 创建一个缓冲读取器
	reader := bufio.NewReader(f)
	buffer := make([]byte, 0)

	basename := filename[:strings.LastIndex(filename, ".")]
	ext := filename[strings.LastIndex(filename, ".")+1:]

	transaltefile, err := os.OpenFile(fmt.Sprintf("%s-ch.%s", basename, ext), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		return fmt.Errorf("open translate file err: %s", err)
	}

	var completeSize float64
	defer transaltefile.Close()
	completeDisplay(completeSize / float64(srcFileStat.Size()))
	for {
		l, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取文件出错: %w", err)
		}

		buffer = append(buffer, l...)
		if len(buffer) > ChunkSize {
			r := translation(string(buffer))
			_, err := transaltefile.WriteString(r)
			if err != nil {
				return fmt.Errorf("write translate file err: %s", err)
			}
			buffer = buffer[:0]
		}

		// 如果到达文件末尾
		if err == io.EOF {
			r := translation(string(buffer)) + "\n"
			_, err := transaltefile.WriteString(r)
			if err != nil {
				return fmt.Errorf("write translate file err: %s", err)
			}
			completeDisplay(completeSize / float64(srcFileStat.Size()))
			time.Sleep(100 * time.Millisecond)
			break
		}
		completeSize += float64(len(l))
		completeDisplay(completeSize / float64(srcFileStat.Size()))
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\ntranslate file: %s\n", transaltefile.Name())
	return nil
}

func translation(text string) string {
	res, err := client.ChatCompletion(
		context.Background(),
		fengchao.NewPromptTemplate(
			fengchao.NewSystemMessage(prompt),
			fengchao.NewUserMessage("{{.Text}}"),
		),
		fengchao.WithModel("gpt-4o"),
		fengchao.WithParams(map[string]interface{}{
			"Text": text,
		}),
	)

	if err != nil {
		panic(err)
	}

	return outputParse(res.String())
}

func outputParse(output string) string {

	re := regexp.MustCompile(`(?s)<step3_refined_translation>(.*?)<\/step3_refined_translation>`)
	translateResults := re.FindStringSubmatch(output)
	if len(translateResults) > 1 {
		// 使用 translateResults[1] 来获取捕获组的内容
		output = strings.Trim(translateResults[1], "\n")
	}
	if strings.Contains(output, "<step3_refined_translation>") {
		output = output[strings.Index(output, "<step3_refined_translation>"):]
	}

	if strings.Contains(output, "</step3_refined_translation>") {
		output = output[:strings.Index(output, "</step3_refined_translation>")]
	}

	return output
}

func completeDisplay(completed float64) {
	line := fmt.Sprintf("|%s%s|", strings.Repeat("=", int(completed*40)), strings.Repeat("-", 40-int(completed*40)))
	fmt.Printf("\r%s %.2f%%", line, completed*100)
}
