package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

type EventJson []byte

var NoEventJson EventJson

type Composer struct {
	vm       *jsonnet.VM
	contents jsonnet.Contents
	name     string
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewComposerByFile(dataset string, group string, filePath string) (*Composer, error) {
	vm := jsonnet.MakeVM()
	for _, f := range getFuncs(dataset, group) {
		vm.NativeFunction(f)
	}

	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	file := filepath.Base(filePath)
	contents := jsonnet.MakeContents(string(fileData))

	return &Composer{vm: vm, name: file, contents: contents}, nil

}

func NewComposerByContent(dataset string, instanceId string, name string, data []byte) (*Composer, error) {
	vm := jsonnet.MakeVM()
	for _, f := range getFuncs(dataset, instanceId) {
		vm.NativeFunction(f)
	}

	contents := jsonnet.MakeContents(string(data))

	return &Composer{vm: vm, name: name, contents: contents}, nil

}

func (c *Composer) NewEvent() (EventJson, EventObject, error) {
	c.vm.Importer(&jsonnet.MemoryImporter{Data: map[string]jsonnet.Contents{c.name: c.contents}})
	code, err := jsonnet.SnippetToAST("<snippet>", fmt.Sprintf(`import "%s"`, c.name))
	if err != nil {
		return nil, nil, err
	}
	v, err := c.vm.Evaluate(code)
	eventJson := EventJson(v)
	var eventObject EventObject
	err = json.Unmarshal(eventJson, &eventObject)
	if err != nil {
		return nil, nil, err
	}
	return eventJson, eventObject, err
}

func getFuncs(dataset string, instanceId string) []*jsonnet.NativeFunction {
	return []*jsonnet.NativeFunction{
		{
			Params: ast.Identifiers{"default"},
			Name:   "get_dataset",
			Func: func(args []interface{}) (interface{}, error) {
				if dataset != "" {
					return dataset, nil
				}
				v, err := getOneStringArg(args)
				if err != nil {
					return "", err
				}
				return v, nil
			},
		},
		{
			Params: ast.Identifiers{"default"},
			Name:   "get_instance_id",
			Func: func(args []interface{}) (interface{}, error) {
				if instanceId != "" {
					return instanceId, nil
				}
				v, err := getOneStringArg(args)
				if err != nil {
					return "", err
				}
				return v, nil
			},
		},
		{
			Params: ast.Identifiers{"values"},
			Name:   "get_one_of",
			Func: func(args []interface{}) (interface{}, error) {
				v, err := getOneStringArg(args)
				if err != nil {
					return []interface{}{}, err
				}
				vals := strings.Split(v, ",")
				val := strings.Trim(vals[rand.Intn(len(vals))], " ")
				return val, nil
			},
		},
		{
			Params: ast.Identifiers{"from", "to", "duration"},
			Name:   "get_timestamp",
			Func: func(args []interface{}) (interface{}, error) {
				if len(args) != 3 {
					return float64(0), fmt.Errorf("unexpected number of arguments, expected: 3, get: %d", len(args))
				}
				params, err := getStringArgs(args)
				if err != nil {
					return float64(0), err
				}
				fromTime, err := time.Parse("2006-01-02", params[0])
				if err != nil {
					return float64(0), fmt.Errorf("unexpected from time format, should be in a form YYYY-MM-DD: %v", err)
				}
				fromTime = time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, fromTime.Location())
				toTime, err := time.Parse("2006-01-02", params[1])
				if err != nil {
					return float64(0), fmt.Errorf("unexpected to time format, should be in a form YYYY-MM-DD: %v", err)
				}
				toTime = time.Date(toTime.Year(), toTime.Month(), toTime.Day()+1, 0, 0, 0, 0, toTime.Location()).Add(-time.Nanosecond)
				d, err := time.ParseDuration(params[2])
				if err != nil {
					return float64(0), fmt.Errorf("unexpected duration format, should be in the one of 1h, 1s, 1ms etc: %v", err)
				}
				dur := d.Nanoseconds()
				diff := toTime.UnixNano() - fromTime.UnixNano()
				if diff < dur {
					dur = diff
				}
				ts := rand.Int63n(diff / dur)
				res := time.Duration(fromTime.Add(time.Duration(ts * dur)).UnixNano()).Seconds()
				return res, nil
			},
		},
		{
			Params: ast.Identifiers{"min", "max"},
			Name:   "get_integer",
			Func: func(args []interface{}) (interface{}, error) {
				if len(args) != 2 {
					return float64(0), fmt.Errorf("unexpected number of arguments, expected: 2, get: %d", len(args))
				}
				from, err := getInt64Arg(args, 0)
				if err != nil {
					return float64(0), err
				}
				to, err := getInt64Arg(args, 1)
				if err != nil {
					return float64(0), err
				}
				if to == from {
					return to, nil
				}
				diff := rand.Int63n(to - from)
				return float64(from + diff), nil
			},
		},
		{
			Params: ast.Identifiers{"min", "max"},
			Name:   "get_number",
			Func: func(args []interface{}) (interface{}, error) {
				if len(args) != 2 {
					return float64(0), fmt.Errorf("unexpected number of arguments, expected: 2, get: %d", len(args))
				}
				from, err := getFloat64Arg(args, 0)
				if err != nil {
					return float64(0), err
				}
				to, err := getFloat64Arg(args, 1)
				if err != nil {
					return float64(0), err
				}
				if to == from {
					return to, nil
				}

				v := (from + to) * rand.Float64()

				return v, nil
			},
		},
		{
			Params: ast.Identifiers{},
			Name:   "get_rand_data",
			Func: func(args []interface{}) (interface{}, error) {
				return getRandData()
			},
		},
		{
			Params: ast.Identifiers{},
			Name:   "get_rand_user_agent",
			Func: func(args []interface{}) (interface{}, error) {
				ua := getRandUserAgent()
				return ua, nil
			},
		},
	}
}

func getOneStringArg(args []interface{}) (string, error) {
	return getStringArg(args, 0)
}

func getStringArg(args []interface{}, i int) (string, error) {
	if i >= len(args) {
		return "", errors.New("invalid number of arguments")
	}

	s, ok := args[i].(string)
	if !ok {
		return "", errors.New("arg is not a string")
	}

	return s, nil
}

func getInt64Arg(args []interface{}, i int) (int64, error) {
	if i >= len(args) {
		return 0, errors.New("invalid number of arguments")
	}

	switch t := args[i].(type) {
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, err
		}
		return n, nil
	case []byte:
		n, err := strconv.ParseInt(string(t), 10, 64)
		if err != nil {
			return 0, err
		}
		return n, nil
	case float64:
		return int64(t), nil
	case int:
		return int64(t), nil
	case int64:
		return t, nil
	}

	return 0, fmt.Errorf("unexpected type %T", args[i])
}

func getFloat64Arg(args []interface{}, i int) (float64, error) {
	if i >= len(args) {
		return 0, errors.New("invalid number of arguments")
	}

	switch t := args[i].(type) {
	case string:
		n, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, err
		}
		return n, nil
	case []byte:
		n, err := strconv.ParseFloat(string(t), 64)
		if err != nil {
			return 0, err
		}
		return n, nil
	case float64:
		return t, nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	}

	return 0, fmt.Errorf("unexpected type %T", args[i])
}

func getStringArgs(args []interface{}) ([]string, error) {
	strArgs := make([]string, len(args))
	for i := range args {
		s, err := getStringArg(args, i)
		if err != nil {
			return nil, err
		}
		strArgs[i] = s
	}
	return strArgs, nil
}
