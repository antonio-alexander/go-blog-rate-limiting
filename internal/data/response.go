package data

import "time"

type Response struct {
	Id            string        `json:"id"`
	ApplicationId string        `json:"application_id"`
	Wait          time.Duration `json:"wait,string"`
	Weight        int64         `json:"weight,string"`
}
