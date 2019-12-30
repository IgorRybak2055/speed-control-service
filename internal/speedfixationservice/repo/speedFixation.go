// Package repo provides all needs methods to work with data storage
package repo

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/luno/jettison/errors"
)

// ContactRepo representations ContractRepository interface
type speedFixationRepo struct {
	storage string
}

// NewTestSpeedFixationRepository will create an object that represent the SpeedControlRepo interface for testing
func NewTestSpeedFixationRepository(tempDir string) SpeedControlRepo {
	return &speedFixationRepo{storage: tempDir}
}

// NewSpeedFixationRepository will create an object that represent the SpeedControlRepo interface
func NewSpeedFixationRepository() SpeedControlRepo {
	return &speedFixationRepo{storage: filepath.Join("internal", "speedfixationservice", "data")}
}

func (sf speedFixationRepo) createFile() error {
	file, err := os.Create(filepath.Join(sf.storage, time.Now().Format("02.01.2006")+".json"))
	if err != nil {
		return err
	}

	if _, err := file.Write([]byte("[]")); err != nil {
		return err
	}

	return nil
}

func (sf speedFixationRepo) openFile() (*os.File, error) {
	file, err := os.OpenFile(filepath.Join(sf.storage, time.Now().Format("02.01.2006")+".json"),
		os.O_WRONLY, os.ModeAppend)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		if err = sf.createFile(); err != nil {
			return nil, err
		}

		return sf.openFile()
	}

	return file, nil
}

func (sf speedFixationRepo) CreateRecord(fixation SpeedFixation) error {
	var (
		file   *os.File
		err    error
		offset int64 = 1
	)

	if file, err = sf.openFile(); err != nil {
		return err
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	jsf, err := json.Marshal(fixation)
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()

	if fileInfo.Size() > 2 {
		_, err = file.WriteAt([]byte(","), fileSize-offset)
		if err != nil {
			log.Fatal(err)
		}

		offset = 0
	}

	n, err := file.WriteAt(jsf, fileSize-offset)
	if err != nil {
		return err
	}

	_, err = file.WriteAt([]byte("]"), fileSize-offset+int64(n))
	if err != nil {
		return err
	}

	return nil
}

func (sf speedFixationRepo) readFile(fileName string) ([]SpeedFixation, error) {
	var data dataSlice

	path := filepath.Join(sf.storage, fileName+".json")

	file, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (sf speedFixationRepo) selectViolators(fileName string, speedLimit float64) ([]SpeedFixation, error) {
	var (
		violators []SpeedFixation
		data      SpeedFixation
	)

	path := filepath.Join(sf.storage, fileName+".json")

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	decoder := json.NewDecoder(file)

	if _, err := decoder.Token(); err != nil {
		return nil, err
	}

	for decoder.More() {
		if err := decoder.Decode(&data); err != nil {
			return nil, err
		}

		if data.Speed > speedLimit {
			violators = append(violators, data)
		}
	}

	return violators, nil
}

func (sf speedFixationRepo) selectMinMaxSpeed(fileName string) ([]SpeedFixation, error) {
	var (
		ret                = make([]SpeedFixation, 2)
		data               SpeedFixation
		minSpeed, maxSpeed = 200.0, 0.0
	)

	path := filepath.Join(sf.storage, fileName+".json")

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	decoder := json.NewDecoder(file)

	if _, err := decoder.Token(); err != nil {
		return nil, err
	}

	for decoder.More() {
		if err := decoder.Decode(&data); err != nil {
			return nil, err
		}

		if data.Speed < minSpeed {
			minSpeed = data.Speed
			ret[0] = data
		}

		if data.Speed > maxSpeed {
			maxSpeed = data.Speed
			ret[1] = data
		}
	}

	return ret, nil
}

func (sf speedFixationRepo) LookUpOverSpeedByDate(fixation SpeedFixation) ([]SpeedFixation, error) {
	return sf.selectViolators(fixation.Date.Format("02.01.2006"), fixation.Speed)
}

func (sf speedFixationRepo) LookUpMinMaxSpeedByDate(date time.Time) ([]SpeedFixation, error) {
	return sf.selectMinMaxSpeed(date.Format("02.01.2006"))
}
