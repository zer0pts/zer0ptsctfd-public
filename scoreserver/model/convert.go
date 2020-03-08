package model

import (
	"bytes"
	"text/template"
)

func UserChallenge(chal *Challenge) (*UserChallengeInfo, error) {
	t, err := template.New(chal.Name).Parse(chal.Description)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, map[string]interface{}{
		"Port": chal.Port,
		"Host": chal.Host,
	})
	if err != nil {
		return nil, err
	}
	return &UserChallengeInfo{
		ID:            chal.ID,
		Name:          chal.Name,
		Description:   buf.String(),
		Author:        chal.Author,
		Category:      chal.Category,
		Difficulty:    chal.Difficulty,
		Score:         chal.Score,
		Attachments:   chal.Attachments,
		Tags:          chal.Tags,
		IsQuestionary: chal.IsQuestionary,
		SolveTeams:    chal.SolveTeams,
	}, nil
}
