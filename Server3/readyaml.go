package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type TokenID struct {
	TokenId string
	Writer  string
	Reader  []string
}

func ReadYaml() map[string]TokenID {

	yfile, err := ioutil.ReadFile("../TokenInfo.yaml")

	if err != nil {

		log.Fatal(err)
	}
	tokenIdS := make(map[string]TokenID)

	err2 := yaml.Unmarshal(yfile, &tokenIdS)

	if err2 != nil {

		log.Fatal(err2)
	}

	//log.Println(tokenIdS)
	return tokenIdS
}
