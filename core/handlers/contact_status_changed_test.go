package handlers_test

import (
	"testing"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/modifiers"
	"github.com/nyaruka/mailroom/core/handlers"
	"github.com/nyaruka/mailroom/core/models"
	"github.com/nyaruka/mailroom/testsuite"
)

func TestContactStatusChanged(t *testing.T) {

	db := testsuite.DB()

	// make sure cathyID contact is active
	db.Exec(`UPDATE contacts_contact SET status = 'A' WHERE id = $1`, models.CathyID)

	tcs := []handlers.TestCase{
		{
			Modifiers: handlers.ContactModifierMap{
				models.CathyID: []flows.Modifier{modifiers.NewStatus(flows.ContactStatusBlocked)},
			},
			SQLAssertions: []handlers.SQLAssertion{
				{
					SQL:   `select count(*) from contacts_contact where id = $1 AND status = 'B'`,
					Args:  []interface{}{models.CathyID},
					Count: 1,
				},
			},
		},
		{
			Modifiers: handlers.ContactModifierMap{
				models.CathyID: []flows.Modifier{modifiers.NewStatus(flows.ContactStatusStopped)},
			},
			SQLAssertions: []handlers.SQLAssertion{
				{
					SQL:   `select count(*) from contacts_contact where id = $1 AND status = 'S'`,
					Args:  []interface{}{models.CathyID},
					Count: 1,
				},
			},
		},
		{
			Modifiers: handlers.ContactModifierMap{
				models.CathyID: []flows.Modifier{modifiers.NewStatus(flows.ContactStatusActive)},
			},
			SQLAssertions: []handlers.SQLAssertion{
				{
					SQL:   `select count(*) from contacts_contact where id = $1 AND status = 'A'`,
					Args:  []interface{}{models.CathyID},
					Count: 1,
				},
				{
					SQL:   `select count(*) from contacts_contact where id = $1 AND status = 'A'`,
					Args:  []interface{}{models.CathyID},
					Count: 1,
				},
			},
		},
	}

	handlers.RunTestCases(t, tcs)
}
