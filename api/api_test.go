package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeXdccEngine struct {
	Requested []string
}

func (e *fakeXdccEngine) Start() {
	e.Requested = make([]string, 0)
}

func (e *fakeXdccEngine) RequestFile(botNick string, packageNo int, fileName string) <-chan bool {
	r := make(chan bool)
	go func() {
		e.Requested = append(e.Requested, fmt.Sprintf("%s|%d|%s", botNick, packageNo, fileName))
		r <- true
	}()

	return r
}

func (e *fakeXdccEngine) DownloadsJSON(writer io.Writer) error {
	json := `[{"foo":"bar"}]`
	writer.Write([]byte(json))
	return nil
}

func TestPostDownloads(t *testing.T) {
	engine := &fakeXdccEngine{}
	engine.Start()
	router := NewRouter(engine)

	r, _ := http.NewRequest("POST", "/downloads", strings.NewReader(`{"fileName": "foo.mkv", "botNick": "bar", "packageNumber": 2137}`))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.Equal(t, "bar|2137|foo.mkv", engine.Requested[0])
	assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
}

func TestPostDownloadsUncompletePayload(t *testing.T) {
	engine := &fakeXdccEngine{}
	engine.Start()
	router := NewRouter(engine)

	r, _ := http.NewRequest("POST", "/downloads", strings.NewReader(`{"fileName": "foo.mkv", "botNick": "bar"}`))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.Len(t, engine.Requested, 0)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestPostDownloadsNotJSON(t *testing.T) {
	engine := &fakeXdccEngine{}
	engine.Start()
	router := NewRouter(engine)

	r, _ := http.NewRequest("POST", "/downloads", strings.NewReader("foo"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.Len(t, engine.Requested, 0)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestGetDownloads(t *testing.T) {
	engine := &fakeXdccEngine{}
	engine.Start()
	router := NewRouter(engine)

	r, _ := http.NewRequest("GET", "/downloads", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.Equal(t, `[{"foo":"bar"}]`, w.Body.String())
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}
