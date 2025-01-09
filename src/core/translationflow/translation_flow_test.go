package translationflow_test

import (
	"testing"

	"raggo/src/core/translationflow"
)

func TestExecuteTemplatesForTest(t *testing.T) {
	tf := translationflow.NewTranslationFlow(nil)
	data := translationflow.TemplateData{
		SourceLang: "en",
		TargetLang: "zh",
		Country:    "TW",
		SourceText: "Hello World",
	}

	tests := []struct {
		name       string
		systemTmpl string
		promptTmpl string
		wantSystem string
		wantPrompt string
		wantErr    bool
	}{
		{
			name:       "basic template",
			systemTmpl: "Translate from {{.SourceLang}} to {{.TargetLang}}",
			promptTmpl: "Text: {{.SourceText}}",
			wantSystem: "Translate from en to zh",
			wantPrompt: "Text: Hello World",
			wantErr:    false,
		},
		{
			name:       "with country code",
			systemTmpl: "Translate from {{.SourceLang}} to {{.TargetLang}} ({{.Country}})",
			promptTmpl: "Please translate: {{.SourceText}}",
			wantSystem: "Translate from en to zh (TW)",
			wantPrompt: "Please translate: Hello World",
			wantErr:    false,
		},
		{
			name:       "invalid template",
			systemTmpl: "Translate from {{.InvalidField}}",
			promptTmpl: "Text: {{.SourceText}}",
			wantSystem: "",
			wantPrompt: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSystem, gotPrompt, err := tf.ExecuteTemplatesForTest(tt.systemTmpl, tt.promptTmpl, data)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteTemplatesForTest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotSystem != tt.wantSystem {
					t.Errorf("ExecuteTemplatesForTest() system = %v, want %v", gotSystem, tt.wantSystem)
				}
				if gotPrompt != tt.wantPrompt {
					t.Errorf("ExecuteTemplatesForTest() prompt = %v, want %v", gotPrompt, tt.wantPrompt)
				}
			}
		})
	}
}

func TestCalculateChunkSize(t *testing.T) {
	tests := []struct {
		name       string
		tokenCount int
		tokenLimit int
		wantSize   int
	}{
		{
			name:       "below limit",
			tokenCount: 1000,
			tokenLimit: 500,
			wantSize:   500,
		},
		{
			name:       "above limit - case 1",
			tokenCount: 1530,
			tokenLimit: 500,
			wantSize:   389,
		},
		{
			name:       "above limit - case 2",
			tokenCount: 2242,
			tokenLimit: 500,
			wantSize:   496,
		},
		{
			name:       "equal to limit",
			tokenCount: 500,
			tokenLimit: 500,
			wantSize:   500,
		},
		{
			name:       "small numbers",
			tokenCount: 10,
			tokenLimit: 20,
			wantSize:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translationflow.CalculateChunkSize(tt.tokenCount, tt.tokenLimit)
			if got != tt.wantSize {
				t.Errorf("CalculateChunkSize(%d, %d) = %d, want %d",
					tt.tokenCount, tt.tokenLimit, got, tt.wantSize)
			}
		})
	}
}
