package main

import (
	"fmt"
)

type User struct {
	Name string
	Age  int
}

func main() {
	fmt.Println("Hello, World!")
}

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func processUser(u *User) error {
	if u == nil {
		return fmt.Errorf("user cannot be nil")
	}
	return nil
}
