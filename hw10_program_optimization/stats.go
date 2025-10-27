package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

//easyjson:json
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (result users, err error) {
	var i int
	sc := bufio.NewScanner(r) // создание сканера, который будет построчно считывать данные из источника r
	for sc.Scan() {
		var user User
		if err = user.UnmarshalJSON(sc.Bytes()); err != nil {
			return
		}
		result[i] = user
		i++
	}

	return
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)
	dotDomain := "." + strings.ToLower(domain)

	for _, user := range u {
		email := user.Email
		at := strings.LastIndexByte(email, '@')
		if at == -1 {
			continue
		}
		domainPart := email[at+1:]

		lowerDomainPart := strings.ToLower(domainPart)
		if strings.HasSuffix(lowerDomainPart, dotDomain) {
			result[lowerDomainPart]++
		}
	}

	return result, nil
}
