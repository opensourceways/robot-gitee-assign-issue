package main

import (
	"errors"

	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

const botName = "issue-assign"

type iClient interface {
	AssignGiteeIssue(org, repo string, number string, login string) error
	UnassignGiteeIssue(org, repo string, number string, login string) error
	CreateIssueComment(org, repo string, number string, comment string) error
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config) (*configuration, error) {
	if c, ok := cfg.(*configuration); ok {
		return c, nil
	}
	return nil, errors.New("can't convert to configuration")
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegitster) {
	p.RegisterNoteEventHandler(bot.handleNoteEvent)
}

func (bot *robot) handleNoteEvent(e *sdk.NoteEvent, cfg config.Config, log *logrus.Entry) error {
	if !e.IsIssue() || !e.IsCreatingCommentEvent() {
		return nil
	}

	config, err := bot.getConfig(cfg)
	if err != nil {
		return err
	}

	if config.configFor(e.GetOrgRepo()) == nil {
		return nil
	}

	return bot.handleAssign(e)
}
