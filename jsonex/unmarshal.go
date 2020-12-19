package jsonex

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func UnmarshalFromReader(r io.Reader, out interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	logrus.Debugf("UnmarshalFromReader: %s", data)

	return json.Unmarshal(data, out)
}
