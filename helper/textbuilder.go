package helper

import (
	"os"
	"regexp"
	"strings"

	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

const Expression = "\\${[a-z,0-9,A-Z,:,\\-,_]*}"

func Translate(txt string, h *datahub.Hub, data toolkit.M) (string, error) {
	r, e := regexp.Compile(Expression)
	if e != nil {
		return txt, e
	}

	res := txt
	elems := r.FindAllString(txt, -1)
	for _, el := range elems {
		pureEl := el[2 : len(el)-1]
		elemParts := strings.Split(pureEl, ":")

		switch elemParts[0] {
		case "ctx":
			newTxt := data.GetString(elemParts[1])
			res = strings.ReplaceAll(res, el, newTxt)

		case "var":
			varObj := new(model.Variable)
			varObj.ID = elemParts[1]
			h.Get(varObj)
			res = strings.ReplaceAll(res, el, varObj.Value)

		case "env":
			newTxt := os.Getenv(elemParts[1])
			res = strings.ReplaceAll(res, el, newTxt)
		}
	}
	return res, nil
}
