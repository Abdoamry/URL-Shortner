package routes

import (
	"time"
	
	)

type request struct{
	URL  string `json:"url"`
	CustomShort string	`json:"short"`
	Expiry time.Duration `json:"expiry"`

}

type respone struct{
	URL  string `json:"url"`
	CustomShort string `json:"short"`
	Expiry time.Duration `json:"expiry"`
	XRateRemaining int `json:"rate_limit"`
	XRateLimitRest int `json:"rate_limit_reset"`

}