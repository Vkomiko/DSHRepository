package main

import (
	"path/filepath"
	"fmt"
	"strings"
	"regexp"
)

func main()  {
	s := "file:///z:\\Users\\../tigerjang/://aaa"
	p, _ := filepath.Abs("/C:/Users\\../tigerjang/")
	fmt.Println(p)
	fmt.Println(s[3:])
	sp := strings.SplitN(s, "://", 2)

	var validID = regexp.MustCompile(`^[\\/]\w:([\\/].*?)*`)

	fmt.Println(sp[1])
	fmt.Println(validID.MatchString(sp[1]))
	fmt.Println(filepath.ToSlash(sp[1]))
}