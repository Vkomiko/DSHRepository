package env

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

func Install(config *EnvConfig) error {
	sysEnvKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\ControlSet001\Control\Session Manager\Environment`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer sysEnvKey.Close()

	userEnvKey, err := registry.OpenKey(registry.CURRENT_USER,
		`Environment`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer userEnvKey.Close()

	var envKey registry.Key

	for _, evOpt := range config.EnvVarOpts {
		if evOpt.IsSys {
			envKey = sysEnvKey
		} else {
			envKey = userEnvKey
		}
		switch evOpt.Opration {
		case EnvVarSet:
		case EnvVarAppend:
			if oldVal, valTyp, err := envKey.GetStringValue(evOpt.Name); err == nil {
				fmt.Println(oldVal)
				fmt.Println(valTyp)
			}
		}
	}
	return nil
}


