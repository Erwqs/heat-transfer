package freader

import "os"

func GetToken() string {
	buf, err := os.ReadFile("./token")
	if err != nil {
		panic(err)
	}

	return string(buf)
}