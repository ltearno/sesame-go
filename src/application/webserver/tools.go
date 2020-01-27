package webserver

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
)

func interpolateTemplate(name string, templateContent string, context interface{}) *string {
	t, err := template.New(name).Parse(templateContent)
	if err != nil {
		return nil
	}

	buffer := bytes.NewBufferString("")

	t.Execute(buffer, context)

	out, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil
	}

	result := string(out)

	return &result
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func encodeUint64ToBytes(v uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, v)
	i := 0
	for ; i < len(data); i++ {
		if data[i] != 0x0 {
			break
		}
	}

	return data[i:]
}

func encodeUint64ToString(v uint64) string {
	return base64.RawURLEncoding.EncodeToString(encodeUint64ToBytes(v))
}
