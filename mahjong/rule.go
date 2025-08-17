package mahjong

import (
	"strconv"
	"strings"
)

type Rule struct {
	values []int
}

func NewRule() *Rule {
	return &Rule{
		values: make([]int, 0),
	}
}

func (c *Rule) GetValue(index int) int {
	if int(index) < len(c.values) {
		return c.values[index]
	}
	return 0
}

func (c *Rule) ToString() string {
	var builder strings.Builder
	for i, v := range c.values {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.Itoa(v))
	}
	return builder.String()
}

func (c *Rule) LoadRule(rule string, defalt []int) {
	c.values = make([]int, len(defalt))
	copy(c.values, defalt)

	parts := strings.Split(rule, ",")
	c.values = make([]int, len(parts))

	for i, part := range parts {
		if val, err := strconv.Atoi(strings.TrimSpace(part)); err != nil {
			c.values[i] = val
		}
	}
}
