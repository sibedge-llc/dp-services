package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type sqlColumnDefinition struct {
	ColumnDefinition string
	Converter        Converter
	IsKey            bool
}

type sqlColumn struct {
	DataType  string
	Converter Converter
}

func toSqlColumnDefinition(name string, v interface{}) (*sqlColumnDefinition, error) {
	isKey := strings.HasSuffix(name, "id") && v != nil
	column, err := toSqlColumnTypeAndConverter(name, isKey, v)
	if err != nil {
		return nil, err
	}
	var columnDefinition string
	if isKey {
		columnDefinition = fmt.Sprintf("%s %s", name, column.DataType)
	} else {
		columnDefinition = fmt.Sprintf("%s %s NULL", name, column.DataType)
	}
	return &sqlColumnDefinition{
		ColumnDefinition: columnDefinition,
		Converter:        column.Converter,
		IsKey:            isKey,
	}, nil
}

func toSqlColumnTypeAndConverter(name string, isKey bool, v interface{}) (sqlColumn, error) {
	switch t := v.(type) {
	case map[string]interface{}:
		converter, err := toConverterByValue("jsonb", isKey, t)
		return sqlColumn{DataType: "jsonb", Converter: converter}, err
	case string:
		dataType := "text"
		if strings.HasPrefix(name, "time") || strings.HasSuffix(name, "time") {
			dataType = "timestamp"
		}
		converter, err := toConverterByValue(dataType, isKey, t)
		return sqlColumn{DataType: dataType, Converter: converter}, err
	case float64:
		dataType := "decimal"
		if strings.HasSuffix(name, "id") || t == float64(int64(t)) {
			dataType = "integer"
		}
		converter, err := toConverterByValue(dataType, isKey, t)
		return sqlColumn{DataType: dataType, Converter: converter}, err
	case nil:
		converter, err := toConverterByValue("text", isKey, t)
		return sqlColumn{DataType: "text", Converter: converter}, err
	case []interface{}:
		var v0 interface{}
		if len(t) > 0 {
			v0 = t[0]
		} else {
			v0 = ""
		}
		column, err := toSqlColumnTypeAndConverter(name, false, v0)
		if err != nil {
			return sqlColumn{}, err
		}
		if strings.HasPrefix(column.DataType, "[]") {
			return sqlColumn{
				DataType: "[]text",
				Converter: func(in interface{}) (string, error) {
					strVals := make([]string, 0, len(in.([]interface{})))
					for _, v := range in.([]interface{}) {
						d, err := json.Marshal(v)
						if err != nil {
							return "", err
						}
						str := string(d)
						strVals = append(strVals, toSqlString(str))
					}
					return fmt.Sprintf("ARRAY[%s]", strings.Join(strVals, ",")), nil
				},
			}, nil
		}
		dataType := fmt.Sprintf("[]%s", column.DataType)
		converter, err := toConverterByValue(dataType, isKey, t)
		return sqlColumn{DataType: dataType, Converter: converter}, err
	default:
		zap.L().Error("unsupported type", zap.String("type", fmt.Sprintf("%T", t)))

	}

	converter, err := toConverterByValue("text", isKey, nil)
	return sqlColumn{DataType: "text", Converter: converter}, err
}

