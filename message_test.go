package fengchaogo

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"text/template"
)

func TestMessage_execute(t *testing.T) {
	type fields struct {
		Role     string
		Content  string
		Template *template.Template
		buffer   *bytes.Buffer
	}
	type args struct {
		vairables map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "normal",
			fields: fields{
				Role:     "user",
				Content:  "hello",
				Template: template.Must(template.New("").Parse("hello {{.Name}}")),
				buffer:   &bytes.Buffer{},
			},
			args: args{
				vairables: map[string]interface{}{
					"Name": "world",
				},
			},
			want: "hello world",
		},
		{
			name: "without template",
			fields: fields{
				Role:     "user",
				Content:  "hello",
				Template: template.Must(template.New("").Parse("hello wwwww")),
				buffer:   &bytes.Buffer{},
			},
			args: args{
				vairables: map[string]interface{}{
					"Name": "world",
				},
			},
			want: "hello wwwww",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Role:     tt.fields.Role,
				Content:  tt.fields.Content,
				template: tt.fields.Template,
				buffer:   tt.fields.buffer,
			}
			m.execute(tt.args.vairables)

			if got := m.Content; got != tt.want {
				t.Errorf("Message.execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMessage(t *testing.T) {
	type args struct {
		role       string
		messageStr string
	}
	tests := []struct {
		name    string
		args    args
		want    *Message
		wantErr bool
	}{
		{
			"normal",
			args{
				role:       "user",
				messageStr: "hello {{.Name}}",
			},
			&Message{
				Role:     "user",
				Content:  "",
				template: template.Must(template.New("").Parse("hello {{.Name}}")),
				buffer:   nil,
			},
			false,
		},
		{
			"error template",
			args{
				role:       "user",
				messageStr: "hello {{.Name",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := NewMessage(tt.args.role, tt.args.messageStr)(); !reflect.DeepEqual(got, tt.want) || (err != nil) != tt.wantErr {
				t.Errorf("NewMessageLazy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderLazyMessage(t *testing.T) {
	type args struct {
		role       string
		messageStr string
		vairables  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"normal",
			args{
				role:       "user",
				messageStr: "hello {{.Name}}",
				vairables: map[string]interface{}{
					"Name": "fengchao",
				},
			},
			[]byte(`{"role":"user","content":"hello fengchao"}`),
			false,
		},
		{
			"error template",
			args{
				role:       "user",
				messageStr: "hello {{.Name",
				vairables: map[string]interface{}{
					"Name": "fengchao",
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lazyCreate := NewMessage(tt.args.role, tt.args.messageStr)

			got, err := lazyCreate.Render(tt.args.vairables)

			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != string(tt.want) {
				fmt.Println(string(got))
				fmt.Println(string(tt.want))
				t.Errorf("Render() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
