// Package speedfixationservice provides methods for handling traffic camera requests
package speedfixationservice

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/IgorRybak2055/speed-control-service/internal/speedfixationservice/repo"
	"github.com/IgorRybak2055/speed-control-service/internal/speedfixationservice/usecase"
)

func fillTestData(t testing.TB, tempDir string) {
	t.Helper()

	var testData = []repo.SpeedFixation{
		{
			Date:          time.Now().UTC(),
			VehicleNumber: "6048 EC-3",
			Speed:         54.2,
		},
		{
			Date:          time.Now().UTC(),
			VehicleNumber: "0003 AE-3",
			Speed:         84.5,
		},
		{
			Date:          time.Now().UTC(),
			VehicleNumber: "8911 EE-3",
			Speed:         65.7,
		},
	}

	file, err := os.Create(filepath.Join(tempDir, time.Now().Format("02.01.2006")+".json"))
	require.NoError(t, err)

	encoder := json.NewEncoder(file)
	require.NoError(t, encoder.Encode(testData))
}

func createTempDir(t testing.TB) (string, func()) {
	t.Helper()

	tempDir, err := ioutil.TempDir(filepath.Join("data", "testdata"), "tmpData")
	require.NoError(t, err)

	return tempDir, func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}
}

func buildRequest(serverURL string) (*http.Request, error) {
	var (
		err error
		req = &http.Request{
			Method: http.MethodPost,
		}
	)

	req.URL, err = url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	q.Set("date", "27.12.2019 15:03:27")
	q.Set("vehicle_number", "6048 EC-3")
	q.Set("speed", "100")

	req.URL.RawQuery = q.Encode()

	return req, nil
}

func TestSpeedFixationService(t *testing.T) {
	var (
		err      error
		resp     *http.Response
		respBody []byte
		ans      interface{}
		req      *http.Request
		client   = &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:    100,
				IdleConnTimeout: 90 * time.Second,
			},
		}
	)

	tempDir, dropFile := createTempDir(t)
	defer dropFile()

	fillTestData(t, tempDir)

	srv := service{uc: usecase.NewSpeedFixationUsecase(repo.NewTestSpeedFixationRepository(tempDir))}

	server := httptest.NewServer(http.HandlerFunc(srv.registerSpeed))
	defer server.Close()

	req, err = buildRequest(server.URL)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)

	respBody, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	err = json.Unmarshal(respBody, &ans)
	require.NoError(t, err)

	log.Println(ans)
}
