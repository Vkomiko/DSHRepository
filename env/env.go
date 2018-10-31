package env

import (
	"crypto/md5"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)


func (v *EnvVar) List() []string {
	s := strings.Split(v.value, EnvListSep)
	nei := 0
	for _, e := range s {
		if strings.Trim(e, " ") != "" {
			s[nei] = e
			nei ++
		}
	}
	return s[: nei]
}

func (v *EnvVar) Set(ele string) {
	v.value = ele
}

func (v *EnvVar) Append(ele string) {
	v.value = strings.Join(v.List(), EnvListSep) + EnvListSep + ele
}

func (v *EnvVar) Prepend(ele string) {
	v.value = ele + EnvListSep + strings.Join(v.List(), EnvListSep)
}

func (v *EnvVar) Remove(ele string) string {
	count := 0
	l := v.List()
	for _, e := range l {
		if strings.Trim(e, " ") != ele {
			l[count] = e
			count ++
		}
	}
	return strings.Join(l[: count], EnvListSep)
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
	EnvVarAdd
	EnvVarDel
	EnvVarRemove
)

type EnvVarOperation struct {
	Name string
	Value string
	Opration EnvVarOperationEnum
	IsSys bool
	NoPath bool
}

var (
	envVarReg = regexp.MustCompile(
	`(?P<name>[A-Za-z_]\w*)\s*(?P<options><[^>]*>)?\s*(?P<operation>=\+|=|\+=|\+\+)\s*(?P<value>\S.*)`)
	envTmpDefReg = regexp.MustCompile(`(?P<name>[A-Za-z_]\w*)\s*=\s*(?P<value>\S.*)`)
	envVarOptMap = map [string] EnvVarOperationEnum {
		"=": EnvVarSet,
		"+=": EnvVarAppend,
		"=+": EnvVarPrepend,
		"++": EnvVarAdd,
	}
	tmpDefRefReg = regexp.MustCompile(`\$\$(\w+|{\w+})`)
	envRefReg = regexp.MustCompile(`\$(\w+|{\w+})`)
)

type EnvConfig struct {
	EnvName string `yaml:"EnvName"`
	Platform string `yaml:"Platform"`
	EnvRoot string `yaml:"EnvRoot"`
	TempDefs []string `yaml:"TempDef"`
	EnvVars []string `yaml:"EnvVars"`
	EnvVarOpts []*EnvVarOperation
	tmpDefs map[string]string
	cfgPath string
}

func NewEnvConfig(path string) (*EnvConfig, error) {
	conf := new(EnvConfig)
	if cfgPath, err := filepath.Abs(path); err == nil {
		conf.cfgPath = cfgPath
	} else {return nil, err}
	if b, err := ioutil.ReadFile(path); err == nil {
		if err = yaml.Unmarshal(b, conf); err != nil {
			return nil, err
		}
		if err := conf.init(); err == nil {
			return conf, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (c *EnvConfig) init() error {
	if err := c.initBuildInDef(); err != nil {
		return err
	}
	if err := c.parseTmpDefExps(); err != nil {
		return err
	}
	c.EnvVarOpts = make([]*EnvVarOperation, len(c.EnvVars))
	for i, envExp := range c.EnvVars {
		if opt, err := c.parseEnvVarExp(envExp); err == nil {
			c.EnvVarOpts[i] = opt
		} else {
			return err
		}
	}
	return nil
}

func (c *EnvConfig) initBuildInDef() error {
	c.tmpDefs = make(map[string]string)
	//switch c.EnvRoot {
	//case "ENV_CFG_VOLUME":
	//	c.tmpDefs["ENV_ROOT"] = filepath.VolumeName(c.cfgPath)
	//default:
	//	return errors.New("invalid EnvRoot")
	//}
	return nil
}

func (c *EnvConfig) getEnvUID() string {
	sum := md5.Sum(([]byte)(c.EnvName))
	d := md5.New()
	d.hex ????
}

func (c *EnvConfig) parseTmpDefExps() (err error) {
	err = nil
	var val string
	for _, defExp := range c.TempDefs {
		if m := envTmpDefReg.FindStringSubmatch(defExp); m != nil {
			val, err = c.parseVarVal(m[2], false)
			if err != nil {
				return
			}
			c.tmpDefs[m[1]] = val
		} else {
			err = errors.New("invalid temp define expression")
		}
	}
	return
}

func (c *EnvConfig) parseVarVal(oldVal string, noPath bool) (val string, err error) {
	val = oldVal
	val = tmpDefRefReg.ReplaceAllStringFunc(val, func(s string) string {
		defName := strings.Trim(s, "${} ")
		if defVal, ok := c.tmpDefs[defName]; ok {
			return defVal
		} else {
			err = errors.New(fmt.Sprintf("no such variable: %s", defName))
			return ""
		}
	})
	if err != nil {
		val = ""
		return
	}
	val = envRefReg.ReplaceAllStringFunc(val, func(s string) string {
		name := strings.Trim(s, "${} ")
		return getEnvRefString(name)
	})

	if !noPath {
		val = filepath.FromSlash(val)
	}
	return
}

func (c *EnvConfig) parseEnvVarExp(s string) (*EnvVarOperation, error) {
	if m := envVarReg.FindStringSubmatch(s); m != nil {
		envOpt := &EnvVarOperation{Name: m[1], Opration: envVarOptMap[m[3]],
			IsSys:false, NoPath:false}
		if m[2] != "" {
			for _, _f := range strings.Split(strings.Trim(m[2], "<>"), ",") {
				opt := strings.Split(strings.Trim(_f, " "), "=")
				opt_n := strings.ToLower(strings.Trim(opt[0], " "))
				switch opt_n {
				case "sys":
					envOpt.IsSys = true
				case "noPath":
					envOpt.NoPath = true
					// TODO: More
				default:
					return nil, errors.New(fmt.Sprintf("no such option: %s", opt_n))
				}
			}
		}
		var err error
		if envOpt.Value, err = c.parseVarVal(m[4], envOpt.NoPath); err != nil {
			return nil, err
		}
		return envOpt, nil
	}
	return nil, errors.New("invalid environment variable operation expression")
}