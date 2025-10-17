package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	command := makeCommandFrom(cmd)
	setEnvFor(command, env)

	err := command.Run()
	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		log.Println(err)
	}

	return 0
}

// makeCommandFrom создаёт команду из переданных параметров.
func makeCommandFrom(cmd []string) *exec.Cmd {
	name := cmd[0]
	arg := cmd[1:]
	command := exec.CommandContext(context.TODO(), name, arg...)

	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr

	return command
}

// setEnvFor устанавливает переменные окружения для команды.
func setEnvFor(command *exec.Cmd, env Environment) {
	currentEnv := createEnvFromSlice(os.Environ())

	for name, envValue := range env {
		if envValue.NeedRemove {
			_, exists := currentEnv[name]
			if exists {
				delete(currentEnv, name)
			}
		}
		currentEnv[name] = envValue
	}

	command.Env = currentEnv.toSlice()
}
