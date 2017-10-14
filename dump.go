package config

import (
	"bufio"
	"bytes"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/redforks/appinfo"
)

func getDefaultOptionKVs() map[string]Option {
	opts := make(map[string]Option, len(options))
	for _, rec := range options {
		opts[rec.name] = rec.creator()
	}
	return opts
}

func commentOutAll(prefix string, reader io.Reader) string {
	writer := bytes.Buffer{}
	_, _ = writer.WriteString(prefix) // no need to check error returned by Buffer WriteXXX() methods

	liner := bufio.NewScanner(reader)
	for liner.Scan() {
		if len(liner.Bytes()) != 0 {
			_, _ = writer.WriteString("# ")
		}
		_, _ = writer.Write(liner.Bytes())
		_, _ = writer.WriteRune('\n')
	}
	return writer.String()
}

// DumpDefaultOptions dump default options in config file format. All options
// are comment out.
func DumpDefaultOptions() (string, error) {
	opts := getDefaultOptionKVs()

	buf := bytes.Buffer{}
	encoder := toml.NewEncoder(&buf)
	encoder.Indent = ""
	if err := encoder.Encode(opts); err != nil {
		return "", err
	}

	return commentOutAll("# default options for "+appinfo.CodeName()+"\n\n", &buf), nil
}
