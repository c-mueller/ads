package ads

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestListFetch(t *testing.T) {
	hostlist, err := os.Open("testdata/test_blocklist")
	defer hostlist.Close()
	assert.NoError(t, err)
	hostlistData, err := ioutil.ReadAll(hostlist)

	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		w.Write(hostlistData)
		w.Header().Set("Content-Type", "text/plain")
	}

	srv := httptest.NewServer(http.HandlerFunc(handlerFunc))
	defer srv.Close()

	queryUrl := fmt.Sprintf("%s/hosts.txt", srv.URL)

	list, err := GenerateBlockageMap([]string{queryUrl})
	assert.NoError(t, err)

	expectedList, err := os.Open("testdata/test_blocklist_expected_domains")
	defer expectedList.Close()
	assert.NoError(t, err)
	expData, err := ioutil.ReadAll(expectedList)

	for _, url := range strings.Split(string(expData), "\n") {
		assert.True(t, list[url])
	}
	assert.False(t, list["testme.com"])
}
