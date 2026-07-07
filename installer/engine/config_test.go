package engine

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

// TestImportKey 验证单 Key 导入:模板逐字段正确、Key 正确填充、文件权限 0600。
func TestImportKey(t *testing.T) {
	home := t.TempDir()
	if err := ImportKey(home, "sk-test-123"); err != nil {
		t.Fatal("ImportKey:", err)
	}

	data, err := os.ReadFile(SettingsPath(home))
	if err != nil {
		t.Fatal("读取 settings.json:", err)
	}
	var got struct {
		Env                               map[string]string `json:"env"`
		Theme                             string            `json:"theme"`
		SkipDangerousModePermissionPrompt bool              `json:"skipDangerousModePermissionPrompt"`
		IncludeCoAuthoredBy               bool              `json:"includeCoAuthoredBy"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal("生成的 settings.json 不是合法 JSON:", err)
	}

	wantEnv := map[string]string{
		"ANTHROPIC_AUTH_TOKEN":               "sk-test-123",
		"ANTHROPIC_BASE_URL":                 "https://api.kimi.com/coding/",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":      "kimi-for-coding",
		"ANTHROPIC_DEFAULT_SONNET_MODEL":     "kimi-for-coding",
		"ANTHROPIC_DEFAULT_OPUS_MODEL":       "kimi-for-coding",
		"ANTHROPIC_MODEL":                    "kimi-for-coding",
		"CLAUDE_CODE_DISABLE_TERMINAL_TITLE": "1",
	}
	for k, want := range wantEnv {
		if got.Env[k] != want {
			t.Errorf("env[%s] = %q, 期望 %q", k, got.Env[k], want)
		}
	}
	if len(got.Env) != len(wantEnv) {
		t.Errorf("env 字段数 = %d, 期望 %d", len(got.Env), len(wantEnv))
	}
	if got.Theme != "auto" || !got.SkipDangerousModePermissionPrompt || got.IncludeCoAuthoredBy {
		t.Errorf("顶层字段与模板不符: %+v", got)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(SettingsPath(home))
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("文件权限 = %o, 期望 0600", perm)
		}
	}
}

// TestImportKeySpecialChars 验证 Key 含引号、反斜杠时 JSON 仍合法且逐字保留。
func TestImportKeySpecialChars(t *testing.T) {
	home := t.TempDir()
	key := `sk-"quote"\back\slash`
	if err := ImportKey(home, key); err != nil {
		t.Fatal("ImportKey:", err)
	}
	data, err := os.ReadFile(SettingsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal("特殊字符破坏了 JSON:", err)
	}
	if got.Env["ANTHROPIC_AUTH_TOKEN"] != key {
		t.Errorf("Key 未逐字保留: %q", got.Env["ANTHROPIC_AUTH_TOKEN"])
	}
}

// TestImportKeyRejectsBadInput 验证空 Key 与含空白字符的 Key 被拒绝。
func TestImportKeyRejectsBadInput(t *testing.T) {
	home := t.TempDir()
	for _, key := range []string{"", "   ", "sk-abc def", "sk-abc\ndef", "sk-abc\tdef"} {
		if err := ImportKey(home, key); err == nil {
			t.Errorf("ImportKey(%q) 应报错", key)
		}
	}
	if _, err := os.Stat(SettingsPath(home)); !os.IsNotExist(err) {
		t.Error("非法输入不应产生 settings.json")
	}
}

// TestImportKeyTrimsWhitespace 验证首尾空白(复制时常见)被自动去除。
func TestImportKeyTrimsWhitespace(t *testing.T) {
	home := t.TempDir()
	if err := ImportKey(home, "  sk-trimmed\n"); err != nil {
		t.Fatal("ImportKey:", err)
	}
	data, _ := os.ReadFile(SettingsPath(home))
	var got struct {
		Env map[string]string `json:"env"`
	}
	json.Unmarshal(data, &got)
	if got.Env["ANTHROPIC_AUTH_TOKEN"] != "sk-trimmed" {
		t.Errorf("首尾空白未去除: %q", got.Env["ANTHROPIC_AUTH_TOKEN"])
	}
}
