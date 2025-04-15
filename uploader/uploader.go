package uploader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type FileUploader struct {
	accountUrl *url.URL
	client     *http.Client
}

type accountData struct {
	Url      string       `json:"url"`
	Username string       `json:"username"`
	Password string       `json:"password"`
	Token    accountToken `json:"token"`
}

type accountToken struct {
	Expires time.Time `json:"expires"`
	Value   string    `json:"value"`
}

func NewFileUploader(accountUrl string) (*FileUploader, error) {
	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, errors.New("Falha ao criar o FileUploader. A inicialização do cookiejar falhou")
	}

	parsedUrl, err := url.Parse(accountUrl)

	if err != nil {
		return nil, errors.New("Falha ao criar o FileUploader. URL inválida")
	}

	return &FileUploader{
		accountUrl: parsedUrl,
		client: &http.Client{
			Jar: jar,
		},
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

	err = fileUploader.storeAccountData(username, password)

	if err != nil {
		return errors.New("Falha ao armazenar dados da conta")
	}

	return nil
}

func (fileUploader *FileUploader) startLogin(user string) error {
	path := "/api/vtexid/pub/authentication/startlogin"
	url := fmt.Sprintf("%s/%s",
		fileUploader.accountUrl,
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
	url := fmt.Sprintf("%s/%s",
		fileUploader.accountUrl,
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
	url := fmt.Sprintf("%s/%s",
		fileUploader.accountUrl,
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

func (fileUploader *FileUploader) storeAccountData(username string, password string) error {
	data := accountData{
		Url:      fileUploader.accountUrl.String(),
		Username: username,
		Password: password,
		Token: accountToken{
			Expires: time.Now().Add(time.Hour * 6),
			Value:   fileUploader.getCookie("VtexIdclientAutCookie"),
		},
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return errors.New("Falha ao armazenar dados da conta.")
	}

	err = os.WriteFile("account.json", jsonData, fs.ModePerm)

	if err != nil {
		return errors.New("Falha ao armazenar dados da conta.")
	}

	return nil
}

func (fileUploader *FileUploader) getCookie(name string) string {
	for _, cookie := range fileUploader.client.Jar.Cookies(fileUploader.accountUrl) {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}

func IsLoggedIn() bool {
	file, err := os.ReadFile("account.json")

	if err != nil {
		return false
	}

	var accData accountData

	err = json.Unmarshal(file, &accData)

	if err != nil {
		return false
	}

	if accData.Url == "" ||
		accData.Username == "" ||
		accData.Password == "" ||
		accData.Token.Value == "" ||
		accData.Token.Expires.Compare(time.Now()) == -1 {
		return false
	}

	return true
}

func Load() (*FileUploader, error) {
	var accData accountData

	file, err := os.ReadFile("account.json")

	if err != nil {
		return nil, errors.New("Falha ao carregar dados. Erro ao ler o arquivo account.json")
	}

	err = json.Unmarshal(file, &accData)

	if err != nil {
		return nil, errors.New("Falha ao carregar dados. Arquivo JSON inválido.")
	}

	fileUploader, err := NewFileUploader(accData.Url)

	if err != nil {
		return nil, errors.New("Falha ao carregar dados. Erro ao criar o FileUploader.")
	}

	cookie := &http.Cookie{
		Name:   "VtexIdclientAutCookie",
		Quoted: false,
		Domain: fileUploader.accountUrl.Host,
	}

	fileUploader.client.Jar.SetCookies(fileUploader.accountUrl, []*http.Cookie{cookie})

	for _, cook := range fileUploader.client.Jar.Cookies(fileUploader.accountUrl) {
		fmt.Println(cook.Value)
	}

	return fileUploader, nil
}

func (fileUploader *FileUploader) Upload(filePath string) error {

	return nil
}
