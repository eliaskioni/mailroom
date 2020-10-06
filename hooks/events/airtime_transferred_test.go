package events_test

import (
	"testing"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/mailroom/models"
	"github.com/nyaruka/mailroom/testsuite"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/require"
)

func TestAirtimeTransferred(t *testing.T) {
	ctx := testsuite.CTX()
	db := testsuite.DB()

	oa, err := models.GetOrgAssets(ctx, db, models.Org1)
	require.NoError(t, err)

	contacts, err := models.LoadContacts(ctx, db, oa, []models.ContactID{models.CathyID})
	require.NoError(t, err)

	contact, err := contacts[0].FlowContact(oa)
	require.NoError(t, err)

	scene := models.NewSceneForContact(contact)

	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)

	err = models.HandleEvents(ctx, tx, nil, oa, scene, []flows.Event{
		events.NewAirtimeTransferred(&flows.AirtimeTransfer{
			Sender:        "tel:+1234567890",
			Recipient:     "tel:+2345678901",
			Currency:      "RWF",
			DesiredAmount: decimal.RequireFromString(`110`),
			ActualAmount:  decimal.RequireFromString(`100`),
		}, nil),
	})
	require.NoError(t, err)

	// TODO assert precommit hooks?

	require.NoError(t, tx.Rollback())
}
