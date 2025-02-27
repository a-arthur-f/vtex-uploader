package main

import (
	"fmt"
	"arthur.fusco/vtex-uploader/uploader"
)

func main() {
	fileUploader, err := uploader.NewFileUploader("https://mrcatstore.myvtex.com")

	if err != nil {
		fmt.Println(err)
	}

  var user, password string

  fmt.Print("Insira o usuário: ")
  fmt.Scan(&user)
  fmt.Print("Insira a senha: ")
  fmt.Scan(&password)

  fileUploader.Login(user, password)
}
