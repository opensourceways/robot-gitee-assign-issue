package main

import (
	"fmt"
	"strings"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	sdk "github.com/opensourceways/go-gitee/gitee"
)

const (
	msgMultipleAssignee           = "Can only assign one assignee to the issue."
	msgAssignRepeatedly           = "This issue is already assigned to ***%s***. Please do not assign repeatedly."
	msgNotAllowAssign             = "This issue can not be assigned to ***%s***. Please try to assign to the repository collaborators."
	msgNotAllowUnassign           = "***%s*** can not be unassigned from this issue. Please try to unassign the assignee of this issue."
	msgCollaboratorCantAsAssignee = "The issue collaborator ***%s*** cannot be assigned as the assignee at the same time."
)

func (bot *robot) handleAssign(e *sdk.NoteEvent) error {
	org, repo := e.GetOrgRepo()
	number := e.GetIssueNumber()

	currentAssignee := ""
	if e.Issue.Assignee != nil {
		currentAssignee = e.GetIssue().GetAssignee().GetLogin()
	}

	writeComment := func(s string) error {
		return bot.cli.CreateIssueComment(org, repo, number, s)
	}

	assign, unassign := parseCmd(e.GetComment().GetBody(), e.GetCommenter())
	if n := assign.Len(); n > 0 {
		if n > 1 {
			return writeComment(msgMultipleAssignee)
		}

		if assign.Has(currentAssignee) {
			return writeComment(fmt.Sprintf(msgAssignRepeatedly, currentAssignee))
		}

		newOne := assign.UnsortedList()[0]
		if isIssueCollaborator(e.GetIssue().GetCollaborators(), newOne) {
			return writeComment(fmt.Sprintf(msgCollaboratorCantAsAssignee, newOne))
		}

		err := bot.cli.AssignGiteeIssue(org, repo, number, newOne)
		if err == nil {
			return nil
		}
		if _, ok := err.(giteeclient.ErrorForbidden); ok {
			return writeComment(fmt.Sprintf(msgNotAllowAssign, newOne))
		}
		return err
	}

	if unassign.Len() > 0 {
		if unassign.Has(currentAssignee) {
			return bot.cli.UnassignGiteeIssue(org, repo, number, "")
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
