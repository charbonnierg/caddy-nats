package natsmagic_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/charbonnierg/nats-magic/natsmagic"
)

func TestOptionsSerializeDeserialize(t *testing.T) {
	opts := natsmagic.Options{
		ConfigFile:      "/some/config/file",
		ServerName:      "nats-server",
		Host:            "127.0.0.1",
		Port:            4222,
		ClientAdvertise: "127.0.0.1:4222",
		Trace:           false,
		Debug:           false,
		TraceVerbose:    false,
	}
	rawdata, err := json.Marshal(opts)
	if err != nil {
		panic(err)
	}
	newopts := &natsmagic.Options{}
	json.Unmarshal(rawdata, newopts)
	if !reflect.DeepEqual(opts, newopts) {
		panic(string(rawdata))
	}
}