func toConverterByColumnDef(colDef sqlColumnType, v interface{}) (Converter, error) {
	switch colDef.DataType {
	case "integer", "numeric":
		switch t := v.(type) {
		case string:
			_, err := strconv.ParseFloat(t, 64)
			if err != nil {
				return nil, err
			}
			return func(in interface{}) (string, error) {
				return in.(string), nil
			}, nil
		case map[string]interface{}:
			return nil, errors.New("cannot convert map to number/numeric value")
		case []interface{}:
			return nil, errors.New("cannot convert array(slice) to number/numeric value")
		case nil:
			return toConverterByColumnDefForNil(colDef)
		case bool:
			return func(in interface{}) (string, error) {
				if in.(bool) {
					return "1", nil
				} else {
					return "0", nil
				}
			}, nil
		case float64:
			return func(in interface{}) (string, error) {
				return fmt.Sprint(in), nil
			}, nil
		}
	case "boolean":
		switch t := v.(type) {
		case string:
			_, err := strconv.ParseBool(t)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string %s to bool: %w", t, err)
			}
			return func(in interface{}) (string, error) {
				return in.(string), nil
			}, nil
		case map[string]interface{}:
			return nil, errors.New("cannot convert map to boolean value")
		case []interface{}:
			return nil, errors.New("cannot convert array(slice) to boolean value")
		case nil:
			return toConverterByColumnDefForNil(colDef)
		case bool:
			return func(in interface{}) (string, error) {
				if in.(bool) {
					return toSqlString("t"), nil
				} else {
					return toSqlString("f"), nil
				}
			}, nil
		case float64:
			return func(in interface{}) (string, error) {
				if int(in.(float64)) == 0 {
					return toSqlString("f"), nil
				}
				return toSqlString("t"), nil
			}, nil
		}
	case "array":
		switch t := v.(type) {
		case string:
			switch colDef.UdtName {
			case "_text", "_varchar":
				return func(in interface{}) (string, error) {
					return toSqlString("{", in.(string), "}"), nil
				}, nil
			default:
				return func(in interface{}) (string, error) {
					return fmt.Sprintf("{%v}", in.(string)), nil
				}, nil
			}
		case map[string]interface{}:
			_, err := json.Marshal(t)
			if err != nil {
				return nil, fmt.Errorf("cannot convert map value as json string representation: %w", err)
			}
			switch colDef.UdtName {
			case "_text", "_varchar":
				return func(in interface{}) (string, error) {
					v, err := json.Marshal(t)
					if err != nil {
						return "", err
					}
					return toSqlString("{", string(v), "}"), nil
				}, nil
			default:
				return func(in interface{}) (string, error) {
					v, err := json.Marshal(t)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("{%v}", v), nil
				}, nil
			}
		case []interface{}:
			switch colDef.UdtName {
			case "_text", "_varchar":
				return func(in interface{}) (string, error) {
					vals := make([]string, len(in.([]interface{})))
					for i, vi := range in.([]interface{}) {
						v, err := json.Marshal(vi)
						if err != nil {
							return "", err
						}
						vals[i] = toSqlString(string(v))
					}
					return fmt.Sprintf("{%s}", strings.Join(vals, ",")), nil
				}, nil
			default:
				return func(in interface{}) (string, error) {
					v, err := json.Marshal(in)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("ARRAY%s", string(v)), nil
				}, nil
			}
		case nil:
			return toConverterByColumnDefForNil(colDef)
		case bool:
			switch colDef.UdtName {
			case "_text", "_varchar":
				return func(in interface{}) (string, error) {
					return toSqlString("{", fmt.Sprintf("%v", in.(bool)), "}"), nil
				}, nil
			default:
				return func(in interface{}) (string, error) {
					return fmt.Sprintf("{%v}", in.(bool)), nil
				}, nil
			}
		case float64:
			switch colDef.UdtName {
			case "_text", "_varchar":
				return func(in interface{}) (string, error) {
					return toSqlString("{", fmt.Sprintf("%v", in.(float64)), "}"), nil
				}, nil
			default:
				return func(in interface{}) (string, error) {
					return fmt.Sprintf("{%v}", in.(float64)), nil
				}, nil
			}
		}
	case "text", "varchar":
		switch v.(type) {
		case string:
			return func(in interface{}) (string, error) {
				return toSqlString(in.(string)), nil
			}, nil
		case map[string]interface{}:
			return func(in interface{}) (string, error) {
				v, err := json.Marshal(in)
				if err != nil {
					return "", err
				}
				return toSqlString(string(v)), nil
			}, nil
		case []interface{}:
			return func(in interface{}) (string, error) {
				v, err := json.Marshal(in)
				if err != nil {
					return "", err
				}
				return toSqlString(string(v)), nil
			}, nil
		case nil:
			return toConverterByColumnDefForNil(colDef)
		case bool:
			return func(in interface{}) (string, error) {
				return toSqlString(fmt.Sprint(in)), nil
			}, nil
		case float64:
			return func(in interface{}) (string, error) {
				return toSqlString(fmt.Sprint(in)), nil
			}, nil
		}
	case "timestamp":
		switch v.(type) {
		case string:
			return func(in interface{}) (string, error) {
				return toSqlString(in.(string)), nil
			}, nil
		case map[string]interface{}:
			return nil, errors.New("cannot convert map to timestamp value")
		case []interface{}:
			return nil, errors.New("cannot convert array/slice to timestamp value")
		case nil:
			return toConverterByColumnDefForNil(colDef)
		case bool:
			return nil, errors.New("cannot convert boolean to timestamp value")
		case float64:
			return func(in interface{}) (string, error) {
				return fmt.Sprintf("to_timestamp(%d)", int(in.(float64))), nil
			}, nil
		}
	case "json", "jsonb":
		switch v.(type) {
		case string:
			return func(in interface{}) (string, error) {
				return toSqlString(in.(string)), nil
			}, nil
		case map[string]interface{}:
			return func(in interface{}) (string, error) {
				v, err := json.Marshal(in)
				if err != nil {
					return "", err
				}
				return toSqlString(string(v)), nil
			}, nil
		case []interface{}:
			return func(in interface{}) (string, error) {
				v, err := json.Marshal(in)
				if err != nil {
					return "", err
				}
				return toSqlString(string(v)), nil
			}, nil
		case nil:
			return toConverterByColumnDefForNil(colDef)
		case bool:
			return func(in interface{}) (string, error) {
				return toSqlString(fmt.Sprint(in)), nil
			}, nil
		case float64:
			return func(in interface{}) (string, error) {
				return toSqlString(fmt.Sprint(in)), nil
			}, nil
		}
	}
	return nil, fmt.Errorf("unsupported column type: %s(%s) and data type: %T", colDef.DataType, colDef.UdtName, v)
}

