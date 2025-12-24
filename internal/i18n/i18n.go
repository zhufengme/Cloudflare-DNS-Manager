package i18n

import (
	"embed"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	Bundle    *i18n.Bundle
	Localizer map[string]*i18n.Localizer
)

// Init 初始化 i18n
func Init(localeFS embed.FS) error {
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 加载翻译文件
	languages := []string{"en", "zh"}
	for _, lang := range languages {
		if _, err := Bundle.LoadMessageFileFS(localeFS, "web/locales/"+lang+".yaml"); err != nil {
			return err
		}
	}

	// 创建 Localizer
	Localizer = make(map[string]*i18n.Localizer)
	Localizer["en"] = i18n.NewLocalizer(Bundle, "en")
	Localizer["zh"] = i18n.NewLocalizer(Bundle, "zh", "zh-CN")

	return nil
}

// T 翻译函数
func T(localizer *i18n.Localizer, messageID string) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		return messageID // 返回原始 ID
	}
	return msg
}

// GetLocalizer 根据语言代码获取 Localizer
func GetLocalizer(lang string) *i18n.Localizer {
	if localizer, ok := Localizer[lang]; ok {
		return localizer
	}
	return Localizer["en"] // 默认英文
}
