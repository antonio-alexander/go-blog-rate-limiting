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
	if s := envs["ID"]; s != "" {
		r.Id = s
	}
	if s := envs["APPLICATION_ID"]; s != "" {
		r.ApplicationId = s
	}
	if s := envs["WEIGHT"]; s != "" {
		r.Weight, _ = strconv.ParseInt(s, 10, 64)
	}
	if s := envs["WAIT"]; s != "" {
		i, _ := strconv.Atoi(s)
		r.Wait = time.Duration(i) * time.Second
	}
}

func (r *Request) FromRequest(request *http.Request) error {
	if r == nil {
		return errors.New("request is nil")
	}
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, r)
}

func (r *Request) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Request) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, r)
}
