package utils

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cespare/xxhash/v2"
)

type Ids []uint64

func (a Ids) Len() int           { return len(a) }
func (a Ids) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Ids) Less(i, j int) bool { return a[i] < a[j] }

func ToJsonId(data []byte, keepSliceOrder bool) (uint64, error) {
	d := xxhash.New()
	var o map[string]interface{}
	err := json.Unmarshal(data, &o)
	if err != nil {
		return 0, err
	}
	err = writeObject(o, d, keepSliceOrder)
	if err != nil {
		return 0, err
	}
	return d.Sum64(), nil
}

func ObjectToJsonId(o interface{}, keepSliceOrder bool) (uint64, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return 0, err
	}
	return ToJsonId(data, keepSliceOrder)
}

func writeObject(o map[string]interface{}, d *xxhash.Digest, keepSliceOrder bool) error {
	keys := make([]string, 0, len(o))
	for k := range o {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := o[k]
		_, err := d.WriteString(k)
		if err != nil {
			return err
		}
		err = writeInterface(v, d, keepSliceOrder)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeSlice(o []interface{}, d *xxhash.Digest) error {
	if len(o) == 0 {
		_, err := d.WriteString("[]")
		return err
	}
	if len(o) == 1 {
		return writeInterface(o[0], d, false)
	}
	ids := make(Ids, len(o))
	idsIndices := make(map[uint64]int, len(o))
	tmpD := xxhash.New()
	for i, v := range o {
		tmpD.Reset()
		err := writeInterface(v, tmpD, false)
		if err != nil {
			return err
		}
		ids[i] = tmpD.Sum64()
		idsIndices[ids[i]] = i
	}
	sort.Sort(ids)
	for _, id := range ids {
		err := writeInterface(o[idsIndices[id]], d, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeSliceKeepOrder(o []interface{}, d *xxhash.Digest) error {
	if len(o) == 0 {
		_, err := d.WriteString("[]")
		return err
	}
	for _, v := range o {
		err := writeInterface(v, d, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeInterface(v interface{}, d *xxhash.Digest, keepSliceOrder bool) error {
	var err error
	switch t := v.(type) {
	case nil:
		_, err = d.WriteString("null")
	case map[string]interface{}:
		err = writeObject(t, d, keepSliceOrder)
	case []interface{}:
		if keepSliceOrder {
			err = writeSliceKeepOrder(t, d)
		} else {
			err = writeSlice(t, d)
		}
	default:
		_, err = d.WriteString(fmt.Sprint(v))
	}
	return err
}
