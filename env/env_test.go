package env

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func FuncTest(t *testing.T, f interface{}, args...interface{}) string {
	str := ""
	vf := reflect.ValueOf(f)
	tf := vf.Type()
	numOut := tf.NumOut()
	in := make([]reflect.Value, len(args))
	str += fmt.Sprintf("\nTEST FUNC: %s (\n",  runtime.FuncForPC(vf.Pointer()).Name())
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
		str += fmt.Sprintf("    %#v,\n", arg)
	}
	str += fmt.Sprintf(") => ")
	out := vf.Call(in)

	printOutNum := numOut
	if tf.Out(numOut - 1).Name() == "error" {
		printOutNum = numOut - 1
		if err := out[numOut - 1].Interface(); err != nil {
			str += fmt.Sprintf("[Error]: %s\n", err.(error))
			return str
		}
	}
	str += "[\n"
	for i := 0; i < printOutNum; i++ {
		str += fmt.Sprintf("    %#v,\n", out[i])
	}
	str += "]"
	return str
}

//func TestParseEnvVarExp(t *testing.T) {
//	t.Log(FuncTest(t, ParseEnvVarExp, "PATH<sys, user=tiger> += $$PF_ROOT/envs/go/bin"))
//	t.Log(FuncTest(t, ParseEnvVarExp, "PATH<sys> += $$PF_ROOT/envs/go/bin"))
//}

func TestNewEnvConfig(t *testing.T) {
	t.Log(FuncTest(t, NewEnvConfig, "env_config_win.yaml"))
}

func TestInstall(t *testing.T) {
	if cfg, err := NewEnvConfig("env_config_win.yaml"); err == nil {
		t.Log(FuncTest(t, Install, cfg))
	} else {
		t.Log(err)
	}
}