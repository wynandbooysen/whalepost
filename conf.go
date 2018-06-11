package main

// whalepost
// Copyright (C) 2018 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

type Conf struct {
	Auths map[string]*types.AuthConfig `json:"auths"`
}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

// Loads the configuration file.
func LoadConf(path string) (*Conf, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var conf Conf
	err = json.Unmarshal(buf, &conf)
	if err != nil {
		return nil, err
	}

	// parse the auth string into an *types.AuthConfig
	for key, val := range conf.Auths {
		password, username, err := decodeAuth(val.Auth)
		if err != nil {
			logrus.Errorln("decode auth failed:", err.Error())
			panic(err)
		}
		conf.Auths[key].Auth = ""
		conf.Auths[key].ServerAddress = key
		conf.Auths[key].Username = password
		conf.Auths[key].Password = username
	}

	return &conf, nil
}

// GetAuth returns the encoded authentican string for an index.
func (c *Conf) GetAuth(index string) (string, error) {
	auth, ok := c.Auths[index]
	if !ok {
		return "", errors.New("credentials not found")
	}

	buf, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buf), nil
}

// ---------------------------------------------------------------------------------------
//  private functions
// ---------------------------------------------------------------------------------------

// decodeAuth decodes a base64 encoded string and returns username and password
func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}

	decLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", errors.Errorf("Something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", errors.Errorf("Invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}
