package data

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	defaultWait   time.Duration = 5 * time.Second
	defaultWeight int64         = 1
)

const (
	ID             string = "ID"
	WAIT           string = "WAIT"
	WEIGHT         string = "WEIGHT"
	APPLICATION_ID string = "APPLICATION_ID"
)

type Request struct {
	Id            string        `json:"id"`
	ApplicationId string        `json:"application_id"`
	Wait          time.Duration `json:"wait,string"`
	Weight        int64         `json:"weight,string"`
}

func NewRequest() *Request {
	return &Request{
		Wait:   defaultWait,
		Weight: defaultWeight,
	}
}

func (r *Request) FromArgs(args []string) {
	var wait int

	flag.StringVar(&r.Id, "id", uuid.Must(uuid.NewRandom()).String(), "")
	flag.StringVar(&r.ApplicationId, "application_id", uuid.Must(uuid.NewRandom()).String(), "")
	flag.IntVar(&wait, "wait", int(defaultWait/time.Second), "")
	flag.Int64Var(&r.Weight, "weight", defaultWeight, "")
	flag.Parse()
	r.Wait = time.Second * time.Duration(wait)
}

func (r *Request) FromEnvs(envs map[string]string) {
	if s := envs[ID]; s != "" {
		r.Id = s
	}
	if s := envs[APPLICATION_ID]; s != "" {
		r.ApplicationId = s
	}
	if s := envs[WEIGHT]; s != "" {
		r.Weight, _ = strconv.ParseInt(s, 10, 64)
	}
	if s := envs[WAIT]; s != "" {
		i, _ := strconv.Atoi(s)
		r.Wait = time.Duration(i) * time.Second
	}
}

func (r *Request) FromRequest(request *http.Request) error {
	if r == nil {
		return errors.New("request is nil")
	}
	switch request.Method {
	default:
		return errors.Errorf("unsupported method: %s", request.Method)
	case http.MethodGet:
		request.ParseForm()
		if s := request.Form.Get("id"); s != "" {
			r.Id = s
		}
		if s := request.Form.Get("application_id"); s != "" {
			r.ApplicationId = s
		}
		if s := request.Form.Get("weight"); s != "" {
			r.Weight, _ = strconv.ParseInt(s, 10, 64)
		}
		if s := request.Form.Get("wait"); s != "" {
			i, _ := strconv.ParseInt(s, 10, 64)
			r.Wait = time.Duration(i) * time.Second
		}
	case http.MethodPost:
		bytes, err := io.ReadAll(request.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, r); err != nil {
			return err
		}
	}
	return nil
}
