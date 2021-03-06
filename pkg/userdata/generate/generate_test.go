/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package generate_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/talos-systems/talos/pkg/userdata"
	"github.com/talos-systems/talos/pkg/userdata/generate"
	"gopkg.in/yaml.v2"
)

var (
	input *generate.Input
)

type GenerateSuite struct {
	suite.Suite
}

func TestGenerateSuite(t *testing.T) {
	suite.Run(t, new(GenerateSuite))
}

func (suite *GenerateSuite) SetupSuite() {
	var err error
	input, err = generate.NewInput("test", []string{"10.0.1.5", "10.0.1.6", "10.0.1.7"})
	suite.Require().NoError(err)
}

func (suite *GenerateSuite) TestGenerateInitSuccess() {
	dataString, err := generate.Userdata(generate.TypeInit, input)
	suite.Require().NoError(err)
	data := &userdata.UserData{}
	err = yaml.Unmarshal([]byte(dataString), data)
	suite.Require().NoError(err)
}

func (suite *GenerateSuite) TestGenerateControlPlaneSuccess() {
	dataString, err := generate.Userdata(generate.TypeControlPlane, input)
	suite.Require().NoError(err)
	data := &userdata.UserData{}
	err = yaml.Unmarshal([]byte(dataString), data)
	suite.Require().NoError(err)
}

func (suite *GenerateSuite) TestGenerateWorkerSuccess() {
	dataString, err := generate.Userdata(generate.TypeJoin, input)
	suite.Require().NoError(err)
	data := &userdata.UserData{}
	err = yaml.Unmarshal([]byte(dataString), data)
	suite.Require().NoError(err)
}

func (suite *GenerateSuite) TestGenerateTalosconfigSuccess() {
	_, err := generate.Talosconfig(input)
	suite.Require().NoError(err)
}
