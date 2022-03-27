package main

import (
	"fmt"

	"github.com/willfantom/executive"
)

type Person struct {
	Name      *Name
	Character string
}

type Name struct {
	First string
	Last  string
}

const title = "Shrek Cast"

var people = []Person{
	{
		Name: &Name{
			First: "Mike",
			Last:  "Myers",
		},
		Character: "Shrek",
	},
	{
		Name: &Name{
			First: "Eddie",
			Last:  "Murphy",
		},
		Character: "Donkey",
	},
	{
		Name: &Name{
			First: "Cameron",
			Last:  "Diaz",
		},
		Character: "Fiona",
	},
}

func main() {

	script, err := executive.NewScriptFromFile("shrek", "./example.sh")
	if err != nil {
		panic(err)
	}
	executable, err := script.WithField("People", people).WithField("Title", title).Compile()
	if err != nil {
		panic(err)
	}
	process, err := executable.WithOSEnv().WithEnv("IMDB", "7.9").Execute()
	if err != nil {
		panic(err)
	}
	result, _ := process.Result()
	fmt.Printf("%+v\n", result)

}