func toConverterByValue(dataType string, isKey bool, v interface{}) (Converter, error) {
	switch t := v.(type) {
	case nil:
		return toConverterByColumnDefForNil(sqlColumnType{DataType: dataType, IsNullable: !isKey})
	case string:
		return func(in interface{}) (string, error) {
			return toSqlString(in.(string)), nil
		}, nil
	case float64:
		if isKey {
			return func(in interface{}) (string, error) {
				return fmt.Sprint(int64(in.(float64))), nil
			}, nil
		}
		return func(in interface{}) (string, error) {
			return fmt.Sprint(in.(float64)), nil
		}, nil
	case []byte:
		return func(in interface{}) (string, error) {
			return fmt.Sprintf("'\\x%x'", in), nil
		}, nil
	case bool:
		return func(in interface{}) (string, error) {
			if in.(bool) {
				return toSqlString("t"), nil
			}
			return toSqlString("f"), nil
		}, nil
	case map[string]interface{}:
		return func(in interface{}) (string, error) {
			data, err := json.Marshal(in)
			if err != nil {
				return "", err
			}
			return toSqlString(string(data)), nil
		}, nil
	case []interface{}:
		return func(in interface{}) (string, error) {
			data, err := json.Marshal(t)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("ARRAY%s", string(data)), nil
		}, nil
	default:
		zap.L().Error("usupported type", zap.String("type", fmt.Sprintf("%T", t)))
		return nil, fmt.Errorf("usupported type: %T", t)
	}
}

func toSqlString(vals ...string) string {
	var (
		prefix string
		val    string
		suffix string
	)
	switch len(vals) {
	case 0:
		return ""
	case 1:
		val = vals[0]
	case 2:
		prefix = vals[0]
		val = vals[1]
	default:
		prefix = vals[0]
		val = vals[1]
		suffix = strings.Join(vals[2:], "")
	}

	return fmt.Sprintf("%s'%s'%s", prefix, strings.ReplaceAll(val, "'", "''"), suffix)
}

func toConverterByColumnDefForNil(colDef sqlColumnType) (Converter, error) {
	if colDef.IsNullable {
		return func(in interface{}) (string, error) {
			return "NULL", nil
		}, nil
	}
	switch colDef.DataType {
	case "integer", "numeric":
		return func(in interface{}) (string, error) {
			return "0", nil
		}, nil
	case "boolean":
		return func(in interface{}) (string, error) {
			return "'f'", nil
		}, nil
	case "array":
		return func(in interface{}) (string, error) {
			return "{}", nil
		}, nil
	case "text", "varchar":
		return func(in interface{}) (string, error) {
			return "", nil
		}, nil
	case "json", "jsonb":
		return func(in interface{}) (string, error) {
			return "{}", nil
		}, nil
	case "timestamp":
		return func(in interface{}) (string, error) {
			return "to_timestamp(0)", nil
		}, nil
	default:
		zap.L().Error("usupported type to convert nil to default", zap.String("data_type", colDef.DataType), zap.String("udt_name", colDef.UdtName))
		return nil, fmt.Errorf("usupported type to convert nil to default: %s (%s)", colDef.DataType, colDef.UdtName)
	}
}
