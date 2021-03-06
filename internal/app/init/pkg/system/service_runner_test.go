/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package system_test

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/talos-systems/talos/internal/app/init/pkg/system"
	"github.com/talos-systems/talos/internal/app/init/pkg/system/conditions"
	"github.com/talos-systems/talos/internal/app/init/pkg/system/events"
)

type ServiceRunnerSuite struct {
	suite.Suite
}

func (suite *ServiceRunnerSuite) assertStateSequence(expectedStates []events.ServiceState, sr *system.ServiceRunner) {
	states := []events.ServiceState{}

	for _, event := range sr.GetEventHistory(1000) {
		states = append(states, event.State)
	}

	suite.Assert().Equal(expectedStates, states)
}

func (suite *ServiceRunnerSuite) TestFullFlow() {
	sr := system.NewServiceRunner(&MockService{}, nil)

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		sr.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	select {
	case <-finished:
		suite.Require().Fail("service running should be still running")
	default:
	}

	sr.Shutdown()

	<-finished

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateWaiting,
		events.StatePreparing,
		events.StateRunning,
		events.StateFinished,
	}, sr)
}

func (suite *ServiceRunnerSuite) TestFullFlowHealthy() {
	sr := system.NewServiceRunner(&MockHealthcheckedService{}, nil)

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		sr.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	select {
	case <-finished:
		suite.Require().Fail("service running should be still running")
	default:
	}

	sr.Shutdown()

	<-finished

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateWaiting,
		events.StatePreparing,
		events.StateRunning,
		events.StateRunning, // one more notification when service is healthy
		events.StateFinished,
	}, sr)
}

func (suite *ServiceRunnerSuite) TestFullFlowHealthChanges() {
	m := MockHealthcheckedService{}
	sr := system.NewServiceRunner(&m, nil)

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		sr.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	m.SetHealthy(false)

	time.Sleep(50 * time.Millisecond)

	m.SetHealthy(true)

	time.Sleep(50 * time.Millisecond)

	sr.Shutdown()

	<-finished

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateWaiting,
		events.StatePreparing,
		events.StateRunning,
		events.StateRunning, // initial: healthy
		events.StateRunning, // not healthy
		events.StateRunning, // one again healthy
		events.StateFinished,
	}, sr)
}

func (suite *ServiceRunnerSuite) TestPreStageFail() {
	svc := &MockService{
		preError: errors.New("pre failed"),
	}
	sr := system.NewServiceRunner(svc, nil)
	sr.Start()

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateFailed,
	}, sr)
}

func (suite *ServiceRunnerSuite) TestRunnerStageFail() {
	svc := &MockService{
		runnerError: errors.New("runner failed"),
	}
	sr := system.NewServiceRunner(svc, nil)
	sr.Start()

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateWaiting,
		events.StatePreparing,
		events.StateFailed,
	}, sr)
}

func (suite *ServiceRunnerSuite) TestAbortOnCondition() {
	svc := &MockService{
		condition: conditions.WaitForFileToExist("/doesntexistever"),
	}
	sr := system.NewServiceRunner(svc, nil)

	finished := make(chan struct{})

	go func() {
		defer close(finished)
		sr.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	select {
	case <-finished:
		suite.Require().Fail("service running should be still running")
	default:
	}

	sr.Shutdown()

	<-finished

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateWaiting,
		events.StateFailed,
	}, sr)
}

func (suite *ServiceRunnerSuite) TestPostStateFail() {
	svc := &MockService{
		postError: errors.New("post failed"),
	}
	sr := system.NewServiceRunner(svc, nil)

	finished := make(chan struct{})

	go func() {
		defer close(finished)
		sr.Start()
	}()

	sr.Shutdown()

	<-finished

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateWaiting,
		events.StatePreparing,
		events.StateRunning,
		events.StateFinished,
		events.StateFailed,
	}, sr)
}

func (suite *ServiceRunnerSuite) TestRunFail() {
	runner := &MockRunner{exitCh: make(chan error)}
	svc := &MockService{runner: runner}
	sr := system.NewServiceRunner(svc, nil)

	finished := make(chan struct{})

	go func() {
		defer close(finished)
		sr.Start()
	}()

	runner.exitCh <- errors.New("run failed")

	<-finished

	suite.assertStateSequence([]events.ServiceState{
		events.StatePreparing,
		events.StateWaiting,
		events.StatePreparing,
		events.StateRunning,
		events.StateFailed,
	}, sr)
}

func TestServiceRunnerSuite(t *testing.T) {
	suite.Run(t, new(ServiceRunnerSuite))
}
