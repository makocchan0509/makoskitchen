package util

import (
	"fmt"

	"github.com/google/uuid"
)

func GenUUID() string {
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
	}
	return u.String()
}
