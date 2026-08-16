package tickets

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Statistics struct {
	Characters int    `json:"characters"`
	Words      int    `json:"words"`
	Category   string `json:"category"`
}

type Result struct {
	CleanedText string     `json:"cleaned_text"`
	Statistics  Statistics `json:"statistics"`
}

func cleanDescription(cleaner TextCleaner, description string) (Result, error) {
	var err error
	if strings.TrimSpace(description) == "" {
		err = ErrDescriptionEmpty
	}
	var cleaned string
	if cleaner == nil {
		err = ErrCleanerUnavailable
	} else {
		cleaned, err = cleaner.Clean(description)
	}
	if err != nil {
		return Result{}, err
	}
	return Result{
		CleanedText: cleaned,
		Statistics: Statistics{
			Characters: utf8.RuneCountInString(cleaned),
			Words:      countWords(cleaned),
			Category:   classify(cleaned),
		},
	}, nil
}

func countWords(text string) int {
	count := 0
	latinToken := false
	for _, r := range text {
		if unicode.In(r, unicode.Han) {
			count++
			latinToken = false
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !latinToken {
				count++
				latinToken = true
			}
			continue
		}
		latinToken = false
	}
	return count
}

func classify(text string) string {
	text = strings.ToLower(text)
	switch {
	case strings.Contains(text, "退款") || strings.Contains(text, "退货") || strings.Contains(text, "refund"):
		return "退款售后"
	case strings.Contains(text, "物流") || strings.Contains(text, "快递") || strings.Contains(text, "配送") || strings.Contains(text, "delivery"):
		return "物流配送"
	case strings.Contains(text, "质量") || strings.Contains(text, "损坏") || strings.Contains(text, "破损") || strings.Contains(text, "quality"):
		return "商品质量"
	case strings.Contains(text, "账户") || strings.Contains(text, "账号") || strings.Contains(text, "登录") || strings.Contains(text, "account"):
		return "账户问题"
	default:
		return "其他"
	}
}
