package main

import (
	"fmt"
	"arthur.fusco/vtex-uploader/uploader"
)

func main() {
  var url string

  fmt.Print("URL da loja: ")
  fmt.Scan(&url)

	fileUploader, err := uploader.NewFileUploader(url)

	if err != nil {
		fmt.Println(err)
	}

  var user, password string

  fmt.Print("Usuário: ")
  fmt.Scan(&user)
  fmt.Print("Senha: ")
  fmt.Scan(&password)

  fileUploader.Login(user, password)
}
