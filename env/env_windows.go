package env

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"runtime"
	"strings"
)

const EnvListSep = ";"

type EnvVar struct {
	name string
	value string
	//isList bool
	valeTyp uint32
}

type Environment struct {
	entry registry.Key
	key registry.Key
	path string
	flag uint32
	isOpen bool
}

var entryMap = map [string] registry.Key {
	"HKEY_LOCAL_MACHINE": registry.LOCAL_MACHINE,
	"HKEY_CURRENT_USER": registry.CURRENT_USER,
}

func getEnvRefString(s string) string {
	return "%" + s +"%"
}

func Env(uri string) *Environment {
	s := strings.SplitN(uri, "\\", 2)
	e := &Environment{entry: entryMap[s[0]], path: s[1]}
	runtime.SetFinalizer(e, closeEnv)
	return e
}

func (e *Environment) Open(flag uint32) error {
	if e.isOpen {
		if (e.flag | flag) == e.flag {
			return nil
		}
		fmt.Println("@@@@", e.flag, flag)
		e.Close()
	}
	if key, err := registry.OpenKey(e.entry, e.path, flag); err == nil {
		e.isOpen = true
		e.key = key
		e.flag = flag
		return nil
	} else {
		return err
	}
}

func (e *Environment) OpenRead() error {
	return e.Open(registry.READ)
}

func (e *Environment) OpenEdit() error {
	return e.Open(registry.SET_VALUE | registry.READ)
}

func closeEnv(e *Environment) {
	if e.isOpen {
		e.key.Close()
		e.isOpen = false
		e.flag = 0
	}
}

func (e *Environment) Close() {
	closeEnv(e)
}

func (e *Environment) Get(name string) (*EnvVar, error) {
	e.OpenRead()
	val, valeTyp, err := e.key.GetStringValue(name)
	if err != nil {
		return nil, err
	}
	return &EnvVar{name: name, value: val, valeTyp: valeTyp}, nil
}

func (e *Environment) Do(op *EnvVarOperation) error {
	e.OpenEdit()
	if envVal, err := e.Get(op.Name); err == nil {
		switch op.Opration {
		case EnvVarSet:
			envVal.Set(op.Value)
		case EnvVarAppend:
			envVal.Append(op.Value)
		case EnvVarPrepend:
			envVal.Prepend(op.Value)
		case EnvVarRemove:
			envVal.Remove(op.Value)
		default:
			panic("Not Implement")
		}
		fmt.Printf("%#v\n", envVal)
		return nil
	} else {
		return  err
	}
}

var (
	SysEnv = Env(`HKEY_LOCAL_MACHINE\SYSTEM\ControlSet001\Control\Session Manager\Environment`)
	CurrentUserEnv = Env(`HKEY_CURRENT_USER\Environment`)
)

func Install(config *EnvConfig) error {
	var envKey *Environment

	for _, evOpt := range config.EnvVarOpts {
		if evOpt.IsSys {
			envKey = SysEnv
		} else {
			envKey = CurrentUserEnv
		}
		envKey.Do(evOpt)
	}
	return nil
}


