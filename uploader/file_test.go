package uploader

import "testing"

func TestNewFileError(t *testing.T) {
	_, err := NewFile("teste")

	if err == nil {
		t.Log("err should not be nil")
		t.Fail()
	}
}

func TestGetFileTypeJS(t *testing.T) {
	ext := getFileType("teste.json")

	if ext != JS {
		t.Log("fileType should be js")
		t.Fail()
	}
}

func TestGetFileTypeCSS(t *testing.T) {
	ext := getFileType("teste.css")

	if ext != CSS {
		t.Log("fileType should be css")
		t.Fail()
	}
}

func TestGetFileTypeImage(t *testing.T) {
	ext := getFileType("teste.png")

	if ext != IMG {
		t.Log("with .png fileType should be image")
		t.Fail()
	}

	ext = getFileType("teste.jpg")

	if ext != IMG {
		t.Log("with .jpg fileType should be image")
		t.Fail()
	}

	ext = getFileType("teste.gif")

	if ext != IMG {
		t.Log("with .gif fileType should be image")
		t.Fail()
	}
}
