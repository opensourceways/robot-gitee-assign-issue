package main

import (
	"fmt"
	"strings"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	msgAssignDone                 = "This issue is assigned to: ***%s***."
	msgMultipleAssignee           = "Can only assign one assignee to the issue."
	msgAssignRepeatedly           = "This issue is already assigned to ***%s***. Please do not assign repeatedly."
	msgNotAllowAssign             = "This issue can not be assigned to ***%s***. Please try to assign to the repository collaborators."
	msgUnassignDone               = "***%s*** is unassigned from this issue."
	msgNotAllowUnassign           = "***%s*** can not be unassigned from this issue. Please try to unassign the assignee from this issue."
	msgCollaboratorCantAsAssignee = "The issue collaborator ***%s*** cannot be assigned as the assignee at the same time."
)

func (bot *robot) handleAssign(e *sdk.NoteEvent) error {
	ne := giteeclient.NewIssueNoteEvent(e)
	org, repo := ne.GetOrgRep()
	number := ne.GetIssueNumber()

	currentAssignee := ""
	if e.Issue.Assignee != nil {
		currentAssignee = e.Issue.Assignee.Login
	}

	currentCollaborators := func() sets.String {
		c := sets.NewString()
		for _, v := range e.Issue.Collaborators {
			c.Insert(v.Login)
		}
		return c
	}

	writeComment := func(s string) error {
		return bot.cli.CreateIssueComment(org, repo, number, s)
	}

	assign, unassign := parseCmd(ne.GetComment(), ne.GetCommenter())
	if n := assign.Len(); n > 0 {
		if n > 1 {
			return writeComment(msgMultipleAssignee)
		}

		if assign.Has(currentAssignee) {
			return writeComment(fmt.Sprintf(msgAssignRepeatedly, currentAssignee))
		}

		newOne := assign.UnsortedList()[0]
		if isIssueCollaborator(e.Issue.Collaborators, newOne) {
			return writeComment(fmt.Sprintf(msgCollaboratorCantAsAssignee, newOne))
		}

		err := bot.cli.AssignGiteeIssue(org, repo, number, newOne)
		if err == nil {
			return writeComment(fmt.Sprintf(msgAssignDone, newOne))
		}
		if _, ok := err.(giteeclient.ErrorForbidden); ok {
			return writeComment(fmt.Sprintf(msgNotAllowAssign, newOne))
		}
		return err
	}

	if unassign.Len() > 0 {
		if unassign.Has(currentAssignee) {
			if err := bot.cli.UnassignGiteeIssue(org, repo, number, ""); err != nil {
				return err
			}
			return writeComment(fmt.Sprintf(msgUnassignDone, currentAssignee))
		} else {
			return writeComment(fmt.Sprintf(msgNotAllowUnassign, strings.Join(unassign.UnsortedList(), ",")))
		}
	}

	return nil
}

func isIssueCollaborator(collaborators []sdk.UserHook, assignee string) bool {
	for _, v := range collaborators {
		if v.Name == assignee {
			return true
		}
	}

	return false
}
