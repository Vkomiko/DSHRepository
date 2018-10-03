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

	envVarReg := regexp.MustCompile(
		`(?P<name>[A-Za-z_]\w*)\s*(?P<options><[^>]*>)*?\s*(?P<operation>=\+|=|\+=|\+\+)\s*(?P<value>\S.*)`)
	m := envVarReg.FindStringSubmatch(`PATH<sys, user=tiger> += $$PF_ROOT/envs/go/bin`)
	fmt.Printf("%#v\n", m)
	fmt.Printf("%#v\n", envVarReg.SubexpNames())

	for _, _f := range strings.Split(strings.Trim(m[2], "<>"), ",") {
		aaa := strings.Split(strings.Trim(_f, " "), "=")

		fmt.Printf("%#v\n", aaa)
	}

}