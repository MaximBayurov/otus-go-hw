package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(reader io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	rd := bufio.NewReader(reader)
	var user User
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}
		if err = json.Unmarshal(line, &user); err != nil {
			return nil, err
		}

		if strings.HasSuffix(user.Email, "."+domain) {
			user.Email = strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
			if len(user.Email) > 0 {
				_, ok := result[user.Email]
				if ok {
					result[user.Email]++
				} else {
					result[user.Email] = 1
				}
			}
		}
	}
	return result, nil
}
