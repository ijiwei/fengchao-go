package fengchaogo

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPromptTemplate_Render(t *testing.T) {
	type fields struct {
		Messages []lazyMessage
	}
	type args struct {
		vairables map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"normal",
			fields{
				Messages: []lazyMessage{
					NewMessage("system", "you are a helpful assistant"),
					NewMessage("user", "hello {{.Name}}"),
				},
			},
			args{
				vairables: map[string]interface{}{
					"Name": "fengchao",
				},
			},
			[]byte(`[{"role":"system","content":"you are a helpful assistant"},{"role":"user","content":"hello fengchao"}]`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个 []Speaker 的切片
			p := make([]Prompt, len(tt.fields.Messages))
			for i, m := range tt.fields.Messages {
				p[i] = m
			}
			m := NewPromptTemplate(p...)
			got, err := m.Render(tt.args.vairables)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptTemplate.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				fmt.Println(string(got))
				fmt.Println(string(tt.want))
				t.Errorf("PromptTemplate.Render() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
