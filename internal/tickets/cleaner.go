package tickets

import (
	"errors"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	ErrDescriptionEmpty   = errors.New("描述不能为空")
	ErrCleanerUnavailable = errors.New("文本清理器不可用")
)

type TextCleaner interface {
	Clean(string) (string, error)
}

type LocalTextCleaner struct {
	policy *bluemonday.Policy
}

func NewLocalTextCleaner() *LocalTextCleaner {
	return &LocalTextCleaner{policy: bluemonday.StrictPolicy()}
}

func (c *LocalTextCleaner) Clean(input string) (string, error) {
	if c == nil || c.policy == nil {
		return "", ErrCleanerUnavailable
	}
	cleaned := c.policy.Sanitize(input)
	return strings.TrimSpace(cleaned), nil
}
