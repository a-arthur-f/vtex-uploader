package uploader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type FileUploader struct {
	accountName        string
	accountEnvironment string
	client             *http.Client
}

func NewFileUploader(accountUrl string) (*FileUploader, error) {
	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, errors.New("Falha ao criar o FileUploader. A inicialização do cookiejar falhou")
	}

	accountInfos, err := getAccountInfos(accountUrl)

	if err != nil {
		error := fmt.Sprintf("Falha ao criar o FileUploader: %s", err)
		return nil, errors.New(error)
	}

	return &FileUploader{
		accountName:        accountInfos["accountName"],
		accountEnvironment: accountInfos["accountEnvironment"],
		client: &http.Client{
			Jar: jar,
		},
	}, nil
}

func getAccountInfos(accountUrl string) (map[string]string, error) {
	url, err := url.Parse(accountUrl)

	if err != nil {
		return nil, errors.New("Falha ao obter as informações da conta")
	}

	host := url.Hostname()
	splitedHost := strings.SplitN(host, ".", 2)

	name := splitedHost[0]
	env := splitedHost[1]

	return map[string]string{
		"accountName":        name,
		"accountEnvironment": env,
	}, nil
}

func (fileUploader *FileUploader) Login(username string, password string) error {
	err := fileUploader.startLogin(username)

	if err != nil {
		return err
	}

	requiresMfa, err := fileUploader.validateLogin(username, password)

	if err != nil {
		return err
	}

	if requiresMfa {
    for {
      var token string

      fmt.Print("Insira o código de autenticação: ")
      fmt.Scan(&token)
      
      mfaSuccess, err := fileUploader.validateMfa(token)

      if err != nil {
        return err
      }

      if mfaSuccess {
        fmt.Println("Autenticado com sucesso.")
        break
      } else {
        fmt.Println("Código inválido.")
      }
    }
	}

	return nil
}

func (fileUploader *FileUploader) startLogin(user string) error {
	path := "/api/vtexid/pub/authentication/startlogin"
	url := fmt.Sprintf("https://%s.%s%s",
		fileUploader.accountName,
		fileUploader.accountEnvironment,
		path)

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	writer.WriteField("user", user)
	writer.Close()

	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())

	_, err := fileUploader.client.Post(url, contentType, body)

	if err != nil {
		return errors.New("Falha ao iniciar o login.")
	}

	return nil
}

func (fileUploader *FileUploader) validateLogin(user string, password string) (requireMfa bool, err error) {
	path := "/api/vtexid/pub/authentication/classic/validate"
	url := fmt.Sprintf("https://%s.%s%s",
		fileUploader.accountName,
		fileUploader.accountEnvironment,
		path)

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	writer.WriteField("login", user)
	writer.WriteField("password", password)
	writer.Close()

	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())

	resp, err := fileUploader.client.Post(url, contentType, body)

	if err != nil {
		return false, errors.New("Falha ao validar o login.")
	}

	respJson := make(map[string]any)

	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&respJson)

	if respJson["authStatus"] == "RequiresMFA" {
		return true, nil
	}

	return false, nil
}

func (fileUploader *FileUploader) validateMfa(token string) (bool, error) {
	path := "/api/vtexid/pub/mfa/validate"
	url := fmt.Sprintf("https://%s.%s%s",
		fileUploader.accountName,
		fileUploader.accountEnvironment,
		path)

  body := &bytes.Buffer{}

  writer := multipart.NewWriter(body)
  writer.WriteField("mfaToken", token)
  writer.Close()

  contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())
  
  resp, err := fileUploader.client.Post(url, contentType, body)

  if err != nil {
    return false, errors.New("Falha na validação MFA.")
  }

  respJson := make(map[string]any)

  decoder := json.NewDecoder(resp.Body)
  decoder.Decode(&respJson)

  if respJson["authStatus"] != "Success" {
    return false, nil
  }

  return true, nil
}

func (fileUploader *FileUploader) Upload(filePath string) error {

	return nil
}
