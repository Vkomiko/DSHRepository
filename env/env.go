package env

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"errors"
	"io/ioutil"
	"regexp"
	"strings"
	"fmt"
)

type EnvConfig struct {
	Platform string `yaml:"Platform"`
	Macros []string `yaml:"Macros"`
	EnvVars []string `yaml:"EnvVars"`
	EnvVarOpts []*EnvVarOperation
}

func NewEnvConfig(path string) (*EnvConfig, error) {
	conf := new(EnvConfig)
	if b, err := ioutil.ReadFile(path); err == nil {
		if err = yaml.Unmarshal(b, conf); err != nil {
			return nil, err
		}
		conf.EnvVarOpts = make([]*EnvVarOperation, len(conf.EnvVars))
		for i, envExp := range conf.EnvVars {
			if opt, err := ParseEnvVarExp(envExp); err == nil {
				conf.EnvVarOpts[i] = opt
			} else {
				return nil, err
			}
		}
		return conf, nil
	} else {
		return nil, err
	}
}

type VolumeConfig struct {
	Name string `yaml:"VolName"`
}

func GetVolumeConfig(path string) (*VolumeConfig, error) {
	var (
		err error
		b []byte
		p string
	)
	if p = path; p == "" {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}
		p = filepath.VolumeName(ex)
		if p == "" {
			return nil, errors.New("volume config file must be specified")
		}
	} else {
		p = filepath.Dir(p)
	}
	p =  filepath.Join(p, "__VOL_CONFIG__.yaml")
	conf := new(VolumeConfig)

	if b, err = ioutil.ReadFile(p); err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(b, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

type EnvVarOperationEnum int
const (
	EnvVarSet EnvVarOperationEnum = iota
	EnvVarAppend
	EnvVarPrepend
	EnvVarDel
	EnvVarAdd
)

type EnvVarOperation struct {
	Name string
	Value string
	Opration EnvVarOperationEnum
	IsSys bool
}

var (
	envVarReg = regexp.MustCompile(
	`(?P<name>[A-Za-z_]\w*)\s*(?P<options><[^>]*>)*?\s*(?P<operation>=\+|=|\+=|\+\+)\s*(?P<value>\S.*)`)
	envVarOptMap = map [string] EnvVarOperationEnum {
		"=": EnvVarSet,
		"+=": EnvVarAppend,
		"=+": EnvVarPrepend,
		"++": EnvVarAdd,
	}
)


func ParseEnvVarExp(s string) (*EnvVarOperation, error) {
	if m := envVarReg.FindStringSubmatch(s); m != nil {
		envOpt := &EnvVarOperation{Name: m[1], Value: m[4], Opration: envVarOptMap[m[3]]}
		if m[2] != "" {
			for _, _f := range strings.Split(strings.Trim(m[2], "<>"), ",") {
				opt := strings.Split(strings.Trim(_f, " "), "=")
				opt_n := strings.ToLower(strings.Trim(opt[0], " "))
				switch opt_n {
				case "sys":
					envOpt.IsSys = true
					// TODO: More
				default:
					return nil, errors.New(fmt.Sprintf("no such option: %s", opt_n))
				}
			}
		}
		return envOpt, nil
	}
	return nil, errors.New("invalid environment variable operation expression")
}