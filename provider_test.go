package sdkgoconfig

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func hostAndFromURL(url string) (string, string) {
	hostAndPort := url[7:]
	return strings.Split(hostAndPort, ":")[0], strings.Split(hostAndPort, ":")[1]
}
func TestInit(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("40004"))
	}))
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")

	tests := []struct {
		name        string
		blockDir    string
		expectedErr error
	}{
		{
			name:        "valid block dir",
			blockDir:    "testdata/block",
			expectedErr: nil,
		},
		{
			name:        "invalid block dir",
			blockDir:    "testdata/invalid",
			expectedErr: fmt.Errorf("kapeta.yml file contained invalid YML: testdata/invalid"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			CONFIG.provider = nil
			_, err := Init(test.blockDir)
			if err != nil {
				if test.expectedErr == nil {
					t.Errorf("Init() returned error: %v", err)
				} else if err.Error() != test.expectedErr.Error() {
					t.Errorf("Init() returned unexpected error: %v, expected: %v", err, test.expectedErr)
				}
			} else if test.expectedErr != nil {
				t.Errorf("Init() did not return error, expected: %v", test.expectedErr)
			}
		})
	}
}

func TestGetOrDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("40004"))
	}))
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")

	tests := []struct {
		name         string
		blockDir     string
		path         string
		defaultValue interface{}
		expected     interface{}
	}{
		{
			name:         "valid path",
			blockDir:     "testdata/block",
			path:         "/foo/bar",
			defaultValue: "baz",
			expected:     "baz",
		},
		{
			name:         "invalid path",
			blockDir:     "testdata/block",
			path:         "/foo/bar/baz",
			defaultValue: "baz",
			expected:     "baz",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider, err := Init(test.blockDir)
			if err != nil {
				t.Fatalf("Init() returned error: %v", err)
			}

			actual := provider.GetOrDefault(test.path, test.defaultValue)
			if actual != test.expected {
				t.Errorf("GetOrDefault() returned unexpected value: %v, expected: %v", actual, test.expected)
			}
		})
	}
}
