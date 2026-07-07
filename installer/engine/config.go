package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SettingsPath 返回 Claude Code 的配置文件路径。
func SettingsPath(home string) string {
	return filepath.Join(home, ".claude", "settings.json")
}

// kimiSettings 返回锻码工坊活动统一的 settings.json 模板,ANTHROPIC_AUTH_TOKEN 填入用户粘贴的 Key。
// Key 通过 JSON 序列化注入(而非字符串拼接),含引号、反斜杠等特殊字符也不会破坏 JSON。
func kimiSettings(key string) map[string]any {
	return map[string]any{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":               key,
			"ANTHROPIC_BASE_URL":                 "https://api.kimi.com/coding/",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":      "kimi-for-coding",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":     "kimi-for-coding",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":       "kimi-for-coding",
			"ANTHROPIC_MODEL":                    "kimi-for-coding",
			"CLAUDE_CODE_DISABLE_TERMINAL_TITLE": "1",
		},
		"theme":                             "auto",
		"skipDangerousModePermissionPrompt": true,
		"includeCoAuthoredBy":               false,
	}
}

// ImportKey 把用户粘贴的 Kimi API Key 注入内置模板后写入 ~/.claude/settings.json。
// 活动主流程只需这一个 Key,参会者全程不接触 JSON;写入复用 ImportSettings
// (0600 权限、Windows 剔除 hooks 的逻辑保持一致)。
func ImportKey(home, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("API Key 不能为空")
	}
	if strings.ContainsAny(key, " \t\r\n") {
		return fmt.Errorf("API Key 不能包含空格或换行,请检查是否粘贴完整")
	}
	out, err := json.Marshal(kimiSettings(key))
	if err != nil {
		return err
	}
	return ImportSettings(home, string(out))
}

// ImportSettings 校验内容为合法 JSON 后,整份写入 ~/.claude/settings.json。
// 用户从别处(团队分发、文档)复制一份 settings.json 直接粘贴导入即可,无需逐项填写。
// Windows 上自动剔除 hooks 字段：hooks 通常调用 /bin/sh，在 Windows 上无法运行。
func ImportSettings(home, content string) error {
	var probe map[string]any
	if err := json.Unmarshal([]byte(content), &probe); err != nil {
		return fmt.Errorf("不是合法的 JSON:%v", err)
	}
	if runtime.GOOS == "windows" {
		delete(probe, "hooks")
	}
	p := SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(probe, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, out, 0o600) // 0600:含 Key,仅本人可读
}

// RemoveSettings 删除 ~/.claude/settings.json(想换配置时先移除再导入)。
func RemoveSettings(home string) error {
	err := os.Remove(SettingsPath(home))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ReadSettings 读取当前配置内容;不存在返回 false。
func ReadSettings(home string) (string, bool) {
	data, err := os.ReadFile(SettingsPath(home))
	if err != nil {
		return "", false
	}
	return string(data), true
}
