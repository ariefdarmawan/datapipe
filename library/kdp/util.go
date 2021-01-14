package kdp

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

type OutputTo struct {
	Name  string
	Scope string
	Index int
	Field string
}

type SetVarRequest struct {
	Scope  string
	Name   string
	PipeID string
	JobID  string
	Value  interface{}
}

func MapsOutputTo(data []toolkit.M, h *datahub.Hub, pipeID, jobID string, opt OutputTo) error {
	if opt.Name != "" {
		idx := opt.Index
		if idx < 0 {
			idx = len(data) + idx
		}
		if idx > len(data)-1 {
			return fmt.Errorf("Fail to map output to %s.%s: out of range", opt.Scope, opt.Name)
		}

		val := data[idx].Get(opt.Field)

		SetVar(h, SetVarRequest{
			Scope:  opt.Scope,
			Name:   opt.Name,
			PipeID: pipeID,
			JobID:  jobID,
			Value:  val,
		})
	}

	return nil
}

const Expression = "\\${[a-z,0-9,A-Z,:,\\-,_]*}"

func Translate(txt string, h *datahub.Hub, pipeID, jobID string) (string, error) {
	r, e := regexp.Compile(Expression)
	if e != nil {
		return txt, e
	}

	res := txt

	for {
		elems := r.FindAllString(res, -1)
		if len(elems) == 0 {
			break
		}

		for _, el := range elems {
			pureEl := el[2 : len(el)-1]
			elemParts := strings.Split(pureEl, ":")

			switch elemParts[0] {
			case "ctx", "job", "var":
				v, e := GetVar(h, SetVarRequest{
					Name:   elemParts[1],
					Scope:  elemParts[0],
					PipeID: pipeID,
					JobID:  jobID,
				})
				return toolkit.ToString(v.Value), e

			case "conn":
				varObj := new(model.Connection)
				varObj.ID = elemParts[1]
				h.Get(varObj)
				res = strings.ReplaceAll(res, el, varObj.Connection)

			case "env":
				newTxt := os.Getenv(elemParts[1])
				res = strings.ReplaceAll(res, el, newTxt)
			}
		}
	}

	return res, nil
}

func TranslateM(m toolkit.M, h *datahub.Hub, pipeID, jobID string) (toolkit.M, error) {
	var e error
	trM := toolkit.M{}
	for k, v := range m {
		rv := reflect.ValueOf(v)
		rvTypeStr := rv.Type().String()
		rvKind := rv.Kind()
		if rvTypeStr == "json.Number" {
			trM.Set(k, toolkit.ToFloat64(rv.Interface(), 4, toolkit.RoundingAuto))
		} else if rv.Kind() == reflect.String {
			translatedV := rv.String()
			translatedV, e = Translate(translatedV, h, pipeID, jobID)
			if e != nil {
				return m, fmt.Errorf("fail to translate %s: %s", k, v)
			}
			trM.Set(k, translatedV)
		} else if rvKind == reflect.Map || rvKind == reflect.Ptr {
			m2 := toolkit.M{}
			if e := toolkit.Serde(v, &m2, ""); e == nil {
				if m2, e = TranslateM(m2, h, pipeID, jobID); e != nil {
					return m, fmt.Errorf("fail to translate %s: %s", k, v)
				}
				trM.Set(k, m2)
			} else {
				trM.Set(k, v)
			}
		} else if rvKind == reflect.Slice {
			sliceLen := rv.Len()
			trMs := make([]interface{}, sliceLen)
			for midx := 0; midx < sliceLen; midx++ {
				sliceItem := reflect.ValueOf(rv.Index(midx).Interface())
				sliceItemKind := sliceItem.Kind()
				if sliceItemKind == reflect.Map || sliceItemKind == reflect.Ptr || sliceItemKind == reflect.Struct {
					mm := toolkit.M{}
					if err := toolkit.Serde(sliceItem.Interface(), &mm, ""); err != nil {
						return m, fmt.Errorf("fail to translate %s: %s", k, v)
					}
					trm, err := TranslateM(mm, h, pipeID, jobID)
					if err != nil {
						return m, fmt.Errorf("fail to translate %s: %s", k, v)
					}
					trMs[midx] = trm
				} else {
					trMs[midx] = sliceItem.Interface()
				}
			}
			trM.Set(k, trMs)
		} else {
			trM.Set(k, v)
		}
	}
	return trM, nil
}

func SetVar(h *datahub.Hub, request SetVarRequest) (string, error) {
	res := ""

	switch request.Scope {
	case "var":
		v := new(model.Variable)
		if e := h.GetByID(v, request.Name); e != nil {
			v.ID = request.Name
			v.Kind = "Text"
		}
		v.Value = toolkit.ToString(request.Value)
		if e := h.Save(v); e != nil {
			return res, nil
		}

	case "ctx":
		p := new(Pipe)
		if e := h.GetByID(p, request.PipeID); e != nil {
			return res, nil
		}
		data := p.Data
		data.Set(request.Name, request.Value)
		if e := h.Save(p, "Data"); e != nil {
			return res, nil
		}

	case "job":
		p := new(Job)
		if e := h.GetByID(p, request.JobID); e != nil {
			return res, nil
		}
		data := p.Data
		data.Set(request.Name, request.Value)
		if e := h.Save(p, "Data"); e != nil {
			return res, nil
		}
	}

	return res, nil
}

func GetVar(h *datahub.Hub, request SetVarRequest) (SetVarRequest, error) {
	switch request.Scope {
	case "var":
		v := new(model.Variable)
		h.GetByID(v, request.Name)
		request.Value = v.Value
		return request, nil

	case "ctx":
		p := new(Pipe)
		if e := h.GetByID(p, request.PipeID); e != nil {
			return request, nil
		}
		data := p.Data
		request.Value = data.Get(request.Name)
		return request, nil

	case "job":
		p := new(Job)
		if e := h.GetByID(p, request.JobID); e != nil {
			return request, nil
		}
		data := p.Data
		request.Value = data.Get(request.Name)
		return request, nil
	}

	return request, nil
}
