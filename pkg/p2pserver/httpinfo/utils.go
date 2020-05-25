package httpinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const namespace = "stat"

func txPath(method string) string {
	return fmt.Sprintf("/%s/tx/%s", namespace, method)
}

func errors(w http.ResponseWriter, err error) {
	bz, _ := json.Marshal(&Resp{
		Err:     err.Error(),
		Succeed: false,
		Data:    "",
	})
	_, _ = fmt.Fprintf(w, string(bz))
}

func result(w http.ResponseWriter, data interface{}) {
	bz, _ := json.Marshal(&Resp{
		Err:     "",
		Succeed: true,
		Data:    data,
	})
	_, _ = fmt.Fprintf(w, string(bz))
}

type Resp struct {
	Err     string      `json:"err"`
	Succeed bool        `json:"succeed"`
	Data    interface{} `json:"data"`
}
