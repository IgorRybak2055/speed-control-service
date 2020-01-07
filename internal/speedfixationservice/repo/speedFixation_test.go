package repo

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

var testData = []SpeedFixation{
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

func readFile(t testing.TB, tempDir string, fileName string) []SpeedFixation {
	var (
		data     dataSlice
		jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary
	)

	path := filepath.Join(tempDir, fileName+".json")

	file, err := ioutil.ReadFile(filepath.Clean(path))
	require.NoError(t, err)

	require.NoError(t, jsonIter.Unmarshal(file, &data))

	return data
}

func BenchmarkLookUpOverSpeedByDate(b *testing.B) {
	var err error

	tempDir, dropFile := createTempDir(b)
	defer dropFile()

	fillTestData(b, tempDir)

	sfr := NewTestSpeedFixationRepository(tempDir)

	a := SpeedFixation{}

	a.Date = time.Now()

	require.NoError(b, err)

	a.Speed = 100

	for i := 0; i < b.N; i++ {
		_, err := sfr.LookUpOverSpeedByDate(a)
		require.NoError(b, err)
	}
}

func BenchmarkLookUpMinMaxSpeedByDate(b *testing.B) {
	var err error

	tempDir, dropFile := createTempDir(b)
	defer dropFile()

	fillTestData(b, tempDir)

	sfr := NewTestSpeedFixationRepository(tempDir)

	for i := 0; i < b.N; i++ {
		_, err = sfr.LookUpMinMaxSpeedByDate(time.Now())
		require.NoError(b, err)
	}
}

func BenchmarkCreateRecord(b *testing.B) {
	var err error

	tempDir, dropFile := createTempDir(b)
	defer dropFile()

	sfr := NewTestSpeedFixationRepository(tempDir)

	sf := SpeedFixation{}

	sf.Date = time.Now()

	require.NoError(b, err)

	sf.VehicleNumber = "6048 EC-3"

	sf.Speed = 100

	for i := 0; i < b.N; i++ {
		require.NoError(b, sfr.CreateRecord(sf))
	}
}

func createTempDir(t testing.TB) (string, func()) {
	t.Helper()

	tempDir, err := ioutil.TempDir(filepath.Join("..", "data", "testdata"), "tmpData")
	require.NoError(t, err)

	return tempDir, func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}
}

func Test_speedFixationRepo_CreateRecord(t *testing.T) {
	type args struct {
		fixation SpeedFixation
	}

	tempDir, dropFile := createTempDir(t)
	defer dropFile()

	tt := struct {
		args    args
		wantErr bool
	}{
		args: args{
			fixation: SpeedFixation{
				Date:          time.Now().UTC(),
				VehicleNumber: "6048 EC-3",
				Speed:         62.8,
			},
		},
	}

	sf := speedFixationRepo{
		storage: tempDir,
		mu:      &sync.Mutex{},
	}

	err := sf.CreateRecord(tt.args.fixation)

	if tt.wantErr {
		require.Error(t, err)
		return
	}

	require.NoError(t, err)

	got := readFile(t, tempDir, time.Now().Format("02.01.2006"))
	require.Equal(t, tt.args.fixation, got[0])
}

func fillTestData(t testing.TB, tempDir string) {
	t.Helper()

	file, err := os.Create(filepath.Join(tempDir, time.Now().Format("02.01.2006")+".json"))
	require.NoError(t, err)

	encoder := json.NewEncoder(file)
	require.NoError(t, encoder.Encode(testData))
}

func Test_speedFixationRepo_LookUpMinMaxSpeedByDate(t *testing.T) {
	type args struct {
		date time.Time
	}

	tempDir, dropFile := createTempDir(t)
	defer dropFile()

	fillTestData(t, tempDir)

	tt := struct {
		args    args
		want    []SpeedFixation
		wantErr bool
	}{
		args:    args{date: time.Now()},
		want:    []SpeedFixation{testData[0], testData[1]},
		wantErr: false,
	}

	sf := NewTestSpeedFixationRepository(tempDir)

	got, err := sf.LookUpMinMaxSpeedByDate(tt.args.date)

	if tt.wantErr {
		require.Error(t, err)
		return
	}

	require.NoError(t, err)

	require.Equal(t, tt.want, got)
}

func Test_speedFixationRepo_LookUpOverSpeedByDate(t *testing.T) {
	type args struct {
		fixation SpeedFixation
	}

	tempDir, dropFile := createTempDir(t)
	defer dropFile()

	fillTestData(t, tempDir)

	tt := struct {
		args    args
		want    []SpeedFixation
		wantErr bool
	}{
		args: args{fixation: SpeedFixation{
			Date:  time.Now(),
			Speed: 60,
		}},
		want:    []SpeedFixation{testData[1], testData[2]},
		wantErr: false,
	}

	sf := NewTestSpeedFixationRepository(tempDir)

	got, err := sf.LookUpOverSpeedByDate(tt.args.fixation)

	if tt.wantErr {
		require.Error(t, err)
		return
	}

	require.NoError(t, err)

	require.Equal(t, tt.want, got)
}
