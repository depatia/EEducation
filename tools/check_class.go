package tools

import (
	"strconv"
)

var classes = make(map[string]bool)

func SetClasses() {
	var i rune
	for j := 1; j <= 11; j++ {
		for i = 'А'; i <= 'Я'; i++ {
			switch i {
			case 'Ь':
				continue
			case 'Ъ':
				continue
			case 'Ы':
				continue
			case 'Й':
				continue
			}
			letter, _ := strconv.Unquote(strconv.QuoteRune(i))
			classes[strconv.Itoa(j)+letter] = true
		}
	}
}

func CheckClass(classname string) bool {
	if _, ok := classes[classname]; ok {
		return true
	}
	return false
}
