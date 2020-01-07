// Package speedfixationservice provides methods for handling traffic camera requests
package speedfixationservice

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/luno/jettison/errors"

	"github.com/IgorRybak2055/speed-control-service/internal/speedfixationservice/repo"
	"github.com/IgorRybak2055/speed-control-service/internal/speedfixationservice/usecase"
	"github.com/IgorRybak2055/speed-control-service/pkg/env"
)

var (
	// httpAddr - server address in network
	httpAddr = flag.String("addr", ":8001", "server addr")
)

type service struct {
	uc    usecase.SpeedControl
	start time.Time
	end   time.Time
}

// Run start service
func Run() {
	var (
		srv service
		err error
	)

	srv.start, err = time.Parse(time.Kitchen, env.GetString("startWork", "12:00AM"))
	if err != nil {
		log.Fatal(err)
	}

	srv.end, err = time.Parse(time.Kitchen, env.GetString("endWork", "11:00PM"))
	if err != nil {
		log.Fatal(err)
	}

	sfr := repo.NewSpeedFixationRepository()
	srv.uc = usecase.NewSpeedFixationUsecase(sfr)

	srv.serviceHandlers()
}

func (srv *service) serviceHandlers() {
	mainMux := http.NewServeMux()
	mainMux.HandleFunc("/register", srv.registerSpeed)

	limitedMux := http.NewServeMux()
	limitedMux.HandleFunc("/overspeed", srv.overSpeed)
	limitedMux.HandleFunc("/minmaxspeed", srv.minMaxSpeed)
	loginHandler := srv.checkTimeMiddleware(limitedMux)
	mainMux.Handle("/", loginHandler)

	log.Printf("Start server at :%v ...", *httpAddr)
	err := http.ListenAndServe(*httpAddr, mainMux)

	if err != nil {
		log.Fatal("Error happened:", err.Error())
	}
}

func (srv *service) checkTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if time.Now().Hour() >= srv.start.Hour() && time.Now().Hour() < srv.end.Hour() {
			next.ServeHTTP(w, r)
		} else {
			responseError(w, errors.New(fmt.Sprintf("Service work from %v till %v",
				srv.start.Hour(), srv.end.Hour())), http.StatusNotAcceptable)
		}
	})
}

func responseError(w http.ResponseWriter, err error, status int) {
	resp, err := json.Marshal(err)
	if err != nil {
		http.Error(w, "cannot serialize srv", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(resp)

	if err != nil {
		http.Error(w, "response error: ", http.StatusInternalServerError)
	}
}

func makeResponse(w http.ResponseWriter, ans interface{}) {
	resp, err := json.Marshal(ans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		resp, err = json.Marshal(errors.New("error serializing srv"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (srv service) registerSpeed(w http.ResponseWriter, r *http.Request) {
	var (
		speedFixation repo.SpeedFixation
		err           error
	)

	if r.Method != http.MethodPost {
		responseError(w, errors.New("incorrect request method"), http.StatusBadRequest)
		return
	}

	datetime := r.FormValue("date")
	if datetime == "" {
		responseError(w, errors.New("datetime not defined in this request"), http.StatusBadRequest)
		return
	}

	speedFixation.Date, err = time.Parse("02.01.2006 15:04:05", datetime)
	if err != nil {
		responseError(w, errors.New("unable parse datetime"), http.StatusBadRequest)
		return
	}

	speedFixation.VehicleNumber = r.FormValue("vehicle_number")
	if speedFixation.VehicleNumber == "" {
		responseError(w, errors.New("vehicle number not defined in this request"), http.StatusBadRequest)
		return
	}

	if speedFixation.Speed, err = strconv.ParseFloat(r.FormValue("speed"), 64); err != nil {
		responseError(w, errors.New("unable parse speed"), http.StatusBadRequest)
		return
	}

	if speedFixation.Speed == 0 {
		responseError(w, errors.New("speed not defined in this request"), http.StatusBadRequest)
		return
	}

	if err := srv.uc.CreateRecord(speedFixation); err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}

	makeResponse(w, "register success")
}

func (srv service) overSpeed(w http.ResponseWriter, r *http.Request) {
	var (
		samplingConditions repo.SpeedFixation
		err                error
	)

	if r.Method != http.MethodGet {
		responseError(w, errors.New("incorrect request method"), http.StatusBadRequest)
		return
	}

	datetime := r.FormValue("date")
	if datetime == "" {
		responseError(w, errors.New("datetime not defined in this request"), http.StatusBadRequest)
		return
	}

	samplingConditions.Date, err = time.Parse("02.01.2006", datetime)
	if err != nil {
		responseError(w, errors.New("unable parse datetime"), http.StatusBadRequest)
		return
	}

	if samplingConditions.Speed, err = strconv.ParseFloat(r.FormValue("speed"), 64); err != nil {
		responseError(w, errors.New("unable parse speed"), http.StatusBadRequest)
		return
	}

	if samplingConditions.Speed == 0 {
		responseError(w, errors.New("speed not defined in this request"), http.StatusBadRequest)
		return
	}

	resp, err := srv.uc.LookUpOverSpeedByDate(samplingConditions)
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}

	makeResponse(w, resp)
}

func (srv service) minMaxSpeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responseError(w, errors.New("incorrect request method"), http.StatusBadRequest)
		return
	}

	date, err := time.Parse("02.01.2006", r.FormValue("date"))
	if err != nil {
		responseError(w, errors.New("unable parse datetime"), http.StatusBadRequest)
		return
	}

	resp, err := srv.uc.LookUpMinMaxSpeedByDate(date)
	if err != nil {
		responseError(w, err, http.StatusBadRequest)
		return
	}

	makeResponse(w, resp)
}
