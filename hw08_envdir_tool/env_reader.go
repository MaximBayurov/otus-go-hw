package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var ErrIncorrectDirPath = errors.New("incorrect dir path")

type Environment map[string]EnvValue

// toSlice возвращает Environment преобразованный к слайсу.
func (e Environment) toSlice() []string {
	result := make([]string, 0, len(e))
	for name, envValue := range e {
		result = append(result, strings.Join([]string{name, envValue.Value}, "="))
	}
	return result
}

// createEnvFromSlice создаёт окружение из слайса.
func createEnvFromSlice(unparsed []string) Environment {
	env := Environment{}
	for _, item := range unparsed {
		pieces := strings.Split(item, "=")
		if len(pieces) != 2 {
			continue
		}
		envVal := EnvValue{
			Value:      pieces[1],
			NeedRemove: false,
		}
		env[pieces[0]] = envVal
	}
	return env
}

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	if !fileInfo.IsDir() {
		return nil, ErrIncorrectDirPath
	}
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	env := make(Environment, len(dirEntries))
	for _, dirEntry := range dirEntries {
		if !dirEntry.Type().IsRegular() {
			continue
		}
		if strings.Contains(dirEntry.Name(), "=") {
			continue
		}

		envValue, err := makeEnvValueFor(dir + "/" + dirEntry.Name())
		if err != nil {
			continue
		}
		env[dirEntry.Name()] = *envValue
	}
	return env, nil
}

// makeEnvValueFor создаёт значение переменной окружения по переданному пути файла и возвращает ссылку на него.
func makeEnvValueFor(path string) (*EnvValue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer closeFile(file)

	envVal := &EnvValue{
		Value:      "",
		NeedRemove: false,
	}

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		envVal.Value = strings.TrimRight(scanner.Text(), " \t")
		envVal.Value = strings.ReplaceAll(envVal.Value, string([]byte{0x00}), "\n")
	} else {
		envVal.NeedRemove = true
	}

	return envVal, nil
}

// closeFile закрывает файл и логирует ошибку в случае её возникновения.
func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Println(err)
	}
}
